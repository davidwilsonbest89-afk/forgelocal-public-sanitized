package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupLoggerUsesPrivateDirectoryAndFile(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "nested", "runtime.log")
	logger := SetupLogger(logPath)
	logger.LogAttrs(nil, slog.LevelInfo, "synthetic logger check")

	dirInfo, err := os.Stat(filepath.Dir(logPath))
	if err != nil {
		t.Fatal(err)
	}
	if mode := dirInfo.Mode().Perm(); mode != 0700 {
		t.Fatalf("log directory mode = %04o, want 0700", mode)
	}
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if mode := fileInfo.Mode().Perm(); mode != 0600 {
		t.Fatalf("log file mode = %04o, want 0600", mode)
	}
}

func TestOpenConfigFileRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "real.json")
	if err := os.WriteFile(target, []byte(`{"port":"19280"}`), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	openedRoot, file, err := openConfigFile(link, 0, 0)
	if err == nil {
		_ = file.Close()
		_ = openedRoot.Close()
		t.Fatal("symlinked configuration unexpectedly opened")
	}
}

func TestLoadMissingConfigKeepsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "19280" || cfg.DataDir != "data" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
