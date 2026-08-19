// Package templates owns the T21 TemplateRepository. It is a dedicated SQLite
// repository under the Core data directory; it neither migrates nor replaces
// the file-based Profile Store.
package templates

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"forgelocal/internal/profile"

	_ "modernc.org/sqlite"
)

const defaultPageSize = 50

type VersionState string

const (
	StateActive   VersionState = "active"
	StateArchived VersionState = "archived"
)

// Content is deliberately a closed set. Pointer scalar fields preserve the
// difference between an absent field and an explicit empty value in drafts.
type Content struct {
	Group        *string                        `json:"group,omitempty"`
	Tags         *[]string                      `json:"tags,omitempty"`
	Note         *string                        `json:"note,omitempty"`
	CustomFields map[string]profile.CustomField `json:"custom_fields,omitempty"`
}

type Series struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"active_version,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type Version struct {
	TemplateID string       `json:"template_id"`
	Version    int          `json:"version"`
	State      VersionState `json:"state"`
	Content    Content      `json:"content"`
	CreatedAt  string       `json:"created_at"`
}

type CatalogItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"active_version,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type ListResult struct {
	Data   []CatalogItem `json:"data"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type Conflict struct {
	Path string `json:"path"`
}

type DraftResult struct {
	TemplateID       string     `json:"template_id"`
	Version          int        `json:"version"`
	ValidationStatus string     `json:"validation_status"`
	Draft            *Content   `json:"draft,omitempty"`
	AppliedFields    []string   `json:"applied_fields,omitempty"`
	Conflicts        []Conflict `json:"conflicts,omitempty"`
}

type Store struct {
	db *sql.DB
	mu sync.Mutex
	// failBeforeCommit is test-only. It proves that the mutation and redacted
	// audit share one transaction. Production code leaves it nil.
	failBeforeCommit func() error
}

func Open(dataDir string) (*Store, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, fmt.Errorf("template repository data directory is required")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	// #nosec G302 -- the local Core repository directory needs owner execute permission and denies group/other access.
	if err := os.Chmod(dataDir, 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dataDir, "templates.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func migrate(db *sql.DB) error {
	statements := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS template_schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS template_series (
			id TEXT PRIMARY KEY, display_name TEXT NOT NULL, canonical_name TEXT NOT NULL,
			active_version INTEGER NULL, created_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_template_active_name ON template_series(canonical_name) WHERE active_version IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS template_versions (
			template_id TEXT NOT NULL, version INTEGER NOT NULL, state TEXT NOT NULL,
			content_json TEXT NOT NULL, created_at TEXT NOT NULL,
			PRIMARY KEY(template_id, version),
			FOREIGN KEY(template_id) REFERENCES template_series(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_template_one_active ON template_versions(template_id) WHERE state = 'active'`,
		`CREATE TABLE IF NOT EXISTS template_audit_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT, template_id TEXT NOT NULL, version INTEGER NULL,
			action TEXT NOT NULL, result TEXT NOT NULL, correlation_id TEXT NOT NULL,
			paths_json TEXT NOT NULL, created_at TEXT NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO template_schema_migrations(version, applied_at) VALUES(1, ?)`, now())
	return err
}

func (s *Store) Create(ctx context.Context, name string, content Content, correlation string) (*Version, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateContent(content); err != nil {
		return nil, err
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, ErrInvalidTemplate
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	created := now()
	if _, err = tx.ExecContext(ctx, `INSERT INTO template_series(id, display_name, canonical_name, active_version, created_at) VALUES(?,?,?,?,?)`, id, strings.TrimSpace(name), canonical(name), 1, created); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrNameActive
		}
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO template_versions(template_id, version, state, content_json, created_at) VALUES(?,?,?,?,?)`, id, 1, StateActive, string(contentJSON), created); err != nil {
		return nil, err
	}
	if err = s.audit(ctx, tx, id, 1, "template.version.created", "success", correlation, nil); err != nil {
		return nil, err
	}
	if err = s.beforeCommit(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Version{TemplateID: id, Version: 1, State: StateActive, Content: cloneContent(content), CreatedAt: created}, nil
}

func (s *Store) NewVersion(ctx context.Context, id string, expectedActive int, content Content, correlation string) (*Version, error) {
	if !validID(id) || expectedActive < 1 {
		return nil, ErrInvalidVersion
	}
	if err := validateContent(content); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(content)
	if err != nil {
		return nil, ErrInvalidTemplate
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var active sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT active_version FROM template_series WHERE id=?`, id).Scan(&active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !active.Valid {
		return nil, ErrVersionNotActive
	}
	if int(active.Int64) != expectedActive {
		return nil, ErrStaleVersion
	}
	next := expectedActive + 1
	created := now()
	if res, execErr := tx.ExecContext(ctx, `UPDATE template_versions SET state=? WHERE template_id=? AND version=? AND state=?`, StateArchived, id, expectedActive, StateActive); execErr != nil {
		return nil, execErr
	} else if n, _ := res.RowsAffected(); n != 1 {
		return nil, ErrStaleVersion
	}
	// The partial unique index enforces one active version. Archiving before the
	// insert is safe because both statements share this transaction; no reader
	// can observe an intermediate state and rollback restores the old active row.
	if _, err = tx.ExecContext(ctx, `INSERT INTO template_versions(template_id, version, state, content_json, created_at) VALUES(?,?,?,?,?)`, id, next, StateActive, string(payload), created); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE template_series SET active_version=? WHERE id=? AND active_version=?`, next, id, expectedActive); err != nil {
		return nil, err
	}
	if err = s.audit(ctx, tx, id, next, "template.version.created", "success", correlation, nil); err != nil {
		return nil, err
	}
	if err = s.beforeCommit(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Version{TemplateID: id, Version: next, State: StateActive, Content: cloneContent(content), CreatedAt: created}, nil
}

func (s *Store) Archive(ctx context.Context, id string, version int, correlation string) error {
	if !validID(id) || version < 1 {
		return ErrInvalidVersion
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE template_versions SET state=? WHERE template_id=? AND version=? AND state=?`, StateArchived, id, version, StateActive)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return ErrVersionNotActive
	}
	if _, err = tx.ExecContext(ctx, `UPDATE template_series SET active_version=NULL WHERE id=? AND active_version=?`, id, version); err != nil {
		return err
	}
	if err = s.audit(ctx, tx, id, version, "template.version.archived", "success", correlation, nil); err != nil {
		return err
	}
	if err = s.beforeCommit(); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) List(ctx context.Context, limit, offset int) (ListResult, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultPageSize
	}
	if offset < 0 {
		return ListResult{}, ErrInvalidTemplate
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM template_series`).Scan(&total); err != nil {
		return ListResult{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, display_name, active_version, created_at FROM template_series ORDER BY canonical_name, id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()
	result := ListResult{Data: []CatalogItem{}, Total: total, Limit: limit, Offset: offset}
	for rows.Next() {
		var item CatalogItem
		var active sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Name, &active, &item.CreatedAt); err != nil {
			return ListResult{}, err
		}
		if active.Valid {
			value := int(active.Int64)
			item.ActiveVersion = &value
		}
		result.Data = append(result.Data, item)
	}
	return result, rows.Err()
}

func (s *Store) GetVersion(ctx context.Context, id string, version int) (*Version, error) {
	if !validID(id) || version < 1 {
		return nil, ErrInvalidVersion
	}
	var state string
	var raw string
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT state, content_json, created_at FROM template_versions WHERE template_id=? AND version=?`, id, version).Scan(&state, &raw, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var content Content
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return nil, err
	}
	return &Version{TemplateID: id, Version: version, State: VersionState(state), Content: content, CreatedAt: created}, nil
}

func (s *Store) Draft(ctx context.Context, id string, version int, base Content, correlation string) (*DraftResult, error) {
	v, err := s.GetVersion(ctx, id, version)
	if err != nil {
		return nil, err
	}
	if v.State != StateActive {
		_ = s.auditStandalone(ctx, id, version, "template.draft", "refused", correlation, nil)
		return nil, ErrVersionNotActive
	}
	if err := validateContent(base); err != nil {
		_ = s.auditStandalone(ctx, id, version, "template.draft", "refused", correlation, nil)
		return nil, err
	}
	draft, fields, conflicts := mergeDraft(base, v.Content)
	if len(conflicts) > 0 {
		paths := make([]string, 0, len(conflicts))
		for _, conflict := range conflicts {
			paths = append(paths, conflict.Path)
		}
		_ = s.auditStandalone(ctx, id, version, "template.draft", "conflict", correlation, paths)
		return &DraftResult{TemplateID: id, Version: version, ValidationStatus: "CONFLICT", Conflicts: conflicts}, ErrConflict
	}
	if err := validateContent(draft); err != nil {
		_ = s.auditStandalone(ctx, id, version, "template.draft", "refused", correlation, nil)
		return nil, err
	}
	if err := s.auditStandalone(ctx, id, version, "template.draft", "success", correlation, nil); err != nil {
		return nil, err
	}
	return &DraftResult{TemplateID: id, Version: version, ValidationStatus: "VALID", Draft: &draft, AppliedFields: fields}, nil
}

func (s *Store) AuditCount(ctx context.Context, id string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM template_audit_events WHERE template_id=?`, id).Scan(&n)
	return n, err
}

func (s *Store) SetFailBeforeCommitForTest(fn func() error) { s.failBeforeCommit = fn }

func (s *Store) beforeCommit() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failBeforeCommit == nil {
		return nil
	}
	return s.failBeforeCommit()
}

func (s *Store) auditStandalone(ctx context.Context, id string, version int, action, result, correlation string, paths []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = s.audit(ctx, tx, id, version, action, result, correlation, paths); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) audit(ctx context.Context, tx *sql.Tx, id string, version int, action, result, correlation string, paths []string) error {
	payload, err := json.Marshal(paths)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO template_audit_events(template_id, version, action, result, correlation_id, paths_json, created_at) VALUES(?,?,?,?,?,?,?)`, id, version, action, result, correlation, string(payload), now())
	return err
}

func validateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > 128 {
		return ErrInvalidTemplateName
	}
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			return ErrInvalidTemplateName
		}
	}
	return nil
}

func validateContent(content Content) error {
	if err := profile.ValidateTemplateMetadata(content.Group, dereferenceTags(content.Tags), content.Note, content.CustomFields); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}
	return nil
}

func dereferenceTags(tags *[]string) []string {
	if tags == nil {
		return nil
	}
	return *tags
}

func mergeDraft(base, template Content) (Content, []string, []Conflict) {
	draft := cloneContent(base)
	applied := []string{}
	conflicts := []Conflict{}
	mergeScalar := func(path string, current, incoming *string, set func(*string)) {
		if incoming == nil {
			return
		}
		if current == nil {
			value := *incoming
			set(&value)
			applied = append(applied, path)
			return
		}
		if *current != *incoming {
			conflicts = append(conflicts, Conflict{Path: path})
		}
	}
	mergeScalar("group", draft.Group, template.Group, func(v *string) { draft.Group = v })
	mergeScalar("note", draft.Note, template.Note, func(v *string) { draft.Note = v })
	if template.Tags != nil {
		union := append([]string{}, dereferenceTags(draft.Tags)...)
		union = append(union, (*template.Tags)...)
		canonical := make(map[string]string, len(union))
		for _, tag := range union {
			key := profile.CanonicalMetadataName(tag)
			if _, found := canonical[key]; !found {
				canonical[key] = tag
			}
		}
		keys := make([]string, 0, len(canonical))
		for key := range canonical {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make([]string, 0, len(keys))
		for _, key := range keys {
			out = append(out, canonical[key])
		}
		draft.Tags = &out
		applied = append(applied, "tags")
	}
	if template.CustomFields != nil {
		if draft.CustomFields == nil {
			draft.CustomFields = map[string]profile.CustomField{}
		}
		for key, value := range template.CustomFields {
			if current, exists := draft.CustomFields[key]; exists {
				if !reflect.DeepEqual(current, value) {
					conflicts = append(conflicts, Conflict{Path: "custom_fields." + key})
				}
				continue
			}
			draft.CustomFields[key] = value
			applied = append(applied, "custom_fields."+key)
		}
	}
	return draft, applied, conflicts
}

func cloneContent(content Content) Content {
	data, _ := json.Marshal(content)
	var clone Content
	_ = json.Unmarshal(data, &clone)
	return clone
}

func canonical(name string) string { return profile.CanonicalMetadataName(name) }

func validID(value string) bool {
	if len(value) != 32 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'f') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func newID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }
