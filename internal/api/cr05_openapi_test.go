package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

func TestCR05V1FacadeAndOpenAPIMatchLivingRouter(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{DataDir: t.TempDir(), Version: "cr05-test"}
	router, err := NewRouter(cfg, store, &browser.Manager{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	openAPI := httptest.NewRecorder()
	router.ServeHTTP(openAPI, httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil))
	if openAPI.Code != http.StatusOK {
		t.Fatalf("openapi status=%d body=%s", openAPI.Code, openAPI.Body.String())
	}
	var spec struct {
		OpenAPI string         `json:"openapi"`
		Paths   map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(openAPI.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.OpenAPI != "3.1.0" || spec.Paths["/api/v1/profiles"] == nil {
		t.Fatalf("invalid v1 OpenAPI index: %#v", spec)
	}

	for _, path := range []string{"/api/profiles", "/api/v1/profiles"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = "127.0.0.1:7611"
		req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}

	unauth := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	unauth.RemoteAddr = "127.0.0.1:7612"
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauth)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated v1 alias status=%d", unauthRec.Code)
	}

	missing := httptest.NewRequest(http.MethodGet, "/api/v1/not-a-route", nil)
	missing.RemoteAddr = "127.0.0.1:7613"
	missing.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	missingRec := httptest.NewRecorder()
	router.ServeHTTP(missingRec, missing)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("unknown v1 route status=%d", missingRec.Code)
	}
}
