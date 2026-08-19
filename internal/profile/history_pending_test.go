package profile

import (
	"os"
	"testing"
	"time"
)

func TestHistoryPendingPersistsUntilExplicitConfirmation(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	p := &Profile{Name: "History Pending", RuntimeID: "chromium", LifecycleState: LifecycleActive}
	if err := s.Create(p); err != nil { t.Fatal(err) }
	if p.HistoryPending == nil || p.HistoryPending.OperationID == "" || p.HistoryPending.Action != "create" || p.HistoryPending.SnapshotDigest == "" { t.Fatal("create must persist a complete pending operation") }
	if err := s.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	if got, err := s.Get(p.ID); err != nil || got.HistoryPending != nil { t.Fatalf("confirmation = %#v %v", got, err) }
	updated, err := s.Update(p.ID, map[string]any{"name": "History Pending Updated"})
	if err != nil { t.Fatal(err) }
	if updated.HistoryPending == nil || updated.HistoryPending.OperationID == "" || updated.HistoryPending.Action != "update" { t.Fatal("profile mutation must restore a complete pending operation") }
	reopened, err := NewStore(s.dir)
	if err != nil { t.Fatal(err) }
	if pending := reopened.PendingHistoryProfiles(); len(pending) != 1 || pending[0].ID != p.ID { t.Fatalf("pending after restart = %#v", pending) }
}

func TestClearHistoryPendingFailureRetainsMarker(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	p := &Profile{Name: "Clear failure", RuntimeID: "chromium", LifecycleState: LifecycleActive}
	if err := s.Create(p); err != nil { t.Fatal(err) }
	operationID := p.HistoryPending.OperationID
	if err := os.RemoveAll(p.ProfileDir); err != nil { t.Fatal(err) }
	if err := s.ClearHistoryPending(p.ID, operationID); err == nil { t.Fatal("clear write failure must be returned") }
	current, err := s.Get(p.ID)
	if err != nil || current.HistoryPending == nil || current.HistoryPending.OperationID != operationID { t.Fatalf("failed clear lost pending operation: %#v %v", current, err) }
}

func TestWithHistorySequenceSerializesCallerCriticalSection(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	unlock, err := s.WithHistorySequence("prof_sequence")
	if err != nil { t.Fatal(err) }
	entered := make(chan struct{})
	go func() {
		second, err := s.WithHistorySequence("prof_sequence")
		if err == nil {
			close(entered)
			second()
		}
	}()
	select {
	case <-entered:
		t.Fatal("second sequence entered before the first released")
	case <-time.After(25 * time.Millisecond):
	}
	unlock()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("second sequence did not acquire after release")
	}
}
