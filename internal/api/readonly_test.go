package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forgelocal/internal/backup"
	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

type t06Catalog struct{}

func (t06Catalog) ListReadOnlyGroups(context.Context) ([]backup.ReadOnlyGroup, error) {
	return []backup.ReadOnlyGroup{{ID: "group-t06", Name: "T06 Group", ProxyMode: "enforced", ProxyConfigured: true, ProfileCount: 1, CreatedAt: "2026-08-15T00:00:00Z", UpdatedAt: "2026-08-15T00:00:00Z"}}, nil
}

func (t06Catalog) ListReadOnlyRuntimeCandidates(context.Context) ([]backup.ReadOnlyRuntimeCandidate, error) {
	return []backup.ReadOnlyRuntimeCandidate{{ID: "runtime-t06", Name: "T06 Runtime", Version: "1.0", Architecture: "amd64", Status: "candidate"}}, nil
}

func TestReadOnlyProfilesAreRedactedAndPaginated(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Alpha", "Bravo", "Charlie"} {
		item := &profile.Profile{Name: name, RuntimeID: "cloakbrowser", Tags: []string{"qa"}, Proxy: &profile.ProxyConfig{Type: "socks5", Host: "private.proxy.invalid", Port: 1080, Region: "private-region"}}
		if err := store.Create(item); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	h := &handler{store: store}
	rec := httptest.NewRecorder()
	h.readonlyProfiles(rec, httptest.NewRequest(http.MethodGet, "/api/v1/readonly/profiles?limit=2", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "private.proxy.invalid") || strings.Contains(rec.Body.String(), "secret_ref") || strings.Contains(rec.Body.String(), "profile_dir") || strings.Contains(rec.Body.String(), "fingerprint") {
		t.Fatalf("redacted response exposed a forbidden field: %s", rec.Body.String())
	}
	var page struct {
		APIVersion string            `json:"api_version"`
		Data       []ReadOnlyProfile `json:"data"`
		Page       struct {
			NextCursor string `json:"next_cursor"`
		} `json:"page"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.APIVersion != readOnlyAPIVersion || len(page.Data) != 2 || page.Page.NextCursor == "" {
		t.Fatalf("unexpected first page: %+v", page)
	}
	rec = httptest.NewRecorder()
	h.readonlyProfiles(rec, httptest.NewRequest(http.MethodGet, "/api/v1/readonly/profiles?limit=2&cursor="+page.Page.NextCursor, nil))
	if rec.Code != http.StatusOK || strings.Count(rec.Body.String(), `"id"`) != 1 {
		t.Fatalf("unexpected second page: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestReadOnlyRuntimeMarksCamoufoxCandidateNonLaunchable(t *testing.T) {
	h := &handler{mgr: testManagerWithRuntimeConfig(t, &config.Config{Runtimes: map[string]config.RuntimeConfig{"camoufox": {BinaryPath: "/private/camoufox"}}})}
	rec := httptest.NewRecorder()
	h.readonlyRuntimes(rec, httptest.NewRequest(http.MethodGet, "/api/v1/readonly/runtimes", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "binary_path") || strings.Contains(rec.Body.String(), "/private/camoufox") {
		t.Fatalf("runtime response leaked binary path: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"candidate":true`) || !strings.Contains(rec.Body.String(), `"launchable":false`) {
		t.Fatalf("Camoufox candidate state missing: %s", rec.Body.String())
	}
}

func TestReadOnlyCatalogGroupsAndRuntimesArePaginatedAndRedacted(t *testing.T) {
	h := &handler{readonlyCatalog: t06Catalog{}}
	for _, path := range []string{"/api/v1/readonly/groups?limit=1", "/api/v1/readonly/runtimes?limit=1"} {
		rec := httptest.NewRecorder()
		if strings.Contains(path, "groups") {
			h.readonlyGroups(rec, httptest.NewRequest(http.MethodGet, path, nil))
		} else {
			h.readonlyRuntimes(rec, httptest.NewRequest(http.MethodGet, path, nil))
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("path=%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
		for _, forbidden := range []string{"secret_ref", "proxy_host", "binary_path", "binary_sha256", "/t06/private", "t06-sentinel"} {
			if strings.Contains(rec.Body.String(), forbidden) {
				t.Fatalf("path=%s leaked %q: %s", path, forbidden, rec.Body.String())
			}
		}
		if !strings.Contains(rec.Body.String(), `"api_version":"v1"`) || !strings.Contains(rec.Body.String(), `"limit":1`) {
			t.Fatalf("unexpected response: %s", rec.Body.String())
		}
	}
}

func TestRequestIDMiddlewareRejectsMalformedInput(t *testing.T) {
	h := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/readonly/health", nil)
	req.Header.Set("X-Request-ID", "bad value")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Request-ID"); !strings.HasPrefix(got, "req-") || got == "bad value" {
		t.Fatalf("request id=%q, want generated safe id", got)
	}
}

func TestCORSAllowsOnlyLoopbackDashboardOrigins(t *testing.T) {
	h := corsLocal(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))

	allowed := httptest.NewRecorder()
	allowedRequest := httptest.NewRequest(http.MethodOptions, "/api/v1/readonly/session/bootstrap", nil)
	allowedRequest.Header.Set("Origin", "http://127.0.0.1:3000")
	h.ServeHTTP(allowed, allowedRequest)
	if allowed.Code != http.StatusNoContent {
		t.Fatalf("loopback preflight status=%d, want %d", allowed.Code, http.StatusNoContent)
	}
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:3000" {
		t.Fatalf("loopback allow origin=%q", got)
	}

	denied := httptest.NewRecorder()
	deniedRequest := httptest.NewRequest(http.MethodOptions, "/api/v1/readonly/session/bootstrap", nil)
	deniedRequest.Header.Set("Origin", "https://dashboard.example.invalid")
	h.ServeHTTP(denied, deniedRequest)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("remote preflight status=%d, want %d", denied.Code, http.StatusForbidden)
	}
	if got := denied.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("remote origin unexpectedly allowed: %q", got)
	}
}

func TestReadOnlyRoutesRequireCoreBearerAndReturnRequestID(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{DataDir: t.TempDir(), Version: "test-core"}
	router, err := NewRouter(cfg, store, &browser.Manager{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	denied := httptest.NewRecorder()
	router.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/api/v1/readonly/health", nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d body=%s", denied.Code, denied.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/readonly/health", nil)
	request.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	request.Header.Set("X-Request-ID", "ui-readonly-0001")
	allowed := httptest.NewRecorder()
	router.ServeHTTP(allowed, request)
	if allowed.Code != http.StatusOK {
		t.Fatalf("authenticated status=%d body=%s", allowed.Code, allowed.Body.String())
	}
	if got := allowed.Header().Get("X-Request-ID"); got != "ui-readonly-0001" {
		t.Fatalf("X-Request-ID=%q, want supplied safe id", got)
	}
}
