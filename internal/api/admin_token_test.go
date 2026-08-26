package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func configureTestAdminToken(t *testing.T, dir, token string, now time.Time) *adminTokenState {
	t.Helper()
	state, err := newAdminTokenState(dir, token)
	if err != nil {
		t.Fatal(err)
	}
	state.now = func() time.Time { return now }
	state.mu.Lock()
	state.metadata = adminTokenMetadata{
		Version:     adminTokenMetadataVersion,
		TokenSHA256: digestToken(token),
		IssuedAt:    now,
		ExpiresAt:   now.Add(adminTokenLifetime),
		RevokedAt:   nil,
	}
	if err := state.persistLocked(); err != nil {
		state.mu.Unlock()
		t.Fatal(err)
	}
	state.mu.Unlock()
	return state
}

func TestAdminTokenValidationDistinguishesStates(t *testing.T) {
	now := time.Date(2026, time.August, 25, 18, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		auth string
		want error
	}{
		{name: "missing", auth: "", want: ErrAdminTokenMissing},
		{name: "malformed", auth: "Basic synthetic", want: ErrAdminTokenMalformed},
		{name: "empty bearer", auth: "Bearer   ", want: ErrAdminTokenMalformed},
		{name: "invalid", auth: "Bearer wrong-synthetic-token", want: ErrAdminTokenInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := configureTestAdminToken(t, t.TempDir(), "admin-synthetic-token", now)
			if err := state.validateAuthorization(tc.auth); !errors.Is(err, tc.want) {
				t.Fatalf("validate error = %v, want %v", err, tc.want)
			}
		})
	}

	t.Run("valid at start", func(t *testing.T) {
		state := configureTestAdminToken(t, t.TempDir(), "admin-synthetic-token", now)
		if err := state.validateAuthorization("Bearer admin-synthetic-token"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("expired at exact boundary", func(t *testing.T) {
		dir := t.TempDir()
		state := configureTestAdminToken(t, dir, "admin-synthetic-token", now)
		state.now = func() time.Time { return now.Add(adminTokenLifetime) }
		if err := state.validateAuthorization("Bearer admin-synthetic-token"); !errors.Is(err, ErrAdminTokenExpired) {
			t.Fatalf("validate error = %v, want expired", err)
		}
	})
	t.Run("revoked", func(t *testing.T) {
		state := configureTestAdminToken(t, t.TempDir(), "admin-synthetic-token", now)
		if err := state.revoke(); err != nil {
			t.Fatal(err)
		}
		if err := state.validateAuthorization("Bearer admin-synthetic-token"); !errors.Is(err, ErrAdminTokenRevoked) {
			t.Fatalf("validate error = %v, want revoked", err)
		}
	})
}

func TestAdminTokenRevokePersistsAcrossRestartAndRedactsValue(t *testing.T) {
	dir := t.TempDir()
	token := "admin-synthetic-restart-token"
	now := time.Date(2026, time.August, 25, 18, 0, 0, 0, time.UTC)
	state := configureTestAdminToken(t, dir, token, now)
	if err := state.revoke(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, adminTokenMetadataName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), token) {
		t.Fatal("admin token value persisted in metadata")
	}
	if !strings.Contains(string(data), `"version": 1`) || !strings.Contains(string(data), `"revoked_at"`) {
		t.Fatalf("metadata missing version/revocation fields: %s", data)
	}
	restarted, err := newAdminTokenState(dir, token)
	if err != nil {
		t.Fatal(err)
	}
	if err := restarted.validateAuthorization("Bearer " + token); !errors.Is(err, ErrAdminTokenRevoked) {
		t.Fatalf("restart validation error = %v, want revoked", err)
	}
}

func TestAdminTokenMiddlewareAndRevokeEndpoint(t *testing.T) {
	now := time.Date(2026, time.August, 25, 18, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	token := "admin-synthetic-http-token"
	state := configureTestAdminToken(t, dir, token, now)
	h := &handler{token: token, adminToken: state}
	next := h.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	cases := []struct {
		name   string
		auth   string
		want   int
		code   string
		reason string
	}{
		{name: "missing", want: http.StatusUnauthorized, code: "UNAUTHORIZED", reason: "missing"},
		{name: "malformed", auth: "Basic synthetic", want: http.StatusUnauthorized, code: "UNAUTHORIZED", reason: "malformed"},
		{name: "invalid", auth: "Bearer invalid-synthetic", want: http.StatusUnauthorized, code: "UNAUTHORIZED", reason: "invalid"},
		{name: "valid", auth: "Bearer admin-synthetic-http-token", want: http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			req.Header.Set("Authorization", tc.auth)
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.want, rec.Body.String())
			}
			if tc.code != "" && !strings.Contains(rec.Body.String(), `"code":"`+tc.code+`"`) {
				t.Fatalf("body = %s, missing code %s", rec.Body.String(), tc.code)
			}
			if tc.reason != "" && !strings.Contains(rec.Body.String(), `"reason":"`+tc.reason+`"`) {
				t.Fatalf("body = %s, missing reason %s", rec.Body.String(), tc.reason)
			}
		})
	}

	revokeReq := httptest.NewRequest(http.MethodPost, "/api/auth/revoke", nil)
	revokeReq.Header.Set("Authorization", "Bearer "+token)
	revokeRec := httptest.NewRecorder()
	h.revokeAdminToken(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusNoContent {
		t.Fatalf("revoke status = %d, body=%s", revokeRec.Code, revokeRec.Body.String())
	}
	after := httptest.NewRecorder()
	next.ServeHTTP(after, revokeReq)
	if after.Code != http.StatusUnauthorized || !strings.Contains(after.Body.String(), `"code":"UNAUTHORIZED"`) || !strings.Contains(after.Body.String(), `"reason":"revoked"`) {
		t.Fatalf("after revoke status/body = %d/%s", after.Code, after.Body.String())
	}
	if strings.Contains(after.Body.String(), token) {
		t.Fatal("token value leaked in auth error")
	}
}

func TestAdminTokenConcurrentValidationAndRevocation(t *testing.T) {
	now := time.Date(2026, time.August, 25, 18, 0, 0, 0, time.UTC)
	state := configureTestAdminToken(t, t.TempDir(), "admin-synthetic-race-token", now)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = state.validateAuthorization("Bearer admin-synthetic-race-token")
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := state.revoke(); err != nil {
			t.Errorf("revoke: %v", err)
		}
	}()
	wg.Wait()
	if err := state.validateAuthorization("Bearer admin-synthetic-race-token"); !errors.Is(err, ErrAdminTokenRevoked) {
		t.Fatalf("final validation error = %v, want revoked", err)
	}
}

func TestLoadOrCreateTokenRejectsSymlinkedTokenFile(t *testing.T) {
	t.Setenv("BROWSEFORGE_TOKEN", "")
	dir := t.TempDir()
	want := "synthetic-bootstrap-token"
	if err := os.WriteFile(filepath.Join(dir, ".api-token"), []byte(want), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := loadOrCreateToken(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("loadOrCreateToken() = %q, want %q", got, want)
	}
	if err := os.Remove(filepath.Join(dir, ".api-token")); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(dir, "external-token")
	if err := os.WriteFile(external, []byte("external-token"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(dir, ".api-token")); err != nil {
		t.Fatal(err)
	}
	if _, err := loadOrCreateToken(dir); err == nil {
		t.Fatal("loadOrCreateToken() followed a symlinked token file")
	}
}
