// T22 Profile History is a dedicated local SQLite repository. It records a
// redacted, immutable representation of Core-owned Profile metadata and never
// replaces the Profile JSON store or persists vault values.
package history

import (
	"context"
	"database/sql"
	"encoding/json"
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

type Snapshot struct {
	Profile profile.Profile `json:"profile"`
}

type Entry struct {
	ProfileID string    `json:"profile_id"`
	Version   int       `json:"version"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

type Version struct {
	Entry
	Snapshot Snapshot `json:"snapshot"`
}

type ListResult struct {
	Data   []Entry `json:"data"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type DiffResult struct {
	ProfileID string   `json:"profile_id"`
	From      int      `json:"from"`
	To        int      `json:"to"`
	Paths     []string `json:"paths"`
}

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

func Open(dataDir string) (*Store, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, fmt.Errorf("profile history data directory is required")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	// #nosec G302 -- owner-only Core state directory.
	if err := os.Chmod(dataDir, 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dataDir, "profile_history.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	for _, statement := range []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS profile_history_schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS profile_history_versions (
			profile_id TEXT NOT NULL, version INTEGER NOT NULL, action TEXT NOT NULL,
			snapshot_json TEXT NOT NULL, created_at TEXT NOT NULL, operation_id TEXT NOT NULL DEFAULT '',
			PRIMARY KEY(profile_id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS profile_history_audit_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT, profile_id TEXT NOT NULL, version INTEGER NULL,
			action TEXT NOT NULL, result TEXT NOT NULL, correlation_id TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO profile_history_schema_migrations(version, applied_at) VALUES(1, ?)`, timestamp()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`ALTER TABLE profile_history_versions ADD COLUMN operation_id TEXT NOT NULL DEFAULT ''`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO profile_history_schema_migrations(version, applied_at) VALUES(2, ?)`, timestamp()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Capture(ctx context.Context, p *profile.Profile, action, correlation string) (*Entry, error) {
	if p == nil || strings.TrimSpace(p.ID) == "" || !validAction(action) {
		return nil, ErrInvalidSnapshot
	}
	snapshot, err := makeSnapshot(p)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return nil, ErrInvalidSnapshot
	}
	operationID := ""
	if p.HistoryPending != nil {
		operationID = p.HistoryPending.OperationID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	next, err := nextVersion(ctx, tx, p.ID)
	if err != nil {
		return nil, err
	}
	created := now()
	if _, err = tx.ExecContext(ctx, `INSERT INTO profile_history_versions(profile_id, version, action, snapshot_json, created_at, operation_id) VALUES(?,?,?,?,?,?)`, p.ID, next, action, string(payload), created.Format(time.RFC3339Nano), operationID); err != nil {
		return nil, err
	}
	if err = audit(ctx, tx, p.ID, next, "history_created", "success", correlation); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Entry{ProfileID: p.ID, Version: next, Action: action, CreatedAt: created}, nil
}

// ReconcilePending confirms that a pending Profile write already has the same
// durable History snapshot, or records one recovery version when it does not.
// It is called at router startup before a pending marker can be cleared.
func (s *Store) ReconcilePending(ctx context.Context, p *profile.Profile, correlation string) (*Entry, error) {
	if p == nil || strings.TrimSpace(p.ID) == "" || p.HistoryPending == nil {
		return nil, ErrInvalidSnapshot
	}
	digest, err := profile.HistorySnapshotDigest(p)
	if err != nil || digest != p.HistoryPending.SnapshotDigest {
		return nil, ErrInvalidSnapshot
	}
	snapshot, err := makeSnapshot(p)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var version int
	var payload, created, operationID string
	err = tx.QueryRowContext(ctx, `SELECT version, snapshot_json, created_at, operation_id FROM profile_history_versions WHERE profile_id=? ORDER BY version DESC LIMIT 1`, p.ID).Scan(&version, &payload, &created, &operationID)
	if err == nil {
		var latest Snapshot
		if operationID == p.HistoryPending.OperationID && json.Unmarshal([]byte(payload), &latest) == nil && reflect.DeepEqual(latest, snapshot) {
			if err := audit(ctx, tx, p.ID, version, "history_recovery_confirmed", "success", correlation); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			at, err := time.Parse(time.RFC3339Nano, created)
			if err != nil {
				return nil, err
			}
			return &Entry{ProfileID: p.ID, Version: version, Action: "confirmed", CreatedAt: at}, nil
		}
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	next, err := nextVersion(ctx, tx, p.ID)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, ErrInvalidSnapshot
	}
	createdAt := now()
	if _, err := tx.ExecContext(ctx, `INSERT INTO profile_history_versions(profile_id, version, action, snapshot_json, created_at, operation_id) VALUES(?,?,?,?,?,?)`, p.ID, next, "recovery", string(encoded), createdAt.Format(time.RFC3339Nano), p.HistoryPending.OperationID); err != nil {
		return nil, err
	}
	if err := audit(ctx, tx, p.ID, next, "history_recovered", "success", correlation); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Entry{ProfileID: p.ID, Version: next, Action: "recovery", CreatedAt: createdAt}, nil
}

func (s *Store) List(ctx context.Context, profileID string, limit, offset int) (*ListResult, error) {
	if strings.TrimSpace(profileID) == "" || !validPage(limit, offset) {
		return nil, ErrInvalidVersion
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM profile_history_versions WHERE profile_id=?`, profileID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT version, action, created_at FROM profile_history_versions WHERE profile_id=? ORDER BY version DESC LIMIT ? OFFSET ?`, profileID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := &ListResult{Data: make([]Entry, 0), Total: total, Limit: limit, Offset: offset}
	for rows.Next() {
		var item Entry
		var created string
		item.ProfileID = profileID
		if err := rows.Scan(&item.Version, &item.Action, &created); err != nil {
			return nil, err
		}
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, err
		}
		result.Data = append(result.Data, item)
	}
	return result, rows.Err()
}

func (s *Store) Get(ctx context.Context, profileID string, version int) (*Version, error) {
	if strings.TrimSpace(profileID) == "" || version < 1 {
		return nil, ErrInvalidVersion
	}
	var item Version
	var payload, created string
	item.ProfileID = profileID
	err := s.db.QueryRowContext(ctx, `SELECT action, snapshot_json, created_at FROM profile_history_versions WHERE profile_id=? AND version=?`, profileID, version).Scan(&item.Action, &payload, &created)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	item.Version = version
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(payload), &item.Snapshot); err != nil {
		return nil, ErrInvalidSnapshot
	}
	return &item, nil
}

func (s *Store) Diff(ctx context.Context, profileID string, from, to int) (*DiffResult, error) {
	if from < 1 || to < 1 {
		return nil, ErrInvalidVersion
	}
	left, err := s.Get(ctx, profileID, from)
	if err != nil {
		return nil, err
	}
	right, err := s.Get(ctx, profileID, to)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	diffValue("", profileMap(left.Snapshot.Profile), profileMap(right.Snapshot.Profile), &paths)
	sort.Strings(paths)
	return &DiffResult{ProfileID: profileID, From: from, To: to, Paths: paths}, nil
}

// Restore validates the optimistic version first, invokes the caller-owned
// Profile persistence callback, and then creates a fresh immutable history
// entry. The history store never writes profile.json on its own.
func (s *Store) Restore(ctx context.Context, profileID string, target, expected int, correlation string, apply func(*profile.Profile) (*profile.Profile, error)) (*Entry, error) {
	if strings.TrimSpace(profileID) == "" || target < 1 || expected < 1 || apply == nil {
		return nil, ErrInvalidVersion
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	current, err := currentVersion(ctx, tx, profileID)
	if err != nil {
		return nil, err
	}
	if current != expected {
		if err := refuse(ctx, tx, profileID, target, "history_restore_conflict", "refused", correlation); err != nil {
			return nil, err
		}
		return nil, ErrVersionConflict
	}
	var payload string
	if err = tx.QueryRowContext(ctx, `SELECT snapshot_json FROM profile_history_versions WHERE profile_id=? AND version=?`, profileID, target).Scan(&payload); err == sql.ErrNoRows {
		if err := refuse(ctx, tx, profileID, target, "history_restore_refused", "not_found", correlation); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	var snapshot Snapshot
	if err = json.Unmarshal([]byte(payload), &snapshot); err != nil || snapshot.Profile.ID != profileID {
		if err := refuse(ctx, tx, profileID, target, "history_restore_refused", "invalid_snapshot", correlation); err != nil {
			return nil, err
		}
		return nil, ErrInvalidSnapshot
	}
	restored, err := apply(&snapshot.Profile)
	if err != nil {
		if commitErr := refuse(ctx, tx, profileID, target, "history_restore_refused", "profile_refused", correlation); commitErr != nil {
			return nil, commitErr
		}
		return nil, err
	}
	newSnapshot, err := makeSnapshot(restored)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(newSnapshot)
	if err != nil {
		return nil, err
	}
	next := current + 1
	created := now()
	operationID := ""
	if restored.HistoryPending != nil {
		operationID = restored.HistoryPending.OperationID
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO profile_history_versions(profile_id, version, action, snapshot_json, created_at, operation_id) VALUES(?,?,?,?,?,?)`, profileID, next, "restore", string(encoded), created.Format(time.RFC3339Nano), operationID); err != nil {
		return nil, err
	}
	if err = audit(ctx, tx, profileID, next, "history_restored", "success", correlation); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Entry{ProfileID: profileID, Version: next, Action: "restore", CreatedAt: created}, nil
}

func nextVersion(ctx context.Context, tx *sql.Tx, profileID string) (int, error) {
	var current int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM profile_history_versions WHERE profile_id=?`, profileID).Scan(&current); err != nil {
		return 0, err
	}
	return current + 1, nil
}

func currentVersion(ctx context.Context, tx *sql.Tx, profileID string) (int, error) {
	var current sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(version) FROM profile_history_versions WHERE profile_id=?`, profileID).Scan(&current); err != nil {
		return 0, err
	}
	if !current.Valid {
		return 0, ErrNotFound
	}
	return int(current.Int64), nil
}

func audit(ctx context.Context, tx *sql.Tx, profileID string, version int, action, result, correlation string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO profile_history_audit_events(profile_id, version, action, result, correlation_id, created_at) VALUES(?,?,?,?,?,?)`, profileID, version, action, result, correlation, timestamp())
	return err
}

func refuse(ctx context.Context, tx *sql.Tx, profileID string, version int, action, result, correlation string) error {
	if err := audit(ctx, tx, profileID, version, action, result, correlation); err != nil {
		return err
	}
	return tx.Commit()
}

func makeSnapshot(p *profile.Profile) (Snapshot, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return Snapshot{}, err
	}
	var clone profile.Profile
	if err := json.Unmarshal(data, &clone); err != nil {
		return Snapshot{}, err
	}
	clone.ProfileDir = ""
	clone.HistoryPending = nil
	if clone.Proxy != nil {
		clone.Proxy.Username = ""
		clone.Proxy.Password = ""
		clone.Proxy.SecretRef = ""
	}
	return Snapshot{Profile: clone}, nil
}

func profileMap(p profile.Profile) map[string]any {
	data, _ := json.Marshal(p)
	var value map[string]any
	_ = json.Unmarshal(data, &value)
	return value
}

func diffValue(path string, left, right any, paths *[]string) {
	if reflect.DeepEqual(left, right) {
		return
	}
	lm, lok := left.(map[string]any)
	rm, rok := right.(map[string]any)
	if lok && rok {
		keys := make(map[string]struct{}, len(lm)+len(rm))
		for key := range lm { keys[key] = struct{}{} }
		for key := range rm { keys[key] = struct{}{} }
		for key := range keys {
			next := key
			if path != "" { next = path + "." + key }
			diffValue(next, lm[key], rm[key], paths)
		}
		return
	}
	if path != "" { *paths = append(*paths, path) }
}

func validPage(limit, offset int) bool { return limit >= 1 && limit <= 100 && offset >= 0 }
func validAction(action string) bool { return action == "create" || action == "update" || action == "metadata" || action == "archive" || action == "reopen" || action == "tag_add" || action == "tag_remove" || action == "recovery" }
func now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
func timestamp() string { return now().Format(time.RFC3339Nano) }
