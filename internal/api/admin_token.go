package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	adminTokenMetadataVersion = 1
	adminTokenLifetime        = 15 * time.Minute
	adminTokenMetadataName    = ".api-token.meta"
)

var (
	ErrAdminTokenMissing   = errors.New("admin token missing")
	ErrAdminTokenMalformed = errors.New("admin token malformed")
	ErrAdminTokenInvalid   = errors.New("admin token invalid")
	ErrAdminTokenExpired   = errors.New("admin token expired")
	ErrAdminTokenRevoked   = errors.New("admin token revoked")
)

type adminTokenMetadata struct {
	Version     int        `json:"version"`
	TokenSHA256 string     `json:"token_sha256"`
	IssuedAt    time.Time  `json:"issued_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

type adminTokenState struct {
	mu       sync.RWMutex
	token    string
	metaPath string
	metadata adminTokenMetadata
	now      func() time.Time
}

func newAdminTokenState(dataDir, token string) (*adminTokenState, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("initialize admin token: %w", ErrAdminTokenInvalid)
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create admin token metadata directory: %w", err)
	}
	s := &adminTokenState{
		token:    token,
		metaPath: filepath.Join(dataDir, adminTokenMetadataName),
		now:      func() time.Time { return time.Now().UTC() },
	}
	data, err := os.ReadFile(s.metaPath)
	if err == nil {
		if err := json.Unmarshal(data, &s.metadata); err != nil {
			return nil, fmt.Errorf("read admin token metadata: %w", err)
		}
		if s.metadata.Version != adminTokenMetadataVersion || s.metadata.TokenSHA256 != digestToken(token) || s.metadata.IssuedAt.IsZero() || s.metadata.ExpiresAt.IsZero() || !s.metadata.ExpiresAt.After(s.metadata.IssuedAt) {
			return nil, fmt.Errorf("invalid admin token metadata: %w", ErrAdminTokenInvalid)
		}
		return s, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read admin token metadata: %w", err)
	}
	now := s.now().UTC()
	s.metadata = adminTokenMetadata{
		Version:     adminTokenMetadataVersion,
		TokenSHA256: digestToken(token),
		IssuedAt:    now,
		ExpiresAt:   now.Add(adminTokenLifetime),
	}
	if err := s.persistLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *adminTokenState) validateAuthorization(auth string) error {
	if strings.TrimSpace(auth) == "" {
		return ErrAdminTokenMissing
	}
	if !strings.HasPrefix(auth, "Bearer ") || strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")) == "" {
		return ErrAdminTokenMalformed
	}
	presented := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if subtle.ConstantTimeCompare([]byte(presented), []byte(s.token)) != 1 {
		return ErrAdminTokenInvalid
	}
	if s.metadata.RevokedAt != nil {
		return ErrAdminTokenRevoked
	}
	if !s.now().UTC().Before(s.metadata.ExpiresAt) {
		return ErrAdminTokenExpired
	}
	return nil
}

func (s *adminTokenState) presentedTokenMatches(auth string) bool {
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	presented := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	s.mu.RLock()
	defer s.mu.RUnlock()
	return subtle.ConstantTimeCompare([]byte(presented), []byte(s.token)) == 1
}

func (s *adminTokenState) revoke() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metadata.RevokedAt == nil {
		now := s.now().UTC()
		s.metadata.RevokedAt = &now
	}
	return s.persistLocked()
}

func (s *adminTokenState) persistLocked() error {
	data, err := json.MarshalIndent(s.metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode admin token metadata: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.metaPath), ".api-token.meta-*")
	if err != nil {
		return fmt.Errorf("create admin token metadata temporary file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod admin token metadata: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write admin token metadata: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close admin token metadata: %w", err)
	}
	if err := os.Rename(tmpName, s.metaPath); err != nil {
		return fmt.Errorf("commit admin token metadata: %w", err)
	}
	return nil
}

func digestToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func (h *handler) validateAdminAuthorization(auth string) error {
	if h.adminToken != nil {
		return h.adminToken.validateAuthorization(auth)
	}
	if validBearerToken(auth, h.token) {
		return nil
	}
	if strings.TrimSpace(auth) == "" {
		return ErrAdminTokenMissing
	}
	if !strings.HasPrefix(auth, "Bearer ") || strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")) == "" {
		return ErrAdminTokenMalformed
	}
	return ErrAdminTokenInvalid
}

func writeAdminAuthError(w http.ResponseWriter, err error) {
	reason, message := "invalid", "invalid admin token"
	switch {
	case errors.Is(err, ErrAdminTokenMissing):
		reason, message = "missing", "admin token required"
	case errors.Is(err, ErrAdminTokenMalformed):
		reason, message = "malformed", "malformed admin token"
	case errors.Is(err, ErrAdminTokenExpired):
		reason, message = "expired", "admin token expired"
	case errors.Is(err, ErrAdminTokenRevoked):
		reason, message = "revoked", "admin token revoked"
	}
	// Keep the established UNAUTHORIZED code for existing clients while the
	// stable reason field distinguishes the R2 token states without revealing
	// the presented or stored token value.
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]string{"code": "UNAUTHORIZED", "message": message, "reason": reason}})
}

func (h *handler) revokeAdminToken(w http.ResponseWriter, _ *http.Request) {
	if h.adminToken == nil {
		writeError(w, http.StatusNotImplemented, "TOKEN_REVOCATION_UNAVAILABLE", "admin token revocation unavailable")
		return
	}
	if err := h.adminToken.revoke(); err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_REVOCATION_FAILED", "admin token revocation failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func newAdminTokenValue() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
