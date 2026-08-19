package profile

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
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
	LifecycleState  LifecycleState  `json:"lifecycle_state,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	LastUsed        time.Time       `json:"last_used"`
	Fingerprint     map[string]any  `json:"fingerprint,omitempty"`
	FingerprintSeed uint32          `json:"fingerprint_seed,omitempty"` // CloakBrowser seed
	Proxy           *ProxyConfig    `json:"proxy,omitempty"`
	ContainerID     string          `json:"container_id,omitempty"`
	ProfileDir      string          `json:"profile_dir"`
	// Note and CustomFields are non-sensitive profile metadata. They are never
	// projected by the dashboard read-only catalogue; authenticated metadata
	// writes go through the Core-only T20-NCF contract and produce redacted audit.
	Note         string                 `json:"note,omitempty"`
	CustomFields map[string]CustomField `json:"custom_fields,omitempty"`
}

// CustomField is a typed, non-secret profile metadata value. Select values
// carry their allowed options so the Core can validate values deterministically.
type CustomField struct {
	Type    string   `json:"type"`
	Value   any      `json:"value"`
	Options []string `json:"options,omitempty"`
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
	// locksByName is the uniqueness budget for profile names, guarded by mu.
	locksByName map[string]string // name -> owning profile id
	// perProfile is the lazy per-profile isolation map used by T09 write
	// operations. Its own mu is held only long enough to look up or register
	// a mutex; the per-profile mutex is held only by a single writer at a time.
	perProfileMu sync.Mutex
	perProfile   map[string]*sync.Mutex
}

const maxProfileTags = 20

const (
	maxProfileNoteBytes   = 4096
	maxCustomFieldNameLen = 64
	maxCustomFields       = 20
	maxCustomFieldText    = 2048
	maxSelectOptions      = 20
)

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
	s := &Store{dir: dir, profiles: make(map[string]*Profile), vault: vault, locksByName: make(map[string]string), perProfile: make(map[string]*sync.Mutex)}
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
		s.locksByName[normalizeName(p.Name)] = p.ID
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
	if _, exists := s.profiles[p.ID]; exists {
		return ErrDuplicateID
	}
	if err := validateV2Profile(p, "new profile"); err != nil {
		return err
	}
	if owner, taken := s.locksByName[normalizeName(p.Name)]; taken && owner != p.ID {
		return ErrDuplicateName
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
	s.locksByName[normalizeName(p.Name)] = p.ID
	return nil
}

func validateV2Profile(p *Profile, source string) error {
	if p == nil {
		return fmt.Errorf("%s is nil", source)
	}
	if p.RuntimeID == "" {
		return fmt.Errorf("%s uses v1 profile schema: runtime_id is required; run BrowseForge migrate profiles --from v1 --to v2", source)
	}
	return validateProfileInputs(p)
}

// validateProfileInputs is the T09 input contract: names must be short
// printable strings, lifecycle state must belong to the enum, proxy fields
// must agree with each other, and the tag budget is enforced.
func validateProfileInputs(p *Profile) error {
	if !validName(p.Name) {
		return ErrInvalidName
	}
	if p.LifecycleState == "" {
		p.LifecycleState = LifecycleActive
	}
	if !ValidLifecycleState(p.LifecycleState) {
		return fmt.Errorf("%w: %s", ErrInvalidName, p.LifecycleState)
	}
	if err := validateProxyShape(p); err != nil {
		return err
	}
	if len(p.Tags) > maxProfileTags {
		return ErrTooManyTags
	}
	for _, tag := range p.Tags {
		if !validName(tag) {
			return ErrInvalidTag
		}
	}
	if err := validateProfileMetadata(p.Note, p.CustomFields); err != nil {
		return err
	}
	return nil
}

func validateProfileMetadata(note string, fields map[string]CustomField) error {
	if len(note) > maxProfileNoteBytes || !validNote(note) {
		return ErrInvalidNote
	}
	if len(fields) > maxCustomFields {
		return ErrTooManyCustomFields
	}
	for name, field := range fields {
		if name == "" || len(name) > maxCustomFieldNameLen || !validName(name) {
			return ErrInvalidCustomField
		}
		if err := validateCustomField(field); err != nil {
			return err
		}
	}
	return nil
}

// ValidateTemplateMetadata reuses the canonical Profile/T20-NCF syntax and
// value rules for a Template payload without creating or changing a profile.
// Nil values mean that the corresponding field is absent from the template.
func ValidateTemplateMetadata(group *string, tags []string, note *string, fields map[string]CustomField) error {
	if group != nil && *group != "" && !validName(*group) {
		return ErrInvalidGroup
	}
	if len(tags) > maxProfileTags {
		return ErrTooManyTags
	}
	seenTags := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if !validName(tag) {
			return ErrInvalidTag
		}
		key := normalizeName(tag)
		if _, duplicate := seenTags[key]; duplicate {
			return ErrAlreadyTagged
		}
		seenTags[key] = struct{}{}
	}
	if note != nil {
		return validateProfileMetadata(*note, fields)
	}
	return validateProfileMetadata("", fields)
}

// CanonicalMetadataName is the case-insensitive comparison rule used by the
// Profile Store. Templates use it for names and tag union so their behaviour
// cannot silently diverge from existing Core metadata.
func CanonicalMetadataName(value string) string { return normalizeName(value) }

func validNote(value string) bool {
	for _, c := range value {
		if c == '\n' || c == '\t' {
			continue
		}
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

func validateCustomField(field CustomField) error {
	switch field.Type {
	case "text":
		value, ok := field.Value.(string)
		if !ok || len(value) > maxCustomFieldText || !validName(value) || len(field.Options) != 0 {
			return ErrInvalidCustomField
		}
	case "number":
		value, ok := field.Value.(float64)
		if !ok || math.IsNaN(value) || math.IsInf(value, 0) || len(field.Options) != 0 {
			return ErrInvalidCustomField
		}
	case "boolean":
		if _, ok := field.Value.(bool); !ok || len(field.Options) != 0 {
			return ErrInvalidCustomField
		}
	case "select":
		value, ok := field.Value.(string)
		if !ok || len(field.Options) == 0 || len(field.Options) > maxSelectOptions {
			return ErrInvalidCustomField
		}
		seen := make(map[string]struct{}, len(field.Options))
		found := false
		for _, option := range field.Options {
			if !validName(option) {
				return ErrInvalidCustomField
			}
			key := normalizeName(option)
			if _, duplicate := seen[key]; duplicate {
				return ErrInvalidCustomField
			}
			seen[key] = struct{}{}
			if option == value {
				found = true
			}
		}
		if !found {
			return ErrInvalidCustomField
		}
	default:
		return ErrInvalidCustomField
	}
	return nil
}

func validName(value string) bool {
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

// validateProxyShape enforces the proxy consistency rule: an empty proxy is
// allowed, and any configured proxy must carry a port in the 1..65535 range
// with a type that is either empty (protocol-agnostic) or one of the
// supported transports. This preserves the pre-T09 historical contract where
// a host and port without an explicit type remain valid.
func validateProxyShape(p *Profile) error {
	if p.Proxy == nil {
		return nil
	}
	px := p.Proxy
	if px.Type == "" && px.Host == "" && px.Port == 0 {
		return nil
	}
	if px.Type != "" && px.Type != "http" && px.Type != "socks5" {
		return ErrInvalidProxy
	}
	if px.Port < 0 || px.Port > 65535 {
		return ErrInvalidProxy
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

// WithProfile isolates a single writer per profile. The returned unlock
// function releases the per-profile lock and MUST be deferred by the caller.
// Concurrent mutations on the same profile serialize through this lock, and
// each lock acquisition is bounded by the isolation budget.
func (s *Store) WithProfile(id string, budget time.Duration) (unlock func(), err error) {
	s.perProfileMu.Lock()
	mu, known := s.perProfile[id]
	if !known {
		mu = &sync.Mutex{}
		s.perProfile[id] = mu
	}
	s.perProfileMu.Unlock()

	done := make(chan struct{})
	go func() {
		mu.Lock()
		close(done)
	}()
	select {
	case <-done:
		return func() { mu.Unlock() }, nil
	case <-time.After(budget):
		return nil, ErrLocked
	}
}

const perProfileIsolationBudget = 5 * time.Second

func (s *Store) Update(id string, updates map[string]any) (*Profile, error) {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	// Archive prevents any further mutation; updates to an archived profile are
	// refused so the dormant state stays observable.
	if p.LifecycleState == LifecycleArchived || p.LifecycleState == LifecycleQuarantined {
		s.mu.Unlock()
		return nil, ErrNotArchived // routed to INVALID_LIFECYCLE by the handler mapper
	}
	s.mu.Unlock()

	unlock, err := s.WithProfile(id, perProfileIsolationBudget)
	if err != nil {
		return nil, err
	}
	defer unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// Re-check existence and lifecycle after acquiring the store lock.
	if p, ok = s.profiles[id]; !ok {
		return nil, ErrNotFound
	}
	if p.LifecycleState != LifecycleActive {
		return nil, ErrNotArchived
	}
	// Only the currently loaded profile may retain its own vault reference;
	// client-provided update payloads must not select another profile's secret.
	hadProxySecret := p.Proxy != nil && p.Proxy.SecretRef == proxySecretRef(p.ID)
	// Keep the normalized name before the merge so the uniqueness check
	// compares the new name against the profile's actual previous name,
	// not against the already-merged value.
	previousName := normalizeName(p.Name)
	// Mutations are applied to a disposable copy first: an invalid or
	// refused update must never leave the in-memory profile half-changed,
	// since callers hold the store pointer and read it directly.
	tmp := new(Profile)
	*tmp = *p
	data, err := json.Marshal(tmp)
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
	if err := json.Unmarshal(merged, tmp); err != nil {
		return nil, err
	}
	if err := validateV2Profile(tmp, "updated profile"); err != nil {
		return nil, err
	}
	// A name change must keep the global name budget.
	if name, ok := updates["name"].(string); ok && normalizeName(name) != previousName {
		if owner, taken := s.locksByName[normalizeName(name)]; taken && owner != id {
			return nil, ErrDuplicateName
		}
		s.locksByName[normalizeName(name)] = id
		delete(s.locksByName, previousName)
	}
	if tmp.Proxy != nil {
		if tmp.Proxy.Username != "" || tmp.Proxy.Password != "" || hadProxySecret {
			tmp.Proxy.SecretRef = proxySecretRef(id)
		} else {
			tmp.Proxy.SecretRef = ""
		}
	}
	if err := s.restoreProxySecret(tmp); err != nil {
		return nil, err
	}
	if err := s.persistProxySecret(tmp); err != nil {
		return nil, err
	}
	tmp.ID = id
	tmp.ProfileDir = p.ProfileDir
	s.profiles[id] = tmp
	return tmp, s.save(tmp)

}

// SetMetadata is the sole Core mutation for notes and custom fields. It uses
// the same per-profile isolation and lifecycle guard as other profile writes,
// and copies caller-owned maps before persistence.
func (s *Store) SetMetadata(id, note string, fields map[string]CustomField) (*Profile, error) {
	s.mu.RLock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrNotFound
	}
	if p.LifecycleState != LifecycleActive {
		s.mu.RUnlock()
		return nil, ErrNotArchived
	}
	s.mu.RUnlock()

	unlock, err := s.WithProfile(id, perProfileIsolationBudget)
	if err != nil {
		return nil, err
	}
	defer unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok = s.profiles[id]
	if !ok {
		return nil, ErrNotFound
	}
	if p.LifecycleState != LifecycleActive {
		return nil, ErrNotArchived
	}
	cloned, err := cloneCustomFields(fields)
	if err != nil {
		return nil, err
	}
	tmp := *p
	tmp.Note = note
	tmp.CustomFields = cloned
	if err := validateProfileMetadata(tmp.Note, tmp.CustomFields); err != nil {
		return nil, err
	}
	if err := s.save(&tmp); err != nil {
		return nil, err
	}
	s.profiles[id] = &tmp
	return &tmp, nil
}

// RestoreHistory applies a sanitized immutable T22 snapshot to an active
// profile. It preserves current vault credentials when a restored proxy remains
// configured; snapshots themselves never contain vault references or values.
func (s *Store) RestoreHistory(id string, restored *Profile) (*Profile, error) {
	if restored == nil || restored.ID != id {
		return nil, ErrNotFound
	}
	s.mu.RLock()
	current, ok := s.profiles[id]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrNotFound
	}
	if current.LifecycleState != LifecycleActive {
		s.mu.RUnlock()
		return nil, ErrNotArchived
	}
	s.mu.RUnlock()

	unlock, err := s.WithProfile(id, perProfileIsolationBudget)
	if err != nil {
		return nil, err
	}
	defer unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok = s.profiles[id]
	if !ok {
		return nil, ErrNotFound
	}
	if current.LifecycleState != LifecycleActive {
		return nil, ErrNotArchived
	}
	data, err := json.Marshal(restored)
	if err != nil {
		return nil, err
	}
	var tmp Profile
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}
	tmp.ID = id
	tmp.ProfileDir = current.ProfileDir
	tmp.CreatedAt = current.CreatedAt
	if tmp.Proxy != nil && current.Proxy != nil && current.Proxy.SecretRef == proxySecretRef(id) {
		tmp.Proxy.SecretRef = current.Proxy.SecretRef
		tmp.Proxy.Username = current.Proxy.Username
		tmp.Proxy.Password = current.Proxy.Password
	}
	if err := validateV2Profile(&tmp, "history restore"); err != nil {
		return nil, err
	}
	previousName := normalizeName(current.Name)
	if nextName := normalizeName(tmp.Name); nextName != previousName {
		if owner, taken := s.locksByName[nextName]; taken && owner != id {
			return nil, ErrDuplicateName
		}
		s.locksByName[nextName] = id
		delete(s.locksByName, previousName)
	}
	if err := s.save(&tmp); err != nil {
		return nil, err
	}
	s.profiles[id] = &tmp
	return &tmp, nil
}

func cloneCustomFields(fields map[string]CustomField) (map[string]CustomField, error) {
	if fields == nil {
		return nil, nil
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return nil, ErrInvalidCustomField
	}
	var clone map[string]CustomField
	if err := json.Unmarshal(b, &clone); err != nil {
		return nil, ErrInvalidCustomField
	}
	return clone, nil
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
	if locksByNameName, taken := s.locksByName[normalizeName(p.Name)]; taken && locksByNameName == id {
		delete(s.locksByName, normalizeName(p.Name))
	}
	delete(s.profiles, id)
	delete(s.perProfile, id)
	return nil
}

// normalizeName folds a profile name to its uniqueness key (lower-cased,
// trimmed). Comparison remains case-insensitive, matching the SQLite NOCASE
// collation used for the names table column.
func normalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ArchiveProfile transitions an active profile to the archived state. Archived
// profiles keep their directory and vault entry but refuse mutations.
func (s *Store) ArchiveProfile(id string) error {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	if p.LifecycleState == LifecycleArchived {
		s.mu.Unlock()
		return nil // idempotent: the profile is already in the target state
	}
	if p.LifecycleState == LifecycleQuarantined {
		s.mu.Unlock()
		return ErrQuarantined
	}
	unlock, err := s.WithProfile(id, perProfileIsolationBudget)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	// The per-profile lock must serialize every write for this profile; the
	// store lock cannot be released before the archive state is flushed.
	defer s.mu.Unlock()
	defer unlock()
	p.LifecycleState = LifecycleArchived
	return s.save(p)
}

// ReopenProfile returns an archived profile to the active state. Quarantined
// profiles refuse reopening: only an external authority can lift quarantine.
func (s *Store) ReopenProfile(id string) error {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	if p.LifecycleState == LifecycleActive {
		s.mu.Unlock()
		return ErrAlreadyArchived // reused sentinel meaning "not archived"; mapper routes to INVALID_LIFECYCLE
	}
	if p.LifecycleState == LifecycleQuarantined {
		s.mu.Unlock()
		return ErrQuarantined
	}
	unlock, err := s.WithProfile(id, perProfileIsolationBudget)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	defer s.mu.Unlock()
	defer unlock()
	p.LifecycleState = LifecycleActive
	return s.save(p)
}

// AddProfileTag assigns a tag to a profile. The tag budget and active state
// are enforced; quarantined or archived profiles refuse new tags.
func (s *Store) AddProfileTag(id string, tag string) error {
	if !validName(tag) {
		return ErrInvalidTag
	}
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	if p.LifecycleState != LifecycleActive {
		s.mu.Unlock()
		return fmt.Errorf("%w: profile is %s", ErrNotArchived, p.LifecycleState)
	}
	if contains(p.Tags, tag) {
		s.mu.Unlock()
		return ErrAlreadyTagged
	}
	if len(p.Tags) >= maxProfileTags {
		s.mu.Unlock()
		return ErrTooManyTags
	}
	p.Tags = append(p.Tags, tag)
	if err := s.save(p); err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()
	return nil
}

// RemoveProfileTag unassigns a tag from a profile. Archived profiles refuse
// tag changes so their observable state stays stable.
func (s *Store) RemoveProfileTag(id string, tag string) error {
	if !validName(tag) {
		return ErrInvalidTag
	}
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	if p.LifecycleState != LifecycleActive {
		s.mu.Unlock()
		return fmt.Errorf("%w: profile is %s", ErrNotArchived, p.LifecycleState)
	}
	if !contains(p.Tags, tag) {
		s.mu.Unlock()
		return ErrTagNotAssigned
	}
	p.Tags = removeTag(p.Tags, tag)
	if err := s.save(p); err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()
	return nil
}

func removeTag(tags []string, tag string) []string {
	kept := make([]string, 0, len(tags))
	for _, t := range tags {
		if t != tag {
			kept = append(kept, t)
		}
	}
	return kept
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

// quarantineForTest is a controlled test-only transition; production leaves
// quarantine to the external authority flow and never exposes this path.
func (s *Store) quarantineForTest(id string) error {
	s.mu.Lock()
	p, ok := s.profiles[id]
	if !ok {
		s.mu.Unlock()
		return ErrNotFound
	}
	p.LifecycleState = LifecycleQuarantined
	s.mu.Unlock()
	return nil
}

// ArchiveQuarantinedForTest moves a profile into the quarantined state for
// cross-package write-contract tests. Production quarantines only through
// the external authority flow; this helper exists so the API layer can
// assert that quarantined profiles refuse reopen without reusing the
// package-private helper.
func (s *Store) ArchiveQuarantinedForTest(id string) error {
	return s.quarantineForTest(id)
}
