// Package drift compares redacted diagnostic control states.
// It never compares or returns raw environment values.
package drift

import "sort"

const Version = "diagnostic-drift-v1"

type State string

const (
	StateStable  State = "STABLE"
	StateChanged State = "CHANGED"
	StateMissing State = "MISSING"
	StateAdded   State = "ADDED"
)

type Control struct {
	Key   string `json:"key"`
	State string `json:"state"`
}

type Finding struct {
	Key   string `json:"key"`
	State State  `json:"state"`
}

type Report struct {
	Version     string    `json:"version"`
	BaselineID  string    `json:"baseline_id"`
	CurrentID   string    `json:"current_id"`
	MaxChanges  int       `json:"max_changes"`
	WithinLimit bool      `json:"within_limit"`
	Findings    []Finding `json:"findings"`
}

// Compare compares only named state labels. The IDs are opaque caller-provided
// references and no raw control value is returned.
func Compare(baselineID, currentID string, baseline, current []Control, maxChanges int) Report {
	base := map[string]string{}
	now := map[string]string{}
	for _, c := range baseline {
		base[c.Key] = c.State
	}
	for _, c := range current {
		now[c.Key] = c.State
	}
	keys := map[string]bool{}
	for k := range base {
		keys[k] = true
	}
	for k := range now {
		keys[k] = true
	}
	ordered := make([]string, 0, len(keys))
	for k := range keys {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	findings := make([]Finding, 0)
	for _, key := range ordered {
		old, hadOld := base[key]
		fresh, hadFresh := now[key]
		switch {
		case !hadOld:
			findings = append(findings, Finding{Key: key, State: StateAdded})
		case !hadFresh:
			findings = append(findings, Finding{Key: key, State: StateMissing})
		case old != fresh:
			findings = append(findings, Finding{Key: key, State: StateChanged})
		}
	}
	if maxChanges < 0 {
		maxChanges = 0
	}
	return Report{Version: Version, BaselineID: baselineID, CurrentID: currentID, MaxChanges: maxChanges, WithinLimit: len(findings) <= maxChanges, Findings: findings}
}
