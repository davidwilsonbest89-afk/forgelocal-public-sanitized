package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/profile"
	"forgelocal/internal/proxyprovider"
)

const t26Token = "t26-simulated-provider-token"

func t26ProviderRouter(t *testing.T) (*httptest.Server, *profile.Store, string) {
	t.Helper()
	root := t.TempDir()
	profiles, err := profile.NewStore(root + "/profiles")
	if err != nil {
		t.Fatal(err)
	}
	p := &profile.Profile{Name: "T26 Provider Profile", RuntimeID: "cloakbrowser"}
	if err := profiles.Create(p); err != nil {
		t.Fatal(err)
	}
	providers, err := proxyprovider.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{store: profiles, token: t26Token, proxyProviderStore: providers, auditSink: newWriteAuditSink(nil)}
	r := chi.NewRouter()
	r.Use(correlationMiddleware)
	r.Use(originGuard)
	r.Use(h.authMiddleware)
	r.Use(h.requireLoopbackMiddleware)
	r.Post("/api/proxy-providers", h.createProxyProvider)
	r.Get("/api/proxy-providers", h.listProxyProviders)
	r.Post("/api/proxy-providers/{id}/simulate-resolve", h.simulateProxyProviderResolve)
	return httptest.NewServer(r), profiles, p.ID
}

func t26Request(t *testing.T, method, url, body, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Referer", "http://localhost:3000/")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestT26DynamicRegistrationAndSimulatedResolution(t *testing.T) {
	server, profiles, id := t26ProviderRouter(t)
	defer server.Close()
	client := server.Client()
	create := `{"id":"acme-test","name":"Acme QA","secret_ref":"provider.ref.acme-test"}`
	resp, err := client.Do(t26Request(t, http.MethodPost, server.URL+"/api/proxy-providers", create, t26Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create=%d", resp.StatusCode)
	}
	resolve := `{"profile_id":"` + id + `","region":"eu-test"}`
	resp, err = client.Do(t26Request(t, http.MethodPost, server.URL+"/api/proxy-providers/acme-test/simulate-resolve", resolve, t26Token))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resolve=%d", resp.StatusCode)
	}
	var out struct {
		Data proxyprovider.Lease `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Data.Host != "eu-test.acme-test.provider.test" || out.Data.Mode != "simulated" || out.Data.SecretRef != "provider.ref.acme-test" {
		t.Fatalf("lease=%#v", out.Data)
	}
	p, err := profiles.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if p.Proxy != nil {
		t.Fatalf("simulated resolution mutated proxy: %#v", p.Proxy)
	}
}

func TestT26DynamicGuardsStrictJSONAndConcurrency(t *testing.T) {
	server, _, id := t26ProviderRouter(t)
	defer server.Close()
	client := server.Client()
	path := server.URL + "/api/proxy-providers"
	valid := `{"id":"acme-test","name":"Acme QA","secret_ref":"provider.ref.acme-test"}`
	resp, err := client.Do(t26Request(t, http.MethodPost, path, valid, ""))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing token=%d", resp.StatusCode)
	}
	for _, body := range []string{`{"id":"bad-test","name":"Bad","secret_ref":"provider.ref.bad-test","api_key":"real"}`, valid + valid} {
		resp, err = client.Do(t26Request(t, http.MethodPost, path, body, t26Token))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid accepted status=%d", resp.StatusCode)
		}
	}
	foreign := t26Request(t, http.MethodPost, path, valid, t26Token)
	foreign.Header.Set("Origin", "https://external.example")
	foreign.Header.Del("Referer")
	resp, err = client.Do(foreign)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("foreign origin=%d", resp.StatusCode)
	}
	resp, err = client.Do(t26Request(t, http.MethodPost, path, valid, t26Token))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("valid=%d", resp.StatusCode)
	}
	var wg sync.WaitGroup
	for range []int{1, 2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := `{"profile_id":"` + id + `","region":"us-test"}`
			r, e := client.Do(t26Request(t, http.MethodPost, server.URL+"/api/proxy-providers/acme-test/simulate-resolve", body, t26Token))
			if e != nil {
				t.Errorf("resolve: %v", e)
				return
			}
			r.Body.Close()
			if r.StatusCode != http.StatusOK {
				t.Errorf("resolve=%d", r.StatusCode)
			}
		}()
	}
	wg.Wait()
}
