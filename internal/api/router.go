package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"camoufoxmulti/internal/browser"
	"camoufoxmulti/internal/config"
	"camoufoxmulti/internal/fingerprint"
	"camoufoxmulti/internal/profile"
)

func NewRouter(cfg *config.Config, store *profile.Store, mgr *browser.Manager, fpPool *fingerprint.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsLocal)

	token := loadOrCreateToken(cfg.DataDir)
	cfg.APIToken = token

	h := &handler{cfg: cfg, store: store, mgr: mgr, token: token, fpPool: fpPool}

	r.Get("/api/status", h.status)
	r.Get("/", h.dashboard)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Post("/api/profiles", h.createProfile)
		r.Get("/api/profiles", h.listProfiles)
		r.Get("/api/profiles/{id}", h.getProfile)
		r.Put("/api/profiles/{id}", h.updateProfile)
		r.Delete("/api/profiles/{id}", h.deleteProfile)
		r.Post("/api/profiles/{id}/duplicate", h.duplicateProfile)
		r.Post("/api/profiles/{id}/export", h.exportProfile)
		r.Post("/api/profiles/import", h.importProfile)

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
		r.Post("/api/backup", h.backup)
		r.Post("/api/restore", h.restore)
		r.Post("/api/shutdown", h.shutdown)
	})

	return r
}

type handler struct {
	cfg    *config.Config
	store  *profile.Store
	mgr    *browser.Manager
	fpPool *fingerprint.Pool
	token  string
}

func (h *handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || auth[7:] != h.token {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": "0.1.0",
		"status":  "ok",
	})
}

func (h *handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var p profile.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if p.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	// Auto-assign fingerprint from pool if not provided
	if p.Fingerprint == nil && h.fpPool != nil {
		engine := p.Engine
		if engine == "" {
			engine = "firefox"
		}
		fp, err := h.fpPool.Pick(engine, "windows")
		if err == nil {
			if p.Proxy != nil && p.Proxy.Host != "" {
				// Has proxy: adjust geo fields based on proxy IP country
				fingerprint.AdjustForCountry(fp, "US") // TODO: real GeoIP lookup
			} else {
				// No proxy: adjust to match local machine's actual location
				fingerprint.AdjustToLocal(fp)
			}
			p.Fingerprint = fp
		}
	}
	if err := h.store.Create(&p); err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": p})
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
	p, err := h.store.Update(chi.URLParam(r, "id"), updates)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
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
	p, err := h.store.Duplicate(chi.URLParam(r, "id"))
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

func corsLocal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "moz-extension://*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loadOrCreateToken(dataDir string) string {
	os.MkdirAll(dataDir, 0755)
	path := dataDir + "/.api-token"
	if data, err := os.ReadFile(path); err == nil {
		return strings.TrimSpace(string(data))
	}
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	os.WriteFile(path, []byte(token), 0600)
	slog.Info("generated API token", "path", path)
	return token
}
