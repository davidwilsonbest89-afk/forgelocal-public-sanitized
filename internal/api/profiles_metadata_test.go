package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProfileMetadataEndpointAuditsRedactedShape(t *testing.T) {
	r, db, _ := testWriteRouter(t)
	id := createTestProfile(t, r, "Metadata API", "cloakbrowser")
	body := `{"note":"private but non-secret","custom_fields":{"tier":{"type":"select","value":"gold","options":["silver","gold"]},"approved":{"type":"boolean","value":true}}}`
	rec := httptest.NewRecorder()
	req := newLoopbackRequest(http.MethodPut, "/api/profiles/"+id+"/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("metadata status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(correlationHeader); got == "" {
		t.Fatal("missing correlation id")
	}
	checkAuditEvent(t, db, "profile.metadata_updated", id, rec.Header().Get(correlationHeader))
	if strings.Contains(rec.Body.String(), "proxy_secret_ref") {
		t.Fatalf("response leaked secret ref: %s", rec.Body.String())
	}
	assertMetadataAuditIsRedacted(t, db, id, "private but non-secret", "gold")
}

func TestProfileMetadataEndpointRejectsInvalidAndGenericBypass(t *testing.T) {
	r, _, _ := testWriteRouter(t)
	id := createTestProfile(t, r, "Metadata Reject", "cloakbrowser")
	for _, tc := range []struct{ method, path, body, code string }{
		{http.MethodPut, "/api/profiles/" + id + "/metadata", `{"custom_fields":{"x":{"type":"select","value":"bad","options":["ok"]}}}`, "INVALID_PROFILE"},
		{http.MethodPut, "/api/profiles/" + id, `{"note":"bypass"}`, "METADATA_ENDPOINT_REQUIRED"},
	} {
		rec := httptest.NewRecorder()
		req := newLoopbackRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Fatalf("%s accepted: %s", tc.path, rec.Body.String())
		}
		assertErrorCode(t, rec, tc.code)
	}
	// Creation must not become a second metadata write path without audit.
	rec := httptest.NewRecorder()
	req := newLoopbackRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Bypass","runtime_id":"cloakbrowser","note":"bypass"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assertErrorCode(t, rec, "METADATA_ENDPOINT_REQUIRED")
}

func assertMetadataAuditIsRedacted(t *testing.T, db *sql.DB, id, forbiddenNote, forbiddenValue string) {
	t.Helper()
	var details string
	if err := db.QueryRow(`SELECT details_json FROM audit_events WHERE event_type = ? AND entity_id = ? ORDER BY id DESC LIMIT 1`, "profile.metadata_updated", id).Scan(&details); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(details, forbiddenNote) || strings.Contains(details, forbiddenValue) {
		t.Fatalf("audit leaked metadata value: %s", details)
	}
}
