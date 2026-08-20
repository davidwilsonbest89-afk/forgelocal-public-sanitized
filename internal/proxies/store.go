// T10 — Proxies. File-based proxy registry store.
//
// The registry owns reusable proxy entries (http/socks5 endpoints identified
// by name) that profiles may reference. Secrets are never stored as values:
// every proxy with credentials carries an opaque `secret_ref` matching the
// project vault reference grammar (proxy.ref.*) and the credential material
// is only ever touched through a secrets.SecretVault by explicit test-only
// or system paths. List responses and API payloads never expose credential
// values; they expose secret_ref alone.
//
// Concurrency follows the T09 profile contract: a per-proxy mutex with a
// bounded isolation budget so concurrent mutations on the same proxy serialize
// and never stall indefinitely.
//
// No proxy network activity, browser, runtime, Camoufox, provider integration
// or release activity lives in this package.
package proxies

import (
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Proxy is a reusable endpoint entry in the local proxy registry. Credential
// fields deliberately use json:"-" so that values never appear in any JSON
// projection (list, get, profile mapping). The opaque SecretRef is the only
// credential-adjacent field exposed, matching the vault grammar
// proxy.ref.<id>.
type Proxy struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "http" | "socks5"
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	SecretRef string    `json:"secret_ref,omitempty"`
	Region    string    `json:"region,omitempty"`
	HasSecret bool      `json:"has_secret"` // redacted presence flag only
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is the proxy registry. The backing file format mirrors the file-based
// profile and group stores so the registry stays a local-first artifact with
// no SQLite dependency from the write contract.
type Store struct {
	path string
	mu   sync.RWMutex
	// byName is the case-insensitive uniqueness index maintained lazily
	// (rebuildIndex) so Create/Update keep a single source of truth.
	byName map[string]string
	// perProxy is the lazy per-proxy isolation map used by write operations.
	// Its own mu is held only long enough to look up or register a mutex.
	perProxyMu sync.Mutex
	perProxy   map[string]*sync.Mutex
	proxies    map[string]*Proxy
	// assigned tracks profile->proxy assignments so deletions cannot silently
	// detach profiles (ErrProxyInUse).
	assigned map[string]string // profile id -> proxy id
}

const perProxyIsolationBudget = 5 * time.Second

var secretRefPattern = regexp.MustCompile(`^proxy\.ref\.[A-Za-z0-9._-]{1,128}$`)

// NewStore creates or loads the proxy registry at dataDir/proxies.json.
func NewStore(dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{
		path:     filepath.Join(dataDir, "proxies.json"),
		proxies:  make(map[string]*Proxy),
		assigned: make(map[string]string),
		perProxy: make(map[string]*sync.Mutex),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	s.rebuildIndex()
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.saveLocked()
		}
		return err
	}
	var payload struct {
		Proxies  []*Proxy        `json:"proxies"`
		Assigned map[string]string `json:"assigned,omitempty"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	for _, p := range payload.Proxies {
		// Reject any persisted entry whose secret_ref no longer matches the
		// vault grammar: tampered or migrated entries must not silently pass.
		if p.SecretRef != "" && !secretRefPattern.MatchString(p.SecretRef) {
			continue
		}
		s.proxies[p.ID] = p
	}
	if payload.Assigned != nil {
		// Only assignments whose proxy still exists are restored.
		for profileID, proxyID := range payload.Assigned {
			if _, known := s.proxies[proxyID]; known {
				s.assigned[profileID] = proxyID
			}
		}
	}
	return nil
}

func (s *Store) saveLocked() error {
	payload := struct {
		Proxies  []*Proxy        `json:"proxies"`
		Assigned map[string]string `json:"assigned,omitempty"`
	}{Proxies: make([]*Proxy, 0, len(s.proxies)), Assigned: nil}
	for _, p := range s.proxies {
		payload.Proxies = append(payload.Proxies, p)
	}
	if len(s.assigned) > 0 {
		payload.Assigned = s.assigned
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// #nosec G306 -- registry file is owner-only readable (vault-adjacent data).
	return os.WriteFile(s.path, data, 0600)
}

// Create registers a validated proxy. The id is generated when empty; an
// explicit id with an already-used value is refused (ErrDuplicateName applies
// to names, ErrDuplicateID to ids).
func (s *Store) Create(p *Proxy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateProxyInputs(p); err != nil {
		return err
	}
	if p.ID == "" {
		id, err := generateID()
		if err != nil {
			return err
		}
		p.ID = id
	}
	if _, exists := s.proxies[p.ID]; exists {
		return ErrDuplicateID
	}
	if _, taken := s.byName[normalizeName(p.Name)]; taken {
		return ErrDuplicateName
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	s.proxies[p.ID] = p
	s.byName[normalizeName(p.Name)] = p.ID
	return s.saveLocked()
}

// byName is maintained lazily (see rebuildIndex) to keep Create/Update simple.
func (s *Store) rebuildIndex() {
	s.byName = make(map[string]string, len(s.proxies))
	for _, p := range s.proxies {
		s.byName[normalizeName(p.Name)] = p.ID
	}
}

// Get returns a proxy copy. The copy keeps credential fields excluded by the
// struct tags; callers of Get receive the same redacted view as the API.
func (s *Store) Get(id string) (*Proxy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.proxies[id]
	if !ok {
		return nil, ErrNotFound
	}
	return clone(p), nil
}

// List returns all registered proxies sorted by name (redacted view).
func (s *Store) List() []*Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Proxy, 0, len(s.proxies))
	for _, p := range s.proxies {
		out = append(out, clone(p))
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}

// Update applies validated field changes. An update cannot move a proxy that
// is assigned to profiles (ErrProxyInUse protects profile consistency).
func (s *Store) Update(id string, updates map[string]any) (*Proxy, error) {
	s.mu.Lock()
	p, ok := s.proxies[id]
	if !ok {
		s.mu.Unlock()
		return nil, ErrNotFound
	}
	s.mu.Unlock()

	unlock, err := s.WithProxy(id, perProxyIsolationBudget)
	if err != nil {
		return nil, err
	}
	defer unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok = s.proxies[id]; !ok {
		return nil, ErrNotFound
	}
	// A proxy assigned to profiles is read-only: unassign first.
	if s.isAssignedToAny(id) {
		return nil, ErrProxyInUse
	}
	tmp := clone(p)
	mergeInto(tmp, updates)
	if err := validateProxyInputs(tmp); err != nil {
		return nil, err
	}
	if normalizeName(tmp.Name) != normalizeName(p.Name) {
		if _, taken := s.byName[normalizeName(tmp.Name)]; taken {
			return nil, ErrDuplicateName
		}
		delete(s.byName, normalizeName(p.Name))
		s.byName[normalizeName(tmp.Name)] = id
	}
	tmp.ID = id
	tmp.UpdatedAt = time.Now()
	s.proxies[id] = tmp
	return tmp, s.saveLocked()
}

// Delete removes a proxy. Assigned proxies are refused (ErrProxyInUse).
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.proxies[id]
	if !ok {
		return ErrNotFound
	}
	for _, proxyID := range s.assigned {
		if proxyID == id {
			return ErrProxyInUse
		}
	}
	delete(s.proxies, id)
	delete(s.byName, normalizeName(p.Name))
	delete(s.perProxy, id)
	return s.saveLocked()
}

// Assign binds a profile to the proxy, replacing any previous assignment of
// that profile. The proxy must exist; profile existence is enforced by the
// profile store at the API boundary (this registry only tracks assignments).
// Re-assigning the same pair is a no-op.
func (s *Store) Assign(profileID, proxyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.proxies[proxyID]; !ok {
		return ErrNotFound
	}
	if s.assigned[profileID] == proxyID {
		return nil // idempotent
	}
	s.assigned[profileID] = proxyID
	return s.saveLocked()
}

// Unassign detaches a profile from its proxy. Re-detaching is a no-op.
func (s *Store) Unassign(profileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, known := s.assigned[profileID]; !known {
		return nil // idempotent
	}
	delete(s.assigned, profileID)
	return s.saveLocked()
}

// UnassignFor detaches a profile from its proxy only when the profile is
// actually bound to the given proxy id. A mismatch (profile bound to a
// different proxy, or not bound at all) is refused so an assignment cannot be
// silently redirected away from the proxy the caller targeted.
func (s *Store) UnassignFor(profileID, proxyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, known := s.assigned[profileID]
	if !known || current != proxyID {
		return ErrNotFound
	}
	delete(s.assigned, profileID)
	return s.saveLocked()
}

// AssignedProxy returns the proxy bound to the profile, or nil when none.
func (s *Store) AssignedProxy(profileID string) *Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proxyID, known := s.assigned[profileID]
	if !known {
		return nil
	}
	p, ok := s.proxies[proxyID]
	if !ok {
		return nil
	}
	return clone(p)
}

// AssignedProfiles lists the profile ids bound to the proxy (count-only view).
func (s *Store) AssignedProfiles(proxyID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for profileID, pid := range s.assigned {
		if pid == proxyID {
			out = append(out, profileID)
		}
	}
	sort.Strings(out)
	return out
}

// WithProxy isolates a single writer per proxy. The returned unlock function
// releases the per-proxy lock and MUST be deferred by the caller.
func (s *Store) WithProxy(id string, budget time.Duration) (unlock func(), err error) {
	s.perProxyMu.Lock()
	mu, known := s.perProxy[id]
	if !known {
		mu = &sync.Mutex{}
		s.perProxy[id] = mu
	}
	s.perProxyMu.Unlock()

	done := make(chan struct{})
	go func() {
		mu.Lock()
		close(done)
	}()
	select {
	case <-done:
		return func() { mu.Unlock() }, nil
	case <-time.After(budget):
		return nil, ErrProxyLocked
	}
}

func (s *Store) isAssignedToAny(proxyID string) bool {
	for _, pid := range s.assigned {
		if pid == proxyID {
			return true
		}
	}
	return false
}

func normalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// validateProxyInputs enforces the T10 registry contract: short printable
// name, explicit http/socks5 type, non-empty host, port 1..65535 and an
// optional short printable region. Credential fields are presence-checked
// only; their values are never inspected, stored as values or logged.
func validateProxyInputs(p *Proxy) error {
	if p == nil {
		return ErrInvalidProxy
	}
	if !validShortString(p.Name) {
		return ErrInvalidName
	}
	switch p.Type {
	case "http", "socks5":
	default:
		return ErrInvalidType
	}
	if !validShortString(p.Host) {
		return ErrInvalidHost
	}
	if p.Port < 1 || p.Port > 65535 {
		return ErrInvalidPort
	}
	if p.Region != "" && !validShortString(p.Region) {
		return ErrInvalidRegion
	}
	// A present secret_ref must match the vault grammar; malformed references
	// (tampered, leaked or legacy values) are refused at the boundary.
	if p.SecretRef != "" && !secretRefPattern.MatchString(p.SecretRef) {
		return ErrInvalidProxy
	}
	return nil
}

func validShortString(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

func clone(p *Proxy) *Proxy {
	c := *p
	return &c
}

func mergeInto(p *Proxy, updates map[string]any) {
	if v, ok := updates["name"].(string); ok {
		p.Name = v
	}
	if v, ok := updates["type"].(string); ok {
		p.Type = v
	}
	if v, ok := updates["host"].(string); ok {
		p.Host = v
	}
	if v, ok := updates["port"]; ok {
		switch n := v.(type) {
		case int:
			p.Port = n
		case float64:
			p.Port = int(n)
		}
	}
	if v, ok := updates["region"].(string); ok {
		p.Region = v
	}
	if v, ok := updates["secret_ref"].(string); ok {
		p.SecretRef = v
		p.HasSecret = v != ""
	}
	if v, ok := updates["has_secret"].(bool); ok {
		p.HasSecret = v
	}
}

// generateID produces a random opaque identifier for new registry entries.
func generateID() (string, error) {
	return generateRandomID(16)
}

// generateRandomID returns a lowercase hex identifier of the requested byte length.
func generateRandomID(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := readRandom(buf); err != nil {
		return "", err
	}
	return hexEncode(buf), nil
}

// readRandom is overridable in tests; it wraps crypto/rand.Read.
var readRandom = func(buf []byte) (int, error) {
	return rand.Read(buf)
}

// hexEncode returns the lowercase hex representation of buf.
func hexEncode(buf []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(buf)*2)
	for i, b := range buf {
		out[i*2] = hex[b>>4]
		out[i*2+1] = hex[b&0x0f]
	}
	return string(out)
}
