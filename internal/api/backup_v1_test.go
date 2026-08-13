package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forgelocal/internal/backup"
	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
	"forgelocal/internal/secrets"
)

func TestBackupV1CreateModifyRestoreIsolation(t *testing.T) {
	root := t.TempDir()
	profiles, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil {
		t.Fatal(err)
	}
	source := &profile.Profile{ID: "source-api", Name: "API source", RuntimeID: "chromium"}
	if err := profiles.Create(source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source.ProfileDir, "browser-data", "Cookies"), []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := backup.OpenSQLite(filepath.Join(root, "metadata.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	vault := secrets.NewMemoryVault()
	svc := &backup.Service{Root: filepath.Join(root, "backups"), Vault: vault, Store: db, Locks: backup.NewProfileLocks()}
	cfg := &config.Config{DataDir: root, ProfilesDir: filepath.Join(root, "profiles"), Version: "test"}
	router, err := NewRouter(cfg, profiles, &browser.Manager{}, nil, svc)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/profiles/source-api/backups", nil)
	request.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create backup status=%d body=%s", response.Code, response.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.ID == "" {
		t.Fatal("backup id missing")
	}
	if err := os.WriteFile(filepath.Join(source.ProfileDir, "browser-data", "Cookies"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}

	body := httptest.NewRequest(http.MethodPost, "/api/v1/backups/"+created.Data.ID+"/restore", strings.NewReader(`{"target_profile_id":"target-api"}`))
	body.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	restoredResponse := httptest.NewRecorder()
	router.ServeHTTP(restoredResponse, body)
	if restoredResponse.Code != http.StatusCreated {
		t.Fatalf("restore status=%d body=%s", restoredResponse.Code, restoredResponse.Body.String())
	}
	got, err := os.ReadFile(filepath.Join(root, "profiles", "target-api", "browser-data", "Cookies"))
	if err != nil || string(got) != "before" {
		t.Fatalf("restored source snapshot mismatch: %q, %v", got, err)
	}
	original, err := os.ReadFile(filepath.Join(source.ProfileDir, "browser-data", "Cookies"))
	if err != nil || string(original) != "changed" {
		t.Fatalf("source was overwritten: %q, %v", original, err)
	}
	assertRestoredProfileStartsLocalChromium(t, filepath.Join(root, "profiles", "target-api", "browser-data"))
}

// assertRestoredProfileStartsLocalChromium performs the local-only final step
// of AC-BACK-01. CI runners without Chromium keep the crypto/API proof green,
// while release validation must execute this test on a Chromium-equipped host.
func assertRestoredProfileStartsLocalChromium(t *testing.T, userDataDir string) {
	t.Helper()
	binary, err := exec.LookPath("chromium")
	if err != nil {
		t.Log("local Chromium unavailable; runtime relaunch must run in release validation")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary,
		"--headless=new", "--no-first-run", "--no-default-browser-check",
		"--disable-gpu", "--no-sandbox", "--user-data-dir="+userDataDir,
		"--dump-dom", "about:blank")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("restored Chromium launch timed out: %v; output=%s", ctx.Err(), output)
	}
	if err != nil {
		t.Fatalf("restored Chromium launch failed: %v; output=%s", err, output)
	}
}
