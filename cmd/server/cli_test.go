package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHelpAndVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("help exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "BrowseForge CLI") {
		t.Fatalf("help output missing title: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("version exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "BrowseForge") {
		t.Fatalf("version output = %s", stdout.String())
	}
}

func TestCLISubcommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"serve", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("serve help exit = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "BrowseForge CLI") {
		t.Fatalf("serve help output = %s", stdout.String())
	}
}

func TestCLIUnknownCommandFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"nope"}, &stdout, &stderr); code == 0 {
		t.Fatalf("unknown command unexpectedly succeeded")
	}
	if !strings.Contains(stderr.String(), "Unknown command: nope") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestCLIInitConfigTokenAndDoctorJSON(t *testing.T) {
	baseDir := t.TempDir()
	var stdout, stderr bytes.Buffer

	if code := runCLI([]string{"--base-dir", baseDir, "init", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(baseDir, "config.json")); err != nil {
		t.Fatalf("config not created: %v", err)
	}
	for _, dir := range []string{"profiles", "data", "logs"} {
		if _, err := os.Stat(filepath.Join(baseDir, dir)); err != nil {
			t.Fatalf("%s dir not created: %v", dir, err)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--base-dir", baseDir, "config", "validate", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config validate exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var validation map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &validation); err != nil {
		t.Fatalf("decode config validate: %v\n%s", err, stdout.String())
	}
	if validation["ok"] != true {
		t.Fatalf("validation = %#v", validation)
	}

	tokenDir := filepath.Join(baseDir, "data")
	if err := os.WriteFile(filepath.Join(tokenDir, ".api-token"), []byte("test-token\n"), 0600); err != nil {
		t.Fatalf("write token: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--base-dir", baseDir, "token", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("token exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var token map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &token); err != nil {
		t.Fatalf("decode token: %v\n%s", err, stdout.String())
	}
	if token["token"] != "test-token" {
		t.Fatalf("token output = %#v", token)
	}

	stdout.Reset()
	stderr.Reset()
	code := runCLI([]string{"--base-dir", baseDir, "doctor", "--json"}, &stdout, &stderr)
	if code != 0 && code != 1 {
		t.Fatalf("doctor exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var report doctorReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode doctor: %v\n%s", err, stdout.String())
	}
	if report.Version == "" || len(report.Checks) == 0 {
		t.Fatalf("doctor report = %#v", report)
	}
}

func TestCLICapabilitiesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"capabilities", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("capabilities exit = %d, stderr = %s", code, stderr.String())
	}
	var caps map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &caps); err != nil {
		t.Fatalf("decode capabilities: %v\n%s", err, stdout.String())
	}
	commands, ok := caps["commands"].([]any)
	if !ok || len(commands) == 0 {
		t.Fatalf("commands missing: %#v", caps)
	}
}

func TestCLIStatusAndMCPConfigJSON(t *testing.T) {
	baseDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"--base-dir", baseDir, "init"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init exit = %d, stderr = %s", code, stderr.String())
	}
	if err := os.WriteFile(filepath.Join(baseDir, "data", ".api-token"), []byte("test-token\n"), 0600); err != nil {
		t.Fatalf("write token: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--base-dir", baseDir, "status", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("status exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var status map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		t.Fatalf("decode status: %v\n%s", err, stdout.String())
	}
	if status["mcp_url"] == "" || status["dashboard"] == "" {
		t.Fatalf("status missing URLs: %#v", status)
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--base-dir", baseDir, "mcp-config", "http", "--url", "http://example.test/mcp", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("mcp-config exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var cfg map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &cfg); err != nil {
		t.Fatalf("decode mcp config: %v\n%s", err, stdout.String())
	}
	if _, ok := cfg["browseforge"].(map[string]any); !ok {
		t.Fatalf("mcp config = %#v", cfg)
	}
}

func TestCLIFullBackupCreateAndRestore(t *testing.T) {
	baseDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(baseDir, "profiles", "prof_test"), 0755); err != nil {
		t.Fatalf("mkdir profile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "data"), 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "profiles", "prof_test", "profile.json"), []byte(`{"id":"prof_test"}`), 0644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "data", ".api-token"), []byte("test-token"), 0600); err != nil {
		t.Fatalf("write token: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "browsers", "camoufox"), 0755); err != nil {
		t.Fatalf("mkdir browser: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "browsers", "camoufox", "target.txt"), []byte("browser-data"), 0644); err != nil {
		t.Fatalf("write browser target: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(baseDir, "browsers", "camoufox", "link.txt")); err != nil {
		t.Fatalf("create browser symlink: %v", err)
	}

	backupDir := filepath.Join(t.TempDir(), "backups")
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"--base-dir", baseDir, "backup", "create", "--full", "--output", backupDir, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("backup create exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var created map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &created); err != nil {
		t.Fatalf("decode backup create: %v\n%s", err, stdout.String())
	}
	backupPath, _ := created["path"].(string)
	if backupPath == "" {
		t.Fatalf("backup path missing: %#v", created)
	}
	if filepath.Dir(backupPath) != backupDir {
		t.Fatalf("backup path = %s, want dir %s", backupPath, backupDir)
	}
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("backup not written: %v", err)
	}

	restoreDir := t.TempDir()
	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--base-dir", restoreDir, "backup", "restore", "--full", "--json", backupPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("backup restore exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(restoreDir, "profiles", "prof_test", "profile.json")); err != nil {
		t.Fatalf("profile not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(restoreDir, "data", ".api-token")); err != nil {
		t.Fatalf("token not restored: %v", err)
	}
	linkTarget, err := os.Readlink(filepath.Join(restoreDir, "browsers", "camoufox", "link.txt"))
	if err != nil {
		t.Fatalf("browser symlink not restored: %v", err)
	}
	if linkTarget != "target.txt" {
		t.Fatalf("browser symlink target = %q", linkTarget)
	}
}

func TestCLIInvalidConfigDoesNotPanic(t *testing.T) {
	baseDir := t.TempDir()
	configPath := filepath.Join(baseDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"--base-dir", baseDir, "smoke", "rest", "--json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("smoke exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode smoke error: %v\n%s", err, stdout.String())
	}
	errMsg, _ := result["error"].(string)
	if result["ok"] != false || !strings.Contains(errMsg, "config error") {
		t.Fatalf("smoke result = %#v", result)
	}
}

func TestCLIAPIListReportsMissingToken(t *testing.T) {
	baseDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"--base-dir", baseDir, "init"}, &stdout, &stderr); code != 0 {
		t.Fatalf("init exit = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code := runCLI([]string{"--base-dir", baseDir, "profiles", "list", "--base-url", "http://127.0.0.1:1", "--json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("profiles list exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode profiles error: %v\n%s", err, stdout.String())
	}
	errMsg, _ := result["error"].(string)
	if result["ok"] != false || !strings.Contains(errMsg, "token error") {
		t.Fatalf("profiles result = %#v", result)
	}
}

func TestTokenPreviewShortToken(t *testing.T) {
	if got := tokenPreview("short"); got != "short" {
		t.Fatalf("tokenPreview short = %q", got)
	}
	if got := tokenPreview("1234567890abcdefX"); got != "1234567890abcdef" {
		t.Fatalf("tokenPreview long = %q", got)
	}
}
