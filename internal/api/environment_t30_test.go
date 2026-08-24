package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	bfruntime "forgelocal/internal/runtime"
	_ "modernc.org/sqlite"
)

func TestT30EnvironmentHTTPResponseIsVersionedAndRedacted(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE profiles (id TEXT PRIMARY KEY, runtime_id TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO profiles (id, runtime_id) VALUES (?, ?)`, "profile-t30", "runtime-t30"); err != nil {
		t.Fatal(err)
	}
	qualifier := bfruntime.NewQualifier(db)
	if err := qualifier.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO runtime_qualifications (runtime_id, state, version, binary_hash_sha256, qualified_at, failed_reason, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, "runtime-t30", string(bfruntime.QSQualified), "126.0", "private-runtime-hash", "2026-08-20T00:00:00Z", "", "2026-08-20T00:00:00Z"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{APIToken: "t30-test-token"}
	h := &handler{
		cfg:               cfg,
		mgr:               &browser.Manager{},
		token:             cfg.APIToken,
		auditSink:         &writeAuditSink{db: db},
		backupDB:          db,
		qualifiedRegistry: bfruntime.NewQualifiedRegistry(db),
	}
	router := chi.NewRouter()
	router.Use(correlationMiddleware)
	router.Use(originGuard)
	router.Use(h.authMiddleware)
	router.Use(h.requireLoopbackMiddleware)
	router.Get("/api/v1/environment/profiles/{id}", h.getEnvironmentDiagnostic)

	denied := httptest.NewRecorder()
	router.ServeHTTP(denied, newLoopbackRequest(http.MethodGet, "/api/v1/environment/profiles/profile-t30", nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d body=%s", denied.Code, denied.Body.String())
	}

	req := newLoopbackRequest(http.MethodGet, "/api/v1/environment/profiles/profile-t30", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	allowed := httptest.NewRecorder()
	router.ServeHTTP(allowed, req)
	if allowed.Code != http.StatusOK {
		t.Fatalf("authenticated status=%d body=%s", allowed.Code, allowed.Body.String())
	}
	body := allowed.Body.String()
	for _, required := range []string{
		`"diagnostic_version":"environment-projection-v3"`,
		`"observation_mode":"PROJECTED_METADATA_ONLY"`,
		`"name":"navigator","state":"UNSUPPORTED"`,
		`"name":"rendering-apis","state":"UNSUPPORTED"`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("response missing %q: %s", required, body)
		}
	}
	for _, forbidden := range []string{"private-runtime-hash", "runtime-t30", "127.0.0.1", "canvas_value", "canvas fingerprint", "UserAgent", "binary_hash_sha256"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
}

func TestT30EnvironmentOpenAPIContractIsVersionedAndRedacted(t *testing.T) {
	h := &handler{cfg: &config.Config{Version: "t30-openapi-test"}}
	rec := httptest.NewRecorder()
	h.openAPIV1(rec, httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status=%d body=%s", rec.Code, rec.Body.String())
	}
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	path, ok := spec.Paths["/api/v1/environment/profiles/{id}"]
	if !ok {
		t.Fatalf("T30 environment route absent from OpenAPI: %s", rec.Body.String())
	}
	contract := string(path)
	for _, required := range []string{"diagnostic_version", "environment-projection-v3", "observation_mode", "PROJECTED_METADATA_ONLY", "UNSUPPORTED", "401", "404"} {
		if !strings.Contains(contract, required) {
			t.Fatalf("OpenAPI route contract missing %q: %s", required, contract)
		}
	}
	for _, forbidden := range []string{"binary_hash_sha256", "user_agent", "canvas_value", "127.0.0.1"} {
		if strings.Contains(contract, forbidden) {
			t.Fatalf("OpenAPI route contract must stay redacted, found %q: %s", forbidden, contract)
		}
	}
}
