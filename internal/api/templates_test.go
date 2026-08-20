package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

func templateRouter(t *testing.T) (*config.Config, http.Handler, string) {
	t.Helper()
	root := t.TempDir()
	profiles, err := profile.NewStore(filepath.Join(root, "profiles"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{DataDir: filepath.Join(root, "data"), ProfilesDir: filepath.Join(root, "profiles"), Version: "t21-test"}
	router, err := NewRouter(cfg, profiles, &browser.Manager{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, router, root
}

func templateRequest(t *testing.T, cfg *config.Config, method, path string, body any) *httptest.ResponseRecorder {
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
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	if method != http.MethodGet {
		req.Header.Set("Origin", "http://127.0.0.1:3100")
	}
	return httptest.NewRecorder()
}

func serveTemplate(t *testing.T, router http.Handler, cfg *config.Config, method, path string, body any) *httptest.ResponseRecorder {
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
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	if method != http.MethodGet {
		req.Header.Set("Origin", "http://127.0.0.1:3100")
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func createTemplateHTTP(t *testing.T, router http.Handler, cfg *config.Config) map[string]any {
	t.Helper()
	res := serveTemplate(t, router, cfg, http.MethodPost, "/api/templates", map[string]any{
		"name": "QA Template",
		"content": map[string]any{
			"group": "qa", "tags": []string{"smoke"}, "note": "template-business-value",
			"custom_fields": map[string]any{"owner": map[string]any{"type": "text", "value": "quality"}},
		},
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", res.Code, res.Body.String())
	}
	var decoded struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	return decoded.Data
}

func TestTemplatesCatalogReadDetailAndDraft(t *testing.T) {
	cfg, router, root := templateRouter(t)
	created := createTemplateHTTP(t, router, cfg)
	id, _ := created["template_id"].(string)
	if id == "" {
		t.Fatalf("create response=%v", created)
	}
	list := serveTemplate(t, router, cfg, http.MethodGet, "/api/templates?limit=10&offset=0", nil)
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), "template-business-value") || strings.Contains(list.Body.String(), "custom_fields") {
		t.Fatalf("catalog must be a projection, status=%d body=%s", list.Code, list.Body.String())
	}
	detail := serveTemplate(t, router, cfg, http.MethodGet, "/api/templates/"+id+"/versions/1", nil)
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), "template-business-value") {
		t.Fatalf("detail must expose content to authenticated loopback caller: %d %s", detail.Code, detail.Body.String())
	}
	before, err := filepath.Glob(filepath.Join(root, "profiles", "*", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	draft := serveTemplate(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/1/draft", map[string]any{"base_draft": map[string]any{"tags": []string{"base"}}})
	if draft.Code != http.StatusOK || !strings.Contains(draft.Body.String(), "VALID") || !strings.Contains(draft.Body.String(), "base") {
		t.Fatalf("draft status=%d body=%s", draft.Code, draft.Body.String())
	}
	after, err := filepath.Glob(filepath.Join(root, "profiles", "*", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("draft changed Profile Store files before=%v after=%v", before, after)
	}
}

func TestTemplatesVersionArchiveConflictAndLoopbackRefusal(t *testing.T) {
	cfg, router, _ := templateRouter(t)
	created := createTemplateHTTP(t, router, cfg)
	id := created["template_id"].(string)
	version := serveTemplate(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions", map[string]any{
		"expected_active_version": 1,
		"content":                 map[string]any{"note": "v2", "tags": []string{"next"}},
	})
	if version.Code != http.StatusCreated {
		t.Fatalf("new version status=%d body=%s", version.Code, version.Body.String())
	}
	conflict := serveTemplate(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/2/draft", map[string]any{"base_draft": map[string]any{"note": "different"}})
	if conflict.Code != http.StatusConflict || !strings.Contains(conflict.Body.String(), "CONFLICT") || strings.Contains(conflict.Body.String(), "v2") {
		t.Fatalf("draft conflict=%d %s", conflict.Code, conflict.Body.String())
	}
	archived := serveTemplate(t, router, cfg, http.MethodPost, "/api/templates/"+id+"/versions/2/archive", nil)
	if archived.Code != http.StatusOK {
		t.Fatalf("archive=%d %s", archived.Code, archived.Body.String())
	}
	archivedGet := serveTemplate(t, router, cfg, http.MethodGet, "/api/templates/"+id+"/versions/2", nil)
	if archivedGet.Code != http.StatusOK || !strings.Contains(archivedGet.Body.String(), "archived") {
		t.Fatalf("archived get=%d %s", archivedGet.Code, archivedGet.Body.String())
	}
	req := httptest.NewRequest(http.MethodPost, "/api/templates", strings.NewReader(`{"name":"blocked","content":{}}`))
	req.RemoteAddr = "203.0.113.9:9000"
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("Origin", "http://203.0.113.9")
	denied := httptest.NewRecorder()
	router.ServeHTTP(denied, req)
	if denied.Code != http.StatusForbidden || !strings.Contains(denied.Body.String(), "ORIGIN_REQUIRED_LOCAL_ONLY") {
		t.Fatalf("off-loopback mutation=%d %s", denied.Code, denied.Body.String())
	}
}
