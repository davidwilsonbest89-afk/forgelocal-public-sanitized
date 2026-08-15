package backup

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteReadOnlyCatalogProjectsGroupsAndRuntimeCandidatesWithoutSentinels(t *testing.T) {
	store, err := OpenSQLite(filepath.Join(t.TempDir(), "t06.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := "2026-08-15T00:00:00Z"
	if _, err := store.db.Exec(`INSERT INTO runtime_candidates(id,name,version,architecture,binary_path,binary_sha256,status,created_at) VALUES(?,?,?,?,?,?,?,?)`, "runtime-t06", "T06 Runtime", "1.0", "amd64", "/t06/private/runtime", "T06-RUNTIME-HASH", "candidate", now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO groups(id,name,proxy_mode,proxy_secret_ref,proxy_type,proxy_host,proxy_port,proxy_region,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, "group-t06", "T06 Group", "enforced", "proxy.group.t06-sentinel", "socks5", "t06.proxy.invalid", 1080, "private", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO profiles(id,name,runtime_id,group_id,profile_dir,identity_json,fingerprint_json,metadata_json,proxy_secret_ref,proxy_type,proxy_host,proxy_port,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, "profile-t06", "T06 Profile", "runtime-t06", "group-t06", "/t06/private/profile", "{}", "{}", "{}", "proxy.t06-sentinel", "socks5", "t06.profile.proxy.invalid", 1080, now, now); err != nil {
		t.Fatal(err)
	}
	groups, err := store.ListReadOnlyGroups(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	runtimes, err := store.ListReadOnlyRuntimeCandidates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].ID != "group-t06" || groups[0].ProfileCount != 1 || !groups[0].ProxyConfigured {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if len(runtimes) != 1 || runtimes[0].ID != "runtime-t06" || runtimes[0].Status != "candidate" {
		t.Fatalf("unexpected runtimes: %+v", runtimes)
	}
	serialized := strings.Join([]string{groups[0].ID, groups[0].Name, runtimes[0].ID, runtimes[0].Name, runtimes[0].Version}, "|")
	for _, forbidden := range []string{"t06-sentinel", "/t06/private", "T06-RUNTIME-HASH", "t06.proxy.invalid"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("projection leaked %q", forbidden)
		}
	}
}
