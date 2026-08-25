package extensions

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	MaxZIPBytes       = 100 << 20
	MaxExpandedBytes  = 200 << 20
	MaxEntries        = 500
	MaxManifestBytes  = 1 << 20
	DefaultPageSize   = 50
	digestPreviewSize = 16
)

var (
	ErrInvalidArchive     = errors.New("invalid archive")
	ErrArchiveLimit       = errors.New("archive limit exceeded")
	ErrManifestInvalid    = errors.New("invalid manifest")
	ErrNotFound           = errors.New("extension not found")
	ErrSeriesNotFound     = errors.New("extension series not found")
	ErrVersionNotFound    = errors.New("extension version not found")
	ErrPermissionAck      = errors.New("permission acknowledgement required")
	ErrHighRiskAck        = errors.New("high risk acknowledgement required")
	ErrNotApproved        = errors.New("extension version is not approved")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrRevoked            = errors.New("extension version is revoked or quarantined")
	ErrConcurrentMutation = errors.New("concurrent mutation")
	ErrPurgeNotAllowed    = errors.New("purge not allowed")
	ErrInvalidID          = errors.New("invalid extension id")
)

type Manifest struct {
	Name                    string   `json:"name,omitempty"`
	Version                 string   `json:"version,omitempty"`
	ManifestVersion         int      `json:"manifest_version,omitempty"`
	Permissions             []string `json:"permissions,omitempty"`
	OptionalPermissions     []string `json:"optional_permissions,omitempty"`
	HostPermissions         []string `json:"host_permissions,omitempty"`
	OptionalHostPermissions []string `json:"optional_host_permissions,omitempty"`
	ContentScriptMatches    []string `json:"content_script_matches,omitempty"`
}

type Version struct {
	ID             string   `json:"id"`
	SeriesID       string   `json:"series_id"`
	Number         int      `json:"number"`
	State          string   `json:"state"`
	DigestPreview  string   `json:"digest_preview"`
	Size           int64    `json:"size"`
	Format         string   `json:"format"`
	Manifest       Manifest `json:"manifest"`
	RiskState      string   `json:"risk_state"`
	RiskCategories []string `json:"risk_categories,omitempty"`
	CreatedAt      string   `json:"created_at"`
	ApprovedAt     string   `json:"approved_at,omitempty"`
}

type Assignment struct {
	ID        string `json:"id"`
	VersionID string `json:"version_id"`
	ProfileID string `json:"profile_id"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

type Series struct {
	ID              string       `json:"id"`
	ActiveVersionID string       `json:"active_version_id,omitempty"`
	CreatedAt       string       `json:"created_at"`
	Versions        []Version    `json:"versions"`
	Assignments     []Assignment `json:"assignments,omitempty"`
}

type ListResult struct {
	Data   []Series `json:"data"`
	Total  int      `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
}

type ProfileExists func(context.Context, string) (bool, error)

type Repository struct {
	db           *sql.DB
	baseDir      string
	objectDir    string
	stagingDir   string
	mu           sync.Mutex
	beforeCommit func() error
}

func Open(baseDir string) (*Repository, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("T28 base directory is required")
	}
	root := filepath.Join(baseDir, "extensions")
	objectDir := filepath.Join(root, "objects")
	stagingDir := filepath.Join(root, "staging")
	for _, dir := range []string{root, objectDir, stagingDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
		if err := os.Chmod(dir, 0700); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "extensions.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Repository{db: db, baseDir: root, objectDir: objectDir, stagingDir: stagingDir}, nil
}

func (r *Repository) Close() error                               { return r.db.Close() }
func (r *Repository) DB() *sql.DB                                { return r.db }
func (r *Repository) SetFailBeforeCommitForTest(fn func() error) { r.beforeCommit = fn }

func migrate(db *sql.DB) error {
	statements := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA busy_timeout = 5000`,
		`CREATE TABLE IF NOT EXISTS t28_schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS extension_series (id TEXT PRIMARY KEY, active_version_id TEXT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS extension_versions (
            id TEXT PRIMARY KEY, series_id TEXT NOT NULL, version_number INTEGER NOT NULL,
            state TEXT NOT NULL, digest TEXT NOT NULL UNIQUE, digest_preview TEXT NOT NULL,
            size_bytes INTEGER NOT NULL, format TEXT NOT NULL, blob_relpath TEXT NOT NULL,
            manifest_json TEXT NOT NULL, risk_state TEXT NOT NULL, risk_categories_json TEXT NOT NULL,
            created_at TEXT NOT NULL, approved_at TEXT NULL, revoked_at TEXT NULL,
            FOREIGN KEY(series_id) REFERENCES extension_series(id)
        )`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_t28_series_version ON extension_versions(series_id, version_number)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_t28_active_version ON extension_series(active_version_id) WHERE active_version_id IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS extension_assignments (
            id TEXT PRIMARY KEY, version_id TEXT NOT NULL, profile_id TEXT NOT NULL,
            state TEXT NOT NULL, created_at TEXT NOT NULL,
            UNIQUE(version_id, profile_id),
            FOREIGN KEY(version_id) REFERENCES extension_versions(id)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_t28_assign_profile ON extension_assignments(profile_id, state)`,
		`CREATE TABLE IF NOT EXISTS extension_audit_events (
            id INTEGER PRIMARY KEY AUTOINCREMENT, series_id TEXT NOT NULL, version_id TEXT NULL,
            action TEXT NOT NULL, result TEXT NOT NULL, digest_preview TEXT NULL,
            permission_categories_json TEXT NOT NULL, profile_pseudonym TEXT NULL,
            error_code TEXT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL
        )`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO t28_schema_migrations(version, applied_at) VALUES(1, ?)`, now())
	return err
}

func (r *Repository) Import(ctx context.Context, src io.Reader, seriesID, correlation string) (*Version, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if src == nil {
		return nil, ErrInvalidArchive
	}
	temp, err := os.CreateTemp(r.stagingDir, "import-*.zip")
	if err != nil {
		return nil, fmt.Errorf("%w: staging", ErrInvalidArchive)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	hash := sha256.New()
	limited := io.LimitReader(src, MaxZIPBytes+1)
	n, err := io.Copy(io.MultiWriter(temp, hash), limited)
	if err != nil {
		_ = temp.Close()
		return nil, fmt.Errorf("%w: read", ErrInvalidArchive)
	}
	if n > MaxZIPBytes {
		_ = temp.Close()
		return nil, ErrArchiveLimit
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return nil, fmt.Errorf("%w: sync", ErrInvalidArchive)
	}
	if err := temp.Close(); err != nil {
		return nil, fmt.Errorf("%w: close", ErrInvalidArchive)
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	manifest, err := inspectArchive(tempPath)
	if err != nil {
		return nil, err
	}
	destDir := filepath.Join(r.objectDir, digest[:2])
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return nil, fmt.Errorf("%w: object directory", ErrInvalidArchive)
	}
	dest := filepath.Join(destDir, digest+".zip")
	if _, statErr := os.Stat(dest); os.IsNotExist(statErr) {
		if err := os.Rename(tempPath, dest); err != nil {
			return nil, fmt.Errorf("%w: object copy", ErrInvalidArchive)
		}
	} else if statErr != nil {
		return nil, fmt.Errorf("%w: object stat", ErrInvalidArchive)
	}
	if seriesID != "" && !validID(seriesID) {
		return nil, ErrInvalidID
	}
	if correlation == "" {
		correlation = "t28-local"
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: database", ErrInvalidArchive)
	}
	defer tx.Rollback()
	created := now()
	if seriesID == "" {
		seriesID, err = newID("series")
		if err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO extension_series(id, created_at) VALUES(?,?)`, seriesID, created); err != nil {
			return nil, err
		}
	} else {
		var exists int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM extension_series WHERE id=?`, seriesID).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, ErrSeriesNotFound
		}
	}
	var next int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number),0)+1 FROM extension_versions WHERE series_id=?`, seriesID).Scan(&next); err != nil {
		return nil, err
	}
	versionID, err := newID("version")
	if err != nil {
		return nil, err
	}
	manifestJSON, _ := json.Marshal(manifest.Manifest)
	riskJSON, _ := json.Marshal(manifest.RiskCategories)
	_, err = tx.ExecContext(ctx, `INSERT INTO extension_versions(id,series_id,version_number,state,digest,digest_preview,size_bytes,format,blob_relpath,manifest_json,risk_state,risk_categories_json,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		versionID, seriesID, next, "imported", digest, digest[:digestPreviewSize], n, "zip", filepath.ToSlash(filepath.Join("objects", digest[:2], digest+".zip")), string(manifestJSON), manifest.RiskState, string(riskJSON), created)
	if err != nil {
		return nil, err
	}
	if err = r.audit(ctx, tx, seriesID, versionID, "extension.imported", "success", digest[:digestPreviewSize], manifest.RiskCategories, "", "", correlation); err != nil {
		return nil, err
	}
	if err = r.beforeCommitHook(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Version{ID: versionID, SeriesID: seriesID, Number: next, State: "imported", DigestPreview: digest[:digestPreviewSize], Size: n, Format: "zip", Manifest: manifest.Manifest, RiskState: manifest.RiskState, RiskCategories: manifest.RiskCategories, CreatedAt: created}, nil
}

type parsedManifest struct {
	Manifest       Manifest
	RiskState      string
	RiskCategories []string
}

func inspectArchive(path string) (parsedManifest, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return parsedManifest{}, ErrInvalidArchive
	}
	defer zr.Close()
	if len(zr.File) == 0 || len(zr.File) > MaxEntries {
		return parsedManifest{}, ErrArchiveLimit
	}
	var total uint64
	names := map[string]struct{}{}
	var manifestFile *zip.File
	for _, f := range zr.File {
		name := f.Name
		clean := filepath.ToSlash(filepath.Clean(name))
		if name == "" || strings.Contains(name, "\x00") || filepath.IsAbs(name) || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(name, "\\") {
			return parsedManifest{}, ErrInvalidArchive
		}
		if clean != name || strings.HasPrefix(name, "/") {
			return parsedManifest{}, ErrInvalidArchive
		}
		if _, ok := names[name]; ok {
			return parsedManifest{}, ErrInvalidArchive
		}
		names[name] = struct{}{}
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			return parsedManifest{}, ErrInvalidArchive
		}
		if f.UncompressedSize64 > MaxExpandedBytes || total > MaxExpandedBytes-f.UncompressedSize64 {
			return parsedManifest{}, ErrArchiveLimit
		}
		total += f.UncompressedSize64
		if name == "manifest.json" {
			manifestFile = f
		}
	}
	if manifestFile == nil || manifestFile.FileInfo().IsDir() {
		return parsedManifest{}, ErrManifestInvalid
	}
	rc, err := manifestFile.Open()
	if err != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	raw, err := io.ReadAll(io.LimitReader(rc, MaxManifestBytes+1))
	_ = rc.Close()
	if err != nil || len(raw) > MaxManifestBytes {
		return parsedManifest{}, ErrManifestInvalid
	}
	var obj map[string]json.RawMessage
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	if err := dec.Decode(&obj); err != nil || obj == nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return parsedManifest{}, ErrManifestInvalid
	}
	m := Manifest{}
	if err := decodeStringField(obj, "name", &m.Name); err != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if err := decodeStringField(obj, "version", &m.Version); err != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if err := decodeIntField(obj, "manifest_version", &m.ManifestVersion); err != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	var errList error
	if m.Permissions, errList = decodeStringList(obj, "permissions"); errList != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if m.OptionalPermissions, errList = decodeStringList(obj, "optional_permissions"); errList != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if m.HostPermissions, errList = decodeStringList(obj, "host_permissions"); errList != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if m.OptionalHostPermissions, errList = decodeStringList(obj, "optional_host_permissions"); errList != nil {
		return parsedManifest{}, ErrManifestInvalid
	}
	if rawScripts, ok := obj["content_scripts"]; ok {
		var scripts []struct {
			Matches []string `json:"matches"`
		}
		if err := json.Unmarshal(rawScripts, &scripts); err != nil {
			return parsedManifest{}, ErrManifestInvalid
		}
		for _, script := range scripts {
			m.ContentScriptMatches = append(m.ContentScriptMatches, script.Matches...)
		}
	}
	m.Permissions = normalizeList(m.Permissions)
	m.OptionalPermissions = normalizeList(m.OptionalPermissions)
	m.HostPermissions = normalizeList(m.HostPermissions)
	m.OptionalHostPermissions = normalizeList(m.OptionalHostPermissions)
	m.ContentScriptMatches = normalizeList(m.ContentScriptMatches)
	categories := riskCategories(m)
	state := "NORMAL"
	if len(categories) > 0 {
		state = "HIGH_RISK"
	}
	return parsedManifest{Manifest: m, RiskState: state, RiskCategories: categories}, nil
}

func decodeStringField(obj map[string]json.RawMessage, key string, out *string) error {
	raw, ok := obj[key]
	if !ok {
		return nil
	}
	return json.Unmarshal(raw, out)
}
func decodeIntField(obj map[string]json.RawMessage, key string, out *int) error {
	raw, ok := obj[key]
	if !ok {
		return nil
	}
	return json.Unmarshal(raw, out)
}
func decodeStringList(obj map[string]json.RawMessage, key string) ([]string, error) {
	raw, ok := obj[key]
	if !ok {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}
func normalizeList(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func riskCategories(m Manifest) []string {
	sensitive := map[string]bool{"cookies": true, "webRequest": true, "webRequestBlocking": true, "debugger": true, "nativeMessaging": true, "management": true, "proxy": true, "downloads": true, "clipboardRead": true}
	categories := map[string]struct{}{}
	for _, permission := range append(append([]string{}, m.Permissions...), m.OptionalPermissions...) {
		if sensitive[permission] {
			categories[permission] = struct{}{}
		} else if !knownPermission(permission) {
			categories["UNCLASSIFIED_HIGH_RISK"] = struct{}{}
		}
	}
	for _, host := range append(append(append([]string{}, m.HostPermissions...), m.OptionalHostPermissions...), m.ContentScriptMatches...) {
		if host == "<all_urls>" || host == "*://*/*" || host == "file:///*" || strings.Contains(host, "://*") {
			categories[host] = struct{}{}
		}
	}
	out := make([]string, 0, len(categories))
	for value := range categories {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
func knownPermission(value string) bool {
	switch value {
	case "activeTab", "alarms", "bookmarks", "contextMenus", "idle", "storage", "tabs", "unlimitedStorage", "notifications", "scripting", "sessions", "webNavigation":
		return true
	default:
		return false
	}
}

func (r *Repository) List(ctx context.Context, limit, offset int) (ListResult, error) {
	if limit <= 0 || limit > 100 {
		limit = DefaultPageSize
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM extension_series`).Scan(&total); err != nil {
		return ListResult{}, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, COALESCE(active_version_id,''), created_at FROM extension_series ORDER BY created_at, id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return ListResult{}, err
	}
	result := ListResult{Data: []Series{}, Total: total, Limit: limit, Offset: offset}
	series := make([]Series, 0, limit)
	for rows.Next() {
		var s Series
		if err := rows.Scan(&s.ID, &s.ActiveVersionID, &s.CreatedAt); err != nil {
			_ = rows.Close()
			return ListResult{}, err
		}
		series = append(series, s)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return ListResult{}, err
	}
	if err := rows.Close(); err != nil {
		return ListResult{}, err
	}
	for _, s := range series {
		versions, err := r.listVersions(ctx, s.ID)
		if err != nil {
			return ListResult{}, err
		}
		s.Versions = versions
		result.Data = append(result.Data, s)
	}
	return result, nil
}

func (r *Repository) GetSeries(ctx context.Context, seriesID string) (*Series, error) {
	if !validID(seriesID) {
		return nil, ErrInvalidID
	}
	var s Series
	if err := r.db.QueryRowContext(ctx, `SELECT id, COALESCE(active_version_id,''), created_at FROM extension_series WHERE id=?`, seriesID).Scan(&s.ID, &s.ActiveVersionID, &s.CreatedAt); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSeriesNotFound
	} else if err != nil {
		return nil, err
	}
	var err error
	s.Versions, err = r.listVersions(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, version_id, profile_id, state, created_at FROM extension_assignments WHERE version_id IN (SELECT id FROM extension_versions WHERE series_id=?) ORDER BY created_at, id`, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a Assignment
		if err := rows.Scan(&a.ID, &a.VersionID, &a.ProfileID, &a.State, &a.CreatedAt); err != nil {
			return nil, err
		}
		s.Assignments = append(s.Assignments, a)
	}
	return &s, rows.Err()
}

func (r *Repository) listVersions(ctx context.Context, seriesID string) ([]Version, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, series_id, version_number, state, digest_preview, size_bytes, format, manifest_json, risk_state, risk_categories_json, created_at, COALESCE(approved_at,'') FROM extension_versions WHERE series_id=? ORDER BY version_number`, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	versions := []Version{}
	for rows.Next() {
		var v Version
		var manifestJSON, riskJSON string
		if err := rows.Scan(&v.ID, &v.SeriesID, &v.Number, &v.State, &v.DigestPreview, &v.Size, &v.Format, &manifestJSON, &v.RiskState, &riskJSON, &v.CreatedAt, &v.ApprovedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(manifestJSON), &v.Manifest); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(riskJSON), &v.RiskCategories)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *Repository) Approve(ctx context.Context, versionID string, acknowledged []string, acceptHighRisk bool, correlation string) (*Version, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var seriesID, state, manifestJSON, riskState, riskJSON, digestPreview, created, approved string
	var number int
	var size int64
	var format string
	err = tx.QueryRowContext(ctx, `SELECT series_id,version_number,state,manifest_json,risk_state,risk_categories_json,digest_preview,size_bytes,format,created_at,COALESCE(approved_at,'') FROM extension_versions WHERE id=?`, versionID).Scan(&seriesID, &number, &state, &manifestJSON, &riskState, &riskJSON, &digestPreview, &size, &format, &created, &approved)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	if state == "revoked" || state == "quarantined" {
		return nil, ErrRevoked
	}
	if state != "imported" {
		return nil, ErrPermissionAck
	}
	var risk []string
	_ = json.Unmarshal([]byte(riskJSON), &risk)
	var manifest Manifest
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return nil, ErrManifestInvalid
	}
	if !equalStrings(normalizeList(acknowledged), manifestAcknowledgement(manifest)) {
		return nil, ErrPermissionAck
	}
	if len(risk) > 0 && !acceptHighRisk {
		return nil, ErrHighRiskAck
	}
	approved = now()
	if _, err = tx.ExecContext(ctx, `UPDATE extension_versions SET state='approved', approved_at=? WHERE id=? AND state='imported'`, approved, versionID); err != nil {
		return nil, err
	}
	if err = r.audit(ctx, tx, seriesID, versionID, "extension.approved", "success", digestPreview, risk, "", "", correlation); err != nil {
		return nil, err
	}
	if err = r.beforeCommitHook(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	var m Manifest
	_ = json.Unmarshal([]byte(manifestJSON), &m)
	return &Version{ID: versionID, SeriesID: seriesID, Number: number, State: "approved", DigestPreview: digestPreview, Size: size, Format: format, Manifest: m, RiskState: map[bool]string{true: "HIGH_RISK", false: "NORMAL"}[len(risk) > 0], RiskCategories: risk, CreatedAt: created, ApprovedAt: approved}, nil
}

func (r *Repository) Assign(ctx context.Context, versionID, profileID, correlation string, exists ProfileExists) (*Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if strings.TrimSpace(profileID) == "" {
		return nil, ErrProfileNotFound
	}
	if exists == nil {
		return nil, ErrProfileNotFound
	}
	ok, err := exists(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrProfileNotFound
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var seriesID, state, current string
	var riskJSON, digestPreview string
	err = tx.QueryRowContext(ctx, `SELECT series_id,state,risk_categories_json,digest_preview FROM extension_versions WHERE id=?`, versionID).Scan(&seriesID, &state, &riskJSON, &digestPreview)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	if state == "revoked" || state == "quarantined" {
		return nil, ErrRevoked
	}
	if state != "approved" {
		return nil, ErrNotApproved
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(active_version_id,'') FROM extension_series WHERE id=?`, seriesID).Scan(&current); err != nil {
		return nil, err
	}
	if current != "" && current != versionID {
		_, err = tx.ExecContext(ctx, `UPDATE extension_assignments SET state='archived' WHERE version_id=? AND state='ready'`, current)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE extension_versions SET state='archived' WHERE id=? AND state='approved'`, current)
		if err != nil {
			return nil, err
		}
	}
	id, err := newID("assign")
	if err != nil {
		return nil, err
	}
	created := now()
	if _, err = tx.ExecContext(ctx, `INSERT INTO extension_assignments(id,version_id,profile_id,state,created_at) VALUES(?,?,?,?,?)`, id, versionID, profileID, "ready", created); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_series SET active_version_id=? WHERE id=?`, versionID, seriesID); err != nil {
		return nil, err
	}
	var risk []string
	_ = json.Unmarshal([]byte(riskJSON), &risk)
	if err = r.audit(ctx, tx, seriesID, versionID, "extension.assigned", "success", digestPreview, risk, profilePseudonym(profileID), "", correlation); err != nil {
		return nil, err
	}
	if err = r.beforeCommitHook(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Assignment{ID: id, VersionID: versionID, ProfileID: profileID, State: "ready", CreatedAt: created}, nil
}

func (r *Repository) Update(ctx context.Context, seriesID string, src io.Reader, correlation string) (*Version, error) {
	return r.Import(ctx, src, seriesID, correlation)
}

func (r *Repository) Rollback(ctx context.Context, seriesID, targetVersionID, correlation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var targetState, digestPreview, riskJSON, current string
	err = tx.QueryRowContext(ctx, `SELECT state,digest_preview,risk_categories_json FROM extension_versions WHERE id=? AND series_id=?`, targetVersionID, seriesID).Scan(&targetState, &digestPreview, &riskJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVersionNotFound
	}
	if err != nil {
		return err
	}
	if targetState == "revoked" || targetState == "quarantined" {
		return ErrRevoked
	}
	if targetState != "approved" && targetState != "archived" {
		return ErrNotApproved
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(active_version_id,'') FROM extension_series WHERE id=?`, seriesID).Scan(&current); err != nil {
		return err
	}
	if current == targetVersionID {
		return ErrConcurrentMutation
	}
	if current != "" {
		if _, err = tx.ExecContext(ctx, `UPDATE extension_versions SET state='archived' WHERE id=? AND state IN ('approved','ready')`, current); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE extension_assignments SET state='archived' WHERE version_id=? AND state='ready'`, current); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_versions SET state='approved' WHERE id=?`, targetVersionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_series SET active_version_id=? WHERE id=?`, targetVersionID, seriesID); err != nil {
		return err
	}
	var risk []string
	_ = json.Unmarshal([]byte(riskJSON), &risk)
	if err = r.audit(ctx, tx, seriesID, targetVersionID, "extension.rollback", "success", digestPreview, risk, "", "", correlation); err != nil {
		return err
	}
	if err = r.beforeCommitHook(); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Revoke(ctx context.Context, versionID, correlation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var seriesID, state, digestPreview, riskJSON string
	err = tx.QueryRowContext(ctx, `SELECT series_id,state,digest_preview,risk_categories_json FROM extension_versions WHERE id=?`, versionID).Scan(&seriesID, &state, &digestPreview, &riskJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVersionNotFound
	}
	if err != nil {
		return err
	}
	if state == "quarantined" || state == "revoked" {
		return nil
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_versions SET state='quarantined',revoked_at=? WHERE id=?`, now(), versionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_assignments SET state='revoked' WHERE version_id=? AND state IN ('ready','assigned')`, versionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE extension_series SET active_version_id=NULL WHERE id=? AND active_version_id=?`, seriesID, versionID); err != nil {
		return err
	}
	var risk []string
	_ = json.Unmarshal([]byte(riskJSON), &risk)
	if err = r.audit(ctx, tx, seriesID, versionID, "extension.quarantined", "success", digestPreview, risk, "", "", correlation); err != nil {
		return err
	}
	if err = r.beforeCommitHook(); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Purge(ctx context.Context, versionID, correlation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var seriesID, state, rel string
	var active string
	err = tx.QueryRowContext(ctx, `SELECT v.series_id,v.state,v.blob_relpath,COALESCE(s.active_version_id,'') FROM extension_versions v JOIN extension_series s ON s.id=v.series_id WHERE v.id=?`, versionID).Scan(&seriesID, &state, &rel, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVersionNotFound
	}
	if err != nil {
		return err
	}
	if active == versionID {
		return ErrPurgeNotAllowed
	}
	var assignments int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM extension_assignments WHERE version_id=? AND state NOT IN ('archived','revoked')`, versionID).Scan(&assignments); err != nil {
		return err
	}
	if assignments > 0 {
		return ErrPurgeNotAllowed
	}
	if state != "quarantined" && state != "revoked" && state != "archived" {
		return ErrPurgeNotAllowed
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM extension_assignments WHERE version_id=?`, versionID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM extension_versions WHERE id=?`, versionID); err != nil {
		return err
	}
	if err = r.audit(ctx, tx, seriesID, versionID, "extension.purged", "success", "", nil, "", "", correlation); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(r.baseDir, rel))
	return nil
}

func (r *Repository) audit(ctx context.Context, tx *sql.Tx, seriesID, versionID, action, result, digest string, categories []string, profile, errorCode, correlation string) error {
	if correlation == "" {
		correlation = "t28-local"
	}
	payload, _ := json.Marshal(normalizeList(categories))
	_, err := tx.ExecContext(ctx, `INSERT INTO extension_audit_events(series_id,version_id,action,result,digest_preview,permission_categories_json,profile_pseudonym,error_code,correlation_id,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, seriesID, versionID, action, result, digest, string(payload), profile, errorCode, correlation, now())
	return err
}
func (r *Repository) beforeCommitHook() error {
	if r.beforeCommit != nil {
		return r.beforeCommit()
	}
	return nil
}
func newID(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(b), nil
}
func validID(v string) bool {
	return strings.HasPrefix(v, "series_") || strings.HasPrefix(v, "version_") || strings.HasPrefix(v, "assign_")
}
func profilePseudonym(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])[:12]
}
func manifestAcknowledgement(m Manifest) []string {
	values := make([]string, 0, len(m.Permissions)+len(m.OptionalPermissions)+len(m.HostPermissions)+len(m.OptionalHostPermissions)+len(m.ContentScriptMatches))
	values = append(values, m.Permissions...)
	values = append(values, m.OptionalPermissions...)
	values = append(values, m.HostPermissions...)
	values = append(values, m.OptionalHostPermissions...)
	values = append(values, m.ContentScriptMatches...)
	return normalizeList(values)
}

func equalStrings(a, b []string) bool {
	a = normalizeList(a)
	b = normalizeList(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
