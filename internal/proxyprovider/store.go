// Package proxyprovider implements a local simulated provider catalogue. It
// never stores credentials and has no network client by design.
package proxyprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidProvider = errors.New("invalid proxy provider")
	ErrDuplicate       = errors.New("proxy provider already exists")
	ErrNotFound        = errors.New("proxy provider not found")
)

type Provider struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SecretRef string `json:"secret_ref"`
	Mode      string `json:"mode"`
}

type Lease struct {
	ProviderID string `json:"provider_id"`
	ProfileID  string `json:"profile_id"`
	Region     string `json:"region"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	SecretRef  string `json:"secret_ref"`
	Mode       string `json:"mode"`
}

type Store struct {
	mu        sync.RWMutex
	path      string
	providers map[string]Provider
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dataDir, "proxy-providers.json"), providers: map[string]Provider{}}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.providers); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Register(input Provider) (Provider, error) {
	input.ID = strings.ToLower(strings.TrimSpace(input.ID))
	input.Name = strings.TrimSpace(input.Name)
	input.SecretRef = strings.TrimSpace(input.SecretRef)
	if !validID(input.ID) || !validName(input.Name) || input.SecretRef != "provider.ref."+input.ID {
		return Provider{}, ErrInvalidProvider
	}
	input.Mode = "simulated"
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.providers[input.ID]; exists {
		return Provider{}, ErrDuplicate
	}
	next := clone(s.providers)
	next[input.ID] = input
	if err := s.persist(next); err != nil {
		return Provider{}, err
	}
	s.providers = next
	return input, nil
}

func (s *Store) List() []Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Provider, 0, len(s.providers))
	for _, p := range s.providers {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (s *Store) SimulateResolve(providerID, profileID, region string) (Lease, error) {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	region = strings.ToLower(strings.TrimSpace(region))
	if !validID(providerID) || !validProfileID(profileID) || !validRegion(region) {
		return Lease{}, ErrInvalidProvider
	}
	s.mu.RLock()
	provider, ok := s.providers[providerID]
	s.mu.RUnlock()
	if !ok {
		return Lease{}, ErrNotFound
	}
	return Lease{ProviderID: provider.ID, ProfileID: profileID, Region: region, Type: "http", Host: fmt.Sprintf("%s.%s.provider.test", region, provider.ID), Port: 18080, SecretRef: provider.SecretRef, Mode: "simulated"}, nil
}

func (s *Store) persist(providers map[string]Provider) error {
	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
func clone(in map[string]Provider) map[string]Provider {
	out := make(map[string]Provider, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func validID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

func validProfileID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func validName(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
func validRegion(value string) bool { return validID(value) && strings.HasSuffix(value, "-test") }
