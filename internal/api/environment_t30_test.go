package api

import (
	"context"
	"database/sql"
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
		`"diagnostic_version":"environment-projection-v2"`,
		`"observation_mode":"PROJECTED_METADATA_ONLY"`,
		`"name":"navigator","state":"UNSUPPORTED"`,
		`"name":"rendering-apis","state":"UNSUPPORTED"`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("response missing %q: %s", required, body)
		}
	}
	for _, forbidden := range []string{"private-runtime-hash", "runtime-t30", "127.0.0.1", "canvas", "UserAgent", "binary_hash_sha256"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
}
