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
		t.Fatalf("apply Core migrations through version 3: %v", err)
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

func TestProxyReferenceColumnsConstraintsAndIndexesAreCanonical(t *testing.T) {
	db := openProductSchema(t)

	expectedColumns := map[string]map[string]string{
		"profiles": {
			"proxy_provider_id": "TEXT",
			"proxy_secret_ref":  "TEXT",
		},
		"groups": {
			"proxy_provider_id": "TEXT",
			"proxy_secret_ref":  "TEXT",
		},
	}
	for table, expected := range expectedColumns {
		rows, err := db.Query("PRAGMA table_info(" + table + ")")
		if err != nil {
			t.Fatalf("inspect %s columns: %v", table, err)
		}
		found := map[string]string{}
		for rows.Next() {
			var cid, notNull, pk int
			var name, typ string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
				_ = rows.Close()
				t.Fatalf("scan %s columns: %v", table, err)
			}
			found[name] = typ
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close %s columns: %v", table, err)
		}
		for column, expectedType := range expected {
			if found[column] != expectedType {
				t.Fatalf("%s.%s type = %q, want %q", table, column, found[column], expectedType)
			}
		}

		var createSQL string
		if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&createSQL); err != nil {
			t.Fatalf("read %s DDL: %v", table, err)
		}
		if !strings.Contains(createSQL, "proxy_provider_id TEXT REFERENCES proxy_providers(id) ON DELETE RESTRICT") {
			t.Fatalf("%s.proxy_provider_id must retain proxy provider FK: %s", table, createSQL)
		}
		if !strings.Contains(createSQL, "proxy_secret_ref GLOB") {
			t.Fatalf("%s.proxy_secret_ref must retain a bounded vault-reference CHECK: %s", table, createSQL)
		}
	}

	expectedIndexes := map[string]string{
		"idx_profiles_proxy_provider_id_not_null": "ON profiles(proxy_provider_id) WHERE proxy_provider_id IS NOT NULL",
		"idx_profiles_proxy_secret_ref_not_empty": "ON profiles(proxy_secret_ref) WHERE proxy_secret_ref <> ''",
		"idx_groups_proxy_provider_id_not_null":   "ON groups(proxy_provider_id) WHERE proxy_provider_id IS NOT NULL",
		"idx_groups_proxy_secret_ref_not_empty":   "ON groups(proxy_secret_ref) WHERE proxy_secret_ref <> ''",
	}
	for indexName, expectedSQL := range expectedIndexes {
		var sqlText string
		if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'index' AND name = ?`, indexName).Scan(&sqlText); err != nil {
			t.Fatalf("read index %s: %v", indexName, err)
		}
		if !strings.Contains(sqlText, expectedSQL) {
			t.Fatalf("index %s DDL = %q, want fragment %q", indexName, sqlText, expectedSQL)
		}
	}
}

func TestProxyReferenceLookupPlansUsePartialIndexes(t *testing.T) {
	db := openProductSchema(t)

	checks := []struct {
		name      string
		query     string
		argument  string
		wantIndex string
	}{
		{
			name:      "profiles by provider",
			query:     `EXPLAIN QUERY PLAN SELECT id FROM profiles WHERE proxy_provider_id = ? AND proxy_provider_id IS NOT NULL`,
			argument:  "provider.demo",
			wantIndex: "idx_profiles_proxy_provider_id_not_null",
		},
		{
			name:      "profiles by vault reference",
			query:     `EXPLAIN QUERY PLAN SELECT id FROM profiles WHERE proxy_secret_ref = ? AND proxy_secret_ref <> ''`,
			argument:  "proxy.profile-demo",
			wantIndex: "idx_profiles_proxy_secret_ref_not_empty",
		},
		{
			name:      "groups by provider",
			query:     `EXPLAIN QUERY PLAN SELECT id FROM groups WHERE proxy_provider_id = ? AND proxy_provider_id IS NOT NULL`,
			argument:  "provider.demo",
			wantIndex: "idx_groups_proxy_provider_id_not_null",
		},
		{
			name:      "groups by vault reference",
			query:     `EXPLAIN QUERY PLAN SELECT id FROM groups WHERE proxy_secret_ref = ? AND proxy_secret_ref <> ''`,
			argument:  "proxy.group-demo",
			wantIndex: "idx_groups_proxy_secret_ref_not_empty",
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			rows, err := db.Query(check.query, check.argument)
			if err != nil {
				t.Fatalf("explain query plan: %v", err)
			}
			defer func() { _ = rows.Close() }()

			usedExpectedIndex := false
			for rows.Next() {
				var id, parent, unused int
				var detail string
				if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
					t.Fatalf("scan query plan: %v", err)
				}
				if strings.Contains(detail, check.wantIndex) {
					usedExpectedIndex = true
				}
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("iterate query plan: %v", err)
			}
			if !usedExpectedIndex {
				t.Fatalf("query plan did not use %s", check.wantIndex)
			}
		})
	}
}

func TestProfileTagsEnforcesCaseInsensitiveUniqueness(t *testing.T) {
	db := openProductSchema(t)
	if _, err := db.Exec(`INSERT INTO profile_tags(id, name, created_at) VALUES ('tag.initial', 'QA', '2026-08-14T00:00:00Z')`); err != nil {
		t.Fatalf("insert QA: %v", err)
	}
	for index, variant := range []string{"qa", "Qa", "qA", "QA"} {
		id := "tag.variant." + string(rune('a'+index))
		if _, err := db.Exec(`INSERT INTO profile_tags(id, name, created_at) VALUES (?, ?, '2026-08-14T00:00:00Z')`, id, variant); err == nil {
			t.Fatalf("duplicate tag %q must be rejected", variant)
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_tags WHERE name = 'qa'`).Scan(&count); err != nil {
		t.Fatalf("count case-insensitive QA tags: %v", err)
	}
	if count != 1 {
		t.Fatalf("case-insensitive QA tags = %d, want 1", count)
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
