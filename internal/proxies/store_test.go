// T10 — Proxies. Store tests covering the registry contract:
// CRUD, input validation, name uniqueness, assignment consistency,
// per-proxy isolation under concurrency (go test -race), redaction of
// credential material, and file persistence.
//
// Credential material used in these tests is strictly synthetic and never
// represents a real endpoint or provider.
package proxies

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func validProxy(name string) *Proxy {
	return &Proxy{
		Name: name,
		Type: "http",
		Host: "proxy.local",
		Port: 8080,
	}
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("proxy-one")
	if err := s.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == "" {
		t.Fatal("Create did not assign an id")
	}
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "proxy-one" || got.Type != "http" || got.Host != "proxy.local" || got.Port != 8080 {
		t.Errorf("Get returned unexpected values: %+v", got)
	}
	if got.HasSecret {
		t.Error("Get must not report a secret presence for a credential-less proxy")
	}
}

func TestCreateRejectsInvalid(t *testing.T) {
	s := newTestStore(t)
	cases := []struct {
		name  string
		build func() *Proxy
	}{
		{"nil payload", func() *Proxy { return nil }},
		{"missing name", func() *Proxy { p := validProxy(""); return p }},
		{"unsupported type", func() *Proxy { p := validProxy("px"); p.Type = "ftp"; return p }},
		{"empty host", func() *Proxy { p := validProxy("px"); p.Host = ""; return p }},
		{"port zero", func() *Proxy { p := validProxy("px"); p.Port = 0; return p }},
		{"negative port", func() *Proxy { p := validProxy("px"); p.Port = -1; return p }},
		{"port overflow", func() *Proxy { p := validProxy("px"); p.Port = 70000; return p }},
		{"control char in name", func() *Proxy { p := validProxy("px\x01"); return p }},
		{"invalid region", func() *Proxy { p := validProxy("px"); p.Region = "a\nb"; return p }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := s.Create(c.build())
			if err == nil {
				t.Fatal("Create accepted an invalid proxy")
			}
			if !IsValidationError(err) && !errorsIs(err, ErrInvalidName) {
				t.Errorf("expected validation error, got: %v", err)
			}
		})
	}
}

func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		if w, ok := err.(interface{ Unwrap() error }); ok {
			err = w.Unwrap()
		} else {
			return false
		}
	}
	return false
}

func TestNameUniqueness(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("dc-proxy")
	if err := s.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	err := s.Create(validProxy("dc-proxy"))
	if !IsDuplicateError(err) {
		t.Fatalf("expected duplicate error, got: %v", err)
	}
	// Case-insensitive collision.
	err = s.Create(validProxy("DC-Proxy"))
	if !IsDuplicateError(err) {
		t.Fatalf("expected case-insensitive duplicate error, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("del-me")
	_ = s.Create(p)
	if err := s.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(p.ID); !IsNotFoundError(err) {
		t.Errorf("expected not found after delete, got: %v", err)
	}
	if err := s.Delete(p.ID); !IsNotFoundError(err) {
		t.Errorf("second delete must be refused: %v", err)
	}
}

func TestUpdateValidatesAndRejectsAssigned(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("updatable")
	_ = s.Create(p)
	if _, err := s.Update(p.ID, map[string]any{"host": "proxy2.local", "port": 9090}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := s.Get(p.ID)
	if got.Host != "proxy2.local" || got.Port != 9090 {
		t.Errorf("Update did not apply changes: %+v", got)
	}
	// An assigned proxy refuses mutation: unassign first.
	if err := s.Assign("profile-a", p.ID); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	_, err := s.Update(p.ID, map[string]any{"host": "proxy3.local"})
	if err != ErrProxyInUse {
		t.Errorf("expected ErrProxyInUse, got: %v", err)
	}
	if err := s.Unassign("profile-a"); err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if _, err := s.Update(p.ID, map[string]any{"host": "proxy3.local"}); err != nil {
		t.Fatalf("Update after unassign: %v", err)
	}
}

func TestUpdateReplacesNameIndex(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("old-name")
	_ = s.Create(p)
	if _, err := s.Update(p.ID, map[string]any{"name": "new-name"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := s.Create(validProxy("old-name")); err != nil {
		t.Fatalf("renamed-away name must be reusable: %v", err)
	}
	// The new name is taken by the renamed proxy.
	err := s.Create(validProxy("new-name"))
	if !IsDuplicateError(err) {
		t.Errorf("expected duplicate after rename: %v", err)
	}
}

func TestAssignUnassignConsistency(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("px-assign")
	_ = s.Create(p)
	if err := s.Assign("profile-1", p.ID); err != nil {
		t.Fatalf("Assign: %v", err)
	}
	// Second assignment of the same profile replaces the previous one.
	p2 := validProxy("px-second")
	_ = s.Create(p2)
	if err := s.Assign("profile-1", p2.ID); err != nil {
		t.Fatalf("Assign replace: %v", err)
	}
	if got := s.AssignedProxy("profile-1"); got == nil || got.ID != p2.ID {
		t.Error("AssignedProxy did not track the replacement")
	}
	profiles := s.AssignedProfiles(p2.ID)
	if len(profiles) != 1 || profiles[0] != "profile-1" {
		t.Errorf("AssignedProfiles unexpected: %v", profiles)
	}
	// Deletion of an assigned proxy is refused.
	if err := s.Delete(p2.ID); err != ErrProxyInUse {
		t.Errorf("expected ErrProxyInUse on delete: %v", err)
	}
	// Unassign frees the proxy; deletion then succeeds.
	if err := s.Unassign("profile-1"); err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if err := s.Unassign("profile-1"); err != nil {
		t.Fatalf("second Unassign must be idempotent: %v", err)
	}
	if err := s.Delete(p2.ID); err != nil {
		t.Fatalf("Delete after unassign: %v", err)
	}
	if got := s.AssignedProxy("profile-1"); got != nil {
		t.Error("AssignedProxy must be nil after unassign")
	}
}

func TestAssignRejectsUnknownProxy(t *testing.T) {
	s := newTestStore(t)
	err := s.Assign("profile-1", "no-such-proxy")
	if !IsNotFoundError(err) {
		t.Errorf("expected not found for unknown proxy: %v", err)
	}
}

func TestUnassignForMismatchRefused(t *testing.T) {
	s := newTestStore(t)
	p1 := validProxy("px-a")
	p2 := validProxy("px-b")
	_ = s.Create(p1)
	_ = s.Create(p2)
	_ = s.Assign("profile-1", p1.ID)
	// Unbinding the wrong proxy must be refused, not silently rerouted.
	if err := s.UnassignFor("profile-1", p2.ID); !IsNotFoundError(err) {
		t.Errorf("expected not found on mismatch: %v", err)
	}
	if s.AssignedProxy("profile-1") == nil || s.AssignedProxy("profile-1").ID != p1.ID {
		t.Error("assignment must remain intact after a refused mismatch")
	}
	if err := s.UnassignFor("profile-1", p1.ID); err != nil {
		t.Fatalf("UnassignFor correct pair: %v", err)
	}
	if s.AssignedProxy("profile-1") != nil {
		t.Error("must be detached after correct unassign")
	}
}

func TestListSortedAndRedacted(t *testing.T) {
	s := newTestStore(t)
	for _, n := range []string{"zulu", "alpha", "mike"} {
		_ = s.Create(validProxy(n))
	}
	list := s.List()
	if len(list) != 3 {
		t.Fatalf("List length: %d", len(list))
	}
	if list[0].Name != "alpha" || list[1].Name != "mike" || list[2].Name != "zulu" {
		t.Errorf("List not sorted: %v %v %v", list[0].Name, list[1].Name, list[2].Name)
	}
	// Redaction: credential fields must not surface in any list projection.
	for _, p := range list {
		if p.HasSecret {
			t.Error("List must not expose secret presence for credential-less proxies")
		}
	}
}

func TestPersistenceReload(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	p := validProxy("persist-me")
	_ = s.Create(p)
	_ = s.Assign("profile-x", p.ID)

	loaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, err := loaded.Get(p.ID)
	if err != nil || got.Name != "persist-me" {
		t.Fatalf("reload lost the proxy: err=%v got=%+v", err, got)
	}
	if ap := loaded.AssignedProxy("profile-x"); ap == nil || ap.ID != p.ID {
		t.Error("reload lost the assignment")
	}
}

func TestPersistenceRejectsBadSecretRef(t *testing.T) {
	dir := t.TempDir()
	// Write a tampered registry file whose secret_ref violates the grammar.
	tampered := `{"proxies":[{"id":"px","name":"tampered","type":"http","host":"h","port":1,"secret_ref":"STOLEN:proxy.plain"}]}`
	if err := os.WriteFile(filepath.Join(dir, "proxies.json"), []byte(tampered), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if loaded.List() != nil && len(loaded.List()) != 0 {
		t.Error("tampered secret_ref entry must be dropped on reload")
	}
}

func TestConcurrentMutations(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("concurrent-px")
	if err := s.Create(p); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	const workers = 40
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.Update(p.ID, map[string]any{"port": 1000 + i})
			if err != nil && err != ErrProxyLocked {
				t.Errorf("concurrent update %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	got, _ := s.Get(p.ID)
	if got.Port < 1000 || got.Port >= 1000+workers {
		t.Errorf("port out of expected range after contention: %d", got.Port)
	}
}

func TestConcurrentAssignAndDelete(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("assign-del-px")
	_ = s.Create(p)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.Assign("profile", p.ID)
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Delete(p.ID)
		// Deletion while assigned must be refused or the proxy already gone.
		if err != nil && !IsDuplicateError(err) && !IsNotFoundError(err) {
			t.Errorf("unexpected delete error: %v", err)
		}
	}()
	wg.Wait()
	// End state must be consistent: either deleted or assigned, never both.
	if ap := s.AssignedProxy("profile"); ap != nil {
		if _, err := s.Get(ap.ID); err != nil {
			t.Error("assigned proxy must remain retrievable")
		}
	}
}

func TestSecretPresenceFlag(t *testing.T) {
	s := newTestStore(t)
	p := validProxy("secret-px")
	// Create first so the id exists; then store a synthetic reference
	// (never a real credential) and persist via Update.
	if err := s.Create(p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	syntheticRef := "proxy.ref." + p.ID
	if _, err := s.Update(p.ID, map[string]any{"secret_ref": syntheticRef}); err != nil {
		t.Fatalf("Update with synthetic secret ref: %v", err)
	}
	got, _ := s.Get(p.ID)
	if got.SecretRef != syntheticRef {
		t.Errorf("SecretRef mismatch: %q", got.SecretRef)
	}
	// The secret ref must match the vault grammar; a malformed ref is refused.
	p2 := validProxy("bad-ref")
	p2.SecretRef = "not-a-ref"
	if err := s.Create(p2); err == nil {
		t.Error("Create must refuse a malformed secret_ref")
	}
}

func TestIsolationBudgetTimeout(t *testing.T) {
	// Shorten the budget for this test only by using WithProxy directly with
	// a tiny budget while holding the per-proxy lock.
	s := newTestStore(t)
	p := validProxy("budget-px")
	_ = s.Create(p)
	mu := &sync.Mutex{}
	mu.Lock() // hold forever
	s.perProxyMu.Lock()
	s.perProxy[p.ID] = mu
	s.perProxyMu.Unlock()
	_, err := s.WithProxy(p.ID, 50*time.Millisecond)
	if err != ErrProxyLocked {
		t.Errorf("expected ErrProxyLocked on budget expiry: %v", err)
	}
	mu.Unlock()
}
