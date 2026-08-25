package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"forgelocal/internal/extensions"
	"github.com/go-chi/chi/v5"
)

type extensionApprovalRequest struct {
	PermissionsAcknowledged []string `json:"permissions_acknowledged"`
	AcceptHighRisk          bool     `json:"accept_high_risk"`
}

type extensionAssignmentRequest struct {
	ProfileID string `json:"profile_id"`
}

type extensionRollbackRequest struct {
	TargetVersionID string `json:"target_version_id"`
}

func (h *handler) importExtension(w http.ResponseWriter, r *http.Request) {
	if h.extensionStore == nil {
		writeError(w, http.StatusServiceUnavailable, "EXTENSION_REPOSITORY_UNAVAILABLE", "extension repository unavailable")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, extensions.MaxZIPBytes+1<<20)
	// #nosec G120 -- MaxBytesReader caps the request body before parsing the multipart form.
	if err := r.ParseMultipartForm(extensions.MaxZIPBytes + 1<<20); err != nil {
		writeExtensionError(w, err)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("package")
	if err != nil {
		file, _, err = r.FormFile("file")
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARCHIVE", "local ZIP package is required")
		return
	}
	defer file.Close()
	result, err := h.extensionStore.Import(r.Context(), file, r.FormValue("series_id"), correlationIDFrom(r.Context()))
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": result})
}

func (h *handler) listExtensions(w http.ResponseWriter, r *http.Request) {
	if h.extensionStore == nil {
		writeError(w, http.StatusServiceUnavailable, "EXTENSION_REPOSITORY_UNAVAILABLE", "extension repository unavailable")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.extensionStore.List(r.Context(), limit, offset)
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) getExtensionSeries(w http.ResponseWriter, r *http.Request) {
	if h.extensionStore == nil {
		writeError(w, http.StatusServiceUnavailable, "EXTENSION_REPOSITORY_UNAVAILABLE", "extension repository unavailable")
		return
	}
	result, err := h.extensionStore.GetSeries(r.Context(), chi.URLParam(r, "seriesID"))
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *handler) approveExtension(w http.ResponseWriter, r *http.Request) {
	var input extensionApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "approval body is invalid")
		return
	}
	result, err := h.extensionStore.Approve(r.Context(), chi.URLParam(r, "versionID"), input.PermissionsAcknowledged, input.AcceptHighRisk, correlationIDFrom(r.Context()))
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *handler) assignExtension(w http.ResponseWriter, r *http.Request) {
	var input extensionAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "assignment body is invalid")
		return
	}
	result, err := h.extensionStore.Assign(r.Context(), chi.URLParam(r, "versionID"), input.ProfileID, correlationIDFrom(r.Context()), func(_ context.Context, id string) (bool, error) {
		_, profileErr := h.store.Get(id)
		return profileErr == nil, nil
	})
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *handler) updateExtension(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, extensions.MaxZIPBytes+1<<20)
	// #nosec G120 -- MaxBytesReader caps the request body before parsing the multipart form.
	if err := r.ParseMultipartForm(extensions.MaxZIPBytes + 1<<20); err != nil {
		writeExtensionError(w, err)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("package")
	if err != nil {
		file, _, err = r.FormFile("file")
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARCHIVE", "local ZIP package is required")
		return
	}
	defer file.Close()
	result, err := h.extensionStore.Update(r.Context(), chi.URLParam(r, "seriesID"), file, correlationIDFrom(r.Context()))
	if err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": result})
}

func (h *handler) rollbackExtension(w http.ResponseWriter, r *http.Request) {
	var input extensionRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "rollback body is invalid")
		return
	}
	if err := h.extensionStore.Rollback(r.Context(), chi.URLParam(r, "seriesID"), input.TargetVersionID, correlationIDFrom(r.Context())); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"state": "rolled_back"}})
}

func (h *handler) revokeExtension(w http.ResponseWriter, r *http.Request) {
	if err := h.extensionStore.Revoke(r.Context(), chi.URLParam(r, "versionID"), correlationIDFrom(r.Context())); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"state": "quarantined"}})
}

func (h *handler) purgeExtension(w http.ResponseWriter, r *http.Request) {
	if err := h.extensionStore.Purge(r.Context(), chi.URLParam(r, "versionID"), correlationIDFrom(r.Context())); err != nil {
		writeExtensionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"state": "purged"}})
}

func writeExtensionError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	code := "INVALID_ARCHIVE"
	message := "extension operation refused"
	switch {
	case errors.Is(err, extensions.ErrArchiveLimit):
		code, message = "ARCHIVE_LIMIT_EXCEEDED", "archive limits exceeded"
	case errors.Is(err, extensions.ErrManifestInvalid):
		code, message = "MANIFEST_INVALID", "manifest is invalid"
	case errors.Is(err, extensions.ErrSeriesNotFound):
		status, code, message = http.StatusNotFound, "SERIES_NOT_FOUND", "extension series not found"
	case errors.Is(err, extensions.ErrVersionNotFound), errors.Is(err, extensions.ErrNotFound):
		status, code, message = http.StatusNotFound, "VERSION_NOT_FOUND", "extension version not found"
	case errors.Is(err, extensions.ErrPermissionAck):
		code, message = "PERMISSION_ACK_REQUIRED", "the normalized permission list must be acknowledged exactly"
	case errors.Is(err, extensions.ErrHighRiskAck):
		code, message = "HIGH_RISK_ACK_REQUIRED", "high-risk permissions require explicit acceptance"
	case errors.Is(err, extensions.ErrNotApproved):
		code, message = "VERSION_NOT_APPROVED", "extension version is not approved"
	case errors.Is(err, extensions.ErrProfileNotFound):
		status, code, message = http.StatusNotFound, "PROFILE_NOT_FOUND", "profile not found"
	case errors.Is(err, extensions.ErrRevoked):
		code, message = "VERSION_REVOKED", "extension version is revoked or quarantined"
	case errors.Is(err, extensions.ErrConcurrentMutation):
		status, code, message = http.StatusConflict, "CONCURRENT_MUTATION", "concurrent mutation refused"
	case errors.Is(err, extensions.ErrPurgeNotAllowed):
		code, message = "PURGE_NOT_ALLOWED", "version is still referenced"
	case errors.Is(err, extensions.ErrIntegrity):
		status, code, message = http.StatusInternalServerError, "INTEGRITY_MISMATCH", "managed extension package integrity check failed"
	case errors.Is(err, extensions.ErrInvalidID):
		code, message = "INVALID_EXTENSION_ID", "extension identifier is invalid"
	default:
		status, code, message = http.StatusInternalServerError, "EXTENSION_REPOSITORY_ERROR", "extension repository operation failed"
	}
	writeError(w, status, code, message)
}
