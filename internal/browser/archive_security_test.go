package browser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeArchiveFile(t *testing.T, suffix string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture"+suffix)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSecureExtractZipAcceptsLocalFileAndReplacesDestinationAtomically(t *testing.T) {
	archivePath := writeArchiveFile(t, ".zip", zipArchive(t, "bin/chrome", []byte("synthetic-runtime"), 0755))
	parent := t.TempDir()
	dest := filepath.Join(parent, "runtime")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.marker"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := secureExtractArchive(archivePath, dest); err != nil {
		t.Fatalf("secureExtractArchive() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dest, "bin", "chrome"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "synthetic-runtime" {
		t.Fatalf("extracted data = %q", data)
	}
	if _, err := os.Stat(filepath.Join(dest, "old.marker")); !os.IsNotExist(err) {
		t.Fatalf("old destination was not replaced: err=%v", err)
	}
}

func TestSecureExtractZipRejectsTraversalAndPreservesDestination(t *testing.T) {
	archivePath := writeArchiveFile(t, ".zip", zipArchive(t, "../escape.txt", []byte("must-not-extract"), 0644))
	parent := t.TempDir()
	dest := filepath.Join(parent, "runtime")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dest, "keep.marker")
	if err := os.WriteFile(marker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := secureExtractArchive(archivePath, dest); err == nil {
		t.Fatal("traversal archive was accepted")
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
		t.Fatalf("destination changed after rejected archive: data=%q err=%v", data, err)
	}
	if _, err := os.Stat(filepath.Join(parent, "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("traversal output exists: err=%v", err)
	}
}

func TestSecureExtractZipRejectsAbsoluteAndTooDeepPaths(t *testing.T) {
	cases := []string{"/tmp/escape.txt", strings.Repeat("nested/", maxArchivePathDepth) + "escape.txt"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			archivePath := writeArchiveFile(t, ".zip", zipArchive(t, name, []byte("blocked"), 0644))
			dest := filepath.Join(t.TempDir(), "runtime")
			if err := os.MkdirAll(dest, 0755); err != nil {
				t.Fatal(err)
			}
			if err := secureExtractArchive(archivePath, dest); err == nil {
				t.Fatal("unsafe archive path was accepted")
			}
		})
	}
}

func TestSecureExtractTarRejectsSymlinkAndTraversal(t *testing.T) {
	cases := []struct {
		name     string
		entry    string
		typeflag byte
		linkname string
	}{
		{name: "symlink", entry: "bin/link", typeflag: tar.TypeSymlink, linkname: "../../escape"},
		{name: "traversal", entry: "../escape", typeflag: tar.TypeReg, linkname: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			archivePath := writeArchiveFile(t, ".tar.gz", tarGzArchive(t, tc.entry, tc.typeflag, []byte("blocked"), tc.linkname))
			dest := filepath.Join(t.TempDir(), "runtime")
			if err := os.MkdirAll(dest, 0755); err != nil {
				t.Fatal(err)
			}
			if err := secureExtractArchive(archivePath, dest); err == nil {
				t.Fatal("unsafe tar archive was accepted")
			}
			if _, err := os.Stat(filepath.Join(dest, "bin", "link")); !os.IsNotExist(err) {
				t.Fatalf("partial symlink output exists: %v", err)
			}
		})
	}
}

func TestSecureExtractTarAcceptsRegularFile(t *testing.T) {
	archivePath := writeArchiveFile(t, ".tar.gz", tarGzArchive(t, "bin/chrome", tar.TypeReg, []byte("synthetic-tar-runtime"), ""))
	dest := filepath.Join(t.TempDir(), "runtime")
	if err := secureExtractArchive(archivePath, dest); err != nil {
		t.Fatalf("secureExtractArchive() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dest, "bin", "chrome"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "synthetic-tar-runtime" {
		t.Fatalf("extracted data = %q", data)
	}
}

func tarGzArchive(t *testing.T, name string, typeflag byte, data []byte, linkname string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: name, Mode: 0644, Typeflag: typeflag, Linkname: linkname}
	if typeflag == tar.TypeReg {
		hdr.Size = int64(len(data))
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if typeflag == tar.TypeReg {
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
