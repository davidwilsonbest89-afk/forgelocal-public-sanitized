package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"forgelocal/internal/browser"
	"forgelocal/internal/config"
	"forgelocal/internal/profile"
)

func TestReadOnlyBootstrapIsLoopbackOnlySingleUseAndScopeLimited(t *testing.T) {
	store, err := profile.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{DataDir: t.TempDir(), Version: "test-core"}
	router, err := NewRouter(cfg, store, &browser.Manager{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	issue := httptest.NewRequest(http.MethodPost, "/api/v1/readonly/session/codes", nil)
	issue.RemoteAddr = "127.0.0.1:41234"
	issue.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	issue.Header.Set("Origin", "http://localhost:3000")
	issued := httptest.NewRecorder()
	router.ServeHTTP(issued, issue)
	if issued.Code != http.StatusCreated {
		t.Fatalf("issue status=%d body=%s", issued.Code, issued.Body.String())
	}
	var codeReply struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(issued.Body.Bytes(), &codeReply); err != nil || len(codeReply.Code) != 64 {
		t.Fatalf("invalid issued code: %q err=%v", codeReply.Code, err)
	}

	payload, _ := json.Marshal(map[string]string{"code": codeReply.Code})
	exchange := httptest.NewRequest(http.MethodPost, "/api/v1/readonly/session/bootstrap", bytes.NewReader(payload))
	exchange.RemoteAddr = "127.0.0.1:41235"
	exchange.Header.Set("Origin", "http://localhost:3000")
	bootstrapped := httptest.NewRecorder()
	router.ServeHTTP(bootstrapped, exchange)
	if bootstrapped.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapped.Code, bootstrapped.Body.String())
	}
	var sessionReply struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(bootstrapped.Body.Bytes(), &sessionReply); err != nil || len(sessionReply.Token) != 64 {
		t.Fatalf("invalid session token: %q err=%v", sessionReply.Token, err)
	}

	readonly := httptest.NewRequest(http.MethodGet, "/api/v1/readonly/health", nil)
	readonly.Header.Set("Authorization", "Bearer "+sessionReply.Token)
	readonlyAllowed := httptest.NewRecorder()
	router.ServeHTTP(readonlyAllowed, readonly)
	if readonlyAllowed.Code != http.StatusOK {
		t.Fatalf("readonly token status=%d body=%s", readonlyAllowed.Code, readonlyAllowed.Body.String())
	}

	writeAttempt := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewBufferString(`{"name":"must-not-create"}`))
	writeAttempt.Header.Set("Authorization", "Bearer "+sessionReply.Token)
	writeAttempt.Header.Set("Origin", "http://localhost:3000")
	writeDenied := httptest.NewRecorder()
	router.ServeHTTP(writeDenied, writeAttempt)
	if writeDenied.Code != http.StatusUnauthorized {
		t.Fatalf("readonly token unexpectedly authorized write: status=%d body=%s", writeDenied.Code, writeDenied.Body.String())
	}

	replay := httptest.NewRequest(http.MethodPost, "/api/v1/readonly/session/bootstrap", bytes.NewReader(payload))
	replay.RemoteAddr = "127.0.0.1:41236"
	replay.Header.Set("Origin", "http://localhost:3000")
	replayed := httptest.NewRecorder()
	router.ServeHTTP(replayed, replay)
	if replayed.Code != http.StatusUnauthorized {
		t.Fatalf("replayed code status=%d body=%s", replayed.Code, replayed.Body.String())
	}

	// Remote origin: the request now fails at originGuard (403 ORIGIN_REQUIRED_LOCAL_ONLY)
	// before requireLoopbackMiddleware; the effective refusal of the mutation is unchanged.
	remote := httptest.NewRequest(http.MethodPost, "/api/v1/readonly/session/bootstrap", bytes.NewReader(payload))
	remote.RemoteAddr = "203.0.113.10:41237"
	remote.Header.Set("Origin", "https://203.0.113.10")
	nonLoopback := httptest.NewRecorder()
	router.ServeHTTP(nonLoopback, remote)
	if nonLoopback.Code != http.StatusForbidden {
		t.Fatalf("remote bootstrap status=%d body=%s", nonLoopback.Code, nonLoopback.Body.String())
	}
}

func TestReadOnlySessionBrokerExpiresCodesAndTokens(t *testing.T) {
	now := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	broker := newReadOnlySessionBroker()
	broker.now = func() time.Time { return now }
	code, _, err := broker.issueCode()
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(readOnlyBootstrapCodeTTL + time.Second)
	if _, _, accepted, err := broker.exchangeCode(code); err != nil || accepted {
		t.Fatalf("expired code accepted=%v err=%v", accepted, err)
	}
	now = now.Add(-readOnlyBootstrapCodeTTL - time.Second)
	code, _, err = broker.issueCode()
	if err != nil {
		t.Fatal(err)
	}
	token, _, accepted, err := broker.exchangeCode(code)
	if err != nil || !accepted {
		t.Fatalf("valid code accepted=%v err=%v", accepted, err)
	}
	now = now.Add(readOnlySessionTokenTTL + time.Second)
	if broker.validateToken(token) {
		t.Fatal("expired read-only token was accepted")
	}
}
