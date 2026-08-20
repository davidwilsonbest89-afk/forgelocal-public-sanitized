package api

import (
	"encoding/json"
	"errors"
	"net/http"
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
