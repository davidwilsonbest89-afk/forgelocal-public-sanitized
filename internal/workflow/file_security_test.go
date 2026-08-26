package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "workflow.yaml")
	content := []byte("name: synthetic\nsteps:\n  - name: sleep\n    action: sleep\n    params:\n      seconds: 0\n")
	if err := os.WriteFile(realPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine("http://127.0.0.1:19280", "synthetic-token")
	loaded, err := engine.LoadFile(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "synthetic" || len(loaded.Steps) != 1 {
		t.Fatalf("unexpected workflow: %+v", loaded)
	}
	linkPath := filepath.Join(dir, "link.yaml")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.LoadFile(linkPath); err == nil {
		t.Fatal("LoadFile followed a symlink")
	}
}
