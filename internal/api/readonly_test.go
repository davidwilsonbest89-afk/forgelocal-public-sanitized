package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

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
