package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/cookies"
	"forgelocal/internal/profile"
)

const t25Token = "t25-synthetic-only-token"

func t25FixtureRouter(t *testing.T) (*httptest.Server, *profile.Store, string) {
	t.Helper()
	root := t.TempDir()
	profiles, err := profile.NewStore(root + "/profiles")
	if err != nil {
		t.Fatal(err)
	}
	p := &profile.Profile{Name: "T25 Fixture Profile", RuntimeID: "cloakbrowser"}
	if err := profiles.Create(p); err != nil {
		t.Fatal(err)
	}
	fixtures, err := cookies.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{store: profiles, token: t25Token, cookieFixtureStore: fixtures, auditSink: newWriteAuditSink(nil)}
	r := chi.NewRouter()
	r.Use(correlationMiddleware)
	r.Use(originGuard)
	r.Use(h.authMiddleware)
	r.Use(h.requireLoopbackMiddleware)
	r.Post("/api/profiles/{id}/cookie-fixtures/import", h.importCookieFixtures)
	r.Get("/api/profiles/{id}/cookie-fixtures/export", h.exportCookieFixtures)
	return httptest.NewServer(r), profiles, p.ID
}

func t25Request(t *testing.T, method, url, body string, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Referer", "http://localhost:3000/")
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestT25DynamicFixtureImportExportRedactsValuesAndIsAtomic(t *testing.T) {
	server, _, id := t25FixtureRouter(t)
	defer server.Close()
	client := server.Client()
	valid := `{"fixtures":[{"name":"theme","value":"fixture:blue","domain":"app.test","path":"/","same_site":"lax"}]}`
	resp, err := client.Do(t25Request(t, http.MethodPost, server.URL+"/api/profiles/"+id+"/cookie-fixtures/import", valid, t25Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import=%d", resp.StatusCode)
	}
	var raw bytes.Buffer
	if _, err := raw.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw.String(), "fixture:blue") || strings.Contains(raw.String(), `"value"`) {
		t.Fatalf("value leaked: %s", raw.String())
	}
	resp, err = client.Do(t25Request(t, http.MethodGet, server.URL+"/api/profiles/"+id+"/cookie-fixtures/export", "", t25Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var exported struct {
		Data struct {
			Count    int               `json:"count"`
			Fixtures []cookies.Fixture `json:"fixtures"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}
	if exported.Data.Count != 1 || exported.Data.Fixtures[0].ValueDigest == "" {
		t.Fatalf("export=%#v", exported)
	}
	invalid := `{"fixtures":[{"name":"theme","value":"real-cookie","domain":"app.test","path":"/"}]}`
	resp, err = client.Do(t25Request(t, http.MethodPost, server.URL+"/api/profiles/"+id+"/cookie-fixtures/import", invalid, t25Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid=%d", resp.StatusCode)
	}
	resp, err = client.Do(t25Request(t, http.MethodGet, server.URL+"/api/profiles/"+id+"/cookie-fixtures/export", "", t25Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}
	if exported.Data.Count != 1 {
		t.Fatalf("invalid import changed store: %#v", exported)
	}
}

func TestT25DynamicGuardsStrictJSONAndConcurrency(t *testing.T) {
	server, _, id := t25FixtureRouter(t)
	defer server.Close()
	client := server.Client()
	path := server.URL + "/api/profiles/" + id + "/cookie-fixtures/import"
	valid := `{"fixtures":[{"name":"mode","value":"fixture:one","domain":"qa.test","path":"/"}]}`
	resp, err := client.Do(t25Request(t, http.MethodPost, path, valid, ""))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing token=%d", resp.StatusCode)
	}
	for _, body := range []string{`{"fixtures":[],"unknown":true}`, valid + valid} {
		resp, err = client.Do(t25Request(t, http.MethodPost, path, body, t25Token))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("strict JSON accepted %q status=%d", body, resp.StatusCode)
		}
	}
	var wg sync.WaitGroup
	for _, marker := range []string{"fixture:a", "fixture:b"} {
		wg.Add(1)
		go func(marker string) {
			defer wg.Done()
			body := `{"fixtures":[{"name":"mode","value":"` + marker + `","domain":"qa.test","path":"/"}]}`
			resp, err := client.Do(t25Request(t, http.MethodPost, path, body, t25Token))
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("concurrent status=%d", resp.StatusCode)
				}
			} else {
				t.Errorf("concurrent request: %v", err)
			}
		}(marker)
	}
	wg.Wait()
	resp, err = client.Do(t25Request(t, http.MethodGet, server.URL+"/api/profiles/"+id+"/cookie-fixtures/export", "", t25Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var exported struct {
		Data struct {
			Count    int               `json:"count"`
			Fixtures []cookies.Fixture `json:"fixtures"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&exported); err != nil {
		t.Fatal(err)
	}
	if exported.Data.Count != 1 || exported.Data.Fixtures[0].ValueDigest == "" {
		t.Fatalf("concurrent export=%#v", exported)
	}
}
