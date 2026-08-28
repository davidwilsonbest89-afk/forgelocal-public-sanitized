package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/backup"
	"forgelocal/internal/secrets"
)

func backupKeyID(profileID string) string { return "profile." + profileID }

func (h *handler) ensureBackupKey(profileID string) (string, error) {
	if h.backupSvc == nil || h.backupSvc.Vault == nil {
		return "", errors.New("backup vault is not configured")
	}
	keyID := backupKeyID(profileID)
	if _, err := h.backupSvc.Vault.Get(keyID); err == nil {
		return keyID, nil
	}
	key, err := secrets.NewKey()
	if err != nil {
		return "", err
	}
	if err := h.backupSvc.Vault.Put(keyID, key); err != nil {
		return "", err
	}
	return keyID, nil
}

func (h *handler) profileHasLiveSession(profileID string) bool {
	for _, session := range h.mgr.ListSessions() {
		if session != nil && session.ProfileID == profileID {
			return true
		}
	}
	return false
}

func (h *handler) createBackupV1(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "id")
	if _, err := h.store.Get(profileID); err != nil {
		writeError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", err.Error())
		return
	}
	if h.mgr == nil {
		writeError(w, http.StatusServiceUnavailable, "SESSION_MANAGER_UNAVAILABLE", "session manager is not configured")
		return
	}
	releaseSnapshot, err := h.mgr.AcquireSnapshot(profileID)
	if err != nil {
		writeError(w, http.StatusConflict, "PROFILE_ACTIVE_OR_BUSY", err.Error())
		return
	}
	defer releaseSnapshot()
	keyID, err := h.ensureBackupKey(profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_KEY_FAILED", err.Error())
		return
	}
	b, err := h.backupSvc.CreateSnapshot(profileID, keyID, func() ([]byte, error) {
		return h.store.CreateBackupSnapshot(profileID)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{
		"id": b.ID, "profile_id": b.ProfileID, "sha256": b.SHA256, "created_at": b.CreatedAt.UTC().Format(time.RFC3339Nano),
	}})
}

func (h *handler) restoreBackupV1(w http.ResponseWriter, r *http.Request) {
	backupID := chi.URLParam(r, "id")
	var req struct {
		TargetProfileID string `json:"target_profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.TargetProfileID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TARGET_PROFILE_ID", "target_profile_id is required")
		return
	}
	if _, err := h.store.Get(req.TargetProfileID); err == nil {
		writeError(w, http.StatusConflict, "TARGET_EXISTS", "target profile id already exists")
		return
	}
	if h.mgr == nil {
		writeError(w, http.StatusServiceUnavailable, "SESSION_MANAGER_UNAVAILABLE", "session manager is not configured")
		return
	}
	releaseSnapshot, err := h.mgr.AcquireSnapshot(req.TargetProfileID)
	if err != nil {
		writeError(w, http.StatusConflict, "TARGET_ACTIVE_OR_BUSY", err.Error())
		return
	}
	defer releaseSnapshot()
	targetPath, err := h.store.ProfilePath(req.TargetProfileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TARGET_PROFILE_ID", err.Error())
		return
	}
	lockRelease, err := h.backupSvc.AcquireProfileLock(req.TargetProfileID)
	if err != nil {
		writeError(w, http.StatusConflict, "TARGET_BUSY", err.Error())
		return
	}
	defer lockRelease()
	restored, err := h.backupSvc.RestoreWith(backupID, req.TargetProfileID, targetPath, func(payload []byte) error {
		_, err := h.store.RestoreBackupSnapshot(req.TargetProfileID, payload)
		return err
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, backup.ErrIntegrity) {
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, "RESTORE_FAILED", err.Error())
		return
	}
	p, err := h.store.Get(restored.TargetProfileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "RESTORE_PROFILE_MISSING", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]any{
		"restore_id": restored.ID, "backup_id": restored.BackupID, "profile": p,
	}})
}

// listBackupsV1 returns only dashboard-safe backup metadata. Artifact paths and
// key identifiers never cross the API boundary.
func (h *handler) listBackupsV1(w http.ResponseWriter, r *http.Request) {
	if h.backupDB == nil {
		writeError(w, http.StatusServiceUnavailable, "BACKUP_CATALOG_UNAVAILABLE", "backup catalog is not configured")
		return
	}
	rows, err := h.backupDB.QueryContext(r.Context(), `
		SELECT b.id, b.profile_id, b.sha256, COALESCE(o.state, 'committed'), COALESCE(o.error_code, ''), b.created_at, COALESCE(o.updated_at, b.created_at)
		FROM backups b LEFT JOIN backup_operations o ON o.id = b.id ORDER BY b.created_at DESC, b.id DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_CATALOG_FAILED", "backup catalog unavailable")
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, profileID, digest, state, errorCode, createdAt, updatedAt string
		if err := rows.Scan(&id, &profileID, &digest, &state, &errorCode, &createdAt, &updatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "BACKUP_CATALOG_FAILED", "backup catalog unavailable")
			return
		}
		item := map[string]any{"id": id, "profile_id": profileID, "sha256": digest, "state": state, "created_at": createdAt, "updated_at": updatedAt}
		if errorCode != "" {
			item["error_code"] = errorCode
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_CATALOG_FAILED", "backup catalog unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// getBackupV1 returns one dashboard-safe backup projection and the last restore target.
func (h *handler) getBackupV1(w http.ResponseWriter, r *http.Request) {
	if h.backupDB == nil {
		writeError(w, http.StatusServiceUnavailable, "BACKUP_CATALOG_UNAVAILABLE", "backup catalog is not configured")
		return
	}
	id := chi.URLParam(r, "id")
	var profileID, digest, state, errorCode, createdAt, updatedAt string
	err := h.backupDB.QueryRowContext(r.Context(), `
		SELECT b.profile_id, b.sha256, COALESCE(o.state, 'committed'), COALESCE(o.error_code, ''), b.created_at, COALESCE(o.updated_at, b.created_at)
		FROM backups b LEFT JOIN backup_operations o ON o.id = b.id WHERE b.id = ?`, id).
		Scan(&profileID, &digest, &state, &errorCode, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "BACKUP_NOT_FOUND", "backup not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_LOOKUP_FAILED", "backup lookup failed")
		return
	}
	item := map[string]any{"id": id, "profile_id": profileID, "sha256": digest, "state": state, "created_at": createdAt, "updated_at": updatedAt}
	if errorCode != "" {
		item["error_code"] = errorCode
	}
	var target string
	if err := h.backupDB.QueryRowContext(r.Context(), `SELECT target_profile_id FROM restore_operations WHERE backup_id = ? ORDER BY updated_at DESC, id DESC LIMIT 1`, id).Scan(&target); err == nil && target != "" {
		item["last_restored_target_profile_id"] = target
	}
	item["quarantined"] = state == "quarantined"
	writeJSON(w, http.StatusOK, item)
}

// listBackupRestoresV1 returns restore state only; target paths are never exposed.
func (h *handler) listBackupRestoresV1(w http.ResponseWriter, r *http.Request) {
	if h.backupDB == nil {
		writeError(w, http.StatusServiceUnavailable, "BACKUP_CATALOG_UNAVAILABLE", "backup catalog is not configured")
		return
	}
	id := chi.URLParam(r, "id")
	var exists int
	if err := h.backupDB.QueryRowContext(r.Context(), `SELECT COUNT(1) FROM backups WHERE id = ?`, id).Scan(&exists); err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_LOOKUP_FAILED", "backup lookup failed")
		return
	}
	if exists == 0 {
		writeError(w, http.StatusNotFound, "BACKUP_NOT_FOUND", "backup not found")
		return
	}
	rows, err := h.backupDB.QueryContext(r.Context(), `SELECT id, backup_id, source_profile_id, target_profile_id, state, error_code, created_at, updated_at FROM restore_operations WHERE backup_id = ? ORDER BY created_at DESC, id DESC`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "RESTORE_CATALOG_FAILED", "restore catalog unavailable")
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var restoreID, backupID, sourceID, targetID, state, errorCode, createdAt, updatedAt string
		if err := rows.Scan(&restoreID, &backupID, &sourceID, &targetID, &state, &errorCode, &createdAt, &updatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "RESTORE_CATALOG_FAILED", "restore catalog unavailable")
			return
		}
		item := map[string]any{"restore_id": restoreID, "backup_id": backupID, "source_profile_id": sourceID, "target_profile_id": targetID, "state": state, "created_at": createdAt, "updated_at": updatedAt}
		if errorCode != "" {
			item["error_code"] = errorCode
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "RESTORE_CATALOG_FAILED", "restore catalog unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// deleteBackupV1 deletes only a committed local backup artifact and its metadata.
func (h *handler) deleteBackupV1(w http.ResponseWriter, r *http.Request) {
	if h.backupDB == nil || h.backupSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "BACKUP_CATALOG_UNAVAILABLE", "backup catalog is not configured")
		return
	}
	id := chi.URLParam(r, "id")
	var artifactPath string
	if err := h.backupDB.QueryRowContext(r.Context(), `SELECT artifact_path FROM backups WHERE id = ?`, id).Scan(&artifactPath); errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "BACKUP_NOT_FOUND", "backup not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_LOOKUP_FAILED", "backup lookup failed")
		return
	}
	root, err := filepath.Abs(h.backupSvc.Root)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_DELETE_FAILED", "backup root unavailable")
		return
	}
	artifact, err := filepath.Abs(artifactPath)
	if err != nil || artifact == root || !strings.HasPrefix(artifact, root+string(os.PathSeparator)) {
		writeError(w, http.StatusForbidden, "BACKUP_PATH_REJECTED", "backup artifact path rejected")
		return
	}
	if err := os.Remove(artifact); err != nil && !errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusInternalServerError, "BACKUP_DELETE_FAILED", "backup artifact deletion failed")
		return
	}
	tx, err := h.backupDB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_DELETE_FAILED", "backup metadata deletion failed")
		return
	}
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM restore_operations WHERE backup_id = ?`, id); err == nil {
		_, err = tx.ExecContext(r.Context(), `DELETE FROM backup_operations WHERE id = ?`, id)
	}
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `DELETE FROM backups WHERE id = ?`, id)
	}
	if err != nil {
		_ = tx.Rollback()
		writeError(w, http.StatusInternalServerError, "BACKUP_DELETE_FAILED", "backup metadata deletion failed")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "BACKUP_DELETE_FAILED", "backup metadata deletion failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}
