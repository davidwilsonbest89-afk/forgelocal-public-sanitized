// T25 — local synthetic cookie fixture manifests. These endpoints deliberately
// do not call session, browser, Playwright, MCP or filesystem cookie APIs.
package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/cookies"
	"forgelocal/internal/profile"
)

const maxCookieFixtureRequestBytes = 32 << 10

type cookieFixtureImportRequest struct {
	Fixtures []cookies.Input `json:"fixtures"`
}

func (h *handler) importCookieFixtures(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.cookieFixturesAvailable(w, r, id) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxCookieFixtureRequestBytes)
	defer r.Body.Close()
	var req cookieFixtureImportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		profileMutationError(w, http.StatusBadRequest, "INVALID_COOKIE_FIXTURES", "synthetic cookie fixtures are invalid", correlationIDFrom(r.Context()))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		profileMutationError(w, http.StatusBadRequest, "INVALID_COOKIE_FIXTURES", "synthetic cookie fixtures are invalid", correlationIDFrom(r.Context()))
		return
	}
	fixtures, err := h.cookieFixtureStore.Replace(id, req.Fixtures)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "profile.cookie_fixtures_failed", id, correlationIDFrom(r.Context()), map[string]any{"error": "INVALID_COOKIE_FIXTURES"})
		profileMutationError(w, http.StatusBadRequest, "INVALID_COOKIE_FIXTURES", "synthetic cookie fixtures are invalid", correlationIDFrom(r.Context()))
		return
	}
	h.auditSink.auditRecord(r.Context(), "profile.cookie_fixtures_imported", id, correlationIDFrom(r.Context()), map[string]any{"count": len(fixtures)})
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"count": len(fixtures), "fixtures": fixtures}})
}

func (h *handler) exportCookieFixtures(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.cookieFixturesAvailable(w, r, id) {
		return
	}
	fixtures := h.cookieFixtureStore.Export(id)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"count": len(fixtures), "fixtures": fixtures}})
}

func (h *handler) cookieFixturesAvailable(w http.ResponseWriter, r *http.Request, id string) bool {
	if h.cookieFixtureStore == nil {
		writeError(w, http.StatusServiceUnavailable, "COOKIE_FIXTURES_UNAVAILABLE", "synthetic cookie fixture store is unavailable")
		return false
	}
	p, err := h.store.Get(id)
	if err != nil {
		profileMutationError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "profile not found", correlationIDFrom(r.Context()))
		return false
	}
	if p.LifecycleState != profile.LifecycleActive {
		profileMutationError(w, http.StatusConflict, "INVALID_LIFECYCLE", "profile must be active", correlationIDFrom(r.Context()))
		return false
	}
	return true
}
