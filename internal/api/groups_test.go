package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"browseforge/internal/groups"
	"browseforge/internal/profile"

	"github.com/go-chi/chi/v5"
)

func TestGroupProxyAPI(t *testing.T) {
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{groupStore: groupStore}

	req := requestWithGroupName(http.MethodPut, "/api/groups/Client%20A", "Client A", strings.NewReader(`{
		"proxy_mode": "enforced",
		"proxy": {"type": "socks5", "host": "proxy.example.com", "port": 1080}
	}`))
	rec := httptest.NewRecorder()
	h.upsertGroup(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert status = %d body=%s", rec.Code, rec.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	data := result["data"].(map[string]any)
	if data["name"] != "Client A" || data["proxy_mode"] != groups.ProxyModeEnforced {
		t.Fatalf("group response = %#v", data)
	}

	req = requestWithGroupName(http.MethodDelete, "/api/groups/Client%20A/proxy", "Client A", nil)
	rec = httptest.NewRecorder()
	h.clearGroupProxy(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("clear status = %d body=%s", rec.Code, rec.Body.String())
	}
	if g, ok := groupStore.Get("Client A"); ok {
		t.Fatalf("cleared group still exists = %+v", g)
	}
}

func TestGroupProxyAPIRequiresProxyOnPut(t *testing.T) {
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := &handler{groupStore: groupStore}

	req := requestWithGroupName(http.MethodPut, "/api/groups/Client%20A", "Client A", strings.NewReader(`{"proxy_mode":"default"}`))
	rec := httptest.NewRecorder()
	h.upsertGroup(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "MISSING_PROXY") {
		t.Fatalf("body missing MISSING_PROXY: %s", rec.Body.String())
	}
}

func TestDeleteGroupUngroupsProfilesAndClearsProxy(t *testing.T) {
	profileStore, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := &profile.Profile{Name: "Profile A", Engine: "firefox", Group: "Client A"}
	if err := profileStore.Create(p); err != nil {
		t.Fatal(err)
	}
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := groupStore.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy.example.com", Port: 1080}, groups.ProxyModeDefault); err != nil {
		t.Fatal(err)
	}
	h := &handler{store: profileStore, groupStore: groupStore}

	req := requestWithGroupName(http.MethodDelete, "/api/groups/Client%20A", "Client A", nil)
	rec := httptest.NewRecorder()
	h.deleteGroup(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := profileStore.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Group != "" {
		t.Fatalf("profile group = %q, want empty", updated.Group)
	}
	if g, ok := groupStore.Get("Client A"); ok {
		t.Fatalf("group proxy still exists = %+v", g)
	}
	if !strings.Contains(rec.Body.String(), `"profiles_ungrouped":1`) {
		t.Fatalf("response missing ungroup count: %s", rec.Body.String())
	}
}

func TestBackupIncludesGroupPolicies(t *testing.T) {
	profileStore, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	groupStore, err := groups.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := groupStore.Upsert("Client A", &profile.ProxyConfig{Type: "socks5", Host: "proxy.example.com", Port: 1080}, groups.ProxyModeEnforced); err != nil {
		t.Fatal(err)
	}
	h := &handler{store: profileStore, groupStore: groupStore}

	rec := httptest.NewRecorder()
	h.backup(rec, httptest.NewRequest(http.MethodPost, "/api/backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("backup status = %d body=%s", rec.Code, rec.Body.String())
	}

	zr, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name != "groups.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "proxy.example.com") {
			t.Fatalf("groups.json missing policy: %s", data)
		}
		return
	}
	t.Fatal("groups.json not found in backup")
}

func requestWithGroupName(method, target, name string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("name", name)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}
