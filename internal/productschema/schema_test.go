package productschema

import (
	"database/sql"
	"strings"
	"testing"

	"forgelocal/internal/backup"

	_ "modernc.org/sqlite"
)

func openProductSchema(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:product-schema-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := backup.Migrate(db); err != nil {
		t.Fatalf("apply Core migrations through version 2: %v", err)
	}
	return db
}

func TestProductSchemaAppliesAndProtectsIntegrity(t *testing.T) {
	db := openProductSchema(t)

	for _, requiredTable := range []string{"backups", "restore_operations", "runtime_candidates", "profiles", "groups"} {
		var found string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, requiredTable).Scan(&found)
		if err != nil || found != requiredTable {
			t.Fatalf("expected unified database table %q, got %q / %v", requiredTable, found, err)
		}
	}

	if _, err := db.Exec(`INSERT INTO proxy_providers
		(id, adapter_id, display_name, created_at, updated_at)
		VALUES ('provider.local', 'manual', 'Local manual', '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("insert proxy provider: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO runtime_candidates
		(id, name, version, architecture, binary_path, binary_sha256, status, created_at)
		VALUES ('runtime.chromium', 'Chromium', '151.0.7922.108', 'amd64', '/opt/chromium', 'test-sha256', 'candidate', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("insert runtime candidate: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO groups
		(id, name, proxy_mode, proxy_provider_id, proxy_type, proxy_host, proxy_port, created_at, updated_at)
		VALUES ('group.qa', 'QA', 'enforced', 'provider.local', 'socks5', '127.0.0.1', 1080, '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("insert enforced group: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO profiles
		(id, name, runtime_id, group_id, profile_dir, created_at, updated_at)
		VALUES ('profile.qa', 'QA Profile', 'runtime.chromium', 'group.qa', '/safe/profile.qa', '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO groups
		(id, name, proxy_mode, created_at, updated_at)
		VALUES ('group.invalid', 'Invalid enforced', 'enforced', '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err == nil {
		t.Fatal("enforced group without proxy details must be rejected")
	}
	if _, err := db.Exec(`INSERT INTO profiles
		(id, name, runtime_id, profile_dir, proxy_type, proxy_host, proxy_port, created_at, updated_at)
		VALUES ('profile.invalid', 'Invalid proxy', 'runtime.chromium', '/safe/profile.invalid', 'http', '', 8080, '2026-08-14T00:00:00Z', '2026-08-14T00:00:00Z')`); err == nil {
		t.Fatal("direct proxy assignment without host must be rejected")
	}
	if _, err := db.Exec(`DELETE FROM proxy_providers WHERE id = 'provider.local'`); err == nil {
		t.Fatal("provider referenced by a group must not be deleted")
	}
	if _, err := db.Exec(`DELETE FROM runtime_candidates WHERE id = 'runtime.chromium'`); err == nil {
		t.Fatal("runtime referenced by a profile must not be deleted")
	}
}

func TestProductSchemaDoesNotStoreSecretsInCleartextColumns(t *testing.T) {
	db := openProductSchema(t)
	tables := []string{"proxy_providers", "runtime_candidates", "profiles", "groups", "product_audit_events"}
	for _, table := range tables {
		rows, err := db.Query("PRAGMA table_info(" + table + ")")
		if err != nil {
			t.Fatalf("inspect %s: %v", table, err)
		}
		for rows.Next() {
			var cid int
			var name, typ string
			var notNull, pk int
			var defaultValue any
			if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
				_ = rows.Close()
				t.Fatalf("scan %s columns: %v", table, err)
			}
			lower := strings.ToLower(name)
			if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "apikey") || strings.Contains(lower, "api_key") || strings.Contains(lower, "username") {
				_ = rows.Close()
				t.Fatalf("forbidden cleartext secret-like column %q in %s", name, table)
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatalf("iterate %s columns: %v", table, err)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close %s columns: %v", table, err)
		}
	}
}
