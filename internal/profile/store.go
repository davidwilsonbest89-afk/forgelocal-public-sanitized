package profile

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"forgelocal/internal/secrets"
	"time"
)

type Profile struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	RuntimeID       string          `json:"runtime_id"` // concrete runtime provider, e.g. "camoufox" | "cloakbrowser"
	Identity        *IdentityConfig `json:"identity,omitempty"`
	Group           string          `json:"group,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	LastUsed        time.Time       `json:"last_used"`
	Fingerprint     map[string]any  `json:"fingerprint,omitempty"`
	FingerprintSeed uint32          `json:"fingerprint_seed,omitempty"` // CloakBrowser seed
	Proxy           *ProxyConfig    `json:"proxy,omitempty"`
	ContainerID     string          `json:"container_id,omitempty"`
	ProfileDir      string          `json:"profile_dir"`
}

type IdentityConfig struct {
	TargetOS     string `json:"target_os,omitempty"` // macos | windows | linux
	BrowserMajor int    `json:"browser_major,omitempty"`
	GPUVendor    string `json:"gpu_vendor,omitempty"`
	GPURenderer  string `json:"gpu_renderer,omitempty"`
	FontPack     string `json:"font_pack,omitempty"`
	RiskAccepted bool   `json:"risk_accepted,omitempty"`
}

type ProxyConfig struct {
	Type      string `json:"type"` // "socks5" | "http"
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"-"`
	Password  string `json:"-"`
	SecretRef string `json:"secret_ref,omitempty"`
	Region    string `json:"region,omitempty"`
}

type Store struct {
	dir      string
	mu       sync.RWMutex
	profiles map[string]*Profile
	vault    secrets.SecretVault
}

func NewStore(dir string, vaults ...secrets.SecretVault) (*Store, error) {
	var vault secrets.SecretVault
	if len(vaults) > 0 {
		vault = vaults[0]
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	// #nosec G302 -- directories require owner execute permission; 0700 denies all group/other access.
	if err := os.Chmod(dir, 0700); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, profiles: make(map[string]*Profile), vault: vault}
	return s, s.loadAll()
}

func (s *Store) loadAll() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(s.dir, e.Name(), "profile.json")
		// #nosec G304 -- path is built from a non-symlink directory entry enumerated under Store.dir.
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var p Profile
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		if err := validateV2Profile(&p, path); err != nil {
			return err
		}
		if err := s.restoreProxySecret(&p); err != nil {
			return err
		}
		s.profiles[p.ID] = &p
	}
	return nil
}
func (s *Store) Create(p *Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.ID == "" {
		id, err := generateID()
		if err != nil {
			return err
		}
		p.ID = id
	}
	if err := validateV2Profile(p, "new profile"); err != nil {
		return err
	}
	p.CreatedAt = time.Now()
	p.LastUsed = p.CreatedAt
	p.ProfileDir = filepath.Join(s.dir, p.ID)
	// A profile may only ever refer to its own vault entry. Imported metadata
	// must not be able to select an arbitrary secret reference.
	if p.Proxy != nil {
		if p.Proxy.Username != "" || p.Proxy.Password != "" {
			p.Proxy.SecretRef = proxySecretRef(p.ID)
		} else {
			p.Proxy.SecretRef = ""
		}
	}
	if err := s.persistProxySecret(p); err != nil {
		return err
	}

	if err := os.MkdirAll(p.ProfileDir, 0700); err != nil {
		return err
	}
	// #nosec G302 -- directories require owner execute permission; 0700 denies all group/other access.
	if err := os.Chmod(p.ProfileDir, 0700); err != nil {
		return err
	}
	browserDataDir := filepath.Join(p.ProfileDir, "browser-data")
	if err := os.MkdirAll(browserDataDir, 0700); err != nil {
		return err
	}
	// #nosec G302 -- directories require owner execute permission; 0700 denies all group/other access.
	if err := os.Chmod(browserDataDir, 0700); err != nil {
		return err
	}

	if err := s.save(p); err != nil {
		return err
	}
	s.profiles[p.ID] = p
	return nil
}

func validateV2Profile(p *Profile, source string) error {
	if p == nil {
		return fmt.Errorf("%s is nil", source)
	}
	if p.RuntimeID == "" {
		return fmt.Errorf("%s uses v1 profile schema: runtime_id is required; run BrowseForge migrate profiles --from v1 --to v2", source)
	}
	return nil
}

func (s *Store) Get(id string) (*Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	return p, nil
}

func (s *Store) List(group, tag string) []*Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Profile
	for _, p := range s.profiles {
		if group != "" && p.Group != group {
			continue
		}
		if tag != "" && !contains(p.Tags, tag) {
			continue
		}
		result = append(result, p)
	}
	return result
}

func (s *Store) Update(id string, updates map[string]any) (*Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	// Only the currently loaded profile may retain its own vault reference;
	// client-provided update payloads must not select another profile's secret.
	hadProxySecret := p.Proxy != nil && p.Proxy.SecretRef == proxySecretRef(p.ID)

	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for k, v := range updates {
		m[k] = v
	}
	merged, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(merged, p); err != nil {
		return nil, err
	}
	if err := validateV2Profile(p, "updated profile"); err != nil {
		return nil, err
	}
	if p.Proxy != nil {
		if p.Proxy.Username != "" || p.Proxy.Password != "" || hadProxySecret {
			p.Proxy.SecretRef = proxySecretRef(p.ID)
		} else {
			p.Proxy.SecretRef = ""
		}
	}
	if err := s.restoreProxySecret(p); err != nil {
		return nil, err
	}
	if err := s.persistProxySecret(p); err != nil {
		return nil, err
	}

	return p, s.save(p)

}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.profiles[id]
	if !ok {
		return fmt.Errorf("profile not found: %s", id)
	}
	if err := os.RemoveAll(p.ProfileDir); err != nil {
		return fmt.Errorf("remove profile data: %w", err)
	}
	if p.Proxy != nil && p.Proxy.SecretRef != "" && s.vault != nil {
		_ = s.vault.DeleteSecret(p.Proxy.SecretRef)
	}
	delete(s.profiles, id)
	return nil
}

func (s *Store) Duplicate(id string) (*Profile, error) {
	s.mu.RLock()
	src, ok := s.profiles[id]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", id)
	}

	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dup Profile
	if err := json.Unmarshal(data, &dup); err != nil {
		return nil, err
	}
	dup.ID = ""
	dup.Name = src.Name + " (copy)"
	dup.ContainerID = ""
	return &dup, s.Create(&dup)
}

func (s *Store) save(p *Profile) error {
	path := filepath.Join(p.ProfileDir, "profile.json")
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "prof_" + hex.EncodeToString(b), nil
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// ProfilePath returns the canonical on-disk path for a validated profile ID.
// Callers cannot supply an arbitrary filesystem path to the profile store.
func (s *Store) ProfilePath(id string) (string, error) {
	if !snapshotIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid profile id")
	}
	base, err := filepath.Abs(s.dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, id), nil
}

// UnmarshalJSON accepts credentials only at the Core boundary. MarshalJSON uses
// the json:"-" tags above, so credentials can never return to disk or API JSON.
func (p *Profile) UnmarshalJSON(data []byte) error {
	type plain Profile
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = Profile(decoded)
	var wire struct {
		Proxy *struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"proxy"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Proxy != nil {
		if p.Proxy == nil {
			p.Proxy = &ProxyConfig{}
		}
		p.Proxy.Username = wire.Proxy.Username
		p.Proxy.Password = wire.Proxy.Password
	}
	return nil
}

func proxySecretRef(profileID string) string { return "proxy." + profileID }

func (s *Store) persistProxySecret(p *Profile) error {
	if p == nil || p.Proxy == nil || (p.Proxy.Username == "" && p.Proxy.Password == "") {
		return nil
	}
	if s.vault == nil {
		return fmt.Errorf("proxy credentials require an OS secret vault")
	}
	// Never accept a caller-provided reference: each profile owns exactly one
	// deterministic vault slot.
	p.Proxy.SecretRef = proxySecretRef(p.ID)
	payload, err := json.Marshal(map[string]string{"username": p.Proxy.Username, "password": p.Proxy.Password})
	if err != nil {
		return err
	}
	return s.vault.PutSecret(p.Proxy.SecretRef, payload)
}

func (s *Store) restoreProxySecret(p *Profile) error {
	if p == nil || p.Proxy == nil || p.Proxy.SecretRef == "" || s.vault == nil {
		return nil
	}
	if p.Proxy.SecretRef != proxySecretRef(p.ID) {
		return fmt.Errorf("invalid proxy secret reference for profile %s", p.ID)
	}
	payload, err := s.vault.GetSecret(p.Proxy.SecretRef)
	if err != nil {
		return err
	}
	var values struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(payload, &values); err != nil {
		return fmt.Errorf("decode proxy secret: %w", err)
	}
	p.Proxy.Username = values.Username
	p.Proxy.Password = values.Password
	return nil
}
