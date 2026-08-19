package templates

import (
	"context"
	"errors"
	"sync"
	"testing"

	"forgelocal/internal/profile"
)

func stringPtr(value string) *string     { return &value }
func tagsPtr(values ...string) *[]string { return &values }

func sampleContent(note string) Content {
	return Content{
		Group: stringPtr("qa"),
		Tags:  tagsPtr("release", "smoke"),
		Note:  stringPtr(note),
		CustomFields: map[string]profile.CustomField{
			"owner": {Type: "text", Value: "quality"},
			"tier":  {Type: "select", Value: "gold", Options: []string{"silver", "gold"}},
		},
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestVersioningIsImmutableAndArchiveReleasesName(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	v1, err := store.Create(ctx, "QA default", sampleContent("first"), "corr-create")
	if err != nil {
		t.Fatal(err)
	}
	v2, err := store.NewVersion(ctx, v1.TemplateID, 1, sampleContent("second"), "corr-v2")
	if err != nil {
		t.Fatal(err)
	}
	if v2.Version != 2 || v2.State != StateActive {
		t.Fatalf("v2=%+v", v2)
	}
	gotV1, err := store.GetVersion(ctx, v1.TemplateID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotV1.State != StateArchived || gotV1.Content.Note == nil || *gotV1.Content.Note != "first" {
		t.Fatalf("v1 was modified instead of archived: %+v", gotV1)
	}
	if err := store.Archive(ctx, v1.TemplateID, 2, "corr-archive"); err != nil {
		t.Fatal(err)
	}
	listing, err := store.List(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(listing.Data) != 1 || listing.Data[0].ActiveVersion != nil {
		t.Fatalf("catalog after archive=%+v", listing)
	}
	if _, err := store.Create(ctx, "qa DEFAULT", sampleContent("replacement"), "corr-reuse"); err != nil {
		t.Fatalf("fully archived series must release normalized name: %v", err)
	}
}

func TestNewVersionRollsBackVersionArchiveAndAudit(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	v1, err := store.Create(ctx, "rollback", sampleContent("one"), "corr-create")
	if err != nil {
		t.Fatal(err)
	}
	store.SetFailBeforeCommitForTest(func() error { return errors.New("injected before commit") })
	if _, err := store.NewVersion(ctx, v1.TemplateID, 1, sampleContent("two"), "corr-v2"); err == nil {
		t.Fatal("expected injected rollback")
	}
	store.SetFailBeforeCommitForTest(nil)
	got, err := store.GetVersion(ctx, v1.TemplateID, 1)
	if err != nil || got.State != StateActive {
		t.Fatalf("v1 must remain active after rollback: %+v / %v", got, err)
	}
	if _, err := store.GetVersion(ctx, v1.TemplateID, 2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("rolled back version 2 err=%v", err)
	}
	count, err := store.AuditCount(ctx, v1.TemplateID)
	if err != nil || count != 1 {
		t.Fatalf("audit count after rollback=%d err=%v", count, err)
	}
}

func TestConcurrentVersionsHaveOneWinner(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	v1, err := store.Create(ctx, "concurrency", sampleContent("one"), "corr-create")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.NewVersion(ctx, v1.TemplateID, 1, sampleContent("two"), "corr-race")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	var success, stale int
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, ErrStaleVersion) {
			stale++
		} else {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if success != 1 || stale != 1 {
		t.Fatalf("success=%d stale=%d", success, stale)
	}
}

func TestDraftTagsUnionAndConflictsAreNonWriting(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	template := sampleContent("template note")
	template.Tags = tagsPtr("release", "template")
	v1, err := store.Create(ctx, "draft", template, "corr-create")
	if err != nil {
		t.Fatal(err)
	}
	base := Content{Tags: tagsPtr("release", "base")}
	result, err := store.Draft(ctx, v1.TemplateID, 1, base, "corr-draft")
	if err != nil {
		t.Fatal(err)
	}
	if result.Draft == nil || result.Draft.Tags == nil || len(*result.Draft.Tags) != 3 {
		t.Fatalf("tag union=%+v", result)
	}
	conflictBase := Content{Note: stringPtr("different")}
	result, err = store.Draft(ctx, v1.TemplateID, 1, conflictBase, "corr-conflict")
	if !errors.Is(err, ErrConflict) || result == nil || result.Draft != nil || len(result.Conflicts) != 1 || result.Conflicts[0].Path != "note" {
		t.Fatalf("conflict result=%+v err=%v", result, err)
	}
	var paths string
	if err := store.db.QueryRow(`SELECT paths_json FROM template_audit_events WHERE template_id=? AND result='conflict'`, v1.TemplateID).Scan(&paths); err != nil {
		t.Fatal(err)
	}
	if paths != `["note"]` {
		t.Fatalf("audit must contain only paths, got %s", paths)
	}
}
