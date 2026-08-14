package profilemigration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forgelocal/internal/backup"
	"forgelocal/internal/secrets"
)

var fixtureTime = time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

func TestDryRunJournalsPlanWithoutWritingProductRecords(t *testing.T) {
	db := openMigrationDB(t)
	source := writeFixture(t, fixture{withGroup: true})
	migrator, err := New(db, source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}

	report, err := migrator.Run(context.Background(), Options{Mode: ModeDryRun, CorrelationID: "corr-dry-run", Now: func() time.Time { return fixtureTime }})
	if err != nil {
		t.Fatalf("dry-run migration: %v", err)
	}
	if report.State != "validated" || report.Profiles != 1 || report.Groups != 1 || report.Tags != 2 || report.Runtimes != 1 {
		t.Fatalf("unexpected dry-run report: %#v", report)
	}
	assertCount(t, db, "profiles", 0)
	assertCount(t, db, "groups", 0)
	assertCount(t, db, "runtime_candidates", 0)
	assertCount(t, db, "profile_json_parity_checks", 0)
	assertCount(t, db, "profile_import_operations", 2)
	var dryRun, validated int
	if err := db.QueryRow(`SELECT dry_run, COUNT(*) FROM profile_import_operations GROUP BY dry_run`).Scan(&dryRun, &validated); err != nil {
		t.Fatalf("read dry-run journal: %v", err)
	}
	if dryRun != 1 || validated != 2 {
		t.Fatalf("dry-run journal = dry_run:%d count:%d", dryRun, validated)
	}
}

func TestApplyRequiresVerifiedBackupAndPersistsParity(t *testing.T) {
	db := openMigrationDB(t)
	source := writeFixture(t, fixture{withGroup: true})
	migrator, err := New(db, source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}

	if _, err := migrator.Run(context.Background(), Options{Mode: ModeApply, Now: func() time.Time { return fixtureTime }}); !errors.Is(err, ErrBackupRequired) {
		t.Fatalf("apply without backup error = %v, want ErrBackupRequired", err)
	}
	assertCount(t, db, "profiles", 0)
	assertCount(t, db, "profile_import_operations", 0)

	backupCalls := 0
	backupFn := func(_ context.Context, request BackupRequest) ([]BackupReceipt, error) {
		backupCalls++
		if request.ProfilesDir != source.ProfilesDir || request.GroupsPath != source.GroupsPath || len(request.SourceHashes) != 2 {
			t.Fatalf("unexpected backup request: %#v", request)
		}
		return []BackupReceipt{{ID: "bkp_migration", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}, nil
	}
	report, err := migrator.Run(context.Background(), Options{Mode: ModeApply, CorrelationID: "corr-apply", Backup: backupFn, Now: func() time.Time { return fixtureTime }})
	if err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	if report.State != "committed" || len(report.Backups) != 1 || len(report.Parity) != 1 || report.Parity[0].Result != "match" {
		t.Fatalf("unexpected apply report: %#v", report)
	}
	if backupCalls != 1 {
		t.Fatalf("backup calls = %d, want 1", backupCalls)
	}
	assertCount(t, db, "profiles", 1)
	assertCount(t, db, "groups", 1)
	assertCount(t, db, "profile_tags", 2)
	assertCount(t, db, "profile_tag_assignments", 2)
	assertCount(t, db, "profile_json_parity_checks", 1)
	assertCount(t, db, "profile_import_operations", 2)
	assertCount(t, db, "product_audit_events", 1)

	var runtimeName, binaryPath, binarySHA string
	if err := db.QueryRow(`SELECT name, binary_path, binary_sha256 FROM runtime_candidates WHERE id = 'camoufox'`).Scan(&runtimeName, &binaryPath, &binarySHA); err != nil {
		t.Fatalf("read migrated runtime: %v", err)
	}
	if runtimeName != "camoufox" || binaryPath != "" || binarySHA != "" {
		t.Fatalf("legacy runtime must remain unverified: name=%q path=%q sha=%q", runtimeName, binaryPath, binarySHA)
	}

	if _, err := migrator.Run(context.Background(), Options{Mode: ModeApply, Backup: backupFn, Now: func() time.Time { return fixtureTime }}); !errors.Is(err, ErrExistingProductRecord) {
		t.Fatalf("repeat apply error = %v, want ErrExistingProductRecord", err)
	}
	if backupCalls != 1 {
		t.Fatalf("backup must not run when product records already exist; calls=%d", backupCalls)
	}
}

func TestApplyCreatesVerifiedEncryptedPreMigrationArtifact(t *testing.T) {
	store := openMigrationStore(t)
	source := writeFixture(t, fixture{withGroup: true})
	migrator, err := New(store.DB(), source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	vault := secrets.NewMemoryVault()
	if err := vault.Put("key-migration-01", bytes.Repeat([]byte{0x42}, 32)); err != nil {
		t.Fatalf("seed migration vault key: %v", err)
	}
	backupRoot := filepath.Join(t.TempDir(), "backups")
	if err := os.MkdirAll(backupRoot, 0700); err != nil {
		t.Fatalf("create backup root: %v", err)
	}
	service := &backup.Service{Root: backupRoot, Vault: vault, Store: store, Locks: backup.NewProfileLocks(), Now: func() time.Time { return fixtureTime }}

	report, err := migrator.Run(context.Background(), Options{
		Mode: ModeApply, CorrelationID: "corr-encrypted-preimage", Backup: NewEncryptedPreMigrationBackup(service, "key-migration-01"),
		Now: func() time.Time { return fixtureTime },
	})
	if err != nil {
		t.Fatalf("apply with encrypted pre-migration backup: %v", err)
	}
	if len(report.Backups) != 1 {
		t.Fatalf("backup receipts = %#v", report.Backups)
	}
	if err := service.Verify(report.Backups[0].ID); err != nil {
		t.Fatalf("verify migration preimage artifact: %v", err)
	}
	stored, err := store.GetBackup(report.Backups[0].ID)
	if err != nil {
		t.Fatalf("load migration preimage artifact: %v", err)
	}
	body, err := os.ReadFile(stored.ArtifactPath)
	if err != nil {
		t.Fatalf("read migration preimage artifact: %v", err)
	}
	for _, plaintext := range [][]byte{[]byte("Profile prof_alpha"), []byte("Europe/Paris"), []byte("France")} {
		if bytes.Contains(body, plaintext) {
			t.Fatalf("encrypted artifact exposes source metadata %q", plaintext)
		}
	}
	assertCount(t, store.DB(), "profiles", 1)
	assertCount(t, store.DB(), "backups", 1)
}

func TestApplyCanRetryAfterContextInterruptedAfterPreimage(t *testing.T) {
	store := openMigrationStore(t)
	source := writeFixture(t, fixture{withGroup: true})
	migrator, err := New(store.DB(), source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	vault := secrets.NewMemoryVault()
	if err := vault.Put("key-interrupt-01", bytes.Repeat([]byte{0x24}, 32)); err != nil {
		t.Fatalf("seed interruption vault key: %v", err)
	}
	backupRoot := filepath.Join(t.TempDir(), "backups")
	if err := os.MkdirAll(backupRoot, 0700); err != nil {
		t.Fatalf("create backup root: %v", err)
	}
	service := &backup.Service{Root: backupRoot, Vault: vault, Store: store, Locks: backup.NewProfileLocks(), Now: func() time.Time { return fixtureTime }}
	preimage := NewEncryptedPreMigrationBackup(service, "key-interrupt-01")
	ctx, cancel := context.WithCancel(context.Background())
	_, err = migrator.Run(ctx, Options{
		Mode: ModeApply, CorrelationID: "corr-interrupted-after-preimage",
		Backup: func(ctx context.Context, request BackupRequest) ([]BackupReceipt, error) {
			receipts, backupErr := preimage(ctx, request)
			cancel()
			return receipts, backupErr
		},
		Now: func() time.Time { return fixtureTime },
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted apply error = %v, want context.Canceled", err)
	}
	assertCount(t, store.DB(), "backups", 1)
	assertCount(t, store.DB(), "profiles", 0)
	assertCount(t, store.DB(), "profile_import_operations", 0)

	report, err := migrator.Run(context.Background(), Options{
		Mode: ModeApply, CorrelationID: "corr-retry-after-interruption", Backup: preimage,
		Now: func() time.Time { return fixtureTime },
	})
	if err != nil {
		t.Fatalf("retry after interruption: %v", err)
	}
	if len(report.Backups) != 1 {
		t.Fatalf("retry backup receipts = %#v", report.Backups)
	}
	assertCount(t, store.DB(), "backups", 2)
	assertCount(t, store.DB(), "profiles", 1)
	assertCount(t, store.DB(), "profile_json_parity_checks", 1)
}

func TestApplyRollsBackAndCanResumeAfterSourceCorrection(t *testing.T) {
	db := openMigrationDB(t)
	source := writeFixture(t, fixture{secondProfile: true, duplicateContainer: true})
	migrator, err := New(db, source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	backupFn := func(_ context.Context, _ BackupRequest) ([]BackupReceipt, error) {
		return []BackupReceipt{{ID: "bkp_rollback", SHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}, nil
	}
	if _, err := migrator.Run(context.Background(), Options{Mode: ModeApply, Backup: backupFn, Now: func() time.Time { return fixtureTime }}); err == nil {
		t.Fatal("apply with duplicate container id must fail")
	}
	assertCount(t, db, "profiles", 0)
	assertCount(t, db, "groups", 0)
	assertCount(t, db, "runtime_candidates", 0)
	assertCount(t, db, "profile_import_operations", 0)
	assertCount(t, db, "product_audit_events", 0)

	source = writeFixture(t, fixture{secondProfile: true, duplicateContainer: false})
	migrator, err = New(db, source)
	if err != nil {
		t.Fatalf("new migrator after correction: %v", err)
	}
	report, err := migrator.Run(context.Background(), Options{Mode: ModeApply, Backup: backupFn, Now: func() time.Time { return fixtureTime }})
	if err != nil {
		t.Fatalf("resume after source correction: %v", err)
	}
	if report.Profiles != 2 || len(report.Parity) != 2 {
		t.Fatalf("resume report = %#v", report)
	}
	assertCount(t, db, "profiles", 2)
	assertCount(t, db, "profile_json_parity_checks", 2)
}

func TestRejectsCleartextProxyCredentialBeforeJournalOrBackup(t *testing.T) {
	db := openMigrationDB(t)
	source := writeFixture(t, fixture{cleartextProxyPassword: true})
	migrator, err := New(db, source)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	backupCalls := 0
	_, err = migrator.Run(context.Background(), Options{
		Mode: ModeApply,
		Backup: func(_ context.Context, _ BackupRequest) ([]BackupReceipt, error) {
			backupCalls++
			return nil, nil
		},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("cleartext proxy error = %v, want ErrValidation", err)
	}
	if backupCalls != 0 {
		t.Fatalf("backup calls = %d, want 0", backupCalls)
	}
	assertCount(t, db, "profiles", 0)
	assertCount(t, db, "profile_import_operations", 0)
}

type fixture struct {
	withGroup              bool
	secondProfile          bool
	duplicateContainer     bool
	cleartextProxyPassword bool
}

func writeFixture(t *testing.T, spec fixture) Source {
	t.Helper()
	root := t.TempDir()
	profilesDir := filepath.Join(root, "profiles")
	groupsPath := filepath.Join(root, "data", "groups.json")
	if err := os.MkdirAll(filepath.Dir(groupsPath), 0700); err != nil {
		t.Fatalf("create data directory: %v", err)
	}
	if err := os.MkdirAll(profilesDir, 0700); err != nil {
		t.Fatalf("create profiles directory: %v", err)
	}
	if spec.withGroup {
		groups := map[string]any{"groups": []any{map[string]any{
			"name": "France", "proxy_mode": "enforced", "proxy": map[string]any{"type": "socks5", "host": "127.0.0.1", "port": 1080, "region": "FR"},
			"created_at": fixtureTime.Format(time.RFC3339Nano), "updated_at": fixtureTime.Format(time.RFC3339Nano),
		}}}
		writeJSON(t, groupsPath, groups)
	} else {
		writeJSON(t, groupsPath, map[string]any{"groups": []any{}})
	}
	writeProfileFixture(t, profilesDir, "prof_alpha", "container-alpha", spec.withGroup, spec.cleartextProxyPassword)
	if spec.secondProfile {
		container := "container-beta"
		if spec.duplicateContainer {
			container = "container-alpha"
		}
		writeProfileFixture(t, profilesDir, "prof_beta", container, false, false)
	}
	return Source{ProfilesDir: profilesDir, GroupsPath: groupsPath}
}

func writeProfileFixture(t *testing.T, profilesDir, id, containerID string, group, cleartextProxyPassword bool) {
	t.Helper()
	profileDir := filepath.Join(profilesDir, id)
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		t.Fatalf("create profile directory: %v", err)
	}
	payload := map[string]any{
		"id": id, "name": "Profile " + id, "runtime_id": "camoufox", "tags": []string{"Retail", "France"},
		"created_at": fixtureTime.Format(time.RFC3339Nano), "last_used": fixtureTime.Add(time.Hour).Format(time.RFC3339Nano),
		"fingerprint": map[string]any{"timezone": "Europe/Paris"}, "fingerprint_seed": 7,
		"container_id": containerID, "profile_dir": profileDir,
	}
	if group {
		payload["group"] = "France"
	}
	if cleartextProxyPassword {
		payload["proxy"] = map[string]any{"type": "http", "host": "127.0.0.1", "port": 8080, "password": "never-store-this"}
	}
	writeJSON(t, filepath.Join(profileDir, "profile.json"), payload)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode fixture JSON: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write fixture JSON: %v", err)
	}
}

func openMigrationDB(t *testing.T) *sql.DB {
	t.Helper()
	return openMigrationStore(t).DB()
}

func openMigrationStore(t *testing.T) *backup.SQLiteStore {
	t.Helper()
	store, err := backup.OpenSQLite(filepath.Join(t.TempDir(), "forgelocal.sqlite"))
	if err != nil {
		t.Fatalf("open Core SQLite: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func assertCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != want {
		t.Fatalf("count %s = %d, want %d", table, count, want)
	}
}
