package api

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"forgelocal/internal/backup"
	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/fingerprint"
	"forgelocal/internal/groups"
	"forgelocal/internal/humanize"
	"forgelocal/internal/profile"
	"forgelocal/internal/proxies"
	bfruntime "forgelocal/internal/runtime"
)

func NewRouter(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStores ...*groups.Store) (*chi.Mux, error) {
	var groupStore *groups.Store
	if len(groupStores) > 0 {
		groupStore = groupStores[0]
	}
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, nil, nil, nil)
}

// NewRouterWithReadOnlyCatalog extends only the v1 readonly projections with
// the Core-owned SQLite catalog. Existing business routes remain unchanged.
func NewRouterWithReadOnlyCatalog(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, nil)
}

// NewRouterWithProxyRegistry adds the T10 proxy registry contract (CRUD,
// profile↔proxy assignment) to the existing loopback-only admin group. All
// other routes stay unchanged.
func NewRouterWithProxyRegistry(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, proxyStore)
}

func newRouter(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(correlationMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(corsLocal)
	r.Use(requestIDMiddleware)

	token, err := loadOrCreateToken(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	cfg.APIToken = token

	hcfg := humanize.DefaultConfig()
	if cfg.Humanize != nil {
		hcfg = humanize.ConfigFromRaw(cfg.Humanize.Enabled, cfg.Humanize.MouseSpeed, cfg.Humanize.TypingCPM, cfg.Humanize.TypoRate, cfg.Humanize.ScrollStyle)
	}

	h := &handler{cfg: cfg, store: store, groupStore: groupStore, mgr: mgr, token: token, fpPool: fpPool, hcfg: hcfg, backupSvc: backupService, readonlyCatalog: catalog, readonlySessions: newReadOnlySessionBroker(), auditSink: newWriteAuditSink(backupDB), proxyStore: proxyStore}

	r.Get("/api/status", h.status)
	r.Get("/api/health", h.health)
	r.Get("/", h.dashboard)
	r.Post("/api/v1/readonly/session/bootstrap", h.bootstrapReadOnlySession)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Use(h.requireLoopbackMiddleware)

		r.Post("/api/profiles", h.createProfile)
		r.Get("/api/runtimes", h.listRuntimes)
		r.Get("/api/profiles", h.listProfiles)
		r.Get("/api/profiles/{id}", h.getProfile)
		r.Put("/api/profiles/{id}", h.updateProfile)
		r.Delete("/api/profiles/{id}", h.deleteProfile)
		r.Post("/api/profiles/{id}/archive", h.archiveProfile)
		r.Post("/api/profiles/{id}/reopen", h.reopenProfile)
		r.Post("/api/profiles/{id}/tags/{tag}", h.addProfileTag)
		r.Delete("/api/profiles/{id}/tags/{tag}", h.removeProfileTag)
		r.Post("/api/profiles/{id}/duplicate", h.duplicateProfile)
		r.Post("/api/profiles/{id}/export", h.exportProfile)
		r.Post("/api/profiles/import", h.importProfile)
		r.Get("/api/profiles/{id}/artifacts/*", h.artifact)
		r.Get("/api/groups", h.listGroups)
		r.Get("/api/groups/{name}", h.getGroup)
		r.Put("/api/groups/{name}", h.upsertGroup)
		r.Delete("/api/groups/{name}", h.deleteGroup)
		r.Delete("/api/groups/{name}/proxy", h.clearGroupProxy)
		if h.proxyStore != nil {
			r.Get("/api/proxies", h.listProxies)
			r.Post("/api/proxies", h.createProxy)
			r.Get("/api/proxies/{id}", h.getProxy)
			r.Put("/api/proxies/{id}", h.updateProxy)
			r.Post("/api/proxies/{id}/assign", h.assignProxy)
			r.Delete("/api/proxies/{id}/assign", h.unassignProxy)
			r.Delete("/api/proxies/{id}", h.deleteProxy)
		}

		r.Post("/api/sessions", h.createSession)
		r.Get("/api/sessions", h.listSessions)
		r.Delete("/api/sessions/{id}", h.deleteSession)

		r.Post("/api/sessions/{id}/navigate", h.navigate)
		r.Post("/api/sessions/{id}/click", h.click)
		r.Post("/api/sessions/{id}/type", h.typeText)
		r.Post("/api/sessions/{id}/eval", h.evaluate)
		r.Get("/api/sessions/{id}/screenshot", h.screenshot)
		r.Get("/api/sessions/{id}/content", h.content)
		r.Post("/api/sessions/{id}/wait", h.waitFor)
		r.Get("/api/sessions/{id}/cookies", h.getCookies)
		r.Post("/api/sessions/{id}/cookies", h.setCookies)

		r.Get("/api/playwright/endpoint", h.playwrightEndpoint)
		r.Get("/api/playwright/ws/{id}", h.playwrightWSProxy)
		r.Post("/api/v1/readonly/session/codes", h.issueReadOnlySessionCode)
		if h.backupSvc != nil {
			r.Post("/api/v1/profiles/{id}/backups", h.createBackupV1)
			r.Post("/api/v1/backups/{id}/restore", h.restoreBackupV1)
		}

		r.Post("/api/shutdown", h.shutdown)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.readonlyAuthMiddleware)
		r.Get("/api/v1/readonly/health", h.readonlyHealth)
		r.Get("/api/v1/readonly/summary", h.readonlySummary)
		r.Get("/api/v1/readonly/profiles", h.readonlyProfiles)
		r.Get("/api/v1/readonly/groups", h.readonlyGroups)
		r.Get("/api/v1/readonly/runtimes", h.readonlyRuntimes)
	})

	return r, nil
}

type handler struct {
	cfg              *config.Config
	store            *profile.Store
	groupStore       *groups.Store
	mgr              *browser.Manager
	fpPool           *fingerprint.Pool
	hcfg             humanize.Config
	token            string
	backupSvc        *backup.Service
	readonlyCatalog  backup.ReadOnlyCatalog
	auditSink        *writeAuditSink
	readonlySessions *readOnlySessionBroker
	proxyStore       *proxies.Store
}

func (h *handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validBearerToken(r.Header.Get("Authorization"), h.token) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) readonlyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validBearerToken(r.Header.Get("Authorization"), h.token) || h.readonlySessions.validateToken(bearerToken(r.Header.Get("Authorization"))) {
			next.ServeHTTP(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid, expired, or missing read-only session")
	})
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": h.cfg.Version,
		"status":  "ok",
	})
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *handler) listRuntimes(w http.ResponseWriter, r *http.Request) {
	reg := h.mgr.RuntimeRegistry()
	writeJSON(w, http.StatusOK, map[string]any{"data": reg.List(), "default_runtime_id": reg.DefaultID()})
}

func requireEnabledRuntime(desc bfruntime.Descriptor) error {
	if !desc.Enabled {
		return fmt.Errorf("runtime %q is disabled", desc.ID)
	}
	return nil
}

func (h *handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if _, ok := raw["engine"]; ok {
		writeError(w, http.StatusBadRequest, "DEPRECATED_FIELD", "engine was removed in v2; use runtime_id")
		return
	}
	data, err := json.Marshal(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	var p profile.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if p.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	desc, err := h.mgr.RuntimeRegistry().ApplyProfileDefaults(&p)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
		return
	}
	if err := requireEnabledRuntime(desc); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
		return
	}
	if err := h.prepareProfileIdentity(&p, desc); err != nil {
		writeError(w, http.StatusInternalServerError, "PREPARE_PROFILE_FAILED", err.Error())
		return
	}
	if err := h.store.Create(&p); err != nil {
		writeProfileError(w, err, correlationIDFrom(r.Context()))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": p})
}

func (h *handler) prepareProfileIdentity(p *profile.Profile, desc bfruntime.Descriptor) error {
	if p.Fingerprint == nil && h.fpPool != nil && desc.FingerprintPoolKey != "" {
		fp, err := h.fpPool.Pick(desc.FingerprintPoolKey, "windows")
		if err != nil {
			fp, err = h.fpPool.Pick(desc.FingerprintPoolKey, "macos")
		}
		if err == nil {
			effectiveProxy := h.effectiveProxyForProfile(p)
			if effectiveProxy.Proxy != nil {
				proxy := effectiveProxy.Proxy
				tz, locale := fingerprint.DetectProxyGeoResult(proxy.Type, proxy.Host, proxy.Port, proxy.Username, proxy.Password)
				fp["timezone"] = tz
				fp["navigator.language"] = locale
			} else {
				fingerprint.AdjustToLocal(fp)
			}
			p.Fingerprint = fp
		}
	}
	if desc.Capabilities.SupportsSeedFingerprint && p.FingerprintSeed == 0 {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("generate fingerprint seed: %w", err)
		}
		p.FingerprintSeed = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}
	return nil
}

func (h *handler) listProfiles(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	tag := r.URL.Query().Get("tag")
	profiles := h.store.List(group, tag)
	writeJSON(w, http.StatusOK, map[string]any{"data": profiles, "total": len(profiles)})
}

func (h *handler) getProfile(w http.ResponseWriter, r *http.Request) {
	p, err := h.store.Get(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": p})
}

func (h *handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if _, ok := updates["engine"]; ok {
		writeError(w, http.StatusBadRequest, "DEPRECATED_FIELD", "engine was removed in v2; use runtime_id")
		return
	}
	if _, runtimeChanged := updates["runtime_id"]; runtimeChanged {
		current, err := h.store.Get(chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		draft := *current
		v, ok := updates["runtime_id"].(string)
		if !ok {
			writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", "runtime_id must be a string")
			return
		}
		draft.RuntimeID = v
		desc, err := h.mgr.RuntimeRegistry().ApplyProfileDefaults(&draft)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
			return
		}
		if err := requireEnabledRuntime(desc); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
			return
		}
		updates["runtime_id"] = draft.RuntimeID
	}
	p, err := h.store.Update(chi.URLParam(r, "id"), updates)
	if err != nil {
		writeProfileError(w, err, correlationIDFrom(r.Context()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": p})
}

func (h *handler) deleteProfile(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": "deleted"})
}

func (h *handler) duplicateProfile(w http.ResponseWriter, r *http.Request) {
	src, err := h.store.Get(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	desc, err := h.mgr.RuntimeRegistry().ApplyProfileDefaults(src)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
		return
	}
	if err := requireEnabledRuntime(desc); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
		return
	}
	p, err := h.store.Duplicate(src.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": p})
}

// Session and browser operations are in sessions.go

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func validBearerToken(auth, token string) bool {
	if token == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	got := auth[len("Bearer "):]
	return subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}

func bearerToken(auth string) string {
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if !validRequestID(requestID) {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func validRequestID(value string) bool {
	if len(value) < 8 || len(value) > 96 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func newRequestID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "req-unavailable"
	}
	return "req-" + hex.EncodeToString(b)
}

func corsLocal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isLoopbackOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Correlation-ID, X-Request-ID")
		}
		if r.Method == "OPTIONS" {
			if origin != "" && !isLoopbackOrigin(origin) {
				writeError(w, http.StatusForbidden, "CORS_ORIGIN_NOT_ALLOWED", "only loopback dashboard origins are allowed")
				return
			}
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopbackOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return false
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func loadOrCreateToken(dataDir string) (string, error) {
	// Environment variable takes priority
	if envToken := os.Getenv("BROWSEFORGE_TOKEN"); envToken != "" {
		return envToken, nil
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("create token directory: %w", err)
	}
	path := dataDir + "/.api-token"
	if data, err := os.ReadFile(path); err == nil {
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", fmt.Errorf("API token file is empty: %s", path)
		}
		return token, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("read API token: %w", err)
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate API token: %w", err)
	}
	token := hex.EncodeToString(b)
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return "", fmt.Errorf("write API token: %w", err)
	}
	slog.Info("generated API token", "path", path)
	return token, nil
}
