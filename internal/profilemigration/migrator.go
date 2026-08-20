// Package profilemigration imports the legacy local JSON metadata into the
// ForgeLocal product schema. It never reads, writes, logs, or returns a secret;
// it only validates deterministic SystemVault references.
package profilemigration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"forgelocal/internal/backup"
	groupstore "forgelocal/internal/groups"
	"forgelocal/internal/profile"
)

var (
	ErrValidation            = errors.New("profile migration validation failed")
	ErrBackupRequired        = errors.New("encrypted pre-migration backup is required for apply mode")
	ErrExistingProductRecord = errors.New("product record already exists")
	// ErrInterruptedOperation identifies journal rows recovered after a crash or process interruption.
	ErrInterruptedOperation = errors.New("INTERRUPTED_BEFORE_COMPLETION")
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

type Mode string

const (
	// ModeDryRun validates and journals the import plan without importing profiles.
	ModeDryRun Mode = "dry-run"
	// ModeApply writes the schema only after a caller-provided encrypted backup succeeds.
	ModeApply Mode = "apply"
)

type Source struct {
	ProfilesDir string
	GroupsPath  string
}

type BackupReceipt struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

type BackupRequest struct {
	ProfilesDir   string            `json:"profiles_dir"`
	GroupsPath    string            `json:"groups_path"`
	SourceHashes  map[string]string `json:"source_hashes"`
	CorrelationID string            `json:"correlation_id"`
}

// BackupFunc is implemented by the Core integration. It must create and verify
// encrypted backups that cover every source named in BackupRequest.
type BackupFunc func(context.Context, BackupRequest) ([]BackupReceipt, error)

// NewEncryptedPreMigrationBackup adapts BACK-01 to the import contract. The
// created FLBK artifact contains only the validated local JSON metadata and is
// authenticated before the import transaction begins. Its plaintext is never
// logged or included in Report.
func NewEncryptedPreMigrationBackup(service *backup.Service, keyID string) BackupFunc {
	return func(ctx context.Context, request BackupRequest) ([]BackupReceipt, error) {
		if service == nil || !validIdentifier(keyID) {
			return nil, ErrBackupRequired
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		payload, err := buildEncryptedPreimagePayload(request)
		if err != nil {
			return nil, err
		}
		profileID := "migration." + sha256Hex([]byte(request.CorrelationID))[:24]
		created, err := service.Create(profileID, keyID, payload)
		if err != nil {
			return nil, fmt.Errorf("create encrypted pre-migration artifact: %w", err)
		}
		if err := service.Verify(created.ID); err != nil {
			return nil, fmt.Errorf("verify encrypted pre-migration artifact: %w", err)
		}
		return []BackupReceipt{{ID: created.ID, SHA256: created.SHA256}}, nil
	}
}

type Options struct {
	Mode          Mode
	CorrelationID string
	Backup        BackupFunc
	Now           func() time.Time
}

type Report struct {
	OperationID   string            `json:"operation_id"`
	CorrelationID string            `json:"correlation_id"`
	Mode          Mode              `json:"mode"`
	State         string            `json:"state"`
	Profiles      int               `json:"profiles"`
	Groups        int               `json:"groups"`
	Tags          int               `json:"tags"`
	Runtimes      int               `json:"runtimes"`
	Sources       int               `json:"sources"`
	SourceHashes  map[string]string `json:"source_hashes"`
	Backups       []BackupReceipt   `json:"backups,omitempty"`
	Parity        []Parity          `json:"parity"`
}

type Parity struct {
	ProfileID          string `json:"profile_id"`
	SourceSHA256       string `json:"source_sha256"`
	SQLiteRecordSHA256 string `json:"sqlite_record_sha256"`
	Result             string `json:"result"`
}

type Migrator struct {
	db     *sql.DB
	source Source
}

func New(db *sql.DB, source Source) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: sqlite database is required", ErrValidation)
	}
	if strings.TrimSpace(source.ProfilesDir) == "" {
		return nil, fmt.Errorf("%w: profiles directory is required", ErrValidation)
	}
	if strings.TrimSpace(source.GroupsPath) == "" {
		return nil, fmt.Errorf("%w: groups path is required", ErrValidation)
	}
	return &Migrator{db: db, source: source}, nil
}

// Run defaults to ModeDryRun. Apply mode requires a verified encrypted backup
// callback and performs all schema writes in a single SQLite transaction.
func buildEncryptedPreimagePayload(request BackupRequest) ([]byte, error) {
	expectedProfilesHash, ok := request.SourceHashes["profile_json"]
	if !ok || !isSHA256(expectedProfilesHash) {
		return nil, fmt.Errorf("%w: missing profiles source hash", ErrValidation)
	}
	groups := preimageFile{ID: "groups_json", Present: false}
	if expectedGroupsHash, present := request.SourceHashes["groups_json"]; present {
		if !isSHA256(expectedGroupsHash) {
			return nil, fmt.Errorf("%w: invalid groups source hash", ErrValidation)
		}
		var err error
		groups, err = readPreimageGroups(request.GroupsPath, expectedGroupsHash)
		if err != nil {
			return nil, err
		}
	}
	profiles, err := readPreimageProfiles(request.ProfilesDir, expectedProfilesHash)
	if err != nil {
		return nil, err
	}
	payload := struct {
		Version       int               `json:"version"`
		CorrelationID string            `json:"correlation_id"`
		SourceHashes  map[string]string `json:"source_hashes"`
		Groups        preimageFile      `json:"groups"`
		Profiles      []preimageFile    `json:"profiles"`
	}{
		Version: 1, CorrelationID: request.CorrelationID, SourceHashes: cloneStringMap(request.SourceHashes),
		Groups: groups, Profiles: profiles,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode encrypted pre-migration payload: %w", err)
	}
	if len(encoded) == 0 {
		return nil, fmt.Errorf("%w: empty pre-migration payload", ErrValidation)
	}
	return encoded, nil
}

type preimageFile struct {
	ID      string          `json:"id"`
	SHA256  string          `json:"sha256"`
	Data    json.RawMessage `json:"data,omitempty"`
	Present bool            `json:"present"`
}

func readPreimageGroups(path, expectedHash string) (preimageFile, error) {
	// #nosec G304 -- path is a local Core migration source; its SHA-256 is bound to the validated plan before backup creation.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if expectedHash != sha256Hex(nil) {
				return preimageFile{}, fmt.Errorf("%w: groups source changed before backup", ErrValidation)
			}
			return preimageFile{ID: "groups_json", SHA256: expectedHash, Present: false}, nil
		}
		return preimageFile{}, fmt.Errorf("%w: read groups preimage", ErrValidation)
	}
	if hasCleartextProxyCredentials(data) || sha256Hex(data) != expectedHash {
		return preimageFile{}, fmt.Errorf("%w: groups source changed or contains cleartext credential", ErrValidation)
	}
	if !json.Valid(data) {
		return preimageFile{}, fmt.Errorf("%w: invalid groups preimage JSON", ErrValidation)
	}
	return preimageFile{ID: "groups_json", SHA256: expectedHash, Data: append(json.RawMessage(nil), data...), Present: true}, nil
}

func readPreimageProfiles(root, expectedHash string) ([]preimageFile, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("%w: read profiles preimage directory", ErrValidation)
	}
	out := make([]preimageFile, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		if !validIdentifier(entry.Name()) {
			return nil, fmt.Errorf("%w: invalid profile preimage identifier", ErrValidation)
		}
		path := filepath.Join(root, entry.Name(), "profile.json")
		// #nosec G304 -- entry comes from os.ReadDir(root), rejects symlinks, and has a validated identifier before this controlled join.
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("%w: read profile preimage", ErrValidation)
		}
		if hasCleartextProxyCredentials(data) || !json.Valid(data) {
			return nil, fmt.Errorf("%w: profile preimage contains cleartext credential or invalid JSON", ErrValidation)
		}
		out = append(out, preimageFile{ID: entry.Name(), SHA256: sha256Hex(data), Data: append(json.RawMessage(nil), data...), Present: true})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	var digest strings.Builder
	for _, item := range out {
		digest.WriteString(item.ID)
		digest.WriteByte(':')
		digest.WriteString(item.SHA256)
		digest.WriteByte('\n')
	}
	if sha256Hex([]byte(digest.String())) != expectedHash {
		return nil, fmt.Errorf("%w: profile source changed before backup", ErrValidation)
	}
	return out, nil
}

func (m *Migrator) Run(ctx context.Context, options Options) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	mode := options.Mode
	if mode == "" {
		mode = ModeDryRun
	}
	if mode != ModeDryRun && mode != ModeApply {
		return Report{}, fmt.Errorf("%w: unsupported mode %q", ErrValidation, mode)
	}
	now := time.Now().UTC()
	if options.Now != nil {
		now = options.Now().UTC()
	}
	plan, err := m.buildPlan(now)
	if err != nil {
		return Report{Mode: mode, State: "failed"}, err
	}
	opID, err := newOperationID()
	if err != nil {
		return Report{}, err
	}
	correlationID := strings.TrimSpace(options.CorrelationID)
	if correlationID == "" {
		correlationID = "profile-migration:" + opID
	}
	report := plan.report(opID, correlationID, mode)

	if mode == ModeDryRun {
		if err := m.recordDryRun(ctx, plan, report, now); err != nil {
			return Report{}, err
		}
		return report, nil
	}
	if options.Backup == nil {
		return Report{}, ErrBackupRequired
	}
	if err := m.ensureNoExistingProductRecords(ctx, plan); err != nil {
		return Report{}, err
	}
	if err := m.startOperations(ctx, plan, report, now); err != nil {
		return Report{}, err
	}
	fail := func(cause error) (Report, error) {
		if err := m.finalizeOperations(context.Background(), report.OperationID, "failed", failureCode(cause), now); err != nil {
			return Report{}, fmt.Errorf("%w: finalize durable failure state: %v", cause, err)
		}
		return Report{}, cause
	}

	backups, err := options.Backup(ctx, BackupRequest{
		ProfilesDir:   plan.profilesDir,
		GroupsPath:    plan.groupsPath,
		SourceHashes:  cloneStringMap(plan.sourceHashes),
		CorrelationID: correlationID,
	})
	if err != nil {
		return fail(fmt.Errorf("create encrypted pre-migration backup: %w", err))
	}
	if err := validateReceipts(backups); err != nil {
		return fail(err)
	}
	report.Backups = append([]BackupReceipt(nil), backups...)

	if err := m.apply(ctx, plan, report, now); err != nil {
		return fail(err)
	}
	if err := m.finalizeOperations(context.Background(), report.OperationID, "committed", "", now); err != nil {
		return Report{}, err
	}
	return report, nil
}

type migrationPlan struct {
	profilesDir  string
	groupsPath   string
	groups       []preparedGroup
	profiles     []preparedProfile
	runtimes     []string
	tags         []preparedTag
	sourceHashes map[string]string
}

type preparedProxy struct {
	secretRef string
	kind      string
	host      string
	port      int
	region    string
}

type preparedGroup struct {
	id        string
	name      string
	proxyMode string
	proxy     preparedProxy
	createdAt string
	updatedAt string
}

type preparedTag struct {
	id   string
	name string
}

type preparedProfile struct {
	id              string
	name            string
	runtimeID       string
	groupID         string
	profileDir      string
	containerID     string
	fingerprintSeed *int
	identityJSON    string
	fingerprintJSON string
	metadataJSON    string
	proxy           preparedProxy
	createdAt       string
	updatedAt       string
	lastUsedAt      string
	tags            []preparedTag
	sourceSHA       string
	sqliteSHA       string
}

type legacyGroupsFile struct {
	Groups []groupstore.Group `json:"groups"`
}

func (m *Migrator) buildPlan(now time.Time) (migrationPlan, error) {
	profilesDir, err := filepath.Abs(m.source.ProfilesDir)
	if err != nil {
		return migrationPlan{}, fmt.Errorf("%w: resolve profiles directory", ErrValidation)
	}
	groupsPath, err := filepath.Abs(m.source.GroupsPath)
	if err != nil {
		return migrationPlan{}, fmt.Errorf("%w: resolve groups path", ErrValidation)
	}
	plan := migrationPlan{
		profilesDir:  profilesDir,
		groupsPath:   groupsPath,
		sourceHashes: map[string]string{},
	}
	groups, groupHash, groupsPresent, err := readGroups(groupsPath, now)
	if err != nil {
		return migrationPlan{}, err
	}
	plan.groups = groups
	if groupsPresent {
		plan.sourceHashes["groups_json"] = groupHash
	}
	groupIDs := make(map[string]string, len(groups))
	for _, group := range groups {
		groupIDs[strings.ToLower(group.name)] = group.id
	}

	profiles, profileHash, runtimes, tags, err := readProfiles(profilesDir, groupIDs)
	if err != nil {
		return migrationPlan{}, err
	}
	plan.profiles = profiles
	plan.runtimes = runtimes
	plan.tags = tags
	plan.sourceHashes["profile_json"] = profileHash
	return plan, nil
}

func readGroups(path string, now time.Time) ([]preparedGroup, string, bool, error) {
	// #nosec G304 -- path is a local Core migration source and is parsed only in the read-only validation stage.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", false, nil
		}
		return nil, "", false, fmt.Errorf("%w: read groups source", ErrValidation)
	}
	if hasCleartextProxyCredentials(data) {
		return nil, "", false, fmt.Errorf("%w: cleartext proxy credential found in groups source", ErrValidation)
	}
	var source legacyGroupsFile
	if err := json.Unmarshal(data, &source); err != nil {
		return nil, "", false, fmt.Errorf("%w: decode groups source", ErrValidation)
	}
	byName := make(map[string]struct{}, len(source.Groups))
	out := make([]preparedGroup, 0, len(source.Groups))
	for _, raw := range source.Groups {
		name := strings.TrimSpace(raw.Name)
		key := strings.ToLower(name)
		if name == "" || key == "" {
			return nil, "", false, fmt.Errorf("%w: group with empty name", ErrValidation)
		}
		if _, exists := byName[key]; exists {
			return nil, "", false, fmt.Errorf("%w: duplicate group name", ErrValidation)
		}
		byName[key] = struct{}{}
		id := deterministicID("grp", key)
		mode := strings.ToLower(strings.TrimSpace(raw.ProxyMode))
		if mode == "" {
			mode = groupstore.ProxyModeDefault
		}
		if mode != groupstore.ProxyModeDefault && mode != groupstore.ProxyModeEnforced {
			return nil, "", false, fmt.Errorf("%w: invalid group proxy mode", ErrValidation)
		}
		proxy, err := normalizeProxy(raw.Proxy, "proxy.group."+id)
		if err != nil {
			return nil, "", false, err
		}
		if mode == groupstore.ProxyModeEnforced && proxy.kind == "" {
			return nil, "", false, fmt.Errorf("%w: enforced group requires a proxy", ErrValidation)
		}
		created := raw.CreatedAt.UTC()
		updated := raw.UpdatedAt.UTC()
		if created.IsZero() || updated.IsZero() {
			return nil, "", false, fmt.Errorf("%w: group timestamps are required", ErrValidation)
		}
		out = append(out, preparedGroup{
			id: id, name: name, proxyMode: mode, proxy: proxy,
			createdAt: timestamp(created), updatedAt: timestamp(updated),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out, sha256Hex(data), true, nil
}

func readProfiles(root string, groupIDs map[string]string) ([]preparedProfile, string, []string, []preparedTag, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, "", nil, nil, fmt.Errorf("%w: read profiles directory", ErrValidation)
	}
	seenProfiles := make(map[string]struct{})
	seenContainers := make(map[string]struct{})
	runtimeSet := make(map[string]struct{})
	tagByName := make(map[string]preparedTag)
	var sourceDigest strings.Builder
	out := make([]preparedProfile, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		dirID := entry.Name()
		if !validIdentifier(dirID) {
			return nil, "", nil, nil, fmt.Errorf("%w: invalid profile directory identifier", ErrValidation)
		}
		path := filepath.Join(root, dirID, "profile.json")
		// #nosec G304 -- dirID originates from os.ReadDir(root), symlinks are rejected, and the identifier is validated before this controlled join.
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, "", nil, nil, fmt.Errorf("%w: read profile source", ErrValidation)
		}
		if hasCleartextProxyCredentials(data) {
			return nil, "", nil, nil, fmt.Errorf("%w: cleartext proxy credential found in profile source", ErrValidation)
		}
		var raw profile.Profile
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, "", nil, nil, fmt.Errorf("%w: decode profile source", ErrValidation)
		}
		if raw.ID != dirID || !validIdentifier(raw.ID) {
			return nil, "", nil, nil, fmt.Errorf("%w: profile id must match its source directory", ErrValidation)
		}
		if _, exists := seenProfiles[raw.ID]; exists {
			return nil, "", nil, nil, fmt.Errorf("%w: duplicate profile id", ErrValidation)
		}
		seenProfiles[raw.ID] = struct{}{}
		if strings.TrimSpace(raw.Name) == "" || !validIdentifier(raw.RuntimeID) {
			return nil, "", nil, nil, fmt.Errorf("%w: profile name and runtime id are required", ErrValidation)
		}
		if raw.CreatedAt.IsZero() {
			return nil, "", nil, nil, fmt.Errorf("%w: profile creation time is required", ErrValidation)
		}
		expectedDir := filepath.Join(root, raw.ID)
		storedDir, err := filepath.Abs(raw.ProfileDir)
		if err != nil || storedDir != expectedDir {
			return nil, "", nil, nil, fmt.Errorf("%w: profile directory does not match source", ErrValidation)
		}
		groupID := ""
		if groupName := strings.TrimSpace(raw.Group); groupName != "" {
			var ok bool
			groupID, ok = groupIDs[strings.ToLower(groupName)]
			if !ok {
				return nil, "", nil, nil, fmt.Errorf("%w: profile references an unknown group", ErrValidation)
			}
		}
		containerID := strings.TrimSpace(raw.ContainerID)
		if containerID != "" {
			if _, exists := seenContainers[containerID]; exists {
				return nil, "", nil, nil, fmt.Errorf("%w: duplicate non-empty container id", ErrValidation)
			}
			seenContainers[containerID] = struct{}{}
		}
		proxy, err := normalizeProxy(raw.Proxy, "proxy."+raw.ID)
		if err != nil {
			return nil, "", nil, nil, err
		}
		identityJSON, err := marshalObject(raw.Identity)
		if err != nil {
			return nil, "", nil, nil, err
		}
		fingerprintJSON, err := marshalObject(raw.Fingerprint)
		if err != nil {
			return nil, "", nil, nil, err
		}
		profileTags, err := normalizeTags(raw.Tags, tagByName)
		if err != nil {
			return nil, "", nil, nil, err
		}
		var seed *int
		if raw.FingerprintSeed != 0 {
			value := int(raw.FingerprintSeed)
			seed = &value
		}
		prepared := preparedProfile{
			id: raw.ID, name: strings.TrimSpace(raw.Name), runtimeID: raw.RuntimeID,
			groupID: groupID, profileDir: expectedDir, containerID: containerID,
			fingerprintSeed: seed, identityJSON: identityJSON, fingerprintJSON: fingerprintJSON,
			metadataJSON: "{}", proxy: proxy, createdAt: timestamp(raw.CreatedAt),
			updatedAt: timestamp(raw.CreatedAt), lastUsedAt: timestamp(raw.LastUsed),
			tags: profileTags, sourceSHA: sha256Hex(data),
		}
		prepared.sqliteSHA, err = canonicalProfileHash(prepared)
		if err != nil {
			return nil, "", nil, nil, err
		}
		out = append(out, prepared)
		runtimeSet[raw.RuntimeID] = struct{}{}
		sourceDigest.WriteString(raw.ID)
		sourceDigest.WriteByte(':')
		sourceDigest.WriteString(prepared.sourceSHA)
		sourceDigest.WriteByte('\n')
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	runtimes := make([]string, 0, len(runtimeSet))
	for id := range runtimeSet {
		runtimes = append(runtimes, id)
	}
	sort.Strings(runtimes)
	tags := make([]preparedTag, 0, len(tagByName))
	for _, tag := range tagByName {
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].id < tags[j].id })
	return out, sha256Hex([]byte(sourceDigest.String())), runtimes, tags, nil
}

func normalizeTags(source []string, tagByName map[string]preparedTag) ([]preparedTag, error) {
	profileTags := make([]preparedTag, 0, len(source))
	seen := make(map[string]struct{})
	for _, raw := range source {
		name := strings.TrimSpace(raw)
		key := strings.ToLower(name)
		if name == "" {
			return nil, fmt.Errorf("%w: empty profile tag", ErrValidation)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		tag, exists := tagByName[key]
		if !exists {
			tag = preparedTag{id: deterministicID("tag", key), name: name}
			tagByName[key] = tag
		}
		profileTags = append(profileTags, tag)
	}
	sort.Slice(profileTags, func(i, j int) bool { return profileTags[i].id < profileTags[j].id })
	return profileTags, nil
}

func normalizeProxy(source *profile.ProxyConfig, expectedSecretRef string) (preparedProxy, error) {
	if source == nil {
		return preparedProxy{}, nil
	}
	if source.Username != "" || source.Password != "" {
		return preparedProxy{}, fmt.Errorf("%w: cleartext proxy credential in memory", ErrValidation)
	}
	kind := strings.ToLower(strings.TrimSpace(source.Type))
	host := strings.TrimSpace(source.Host)
	region := strings.TrimSpace(source.Region)
	secretRef := strings.TrimSpace(source.SecretRef)
	if kind != "http" && kind != "socks5" {
		return preparedProxy{}, fmt.Errorf("%w: unsupported proxy type", ErrValidation)
	}
	if host == "" || source.Port < 1 || source.Port > 65535 {
		return preparedProxy{}, fmt.Errorf("%w: invalid proxy endpoint", ErrValidation)
	}
	if secretRef != "" && secretRef != expectedSecretRef {
		return preparedProxy{}, fmt.Errorf("%w: proxy secret reference does not belong to its entity", ErrValidation)
	}
	return preparedProxy{secretRef: secretRef, kind: kind, host: host, port: source.Port, region: region}, nil
}

func (m *Migrator) recordDryRun(ctx context.Context, plan migrationPlan, report Report, now time.Time) error {
	if err := m.startOperations(ctx, plan, report, now); err != nil {
		return err
	}
	if err := m.finalizeOperations(context.Background(), report.OperationID, "validated", "", now); err != nil {
		return err
	}
	return nil
}

// RecoverInterruptedOperations fail-closes journal rows left started by a process
// interruption. It never retries or mutates product records, so a later explicit
// migration can safely start a fresh operation after the source is revalidated.
func (m *Migrator) RecoverInterruptedOperations(ctx context.Context) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result, err := m.db.ExecContext(ctx, `UPDATE profile_import_operations
		SET state = 'failed', error_code = ?, updated_at = ?
		WHERE state = 'started'`, ErrInterruptedOperation.Error(), timestamp(time.Now().UTC()))
	if err != nil {
		return 0, fmt.Errorf("recover interrupted import operations: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count recovered import operations: %w", err)
	}
	return int(count), nil
}

func (m *Migrator) startOperations(ctx context.Context, plan migrationPlan, report Report, now time.Time) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin durable import journal: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := insertOperations(ctx, tx, plan, report, "started", now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit durable import journal: %w", err)
	}
	return nil
}

func (m *Migrator) finalizeOperations(ctx context.Context, operationID, state, errorCode string, now time.Time) error {
	if state != "validated" && state != "committed" && state != "failed" {
		return fmt.Errorf("%w: unsupported import operation terminal state %q", ErrValidation, state)
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin import operation finalization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE profile_import_operations
		SET state = ?, error_code = ?, updated_at = ?
		WHERE substr(id, 1, ?) = ? AND state = 'started'`,
		state, errorCode, timestamp(now), len(operationID)+1, operationID+".")
	if err != nil {
		return fmt.Errorf("finalize import operations: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count finalized import operations: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("%w: no started import operations for %s", ErrValidation, operationID)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit import operation finalization: %w", err)
	}
	return nil
}

func failureCode(err error) string {
	if errors.Is(err, context.Canceled) {
		return "CONTEXT_CANCELED"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "CONTEXT_DEADLINE_EXCEEDED"
	}
	if errors.Is(err, ErrValidation) {
		return "VALIDATION_FAILED"
	}
	if errors.Is(err, ErrExistingProductRecord) {
		return "PRODUCT_RECORD_CONFLICT"
	}
	return "IMPORT_FAILED"
}

func (m *Migrator) ensureNoExistingProductRecords(ctx context.Context, plan migrationPlan) error {
	for _, group := range plan.groups {
		var found int
		if err := m.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)`, group.id).Scan(&found); err != nil {
			return fmt.Errorf("preflight group conflict: %w", err)
		}
		if found == 1 {
			return fmt.Errorf("%w: group", ErrExistingProductRecord)
		}
	}
	for _, item := range plan.profiles {
		var found int
		if err := m.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE id = ?)`, item.id).Scan(&found); err != nil {
			return fmt.Errorf("preflight profile conflict: %w", err)
		}
		if found == 1 {
			return fmt.Errorf("%w: profile", ErrExistingProductRecord)
		}
	}
	return nil
}

func (m *Migrator) apply(ctx context.Context, plan migrationPlan, report Report, now time.Time) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin product import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, group := range plan.groups {
		if exists, err := entityExists(ctx, tx, "groups", group.id); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("%w: group", ErrExistingProductRecord)
		}
	}
	for _, item := range plan.profiles {
		if exists, err := entityExists(ctx, tx, "profiles", item.id); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("%w: profile", ErrExistingProductRecord)
		}
	}
	for _, runtimeID := range plan.runtimes {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO runtime_candidates
			(id, name, version, architecture, binary_path, binary_sha256, status, created_at)
			VALUES (?, ?, '', '', '', '', 'candidate', ?)`, runtimeID, runtimeID, timestamp(now)); err != nil {
			return fmt.Errorf("insert runtime candidate: %w", err)
		}
	}
	for _, group := range plan.groups {
		if _, err := tx.ExecContext(ctx, `INSERT INTO groups
			(id, name, proxy_mode, proxy_secret_ref, proxy_type, proxy_host, proxy_port, proxy_region, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			group.id, group.name, group.proxyMode, group.proxy.secretRef, group.proxy.kind, group.proxy.host, group.proxy.port, group.proxy.region, group.createdAt, group.updatedAt); err != nil {
			return fmt.Errorf("insert group: %w", err)
		}
	}
	for _, tag := range plan.tags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_tags(id, name, created_at) VALUES (?, ?, ?) ON CONFLICT(name) DO NOTHING`, tag.id, tag.name, timestamp(now)); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	for _, item := range plan.profiles {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profiles
			(id, name, runtime_id, group_id, profile_dir, container_id, fingerprint_seed, identity_json, fingerprint_json, metadata_json, proxy_secret_ref, proxy_type, proxy_host, proxy_port, proxy_region, created_at, updated_at, last_used_at)
			VALUES (?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.id, item.name, item.runtimeID, item.groupID, item.profileDir, item.containerID, item.fingerprintSeed,
			item.identityJSON, item.fingerprintJSON, item.metadataJSON, item.proxy.secretRef, item.proxy.kind, item.proxy.host,
			item.proxy.port, item.proxy.region, item.createdAt, item.updatedAt, item.lastUsedAt); err != nil {
			return fmt.Errorf("insert profile: %w", err)
		}
		for _, tag := range item.tags {
			if _, err := tx.ExecContext(ctx, `INSERT INTO profile_tag_assignments(profile_id, tag_id, created_at) VALUES (?, ?, ?)`, item.id, tag.id, timestamp(now)); err != nil {
				return fmt.Errorf("assign profile tag: %w", err)
			}
		}
		actualHash, err := canonicalSQLiteProfileHash(ctx, tx, item.id)
		if err != nil {
			return err
		}
		if actualHash != item.sqliteSHA {
			return fmt.Errorf("%w: SQLite parity mismatch for profile", ErrValidation)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_json_parity_checks
			(id, profile_id, source_sha256, sqlite_record_sha256, result, checked_at, correlation_id)
			VALUES (?, ?, ?, ?, 'match', ?, ?)`, deterministicID("parity", report.OperationID+":"+item.id), item.id, item.sourceSHA, actualHash, timestamp(now), report.CorrelationID); err != nil {
			return fmt.Errorf("insert parity record: %w", err)
		}
	}
	summary, err := summaryJSON(report)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO product_audit_events(event_type, entity_type, entity_id, correlation_id, details_json, created_at)
		VALUES ('profile_json_import_committed', 'profile_migration', ?, ?, ?, ?)`, report.OperationID, report.CorrelationID, summary, timestamp(now)); err != nil {
		return fmt.Errorf("insert product audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit product import: %w", err)
	}
	return nil
}

type operationSource struct {
	kind string
	hash string
	id   string
}

func operationSources(plan migrationPlan, operationID string) ([]operationSource, error) {
	kinds := make([]string, 0, len(plan.sourceHashes))
	for kind := range plan.sourceHashes {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	if len(kinds) == 0 {
		return nil, fmt.Errorf("%w: import plan has no sources", ErrValidation)
	}
	sources := make([]operationSource, 0, len(kinds))
	for _, kind := range kinds {
		hash := plan.sourceHashes[kind]
		if (kind != "groups_json" && kind != "profile_json") || !isSHA256(hash) {
			return nil, fmt.Errorf("%w: invalid import source %q", ErrValidation, kind)
		}
		sources = append(sources, operationSource{kind: kind, hash: hash, id: operationID + "." + kind})
	}
	return sources, nil
}

func insertOperations(ctx context.Context, tx *sql.Tx, plan migrationPlan, report Report, state string, now time.Time) error {
	summary, err := summaryJSON(report)
	if err != nil {
		return err
	}
	sources, err := operationSources(plan, report.OperationID)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_import_operations
			(id, source_kind, source_sha256, dry_run, state, summary_json, correlation_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, source.id, source.kind, source.hash, boolInt(report.Mode == ModeDryRun), state, summary, report.CorrelationID, timestamp(now), timestamp(now)); err != nil {
			return fmt.Errorf("record import operation: %w", err)
		}
	}
	return nil
}

func entityExists(ctx context.Context, tx *sql.Tx, table, id string) (bool, error) {
	if table != "groups" && table != "profiles" {
		return false, fmt.Errorf("unsupported entity table")
	}
	var found int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM `+table+` WHERE id = ?)`, id).Scan(&found); err != nil {
		return false, fmt.Errorf("check existing %s: %w", table, err)
	}
	return found == 1, nil
}

func (p migrationPlan) report(operationID, correlationID string, mode Mode) Report {
	parity := make([]Parity, 0, len(p.profiles))
	for _, item := range p.profiles {
		parity = append(parity, Parity{ProfileID: item.id, SourceSHA256: item.sourceSHA, SQLiteRecordSHA256: item.sqliteSHA, Result: "match"})
	}
	return Report{
		OperationID: operationID, CorrelationID: correlationID, Mode: mode,
		State:    map[Mode]string{ModeDryRun: "validated", ModeApply: "committed"}[mode],
		Profiles: len(p.profiles), Groups: len(p.groups), Tags: len(p.tags), Runtimes: len(p.runtimes), Sources: len(p.sourceHashes),
		SourceHashes: cloneStringMap(p.sourceHashes), Parity: parity,
	}
}

func canonicalProfileHash(item preparedProfile) (string, error) {
	canonical := struct {
		ID, Name, RuntimeID, GroupID, ProfileDir, ContainerID string
		FingerprintSeed                                       *int
		IdentityJSON, FingerprintJSON, MetadataJSON           string
		Proxy                                                 preparedProxy
		CreatedAt, UpdatedAt, LastUsedAt                      string
		Tags                                                  []preparedTag
	}{
		ID: item.id, Name: item.name, RuntimeID: item.runtimeID, GroupID: item.groupID,
		ProfileDir: item.profileDir, ContainerID: item.containerID, FingerprintSeed: item.fingerprintSeed,
		IdentityJSON: item.identityJSON, FingerprintJSON: item.fingerprintJSON, MetadataJSON: item.metadataJSON,
		Proxy: item.proxy, CreatedAt: item.createdAt, UpdatedAt: item.updatedAt, LastUsedAt: item.lastUsedAt, Tags: item.tags,
	}
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("canonicalize profile parity: %w", err)
	}
	return sha256Hex(data), nil
}

func canonicalSQLiteProfileHash(ctx context.Context, tx *sql.Tx, profileID string) (string, error) {
	var item preparedProfile
	var groupID sql.NullString
	var fingerprintSeed sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT id, name, runtime_id, group_id, profile_dir, container_id, fingerprint_seed,
		identity_json, fingerprint_json, metadata_json, proxy_secret_ref, proxy_type, proxy_host, proxy_port,
		proxy_region, created_at, updated_at, last_used_at
		FROM profiles WHERE id = ?`, profileID).Scan(
		&item.id, &item.name, &item.runtimeID, &groupID, &item.profileDir, &item.containerID, &fingerprintSeed,
		&item.identityJSON, &item.fingerprintJSON, &item.metadataJSON, &item.proxy.secretRef, &item.proxy.kind,
		&item.proxy.host, &item.proxy.port, &item.proxy.region, &item.createdAt, &item.updatedAt, &item.lastUsedAt,
	); err != nil {
		return "", fmt.Errorf("read SQLite profile for parity: %w", err)
	}
	if groupID.Valid {
		item.groupID = groupID.String
	}
	if fingerprintSeed.Valid {
		value := int(fingerprintSeed.Int64)
		item.fingerprintSeed = &value
	}
	rows, err := tx.QueryContext(ctx, `SELECT t.id, t.name
		FROM profile_tags AS t
		JOIN profile_tag_assignments AS a ON a.tag_id = t.id
		WHERE a.profile_id = ?
		ORDER BY t.id`, profileID)
	if err != nil {
		return "", fmt.Errorf("read SQLite profile tags for parity: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tag preparedTag
		if err := rows.Scan(&tag.id, &tag.name); err != nil {
			return "", fmt.Errorf("read SQLite profile tag for parity: %w", err)
		}
		item.tags = append(item.tags, tag)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate SQLite profile tags for parity: %w", err)
	}
	return canonicalProfileHash(item)
}

func summaryJSON(report Report) (string, error) {
	value := struct {
		Profiles, Groups, Tags, Runtimes, Sources int
		Mode                                      Mode
		SourceHashes                              map[string]string
		Backups                                   []BackupReceipt
	}{report.Profiles, report.Groups, report.Tags, report.Runtimes, report.Sources, report.Mode, report.SourceHashes, report.Backups}
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode redacted import summary: %w", err)
	}
	return string(data), nil
}

func hasCleartextProxyCredentials(data []byte) bool {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return false
	}
	return walkForProxyCredentials(value)
}

func walkForProxyCredentials(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if strings.EqualFold(key, "proxy") {
				if proxy, ok := child.(map[string]any); ok {
					for proxyKey := range proxy {
						if strings.EqualFold(proxyKey, "username") || strings.EqualFold(proxyKey, "password") {
							return true
						}
					}
				}
			}
			if walkForProxyCredentials(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if walkForProxyCredentials(child) {
				return true
			}
		}
	}
	return false
}

func marshalObject(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("canonicalize profile JSON: %w", err)
	}
	return string(data), nil
}

func timestamp(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func deterministicID(prefix, value string) string {
	digest := sha256.Sum256([]byte(value))
	return prefix + "_" + hex.EncodeToString(digest[:12])
}

func newOperationID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate migration operation id: %w", err)
	}
	return "imp_" + hex.EncodeToString(bytes), nil
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func validIdentifier(value string) bool { return identifierPattern.MatchString(value) }

func validateReceipts(receipts []BackupReceipt) error {
	if len(receipts) == 0 {
		return ErrBackupRequired
	}
	for _, receipt := range receipts {
		if !validIdentifier(receipt.ID) || !isSHA256(receipt.SHA256) {
			return fmt.Errorf("%w: invalid backup receipt", ErrBackupRequired)
		}
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func cloneStringMap(source map[string]string) map[string]string {
	out := make(map[string]string, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
