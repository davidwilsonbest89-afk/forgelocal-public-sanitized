// Package runtime — qualification (T14 clean-room reimplementation).
//
// Contract:
//   - A runtime is Qualified only after real local observation: the binary
//     must exist, be executable, have a deterministic SHA-256 of its first
//     block-stable bytes is NOT allowed as a shortcut — the whole binary is
//     hashed, and `headless --version` (or equivalent probe with timeout)
//     must return a real version string.
//   - State machine per runtime id: NotAttempted → Probing →
//     Qualified(version, hash, at) / Failed(reason). Transitions are
//     upserted atomically in SQLite.
//   - Probe is fail-closed on panic: temp dir removed, process killed.
//   - The qualified registry NEVER exposes BinaryPath, absolute paths or
//     internal filesystem details to any caller (G15-A projection).
package runtime

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// QualificationState is the machine state of a runtime qualification record.
type QualificationState string

const (
	QSNotAttempted QualificationState = "NOT_ATTEMPTED"
	QSProbing      QualificationState = "PROBING"
	QSQualified    QualificationState = "QUALIFIED"
	QSFailed       QualificationState = "FAILED"
)

// QualifiedInfo is the redacted projection returned by the registry:
// never contains BinaryPath or filesystem paths.
type QualifiedInfo struct {
	ID           string             `json:"id"`
	State        QualificationState `json:"state"`
	Version      string             `json:"version,omitempty"`
	BinaryHash   string             `json:"binary_hash_sha256,omitempty"`
	QualifiedAt  *time.Time         `json:"qualified_at,omitempty"`
	FailedReason string             `json:"failed_reason,omitempty"`
}

var (
	ErrRuntimeUnknown      = errors.New("runtime: runtime id not registered")
	ErrRuntimeNotQualified = errors.New("runtime: not qualified")
	ErrProbeTimeout        = errors.New("runtime: probe timeout")
	ErrProbeNoVersion      = errors.New("runtime: probe returned no version")
	ErrBinaryNotExecutable = errors.New("runtime: binary missing or not executable")
)

// DB is the minimal persistence surface shared with the Core store.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// probeCmd is the executable probe; injectable for tests without spawning a
// real browser process.
type probeCmd func(ctx context.Context, binaryPath string) (version string, err error)

// defaultHeadlessProbe runs `<binary> --headless --version` with a strict
// timeout and an ephemeral working directory; it never opens a profile or a
// window.
func defaultHeadlessProbe(ctx context.Context, binaryPath string) (string, error) {
	if _, err := os.Stat(binaryPath); err != nil {
		return "", ErrBinaryNotExecutable
	}
	tmp, err := os.MkdirTemp("", "fl-t14-probe-")
	if err != nil {
		return "", fmt.Errorf("runtime: tmp dir: %w", err)
	}
	defer os.RemoveAll(tmp) // fail-closed cleanup
	probeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, binaryPath, "--headless", "--version")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	// CommandContext returns a generic "signal killed"/"killed" error on
	// timeout; map it explicitly.
	if probeCtx.Err() == context.DeadlineExceeded {
		return "", ErrProbeTimeout
	}
	if err != nil {
		return "", fmt.Errorf("runtime: probe exit: %w", err)
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return "", ErrProbeNoVersion
	}
	return v, nil
}

// Qualifier discovers, hashes and probes a runtime binary, then persists the
// machine state.
type Qualifier struct {
	db    DB
	probe probeCmd
}

func NewQualifier(db DB) *Qualifier {
	return &Qualifier{db: db, probe: defaultHeadlessProbe}
}

// Migrate creates the qualification table idempotently.
func (q *Qualifier) Migrate(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS runtime_qualifications (
		runtime_id TEXT PRIMARY KEY,
		state TEXT NOT NULL,
		version TEXT NOT NULL DEFAULT '',
		binary_hash_sha256 TEXT NOT NULL DEFAULT '',
		qualified_at TEXT,
		failed_reason TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL
	)`)
	return err
}

// Qualify runs discovery + hash + probe for id and upserts the machine state.
func (q *Qualifier) Qualify(ctx context.Context, id ID, binaryPath string) (*QualifiedInfo, error) {
	if binaryPath == "" {
		return nil, fmt.Errorf("runtime: %w", ErrBinaryNotExecutable)
	}
	if err := q.record(ctx, id, QSProbing, "", "", nil, ""); err != nil {
		return nil, fmt.Errorf("runtime: state: %w", err)
	}
	// Fail-closed: record FAILED on any panic path below.
	var result *QualifiedInfo
	func() {
		defer func() {
			if recover() != nil {
				_ = q.record(ctx, id, QSFailed, "", "", nil, "probe panic")
				result = q.info(ctx, id)
			}
		}()

		hash, err := hashBinary(binaryPath)
		if err != nil {
			_ = q.record(ctx, id, QSFailed, "", "", nil, "hash: "+err.Error())
			result = q.info(ctx, id)
			return
		}
		version, err := q.probe(ctx, binaryPath)
		if err != nil {
			_ = q.record(ctx, id, QSFailed, "", hash, nil, "probe: "+err.Error())
			result = q.info(ctx, id)
			return
		}
		if strings.TrimSpace(version) == "" {
			_ = q.record(ctx, id, QSFailed, "", hash, nil, "probe: "+ErrProbeNoVersion.Error())
			result = q.info(ctx, id)
			return
		}
		now := time.Now().UTC()
		_ = q.record(ctx, id, QSQualified, version, hash, &now, "")
		result = q.info(ctx, id)
	}()
	return result, nil
}

func hashBinary(path string) (string, error) {
	f, err := openQualifiedBinary(path)
	if err != nil {
		return "", ErrBinaryNotExecutable
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", closeBinaryWithError(f, fmt.Errorf("hash read: %w", err))
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("hash close: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// openQualifiedBinary accepts only an absolute runtime path and opens its final
// component through os.Root. This confines traversal and symlink resolution to
// the binary’s parent directory, rather than handing an unchecked string to a
// filesystem open. The path is supplied by the internal runtime catalog, never
// by an API caller; confinement still makes a catalog/configuration mistake
// fail closed.
func openQualifiedBinary(path string) (*os.File, error) {
	if !filepath.IsAbs(path) {
		return nil, ErrBinaryNotExecutable
	}
	clean := filepath.Clean(path)
	base := filepath.Base(clean)
	if base == "." || base == string(filepath.Separator) {
		return nil, ErrBinaryNotExecutable
	}
	root, err := os.OpenRoot(filepath.Dir(clean))
	if err != nil {
		return nil, ErrBinaryNotExecutable
	}
	f, openErr := root.Open(base)
	rootCloseErr := root.Close()
	if openErr != nil {
		if rootCloseErr != nil {
			return nil, errors.Join(openErr, rootCloseErr)
		}
		return nil, ErrBinaryNotExecutable
	}
	if rootCloseErr != nil {
		return nil, closeBinaryWithError(f, rootCloseErr)
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return nil, closeBinaryWithError(f, ErrBinaryNotExecutable)
	}
	return f, nil
}

func closeBinaryWithError(f *os.File, opErr error) error {
	if closeErr := f.Close(); closeErr != nil {
		return errors.Join(opErr, closeErr)
	}
	return opErr
}

func (q *Qualifier) record(ctx context.Context, id ID, state QualificationState, version, hash string, at *time.Time, reason string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	atS := ""
	if at != nil {
		atS = at.Format(time.RFC3339)
	}
	_, err := q.db.ExecContext(ctx,
		`INSERT INTO runtime_qualifications (runtime_id, state, version, binary_hash_sha256, qualified_at, failed_reason, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT(runtime_id) DO UPDATE SET
			state=excluded.state, version=excluded.version,
			binary_hash_sha256=excluded.binary_hash_sha256,
			qualified_at=excluded.qualified_at,
			failed_reason=excluded.failed_reason, updated_at=$7`,
		string(id), string(state), version, hash, atS, reason, now)
	return err
}

func (q *Qualifier) info(ctx context.Context, id ID) *QualifiedInfo {
	row := q.db.QueryRowContext(ctx,
		`SELECT state, version, binary_hash_sha256, qualified_at, failed_reason FROM runtime_qualifications WHERE runtime_id=$1`, string(id))
	info := &QualifiedInfo{ID: string(id), State: QSNotAttempted}
	var at sql.NullString
	if err := row.Scan(&info.State, &info.Version, &info.BinaryHash, &at, &info.FailedReason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return info
		}
		info.FailedReason = "lookup: " + err.Error()
		info.State = QSFailed
		return info
	}
	if at.Valid {
		if t, err := time.Parse(time.RFC3339, at.String); err == nil {
			info.QualifiedAt = &t
		}
	}
	return info
}

// QualifiedRegistry is the redacted query surface: ListQualified/Get/RequireQualified
// never surface filesystem paths.
type QualifiedRegistry struct {
	db DB
}

func NewQualifiedRegistry(db DB) *QualifiedRegistry { return &QualifiedRegistry{db: db} }

func (r *QualifiedRegistry) Get(ctx context.Context, id ID) (*QualifiedInfo, error) {
	info := (&Qualifier{db: r.db}).info(ctx, id)
	if info.State != QSQualified {
		return info, fmt.Errorf("%w: %s (%s)", ErrRuntimeNotQualified, id, info.State)
	}
	return info, nil
}

func (r *QualifiedRegistry) ListQualified(ctx context.Context) ([]QualifiedInfo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT runtime_id, state, version, binary_hash_sha256, qualified_at, failed_reason FROM runtime_qualifications WHERE state=$1 ORDER BY runtime_id`, string(QSQualified))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []QualifiedInfo
	for rows.Next() {
		var info QualifiedInfo
		var at sql.NullString
		if err := rows.Scan(&info.ID, &info.State, &info.Version, &info.BinaryHash, &at, &info.FailedReason); err != nil {
			return nil, err
		}
		if at.Valid {
			if t, err := time.Parse(time.RFC3339, at.String); err == nil {
				info.QualifiedAt = &t
			}
		}
		out = append(out, info)
	}
	if out == nil {
		out = []QualifiedInfo{}
	}
	return out, rows.Err()
}

func (r *QualifiedRegistry) RequireQualified(ctx context.Context, id ID) (*QualifiedInfo, error) {
	return r.Get(ctx, id)
}
