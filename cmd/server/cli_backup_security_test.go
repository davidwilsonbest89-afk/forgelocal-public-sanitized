package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIFullBackupRestoreRejectsUnsafeEntriesWithoutPartialActivation(t *testing.T) {
	cases := []struct {
		name     string
		entry    string
		typeflag byte
		linkname string
	}{
		{name: "traversal after valid entry", entry: "profiles/../../escape.txt", typeflag: tar.TypeReg},
		{name: "windows separator traversal", entry: `profiles/..\..\escape.txt`, typeflag: tar.TypeReg},
		{name: "double windows separator traversal", entry: `profiles/..\\..\\escape.txt`, typeflag: tar.TypeReg},
		{name: "external symlink", entry: "profiles/escape", typeflag: tar.TypeSymlink, linkname: "../../escape"},
		{name: "windows volume symlink", entry: "profiles/windows-escape", typeflag: tar.TypeSymlink, linkname: `C:\\outside`},
		{name: "hardlink", entry: "profiles/hardlink", typeflag: tar.TypeLink, linkname: "profiles/profile.json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backupPath := writeRestoreTarGz(t, []restoreArchiveEntry{
				{name: "profiles/profile.json", typeflag: tar.TypeReg, body: []byte(`{"id":"staged"}`)},
				{name: tc.entry, typeflag: tc.typeflag, linkname: tc.linkname},
			})
			baseDir := t.TempDir()
			marker := filepath.Join(baseDir, "marker.txt")
			if err := os.WriteFile(marker, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := runCLI([]string{"--base-dir", baseDir, "backup", "restore", "--full", "--json", backupPath}, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("unsafe restore succeeded: stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
			var result map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("decode restore result: %v; stdout=%s", err, stdout.String())
			}
			if ok, _ := result["ok"].(bool); ok {
				t.Fatalf("restore result marked ok: %#v", result)
			}
			if data, err := os.ReadFile(marker); err != nil || string(data) != "original" {
				t.Fatalf("existing destination changed: data=%q err=%v", data, err)
			}
			if _, err := os.Stat(filepath.Join(baseDir, "profiles")); !os.IsNotExist(err) {
				t.Fatalf("partial profiles root activated: err=%v", err)
			}
			if _, err := os.Stat(filepath.Join(baseDir, "escape.txt")); !os.IsNotExist(err) {
				t.Fatalf("escape output exists: err=%v", err)
			}
		})
	}
}

func TestCLIFullBackupRestoreRejectsTooManyEntries(t *testing.T) {
	entries := make([]restoreArchiveEntry, 0, maxRestoreFiles+1)
	for i := 0; i <= maxRestoreFiles; i++ {
		entries = append(entries, restoreArchiveEntry{name: filepath.Join("profiles", "file-"+itoa(i)), typeflag: tar.TypeReg, body: []byte("x")})
	}
	backupPath := writeRestoreTarGz(t, entries)
	baseDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"--base-dir", baseDir, "backup", "restore", "--full", "--json", backupPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("oversized entry-count restore succeeded: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(baseDir, "profiles")); !os.IsNotExist(err) {
		t.Fatalf("entry-count-limited restore activated output: err=%v", err)
	}
}

type restoreArchiveEntry struct {
	name     string
	typeflag byte
	linkname string
	body     []byte
	size     int64
}

func itoa(value int) string {
	return fmt.Sprintf("%d", value)
}

func writeRestoreTarGz(t *testing.T, entries []restoreArchiveEntry) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		size := entry.size
		if entry.typeflag == tar.TypeReg && size == 0 {
			size = int64(len(entry.body))
		}
		header := &tar.Header{Name: entry.name, Mode: 0644, Typeflag: entry.typeflag, Linkname: entry.linkname, Size: size}
		if entry.typeflag == tar.TypeDir {
			header.Mode = 0755
			header.Size = 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if entry.typeflag == tar.TypeReg && len(entry.body) > 0 {
			if _, err := tw.Write(entry.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "backup.tgz")
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
