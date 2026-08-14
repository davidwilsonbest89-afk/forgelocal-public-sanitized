package backup

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateUpgradesExistingBack01DatabaseToProductV4(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "forgelocal.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(backupSchemaSQL); err != nil {
		t.Fatalf("seed BACK-01 schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (1, '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("seed BACK-01 ledger: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("upgrade BACK-01 database: %v", err)
	}
	for _, table := range []string{"schema_migrations", "backups", "runtime_candidates", "groups", "profiles", "proxy_providers"} {
		var found string
		if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&found); err != nil || found != table {
			t.Fatalf("expected table %q after upgrade, got %q / %v", table, found, err)
		}
	}
	assertMigrationRecordedOnce(t, db, productSchemaVersion)
	assertMigrationRecordedOnce(t, db, proxyIndexesSchemaVersion)
	assertMigrationRecordedOnce(t, db, operationJournalSchemaVersion)
	if _, err := db.Exec(`INSERT INTO profile_import_operations
		(id, source_kind, source_sha256, dry_run, state, summary_json, correlation_id, created_at, updated_at)
		VALUES ('operation.started', 'profile_json', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 1, 'started', '{"Sources":1}', 'corr-v4', '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("v4 must accept started durable operation: %v", err)
	}
	for _, index := range []string{
		"idx_profiles_proxy_provider_id_not_null",
		"idx_profiles_proxy_secret_ref_not_empty",
		"idx_groups_proxy_provider_id_not_null",
		"idx_groups_proxy_secret_ref_not_empty",
	} {
		var found string
		if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?`, index).Scan(&found); err != nil || found != index {
			t.Fatalf("expected proxy index %q after upgrade, got %q / %v", index, found, err)
		}
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("repeat upgrade must be idempotent: %v", err)
	}
	assertMigrationRecordedOnce(t, db, productSchemaVersion)
	assertMigrationRecordedOnce(t, db, proxyIndexesSchemaVersion)
	assertMigrationRecordedOnce(t, db, operationJournalSchemaVersion)
}

func assertMigrationRecordedOnce(t *testing.T, db *sql.DB, version int) {
	t.Helper()
	var records int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&records); err != nil {
		t.Fatalf("count migration %d records: %v", version, err)
	}
	if records != 1 {
		t.Fatalf("migration %d ledger records = %d, want 1", version, records)
	}
}
