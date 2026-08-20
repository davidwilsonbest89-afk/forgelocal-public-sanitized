package profile

import (
	"os"
	"testing"
)

func TestT23ArchivePersistsBeforePublishingAndReopenClearsTimestamp(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	p := &Profile{Name: "T23 lifecycle", RuntimeID: "chromium", LifecycleState: LifecycleActive}
	if err := s.Create(p); err != nil { t.Fatal(err) }
	archived, changed, err := s.ArchiveProfileResult(p.ID)
	if err != nil || !changed || archived.LifecycleState != LifecycleArchived || archived.ArchivedAt == nil { t.Fatalf("archive result=%#v changed=%v err=%v", archived, changed, err) }
	reopenedStore, err := NewStore(s.dir)
	if err != nil { t.Fatal(err) }
	onDisk, err := reopenedStore.Get(p.ID)
	if err != nil || onDisk.LifecycleState != LifecycleArchived || onDisk.ArchivedAt == nil { t.Fatalf("archived persistence=%#v %v", onDisk, err) }
	noOp, changed, err := reopenedStore.ArchiveProfileResult(p.ID)
	if err != nil || changed || noOp.ArchivedAt == nil { t.Fatalf("idempotent archive=%#v changed=%v err=%v", noOp, changed, err) }
	reopened, err := reopenedStore.ReopenProfileResult(p.ID)
	if err != nil || reopened.LifecycleState != LifecycleActive || reopened.ArchivedAt != nil { t.Fatalf("reopen result=%#v err=%v", reopened, err) }
	postReopenStore, err := NewStore(reopenedStore.dir)
	if err != nil { t.Fatal(err) }
	postReopen, err := postReopenStore.Get(p.ID)
	if err != nil || postReopen.LifecycleState != LifecycleActive || postReopen.ArchivedAt != nil { t.Fatalf("post-reopen persistence=%#v %v", postReopen, err) }
}

func TestT23ArchiveWriteFailureDoesNotPublishPartialState(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	p := &Profile{Name: "T23 persistence failure", RuntimeID: "chromium", LifecycleState: LifecycleActive}
	if err := s.Create(p); err != nil { t.Fatal(err) }
	if err := os.RemoveAll(p.ProfileDir); err != nil { t.Fatal(err) }
	if _, _, err := s.ArchiveProfileResult(p.ID); err == nil { t.Fatal("archive write failure must be returned") }
	current, err := s.Get(p.ID)
	if err != nil || current.LifecycleState != LifecycleActive || current.ArchivedAt != nil { t.Fatalf("failed archive published partial state: %#v %v", current, err) }
}
