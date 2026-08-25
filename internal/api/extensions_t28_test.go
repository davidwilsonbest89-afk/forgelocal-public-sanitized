package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forgelocal/internal/extensions"
	"forgelocal/internal/profile"
	"github.com/go-chi/chi/v5"
)

func t28APIRouter(t *testing.T) (*chi.Mux, *extensions.Repository, *profile.Store) {
	t.Helper()
	repo, err := extensions.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	profiles, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{token: "t28-test-token", extensionStore: repo, store: profiles}
	r := chi.NewRouter()
	r.Use(originGuard)
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Use(h.requireLoopbackMiddleware)
		r.Post("/api/v1/extensions/import", h.importExtension)
		r.Get("/api/v1/extensions", h.listExtensions)
		r.Get("/api/v1/extensions/{seriesID}", h.getExtensionSeries)
		r.Post("/api/v1/extensions/{versionID}/approve", h.approveExtension)
		r.Post("/api/v1/extensions/{versionID}/assign", h.assignExtension)
	})
	return r, repo, profiles
}

func t28Request(r *http.Request) {
	r.RemoteAddr = "127.0.0.1:19280"
	r.Header.Set("Origin", "http://127.0.0.1:3000")
	r.Header.Set("Authorization", "Bearer t28-test-token")
}

func t28Multipart(t *testing.T, manifest string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("package", "local.zip")
	if err != nil {
		t.Fatal(err)
	}
	zipBytes := syntheticT28ZIP(t, map[string]string{"manifest.json": manifest, "payload.js": "must never execute"})
	if _, err := part.Write(zipBytes); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	t28Request(req)
	return req
}

func syntheticT28ZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := newTestZipWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Kept as a small indirection so this API test does not share any HTTP/runtime helper.
func newTestZipWriter(buf *bytes.Buffer) *zipWriter { return &zipWriter{buf: buf} }

type zipWriter struct {
	buf *bytes.Buffer
	zw  *zip.Writer
}

func (w *zipWriter) Create(name string) (io.Writer, error) {
	if w.zw == nil {
		w.zw = zip.NewWriter(w.buf)
	}
	return w.zw.Create(name)
}
func (w *zipWriter) Close() error { return w.zw.Close() }

func TestT28APIRejectsMissingAuthAndForeignOrigin(t *testing.T) {
	r, _, _ := t28APIRouter(t)
	missing := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/import", strings.NewReader("x"))
	missing.RemoteAddr = "127.0.0.1:19280"
	missing.Header.Set("Origin", "http://127.0.0.1:3000")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, missing)
	if resp.Code != http.StatusUnauthorized || !strings.Contains(resp.Body.String(), `"UNAUTHORIZED"`) {
		t.Fatalf("missing bearer was not refused: code=%d body=%s", resp.Code, resp.Body.String())
	}
	foreign := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/import", strings.NewReader("x"))
	foreign.RemoteAddr = "127.0.0.1:19280"
	foreign.Header.Set("Authorization", "Bearer t28-test-token")
	foreign.Header.Set("Origin", "https://foreign.example")
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, foreign)
	if resp.Code != http.StatusForbidden || !strings.Contains(resp.Body.String(), `"ORIGIN_REQUIRED_LOCAL_ONLY"`) {
		t.Fatalf("foreign origin was not refused: code=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestT28APIOffLoopbackRefused(t *testing.T) {
	r, _, _ := t28APIRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extensions", nil)
	req.RemoteAddr = "192.0.2.10:19280"
	req.Header.Set("Authorization", "Bearer t28-test-token")
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden || !strings.Contains(resp.Body.String(), `"LOOPBACK_REQUIRED"`) {
		t.Fatalf("off-loopback request was not refused: code=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestT28APIImportApproveListAndRedaction(t *testing.T) {
	r, repo, _ := t28APIRouter(t)
	manifest := `{"name":"Synthetic","version":"1","manifest_version":3,"permissions":["cookies","storage"],"host_permissions":["*://*/*"]}`
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, t28Multipart(t, manifest))
	if resp.Code != http.StatusCreated {
		t.Fatalf("import failed: code=%d body=%s", resp.Code, resp.Body.String())
	}
	var imported struct {
		Data extensions.Version `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &imported); err != nil {
		t.Fatal(err)
	}
	v := imported.Data
	if v.ID == "" || v.SeriesID == "" || v.DigestPreview == "" || v.RiskState != "HIGH_RISK" {
		t.Fatalf("unexpected import projection: %+v", v)
	}
	body, _ := json.Marshal(map[string]any{"permissions_acknowledged": []string{"cookies", "storage", "*://*/*"}})
	approve := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/"+v.ID+"/approve", bytes.NewReader(body))
	t28Request(approve)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, approve)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), `"HIGH_RISK_ACK_REQUIRED"`) {
		t.Fatalf("high-risk approval bypassed: code=%d body=%s", resp.Code, resp.Body.String())
	}
	body, _ = json.Marshal(map[string]any{"permissions_acknowledged": []string{"cookies", "storage", "*://*/*"}, "accept_high_risk": true})
	approve = httptest.NewRequest(http.MethodPost, "/api/v1/extensions/"+v.ID+"/approve", bytes.NewReader(body))
	t28Request(approve)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, approve)
	if resp.Code != http.StatusOK {
		t.Fatalf("approval failed: code=%d body=%s", resp.Code, resp.Body.String())
	}
	list := httptest.NewRequest(http.MethodGet, "/api/v1/extensions", nil)
	t28Request(list)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, list)
	if resp.Code != http.StatusOK || strings.Contains(resp.Body.String(), "blob_relpath") || strings.Contains(resp.Body.String(), "payload.js") {
		t.Fatalf("list projection leaked internal/package data: code=%d body=%s", resp.Code, resp.Body.String())
	}
	series := httptest.NewRequest(http.MethodGet, "/api/v1/extensions/"+v.SeriesID, nil)
	t28Request(series)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, series)
	if resp.Code != http.StatusOK || strings.Contains(resp.Body.String(), "blob_relpath") || strings.Contains(resp.Body.String(), "package") {
		t.Fatalf("detail projection leaked internal/package data: code=%d body=%s", resp.Code, resp.Body.String())
	}
	var auditCount int
	if err := repo.DB().QueryRow(`SELECT COUNT(*) FROM extension_audit_events`).Scan(&auditCount); err != nil || auditCount < 2 {
		t.Fatalf("audit was not written: count=%d err=%v", auditCount, err)
	}
	rows, err := repo.DB().Query(`SELECT action, permission_categories_json, profile_pseudonym, correlation_id FROM extension_audit_events`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var action, categories, pseudonym, correlation string
		if err := rows.Scan(&action, &categories, &pseudonym, &correlation); err != nil {
			t.Fatal(err)
		}
		joined := action + categories + pseudonym + correlation
		for _, forbidden := range []string{"local.zip", "payload.js", "must never execute", "Bearer", "cookie_value", "Set-Cookie"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("audit leaked forbidden marker %q: %s", forbidden, joined)
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestT28APIAssignmentRequiresApprovedVersionAndExistingProfile(t *testing.T) {
	r, _, profiles := t28APIRouter(t)
	if err := profiles.Create(&profile.Profile{Name: "T28 profile", RuntimeID: "camoufox"}); err != nil {
		t.Fatal(err)
	}
	profileID := profiles.List("", "")[0].ID
	manifest := `{"name":"Synthetic","version":"1"}`
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, t28Multipart(t, manifest))
	var imported struct {
		Data extensions.Version `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &imported); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"profile_id": "missing"})
	assign := httptest.NewRequest(http.MethodPost, "/api/v1/extensions/"+imported.Data.ID+"/assign", bytes.NewReader(body))
	t28Request(assign)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, assign)
	if resp.Code != http.StatusNotFound || !strings.Contains(resp.Body.String(), `"PROFILE_NOT_FOUND"`) {
		t.Fatalf("missing profile assignment was accepted: code=%d body=%s", resp.Code, resp.Body.String())
	}
	body, _ = json.Marshal(map[string]string{"profile_id": profileID})
	assign = httptest.NewRequest(http.MethodPost, "/api/v1/extensions/"+imported.Data.ID+"/assign", bytes.NewReader(body))
	t28Request(assign)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, assign)
	if resp.Code != http.StatusBadRequest || !strings.Contains(resp.Body.String(), `"VERSION_NOT_APPROVED"`) {
		t.Fatalf("unapproved version was assigned: code=%d body=%s", resp.Code, resp.Body.String())
	}
}
