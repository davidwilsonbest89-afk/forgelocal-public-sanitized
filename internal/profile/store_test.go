package profile

import (
	"os"
	"testing"
)

func TestProfileCRUD(t *testing.T) {
	dir, _ := os.MkdirTemp("", "browseforge-test-*")
	defer os.RemoveAll(dir)

	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create
	p := &Profile{Name: "Test", Engine: "firefox"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	if p.ID == "" {
		t.Fatal("ID not generated")
	}
	t.Logf("Created: %s", p.ID)

	// Read
	got, err := store.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test" {
		t.Errorf("name = %s, want Test", got.Name)
	}

	// List
	all := store.List("", "")
	if len(all) != 1 {
		t.Errorf("list = %d, want 1", len(all))
	}

	// Update
	updated, err := store.Update(p.ID, map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated" {
		t.Errorf("name = %s, want Updated", updated.Name)
	}

	// Duplicate
	dup, err := store.Duplicate(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dup.ID == p.ID {
		t.Error("duplicate has same ID")
	}

	// Delete
	if err := store.Delete(p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(p.ID); err == nil {
		t.Error("deleted profile still found")
	}

	t.Log("✅ Profile CRUD all pass")
}
