package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"

	_ "modernc.org/sqlite"
)

// T21-TEST-EVIDENCE-CORRECTION: these tests exercise existing Templates routes
// and audit schemas. They deliberately add no route, model, or persistence behaviour.
func templateRouterWithProfileAudit(t *testing.T) (*config.Config, http.Handler, *profile.Store, *sql.DB, string) {
	t.Helper()
	root := t.TempDir()
	profiles, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", ":memory:?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := createAuditTable(db); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{DataDir: filepath.Join(root, "data"), ProfilesDir: filepath.Join(root, "profiles"), Version: "t21-correction-test"}
	router, err := newRouter(cfg, profiles, &browser.Manager{}, nil, nil, nil, nil, db, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, router, profiles, db, root
}

func templateRequestWithHeaders(t *testing.T, router http.Handler, cfg *config.Config, method, path string, body any, token, origin, referer string) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.RemoteAddr = "127.0.0.1:7777"
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func TestT21TemplateAPIRejectsUnknownFieldsAndInvalidMetadata(t *testing.T) {
	cfg, router, _ := templateRouter(t)
	validHeaders := []string{"http://127.0.0.1:3100"}
	cases := []struct {
		name string
		body any
	}{
		{
			name: "unknown top level field",
			body: map[string]any{"name": "unknown root", "content": map[string]any{}, "runtime_id": "forbidden"},
		},
		{
			name: "unknown content field",
			body: map[string]any{"name": "unknown content", "content": map[string]any{"runtime_id": "forbidden"}},
		},
		{
			name: "invalid typed custom field",
			body: map[string]any{"name": "bad typed field", "content": map[string]any{"custom_fields": map[string]any{"field": map[string]any{"type": "boolean", "value": "not-a-boolean"}}}},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			res := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates", testCase.body, cfg.APIToken, validHeaders[0], "")
			if res.Code != http.StatusBadRequest || !strings.Contains(res.Body.String(), "INVALID_TEMPLATE") {
				t.Fatalf("invalid input must fail closed: status=%d body=%s", res.Code, res.Body.String())
			}
			if res.Header().Get(correlationHeader) == "" {
				t.Fatal("rejected mutation must retain correlation id")
			}
		})
	}
}

func TestT21TemplateAPIMachineReadableAuthAndOriginRefusals(t *testing.T) {
	cfg, router, _ := templateRouter(t)
	body := map[string]any{"name": "refusal", "content": map[string]any{}}
	cases := []struct {
		name          string
		token, origin string
		referer       string
		status        int
		code          string
	}{
		{"authentication absent", "", "http://127.0.0.1:3100", "", http.StatusUnauthorized, "UNAUTHORIZED"},
		{"origin and referer absent", cfg.APIToken, "", "", http.StatusForbidden, "ORIGIN_REQUIRED_LOCAL_ONLY"},
		{"origin distant", cfg.APIToken, "https://remote.invalid", "", http.StatusForbidden, "ORIGIN_REQUIRED_LOCAL_ONLY"},
		{"referer distant", cfg.APIToken, "", "https://remote.invalid/path", http.StatusForbidden, "ORIGIN_REQUIRED_LOCAL_ONLY"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			res := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates", body, testCase.token, testCase.origin, testCase.referer)
			if res.Code != testCase.status || !strings.Contains(res.Body.String(), testCase.code) {
				t.Fatalf("refusal must be machine-readable: status=%d body=%s", res.Code, res.Body.String())
			}
			if res.Header().Get(correlationHeader) == "" {
				t.Fatal("refusal must retain correlation id")
			}
		})
	}
}

func TestT21TemplateAPIPaginationProjectionAndCorrelation(t *testing.T) {
	cfg, router, _ := templateRouter(t)
	for _, name := range []string{"catalog one", "catalog two", "catalog three"} {
		res := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates", map[string]any{
			"name": name,
			"content": map[string]any{"note": "catalog-business-value-" + strings.ReplaceAll(name, " ", "-")},
		}, cfg.APIToken, "http://127.0.0.1:3100", "")
		if res.Code != http.StatusCreated || res.Header().Get(correlationHeader) == "" {
			t.Fatalf("create requires success correlation: status=%d body=%s", res.Code, res.Body.String())
		}
	}
	res := templateRequestWithHeaders(t, router, cfg, http.MethodGet, "/api/templates?limit=1&offset=1", nil, cfg.APIToken, "", "")
	if res.Code != http.StatusOK || strings.Contains(res.Body.String(), "catalog-business-value") || strings.Contains(res.Body.String(), "custom_fields") {
		t.Fatalf("catalog must remain paged redacted projection: status=%d body=%s", res.Code, res.Body.String())
	}
	var catalog struct {
		Data   []map[string]any `json:"data"`
		Total  int              `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.NewDecoder(res.Body).Decode(&catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Data) != 1 || catalog.Total != 3 || catalog.Limit != 1 || catalog.Offset != 1 {
		t.Fatalf("pagination mismatch: %+v", catalog)
	}
}

func TestT21TemplateDraftArchivedProfileSnapshotAndAuditIsolation(t *testing.T) {
	cfg, router, profiles, profileAudit, root := templateRouterWithProfileAudit(t)
	fixture := &profile.Profile{ID: "prof_t21fixture", Name: "T21 Fixture", RuntimeID: "cloakbrowser", LifecycleState: profile.LifecycleActive}
	if err := profiles.Create(fixture); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(root, "profiles", fixture.ID, "profile.json")
	before, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	created := createTemplateHTTP(t, router, cfg)
	id := created["template_id"].(string)
	valid := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/draft", map[string]any{"base_draft": map[string]any{"tags": []string{"fixture"}}}, cfg.APIToken, "http://127.0.0.1:3100", "")
	if valid.Code != http.StatusOK || valid.Header().Get(correlationHeader) == "" {
		t.Fatalf("draft must succeed with correlation: status=%d body=%s", valid.Code, valid.Body.String())
	}
	after, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("draft must leave existing profile.json byte-for-byte unchanged")
	}
	var profileEvents int
	if err := profileAudit.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE entity_id=?`, fixture.ID).Scan(&profileEvents); err != nil {
		t.Fatal(err)
	}
	if profileEvents != 0 {
		t.Fatalf("draft must not create Profile audit events, got %d", profileEvents)
	}
	archived := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/archive", nil, cfg.APIToken, "http://127.0.0.1:3100", "")
	if archived.Code != http.StatusOK {
		t.Fatalf("archive status=%d body=%s", archived.Code, archived.Body.String())
	}
	refused := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/draft", map[string]any{"base_draft": map[string]any{"note": "refused-draft-value"}}, cfg.APIToken, "http://127.0.0.1:3100", "")
	if refused.Code != http.StatusConflict || !strings.Contains(refused.Body.String(), "TEMPLATE_VERSION_NOT_ACTIVE") || strings.Contains(refused.Body.String(), "refused-draft-value") || strings.Contains(refused.Body.String(), "draft") {
		t.Fatalf("archived draft must refuse without a draft payload: status=%d body=%s", refused.Code, refused.Body.String())
	}
}

func TestT21TemplateAuditRedactsSuccessConflictAndRefusal(t *testing.T) {
	cfg, router, _ := templateRouter(t)
	markers := []string{"group-secret-marker", "tag-secret-marker", "note-secret-marker", "field-secret-marker", "draft-secret-marker"}
	created := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates", map[string]any{
		"name": "redaction evidence",
		"content": map[string]any{
			"group": "group-secret-marker", "tags": []string{"tag-secret-marker"}, "note": "note-secret-marker",
			"custom_fields": map[string]any{"field-secret-marker": map[string]any{"type": "text", "value": "field-secret-marker"}},
		},
	}, cfg.APIToken, "http://127.0.0.1:3100", "")
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var createPayload struct {
		Data struct {
			TemplateID string `json:"template_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(created.Body).Decode(&createPayload); err != nil {
		t.Fatal(err)
	}
	id := createPayload.Data.TemplateID
	conflict := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/draft", map[string]any{"base_draft": map[string]any{"note": "draft-secret-marker"}}, cfg.APIToken, "http://127.0.0.1:3100", "")
	if conflict.Code != http.StatusConflict || strings.Contains(conflict.Body.String(), "note-secret-marker") || strings.Contains(conflict.Body.String(), "draft-secret-marker") {
		t.Fatalf("conflict response must be redacted: status=%d body=%s", conflict.Code, conflict.Body.String())
	}
	archived := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/archive", nil, cfg.APIToken, "http://127.0.0.1:3100", "")
	if archived.Code != http.StatusOK {
		t.Fatalf("archive status=%d body=%s", archived.Code, archived.Body.String())
	}
	refused := templateRequestWithHeaders(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/draft", map[string]any{"base_draft": map[string]any{"note": "draft-secret-marker"}}, cfg.APIToken, "http://127.0.0.1:3100", "")
	if refused.Code != http.StatusConflict || strings.Contains(refused.Body.String(), "draft-secret-marker") {
		t.Fatalf("refusal response must be redacted: status=%d body=%s", refused.Code, refused.Body.String())
	}
	templateDB, err := sql.Open("sqlite", filepath.Join(cfg.DataDir, "templates.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer templateDB.Close()
	rows, err := templateDB.Query(`SELECT action, result, correlation_id, paths_json FROM template_audit_events WHERE template_id=? ORDER BY rowid`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var eventCount int
	for rows.Next() {
		var action, result, correlation, paths string
		if err := rows.Scan(&action, &result, &correlation, &paths); err != nil {
			t.Fatal(err)
		}
		if action == "" || result == "" || correlation == "" {
			t.Fatalf("audit event is missing its allowed redacted fields: action=%q result=%q correlation=%q", action, result, correlation)
		}
		combined := action + "\n" + result + "\n" + correlation + "\n" + paths
		for _, marker := range markers {
			if strings.Contains(combined, marker) {
				t.Fatalf("template audit leaked business marker %q in %q", marker, combined)
			}
		}
		eventCount++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if eventCount < 4 {
		t.Fatalf("expected success, conflict, archive success and refusal audit events, got %d", eventCount)
	}
}
