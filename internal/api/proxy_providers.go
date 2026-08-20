// T26 — simulated, local-only provider catalogue. No network or runtime proxy
// mutation is reachable from these handlers.
package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/profile"
	"forgelocal/internal/proxyprovider"
)

const maxProxyProviderRequestBytes = 16 << 10

func (h *handler) createProxyProvider(w http.ResponseWriter, r *http.Request) {
	if h.proxyProviderStore == nil {
		writeError(w, http.StatusServiceUnavailable, "PROXY_PROVIDER_UNAVAILABLE", "simulated provider catalogue unavailable")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProxyProviderRequestBytes)
	defer r.Body.Close()
	var input proxyprovider.Provider
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROXY_PROVIDER", "provider is invalid", correlationIDFrom(r.Context()))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROXY_PROVIDER", "provider is invalid", correlationIDFrom(r.Context()))
		return
	}
	provider, err := h.proxyProviderStore.Register(input)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy_provider.register_failed", input.ID, correlationIDFrom(r.Context()), map[string]any{"error": "INVALID_PROXY_PROVIDER"})
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROXY_PROVIDER", "provider is invalid", correlationIDFrom(r.Context()))
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy_provider.registered", provider.ID, correlationIDFrom(r.Context()), map[string]any{"mode": "simulated"})
	writeJSON(w, http.StatusCreated, map[string]any{"data": provider})
}

func (h *handler) listProxyProviders(w http.ResponseWriter, r *http.Request) {
	if h.proxyProviderStore == nil {
		writeError(w, http.StatusServiceUnavailable, "PROXY_PROVIDER_UNAVAILABLE", "simulated provider catalogue unavailable")
		return
	}
	providers := h.proxyProviderStore.List()
	writeJSON(w, http.StatusOK, map[string]any{"data": providers, "total": len(providers)})
}

func (h *handler) simulateProxyProviderResolve(w http.ResponseWriter, r *http.Request) {
	if h.proxyProviderStore == nil {
		writeError(w, http.StatusServiceUnavailable, "PROXY_PROVIDER_UNAVAILABLE", "simulated provider catalogue unavailable")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProxyProviderRequestBytes)
	defer r.Body.Close()
	var req struct {
		ProfileID string `json:"profile_id"`
		Region    string `json:"region"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROVIDER_RESOLUTION", "provider request is invalid", correlationIDFrom(r.Context()))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROVIDER_RESOLUTION", "provider request is invalid", correlationIDFrom(r.Context()))
		return
	}
	p, err := h.store.Get(req.ProfileID)
	if err != nil {
		profileMutationError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "profile not found", correlationIDFrom(r.Context()))
		return
	}
	if p.LifecycleState != profile.LifecycleActive {
		profileMutationError(w, http.StatusConflict, "INVALID_LIFECYCLE", "profile must be active", correlationIDFrom(r.Context()))
		return
	}
	lease, err := h.proxyProviderStore.SimulateResolve(chi.URLParam(r, "id"), req.ProfileID, req.Region)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy_provider.resolve_failed", chi.URLParam(r, "id"), correlationIDFrom(r.Context()), map[string]any{"error": "INVALID_PROVIDER_RESOLUTION"})
		profileMutationError(w, http.StatusBadRequest, "INVALID_PROVIDER_RESOLUTION", "provider request is invalid", correlationIDFrom(r.Context()))
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy_provider.resolved", lease.ProviderID, correlationIDFrom(r.Context()), map[string]any{"mode": "simulated", "profile_id": lease.ProfileID})
	writeJSON(w, http.StatusOK, map[string]any{"data": lease})
}
