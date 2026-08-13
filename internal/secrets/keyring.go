package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"sync"

	keyring "github.com/zalando/go-keyring"
)

const KeySize = 32

var (
	ErrNotFound   = errors.New("secret not found")
	ErrInvalidKey = errors.New("secret must be exactly 32 bytes")
	idPattern     = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
)

// Vault persists cryptographic keys outside SQLite, profile JSON, logs and backups.
type Vault interface {
	Put(id string, key []byte) error
	Get(id string) ([]byte, error)
	Delete(id string) error
}

func NewKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate backup key: %w", err)
	}
	return key, nil
}

func validate(id string, key []byte) error {
	if !idPattern.MatchString(id) {
		return fmt.Errorf("invalid key id")
	}
	if len(key) != KeySize {
		return ErrInvalidKey
	}
	return nil
}

// SystemVault uses Keychain, Credential Manager or Secret Service through go-keyring.
type SystemVault struct{ Service string }

func NewSystemVault(service string) *SystemVault {
	if service == "" {
		service = "ForgeLocal"
	}
	return &SystemVault{Service: service}
}

func (v *SystemVault) Put(id string, key []byte) error {
	if err := validate(id, key); err != nil {
		return err
	}
	return keyring.Set(v.Service, id, base64.RawStdEncoding.EncodeToString(key))
}

func (v *SystemVault) Get(id string) ([]byte, error) {
	if !idPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid key id")
	}
	encoded, err := keyring.Get(v.Service, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	key, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil || len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	return key, nil
}

func (v *SystemVault) Delete(id string) error {
	if !idPattern.MatchString(id) {
		return fmt.Errorf("invalid key id")
	}
	return keyring.Delete(v.Service, id)
}

// MemoryVault is restricted to tests and is never selected by production wiring.
type MemoryVault struct {
	mu   sync.RWMutex
	keys map[string][]byte
}

func NewMemoryVault() *MemoryVault { return &MemoryVault{keys: map[string][]byte{}} }

func (v *MemoryVault) Put(id string, key []byte) error {
	if err := validate(id, key); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.keys[id] = append([]byte(nil), key...)
	return nil
}

func (v *MemoryVault) Get(id string) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok := v.keys[id]
	if !ok {
		return nil, ErrNotFound
	}
	return append([]byte(nil), key...), nil
}

func (v *MemoryVault) Delete(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.keys, id)
	return nil
}
