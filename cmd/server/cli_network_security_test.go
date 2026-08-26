package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateCLILoopbackURLBoundaries(t *testing.T) {
	accepted := []string{
		"http://127.0.0.1:19281/api/status",
		"http://localhost:19281/api/status",
		"http://[::1]:19281/api/status",
		"https://127.0.0.1:443/api/status",
	}
	for _, raw := range accepted {
		t.Run("accept/"+raw, func(t *testing.T) {
			if _, err := validateCLILoopbackURL(raw); err != nil {
				t.Fatalf("validateCLILoopbackURL(%q) error = %v", raw, err)
			}
		})
	}

	rejected := []string{
		"http://example.invalid/api/status",
		"ws://127.0.0.1:19281/api/status",
		"file:///tmp/status",
		"http://user:pass@127.0.0.1:19281/api/status",
		"http://127.0.0.1:19281/api/status?token=secret",
		"http://127.0.0.1:19281/api/status#fragment",
		"http://127.0.0.1:0/api/status",
		"http://127.0.0.1:65536/api/status",
		"http://127.0.0.1:not-a-port/api/status",
		"http:///api/status",
	}
	for _, raw := range rejected {
		t.Run("reject/"+raw, func(t *testing.T) {
			if _, err := validateCLILoopbackURL(raw); err == nil {
				t.Fatalf("validateCLILoopbackURL(%q) accepted unsafe URL", raw)
			}
		})
	}
}

func TestCLILocalHTTPClientRejectsExternalRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://example.invalid/external", http.StatusFound)
	}))
	defer server.Close()

	client := newCLILocalHTTPClient(2 * time.Second)
	_, err := client.Get(server.URL + "/redirect")
	if err == nil || !strings.Contains(err.Error(), "external redirect refused") {
		t.Fatalf("redirect error = %v, want external redirect refusal", err)
	}
}

func TestAPIGETRejectsExternalBeforeDial(t *testing.T) {
	if _, err := apiGET("http://example.invalid/api/status", ""); err == nil {
		t.Fatal("apiGET accepted external URL")
	}
}

func TestAPIPOSTRejectsExternalBeforeDial(t *testing.T) {
	if _, err := apiPOST("http://example.invalid/api/mutate", "synthetic-token", map[string]any{"ok": true}); err == nil {
		t.Fatal("apiPOST accepted external URL")
	}
}

func TestOpenCommandRejectsExternalBeforeBrowserLaunch(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runOpenCommand([]string{"--base-url", "http://example.invalid"}, cliGlobal{baseDir: t.TempDir(), configPath: "/nonexistent/config.json"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "loopback") {
		t.Fatalf("open external URL code=%d stderr=%q", code, stderr.String())
	}
}

func TestMetadataBackupRejectsExternalBeforeTokenUse(t *testing.T) {
	global := cliGlobal{baseDir: t.TempDir(), configPath: "/nonexistent/config.json"}
	if err := createMetadataBackup(global, "/tmp/synthetic-backup.tar.gz", "http://example.invalid", "synthetic-token"); err == nil {
		t.Fatal("createMetadataBackup accepted external URL")
	}
	if err := restoreMetadataBackup(global, "/nonexistent/backup.tar.gz", "http://example.invalid", "synthetic-token"); err == nil {
		t.Fatal("restoreMetadataBackup accepted external URL")
	}
}
