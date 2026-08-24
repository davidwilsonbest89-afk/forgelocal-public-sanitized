package sessiontrack

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

type State string

const (
	Queued   State = "QUEUED"
	Starting State = "STARTING"
	Running  State = "RUNNING"
	Stopping State = "STOPPING"
	Stopped  State = "STOPPED"
	Failed   State = "FAILED"
)

type View struct {
	SessionKey string
	State      State
	ReasonCode string
}

type Tracker struct {
	mu     sync.RWMutex
	states map[string]View
}

func New() *Tracker { return &Tracker{states: make(map[string]View)} }

func validKey(k string) bool {
	return k != "" && len(k) <= 64 && !strings.ContainsAny(k, "\r\n/\\")
}

func (t *Tracker) Start(key string) error {
	if !validKey(key) {
		return errors.New("invalid session key")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.states[key]; ok {
		return errors.New("session already exists")
	}
	t.states[key] = View{SessionKey: key, State: Queued}
	return nil
}

func (t *Tracker) Transition(key string, state State, reason string) error {
	if !validKey(key) || reason == "" || strings.ContainsAny(reason, "\r\n") {
		return errors.New("invalid redacted transition")
	}
	switch state {
	case Starting, Running, Stopping, Stopped, Failed:
	default:
		return errors.New("unsupported state")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	v, ok := t.states[key]
	if !ok {
		return errors.New("unknown session")
	}
	v.State, v.ReasonCode = state, reason
	t.states[key] = v
	return nil
}

func (t *Tracker) Snapshot() []View {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]View, 0, len(t.states))
	for _, v := range t.states {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SessionKey < out[j].SessionKey })
	return out
}
