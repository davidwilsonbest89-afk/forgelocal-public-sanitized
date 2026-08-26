package browser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenBrowserRegularFileRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "regular.bin")
	if err := os.WriteFile(realPath, []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	root, file, err := openBrowserRegularFile(realPath, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := root.Close(); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(dir, "link.bin")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Fatal(err)
	}
	if _, _, err := openBrowserRegularFile(linkPath, 0, 0); err == nil {
		t.Fatal("openBrowserRegularFile followed a symlink")
	}
}
