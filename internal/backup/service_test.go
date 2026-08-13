package backup

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forgelocal/internal/secrets"
)

func newTestService(t *testing.T) (*Service, *SQLiteStore, []byte) {
	t.Helper()
	dir := t.TempDir()
	store, err := OpenSQLite(filepath.Join(dir, "metadata.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	key, err := secrets.NewKey()
	if err != nil {
		t.Fatal(err)
	}
	vault := secrets.NewMemoryVault()
	if err := vault.Put("key-test-01", key); err != nil {
		t.Fatal(err)
	}
	return &Service{Root: filepath.Join(dir, "backups"), Vault: vault, Store: store, Locks: NewProfileLocks(), Now: func() time.Time { return time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC) }}, store, key
}

func TestBackupCreateRestoreIsolatedAndNoSecretAtRest(t *testing.T) {
	svc, store, _ := newTestService(t)
	payload := []byte(`{"cookies":"super-secret-cookie-value","local_storage":"ok"}`)
	backup, err := svc.Create("profile-source", "key-test-01", payload)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(backup.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, magic) || !bytes.HasSuffix(body, trailer) {
		t.Fatalf("unexpected container framing")
	}
	if bytes.Contains(body, payload) || bytes.Contains(body, []byte("super-secret-cookie-value")) {
		t.Fatal("payload secret is present in flbackup plaintext")
	}
	if got, err := os.Stat(backup.ArtifactPath); err != nil || got.Mode().Perm() != 0600 {
		t.Fatalf("artifact mode=%v err=%v", got.Mode(), err)
	}
	var storedKeyID string
	if err := store.DB().QueryRow(`SELECT key_id FROM backups WHERE id=?`, backup.ID).Scan(&storedKeyID); err != nil {
		t.Fatal(err)
	}
	if storedKeyID != "key-test-01" {
		t.Fatalf("key_id=%q", storedKeyID)
	}
	var dbBytes []byte
	if err := store.DB().QueryRow(`SELECT CAST(group_concat(sql, '') AS BLOB) FROM sqlite_master`).Scan(&dbBytes); err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if bytes.Contains(dbBytes, []byte("super-secret-cookie-value")) {
		t.Fatal("secret leaked to sqlite schema")
	}
	target := filepath.Join(t.TempDir(), "browser-data-target")
	restored, err := svc.Restore(backup.ID, "profile-restored", target)
	if err != nil {
		t.Fatal(err)
	}
	if restored.TargetProfileID == backup.ProfileID || restored.TargetPath == backup.ArtifactPath {
		t.Fatal("restore reused source identity or artifact path")
	}
	got, err := os.ReadFile(filepath.Join(target, "payload.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("restored payload mismatch")
	}
	if _, err := svc.Restore(backup.ID, "profile-source", filepath.Join(t.TempDir(), "forbidden")); err == nil {
		t.Fatal("expected same profile id restore rejection")
	}
}

func TestBackupRejectsCorruptionTruncationAndAADTampering(t *testing.T) {
	svc, _, key := newTestService(t)
	backup, err := svc.Create("profile-a", "key-test-01", []byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(backup.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutated := range map[string][]byte{
		"magic":     append([]byte("X"), body[1:]...),
		"truncated": body[:len(body)-3],
		"aad":       bytes.Replace(append([]byte(nil), body...), []byte("profile-a"), []byte("profile-b"), 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := decode(mutated, key); err == nil {
				t.Fatalf("%s accepted", name)
			}
		})
	}
}

func TestBackupRejectsHostileIdentifiersAndRestorePaths(t *testing.T) {
	svc, _, _ := newTestService(t)
	if _, err := svc.Create("../escape", "key-test-01", []byte("x")); err == nil {
		t.Fatal("expected hostile profile id rejection")
	}
	backup, err := svc.Create("profile-safe", "key-test-01", []byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Restore(backup.ID, "profile-new", filepath.Join(svc.Root, "escape")); err == nil {
		t.Fatal("expected target inside backup root rejection")
	}
}

func TestSnapshotLockRejectsConcurrentBackup(t *testing.T) {
	locks := NewProfileLocks()
	release, err := locks.Acquire("profile-a")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := locks.Acquire("profile-a"); !errors.Is(err, ErrBusy) {
		t.Fatalf("err=%v", err)
	}
}

func TestReconcileRepairsPublishedUnregisteredAfterCrash(t *testing.T) {
	svc, store, _ := newTestService(t)
	var published Backup
	svc.AfterPublishHook = func(b Backup) error { published = b; return errors.New("injected crash after publish") }
	if _, err := svc.Create("profile-crash", "key-test-01", []byte("safe payload")); err == nil {
		t.Fatal("expected injected failure")
	}
	if published.ID == "" {
		t.Fatal("publish hook did not run")
	}
	if ok, err := store.HasBackup(published.ID); err != nil || ok {
		t.Fatalf("backup exists before recovery=%v err=%v", ok, err)
	}
	svc.AfterPublishHook = nil
	recovered, quarantined, err := svc.Reconcile()
	if err != nil {
		t.Fatal(err)
	}
	if recovered != 1 || quarantined != 0 {
		t.Fatalf("recovered=%d quarantined=%d", recovered, quarantined)
	}
	if ok, err := store.HasBackup(published.ID); err != nil || !ok {
		t.Fatalf("backup not recovered ok=%v err=%v", ok, err)
	}
}

func TestReconcileQuarantinesMalformedArtifact(t *testing.T) {
	svc, _, _ := newTestService(t)
	if err := os.MkdirAll(svc.Root, 0700); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(svc.Root, "evil.flbackup")
	if err := os.WriteFile(bad, []byte("not-a-container"), 0600); err != nil {
		t.Fatal(err)
	}
	recovered, quarantined, err := svc.Reconcile()
	if err != nil {
		t.Fatal(err)
	}
	if recovered != 0 || quarantined != 1 {
		t.Fatalf("recovered=%d quarantined=%d", recovered, quarantined)
	}
	if _, err := os.Stat(filepath.Join(svc.Root, "quarantine", "evil.flbackup.invalid")); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationsEnableCoreTables(t *testing.T) {
	_, store, _ := newTestService(t)
	for _, table := range []string{"schema_migrations", "backup_operations", "backups", "restore_operations", "audit_events"} {
		var name string
		if err := store.DB().QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Fatalf("missing table %s: %v", table, err)
		}
	}
	var foreignKeys int
	if err := store.DB().QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys=%d", foreignKeys)
	}
}

func TestBackupIDsDeterministic(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b.flbackup", "a.flbackup", "ignore.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ids, err := BackupIDs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Fatalf("ids=%v", ids)
	}
}
