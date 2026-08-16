package profile

import (
	"runtime"
	"testing"
	"time"
)

// TestDupAfterUpdate reproduces the Update -> Duplicate hang with a
// goroutine dump to locate the blocking mutex.
func TestDupAfterUpdate(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{Name: "Test", RuntimeID: "camoufox"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(p.ID, map[string]any{"name": "Updated"}); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		dup, err := store.Duplicate(p.ID)
		t.Logf("dup err=%v id=%s", err, dup.ID)
		close(done)
	}()

	select {
	case <-done:
		t.Log("duplicate completed")
	case <-time.After(8 * time.Second):
		buf := make([]byte, 64*1024)
		n := runtime.Stack(buf, true)
		t.Fatalf("hang after update; goroutine stacks:\n%s", buf[:n])
	}
}
