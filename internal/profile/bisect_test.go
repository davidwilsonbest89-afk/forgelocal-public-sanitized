package profile

import (
	"testing"
)

// TestBisectCRUDSteps isolates each CRUD operation to locate a hang.
func TestBisectCRUDSteps(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("newstore: %v", err)
	}
	t.Log("1 newstore ok")

	p := &Profile{Name: "Test", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Logf("2 create ok %s", p.ID)

	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	t.Logf("3 get ok %s", got.Name)

	all := store.List("", "")
	t.Logf("4 list ok %d", len(all))

	updated, err := store.Update(p.ID, map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	t.Logf("5 update ok %s", updated.Name)

	t.Log("6a about to duplicate")
	dup, err := store.Duplicate(p.ID)
	t.Logf("6b duplicate done err=%v id=%s", err, dup.ID)

	if err := store.Delete(p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	t.Log("7 delete ok")
}
