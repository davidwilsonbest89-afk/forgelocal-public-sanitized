package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/history"
	"forgelocal/internal/profile"
)

type historyRestoreRequest struct {
	ExpectedCurrentVersion int `json:"expected_current_version"`
}

func (h *handler) listProfileHistory(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := historyPagination(w, r)
	if !ok { return }
	result, err := h.historyStore.List(r.Context(), chi.URLParam(r, "id"), limit, offset)
	if err != nil { writeHistoryError(w, err, correlationIDFrom(r.Context())); return }
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) getProfileHistoryVersion(w http.ResponseWriter, r *http.Request) {
	version, ok := historyVersionParam(w, r)
	if !ok { return }
	item, err := h.historyStore.Get(r.Context(), chi.URLParam(r, "id"), version)
	if err != nil { writeHistoryError(w, err, correlationIDFrom(r.Context())); return }
	writeJSON(w, http.StatusOK, map[string]any{"data": redactHistoryVersion(*item)})
}

func (h *handler) diffProfileHistory(w http.ResponseWriter, r *http.Request) {
	from, errFrom := strconv.Atoi(r.URL.Query().Get("from"))
	to, errTo := strconv.Atoi(r.URL.Query().Get("to"))
	if errFrom != nil || errTo != nil || from < 1 || to < 1 {
		writeHistoryError(w, history.ErrInvalidVersion, correlationIDFrom(r.Context()))
		return
	}
	result, err := h.historyStore.Diff(r.Context(), chi.URLParam(r, "id"), from, to)
	if err != nil { writeHistoryError(w, err, correlationIDFrom(r.Context())); return }
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *handler) restoreProfileHistory(w http.ResponseWriter, r *http.Request) {
	version, ok := historyVersionParam(w, r)
	if !ok { return }
	correlation := correlationIDFrom(r.Context())
	var req historyRestoreRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil || req.ExpectedCurrentVersion < 1 {
		writeHistoryError(w, history.ErrInvalidVersion, correlation)
		return
	}
	id := chi.URLParam(r, "id")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeHistoryError(w, err, correlation)
		return
	}
	defer unlock()
	operationID := ""
	entry, err := h.historyStore.Restore(r.Context(), id, version, req.ExpectedCurrentVersion, correlation, func(snapshot *profile.Profile) (*profile.Profile, error) {
		restored, err := h.store.RestoreHistory(id, snapshot)
		if err == nil && restored.HistoryPending != nil {
			operationID = restored.HistoryPending.OperationID
		}
		return restored, err
	})
	if err != nil { writeHistoryError(w, err, correlation); return }
	if operationID == "" {
		writeHistoryError(w, profile.ErrInvalidName, correlation)
		return
	}
	if err := h.store.ClearHistoryPending(id, operationID); err != nil {
		writeHistoryError(w, err, correlation)
		return
	}
	w.Header().Set(correlationHeader, correlation)
	writeJSON(w, http.StatusOK, map[string]any{"data": entry})
}

func historyPagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, offset := 50, 0
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 { writeHistoryError(w, history.ErrInvalidVersion, correlationIDFrom(r.Context())); return 0, 0, false }
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 { writeHistoryError(w, history.ErrInvalidVersion, correlationIDFrom(r.Context())); return 0, 0, false }
	}
	return limit, offset, true
}

func historyVersionParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil || version < 1 { writeHistoryError(w, history.ErrInvalidVersion, correlationIDFrom(r.Context())); return 0, false }
	return version, true
}

func redactHistoryProfile(p profile.Profile) profile.Profile {
	if p.Proxy != nil {
		copy := *p.Proxy
		copy.Username, copy.Password, copy.SecretRef = "", "", ""
		p.Proxy = &copy
	}
	p.ProfileDir = ""
	return p
}

func redactHistoryVersion(item history.Version) map[string]any {
	data, _ := json.Marshal(redactHistoryProfile(item.Snapshot.Profile))
	var projected map[string]any
	_ = json.Unmarshal(data, &projected)
	delete(projected, "profile_dir")
	if proxy, ok := projected["proxy"].(map[string]any); ok {
		delete(proxy, "username")
		delete(proxy, "password")
		delete(proxy, "secret_ref")
	}
	return map[string]any{
		"profile_id": item.ProfileID,
		"version": item.Version,
		"action": item.Action,
		"created_at": item.CreatedAt,
		"snapshot": map[string]any{"profile": projected},
	}
}

func writeHistoryError(w http.ResponseWriter, err error, correlation string) {
	w.Header().Set(correlationHeader, correlation)
	status, code := http.StatusBadRequest, "PROFILE_HISTORY_INVALID"
	switch {
	case errors.Is(err, history.ErrNotFound):
		status, code = http.StatusNotFound, "PROFILE_HISTORY_NOT_FOUND"
	case errors.Is(err, history.ErrVersionConflict):
		status, code = http.StatusConflict, "PROFILE_HISTORY_VERSION_CONFLICT"
	case profile.IsNotFoundError(err):
		status, code = http.StatusNotFound, "PROFILE_NOT_FOUND"
	case profile.IsLockedError(err):
		status, code = http.StatusConflict, "PROFILE_LOCKED"
	case profile.IsLifecycleError(err):
		status, code = http.StatusConflict, "INVALID_LIFECYCLE"
	}
	writeError(w, status, code, "profile history request rejected")
}
