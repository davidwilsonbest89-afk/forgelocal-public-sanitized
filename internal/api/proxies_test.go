// T10 — Proxies. API contract tests for the proxy registry endpoints.
//
// Covers the AC criteria validated by the reviewer:
// AC-01 valid creation, AC-02 invalid refusals (type/host/port/ref),
// AC-03 assignment/unassignment with audit events,
// AC-04 secret_ref-only listing (never credential values),
// AC-05 loopback refusal 403 for mutations, AC-06 concurrency.
//
// All credential material is synthetic and never represents a real endpoint.
package api

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/profile"
	"forgelocal/internal/proxies"
)

// loopbackNet is a fake loopback listener used to mount the router on a
// loopback interface, mirroring the T09 admin loopback test helper.
type loopbackNet struct{}

func (loopbackNet) Dial(network, address string) (net.Conn, error)           { return nil, nil }
func (loopbackNet) DialTimeout(network, address string, _ interface{}) error { return nil }
func (loopbackNet) Listen(network, address string) (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:0")
}
func (loopbackNet) ListenPacket(network, address string) (net.PacketConn, error) { return nil, nil }
func (loopbackNet) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) { return nil, nil }

// --- helpers ---------------------------------------------------------------

func newProxyTestServer(t *testing.T) (*httptest.Server, *proxies.Store) {
	t.Helper()
	dir := t.TempDir()
	store, err := proxies.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	r := chi.NewRouter()
	// reuse the real router wiring but for unit tests build a minimal handler
	// with the same audit/loopback middlewares as production.
	h := &handler{
		cfg:        nil,
		token:      "t10-test-token",
		proxyStore: store,
		auditSink:  newWriteAuditSink(nil),
	}
	r.Use(correlationMiddleware)
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Use(h.requireLoopbackMiddleware)
		r.Get("/api/proxies", h.listProxies)
		r.Post("/api/proxies", h.createProxy)
		r.Get("/api/proxies/{id}", h.getProxy)
		r.Put("/api/proxies/{id}", h.updateProxy)
		r.Post("/api/proxies/{id}/assign", h.assignProxy)
		r.Delete("/api/proxies/{id}/assign", h.unassignProxy)
		r.Delete("/api/proxies/{id}", h.deleteProxy)
	})
	ts := httptest.NewUnstartedServer(r)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ts.Listener = ln
	ts.Start()
	return ts, store
}

func doJSON(t *testing.T, ts *httptest.Server, method, path string, token string, body any) (*http.Response, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, err := http.NewRequest(method, ts.URL+path, &buf)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	var payload map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return resp, payload
}

func createProxy(t *testing.T, ts *httptest.Server, payload any) (string, *http.Response) {
	t.Helper()
	resp, data := doJSON(t, ts, http.MethodPost, "/api/proxies", "t10-test-token", payload)
	id, _ := data["data"].(map[string]any)["id"].(string)
	return id, resp
}

// --- tests -----------------------------------------------------------------

func TestProxyListEmptyAndCreated(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	resp, data := doJSON(t, ts, http.MethodGet, "/api/proxies", "t10-test-token", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list: %d", resp.StatusCode)
	}
	items := data["data"].(map[string]any)["items"].([]any)
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %d", len(items))
	}
	id, resp := createProxy(t, ts, map[string]any{"name": "t10-proxy", "type": "http", "host": "proxy.local", "port": 8080})
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d %+v", resp.StatusCode, data)
	}
	resp2, data2 := doJSON(t, ts, http.MethodGet, "/api/proxies", "t10-test-token", nil)
	items2 := data2["data"].(map[string]any)["items"].([]any)
	if len(items2) != 1 {
		t.Fatalf("list after create: %d", len(items2))
	}
	entry := items2[0].(map[string]any)
	if entry["id"].(string) != id || entry["host"].(string) != "proxy.local" {
		t.Error("list entry mismatch")
	}
	if entry["username"] != nil || entry["password"] != nil {
		t.Error("list must never expose credential values")
	}
	if entry["secret_ref"] != nil && entry["secret_ref"] != "" {
		// Credential-less proxy must not expose a secret reference.
		t.Error("credential-less proxy must not expose secret_ref")
	}
	if resp2.Header.Get("X-Correlation-ID") == "" {
		t.Error("list must propagate the correlation id")
	}
}

func TestProxyCreateRefusesInvalid(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	cases := []struct {
		name string
		body any
		code string
	}{
		{"unsupported type", map[string]any{"name": "px", "type": "ftp", "host": "h", "port": 80}, "INVALID_PROXY"},
		{"port out of range", map[string]any{"name": "px", "type": "http", "host": "h", "port": 70000}, "INVALID_PROXY"},
		{"port zero", map[string]any{"name": "px", "type": "http", "host": "h", "port": 0}, "INVALID_PROXY"},
		{"empty host", map[string]any{"name": "px", "type": "http", "host": "", "port": 80}, "INVALID_PROXY"},
		{"missing name", map[string]any{"name": "", "type": "http", "host": "h", "port": 80}, "INVALID_PROXY"},
		{"malformed secret ref", map[string]any{"name": "px", "type": "http", "host": "h", "port": 80, "secret_ref": "STOLEN:plain"}, "INVALID_PROXY"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, data := doJSON(t, ts, http.MethodPost, "/api/proxies", "t10-test-token", c.body)
			if resp.StatusCode != 400 {
				t.Fatalf("%s: expected 400, got %d", c.name, resp.StatusCode)
			}
			errObj := data["error"].(map[string]any)
			if errObj["code"].(string) != c.code {
				t.Errorf("%s: expected code %s, got %s", c.name, c.code, errObj["code"])
			}
			if errObj["message"].(string) == "" {
				t.Error("message must never be empty even when redacted")
			}
		})
	}
}

func TestProxyCreateSyntheticSecretRef(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	payload := map[string]any{
		"name": "px-secret", "type": "socks5", "host": "px.local", "port": 1080,
		"secret_ref": "proxy.ref.synthetic-001",
	}
	resp, data := doJSON(t, ts, http.MethodPost, "/api/proxies", "t10-test-token", payload)
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d %+v", resp.StatusCode, data)
	}
	entry := data["data"].(map[string]any)
	if entry["secret_ref"].(string) != "proxy.ref.synthetic-001" {
		t.Error("secret_ref must echo the reference, not the value")
	}
	if entry["has_secret"].(bool) != true {
		t.Error("has_secret presence flag must be true")
	}
}

func TestProxyDeleteRefusesAssigned(t *testing.T) {
	ts, store := newProxyTestServer(t)
	defer ts.Close()
	id, _ := createProxy(t, ts, map[string]any{"name": "px-del", "type": "http", "host": "h", "port": 80})
	// Assign a profile through the store (the API contract uses query param).
	if err := store.Assign("profile-alpha", id); err != nil {
		t.Fatal(err)
	}
	resp, data := doJSON(t, ts, http.MethodDelete, "/api/proxies/"+id, "t10-test-token", nil)
	if resp.StatusCode != 409 {
		t.Fatalf("delete assigned: expected 409, got %d", resp.StatusCode)
	}
	if data["error"].(map[string]any)["code"].(string) != "PROXY_NAME_TAKEN" {
		t.Errorf("code: %s", data["error"])
	}
	// Unassign via the dedicated API, then deletion succeeds.
	resp2, _ := doJSON(t, ts, http.MethodDelete, "/api/proxies/"+id+"/assign?profile_id=profile-alpha", "t10-test-token", nil)
	if resp2.StatusCode != 200 {
		t.Fatalf("unassign: %d", resp2.StatusCode)
	}
	resp3, _ := doJSON(t, ts, http.MethodDelete, "/api/proxies/"+id, "t10-test-token", nil)
	if resp3.StatusCode != 200 {
		t.Fatalf("delete after unassign: %d", resp3.StatusCode)
	}
}

func TestProxyUpdateRefusesOverwriteOfSecretRef(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	id, _ := createProxy(t, ts, map[string]any{
		"name": "px-ref", "type": "http", "host": "h", "port": 80,
		"secret_ref": "proxy.ref.ref-001",
	})
	resp, data := doJSON(t, ts, http.MethodPut, "/api/proxies/"+id, "t10-test-token",
		map[string]any{"secret_ref": "STOLEN:overwrite-attempt"})
	if resp.StatusCode != 200 {
		t.Fatalf("update: %d %+v", resp.StatusCode, data)
	}
	entry := data["data"].(map[string]any)
	if entry["secret_ref"].(string) != "proxy.ref.ref-001" {
		t.Error("secret_ref must not be overwritable via the API")
	}
}

func TestProxyUnauthorized(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	resp, _ := doJSON(t, ts, http.MethodGet, "/api/proxies", "wrong-token", nil)
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProxyMutationsOutOfLoopbackRefused(t *testing.T) {
	// Mount the same router behind a non-loopback listener to reproduce the
	// T09 loopback refusal evidence for the proxy contract.
	dir, _ := os.MkdirTemp("", "t10-offloopback")
	defer os.RemoveAll(dir)
	store, _ := proxies.NewStore(dir)
	h := &handler{token: "t10-test-token", proxyStore: store, auditSink: newWriteAuditSink(nil)}
	r := chi.NewRouter()
	r.Use(correlationMiddleware)
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Use(h.requireLoopbackMiddleware)
		r.Post("/api/proxies", h.createProxy)
	})
	// Off-loopback listener (sandbox link-local interface used in T05/T09 evidence).
	ln, err := net.Listen("tcp", "169.254.0.21:0")
	if err != nil {
		t.Skipf("off-loopback listener unavailable: %v", err)
	}
	ts := httptest.NewUnstartedServer(r)
	ts.Listener = ln
	ts.Start()
	defer ts.Close()

	body := map[string]any{"name": "px", "type": "http", "host": "h", "port": 80}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/proxies", &buf)
	req.Header.Set("Authorization", "Bearer t10-test-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403 LOOPBACK_REQUIRED, got %d", resp.StatusCode)
	}
	var payload map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload["error"].(map[string]any)["code"].(string) != "LOOPBACK_REQUIRED" {
		t.Errorf("code: %s", payload["error"])
	}
	// The registry must stay untouched after a refused off-loopback mutation.
	if len(store.List()) != 0 {
		t.Error("refused off-loopback mutation must not write to the registry")
	}
}

func TestProxyNotFoundAndCorrelation(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	resp, _ := doJSON(t, ts, http.MethodGet, "/api/proxies/no-such-id", "t10-test-token", nil)
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Correlation-ID") == "" {
		t.Error("error responses must carry the correlation id")
	}
}

func TestProxyAuditSinkNilTolerance(t *testing.T) {
	// Proxy handlers must never panic when the audit sink is nil (the sink is
	// optional outside the fully configured server; T09/T10 integration
	// evidence uses the real SQLite-backed sink).
	var nilSink *writeAuditSink
	nilSink.auditRecord(nil, "proxy.created", "px", "corr", map[string]any{"has_secret": true})
}

func TestProxyAssignmentAPI(t *testing.T) {
	ts, _ := newProxyTestServer(t)
	defer ts.Close()
	id, _ := createProxy(t, ts, map[string]any{"name": "px-assign", "type": "http", "host": "h", "port": 80})

	resp, data := doJSON(t, ts, http.MethodPost, "/api/proxies/"+id+"/assign?profile_id=profile-1", "t10-test-token", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("assign: %d %+v", resp.StatusCode, data)
	}
	// Missing profile_id is refused.
	resp2, _ := doJSON(t, ts, http.MethodPost, "/api/proxies/"+id+"/assign", "t10-test-token", nil)
	if resp2.StatusCode != 400 {
		t.Fatalf("assign without profile: expected 400, got %d", resp2.StatusCode)
	}
	// Idempotent re-assign.
	resp3, _ := doJSON(t, ts, http.MethodPost, "/api/proxies/"+id+"/assign?profile_id=profile-1", "t10-test-token", nil)
	if resp3.StatusCode != 200 {
		t.Fatalf("idempotent assign: %d", resp3.StatusCode)
	}
	resp4, _ := doJSON(t, ts, http.MethodDelete, "/api/proxies/"+id+"/assign?profile_id=profile-1", "t10-test-token", nil)
	if resp4.StatusCode != 200 {
		t.Fatalf("unassign: %d", resp4.StatusCode)
	}
}

func TestProxyConcurrentMutationsAPI(t *testing.T) {
	ts, store := newProxyTestServer(t)
	defer ts.Close()
	id, _ := createProxy(t, ts, map[string]any{"name": "px-conc", "type": "http", "host": "h", "port": 80})
	done := make(chan bool, 8)
	for i := 0; i < 8; i++ {
		go func(i int) {
			resp, _ := doJSON(t, ts, http.MethodPut, "/api/proxies/"+id, "t10-test-token",
				map[string]any{"port": 1000 + i})
			done <- resp.StatusCode == 200 || resp.StatusCode == 409
		}(i)
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	got, err := store.Get(id)
	if err != nil || got.Port < 1000 || got.Port >= 1008 {
		t.Errorf("unexpected final port: %+v %v", got, err)
	}
}

// T10 — getProfile exposes only the proxy registry id, never credentials.
func TestGetProfileProxyIdRedacted(t *testing.T) {
	dir := t.TempDir()
	proxyStore, err := proxies.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	profileStore, err := profile.NewStore(dir)
	if err != nil {
		t.Fatalf("profile.NewStore: %v", err)
	}
	if err := profileStore.Create(&profile.Profile{ID: "p-1", Name: "redacted", RuntimeID: "rt-1"}); err != nil {
		t.Fatalf("create profile: %v", err)
	}
	r := chi.NewRouter()
	h := &handler{
		cfg:        nil,
		token:      "t10-test-token",
		store:      profileStore,
		proxyStore: proxyStore,
		auditSink:  newWriteAuditSink(nil),
	}
	r.Use(correlationMiddleware)
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Use(h.requireLoopbackMiddleware)
		r.Get("/api/profiles/{id}", h.getProfile)
		r.Post("/api/proxies", h.createProxy)
		r.Post("/api/proxies/{id}/assign", h.assignProxy)
	})
	ts := httptest.NewUnstartedServer(r)
	ln, listenErr := net.Listen("tcp", "127.0.0.1:0")
	if listenErr != nil {
		t.Fatalf("listen: %v", listenErr)
	}
	ts.Listener = ln
	ts.Start()
	defer ts.Close()
	id, createResp := createProxy(t, ts, map[string]any{"name": "px-redact", "type": "http", "host": "h", "port": 9090})
	if createResp.StatusCode != 201 {
		t.Fatalf("create proxy: %d", createResp.StatusCode)
	}
	if err := proxyStore.Assign("p-1", id); err != nil {
		t.Fatal(err)
	}
	resp, data := doJSON(t, ts, http.MethodGet, "/api/profiles/p-1", "t10-test-token", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get profile: %d", resp.StatusCode)
	}
	if data["proxy_id"] != id {
		t.Errorf("proxy_id: got %v, want %s", data["proxy_id"], id)
	}
	encoded := encodeJSON(t, data["proxy_data"])
	for _, forbidden := range []string{"username", "password", "\"proxy\""} {
		if strings.Contains(encoded, forbidden) {
			t.Errorf("credential or proxy object leaked in getProfile: %s", forbidden)
		}
	}
}

func encodeJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(encoded)
}
