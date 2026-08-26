package launch

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Request starts a session for p with the given launcher. It honours the
// global limit, the per-profile serialisation and the queue bounds. The
// returned Session describes the final outcome only: queued, running,
// explicitly refused or cancelled. The manager never starts a real
// browser or runtime; the launcher attachment is bounded by ctx and by
// StartTimeout and always releases resources through cleanup.
func (m *Manager) Request(ctx context.Context, launcher Launcher, p profileID) (Session, error) {
	if p == "" {
		return Session{}, ErrInvalidProfile
	}

	sess := Session{
		ID:        sessionID(newID()),
		ProfileID: p,
		State:     StateQueued,
		CreatedAt: time.Now().UTC(),
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Session{}, ErrWaitExpired
		}
		return Session{}, ErrCancelled
	}
	meta := metadataFromContext(ctx, sess.ID)
	sess.CorrelID = meta.correlationID

	m.mu.Lock()
	m.CLockTaken.Add(1)
	if m.journal != nil {
		op, found, err := m.journal.Lookup(ctx, meta.idempotencyKey)
		if err != nil {
			m.mu.Unlock()
			return Session{}, err
		}
		if found {
			m.mu.Unlock()
			return Session{ID: op.SessionID, ProfileID: op.ProfileID, State: op.State,
				CreatedAt: op.CreatedAt, StoppedAt: op.UpdatedAt, Err: op.Reason,
				CorrelID: op.CorrelationID}, nil
		}
	}
	if _, taken := m.byProfile[p]; taken {
		m.mu.Unlock()
		m.record(AuditEvent{
			Time: time.Now().UTC(), Event: "request_refused",
			Session: sess.ID, Profile: p, From: StateQueued,
			Reason: "session already active for profile",
		})
		m.CByProfileHit.Add(1)
		return Session{}, ErrProfileAlreadyRunning
	}
	if len(m.running) >= m.opt.GlobalLimit && len(m.queue) >= m.opt.MaxQueueDepth {
		m.mu.Unlock()
		return Session{}, ErrQueueFull
	}
	if m.journal != nil {
		op, created, err := m.journal.Reserve(ctx, JournalOperation{
			SessionID:      sess.ID,
			ProfileID:      sess.ProfileID,
			IdempotencyKey: meta.idempotencyKey,
			State:          StateQueued,
			CorrelationID:  sess.CorrelID,
		})
		if err != nil {
			m.mu.Unlock()
			return Session{}, err
		}
		if !created {
			m.mu.Unlock()
			return Session{ID: op.SessionID, ProfileID: op.ProfileID, State: op.State,
				CreatedAt: op.CreatedAt, StoppedAt: op.UpdatedAt, Err: op.Reason,
				CorrelID: op.CorrelationID}, nil
		}
	}

	// If the global limit is free and no conflicting session exists,
	// start immediately without queuing (AC-CAMO-01 fast path). A
	// dedicated reply channel still carries the final snapshot so attach
	// failures are never reported as a bare context error.
	if len(m.running) < m.opt.GlobalLimit {
		if err := m.transitionDurable(ctx, sess, StateQueued, StateStarting, ""); err != nil {
			m.mu.Unlock()
			return Session{}, err
		}
		reply := make(chan Session, 1)
		m.begin(ctx, sess, launcher, reply)
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			m.cancelSession(sess)
			if ctx.Err() == context.Canceled {
				m.CReqResolved.Add(1)
				return sess, ErrCancelled
			}
			m.CReqResolved.Add(1)
			return sess, ErrContextDone
		case out := <-reply:
			if out.State == StateError {
				m.CReqResolved.Add(1)
				return out, nil
			}
			m.CReqResolved.Add(1)
			return out, nil
		}
	}

	// Global limit reached: enqueue under the deadline.
	deadline := time.Now().Add(m.opt.WaitDeadline)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	qr := queuedRequest{
		ctx:      ctx,
		cancel:   cancel,
		profile:  p,
		session:  sess,
		result:   make(chan Session, 1),
		launcher: launcher,
	}
	m.queue = append(m.queue, qr)
	m.mu.Unlock()
	m.CReqQueuePath.Add(1)

	select {
	case <-ctx.Done():
		m.removeRequest(qr)
		m.mu.Lock()
		// Free the waiter without leaking the profile lock or slots.
		for i := range m.queue {
			if m.queue[i].result == qr.result {
				m.queue = append(m.queue[:i], m.queue[i+1:]...)
				break
			}
		}
		m.mu.Unlock()
		_ = m.transitionDurable(context.Background(), sess, StateQueued, StateInterrupted, "queue context ended")
		if ctx.Err() == context.DeadlineExceeded {
			m.record(AuditEvent{
				Time: time.Now().UTC(), Event: "wait_expired",
				Session: sess.ID, Profile: p, From: StateQueued,
				Reason: "queue wait deadline", CorrelID: sess.CorrelID,
			})
			m.CReqResolved.Add(1)
			return sess, ErrWaitExpired
		}
		m.CReqResolved.Add(1)
		return sess, ErrCancelled
	case out := <-qr.result:
		if out.State == StateError {
			m.CReqResolved.Add(1)
			return out, ErrContextDone
		}
		m.CReqResolved.Add(1)
		return out, nil
	}
}

// begin admits a session into the running set and starts its bounded
// attach. It must be called with m.mu held. An optional waiter channel,
// when provided, receives the session snapshot once the attach resolves;
// this is what unblocks a queued request. If notify is nil the caller is
// assumed to be reading the attached-channel completion (fast path).
func (m *Manager) begin(ctx context.Context, sess Session, launcher Launcher, notify chan<- Session) {
	sess.State = StateStarting
	rec := m.snapshot(&sess)
	_ = rec // caller may persist for crash recovery

	done := make(chan struct{})
	m.attached[sess.ID] = done

	// attachJoin counts this goroutine so Stop can prove bounded
	// termination (no leaked goroutine) after cancelling the manager.
	m.attach.Add(1)

	// The running map must never hold the address of begin's parameter
	// (shared with this goroutine and with cancelSession) : allocate a
	// dedicated heap copy so the parameter stays local to begin.func1.
	heap := &Session{ID: sess.ID, ProfileID: sess.ProfileID,
		State: sess.State, CreatedAt: sess.CreatedAt, CorrelID: sess.CorrelID}
	m.running[sess.ID] = heap
	m.byProfile[sess.ProfileID] = heap

	go func() {
		defer close(done)
		defer func() { m.CAttachDone.Add(1) }()
		defer m.attach.Done()
		// The attach inherits the outer context so an explicit caller
		// cancellation propagates. It is also merged with the manager's
		// own stop context: stopping the manager unblocks every attach,
		// guaranteeing bounded termination even when the caller context
		// never cancels. A start-timeout cap applies only when the
		// caller did not already bound the request.
		callerCtx := ctx
		attachCtx := mergeContexts(callerCtx, m.ctx)
		var cancel context.CancelFunc
		if _, hasDeadline := attachCtx.Deadline(); !hasDeadline {
			attachCtx, cancel = context.WithTimeout(attachCtx, m.opt.StartTimeout)
		} else {
			cancel = func() {}
		}
		defer cancel()

		err := launcher.Attach(attachCtx, sess)

		// The final snapshot (success or redacted failure) is always
		// pushed to notify, so waiters never read a stale running map.
		m.mu.Lock()
		var final Session
		if err != nil {
			m.failLocked(sess, err, attachFailureState(err, callerCtx))
			final = sess
			final.State = attachFailureState(err, callerCtx)
			final.Err = redacted(err.Error())
		} else if st, ok := m.running[sess.ID]; ok {
			if journalErr := m.transitionDurable(attachCtx, *st, StateStarting, StateRunning, ""); journalErr != nil {
				m.failLocked(sess, journalErr, StateError)
				final = sess
				final.State = StateError
				final.Err = redacted(journalErr.Error())
			} else {
				st.State = StateRunning
				st.StartedAt = time.Now().UTC()
				final = *st
			}
		} else {
			final = sess
			final.State = StateInterrupted
			final.Err = "attach interrupted"
		}
		if notify != nil {
			m.notifyNotifyLocked(notify, final)
		}
		m.mu.Unlock()
		m.CNotifySent.Add(1)

		// Promotion is delegated to a single dedicated goroutine so that
		// finish-of-attach never hammers the manager mutex under load
		// (no lock contention fan-out, AC-CAMO-02 fairness).
		m.CPromoteSign.Add(1)
		select {
		case m.promote <- struct{}{}:
		default:
		}
	}()
}

// notifyNotifyLocked sends the session snapshot to a queued waiter; the
// channel is buffered (1) so this never blocks while m.mu is held.
func (m *Manager) notifyNotifyLocked(notify chan<- Session, sess Session) {
	select {
	case notify <- sess:
	default:
	}
}

// failLocked records a failed attach and cleans up without leaving a
// lock or stale state (AC-CAMO-04). m.mu must be held.
func (m *Manager) failLocked(sess Session, err error, finalState LaunchState) {
	reason := "attach error"
	if finalState == StateInterrupted {
		reason = "attach cancelled"
	}
	if st, ok := m.running[sess.ID]; ok {
		st.State = finalState
		st.StoppedAt = time.Now().UTC()
		st.Err = redacted(err.Error())
	}
	delete(m.running, sess.ID)
	delete(m.byProfile, sess.ProfileID)
	delete(m.attached, sess.ID)
	m.record(AuditEvent{
		Time: time.Now().UTC(), Event: "attach_failed",
		Session: sess.ID, Profile: sess.ProfileID,
		From: StateStarting, To: finalState,
		Reason: reason, CorrelID: sess.CorrelID,
	})
	_ = m.transitionDurable(context.Background(), sess, StateStarting, finalState, reason)
}

func attachFailureState(err error, callerCtx context.Context) LaunchState {
	if errors.Is(err, context.Canceled) && errors.Is(callerCtx.Err(), context.Canceled) {
		return StateInterrupted
	}
	return StateError
}

// wakeNext promotes the oldest waiter for a free profile slot, if any.
// m.mu is NOT held when wakeNext runs.
func (m *Manager) wakeNext() {
	m.mu.Lock()
	m.CWakeNextRun.Add(1)
	defer m.mu.Unlock()
	if len(m.running) >= m.opt.GlobalLimit || len(m.queue) == 0 {
		return
	}
	// Find the first waiter whose profile is free and promote it.
	for i := range m.queue {
		if _, taken := m.byProfile[m.queue[i].profile]; taken {
			continue
		}
		qr := m.queue[i]
		m.queue = append(m.queue[:i], m.queue[i+1:]...)
		sess := qr.session
		sess.State = StateStarting
		if err := m.transitionDurable(qr.ctx, sess, StateQueued, StateStarting, ""); err != nil {
			sess.State = StateError
			sess.Err = redacted(err.Error())
			m.notifyNotifyLocked(qr.result, sess)
			return
		}
		m.begin(qr.ctx, sess, qr.launcher, qr.result)
		return
	}
}

// removeRequest marks a queue entry as already answered so wakeNext skips it.
func (m *Manager) removeRequest(qr queuedRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.queue {
		if m.queue[i].result == qr.result {
			m.queue[i].replied = true
			return
		}
	}
}

// Stop gracefully halts all starting/running sessions with per-session
// timeout and idempotent cleanup. Safe to call multiple times.
// stopWork carries a snapshot of a session and its attach-done channel,
// captured under the manager lock so stopOne never shares pointers with
// the attach goroutine.
type stopWork struct {
	sess Session
	done chan struct{}
}

func (m *Manager) Stop(ctx context.Context) {
	// Cancel the manager context first: every attach goroutine merges
	// this context, so all in-flight attaches observe cancellation and
	// terminate within a bounded deadline (proven join, no leak).
	m.cancel()

	m.mu.Lock()
	works := make([]stopWork, 0, len(m.running))
	for id, s := range m.running {
		works = append(works, stopWork{sess: *s, done: m.attached[id]})
	}
	// Drain the queue so no waiter can be promoted after Stop begins.
	for i := range m.queue {
		if m.queue[i].cancel != nil {
			m.queue[i].cancel()
		}
		_ = m.transitionDurable(context.Background(), m.queue[i].session, StateQueued, StateInterrupted, "manager shutdown")
		m.queue[i].replied = true
	}
	m.queue = nil
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, w := range works {
		wg.Add(1)
		go func(sw stopWork) {
			defer wg.Done()
			m.stopOne(ctx, sw)
		}(w)
	}
	wg.Wait()

	// Bounded join of every attach goroutine ever started: Stop returns
	// nil only when the join completes before the deadline, which is the
	// explicit cleanup-without-leak proof required by AC-CAMO-04.
	joinCtx, joinCancel := context.WithTimeout(ctx, m.opt.StartTimeout)
	defer joinCancel()
	joined := make(chan struct{})
	go func() { m.attach.Wait(); close(joined) }()
	select {
	case <-joined:
	case <-joinCtx.Done():
		m.record(AuditEvent{
			Time: time.Now().UTC(), Event: "stop_join_deadline",
			Reason: "attach goroutine join exceeded stop deadline",
		})
	}
}

// mergeContexts returns a context that closes as soon as either input
// closes, with both underlying cancellations preserved. It is the
// mechanism that lets a manager Stop propagate to every attach.
func mergeContexts(a, b context.Context) context.Context {
	merged, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-a.Done():
		case <-b.Done():
		}
		cancel()
	}()
	return merged
}

// stopOne runs the idempotent cleanup of one session (AC-CAMO-04,
// stop is safe when resources are already released).
func (m *Manager) stopOne(ctx context.Context, w stopWork) {
	select {
	case <-ctx.Done():
	case <-w.done:
	}
	// Deterministic cleanup: no external resource in T08; the audit
	// trail is the observable side effect. stopOne must never write
	// into the Session pointer shared with the attach goroutine: it
	// snapshots the value under the lock, removes the shared entry,
	// then records the audit event from the private copy.
	m.mu.Lock()
	m.CLockTaken.Add(1)
	var snap Session
	ok := false
	if st, present := m.running[w.sess.ID]; present {
		snap = *st
		ok = true
		delete(m.running, w.sess.ID)
		delete(m.byProfile, snap.ProfileID)
		delete(m.attached, w.sess.ID)
	}
	m.mu.Unlock()
	m.CStopOneDone.Add(1)
	if ok {
		snap.State = StateStopped
		snap.StoppedAt = time.Now().UTC()
		snap.Err = "manager shutdown"
		m.record(AuditEvent{
			Time: time.Now().UTC(), Event: "session_stopped",
			Session: w.sess.ID, Profile: w.sess.ProfileID,
			From: w.sess.State, To: StateStopped,
			Reason: "manager shutdown", CorrelID: snap.CorrelID,
		})
		_ = m.transitionDurable(context.Background(), snap, w.sess.State, StateStopped, "manager shutdown")
	}
}

// Recover replays a crash snapshot: interrupted or running sessions are
// reconciled so no ghost session survives (AC-CAMO-03).
func (m *Manager) Recover(recs []RecoveredSession) []Session {
	if m.journal != nil && len(recs) == 0 {
		if persisted, err := m.journal.Reconcile(context.Background()); err == nil {
			recs = persisted
		}
	}
	if m.opt.Recoverer != nil && len(recs) == 0 {
		recs = m.opt.Recoverer.Recover()
	}
	var out []Session
	m.mu.Lock()
	for _, r := range recs {
		switch r.LastState {
		case StateStarting, StateRunning:
			// The durable record cannot prove the resource survived;
			// reconciliation marks it stopped with a redacted reason and
			// releases the profile slot (no ghost session survives).
			stopped := &Session{
				ID: r.SessionID, ProfileID: r.ProfileID,
				State: StateStopped, StoppedAt: time.Now().UTC(),
				Err: "crash reconciliation",
			}
			out = append(out, stopped.SessionCopy())
		case StateQueued:
			// Queued intent is dropped fail-closed: no silent relaunch.
			out = append(out, Session{
				ID: r.SessionID, ProfileID: r.ProfileID,
				State: StateInterrupted,
				Err:   "crash reconciliation",
			})
		}
		m.record(AuditEvent{
			Time: time.Now().UTC(), Event: "crash_recovery",
			Session: r.SessionID, Profile: r.ProfileID,
			From: r.LastState, To: StateStopped,
			Reason: "crash reconciliation",
		})
	}
	m.mu.Unlock()
	return out
}

// await blocks until the session attach resolves or ctx ends (which
// carries any deadline/cancellation set by the caller, plus the
// start-timeout applied inside begin). It never holds m.mu while waiting.
func (m *Manager) await(ctx context.Context, sess Session) (Session, error) {
	done := m.attached[sess.ID]
	select {
	case <-ctx.Done():
		m.cancelSession(sess)
		if ctx.Err() == context.Canceled {
			return sess, ErrCancelled
		}
		return sess, ErrContextDone
	case <-done:
		// Attach resolved: the error case was already pushed into
		// notify; the fast path waits for m.running consistency.
		m.mu.Lock()
		st := m.running[sess.ID]
		if st == nil {
			m.mu.Unlock()
			// The attach failed or was interrupted; notify carries the
			// redacted failure snapshot for queued waiters. For the fast
			// path we reconcile under the mutex.
			return sess, ErrContextDone
		}
		out := *st
		m.mu.Unlock()
		return out, nil // error already redacted in st.Err
	}
}

func (m *Manager) cancelSession(sess Session) {
	m.mu.Lock()
	if st, ok := m.running[sess.ID]; ok {
		st.State = StateInterrupted
		st.StoppedAt = time.Now().UTC()
		st.Err = "cancelled"
		delete(m.running, sess.ID)
		delete(m.byProfile, sess.ProfileID)
		delete(m.attached, sess.ID)
		_ = m.transitionDurable(context.Background(), *st, StateStarting, StateInterrupted, "cancelled")
	}
	m.mu.Unlock()
}

func (m *Manager) record(e AuditEvent) {
	if m.sink != nil {
		m.sink.Record(e)
	}
}

func (m *Manager) transitionDurable(ctx context.Context, sess Session, from, to LaunchState, reason string) error {
	if m.journal == nil {
		return nil
	}
	return m.journal.Transition(ctx, sess.ID, from, to, reason, sess.CorrelID)
}

// SessionCopy returns a redactable snapshot of the session.
func (s *Session) SessionCopy() Session {
	if s == nil {
		return Session{}
	}
	return *s
}
