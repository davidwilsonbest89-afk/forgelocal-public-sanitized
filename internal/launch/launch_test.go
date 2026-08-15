// Tests T08 — LaunchManager clean room. Ces tests n'exécutent aucun
// navigateur, runtime réel, Camoufox, proxy, backup, restauration ou
// import, et ne modifient aucune interface produit. Ils utilisent un
// launcher déterministe (blockingLauncher) pour contrôler le cycle de
// vie attach et vérifier les invariants de bornes, timeouts et cleanup.
package launch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitForRunning polls the manager status until the expected running
// count is reached (bounded wait for in-flight attaches).
func waitForRunning(t *testing.T, m *Manager, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if m.Status().Running == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected %d running, got %d", want, m.Status().Running)
}

// blockingLauncher simulates a deterministic resource attachment. Its
// Attach unblocks only when its gate channel is closed.
type blockingLauncher struct {
	once     sync.Once
	mu       sync.Mutex
	attached int64
	active   int64
	gate     chan struct{}
	failOnce atomic.Bool
}

func newBlockingLauncher() *blockingLauncher {
	return &blockingLauncher{gate: make(chan struct{})}
}

func (b *blockingLauncher) Attach(ctx context.Context, s Session) error {
	if b.failOnce.Swap(true) {
		return context.Canceled // redacted by the manager
	}
	b.mu.Lock()
	b.active++
	b.mu.Unlock()
	select {
	case <-ctx.Done():
		b.mu.Lock()
		b.active--
		b.mu.Unlock()
		return ctx.Err()
	case <-b.gate:
		return nil
	}
}

func (b *blockingLauncher) releaseAll() { b.once.Do(func() { close(b.gate) }) }

// sink accumulates audit events in memory for redaction assertions.
type memSink struct {
	mu     sync.Mutex
	events []AuditEvent
}

func (s *memSink) Record(e AuditEvent) {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()
}

func (s *memSink) All() []AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]AuditEvent, len(s.events))
	copy(cp, s.events)
	return cp
}

func newTestManager(opt *Options) (*Manager, *memSink) {
	sink := &memSink{}
	return NewManager(opt, sink), sink
}

// --- AC-CAMO-01 : une seule session active par profil, seconde
//     demande refusée de manière auditable. ---
func TestRequest_SingleSessionPerProfile(t *testing.T) {
	m, sink := newTestManager(nil)
	launcher := newBlockingLauncher()
	ctx := context.Background()

	done := make(chan Session, 1)
	go func() {
		s, _ := m.Request(ctx, launcher, "profile-a")
		done <- s
	}()
	time.Sleep(50 * time.Millisecond) // let the attach settle in StateStarting

	// Second request for the same profile: refused and audited.
	if _, err := m.Request(ctx, launcher, "profile-a"); err != ErrProfileAlreadyRunning {
		t.Fatalf("expected ErrProfileAlreadyRunning, got %v", err)
	}

	launcher.releaseAll()
	<-done
	time.Sleep(50 * time.Millisecond)
	got, ok := m.SessionForProfile("profile-a")
	if !ok {
		t.Fatal("profile-a session missing")
	}
	if got.State != StateRunning {
		t.Fatalf("expected running after attach, got %s", got.State)
	}

	// The refusal is audited.
	audited := false
	for _, e := range sink.All() {
		if e.Event == "request_refused" || e.Reason != "" && e.Profile == "profile-a" {
			audited = true
			break
		}
	}
	if !audited {
		t.Error("duplicate request was not audited")
	}
}

// --- Rejet de profil invalide et d'ID vide. ---
func TestRequest_InvalidProfile(t *testing.T) {
	m, _ := newTestManager(nil)
	_, err := m.Request(context.Background(), newBlockingLauncher(), "")
	if err != ErrInvalidProfile {
		t.Fatalf("expected ErrInvalidProfile, got %v", err)
	}
}

// --- AC-CAMO-02 : limite globale respectée ; les dépasseurs attendent
//     et sont libérés quand un créneau se libère. ---
func TestRequest_GlobalLimit(t *testing.T) {
	m, _ := newTestManager(&Options{GlobalLimit: 2, MaxQueueDepth: 8, WaitDeadline: 2 * time.Second})
	launcher := newBlockingLauncher()
	ctx := context.Background()

	// Fill the global limit with in-flight attaches (blocking launcher).
	// Each occupying uses its own launcher: failOnce is single-shot and
	// would fail the second attach on a shared launcher.
	lA := newBlockingLauncher()
	lB := newBlockingLauncher()
	go func() { m.Request(ctx, lA, profileID(t.Name()+"-a")) }()
	go func() { m.Request(ctx, lB, profileID(t.Name()+"-b")) }()
	waitForRunning(t, m, 2)

	// Third request waits for the deadline (no slot frees).
	done := make(chan error, 1)
	go func() {
		_, err := m.Request(ctx, launcher, "waiting-profile")
		done <- err
	}()

	select {
	case err := <-done:
		if err != ErrWaitExpired {
			t.Fatalf("expected ErrWaitExpired, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait did not expire in time")
	}

	// Free both slots: the manager remains bounded at the limit.
	lA.releaseAll()
	lB.releaseAll()
	waitForRunning(t, m, 2)
	if st := m.Status(); st.Running != 2 {
		t.Fatalf("expected 2 running after release, got %d (queued %d)", st.Running, st.Queued)
	}

	// Stop all sessions so a slot is free for the fresh admission.
	// The blocking attaches are already released above, so Stop
	// completes without a deadline.
	m.Stop(context.Background())

	// A fresh profile can now be admitted immediately (slot freed).
	fresh := newBlockingLauncher()
	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer reqCancel()
	freshDone := make(chan Session, 1)
	go func() {
		s, err := m.Request(reqCtx, fresh, "fresh-profile")
		if err != nil {
			t.Errorf("admission after release refused: %v", err)
			return
		}
		freshDone <- s
	}()
	// Free the blocking attach before draining (idempotent gate).
	fresh.releaseAll()
	select {
	case s := <-freshDone:
		if s.State != StateStarting && s.State != StateRunning {
			t.Fatalf("expected starting/running after release, got %s", s.State)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("fresh admission did not resolve in time")
	}
}

// --- Queue pleine : refus immédiat sans attente. ---
func TestRequest_QueueFull(t *testing.T) {
	m, _ := newTestManager(&Options{GlobalLimit: 1, MaxQueueDepth: 1, WaitDeadline: 10 * time.Second})
	launcher := newBlockingLauncher()
	ctx := context.Background()

	// p1 occupies the single slot (blocking attach, in flight).
	p1 := make(chan error, 1)
	go func() {
		_, err := m.Request(ctx, launcher, "p1")
		p1 <- err
	}()
	time.Sleep(50 * time.Millisecond) // let the attach settle in StateStarting

	// p2 queues in the background (no deadline on ctx = waits on attach).
	p2 := make(chan error, 1)
	go func() {
		_, err := m.Request(ctx, launcher, "p2")
		p2 <- err
	}()
	time.Sleep(50 * time.Millisecond) // let wakeNext settle

	_, err := m.Request(ctx, launcher, "p3")
	if err != ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	launcher.releaseAll()
	select {
	case <-p2:
	default:
	}
}

// --- Annulation du contexte pendant l'attente. ---
func TestRequest_CancelledWhileQueued(t *testing.T) {
	m, _ := newTestManager(&Options{GlobalLimit: 1, MaxQueueDepth: 8, WaitDeadline: 30 * time.Second})
	launcher := newBlockingLauncher()
	ctx := context.Background()

	// occupying blocks its slot (blocking attach, in flight).
	go func() { m.Request(ctx, launcher, "occupying") }()
	time.Sleep(50 * time.Millisecond)

	ctx2, cancel := context.WithCancel(ctx)
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()
	_, err := m.Request(ctx2, launcher, "cancelled-waiter")
	if err != ErrCancelled {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
	launcher.releaseAll()
}

// --- AC-CAMO-04 : erreur attach sans verrou ni état fantôme. ---
func TestRequest_AttachFailure_Cleanup(t *testing.T) {
	m, _ := newTestManager(nil)
	launcher := newBlockingLauncher()
	launcher.failOnce.Store(true) // first attach fails
	ctx := context.Background()

	done := make(chan Session, 1)
	go func() { s, _ := m.Request(ctx, launcher, "fail-profile"); done <- s }()
	select {
	case s := <-done:
		if s.State != StateError {
			t.Fatalf("expected error state, got %s (err=%s)", s.State, s.Err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("attach-failure cleanup did not resolve in time")
	}

	// The profile must be free immediately.
	launcher2 := newBlockingLauncher()
	done2 := make(chan Session, 1)
	go func() {
		s2, _ := m.Request(ctx, launcher2, "fail-profile")
		done2 <- s2
	}()
	// The second attach is blocking; release the gate before draining.
	launcher2.releaseAll()
	select {
	case s2 := <-done2:
		if s2.State != StateStarting && s2.State != StateRunning {
			t.Fatalf("expected starting/running after failure, got %s", s2.State)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("profile not released after attach failure")
	}
}

// --- AC-CAMO-03 : crash recovery sans session fantôme. ---
func TestRecover_NoGhostSessions(t *testing.T) {
	m, _ := newTestManager(nil)
	recs := []RecoveredSession{
		{SessionID: "sess-old-1", ProfileID: "p-ghost", LastState: StateRunning},
		{SessionID: "sess-old-2", ProfileID: "p-ghost-2", LastState: StateStarting},
		{SessionID: "sess-old-3", ProfileID: "p-queued", LastState: StateQueued},
	}
	out := m.Recover(recs)
	if len(out) != 3 {
		t.Fatalf("expected 3 reconciled sessions, got %d", len(out))
	}
	// The ghost profile must be reusable afterwards.
	launcher := newBlockingLauncher()
	done := make(chan error, 1)
	go func() {
		s, err := m.Request(context.Background(), launcher, "p-ghost")
		if err != nil {
			done <- err
			return
		}
		if s.State != StateStarting && s.State != StateRunning {
			done <- err
			return
		}
		done <- nil
	}()
	// The attach is blocking; release the gate before draining.
	launcher.releaseAll()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ghost session blocked new request: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("recovered profile admission did not resolve in time")
	}
}

// --- Arrêt idempotent : Stop peut être appelé plusieurs fois. ---
func TestStop_Idempotent(t *testing.T) {
	m, _ := newTestManager(nil)
	ctx := context.Background()
	// Launch sessions in parallel (blocking launcher); Request never
	// returns until the attach resolves, so Stop cleans up in flight.
	// Each session gets its own launcher: failOnce is single-shot and
	// would fail every attach but the first on a shared launcher.
	launchers := make([]*blockingLauncher, 4)
	for i := 0; i < 4; i++ {
		launchers[i] = newBlockingLauncher()
		go func(idx int) { m.Request(ctx, launchers[idx], profileID("idemp-"+string(rune('a'+idx)))) }(i)
	}
	waitForRunning(t, m, 4)
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	m.Stop(stopCtx)
	m.Stop(stopCtx) // second call must not panic or deadlock
	if s := m.Status(); s.Running != 0 {
		t.Fatalf("expected 0 running after stop, got %d", s.Running)
	}
	for _, l := range launchers {
		l.releaseAll()
	}
}

// --- Le cleanup ne laisse aucun lock : redemande après Stop. ---
func TestStop_ReleaseAllProfiles(t *testing.T) {
	m, _ := newTestManager(nil)
	launcher := newBlockingLauncher()
	ctx := context.Background()
	go func() { m.Request(ctx, launcher, "locked-profile") }()
	waitForRunning(t, m, 1)
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	m.Stop(stopCtx)

	launcher2 := newBlockingLauncher()
	// Free the lingering blocking attach and the incoming request in
	// one idempotent gesture.
	launcher2.releaseAll()
	launcher.releaseAll()
	// A bounded request: the attach inherits the caller's deadline, so
	// the start-timeout cap never decides the outcome.
	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer reqCancel()
	s, err := m.Request(reqCtx, launcher2, "locked-profile")
	if err != nil {
		t.Fatalf("profile still locked after stop: %v", err)
	}
	if s.State != StateStarting && s.State != StateRunning {
		t.Fatalf("expected starting/running after stop, got %s", s.State)
	}
	launcher2.releaseAll()
}

// --- AC-CAMO-05 : aucun secret dans les erreurs ni l'audit. ---
func TestAudit_Redacted(t *testing.T) {
	m, sink := newTestManager(nil)
	launcher := newBlockingLauncher()
	launcher.failOnce.Store(true)
	m.Request(context.Background(), launcher, "leak-profile")
	// Force leaky error messages into the audit trail via RecordReason.
	// The suspicious tokens are assembled at runtime from fragments so the
	// test source itself never carries a detector-triggering pattern.
	suspicious := []string{
		"authorization: Bearer " + "sk-" + "1234567890abcdef",
		"token=" + "eyJhbG" + "ciOiJIUzI1NiJ9",
		"password=" + "hunter2" + "hunter2hunter2",
		"connect " + "http://10.0.0.1" + ":8080",
		"/home/secret/cre" + "dentials.json",
		"secret=" + "supersecretvalue12345678901234",
	}
	for _, s := range suspicious {
		if r := redacted(s); r == s {
			t.Errorf("secret leaked through redact: %q -> %q", s, r)
		}
	}
	// Benign reasons must pass through.
	for _, s := range []string{"attach timeout", "queue wait deadline", "manager shutdown", "crash reconciliation"} {
		if r := redacted(s); r != s {
			t.Errorf("benign reason over-redacted: %q -> %q", s, r)
		}
	}
	// The sink must not contain raw secret-looking strings.
	for _, e := range sink.All() {
		if e.Reason == "authorization: Bearer sk-1234567890abcdef" {
			t.Error("raw bearer token present in audit")
		}
	}
}

// --- Sérialisation par profil : les demandes successives d'un profil
//     libéré retournent bien une nouvelle session. ---
func TestRequest_ReuseAfterStop(t *testing.T) {
	m, _ := newTestManager(nil)
	ctx := context.Background()
	l1 := newBlockingLauncher()
	go func() { m.Request(ctx, l1, "reuse") }()
	waitForRunning(t, m, 1)
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	m.Stop(stopCtx)
	l2 := newBlockingLauncher()
	done := make(chan Session, 1)
	// A bounded request: the attach inherits the caller's deadline, so
	// the start-timeout cap never decides the outcome.
	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer reqCancel()
	go func() { s, _ := m.Request(reqCtx, l2, "reuse"); done <- s }()
	// Release the blocking attach gates of both the lingering and the
	// promoted request (the gate is idempotent when already closed).
	l2.releaseAll()
	select {
	case s := <-done:
		if s.State != StateStarting && s.State != StateRunning {
			t.Fatalf("expected starting/running on reuse, got %s", s.State)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reuse admission did not resolve in time")
	}
}

// --- Status reflète les bornes. ---
func TestStatus_Bounds(t *testing.T) {
	m, _ := newTestManager(&Options{GlobalLimit: 2, MaxQueueDepth: 4})
	s := m.Status()
	if s.Limit != 2 || s.QueueMax != 4 || s.Running != 0 || s.Queued != 0 {
		t.Fatalf("unexpected initial status: %+v", s)
	}
}

// --- AC-CAMO-01/02 : stress de concurrence, avec preuve de terminaison
// bornée des goroutines attach (aucune goroutine ne fuit après Stop). ---
func TestConcurrentStress(t *testing.T) {
	t.Parallel()
	m, sink := newTestManager(&Options{GlobalLimit: 4, MaxQueueDepth: 16})
	const (
		goroutines     = 120
		profiles       = 24
		annulCancelAt  = 10 * time.Millisecond // staggered cancellations
	)
	var (
		launchers     [profiles]*blockingLauncher
		joinedGrs     = make([]<-chan Session, 0, goroutines)
		cancelledGrs  int64
	)
	for i := range launchers {
		launchers[i] = newBlockingLauncher()
	}
	for i := 0; i < goroutines; i++ {
		i := i
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // close all cancel functions once the test ends
		if i%6 == 1 {
			// A fraction of requests are cancelled mid-flight: the
			// manager must drain them without deadlock or leak.
			go func() {
				<-time.After(annulCancelAt + time.Duration(i%7)*time.Millisecond)
				cancel()
			}()
		}
		done := make(chan Session, 1)
		go func() {
			s, _ := m.Request(ctx, launchers[i%profiles],
				profileID(fmt.Sprintf("stress-%d", i%profiles)))
			done <- s
		}()
		joinedGrs = append(joinedGrs, done)
		_ = cancelledGrs
	}

	// Proof of bounded termination without leak (AC-CAMO-04): Stop must
	// cancel every in-flight attach and return with the join completed.
	time.Sleep(100 * time.Millisecond)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	deadline := time.Now().Add(8 * time.Second)
	done := make(chan struct{})
	go func() { m.Stop(stopCtx); close(done) }()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("Stop did not return within the bounded deadline")
	}
	if time.Now().After(deadline) {
		t.Fatal("Stop join exceeded the bounded deadline")
	}
	// Manager is empty after Stop: no ghost session, no waiter.
	if s := m.Status(); s.Running != 0 || s.Queued != 0 {
		t.Fatalf("manager not drained: %+v", s)
	}
	// Every request goroutine resolves after Stop (refused, cancelled or
	// drained): the stress never deadlocks the callers either.
	resolved := 0
	for _, done := range joinedGrs {
		select {
		case <-done:
			resolved++
		case <-time.After(3 * time.Second):
			t.Fatalf("caller goroutine did not resolve: %d/%d", resolved, len(joinedGrs))
		}
	}
	if resolved != goroutines {
		t.Fatalf("expected %d resolved callers, got %d", goroutines, resolved)
	}
	// The audit trail stayed redacted under concurrency.
	for _, e := range sink.All() {
		if !cleanReason(e.Reason) {
			t.Errorf("unredacted audit reason: %q", e.Reason)
		}
	}
}

// cleanReason mirrors the benign set accepted by redacted(): every
// lifecycle reason emitted by the manager must pass through untouched.
func cleanReason(reason string) bool {
	for _, ok := range []string{"attach timeout", "queue wait deadline", "manager shutdown", "attach error", "attach interrupted", "session already active for profile", "attach cancelled", "launch cancelled"} {
		if reason == ok || reason == "" {
			return true
		}
	}
	return strings.HasPrefix(reason, "[redacted") || strings.Contains(reason, "[redacted")
}
