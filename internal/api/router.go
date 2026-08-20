package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"forgelocal/internal/cookies"
	"forgelocal/internal/environment"
	"forgelocal/internal/fingerprint"
	"forgelocal/internal/groups"
	"forgelocal/internal/history"
	"forgelocal/internal/humanize"
	"forgelocal/internal/profile"
	"forgelocal/internal/proxies"
	"forgelocal/internal/proxyprovider"
	bfruntime "forgelocal/internal/runtime"
	"forgelocal/internal/templates"
)

func NewRouter(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStores ...*groups.Store) (*chi.Mux, error) {
	var groupStore *groups.Store
	if len(groupStores) > 0 {
		groupStore = groupStores[0]
	}
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, nil, nil, nil, nil)
}

// NewRouterWithReadOnlyCatalog extends only the v1 readonly projections with
// the Core-owned SQLite catalog. Existing business routes remain unchanged.
func NewRouterWithReadOnlyCatalog(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, nil, nil)
}

// NewRouterWithProxyRegistry adds the T10 proxy registry contract (CRUD,
// profile↔proxy assignment) to the existing loopback-only admin group. All
// other routes stay unchanged.
func NewRouterWithProxyRegistry(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, proxyStore, nil)
}

// NewRouterWithQualifiedRegistry wires the T14 runtime qualification registry
// (redacted SQLite-backed qualified catalog) into the router.
func NewRouterWithQualifiedRegistry(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store, registry *bfruntime.QualifiedRegistry) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, proxyStore, registry)
}

// NewRouterWithEnvironment exposes the T13 projected environment diagnostic
// alongside the qualification registry. Existing routes are untouched.
func NewRouterWithEnvironment(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store, registry *bfruntime.QualifiedRegistry) (*chi.Mux, error) {
	return newRouter(cfg, store, mgr, fpPool, backupService, groupStore, catalog, backupDB, proxyStore, registry)
}

func newRouter(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool, backupService *backup.Service, groupStore *groups.Store, catalog backup.ReadOnlyCatalog, backupDB *sql.DB, proxyStore *proxies.Store, registry *bfruntime.QualifiedRegistry) (*chi.Mux, error) {
	templateStore, err := templates.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open template repository: %w", err)
	}
	historyStore, err := history.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open profile history repository: %w", err)
	}
	cookieFixtureStore, err := cookies.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open synthetic cookie fixture repository: %w", err)
	}
	proxyProviderStore, err := proxyprovider.Open(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("open simulated proxy provider catalogue: %w", err)
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(correlationMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(corsLocal)
	r.Use(originGuard)
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

	h := &handler{cfg: cfg, store: store, groupStore: groupStore, mgr: mgr, token: token, fpPool: fpPool, hcfg: hcfg, backupSvc: backupService, readonlyCatalog: catalog, readonlySessions: newReadOnlySessionBroker(), auditSink: newWriteAuditSink(backupDB), backupDB: backupDB, proxyStore: proxyStore, qualifiedRegistry: registry, templateStore: templateStore, historyStore: historyStore, cookieFixtureStore: cookieFixtureStore, proxyProviderStore: proxyProviderStore}
	if err := h.recoverPendingProfileHistory(context.Background()); err != nil {
		return nil, fmt.Errorf("recover pending profile history: %w", err)
	}

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
		r.Post("/api/profiles/bulk", h.bulkProfileOperation)
		r.Post("/api/profiles/{id}/archive", h.archiveProfile)
		r.Post("/api/profiles/{id}/reopen", h.reopenProfile)
		r.Post("/api/profiles/{id}/tags/{tag}", h.addProfileTag)
		r.Delete("/api/profiles/{id}/tags/{tag}", h.removeProfileTag)
		r.Put("/api/profiles/{id}/metadata", h.updateProfileMetadata)
		r.Post("/api/profiles/{id}/cookie-fixtures/import", h.importCookieFixtures)
		r.Get("/api/profiles/{id}/cookie-fixtures/export", h.exportCookieFixtures)
		r.Get("/api/profiles/{id}/history", h.listProfileHistory)
		r.Get("/api/profiles/{id}/history/diff", h.diffProfileHistory)
		r.Get("/api/profiles/{id}/history/{version}", h.getProfileHistoryVersion)
		r.Post("/api/profiles/{id}/history/{version}/restore", h.restoreProfileHistory)
		r.Post("/api/templates", h.createTemplate)
		r.Get("/api/templates", h.listTemplates)
		r.Get("/api/templates/{id}/versions/{version}", h.getTemplateVersion)
		r.Post("/api/templates/{id}/versions", h.createTemplateVersion)
		r.Post("/api/templates/{id}/versions/{version}/archive", h.archiveTemplateVersion)
		r.Post("/api/templates/{id}/versions/{version}/draft", h.calculateTemplateDraft)
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
		r.Post("/api/proxy-providers", h.createProxyProvider)
		r.Get("/api/proxy-providers", h.listProxyProviders)
		r.Post("/api/proxy-providers/{id}/simulate-resolve", h.simulateProxyProviderResolve)

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
		// T13/T14 are projected diagnostic catalogues. They remain admin-only:
		// read-only session tokens are explicitly refused by authMiddleware.
		if h.qualifiedRegistry != nil {
			r.Get("/api/v1/runtimes/qualified", h.listQualifiedRuntimes)
			r.Get("/api/v1/runtimes/qualified/{id}", h.getQualifiedRuntime)
		}
		r.Get("/api/v1/environment/profiles/{id}", h.getEnvironmentDiagnostic)
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
		r.Get("/api/v1/readonly/proxies", h.readonlyProxies)
		r.Get("/api/v1/readonly/groups", h.readonlyGroups)
		r.Get("/api/v1/readonly/runtimes", h.readonlyRuntimes)
	})

	return r, nil
}

type handler struct {
	cfg                *config.Config
	store              *profile.Store
	groupStore         *groups.Store
	mgr                *browser.Manager
	fpPool             *fingerprint.Pool
	hcfg               humanize.Config
	token              string
	backupSvc          *backup.Service
	readonlyCatalog    backup.ReadOnlyCatalog
	auditSink          *writeAuditSink
	readonlySessions   *readOnlySessionBroker
	proxyStore         *proxies.Store
	qualifiedRegistry  *bfruntime.QualifiedRegistry
	backupDB           *sql.DB
	templateStore      *templates.Store
	historyStore       *history.Store
	cookieFixtureStore *cookies.Store
	proxyProviderStore *proxyprovider.Store
}

// t13Checker projects environment diagnostics from stored metadata and the
// runtime qualification registry; it never observes a real browser.
type t13Checker struct {
	db       *sql.DB
	registry *bfruntime.QualifiedRegistry
}

func (c *t13Checker) ProfileExists(ctx context.Context, id string) (bool, error) {
	row := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles WHERE id = ?", id)
	var n int
	if err := row.Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *t13Checker) ProfileRuntimeID(ctx context.Context, id string) (string, error) {
	row := c.db.QueryRowContext(ctx, "SELECT runtime_id FROM profiles WHERE id = ?", id)
	var rt sql.NullString
	if err := row.Scan(&rt); err != nil {
		return "", err
	}
	if !rt.Valid {
		return "", nil
	}
	return rt.String, nil
}

func (c *t13Checker) RuntimeQualified(ctx context.Context, runtimeID string) (bool, error) {
	if c.registry == nil {
		return false, nil
	}
	entry, err := c.registry.Get(ctx, bfruntime.ID(runtimeID))
	if err != nil {
		return false, err
	}
	return entry != nil && entry.State == bfruntime.QSQualified, nil
}

func (c *t13Checker) RuntimeVersion(ctx context.Context, runtimeID string) (string, error) {
	if c.registry == nil {
		return "", nil
	}
	entry, err := c.registry.Get(ctx, bfruntime.ID(runtimeID))
	if err != nil {
		return "", err
	}
	if entry == nil {
		return "", nil
	}
	return strings.TrimSpace(entry.Version), nil
}

// getEnvironmentDiagnostic serves the T13 projected diagnostic. The admin
// group middleware already refused read-only tokens (401); an unknown profile
// is refused explicitly with PROFILE_NOT_FOUND instead of a silent 404 body.
func (h *handler) getEnvironmentDiagnostic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if h.auditSink == nil || h.qualifiedRegistry == nil {
		writeError(w, http.StatusNotFound, "ENVIRONMENT_DIAGNOSTIC_UNAVAILABLE", "diagnostic subsystem disabled")
		return
	}
	diag, err := environment.Diagnose(r.Context(), &t13Checker{db: h.backupDB, registry: h.qualifiedRegistry}, string(bfruntime.ID(id)))
	if errors.Is(err, environment.ErrProfileNotFound) {
		writeError(w, http.StatusNotFound, "ENVIRONMENT_DIAGNOSTIC_NOT_FOUND", "environment: PROFILE_NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ENVIRONMENT_DIAGNOSTIC_ERROR", "diagnostic unavailable")
		return
	}
	writeJSON(w, http.StatusOK, diag)
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

// qualifiedRuntimeProjection is the intentionally minimal T14/G15-A surface.
// Binary paths, debugging ports and integrity hashes are internal-only and
// must never cross the API boundary, even to the local administration UI.
type qualifiedRuntimeProjection struct {
	ID          string                       `json:"id"`
	State       bfruntime.QualificationState `json:"state"`
	Version     string                       `json:"version,omitempty"`
	QualifiedAt any                          `json:"qualified_at,omitempty"`
}

func redactQualifiedRuntime(info bfruntime.QualifiedInfo) qualifiedRuntimeProjection {
	return qualifiedRuntimeProjection{
		ID:          info.ID,
		State:       info.State,
		Version:     strings.TrimSpace(info.Version),
		QualifiedAt: info.QualifiedAt,
	}
}

// listQualifiedRuntimes projects only state, version and qualification time.
// No filesystem paths, debug ports or hashes are exposed to the admin surface.
func (h *handler) listQualifiedRuntimes(w http.ResponseWriter, r *http.Request) {
	if h.qualifiedRegistry == nil {
		writeError(w, http.StatusServiceUnavailable, "QUALIFICATION_UNAVAILABLE", "runtime qualification registry is not initialized")
		return
	}
	list, err := h.qualifiedRegistry.ListQualified(r.Context())
	if err != nil {
		slog.Error("list qualified runtimes", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "qualified catalog unavailable")
		return
	}
	projected := make([]qualifiedRuntimeProjection, 0, len(list))
	for _, info := range list {
		projected = append(projected, redactQualifiedRuntime(info))
	}
	writeJSON(w, http.StatusOK, map[string]any{"runtimes": projected})
}

func (h *handler) getQualifiedRuntime(w http.ResponseWriter, r *http.Request) {
	if h.qualifiedRegistry == nil {
		writeError(w, http.StatusServiceUnavailable, "QUALIFICATION_UNAVAILABLE", "runtime qualification registry is not initialized")
		return
	}
	id := chi.URLParam(r, "id")
	info, err := h.qualifiedRegistry.Get(r.Context(), bfruntime.ID(id))
	if err != nil {
		if errors.Is(err, bfruntime.ErrRuntimeNotQualified) {
			writeError(w, http.StatusNotFound, "RUNTIME_NOT_QUALIFIED", id)
			return
		}
		slog.Error("get qualified runtime", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "qualified catalog unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runtime": redactQualifiedRuntime(*info)})
}

func requireEnabledRuntime(desc bfruntime.Descriptor) error {
	if err := bfruntime.RequireExecution(desc.ID); err != nil {
		return err
	}
	if !desc.Enabled {
		return fmt.Errorf("runtime %q is disabled", desc.ID)
	}
	return nil
}

func writeRuntimeValidationError(w http.ResponseWriter, err error) {
	if errors.Is(err, bfruntime.ErrCamoufoxExecutionNotAuthorized) {
		writeError(w, http.StatusForbidden, bfruntime.CamoufoxExecutionNotAuthorizedCode, "Camoufox execution is not authorized")
		return
	}
	writeError(w, http.StatusBadRequest, "INVALID_RUNTIME", err.Error())
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
	if _, ok := raw["note"]; ok {
		writeError(w, http.StatusBadRequest, "METADATA_ENDPOINT_REQUIRED", "use the profile metadata endpoint")
		return
	}
	if _, ok := raw["custom_fields"]; ok {
		writeError(w, http.StatusBadRequest, "METADATA_ENDPOINT_REQUIRED", "use the profile metadata endpoint")
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
		writeRuntimeValidationError(w, err)
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
	if err := h.captureProfileHistory(r.Context(), &p, "create", correlationIDFrom(r.Context())); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
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
	// T10 — reference-only projection: only the proxy registry id is returned
	// (metadata, no credentials). The vault secret remains behind proxy.ref.*.
	data := map[string]any{"data": p}
	if h.proxyStore != nil {
		if assigned := h.proxyStore.AssignedProxy(p.ID); assigned != nil {
			data["proxy_id"] = assigned.ID
		}
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	unlock, err := h.store.WithHistorySequence(id)
	if err != nil {
		writeProfileError(w, err, correlationIDFrom(r.Context()))
		return
	}
	defer unlock()
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if _, ok := updates["engine"]; ok {
		writeError(w, http.StatusBadRequest, "DEPRECATED_FIELD", "engine was removed in v2; use runtime_id")
		return
	}
	if _, ok := updates["note"]; ok {
		writeError(w, http.StatusBadRequest, "METADATA_ENDPOINT_REQUIRED", "use the profile metadata endpoint")
		return
	}
	if _, ok := updates["custom_fields"]; ok {
		writeError(w, http.StatusBadRequest, "METADATA_ENDPOINT_REQUIRED", "use the profile metadata endpoint")
		return
	}
	if _, runtimeChanged := updates["runtime_id"]; runtimeChanged {
		current, err := h.store.Get(id)
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
			writeRuntimeValidationError(w, err)
			return
		}
		updates["runtime_id"] = draft.RuntimeID
	}
	p, err := h.store.Update(id, updates)
	if err != nil {
		writeProfileError(w, err, correlationIDFrom(r.Context()))
		return
	}
	if err := h.captureProfileHistory(r.Context(), p, "update", correlationIDFrom(r.Context())); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_HISTORY_CAPTURE_FAILED", "profile history capture failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": p})
}

func (h *handler) captureProfileHistory(ctx context.Context, p *profile.Profile, action, correlation string) error {
	if h.historyStore == nil {
		// Existing package-local handler tests construct a narrow handler without
		// NewRouter. Production routers always initialize historyStore; this no-op
		// preserves the old unit harness without weakening the runtime contract.
		return nil
	}
	if p == nil || p.HistoryPending == nil {
		return fmt.Errorf("profile history operation marker is required")
	}
	operationID := p.HistoryPending.OperationID
	if _, err := h.historyStore.Capture(ctx, p, action, correlation); err != nil {
		return err
	}
	return h.store.ClearHistoryPending(p.ID, operationID)
}

func (h *handler) recoverPendingProfileHistory(ctx context.Context) error {
	if h.historyStore == nil {
		return nil
	}
	for _, p := range h.store.PendingHistoryProfiles() {
		operationID := p.HistoryPending.OperationID
		if _, err := h.historyStore.ReconcilePending(ctx, p, "startup-recovery"); err != nil {
			return err
		}
		if err := h.store.ClearHistoryPending(p.ID, operationID); err != nil {
			return err
		}
	}
	return nil
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
		writeRuntimeValidationError(w, err)
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

// originGuard enforces the fail-closed G15-B contract: state-changing requests
// (POST, PUT, DELETE, PATCH) must carry a loopback Origin OR Referer header.
// Cross-origin mutation attempts from the network are rejected before auth.
func originGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		origin := r.Header.Get("Origin")
		if isLoopbackOrigin(origin) {
			next.ServeHTTP(w, r)
			return
		}
		if isLoopbackReferrer(r.Referer()) {
			next.ServeHTTP(w, r)
			return
		}
		writeError(w, http.StatusForbidden, "ORIGIN_REQUIRED_LOCAL_ONLY", "mutations require a loopback Origin or Referer")
	})
}

func isLoopbackReferrer(ref string) bool {
	return ref != "" && isLoopbackOrigin(ref)
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
