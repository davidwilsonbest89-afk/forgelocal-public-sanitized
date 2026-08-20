package launch

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newT18Journal(t *testing.T) (*sql.DB, *SQLiteOperationJournal) {
	t.Helper()
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open SQLite test journal: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	journal, err := NewSQLiteOperationJournal(context.Background(), db)
	if err != nil {
		t.Fatalf("create SQLite journal: %v", err)
	}
	return db, journal
}

func reserveT18(t *testing.T, journal *SQLiteOperationJournal, session sessionID, profile profileID, key, correlation string) {
	t.Helper()
	_, created, err := journal.Reserve(context.Background(), JournalOperation{
		SessionID: session, ProfileID: profile, IdempotencyKey: key,
		CorrelationID: correlation,
	})
	if err != nil || !created {
		t.Fatalf("reserve journal operation: created=%v err=%v", created, err)
	}
}

type countingLauncher struct{ calls atomic.Int64 }

func (l *countingLauncher) Attach(context.Context, Session) error {
	l.calls.Add(1)
	return nil
}

func TestT18Journal_IdempotencePreventsDoubleAttachAfterCrash(t *testing.T) {
	_, journal := newT18Journal(t)
	const key = "op-idempotent-001"
	reserveT18(t, journal, "sess-t18-idempotent", "profile-t18-idempotent", key, "corr-t18-001")

	m, _ := newTestManager(&Options{Journal: journal})
	recovered := m.Recover(nil) // simulated process restart; no auto-attach is permitted.
	if len(recovered) != 1 || recovered[0].State != StateInterrupted {
		t.Fatalf("expected one interrupted recovery record, got %#v", recovered)
	}
	launcher := &countingLauncher{}
	ctx := WithOperationMetadata(context.Background(), key, "corr-t18-001")
	session, err := m.Request(ctx, launcher, "profile-t18-idempotent")
	if err != nil {
		t.Fatalf("idempotent request: %v", err)
	}
	if session.ID != "sess-t18-idempotent" || session.State != StateInterrupted {
		t.Fatalf("expected recovered durable operation, got %#v", session)
	}
	if launcher.calls.Load() != 0 {
		t.Fatalf("idempotent recovery attached %d time(s)", launcher.calls.Load())
	}
}

func TestT18Journal_CancelBeforeExecutionLeavesNoPartialWrite(t *testing.T) {
	db, journal := newT18Journal(t)
	m, _ := newTestManager(&Options{GlobalLimit: 1, MaxQueueDepth: 2, WaitDeadline: time.Second, Journal: journal})
	blocker := newBlockingLauncher()
	go func() { _, _ = m.Request(context.Background(), blocker, "profile-t18-busy") }()
	waitForRunning(t, m, 1)

	ctx, cancel := context.WithCancel(WithOperationMetadata(context.Background(), "op-cancel-before", "corr-t18-cancel"))
	cancel()
	_, err := m.Request(ctx, newBlockingLauncher(), "profile-t18-cancelled")
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("expected cancellation before execution, got %v", err)
	}
	var operations, auditEvents int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operations WHERE idempotency_key='op-cancel-before'`).Scan(&operations); err != nil {
		t.Fatalf("count cancelled operation: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operation_audit WHERE session_id IN (SELECT session_id FROM launch_operations WHERE idempotency_key='op-cancel-before')`).Scan(&auditEvents); err != nil {
		t.Fatalf("count cancelled audit: %v", err)
	}
	if operations != 0 || auditEvents != 0 {
		t.Fatalf("cancelled-before-admission left partial SQLite state: operations=%d audit=%d", operations, auditEvents)
	}
	blocker.releaseAll()
	m.Stop(context.Background())
}

func TestT18Journal_CrashRecoveryAfterStartingStopsWithoutAttach(t *testing.T) {
	_, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-starting", "profile-t18-starting", "op-starting", "corr-t18-starting")
	if err := journal.Transition(context.Background(), "sess-t18-starting", StateQueued, StateStarting, "", "corr-t18-starting"); err != nil {
		t.Fatalf("persist starting transition: %v", err)
	}
	m, _ := newTestManager(&Options{Journal: journal})
	recovered := m.Recover(nil)
	if len(recovered) != 1 || recovered[0].State != StateStopped {
		t.Fatalf("expected stopped recovered session, got %#v", recovered)
	}
	if status := m.Status(); status.Running != 0 || status.Queued != 0 {
		t.Fatalf("recovery must not auto-execute: %#v", status)
	}
}

func TestT18Journal_ManagerRestartRecoveryIsIdempotentAndDoesNotAttach(t *testing.T) {
	db, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-restart", "profile-t18-restart", "op-restart", "corr-t18-restart")
	if err := journal.Transition(context.Background(), "sess-t18-restart", StateQueued, StateStarting, "", "corr-t18-restart"); err != nil {
		t.Fatalf("persist pre-crash starting state: %v", err)
	}
	m, _ := newTestManager(&Options{GlobalLimit: 1, Journal: journal})
	first := m.Recover(nil)
	if len(first) != 1 || first[0].State != StateStopped {
		t.Fatalf("unexpected first restart recovery: %#v", first)
	}
	m.mu.Lock()
	got := len(m.running)
	m.mu.Unlock()
	if got != 0 {
		t.Fatalf("restart recovery must not attach or retain a session, got %d", got)
	}
	second := m.Recover(nil)
	if len(second) != 0 {
		t.Fatalf("recovery replayed a terminal operation: %#v", second)
	}
	var state string
	if err := db.QueryRow(`SELECT state FROM launch_operations WHERE idempotency_key='op-restart'`).Scan(&state); err != nil {
		t.Fatalf("read recovered operation: %v", err)
	}
	if state != string(StateStopped) {
		t.Fatalf("restart recovery did not persist terminal stopped state: %q", state)
	}
}

func TestT18Journal_CancelDuringAttachPersistsInterrupted(t *testing.T) {
	db, journal := newT18Journal(t)
	m, _ := newTestManager(&Options{GlobalLimit: 1, StartTimeout: time.Second, Journal: journal})
	launcher := newBlockingLauncher()
	ctx, cancel := context.WithCancel(WithOperationMetadata(context.Background(), "op-cancel-during", "corr-t18-during"))
	done := make(chan error, 1)
	go func() {
		_, err := m.Request(ctx, launcher, "profile-t18-during")
		done <- err
	}()
	waitForRunning(t, m, 1)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, ErrCancelled) {
			t.Fatalf("expected cancellation during attach, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled attach did not resolve")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var state string
		err := db.QueryRow(`SELECT state FROM launch_operations WHERE idempotency_key='op-cancel-during'`).Scan(&state)
		if err == nil && state == string(StateInterrupted) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("cancelled attach was not durably marked interrupted")
}

func TestT18Journal_OperationTimeoutPersistsError(t *testing.T) {
	db, journal := newT18Journal(t)
	m, _ := newTestManager(&Options{GlobalLimit: 1, StartTimeout: 25 * time.Millisecond, Journal: journal})
	session, err := m.Request(WithOperationMetadata(context.Background(), "op-timeout", "corr-t18-timeout"), newBlockingLauncher(), "profile-t18-timeout")
	if err != nil {
		t.Fatalf("timed out attach should return its final session, got %v", err)
	}
	if session.State != StateError {
		t.Fatalf("expected timeout final state error, got %#v", session)
	}
	var state string
	if err := db.QueryRow(`SELECT state FROM launch_operations WHERE idempotency_key='op-timeout'`).Scan(&state); err != nil {
		t.Fatalf("read timeout operation: %v", err)
	}
	if state != string(StateError) {
		t.Fatalf("expected durable timeout state error, got %q", state)
	}
}

func TestT18Journal_InvalidTransitionHasNoPartialAuditWrite(t *testing.T) {
	db, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-invalid", "profile-t18-invalid", "op-invalid", "corr-t18-invalid")
	var before int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operation_audit WHERE session_id='sess-t18-invalid'`).Scan(&before); err != nil {
		t.Fatalf("count audit before invalid transition: %v", err)
	}
	err := journal.Transition(context.Background(), "sess-t18-invalid", StateQueued, StateRunning, "must-not-commit", "corr-t18-invalid")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
	var after int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operation_audit WHERE session_id='sess-t18-invalid'`).Scan(&after); err != nil {
		t.Fatalf("count audit after invalid transition: %v", err)
	}
	if after != before {
		t.Fatalf("invalid transition wrote partial audit: before=%d after=%d", before, after)
	}
}

func TestT18Journal_ActiveProfileContentionLeavesNoPartialOperation(t *testing.T) {
	db, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-profile-a", "profile-t18-contention", "op-profile-a", "corr-t18-a")
	_, created, err := journal.Reserve(context.Background(), JournalOperation{
		SessionID: "sess-t18-profile-b", ProfileID: "profile-t18-contention", IdempotencyKey: "op-profile-b",
		CorrelationID: "corr-t18-b",
	})
	if created || !errors.Is(err, ErrJournalUnavailable) {
		t.Fatalf("expected durable active-profile refusal, created=%v err=%v", created, err)
	}
	var operations, audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operations WHERE profile_id='profile-t18-contention'`).Scan(&operations); err != nil {
		t.Fatalf("count profile operations: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operation_audit WHERE session_id IN ('sess-t18-profile-a','sess-t18-profile-b')`).Scan(&audits); err != nil {
		t.Fatalf("count profile audits: %v", err)
	}
	if operations != 1 || audits != 1 {
		t.Fatalf("profile contention left partial journal records: operations=%d audits=%d", operations, audits)
	}
}

func TestT18Journal_DistinctProfilesReserveIndependently(t *testing.T) {
	db, journal := newT18Journal(t)
	for _, op := range []JournalOperation{
		{SessionID: "sess-t18-distinct-a", ProfileID: "profile-t18-distinct-a", IdempotencyKey: "op-distinct-a", CorrelationID: "corr-t18-distinct-a"},
		{SessionID: "sess-t18-distinct-b", ProfileID: "profile-t18-distinct-b", IdempotencyKey: "op-distinct-b", CorrelationID: "corr-t18-distinct-b"},
	} {
		_, created, err := journal.Reserve(context.Background(), op)
		if err != nil || !created {
			t.Fatalf("distinct profile reserve: created=%v err=%v", created, err)
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operations WHERE profile_id IN ('profile-t18-distinct-a','profile-t18-distinct-b')`).Scan(&count); err != nil {
		t.Fatalf("count distinct profile operations: %v", err)
	}
	if count != 2 {
		t.Fatalf("distinct profiles contended unexpectedly: operations=%d", count)
	}
}

func TestT18Journal_ReconcileRejectsIncoherentStateWithoutPartialWrite(t *testing.T) {
	db, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-incoherent", "profile-t18-incoherent", "op-incoherent", "corr-t18-incoherent")
	if _, err := db.Exec(`UPDATE launch_operations SET state='corrupt-state' WHERE session_id='sess-t18-incoherent'`); err != nil {
		t.Fatalf("inject incoherent state: %v", err)
	}
	if _, err := journal.Reconcile(context.Background()); !errors.Is(err, ErrJournalUnavailable) {
		t.Fatalf("expected fail-closed incoherent-state error, got %v", err)
	}
	var state string
	var audits int
	if err := db.QueryRow(`SELECT state FROM launch_operations WHERE session_id='sess-t18-incoherent'`).Scan(&state); err != nil {
		t.Fatalf("read incoherent state: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operation_audit WHERE session_id='sess-t18-incoherent'`).Scan(&audits); err != nil {
		t.Fatalf("count incoherent-state audits: %v", err)
	}
	if state != "corrupt-state" || audits != 1 {
		t.Fatalf("incoherent-state reconciliation wrote partial data: state=%q audits=%d", state, audits)
	}
}

func TestT18Journal_ReconcileKeepsCorrelationInCrashAudit(t *testing.T) {
	db, journal := newT18Journal(t)
	reserveT18(t, journal, "sess-t18-correlation", "profile-t18-correlation", "op-correlation", "c18-r1")
	if _, err := journal.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile queued operation: %v", err)
	}
	var correlation string
	if err := db.QueryRow(`SELECT correlation_id FROM launch_operation_audit WHERE session_id='sess-t18-correlation' AND event_type='crash_recovery'`).Scan(&correlation); err != nil {
		t.Fatalf("read crash audit correlation: %v", err)
	}
	if correlation != "c18-r1" {
		t.Fatalf("crash audit lost correlation id: %q", correlation)
	}
}

func TestT18Journal_QueueSaturationLeavesNoPartialOperation(t *testing.T) {
	db, journal := newT18Journal(t)
	m, _ := newTestManager(&Options{GlobalLimit: 1, MaxQueueDepth: 1, StartTimeout: time.Second, Journal: journal})
	launcher := newBlockingLauncher()
	firstDone := make(chan error, 1)
	go func() {
		_, err := m.Request(WithOperationMetadata(context.Background(), "op-saturation-running", "corr-t18-running"), launcher, "profile-t18-running")
		firstDone <- err
	}()
	waitForRunning(t, m, 1)
	secondDone := make(chan error, 1)
	go func() {
		_, err := m.Request(WithOperationMetadata(context.Background(), "op-saturation-queued", "corr-t18-queued"), launcher, "profile-t18-queued")
		secondDone <- err
	}()
	waitForQueueDepth(t, m, 1)
	if _, err := m.Request(WithOperationMetadata(context.Background(), "op-saturation-refused", "corr-t18-refused"), launcher, "profile-t18-refused"); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected queue saturation refusal, got %v", err)
	}
	var refused int
	if err := db.QueryRow(`SELECT COUNT(*) FROM launch_operations WHERE idempotency_key='op-saturation-refused'`).Scan(&refused); err != nil {
		t.Fatalf("count refused operation: %v", err)
	}
	if refused != 0 {
		t.Fatalf("queue refusal left %d durable operations", refused)
	}
	m.Stop(context.Background())
	for _, done := range []chan error{firstDone, secondDone} {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("saturated queue did not drain")
		}
	}
}

func TestT18Journal_AuditRedactsSecretAndInvalidCorrelation(t *testing.T) {
	db, journal := newT18Journal(t)
	_, created, err := journal.Reserve(context.Background(), JournalOperation{
		SessionID: "sess-t18-redacted", ProfileID: "profile-t18-redacted", IdempotencyKey: "op-redacted",
		CorrelationID: "bad correlation with spaces", Reason: "Bearer super-secret-token",
	})
	if err != nil || !created {
		t.Fatalf("reserve redaction operation: created=%v err=%v", created, err)
	}
	var correlation, reason string
	if err := db.QueryRow(`SELECT correlation_id,reason FROM launch_operations WHERE session_id='sess-t18-redacted'`).Scan(&correlation, &reason); err != nil {
		t.Fatalf("read redaction fields: %v", err)
	}
	if correlation != "[redacted]" || strings.Contains(strings.ToLower(reason), "secret") || strings.Contains(strings.ToLower(reason), "token") {
		t.Fatalf("journal redaction failed: correlation=%q reason=%q", correlation, reason)
	}
}

func waitForQueueDepth(t *testing.T, m *Manager, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		got := len(m.queue)
		m.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("queue depth did not reach %d", want)
}
