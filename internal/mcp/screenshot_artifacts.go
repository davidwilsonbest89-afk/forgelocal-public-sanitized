package mcp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const defaultScreenshotURLTTL = 10 * time.Minute

type screenshotArtifact struct {
	data      []byte
	mimeType  string
	ext       string
	expiresAt time.Time
}

type screenshotArtifactStore struct {
	mu    sync.Mutex
	items map[string]screenshotArtifact
}

func newScreenshotArtifactStore() *screenshotArtifactStore {
	return &screenshotArtifactStore{items: map[string]screenshotArtifact{}}
}

func (s *screenshotArtifactStore) save(data []byte, mimeType, ext string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, item := range s.items {
		if !item.expiresAt.After(now) {
			delete(s.items, id)
		}
	}
	id, err := randomHex(16)
	if err != nil {
		return "", time.Time{}, err
	}
	s.items[id] = screenshotArtifact{data: data, mimeType: mimeType, ext: ext, expiresAt: expiresAt}
	return id, expiresAt, nil
}

func (s *screenshotArtifactStore) get(id string) (screenshotArtifact, bool) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return screenshotArtifact{}, false
	}
	if !item.expiresAt.After(now) {
		delete(s.items, id)
		return screenshotArtifact{}, false
	}
	return item, true
}

func (s *Server) ServeScreenshotArtifact(w http.ResponseWriter, r *http.Request) {
	id := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if id == "." || id == "/" || strings.ContainsAny(id, `/\\`) {
		http.NotFound(w, r)
		return
	}
	store := s.screenshotArtifactStore()
	item, ok := store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", item.mimeType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	filename := "screenshot-" + id + item.ext
	w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	http.ServeContent(w, r, filename, item.expiresAt, bytes.NewReader(item.data))
}

func (s *Server) screenshotArtifactStore() *screenshotArtifactStore {
	if s.screenshotArtifacts == nil {
		s.screenshotArtifacts = newScreenshotArtifactStore()
	}
	return s.screenshotArtifacts
}

func parseScreenshotTTL(args map[string]any) time.Duration {
	seconds := 600.0
	if raw, ok := args["url_ttl_seconds"].(float64); ok && raw > 0 {
		seconds = raw
	}
	if seconds < 30 {
		seconds = 30
	}
	if seconds > 3600 {
		seconds = 3600
	}
	return time.Duration(seconds * float64(time.Second))
}

func randomHex(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
