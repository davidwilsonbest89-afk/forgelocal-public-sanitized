package profile

import (
	"errors"
	"testing"
)

func TestSetMetadataPersistsTypedFields(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Metadata", RuntimeID: "chromium"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	fields := map[string]CustomField{
		"owner":    {Type: "text", Value: "qa-team"},
		"retries":  {Type: "number", Value: float64(3)},
		"approved": {Type: "boolean", Value: true},
		"tier":     {Type: "select", Value: "gold", Options: []string{"silver", "gold"}},
	}
	if _, err := store.SetMetadata(p.ID, "Local QA profile\nNo secrets.", fields); err != nil {
		t.Fatal(err)
	}
	// Mutating the caller map after the write must not mutate the stored profile.
	fields["owner"] = CustomField{Type: "text", Value: "changed"}
	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reloaded.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Note != "Local QA profile\nNo secrets." || got.CustomFields["owner"].Value != "qa-team" {
		t.Fatalf("persisted metadata = %#v / %#v", got.Note, got.CustomFields)
	}
}

func TestSetMetadataRejectsInvalidFieldsAndArchivedProfiles(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Reject", RuntimeID: "chromium"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	invalid := []map[string]CustomField{
		{"bad": {Type: "select", Value: "outside", Options: []string{"inside"}}},
		{"bad": {Type: "number", Value: "three"}},
		{"bad": {Type: "unknown", Value: true}},
	}
	for _, fields := range invalid {
		if _, err := store.SetMetadata(p.ID, "", fields); !errors.Is(err, ErrInvalidCustomField) {
			t.Fatalf("invalid fields err = %v", err)
		}
	}
	if _, err := store.SetMetadata(p.ID, string(make([]byte, maxProfileNoteBytes+1)), nil); !errors.Is(err, ErrInvalidNote) {
		t.Fatalf("oversized note err = %v", err)
	}
	if err := store.ArchiveProfile(p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetMetadata(p.ID, "refused", nil); !IsLifecycleError(err) {
		t.Fatalf("archived metadata err = %v", err)
	}
}
