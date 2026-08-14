// Command back01-core exposes the smallest local authenticated Core needed to
// create and restore encrypted ForgeLocal profile backups. It deliberately does
// not import browser runtimes, fingerprinting, humanization, MCP or extensions.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forgelocal/internal/backup"
	"forgelocal/internal/profile"
	"forgelocal/internal/secrets"
)

type core struct {
	profiles *profile.Store
	backups  *backup.Service
	token    string
}

func main() {
	dataDir := os.Getenv("FORGELOCAL_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".forgelocal-back01")
	}
	token := os.Getenv("FORGELOCAL_API_TOKEN")
	if token == "" {
		log.Fatal("FORGELOCAL_API_TOKEN is required")
	}

	svc, sqliteStore, err := backup.OpenCoreService(dataDir)
	if err != nil {
		log.Fatalf("initialize BACK-01 Core: %v", err)
	}
	defer sqliteStore.Close()
	secretVault, ok := svc.Vault.(secrets.SecretVault)
	if !ok {
		log.Fatal("configured backup vault does not implement SecretVault")
	}
	profiles, err := profile.NewStore(filepath.Join(dataDir, "profiles"), secretVault)
	if err != nil {
		log.Fatalf("open profile store: %v", err)
	}

	bind := os.Getenv("FORGELOCAL_BIND")
	if bind == "" {
		bind = "127.0.0.1:45100"
	}
	if err := validateLoopbackBind(bind); err != nil {
		log.Fatalf("invalid FORGELOCAL_BIND: %v", err)
	}
	server := &http.Server{
		Addr:              bind,
		Handler:           &core{profiles: profiles, backups: svc, token: token},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Print("BACK-01 minimal Core listening on configured loopback address")
	log.Fatal(server.ListenAndServe())
}

func (c *core) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if !c.authorized(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "valid bearer token required")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodPost && len(parts) == 6 && parts[1] == "api" && parts[2] == "v1" && parts[3] == "profiles" && parts[5] == "backups":
		c.createBackup(w, parts[4])
	case r.Method == http.MethodPost && len(parts) == 6 && parts[1] == "api" && parts[2] == "v1" && parts[3] == "backups" && parts[5] == "restore":
		c.restoreBackup(w, r, parts[4])
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "route not available in BACK-01 minimal Core")
	}
}

func (c *core) authorized(r *http.Request) bool {
	value := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return value != "" && len(value) == len(c.token) && subtle.ConstantTimeCompare([]byte(value), []byte(c.token)) == 1
}

func (c *core) createBackup(w http.ResponseWriter, profileID string) {
	if _, err := c.profiles.Get(profileID); err != nil {
		writeError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "profile not found")
		return
	}
	keyID := "profile." + profileID
	if _, err := c.backups.Vault.Get(keyID); err != nil {
		key, keyErr := secrets.NewKey()
		if keyErr != nil || c.backups.Vault.Put(keyID, key) != nil {
			writeError(w, http.StatusInternalServerError, "BACKUP_KEY_FAILED", "could not prepare backup key")
			return
		}
	}
	artifact, err := c.backups.CreateSnapshot(profileID, keyID, func() ([]byte, error) {
		return c.profiles.CreateBackupSnapshot(profileID)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{
		"id": artifact.ID, "profile_id": artifact.ProfileID, "sha256": artifact.SHA256, "created_at": artifact.CreatedAt.UTC().Format(time.RFC3339Nano),
	}})
}

func (c *core) restoreBackup(w http.ResponseWriter, r *http.Request, backupID string) {
	defer r.Body.Close()
	var req struct {
		TargetProfileID string `json:"target_profile_id"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || req.TargetProfileID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "target_profile_id is required")
		return
	}
	if _, err := c.profiles.Get(req.TargetProfileID); err == nil {
		writeError(w, http.StatusConflict, "TARGET_EXISTS", "restore target profile already exists")
		return
	}
	targetPath, err := c.profiles.ProfilePath(req.TargetProfileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TARGET", "target_profile_id is invalid")
		return
	}
	restore, err := c.backups.RestoreWith(backupID, req.TargetProfileID, targetPath, func(payload []byte) error {
		_, restoreErr := c.profiles.RestoreBackupSnapshot(req.TargetProfileID, payload)
		return restoreErr
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, backup.ErrIntegrity) || errors.Is(err, backup.ErrFormat) {
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, "RESTORE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{
		"id": restore.ID, "backup_id": restore.BackupID, "target_profile_id": restore.TargetProfileID, "correlation_id": restore.CorrelationID,
	}})
}

func validateLoopbackBind(bind string) error {
	if strings.ContainsAny(bind, "\r\n") {
		return fmt.Errorf("control character")
	}
	host, port, err := net.SplitHostPort(bind)
	if err != nil || port == "" {
		return fmt.Errorf("expected host:port")
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("must bind to loopback only")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

var _ = fmt.Sprintf
