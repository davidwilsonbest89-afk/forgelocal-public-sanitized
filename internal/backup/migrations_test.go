package backup

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateUpgradesExistingBack01DatabaseToProductV2(t *testing.T) {
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
	var versionTwoRecords int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 2`).Scan(&versionTwoRecords); err != nil {
		t.Fatalf("count migration 2 records: %v", err)
	}
	if versionTwoRecords != 1 {
		t.Fatalf("migration 2 ledger records = %d, want 1", versionTwoRecords)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("repeat upgrade must be idempotent: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = 2`).Scan(&versionTwoRecords); err != nil {
		t.Fatalf("recount migration 2 records: %v", err)
	}
	if versionTwoRecords != 1 {
		t.Fatalf("idempotent migration 2 ledger records = %d, want 1", versionTwoRecords)
	}
}
