package profile

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"forgelocal/internal/secrets"
)

func TestBackupSnapshotRestoresBrowserDataAndRedactsProxySecret(t *testing.T) {
	store, err := NewStore(t.TempDir(), secrets.NewMemoryVault())
	if err != nil {
		t.Fatal(err)
	}
	source := &Profile{ID: "source-profile", Name: "Source", RuntimeID: "chromium", Proxy: &ProxyConfig{Type: "http", Host: "proxy.local", Port: 8080, Username: "alice", Password: "not-in-backup"}}
	if err := store.Create(source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source.ProfileDir, "browser-data", "Cookies"), []byte("session=private"), 0600); err != nil {
		t.Fatal(err)
	}
	payload, err := store.CreateBackupSnapshot(source.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(payload, []byte("not-in-backup")) {
		t.Fatal("proxy password leaked into snapshot")
	}
	restored, err := store.RestoreBackupSnapshot("restored-profile", payload)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != "restored-profile" || restored.ProfileDir == source.ProfileDir {
		t.Fatalf("unexpected restored profile %#v", restored)
	}
	if restored.Proxy == nil || restored.Proxy.Password != "" || restored.Proxy.Username != "" {
		t.Fatalf("proxy secret was restored: %#v", restored.Proxy)
	}
	got, err := os.ReadFile(filepath.Join(restored.ProfileDir, "browser-data", "Cookies"))
	if err != nil || string(got) != "session=private" {
		t.Fatalf("browser data not restored: %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(restored.ProfileDir, "browser-data", "Cookies"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(source.ProfileDir, "browser-data", "Cookies"))
	if err != nil || string(original) != "session=private" {
		t.Fatalf("source data changed through restored profile: %q, %v", original, err)
	}
}

func TestBackupSnapshotRejectsSymlink(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &Profile{ID: "profile-symlink", Name: "Bad", RuntimeID: "chromium"}
	if err := store.Create(p); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(p.ProfileDir, "browser-data", "linked")); err != nil {
		t.Skipf("symlinks unsupported on this platform: %v", err)
	}
	if _, err := store.CreateBackupSnapshot(p.ID); err == nil {
		t.Fatal("expected symlink snapshot rejection")
	}
}

func TestRestoreSnapshotRejectsTraversalArchive(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "../escape", Mode: 0600, Size: 1, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RestoreBackupSnapshot("target", buf.Bytes()); err == nil {
		t.Fatal("expected traversal archive rejection")
	}
}
