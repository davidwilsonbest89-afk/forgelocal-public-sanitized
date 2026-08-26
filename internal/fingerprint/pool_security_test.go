package fingerprint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewPoolRejectsSymlinkedFingerprintFile(t *testing.T) {
	dir := t.TempDir()
	name := "fingerprints-chromium-windows.json"
	realPath := filepath.Join(dir, name)
	data, err := json.Marshal([]map[string]any{{"navigator": map[string]any{"userAgent": "synthetic"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(realPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	pool, err := NewPool(dir)
	if err != nil {
		t.Fatal(err)
	}
	if pool.Available("chromium", "windows") != 1 {
		t.Fatalf("nominal fingerprint file was not loaded: %d", pool.Available("chromium", "windows"))
	}
	if err := os.Remove(realPath); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(dir, "external.json")
	if err := os.WriteFile(external, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, realPath); err != nil {
		t.Fatal(err)
	}
	pool, err = NewPool(dir)
	if err != nil {
		t.Fatal(err)
	}
	if pool.Available("chromium", "windows") != 0 {
		t.Fatal("NewPool followed a symlinked fingerprint file")
	}
}
