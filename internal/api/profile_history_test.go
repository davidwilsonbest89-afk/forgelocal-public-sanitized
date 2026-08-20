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

	"forgelocal/internal/config"
	"forgelocal/internal/profile"
	_ "modernc.org/sqlite"
)

func testHistoryRouter(t *testing.T) (http.Handler, *config.Config, *profile.Store) {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir, DefaultRuntimeID: "cloakbrowser", Runtimes: map[string]config.RuntimeConfig{
		"cloakbrowser": {BinaryPath: "/opt/cloakbrowser"},
		"camoufox": {BinaryPath: "/opt/camoufox"},
	}}
	store, err := profile.NewStore(filepath.Join(dir, "profiles"))
	if err != nil { t.Fatal(err) }
	r, err := NewRouter(cfg, store, testManagerWithRuntimeConfig(t, cfg), nil, nil)
	if err != nil { t.Fatal(err) }
	return r, cfg, store
}

func historyRequest(method, path, body, token string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:7777"
	req.Header.Set("Authorization", "Bearer "+token)
	if method != http.MethodGet { req.Header.Set("Origin", "http://localhost:3000") }
	if body != "" { req.Header.Set("Content-Type", "application/json") }
	return req
}

func createHistoryProfile(t *testing.T, r http.Handler, token string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles", `{"name":"History One","runtime_id":"cloakbrowser"}`, token))
	if rec.Code != http.StatusCreated { t.Fatalf("create: %d %s", rec.Code, rec.Body.String()) }
	var out struct { Data struct { ID string `json:"id"` } `json:"data"` }
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil { t.Fatal(err) }
	if out.Data.ID == "" { t.Fatalf("missing profile id: %s", rec.Body.String()) }
	return out.Data.ID
}

func TestT22ProfileHistoryReadDiffRestoreAndRedaction(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPut, "/api/profiles/"+id, `{"name":"History Two"}`, cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("update: %d %s", rec.Code, rec.Body.String()) }
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPut, "/api/profiles/"+id+"/metadata", `{"note":"t22-private-note","custom_fields":{"status":{"type":"text","value":"t22-private-value"}}}`, cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("metadata: %d %s", rec.Code, rec.Body.String()) }

	profilePath := filepath.Join(cfg.DataDir, "profiles", id, "profile.json")
	before, err := os.ReadFile(profilePath)
	if err != nil { t.Fatal(err) }
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history?limit=10&offset=0", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"version":3`)) { t.Fatalf("list: %d %s", rec.Code, rec.Body.String()) }
	after, err := os.ReadFile(profilePath)
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(before, after) { t.Fatal("history list modified profile.json") }
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history?limit=1&offset=1", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"limit":1`)) || !bytes.Contains(rec.Body.Bytes(), []byte(`"offset":1`)) || !bytes.Contains(rec.Body.Bytes(), []byte(`"version":2`)) { t.Fatalf("pagination: %d %s", rec.Code, rec.Body.String()) }

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history/diff?from=1&to=2", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte("name")) { t.Fatalf("diff: %d %s", rec.Code, rec.Body.String()) }
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history/1", "", cfg.APIToken))
	if rec.Code != http.StatusOK || bytes.Contains(rec.Body.Bytes(), []byte("proxy.")) || bytes.Contains(rec.Body.Bytes(), []byte("profile_dir")) { t.Fatalf("version redaction: %d %s", rec.Code, rec.Body.String()) }

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+id+"/history/1/restore", `{"expected_current_version":3}`, cfg.APIToken))
	if rec.Code != http.StatusOK || rec.Header().Get(correlationHeader) == "" { t.Fatalf("restore: %d %s", rec.Code, rec.Body.String()) }
	p, err := store.Get(id)
	if err != nil || p.Name != "History One" { t.Fatalf("restored profile: %#v %v", p, err) }
	if p.HistoryPending != nil { t.Fatal("successful history restore must clear the durable pending marker") }

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+id+"/history/1/restore", `{"expected_current_version":3}`, cfg.APIToken))
	if rec.Code != http.StatusConflict || !bytes.Contains(rec.Body.Bytes(), []byte("PROFILE_HISTORY_VERSION_CONFLICT")) { t.Fatalf("conflict: %d %s", rec.Code, rec.Body.String()) }

	db, err := sql.Open("sqlite", filepath.Join(cfg.DataDir, "profile_history.sqlite"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	var auditCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_audit_events`).Scan(&auditCount); err != nil || auditCount < 5 { t.Fatalf("history audit: %d %v", auditCount, err) }
	var leaked int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_audit_events WHERE profile_id LIKE '%t22-private-%' OR action LIKE '%t22-private-%' OR correlation_id LIKE '%t22-private-%'`).Scan(&leaked); err != nil || leaked != 0 { t.Fatalf("history audit redaction: %d %v", leaked, err) }
}

func TestT22ProfileHistoryRequiresAuthAndLocalOrigin(t *testing.T) {
	r, cfg, _ := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/profiles/"+id+"/history", nil))
	if rec.Code != http.StatusUnauthorized { t.Fatalf("unauth list: %d", rec.Code) }
	for _, testCase := range []struct { name, origin, referer string }{
		{name: "origin and referer absent"},
		{name: "origin distant", origin: "https://remote.invalid"},
		{name: "referer distant", referer: "https://remote.invalid/path"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+id+"/history/1/restore", strings.NewReader(`{"expected_current_version":1}`))
			req.RemoteAddr = "127.0.0.1:7777"
			req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
			req.Header.Set("Content-Type", "application/json")
			if testCase.origin != "" { req.Header.Set("Origin", testCase.origin) }
			if testCase.referer != "" { req.Header.Set("Referer", testCase.referer) }
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden || !bytes.Contains(rec.Body.Bytes(), []byte("ORIGIN_REQUIRED_LOCAL_ONLY")) || rec.Header().Get(correlationHeader) == "" { t.Fatalf("refusal: %d %s", rec.Code, rec.Body.String()) }
		})
	}
}

func TestT22HistoryProjectionRedactsActualProxySecrets(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	p := &profile.Profile{Name: "Sensitive Proxy", RuntimeID: "cloakbrowser", LifecycleState: profile.LifecycleActive, Proxy: &profile.ProxyConfig{Type: "http", Host: "proxy.test", Port: 8080, SecretRef: "proxy.ref.sensitive"}}
	if err := store.Create(p); err != nil { t.Fatal(err) }
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPut, "/api/profiles/"+p.ID, `{"name":"Sensitive Proxy Updated"}`, cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("update: %d %s", rec.Code, rec.Body.String()) }
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+p.ID+"/history/1", "", cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("version: %d %s", rec.Code, rec.Body.String()) }
	for _, forbidden := range []string{"proxy.ref.sensitive", `"username"`, `"password"`, `"secret_ref"`, "profile_dir"} {
		if bytes.Contains(rec.Body.Bytes(), []byte(forbidden)) { t.Fatalf("history projection leaks %q: %s", forbidden, rec.Body.String()) }
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+p.ID+"/history/diff?from=1&to=1", "", cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("diff: %d %s", rec.Code, rec.Body.String()) }
	for _, forbidden := range []string{"proxy.ref.sensitive"} {
		if bytes.Contains(rec.Body.Bytes(), []byte(forbidden)) { t.Fatalf("history diff leaks %q: %s", forbidden, rec.Body.String()) }
	}
}

func TestT22HistoryConcurrentRestoreAndMutationAreSerialized(t *testing.T) {
	r, cfg, _ := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, historyRequest(http.MethodPut, "/api/profiles/"+id, `{"name":"Before race"}`, cfg.APIToken))
	if res.Code != http.StatusOK { t.Fatalf("baseline update: %d %s", res.Code, res.Body.String()) }
	start := make(chan struct{})
	statuses := make(chan int, 2)
	go func() {
		<-start
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+id+"/history/1/restore", `{"expected_current_version":2}`, cfg.APIToken))
		statuses <- rec.Code
	}()
	go func() {
		<-start
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, historyRequest(http.MethodPut, "/api/profiles/"+id, `{"name":"Concurrent mutation"}`, cfg.APIToken))
		statuses <- rec.Code
	}()
	close(start)
	for range 2 {
		if status := <-statuses; status != http.StatusOK && status != http.StatusConflict { t.Fatalf("concurrent status=%d", status) }
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history?limit=10&offset=0", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"total":`)) { t.Fatalf("post-race history: %d %s", rec.Code, rec.Body.String()) }
}

func TestT22HistoryCaptureFailureLeavesPendingAndStartupRecovers(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	p := &profile.Profile{Name: "Failure injection", RuntimeID: "cloakbrowser", LifecycleState: profile.LifecycleActive}
	if err := store.Create(p); err != nil { t.Fatal(err) }
	if err := store.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	dbPath := filepath.Join(cfg.DataDir, "profile_history.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil { t.Fatal(err) }
	if _, err := db.Exec(`DROP TABLE profile_history_versions`); err != nil { db.Close(); t.Fatal(err) }
	if err := db.Close(); err != nil { t.Fatal(err) }
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPut, "/api/profiles/"+p.ID, `{"name":"Written before failed capture"}`, cfg.APIToken))
	if rec.Code < http.StatusInternalServerError { t.Fatalf("capture failure must be surfaced: %d %s", rec.Code, rec.Body.String()) }
	pending, err := store.Get(p.ID)
	if err != nil || pending.HistoryPending == nil || pending.Name != "Written before failed capture" { t.Fatalf("profile write must persist as pending: %#v %v", pending, err) }
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(dbPath + suffix); err != nil && !os.IsNotExist(err) { t.Fatal(err) }
	}
	recovered, err := NewRouter(cfg, store, testManagerWithRuntimeConfig(t, cfg), nil, nil)
	if err != nil { t.Fatalf("startup recovery: %v", err) }
	confirmed, err := store.Get(p.ID)
	if err != nil || confirmed.HistoryPending != nil { t.Fatalf("recovery must clear pending marker: %#v %v", confirmed, err) }
	rec = httptest.NewRecorder()
	recovered.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+p.ID+"/history?limit=10&offset=0", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"action":"recovery"`)) { t.Fatalf("recovered history: %d %s", rec.Code, rec.Body.String()) }
}

func TestT23ArchiveHistoryFailureLeavesDurableLifecycleAndStartupRecovers(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	p := &profile.Profile{Name: "Archive failure injection", RuntimeID: "cloakbrowser", LifecycleState: profile.LifecycleActive}
	if err := store.Create(p); err != nil { t.Fatal(err) }
	if err := store.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil { t.Fatal(err) }
	dbPath := filepath.Join(cfg.DataDir, "profile_history.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil { t.Fatal(err) }
	if _, err := db.Exec(`DROP TABLE profile_history_versions`); err != nil { db.Close(); t.Fatal(err) }
	if err := db.Close(); err != nil { t.Fatal(err) }
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+p.ID+"/archive", "", cfg.APIToken))
	if rec.Code < http.StatusInternalServerError { t.Fatalf("archive capture failure must be surfaced: %d %s", rec.Code, rec.Body.String()) }
	pending, err := store.Get(p.ID)
	if err != nil || pending.LifecycleState != profile.LifecycleArchived || pending.ArchivedAt == nil || pending.HistoryPending == nil { t.Fatalf("archive must remain durable and pending: %#v %v", pending, err) }
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(dbPath + suffix); err != nil && !os.IsNotExist(err) { t.Fatal(err) }
	}
	if _, err := NewRouter(cfg, store, testManagerWithRuntimeConfig(t, cfg), nil, nil); err != nil { t.Fatalf("startup archive recovery: %v", err) }
	confirmed, err := store.Get(p.ID)
	if err != nil || confirmed.LifecycleState != profile.LifecycleArchived || confirmed.ArchivedAt == nil || confirmed.HistoryPending != nil { t.Fatalf("recovered archive lifecycle=%#v err=%v", confirmed, err) }
}

func TestT23ArchiveReopenConcurrentSequenceLeavesNoPendingMarker(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	start := make(chan struct{})
	statuses := make(chan int, 2)
	for _, path := range []string{"/api/profiles/" + id + "/archive", "/api/profiles/" + id + "/reopen"} {
		path := path
		go func() {
			<-start
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, historyRequest(http.MethodPost, path, "", cfg.APIToken))
			statuses <- rec.Code
		}()
	}
	close(start)
	for range 2 {
		status := <-statuses
		if status != http.StatusOK && status != http.StatusConflict { t.Fatalf("archive/reopen concurrent status=%d", status) }
	}
	current, err := store.Get(id)
	if err != nil || current.HistoryPending != nil { t.Fatalf("concurrent lifecycle left pending marker: %#v %v", current, err) }
}

func TestT23IdempotentArchiveDoesNotCreateSecondHistoryVersion(t *testing.T) {
	r, cfg, _ := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	first := httptest.NewRecorder()
	r.ServeHTTP(first, historyRequest(http.MethodPost, "/api/profiles/"+id+"/archive", "", cfg.APIToken))
	if first.Code != http.StatusOK { t.Fatalf("first archive: %d %s", first.Code, first.Body.String()) }
	db, err := sql.Open("sqlite", filepath.Join(cfg.DataDir, "profile_history.sqlite"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	var before int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_versions WHERE profile_id=? AND action='archive'`, id).Scan(&before); err != nil || before != 1 { t.Fatalf("archive versions before second request=%d err=%v", before, err) }
	second := httptest.NewRecorder()
	r.ServeHTTP(second, historyRequest(http.MethodPost, "/api/profiles/"+id+"/archive", "", cfg.APIToken))
	if second.Code != http.StatusOK || !bytes.Contains(second.Body.Bytes(), []byte(`"changed":false`)) { t.Fatalf("second archive: %d %s", second.Code, second.Body.String()) }
	var after int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_versions WHERE profile_id=? AND action='archive'`, id).Scan(&after); err != nil || after != before { t.Fatalf("idempotent archive history before=%d after=%d err=%v", before, after, err) }
}

func TestT23ArchiveReopenUpdateConcurrentProductionRouter(t *testing.T) {
	r, cfg, store := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	start := make(chan struct{})
	statuses := make(chan int, 3)
	requests := []struct{ method, path, body string }{
		{http.MethodPost, "/api/profiles/" + id + "/archive", ""},
		{http.MethodPost, "/api/profiles/" + id + "/reopen", ""},
		{http.MethodPut, "/api/profiles/" + id, `{"name":"T23 concurrent update"}`},
	}
	for _, item := range requests {
		item := item
		go func() { <-start; rec := httptest.NewRecorder(); r.ServeHTTP(rec, historyRequest(item.method, item.path, item.body, cfg.APIToken)); statuses <- rec.Code }()
	}
	close(start)
	for range requests {
		status := <-statuses
		if status != http.StatusOK && status != http.StatusConflict { t.Fatalf("concurrent T23 status=%d", status) }
	}
	current, err := store.Get(id)
	if err != nil || current.HistoryPending != nil { t.Fatalf("concurrent T23 state=%#v err=%v", current, err) }
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodGet, "/api/profiles/"+id+"/history?limit=20&offset=0", "", cfg.APIToken))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"total":`)) { t.Fatalf("concurrent history: %d %s", rec.Code, rec.Body.String()) }
}

func TestT23ArchiveReopenProductionGuardsAndRedaction(t *testing.T) {
	r, cfg, _ := testHistoryRouter(t)
	id := createHistoryProfile(t, r, cfg.APIToken)
	for _, path := range []string{"/api/profiles/" + id + "/archive", "/api/profiles/" + id + "/reopen"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = "127.0.0.1:7777"
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized { t.Fatalf("missing bearer %s: %d", path, rec.Code) }
		for _, guard := range []struct{ origin, referer string }{
			{},
			{origin: "https://remote.invalid"},
			{referer: "https://remote.invalid/path"},
		} {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.RemoteAddr = "127.0.0.1:7777"
			req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
			if guard.origin != "" { req.Header.Set("Origin", guard.origin) }
			if guard.referer != "" { req.Header.Set("Referer", guard.referer) }
			rec = httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden || !bytes.Contains(rec.Body.Bytes(), []byte("ORIGIN_REQUIRED_LOCAL_ONLY")) { t.Fatalf("origin guard %s origin=%q referer=%q: %d %s", path, guard.origin, guard.referer, rec.Code, rec.Body.String()) }
		}
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+id+"/archive", "", cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("guarded archive success: %d %s", rec.Code, rec.Body.String()) }
	for _, forbidden := range []string{"profile_dir", "history_pending", "secret_ref", "username", "password"} {
		if bytes.Contains(rec.Body.Bytes(), []byte(forbidden)) { t.Fatalf("archive response leaked %q: %s", forbidden, rec.Body.String()) }
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, historyRequest(http.MethodPost, "/api/profiles/"+id+"/reopen", "", cfg.APIToken))
	if rec.Code != http.StatusOK { t.Fatalf("guarded reopen success: %d %s", rec.Code, rec.Body.String()) }
	for _, forbidden := range []string{"profile_dir", "history_pending", "secret_ref", "username", "password"} {
		if bytes.Contains(rec.Body.Bytes(), []byte(forbidden)) { t.Fatalf("reopen response leaked %q: %s", forbidden, rec.Body.String()) }
	}
	db, err := sql.Open("sqlite", filepath.Join(cfg.DataDir, "profile_history.sqlite"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	var events int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_audit_events WHERE profile_id=? AND action='history_created' AND result='success'`, id).Scan(&events); err != nil || events < 2 { t.Fatalf("redacted archive/reopen history audit events=%d err=%v", events, err) }
	var leaked int
	if err := db.QueryRow(`SELECT COUNT(*) FROM profile_history_audit_events WHERE profile_id=? AND (action LIKE '%secret%' OR correlation_id LIKE '%secret%' OR action LIKE '%profile_dir%' OR correlation_id LIKE '%profile_dir%')`, id).Scan(&leaked); err != nil || leaked != 0 { t.Fatalf("history audit redaction leaked=%d err=%v", leaked, err) }
}
