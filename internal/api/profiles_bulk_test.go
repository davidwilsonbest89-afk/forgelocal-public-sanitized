// T24 — Bounded bulk profile operations: API, concurrency and redaction tests.

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"forgelocal/internal/config"
	"forgelocal/internal/groups"
	"forgelocal/internal/history"
	"forgelocal/internal/profile"
	_ "modernc.org/sqlite"
)

const bulkTestToken = "t24-test-token"

type bulkResponse struct {
	Data struct {
		Operation string                       `json:"operation"`
		Results   []bulkProfileOperationResult `json:"results"`
		Summary   bulkProfileOperationSummary  `json:"summary"`
	} `json:"data"`
}

func testBulkRouter(t *testing.T) (http.Handler, *sql.DB, *profile.Store, *groups.Store, *history.Store) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open audit db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := createAuditTable(db); err != nil {
		t.Fatalf("create audit table: %v", err)
	}
	dataDir := t.TempDir()
	store, err := profile.NewStore(dataDir)
	if err != nil {
		t.Fatalf("new profile store: %v", err)
	}
	historyStore, err := history.Open(dataDir)
	if err != nil {
		t.Fatalf("open history store: %v", err)
	}
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new group store: %v", err)
	}
	cfg := &config.Config{DefaultRuntimeID: "cloakbrowser", Runtimes: map[string]config.RuntimeConfig{"cloakbrowser": {BinaryPath: "/opt/cloakbrowser"}}}
	h := &handler{cfg: cfg, store: store, groupStore: groupStore, mgr: testManagerWithRuntimeConfig(t, cfg), token: bulkTestToken, auditSink: &writeAuditSink{db: db}, historyStore: historyStore}
	r := chi.NewRouter()
	r.Use(correlationMiddleware)
	r.Use(originGuard)
	r.Use(h.authMiddleware)
	r.Use(h.requireLoopbackMiddleware)
	r.Post("/api/profiles/bulk", h.bulkProfileOperation)
	return r, db, store, groupStore, historyStore
}

func createBulkProfile(t *testing.T, store *profile.Store, name string) string {
	t.Helper()
	p := &profile.Profile{Name: name, RuntimeID: "cloakbrowser"}
	if err := store.Create(p); err != nil {
		t.Fatalf("create profile %q: %v", name, err)
	}
	// Fixtures bypass the production create handler. Confirm the durable create
	// marker explicitly so later assertions isolate the History capture generated
	// by the T24 operation under test.
	if p.HistoryPending != nil {
		if err := store.ClearHistoryPending(p.ID, p.HistoryPending.OperationID); err != nil {
			t.Fatalf("confirm fixture create history for %q: %v", name, err)
		}
	}
	return p.ID
}

func newBulkRequest(body string) *http.Request {
	req := newLoopbackRequest(http.MethodPost, "/api/profiles/bulk", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bulkTestToken)
	return req
}

func callBulk(t *testing.T, r http.Handler, body string) (bulkResponse, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, newBulkRequest(body))
	var got bulkResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode bulk response: %v body=%s", err, rec.Body.String())
		}
	}
	return got, rec
}

func TestT24BulkArchiveReopenAndIdempotence(t *testing.T) {
	r, db, store, _, _ := testBulkRouter(t)
	first := createBulkProfile(t, store, "T24 First")
	second := createBulkProfile(t, store, "T24 Second")

	got, rec := callBulk(t, r, `{"operation":"archive","profile_ids":["`+first+`","`+second+`"]}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Changed != 2 || got.Data.Summary.Failed != 0 {
		t.Fatalf("archive response=%d %#v body=%s", rec.Code, got, rec.Body.String())
	}
	for _, id := range []string{first, second} {
		p, err := store.Get(id)
		if err != nil || p.LifecycleState != profile.LifecycleArchived {
			t.Fatalf("profile %s lifecycle=%v err=%v", id, p, err)
		}
	}
	if got := rec.Header().Get(correlationHeader); got == "" {
		t.Fatal("bulk archive missing correlation id")
	}
	var auditCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE event_type='profile.bulk_changed'`).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 2 {
		t.Fatalf("bulk changed audits=%d want=2", auditCount)
	}

	got, rec = callBulk(t, r, `{"operation":"archive","profile_ids":["`+first+`","`+second+`"]}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Noop != 2 || got.Data.Summary.Changed != 0 {
		t.Fatalf("idempotent archive response=%d %#v", rec.Code, got)
	}
	got, rec = callBulk(t, r, `{"operation":"reopen","profile_ids":["`+first+`","`+second+`"]}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Changed != 2 {
		t.Fatalf("reopen response=%d %#v", rec.Code, got)
	}
}

func TestT24BulkTagsGroupsPartialFailureAndRedaction(t *testing.T) {
	r, db, store, groupStore, historyStore := testBulkRouter(t)
	first := createBulkProfile(t, store, "T24 Tags First")
	second := createBulkProfile(t, store, "T24 Tags Archived")
	if err := store.ArchiveProfile(second); err != nil {
		t.Fatal(err)
	}
	if _, err := groupStore.Upsert("T24 Group", &profile.ProxyConfig{Type: "http", Host: "proxy.example.invalid", Port: 8080}, groups.ProxyModeDefault); err != nil {
		t.Fatal(err)
	}

	got, rec := callBulk(t, r, `{"operation":"add_tag","profile_ids":["`+first+`","`+second+`"],"tag":"batch"}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Changed != 1 || got.Data.Summary.Failed != 1 {
		t.Fatalf("partial tag response=%d %#v body=%s", rec.Code, got, rec.Body.String())
	}
	p, err := store.Get(first)
	if err != nil || !containsString(p.Tags, "batch") {
		t.Fatalf("first profile tag missing: %+v err=%v", p, err)
	}

	got, rec = callBulk(t, r, `{"operation":"set_group","profile_ids":["`+first+`"],"group":"T24 Group"}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Changed != 1 {
		t.Fatalf("set group response=%d %#v", rec.Code, got)
	}
	p, err = store.Get(first)
	if err != nil || p.Group != "T24 Group" {
		t.Fatalf("group update=%+v err=%v", p, err)
	}
	got, rec = callBulk(t, r, `{"operation":"clear_group","profile_ids":["`+first+`"]}`)
	if rec.Code != http.StatusOK || got.Data.Summary.Changed != 1 {
		t.Fatalf("clear group response=%d %#v", rec.Code, got)
	}
	versions, err := historyStore.List(context.Background(), first, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if versions.Total != 3 {
		t.Fatalf("history versions=%d want=3", versions.Total)
	}
	actions := map[string]bool{}
	for _, version := range versions.Data {
		actions[version.Action] = true
	}
	for _, action := range []string{"tag_add", "group_set", "group_clear"} {
		if !actions[action] {
			t.Fatalf("missing history action %q in %#v", action, versions.Data)
		}
	}

	rows, err := db.Query(`SELECT details_json FROM audit_events WHERE event_type LIKE 'profile.bulk_%'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var details string
		if err := rows.Scan(&details); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(details, "proxy.example.invalid") || strings.Contains(details, "batch") || strings.Contains(details, "T24 Group") || strings.Contains(details, "secret") {
			t.Fatalf("bulk audit leaked detail: %s", details)
		}
	}
}

func TestT24BulkRejectsInvalidInputBeforeMutation(t *testing.T) {
	r, _, store, _, _ := testBulkRouter(t)
	id := createBulkProfile(t, store, "T24 Invalid")
	for _, body := range []string{
		`{"operation":"unknown","profile_ids":["` + id + `"]}`,
		`{"operation":"archive","profile_ids":["` + id + `","` + id + `"]}`,
		`{"operation":"set_group","profile_ids":["` + id + `"],"group":"missing"}`,
		`{"operation":"add_tag","profile_ids":["` + id + `"],"tag":"bad\u007fvalue"}`,
	} {
		_, rec := callBulk(t, r, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid bulk accepted status=%d body=%s", rec.Code, rec.Body.String())
		}
		assertErrorCode(t, rec, "INVALID_BULK_REQUEST")
		p, err := store.Get(id)
		if err != nil || p.LifecycleState != profile.LifecycleActive || len(p.Tags) != 0 || p.Group != "" {
			t.Fatalf("invalid bulk mutated profile=%+v err=%v", p, err)
		}
	}
}

func TestT24BulkRequiresAuthLoopbackAndOrigin(t *testing.T) {
	r, _, store, _, _ := testBulkRouter(t)
	id := createBulkProfile(t, store, "T24 Guard")
	body := `{"operation":"archive","profile_ids":["` + id + `"]}`

	missingToken := newLoopbackRequest(http.MethodPost, "/api/profiles/bulk", strings.NewReader(body))
	missingToken.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, missingToken)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status=%d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "UNAUTHORIZED")

	remote := newBulkRequest(body)
	remote.RemoteAddr = "203.0.113.10:4141"
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, remote)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("remote status=%d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "LOOPBACK_REQUIRED")

	noOrigin := newBulkRequest(body)
	noOrigin.Header.Del("Origin")
	noOrigin.Header.Del("Referer")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, noOrigin)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("no origin status=%d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "ORIGIN_REQUIRED_LOCAL_ONLY")
}

func TestT24BulkContextCancellationAndConcurrentArchive(t *testing.T) {
	r, _, store, _, historyStore := testBulkRouter(t)
	cancelled := createBulkProfile(t, store, "T24 Cancelled")
	active := createBulkProfile(t, store, "T24 Concurrent")

	req := newBulkRequest(`{"operation":"archive","profile_ids":["` + cancelled + `"]}`)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var cancellation bulkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &cancellation); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || cancellation.Data.Summary.Failed != 1 || cancellation.Data.Results[0].Code != "BULK_CANCELLED" {
		t.Fatalf("cancel response=%d %#v", rec.Code, cancellation)
	}
	p, err := store.Get(cancelled)
	if err != nil || p.LifecycleState != profile.LifecycleActive {
		t.Fatalf("cancelled target mutated: %+v err=%v", p, err)
	}

	var wg sync.WaitGroup
	responses := make([]*httptest.ResponseRecorder, 2)
	for i := range responses {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			responses[index] = httptest.NewRecorder()
			r.ServeHTTP(responses[index], newBulkRequest(`{"operation":"archive","profile_ids":["`+active+`"]}`))
		}(i)
	}
	wg.Wait()
	changed, noop := 0, 0
	for _, response := range responses {
		if response.Code != http.StatusOK {
			t.Fatalf("concurrent status=%d body=%s", response.Code, response.Body.String())
		}
		var got bulkResponse
		if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		changed += got.Data.Summary.Changed
		noop += got.Data.Summary.Noop
	}
	if changed != 1 || noop != 1 {
		t.Fatalf("concurrent totals changed=%d noop=%d", changed, noop)
	}
	versions, err := historyStore.List(context.Background(), active, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if versions.Total != 1 || len(store.PendingHistoryProfiles()) != 0 {
		t.Fatalf("history total=%d pending=%d, want one confirmed bulk capture", versions.Total, len(store.PendingHistoryProfiles()))
	}
}
