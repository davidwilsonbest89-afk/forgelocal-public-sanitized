package launch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SQLiteOperationJournal persists T18 operation states and a redacted audit
// event in one SQLite transaction. It does not retain a runtime handle,
// binary path, proxy configuration or secret.
type SQLiteOperationJournal struct {
	db *sql.DB
}

func NewSQLiteOperationJournal(ctx context.Context, db *sql.DB) (*SQLiteOperationJournal, error) {
	if db == nil {
		return nil, ErrJournalUnavailable
	}
	j := &SQLiteOperationJournal{db: db}
	if err := j.migrate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	return j, nil
}

func (j *SQLiteOperationJournal) migrate(ctx context.Context) error {
	_, err := j.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS launch_operations (
			session_id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			idempotency_key TEXT NOT NULL UNIQUE,
			state TEXT NOT NULL,
			correlation_id TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS launch_operations_active_profile
			ON launch_operations(profile_id)
			WHERE state IN ('queued', 'starting', 'running', 'stopping');
		CREATE TABLE IF NOT EXISTS launch_operation_audit (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			from_state TEXT NOT NULL DEFAULT '',
			to_state TEXT NOT NULL DEFAULT '',
			correlation_id TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
	`)
	return err
}

func (j *SQLiteOperationJournal) Reserve(ctx context.Context, in JournalOperation) (JournalOperation, bool, error) {
	if in.SessionID == "" || in.ProfileID == "" || in.IdempotencyKey == "" {
		return JournalOperation{}, false, ErrJournalUnavailable
	}
	in.CorrelationID = redactCorrelationID(in.CorrelationID)
	in.Reason = redacted(in.Reason)
	now := time.Now().UTC()
	in.State = StateQueued
	in.CreatedAt, in.UpdatedAt = now, now

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `INSERT INTO launch_operations
		(session_id, profile_id, idempotency_key, state, correlation_id, reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(idempotency_key) DO NOTHING`,
		in.SessionID, in.ProfileID, in.IdempotencyKey, in.State, in.CorrelationID, in.Reason,
		in.CreatedAt.Format(time.RFC3339Nano), in.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return JournalOperation{}, false, mapSQLiteJournalError(err)
	}
	created, err := result.RowsAffected()
	if err != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	if created == 0 {
		op, err := selectJournalOperation(ctx, tx, in.IdempotencyKey)
		if err != nil {
			return JournalOperation{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
		}
		return op, false, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO launch_operation_audit
		(session_id,event_type,from_state,to_state,correlation_id,reason,created_at)
		VALUES (?, 'operation_reserved', '', ?, ?, ?, ?)`,
		in.SessionID, in.State, in.CorrelationID, in.Reason, now.Format(time.RFC3339Nano)); err != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	if err := tx.Commit(); err != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	return in, true, nil
}

func (j *SQLiteOperationJournal) Lookup(ctx context.Context, idempotencyKey string) (JournalOperation, bool, error) {
	if idempotencyKey == "" {
		return JournalOperation{}, false, nil
	}
	var op JournalOperation
	var created, updated string
	err := j.db.QueryRowContext(ctx, `SELECT session_id,profile_id,idempotency_key,state,correlation_id,reason,created_at,updated_at
		FROM launch_operations WHERE idempotency_key=?`, idempotencyKey).Scan(
		&op.SessionID, &op.ProfileID, &op.IdempotencyKey, &op.State, &op.CorrelationID, &op.Reason, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return JournalOperation{}, false, nil
	}
	if err != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	var parseErr error
	op.CreatedAt, parseErr = time.Parse(time.RFC3339Nano, created)
	if parseErr == nil {
		op.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, updated)
	}
	if parseErr != nil {
		return JournalOperation{}, false, fmt.Errorf("%w: %v", ErrJournalUnavailable, parseErr)
	}
	return op, true, nil
}

func (j *SQLiteOperationJournal) Transition(ctx context.Context, session sessionID, from, to LaunchState, reason, correlationID string) error {
	if !validTransition(from, to) || session == "" {
		return ErrInvalidTransition
	}
	reason = redacted(reason)
	correlationID = redactCorrelationID(correlationID)
	now := time.Now().UTC()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE launch_operations
		SET state = ?, reason = ?, correlation_id = ?, updated_at = ?
		WHERE session_id = ? AND state = ?`, to, reason, correlationID, now.Format(time.RFC3339Nano), session, from)
	if err != nil {
		return mapSQLiteJournalError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ErrInvalidTransition
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO launch_operation_audit
		(session_id,event_type,from_state,to_state,correlation_id,reason,created_at)
		VALUES (?, 'operation_transition', ?, ?, ?, ?, ?)`, session, from, to, correlationID, reason, now.Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	return nil
}

func (j *SQLiteOperationJournal) Reconcile(ctx context.Context) ([]RecoveredSession, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT session_id, profile_id, state, correlation_id, updated_at FROM launch_operations
		ORDER BY created_at, session_id`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	defer rows.Close()
	var recs []RecoveredSession
	type change struct {
		id          sessionID
		from, to    LaunchState
		correlation string
	}
	var changes []change
	for rows.Next() {
		var id, profile, state, correlation, updated string
		if err := rows.Scan(&id, &profile, &state, &correlation, &updated); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
		}
		stamp, err := time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
		}
		from := LaunchState(state)
		switch from {
		case StateInterrupted, StateStopped, StateError:
			continue
		case StateQueued, StateStarting, StateRunning, StateStopping:
		default:
			return nil, fmt.Errorf("%w: unknown persisted state %q", ErrJournalUnavailable, state)
		}
		recs = append(recs, RecoveredSession{SessionID: sessionID(id), ProfileID: profileID(profile), LastState: from, LastRecorded: stamp})
		to := StateStopped
		if from == StateQueued {
			to = StateInterrupted
		}
		changes = append(changes, change{id: sessionID(id), from: from, to: to, correlation: redactCorrelationID(correlation)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, c := range changes {
		if _, err := tx.ExecContext(ctx, `UPDATE launch_operations SET state=?, reason='crash reconciliation', updated_at=? WHERE session_id=? AND state=?`, c.to, now, c.id, c.from); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO launch_operation_audit
			(session_id,event_type,from_state,to_state,correlation_id,reason,created_at)
			VALUES (?, 'crash_recovery', ?, ?, ?, 'crash reconciliation', ?)`, c.id, c.from, c.to, c.correlation, now); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	return recs, nil
}

func selectJournalOperation(ctx context.Context, tx *sql.Tx, idempotencyKey string) (JournalOperation, error) {
	var op JournalOperation
	var created, updated string
	err := tx.QueryRowContext(ctx, `SELECT session_id,profile_id,idempotency_key,state,correlation_id,reason,created_at,updated_at
		FROM launch_operations WHERE idempotency_key=?`, idempotencyKey).Scan(
		&op.SessionID, &op.ProfileID, &op.IdempotencyKey, &op.State, &op.CorrelationID, &op.Reason, &created, &updated)
	if err != nil {
		return JournalOperation{}, fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
	}
	var parseErr error
	op.CreatedAt, parseErr = time.Parse(time.RFC3339Nano, created)
	if parseErr == nil {
		op.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, updated)
	}
	if parseErr != nil {
		return JournalOperation{}, fmt.Errorf("%w: %v", ErrJournalUnavailable, parseErr)
	}
	return op, nil
}

func mapSQLiteJournalError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrJournalUnavailable
	}
	return fmt.Errorf("%w: %v", ErrJournalUnavailable, err)
}
