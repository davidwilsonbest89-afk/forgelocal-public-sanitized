package history

import (
	"context"
	"sync"
	"testing"

	"forgelocal/internal/profile"
)

func testProfile(name string) *profile.Profile {
	return &profile.Profile{ID: "prof_history", Name: name, RuntimeID: "chromium", LifecycleState: profile.LifecycleActive, Tags: []string{"alpha"}}
}

func TestCaptureCreatesImmutableVersionsAndRedactsSecrets(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil { t.Fatal(err) }
	defer s.Close()
	ctx := context.Background()
	p := testProfile("First")
	p.Proxy = &profile.ProxyConfig{Type: "http", Host: "example.test", Port: 8080, Username: "secret-user", Password: "secret-pass", SecretRef: "proxy.prof_history"}
	first, err := s.Capture(ctx, p, "create", "corr-create")
	if err != nil { t.Fatalf("capture v1: %v", err) }
	p.Name = "Second"
	p.Tags = []string{"beta"}
	second, err := s.Capture(ctx, p, "update", "corr-update")
	if err != nil { t.Fatalf("capture v2: %v", err) }
	if first.Version != 1 || second.Version != 2 { t.Fatalf("versions = %d, %d", first.Version, second.Version) }
	v1, err := s.Get(ctx, p.ID, 1)
	if err != nil { t.Fatal(err) }
	if v1.Snapshot.Profile.Name != "First" || v1.Snapshot.Profile.Proxy == nil { t.Fatalf("v1 was mutated: %#v", v1.Snapshot.Profile) }
	if v1.Snapshot.Profile.Proxy.Username != "" || v1.Snapshot.Profile.Proxy.Password != "" || v1.Snapshot.Profile.Proxy.SecretRef != "" { t.Fatalf("snapshot leaks proxy secret material: %#v", v1.Snapshot.Profile.Proxy) }
	list, err := s.List(ctx, p.ID, 10, 0)
	if err != nil { t.Fatal(err) }
	if list.Total != 2 || len(list.Data) != 2 || list.Data[0].Version != 2 { t.Fatalf("unexpected list: %#v", list) }
	diff, err := s.Diff(ctx, p.ID, 1, 2)
	if err != nil { t.Fatal(err) }
	if len(diff.Paths) == 0 { t.Fatal("expected changed paths") }
	restored, err := s.Restore(ctx, p.ID, 1, 2, "corr-restore", func(snapshot *profile.Profile) (*profile.Profile, error) { return snapshot, nil })
	if err != nil { t.Fatalf("restore: %v", err) }
	if restored.Version != 3 || restored.Action != "restore" { t.Fatalf("unexpected restore: %#v", restored) }
	if _, err := s.Restore(ctx, p.ID, 1, 2, "corr-conflict", func(snapshot *profile.Profile) (*profile.Profile, error) { return snapshot, nil }); err != ErrVersionConflict {
		t.Fatalf("conflict = %v", err)
	}
	var auditCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM profile_history_audit_events WHERE action='history_restore_conflict' AND correlation_id='corr-conflict'`).Scan(&auditCount); err != nil { t.Fatal(err) }
	if auditCount != 1 { t.Fatalf("conflict audit count = %d", auditCount) }
}

func TestCaptureSerializesConcurrentVersions(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil { t.Fatal(err) }
	defer s.Close()
	ctx := context.Background()
	if _, err := s.Capture(ctx, testProfile("Base"), "create", "corr-base"); err != nil { t.Fatal(err) }
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Capture(ctx, testProfile("Same"), "update", "corr-concurrent")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil { t.Fatal(err) }
	}
	list, err := s.List(ctx, "prof_history", 20, 0)
	if err != nil { t.Fatal(err) }
	if list.Total != 9 || list.Data[0].Version != 9 { t.Fatalf("concurrent versions: %#v", list) }
}
