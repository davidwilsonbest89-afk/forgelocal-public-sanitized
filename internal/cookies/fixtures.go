// Package cookies stores only synthetic QA fixture metadata. It never accesses
// a browser cookie jar and never persists an incoming fixture value.
package cookies

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	maxFixtures        = 20
	maxNameLen         = 64
	maxDomainLen       = 128
	maxPathLen         = 128
	maxFixtureValueLen = 128
)

var ErrInvalidFixture = errors.New("invalid synthetic cookie fixture")

// Input accepts only a fixture marker. Value is not part of Stored or Exported.
type Input struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Secure   bool   `json:"secure,omitempty"`
	HTTPOnly bool   `json:"http_only,omitempty"`
	SameSite string `json:"same_site,omitempty"`
}

type Fixture struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Path        string `json:"path"`
	Secure      bool   `json:"secure,omitempty"`
	HTTPOnly    bool   `json:"http_only,omitempty"`
	SameSite    string `json:"same_site,omitempty"`
	ValueDigest string `json:"value_digest"`
}

type diskState struct {
	Profiles map[string][]Fixture `json:"profiles"`
}

type Store struct {
	mu       sync.RWMutex
	path     string
	profiles map[string][]Fixture
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dataDir, "cookie-fixtures.json"), profiles: map[string][]Fixture{}}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var disk diskState
	if err := json.Unmarshal(data, &disk); err != nil {
		return nil, err
	}
	if disk.Profiles != nil {
		s.profiles = disk.Profiles
	}
	return s, nil
}

func (s *Store) Replace(profileID string, input []Input) ([]Fixture, error) {
	if profileID == "" || len(input) == 0 || len(input) > maxFixtures {
		return nil, ErrInvalidFixture
	}
	fixtures := make([]Fixture, 0, len(input))
	seen := map[string]struct{}{}
	for _, item := range input {
		fixture, err := validate(item)
		if err != nil {
			return nil, err
		}
		key := fixture.Domain + "\x00" + fixture.Path + "\x00" + fixture.Name
		if _, duplicate := seen[key]; duplicate {
			return nil, ErrInvalidFixture
		}
		seen[key] = struct{}{}
		fixtures = append(fixtures, fixture)
	}
	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].Domain+"\x00"+fixtures[i].Path+"\x00"+fixtures[i].Name < fixtures[j].Domain+"\x00"+fixtures[j].Path+"\x00"+fixtures[j].Name
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	next := cloneProfiles(s.profiles)
	next[profileID] = fixtures
	if err := s.persist(next); err != nil {
		return nil, err
	}
	s.profiles = next
	return cloneFixtures(fixtures), nil
}

func (s *Store) Export(profileID string) []Fixture {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneFixtures(s.profiles[profileID])
}

func (s *Store) persist(profiles map[string][]Fixture) error {
	data, err := json.MarshalIndent(diskState{Profiles: profiles}, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func validate(item Input) (Fixture, error) {
	item.Name = strings.TrimSpace(item.Name)
	item.Domain = strings.ToLower(strings.TrimSpace(item.Domain))
	item.Path = strings.TrimSpace(item.Path)
	item.SameSite = strings.ToLower(strings.TrimSpace(item.SameSite))
	if !validToken(item.Name, maxNameLen) || !validDomain(item.Domain) || !validPath(item.Path) || !strings.HasPrefix(item.Value, "fixture:") || len(item.Value) > maxFixtureValueLen || !validPrintable(item.Value) {
		return Fixture{}, ErrInvalidFixture
	}
	if item.SameSite != "" && item.SameSite != "lax" && item.SameSite != "strict" && item.SameSite != "none" {
		return Fixture{}, ErrInvalidFixture
	}
	digest := sha256.Sum256([]byte(item.Value))
	return Fixture{Name: item.Name, Domain: item.Domain, Path: item.Path, Secure: item.Secure, HTTPOnly: item.HTTPOnly, SameSite: item.SameSite, ValueDigest: hex.EncodeToString(digest[:])}, nil
}

func validToken(value string, max int) bool {
	if value == "" || len(value) > max {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func validDomain(value string) bool {
	if len(value) < 6 || len(value) > maxDomainLen || !strings.HasSuffix(value, ".test") || strings.Contains(value, "..") {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '-') {
			return false
		}
	}
	return true
}

func validPath(value string) bool {
	return len(value) > 0 && len(value) <= maxPathLen && strings.HasPrefix(value, "/") && validPrintable(value)
}
func validPrintable(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
func cloneFixtures(in []Fixture) []Fixture { return append([]Fixture(nil), in...) }
func cloneProfiles(in map[string][]Fixture) map[string][]Fixture {
	out := make(map[string][]Fixture, len(in))
	for id, fixtures := range in {
		out[id] = cloneFixtures(fixtures)
	}
	return out
}
