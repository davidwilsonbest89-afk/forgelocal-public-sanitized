// T09 — Profile Writes. Write-side profile contract handlers.
//
// This file adds the lifecycle and tagging endpoints that complete the
// profile write contract:
//
//	POST   /api/profiles/{id}/archive        archive an active profile
//	POST   /api/profiles/{id}/reopen         reopen an archived profile (quarantined stays refused)
//	POST   /api/profiles/{id}/tags/{tag}     assign a tag
//	DELETE /api/profiles/{id}/tags/{tag}     unassign a tag
//
// Every mutation records a redacted audit event joined to the request
// correlation id. Handlers never write SQLite through the dashboard: the
// dashboard calls this API only. Errors use explicit machine-readable codes.
//
// Style contract (project): redaction at the boundary; no absolute paths,
// vault references or credential values in responses or audit payloads.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/profile"
)

const maxProfileMetadataBytes = 16 << 10

type profileMetadataRequest struct {
	Note         string                         `json:"note"`
	CustomFields map[string]profile.CustomField `json:"custom_fields"`
}

// updateProfileMetadata is the audited Core-only mutation for non-sensitive
// Notes and Custom Fields. Values are returned to the authenticated caller but
// never copied into audit payloads, which carry only shape metadata.
func (h *handler) updateProfileMetadata(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationID)
		return
	}
	defer unlock()
	r.Body = http.MaxBytesReader(w, r.Body, maxProfileMetadataBytes)
	defer r.Body.Close()
	var req profileMetadataRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.metadata_failed", id, correlationID, map[string]any{"error": "INVALID_PROFILE"})
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROFILE", "profile metadata is invalid", correlationID)
		return
	}
	p, err := h.store.SetMetadata(id, req.Note, req.CustomFields)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.metadata_failed", id, correlationID, map[string]any{"error": mapErrorCode(err)})
		writeProfileError(w, err, correlationID)
		return
	}
	if err := h.captureProfileHistory(r.Context(), p, "metadata", correlationID); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.metadata_updated", id, correlationID, map[string]any{"has_note": req.Note != "", "custom_field_count": len(req.CustomFields)})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": p.ID, "note": p.Note, "custom_fields": p.CustomFields}})
}

// profileMutationError writes an explicit error response with the request's
// correlation id attached, so failing operations stay traceable in the ledger.
func profileMutationError(w http.ResponseWriter, status int, code, message string, correlationID string) {
	w.Header().Set(correlationHeader, correlationID)
	writeError(w, status, code, message)
}

func (h *handler) archiveProfile(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationID)
		return
	}
	defer unlock()
	p, changed, err := h.store.ArchiveProfileResult(id)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.archive_failed", id, correlationID, map[string]any{"error": mapErrorCode(err)})
		writeProfileError(w, err, correlationID)
		return
	}
	if !changed {
		w.Header().Set(correlationHeader, correlationID)
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "lifecycle_state": "archived", "changed": false}})
		return
	}
	if err := h.captureProfileHistory(r.Context(), p, "archive", correlationID); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.archived", id, correlationID, map[string]any{"state": "archived"})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "lifecycle_state": "archived"}})
}

func (h *handler) reopenProfile(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationID)
		return
	}
	defer unlock()
	p, err := h.store.ReopenProfileResult(id)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.reopen_failed", id, correlationID, map[string]any{"error": mapErrorCode(err)})
		writeProfileError(w, err, correlationID)
		return
	}
	if err := h.captureProfileHistory(r.Context(), p, "reopen", correlationID); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.reopened", id, correlationID, map[string]any{"state": "active"})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "lifecycle_state": "active"}})
}

func (h *handler) addProfileTag(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag := chi.URLParam(r, "tag")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationID)
		return
	}
	defer unlock()
	if err := h.store.AddProfileTag(id, tag); err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.tag_failed", id, correlationID, map[string]any{"tag": tag, "error": mapErrorCode(err)})
		writeProfileError(w, err, correlationID)
		return
	}
	if p, err := h.store.Get(id); err != nil || h.captureProfileHistory(r.Context(), p, "tag_add", correlationID) != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.tag_added", id, correlationID, map[string]any{"tag": tag})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "tag": tag}})
}

func (h *handler) removeProfileTag(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	tag := chi.URLParam(r, "tag")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationID)
		return
	}
	defer unlock()
	if err := h.store.RemoveProfileTag(id, tag); err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.tag_failed", id, correlationID, map[string]any{"tag": tag, "error": mapErrorCode(err)})
		writeProfileError(w, err, correlationID)
		return
	}
	if p, err := h.store.Get(id); err != nil || h.captureProfileHistory(r.Context(), p, "tag_remove", correlationID) != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.tag_removed", id, correlationID, map[string]any{"tag": tag})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "tag": tag}})
}

// writeProfileError maps sentinel errors to explicit machine-readable codes
// without leaking internal messages.
func writeProfileError(w http.ResponseWriter, err error, correlationID string) {
	w.Header().Set(correlationHeader, correlationID)
	code, status := mapErrorCodeToResponse(err)
	// The body deliberately never carries err.Error(): internal messages may
	// leak paths, vault references or credential details.
	writeError(w, status, code, redactedMessageFor(err))
}

// redactedMessageFor returns a short, user-facing explanation without leaking
// internal state. Codes stay machine-readable; messages stay redacted.
func redactedMessageFor(err error) string {
	switch {
	case profile.IsNotFoundError(err):
		return "the profile does not exist"
	case profile.IsDuplicateError(err):
		return "a profile with this name already exists"
	case profile.IsLockedError(err):
		return "the profile is currently locked by another operation"
	case profile.IsLifecycleError(err):
		return "the requested state change is not allowed for this profile"
	case profile.IsValidationError(err):
		return "the profile data is invalid"
	default:
		return "the profile operation failed"
	}
}

func mapErrorCode(err error) string {
	code, _ := mapErrorCodeToResponse(err)
	return code
}

func mapErrorCodeToResponse(err error) (string, int) {
	switch {
	case err == nil:
		return "OK", http.StatusOK
	case profile.IsNotFoundError(err):
		return "PROFILE_NOT_FOUND", http.StatusNotFound
	case profile.IsLockedError(err):
		return "PROFILE_LOCKED", http.StatusConflict
	case profile.IsLifecycleError(err):
		return "INVALID_LIFECYCLE", http.StatusConflict
	case profile.IsDuplicateError(err):
		return "DUPLICATE_PROFILE", http.StatusConflict
	case profile.IsValidationError(err):
		return "INVALID_PROFILE", http.StatusBadRequest
	default:
		return "PROFILE_MUTATION_FAILED", http.StatusInternalServerError
	}
}
