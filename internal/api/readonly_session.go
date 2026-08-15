package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	readOnlyBootstrapCodeTTL   = 10 * time.Minute
	readOnlySessionTokenTTL    = 15 * time.Minute
	readOnlyBootstrapBodyLimit = 1024
)

type readOnlySessionBroker struct {
	mu     sync.Mutex
	now    func() time.Time
	codes  map[string]time.Time
	tokens map[string]time.Time
}

func newReadOnlySessionBroker() *readOnlySessionBroker {
	return &readOnlySessionBroker{
		now:    time.Now,
		codes:  make(map[string]time.Time),
		tokens: make(map[string]time.Time),
	}
}

func (b *readOnlySessionBroker) issueCode() (string, time.Time, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked()
	code, err := newSessionSecret()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := b.now().Add(readOnlyBootstrapCodeTTL)
	b.codes[code] = expiresAt
	return code, expiresAt, nil
}

func (b *readOnlySessionBroker) exchangeCode(code string) (string, time.Time, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked()
	if _, ok := b.codes[code]; !ok {
		return "", time.Time{}, false, nil
	}
	delete(b.codes, code)
	token, err := newSessionSecret()
	if err != nil {
		return "", time.Time{}, false, err
	}
	expiresAt := b.now().Add(readOnlySessionTokenTTL)
	b.tokens[token] = expiresAt
	return token, expiresAt, true, nil
}

func (b *readOnlySessionBroker) validateToken(token string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked()
	_, ok := b.tokens[token]
	return ok
}

func (b *readOnlySessionBroker) pruneLocked() {
	now := b.now()
	for code, expiresAt := range b.codes {
		if !expiresAt.After(now) {
			delete(b.codes, code)
		}
	}
	for token, expiresAt := range b.tokens {
		if !expiresAt.After(now) {
			delete(b.tokens, token)
		}
	}
}

func newSessionSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *handler) issueReadOnlySessionCode(w http.ResponseWriter, r *http.Request) {
	if !isLoopbackRequest(r) {
		writeError(w, http.StatusForbidden, "LOOPBACK_REQUIRED", "read-only bootstrap codes are issued only on loopback")
		return
	}
	code, expiresAt, err := h.readonlySessions.issueCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_CODE_ISSUE_FAILED", "could not issue local read-only session code")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"code": code, "expires_at": expiresAt.UTC().Format(time.RFC3339), "scope": "readonly"})
}

func (h *handler) bootstrapReadOnlySession(w http.ResponseWriter, r *http.Request) {
	if !isLoopbackRequest(r) {
		writeError(w, http.StatusForbidden, "LOOPBACK_REQUIRED", "read-only session bootstrap is available only on loopback")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, readOnlyBootstrapBodyLimit)
	defer r.Body.Close()
	var input struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.Code) != 64 {
		writeError(w, http.StatusUnauthorized, "INVALID_BOOTSTRAP_CODE", "invalid or expired bootstrap code")
		return
	}
	token, expiresAt, accepted, err := h.readonlySessions.exchangeCode(input.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_BOOTSTRAP_FAILED", "could not establish read-only session")
		return
	}
	if !accepted {
		writeError(w, http.StatusUnauthorized, "INVALID_BOOTSTRAP_CODE", "invalid or expired bootstrap code")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "expires_at": expiresAt.UTC().Format(time.RFC3339), "scope": "readonly"})
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
