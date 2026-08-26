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
