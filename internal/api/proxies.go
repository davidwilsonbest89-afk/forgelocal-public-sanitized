// T10 — Proxies. Proxy registry contract handlers.
//
// This file adds the endpoints that complete the proxy write contract:
//
//	GET    /api/proxies               list registered proxies (redacted)
//	POST   /api/proxies               create a proxy
//	GET    /api/proxies/{id}          get a proxy (redacted)
//	PUT    /api/proxies/{id}          update a proxy
//	DELETE /api/proxies/{id}          delete a proxy
//	POST   /api/proxies/{id}/assign?profile_id=    bind a profile
//	DELETE /api/proxies/{id}/assign?profile_id=    unbind a profile
//
// Every mutation records a redacted audit event joined to the request
// correlation id. Handlers never write SQLite through the dashboard: the
// dashboard calls this API only. Errors use explicit machine-readable codes.
//
// Style contract (project): redaction at the boundary; no absolute paths,
// vault references beyond the opaque secret_ref, or credential values in
// responses or audit payloads.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/proxies"
)

// proxyError maps registry sentinel errors to explicit machine-readable codes
// without leaking internal messages.
func proxyError(w http.ResponseWriter, err error, correlationID string) {
	w.Header().Set(correlationHeader, correlationID)
	code, status := mapProxyErrorCode(err)
	writeError(w, status, code, redactedMessageForProxy(err))
}

// redactedMessageForProxy returns a short, user-facing explanation without
// leaking internal state. Codes stay machine-readable; messages stay redacted.
func redactedMessageForProxy(err error) string {
	switch {
	case proxies.IsNotFoundError(err):
		return "the proxy does not exist"
	case proxies.IsDuplicateError(err):
		return "the proxy name is already in use or the proxy is assigned to profiles"
	case proxies.IsLockedError(err):
		return "the proxy is currently locked by another operation"
	case proxies.IsValidationError(err):
		return "the proxy data is invalid"
	default:
		return "the proxy operation failed"
	}
}

// mapProxyErrorCode maps registry sentinel errors to API error codes.
func mapProxyErrorCode(err error) (string, int) {
	switch {
	case err == nil:
		return "OK", http.StatusOK
	case proxies.IsNotFoundError(err):
		return "PROXY_NOT_FOUND", http.StatusNotFound
	case proxies.IsLockedError(err):
		return "PROXY_LOCKED", http.StatusConflict
	case proxies.IsDuplicateError(err):
		return "PROXY_NAME_TAKEN", http.StatusConflict
	case proxies.IsValidationError(err):
		return "INVALID_PROXY", http.StatusBadRequest
	default:
		return "PROXY_MUTATION_FAILED", http.StatusInternalServerError
	}
}

func (h *handler) listProxies(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	h.auditSink.auditRecord(r.Context(), "proxy.list", "", correlationID, map[string]any{"count": len(h.proxyStore.List())})
	list := h.proxyStore.List()
	data := make([]map[string]any, 0, len(list))
	for _, p := range list {
		data = append(data, map[string]any{
			"id":          p.ID,
			"name":        p.Name,
			"type":        p.Type,
			"host":        p.Host,
			"port":        p.Port,
			"region":      p.Region,
			"secret_ref":  p.SecretRef,
			"has_secret":  p.HasSecret,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		})
	}
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": data}})
}

func (h *handler) getProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	p, err := h.proxyStore.Get(id)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.get_failed", id, correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy.get", id, correlationID, nil)
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": proxyView(p)})
}

func (h *handler) createProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Region   string `json:"region,omitempty"`
		SecretRef string `json:"secret_ref,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.create_failed", "", correlationID, map[string]any{"error": "INVALID_PROXY"})
		proxyError(w, proxies.ErrInvalidProxy, correlationID)
		return
	}
	p := &proxies.Proxy{
		Name:      req.Name,
		Type:      req.Type,
		Host:      req.Host,
		Port:      req.Port,
		Region:    req.Region,
		SecretRef: req.SecretRef,
		HasSecret: req.SecretRef != "",
	}
	if err := h.proxyStore.Create(p); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.create_failed", "", correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	// The response exposes secret_ref only; no credential material is ever
	// returned, logged or audited.
	h.auditSink.auditRecord(r.Context(), "proxy.created", p.ID, correlationID, map[string]any{
		"name":       p.Name,
		"type":       p.Type,
		"has_secret": p.HasSecret,
	})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusCreated, map[string]any{"data": proxyView(p)})
}

func (h *handler) updateProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.update_failed", id, correlationID, map[string]any{"error": "INVALID_PROXY"})
		proxyError(w, proxies.ErrInvalidProxy, correlationID)
		return
	}
	// The dashboard may never overwrite the credential reference directly:
	// secret_ref changes belong to the system vault boundary only.
	delete(req, "secret_ref")
	delete(req, "id")
	p, err := h.proxyStore.Update(id, req)
	if err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.update_failed", id, correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy.updated", id, correlationID, map[string]any{
		"name":       p.Name,
		"type":       p.Type,
		"has_secret": p.HasSecret,
	})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": proxyView(p)})
}

func (h *handler) deleteProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.proxyStore.Delete(id); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.delete_failed", id, correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy.deleted", id, correlationID, nil)
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id}})
}

func (h *handler) assignProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	proxyID := chi.URLParam(r, "id")
	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" {
		proxyError(w, proxies.ErrInvalidProxy, correlationID)
		return
	}
	if err := h.proxyStore.Assign(profileID, proxyID); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.assign_failed", proxyID, correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy.assigned", proxyID, correlationID, map[string]any{"profile_id": profileID})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"proxy_id":   proxyID,
		"profile_id": profileID,
	}})
}

func (h *handler) unassignProxy(w http.ResponseWriter, r *http.Request) {
	correlationID := correlationIDFrom(r.Context())
	proxyID := chi.URLParam(r, "id")
	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" {
		proxyError(w, proxies.ErrInvalidProxy, correlationID)
		return
	}
	if err := h.proxyStore.UnassignFor(profileID, proxyID); err != nil {
		h.auditSink.auditRecord(r.Context(), "proxy.unassign_failed", proxyID, correlationID, map[string]any{"error": mapProxyErrorCodeStr(err)})
		proxyError(w, err, correlationID)
		return
	}
	h.auditSink.auditRecord(r.Context(), "proxy.unassigned", proxyID, correlationID, map[string]any{"profile_id": profileID})
	w.Header().Set(correlationHeader, correlationID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"proxy_id":   proxyID,
		"profile_id": profileID,
	}})
}

// proxyView is the single redacted projection used by every proxy response.
func proxyView(p *proxies.Proxy) map[string]any {
	return map[string]any{
		"id":          p.ID,
		"name":        p.Name,
		"type":        p.Type,
		"host":        p.Host,
		"port":        p.Port,
		"region":      p.Region,
		"secret_ref":  p.SecretRef,
		"has_secret":  p.HasSecret,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}

// mapProxyErrorCodeStr returns only the code string for audit payloads.
func mapProxyErrorCodeStr(err error) string {
	code, _ := mapProxyErrorCode(err)
	return code
}
