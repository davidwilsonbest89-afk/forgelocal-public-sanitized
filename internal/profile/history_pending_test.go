package profile

import "testing"

func TestHistoryPendingPersistsUntilExplicitConfirmation(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil { t.Fatal(err) }
	p := &Profile{Name: "History Pending", RuntimeID: "chromium", LifecycleState: LifecycleActive}
	if err := s.Create(p); err != nil { t.Fatal(err) }
	if !p.HistoryPending { t.Fatal("create must persist a pending history marker") }
	if err := s.ClearHistoryPending(p.ID); err != nil { t.Fatal(err) }
	if got, err := s.Get(p.ID); err != nil || got.HistoryPending { t.Fatalf("confirmation = %#v %v", got, err) }
	updated, err := s.Update(p.ID, map[string]any{"name": "History Pending Updated"})
	if err != nil { t.Fatal(err) }
	if !updated.HistoryPending { t.Fatal("profile mutation must restore pending marker") }
	reopened, err := NewStore(s.dir)
	if err != nil { t.Fatal(err) }
	if pending := reopened.PendingHistoryProfiles(); len(pending) != 1 || pending[0].ID != p.ID { t.Fatalf("pending after restart = %#v", pending) }
}
