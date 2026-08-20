package history

import (
	"context"
	"path/filepath"
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

func TestReconcilePendingRecoversOnceThenConfirms(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil { t.Fatal(err) }
	defer s.Close()
	ctx := context.Background()
	p := testProfile("Before")
	if _, err := s.Capture(ctx, p, "create", "corr-base"); err != nil { t.Fatal(err) }
	p.Name = "After interrupted profile write"
	digest, err := profile.HistorySnapshotDigest(p)
	if err != nil { t.Fatal(err) }
	p.HistoryPending = &profile.HistoryPendingOperation{OperationID: "pending-recovery", Action: "update", SnapshotDigest: digest}
	first, err := s.ReconcilePending(ctx, p, "corr-recovery")
	if err != nil { t.Fatal(err) }
	if first.Action != "recovery" || first.Version != 2 { t.Fatalf("first recovery = %#v", first) }
	second, err := s.ReconcilePending(ctx, p, "corr-confirm")
	if err != nil { t.Fatal(err) }
	if second.Action != "confirmed" || second.Version != 2 { t.Fatalf("second recovery = %#v", second) }
	list, err := s.List(ctx, p.ID, 10, 0)
	if err != nil { t.Fatal(err) }
	if list.Total != 2 { t.Fatalf("reconciliation duplicated history: %#v", list) }
}

func TestPendingOperationConditionalClearPreservesNewerMutationAcrossRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	profiles, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil { t.Fatal(err) }
	historyStore, err := Open(root)
	if err != nil { t.Fatal(err) }
	defer historyStore.Close()
	p := &profile.Profile{Name: "Base", RuntimeID: "chromium", LifecycleState: profile.LifecycleActive}
	if err := profiles.Create(p); err != nil { t.Fatal(err) }
	if _, err := historyStore.Capture(ctx, p, "create", "corr-create"); err != nil { t.Fatal(err) }
	if err := profiles.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	a, err := profiles.Update(p.ID, map[string]any{"name": "A captured"})
	if err != nil { t.Fatal(err) }
	aOperation := a.HistoryPending.OperationID
	if _, err := historyStore.Capture(ctx, a, "update", "corr-a"); err != nil { t.Fatal(err) }
	b, err := profiles.Update(p.ID, map[string]any{"name": "B pending"})
	if err != nil { t.Fatal(err) }
	bOperation := b.HistoryPending.OperationID
	if err := profiles.ClearHistoryPending(p.ID, aOperation); err != nil { t.Fatal(err) }
	current, err := profiles.Get(p.ID)
	if err != nil || current.HistoryPending == nil || current.HistoryPending.OperationID != bOperation { t.Fatalf("older clear erased B marker: %#v %v", current, err) }
	reopened, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil { t.Fatal(err) }
	pending, err := reopened.Get(p.ID)
	if err != nil || pending.HistoryPending == nil || pending.HistoryPending.OperationID != bOperation { t.Fatalf("B marker did not survive crash: %#v %v", pending, err) }
	entry, err := historyStore.ReconcilePending(ctx, pending, "corr-restart")
	if err != nil || entry.Action != "recovery" { t.Fatalf("B recovery: %#v %v", entry, err) }
	if err := reopened.ClearHistoryPending(p.ID, bOperation); err != nil { t.Fatal(err) }
	confirmed, err := reopened.Get(p.ID)
	if err != nil || confirmed.HistoryPending != nil || confirmed.Name != "B pending" { t.Fatalf("B recovery lost state: %#v %v", confirmed, err) }
}

func TestPendingOperationConditionalClearPreservesMutationAfterRestore(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	profiles, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil { t.Fatal(err) }
	historyStore, err := Open(root)
	if err != nil { t.Fatal(err) }
	defer historyStore.Close()
	p := &profile.Profile{Name: "Version one", RuntimeID: "chromium", LifecycleState: profile.LifecycleActive}
	if err := profiles.Create(p); err != nil { t.Fatal(err) }
	if _, err := historyStore.Capture(ctx, p, "create", "corr-create"); err != nil { t.Fatal(err) }
	if err := profiles.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	v2, err := profiles.Update(p.ID, map[string]any{"name": "Version two"})
	if err != nil { t.Fatal(err) }
	if _, err := historyStore.Capture(ctx, v2, "update", "corr-v2"); err != nil { t.Fatal(err) }
	if err := profiles.ClearHistoryPending(p.ID, v2.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	_, err = historyStore.Restore(ctx, p.ID, 1, 2, "corr-restore", func(snapshot *profile.Profile) (*profile.Profile, error) { return profiles.RestoreHistory(p.ID, snapshot) })
	if err != nil { t.Fatal(err) }
	restored, err := profiles.Get(p.ID)
	if err != nil || restored.HistoryPending == nil { t.Fatalf("restore marker: %#v %v", restored, err) }
	restoreOperation := restored.HistoryPending.OperationID
	b, err := profiles.Update(p.ID, map[string]any{"name": "Mutation after restore"})
	if err != nil { t.Fatal(err) }
	bOperation := b.HistoryPending.OperationID
	if err := profiles.ClearHistoryPending(p.ID, restoreOperation); err != nil { t.Fatal(err) }
	current, err := profiles.Get(p.ID)
	if err != nil || current.HistoryPending == nil || current.HistoryPending.OperationID != bOperation { t.Fatalf("restore clear erased B marker: %#v %v", current, err) }
	entry, err := historyStore.ReconcilePending(ctx, current, "corr-recover-after-restore")
	if err != nil || entry.Action != "recovery" { t.Fatalf("mutation after restore recovery: %#v %v", entry, err) }
	if err := profiles.ClearHistoryPending(p.ID, bOperation); err != nil { t.Fatal(err) }
}
