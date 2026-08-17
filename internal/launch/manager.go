// Package launch fournit la fiabilité Core de ForgeLocal (jalon T08) :
// queue bornée, limite globale de sessions, sérialisation par profil,
// timeouts, annulation via context.Context, cleanup idempotent,
// reprise après crash et journal d'audit sans secret.
//
// Ce package est une réimplémentation Go clean room. Il ne contient
// aucun code issu de l'archive Camoflox, n'importe aucune dépendance
// Node/Electron et ne lance aucun navigateur, runtime, proxy ou backup.
// Source de référence de contrat (lecture documentaire seulement) :
//
//	docs/T08_CONCURRENCY_SPEC.md (T08-SPEC-20260815-001)
//	docs/component-rights-register.json — décision reimplementer pour lib/concurrency.js
package launch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// profileID identifies a ForgeLocal profile. It is an opaque stable string.
type profileID string

// sessionID identifies a launched session (purely metadata, no process).
type sessionID string

// LaunchState describes the explicit lifecycle state of a session.
type LaunchState string

const (
	StateQueued      LaunchState = "queued"
	StateStarting    LaunchState = "starting"
	StateRunning     LaunchState = "running"
	StateStopping    LaunchState = "stopping"
	StateStopped     LaunchState = "stopped"
	StateError       LaunchState = "error"
	StateInterrupted LaunchState = "interrupted"
)

// RequestError values with redacted messages only.
var (
	ErrProfileAlreadyRunning = errors.New("session already active for profile")
	ErrGlobalLimitReached    = errors.New("global session limit reached")
	ErrQueueFull             = errors.New("launch queue is full")
	ErrWaitExpired           = errors.New("queue wait deadline expired")
	ErrCancelled             = errors.New("launch cancelled")
	ErrContextDone           = errors.New("context done while launching")
	ErrInvalidProfile        = errors.New("invalid or empty profile id")
)

// default limits used by NewManager when options are nil.
const (
	defaultGlobalLimit   = 4
	defaultMaxQueueDepth = 32
	defaultWaitDeadline  = 30 * time.Second
	defaultStartTimeout  = 45 * time.Second
)

// Options configures the LaunchManager limits. Zero values use defaults.
type Options struct {
	GlobalLimit   int
	MaxQueueDepth int
	WaitDeadline  time.Duration
	StartTimeout  time.Duration
	// Recoverer reopens a session store snapshot recorded before a crash.
	Recoverer Recoverer
	// Journal persists T18 queue lifecycle transitions atomically. When set,
	// no launcher is attached before its intent reached durable storage.
	Journal OperationJournal
}

// Recoverer replays a durable session snapshot (crash recovery input).
// It returns the sessions that must be resumed and their last known state.
type Recoverer interface {
	// Recover returns persisted sessions after a crash, in any order.
	Recover() []RecoveredSession
}

// RecoveredSession is the durable shape of a session persisted by the
// caller before any operation that could crash. It deliberately holds no
// executable resource: only the identifiers and the last known state.
type RecoveredSession struct {
	SessionID    sessionID
	ProfileID    profileID
	LastState    LaunchState
	LastRecorded time.Time
}

// Launcher is the resource attachment point. T08 forbids real process
// launches; implementations of this interface record intent and verify
// deterministic behaviour (e.g. a deterministic stub in tests).
type Launcher interface {
	// Attach attempts to bind resources to the session. It MUST return
	// quickly on ctx cancellation and never start a real browser/runtime.
	Attach(ctx context.Context, s Session) error
}

// Session is the metadata handled by the manager. Keep it redactable:
// it never carries secrets, ports, tokens or file contents.
type Session struct {
	ID        sessionID
	ProfileID profileID
	State     LaunchState
	CreatedAt time.Time
	StartedAt time.Time
	StoppedAt time.Time
	Err       string // redacted error reason only
	// CorrelID is a validated request correlation identifier. It is safe for
	// audit joins and never carries a secret, token, path or runtime detail.
	CorrelID string
}

// auditSink persists launch lifecycle events without secrets.
type auditSink interface {
	Record(event AuditEvent)
}

// AuditEvent is an append-only, redactable launch event.
type AuditEvent struct {
	Time     time.Time   `json:"time"`
	Event    string      `json:"event"`
	Session  sessionID   `json:"session"`
	Profile  profileID   `json:"profile"`
	From     LaunchState `json:"from"`
	To       LaunchState `json:"to"`
	Reason   string      `json:"reason,omitempty"` // redacted
	CorrelID string      `json:"correlation_id,omitempty"`
}

// Manager serializes launch intent per profile under a global bound.
type Manager struct {
	mu        sync.Mutex
	opt       Options
	running   map[sessionID]*Session
	byProfile map[profileID]*Session // at most one running|starting per profile
	queue     []queuedRequest
	attached  map[sessionID]chan struct{} // closed when attach completes (ok/err)
	sink      auditSink
	journal   OperationJournal
	// attach tracks every attach goroutine ever started so termination
	// can be proven within a bounded deadline (no leaked goroutine).
	attach sync.WaitGroup
	// ctx/cancel stop the whole manager: every attach goroutine receives
	// a merged context that closes when the manager stops, guaranteeing
	// bounded termination even when the caller context never cancels.
	ctx    context.Context
	cancel context.CancelFunc
	// promote signals waiter promotion to a single dedicated goroutine,
	// avoiding lock-contention fan-out when many attaches finish at once.
	promote chan struct{}

	// ---- stress debug counters (atomic, used by TestConcurrentStress)
	CAttachDone   atomic.Int64 // attach goroutine finished
	CNotifySent   atomic.Int64 // final snapshot pushed to waiter
	CPromoteSign  atomic.Int64 // promotion signal sent
	CWakeNextRun  atomic.Int64 // wakeNext executed
	CStopOneDone  atomic.Int64 // stopOne completed
	CReqResolved  atomic.Int64 // Request returned
	CReqQueuePath atomic.Int64 // Request took queue path
	CByProfileHit atomic.Int64 // Request refused: profile taken
	CLockTaken    atomic.Int64 // m.mu.Lock acquired
}

type queuedRequest struct {
	ctx      context.Context
	cancel   context.CancelFunc
	profile  profileID
	session  Session
	replied  bool
	result   chan Session
	launcher Launcher
}

// NewManager builds a bounded launch manager. Nil options use defaults.
func NewManager(opt *Options, sink auditSink) *Manager {
	if opt == nil {
		opt = &Options{}
	}
	g := opt.GlobalLimit
	if g <= 0 {
		g = defaultGlobalLimit
	}
	q := opt.MaxQueueDepth
	if q <= 0 {
		q = defaultMaxQueueDepth
	}
	w := opt.WaitDeadline
	if w <= 0 {
		w = defaultWaitDeadline
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		opt:       *opt,
		running:   map[sessionID]*Session{},
		byProfile: map[profileID]*Session{},
		attached:  map[sessionID]chan struct{}{},
		sink:      sink,
		journal:   opt.Journal,
		ctx:       ctx,
		cancel:    cancel,
	}
	m.opt.GlobalLimit = g
	m.opt.MaxQueueDepth = q
	m.opt.WaitDeadline = w
	if opt.StartTimeout <= 0 {
		m.opt.StartTimeout = defaultStartTimeout
	}
	// A single dedicated promotion goroutine consumes promotion signals
	// and runs wakeNext under the manager lock (AC-CAMO-02 fairness).
	m.promote = make(chan struct{}, 1)
	go func() {
		for range m.promote {
			m.wakeNext()
		}
	}()
	return m
}

// Status reports the current bounded load: running count, queue depth
// and global limit.
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Status{
		Running:  len(m.running),
		Queued:   len(m.queue),
		Limit:    m.opt.GlobalLimit,
		QueueMax: m.opt.MaxQueueDepth,
	}
}

// Status is the redactable snapshot of manager load.
type Status struct {
	Running  int `json:"running"`
	Queued   int `json:"queued"`
	Limit    int `json:"limit"`
	QueueMax int `json:"queue_max"`
}

// SessionForProfile returns the active (starting|running) session of a
// profile, or the zero value and false.
func (m *Manager) SessionForProfile(p profileID) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byProfile[p]
	if !ok {
		return Session{}, false
	}
	return *s, true
}

// snapshot records a durable-recovery record before an operation that
// could crash. The caller is responsible for persisting it.
func (m *Manager) snapshot(s *Session) RecoveredSession {
	return RecoveredSession{
		SessionID:    s.ID,
		ProfileID:    s.ProfileID,
		LastState:    s.State,
		LastRecorded: time.Now().UTC(),
	}
}
