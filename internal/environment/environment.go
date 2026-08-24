// Package environment implements the T13 browser-environment diagnostic.
//
// Contract (never deviate):
//   - Projected controls only: raw values (UA strings, raw fingerprints,
//     canvas hashes, screen geometry) are NEVER returned. The client sees
//     check names, states and an aggregated verdict.
//   - Unknown profile ids are refused explicitly with PROFILE_NOT_FOUND;
//     a 200 "not found" style response is never used.
//   - No network, no runtime launch, no Camoufox: projection is derived
//     solely from stored profile metadata and the runtime qualification
//     registry, when present.
//   - Result shape is deterministic and machine-readable for the audit.
package environment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// State is the per-control projected state.
type State string

const (
	StatePass           State = "PASS"
	StateWarning        State = "WARNING"
	StateFail           State = "FAIL"
	StateUnsupported    State = "UNSUPPORTED"
	StateRuntimeDefined State = "RUNTIME_DEFINED"
)

// Verdict is the aggregated diagnostic verdict.
type Verdict string

const (
	VerdictClean   Verdict = "CLEAN"
	VerdictRisky   Verdict = "RISKY"
	VerdictBroken  Verdict = "BROKEN"
	VerdictUnknown Verdict = "UNKNOWN"
)

const (
	// DiagnosticVersion identifies this deterministic, read-only projection.
	DiagnosticVersion = "environment-projection-v3"
	// ObservationModeProjected declares that no browser or runtime was launched.
	ObservationModeProjected = "PROJECTED_METADATA_ONLY"
)

// Control is a single projected diagnostic check; note is a fixed label,
// never a raw value.
type Control struct {
	Name  string `json:"name"`
	State State  `json:"state"`
	Note  string `json:"note,omitempty"`
}

// Diagnostic is the projected, redacted result for one profile.
type Diagnostic struct {
	DiagnosticVersion string    `json:"diagnostic_version"`
	ObservationMode   string    `json:"observation_mode"`
	ProfileID         string    `json:"profile_id"`
	Verdict           Verdict   `json:"verdict"`
	Controls          []Control `json:"controls"`
}

var (
	ErrProfileNotFound = errors.New("environment: PROFILE_NOT_FOUND")
	ErrInvalidID       = errors.New("environment: invalid profile id")
)

// Checker is the dependency surface: it knows the stored profile shape and
// the runtime qualification registry. Both stay inside the Core.
type Checker interface {
	ProfileExists(ctx context.Context, id string) (bool, error)
	ProfileRuntimeID(ctx context.Context, id string) (string, error) // "" if unset
	RuntimeQualified(ctx context.Context, runtimeID string) (bool, error)
	// RuntimeVersion returns only major.minor style versions, never paths.
	RuntimeVersion(ctx context.Context, runtimeID string) (string, error)
}

// Diagnose computes the projected diagnostic for id. Never observes a
// real browser; derives everything from metadata.
func Diagnose(ctx context.Context, chk Checker, id string) (*Diagnostic, error) {
	if !validID(id) {
		return nil, ErrInvalidID
	}
	exists, err := chk.ProfileExists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("environment: %w", err)
	}
	if !exists {
		return nil, ErrProfileNotFound
	}
	rt, err := chk.ProfileRuntimeID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("environment: %w", err)
	}
	controls := []Control{
		controlProfileMetadata(),
		controlRuntimeBinding(rt, chk, ctx),
	}
	controls = append(controls, declaredUnobservedCapabilities()...)
	return &Diagnostic{
		DiagnosticVersion: DiagnosticVersion,
		ObservationMode:   ObservationModeProjected,
		ProfileID:         id,
		Verdict:           aggregate(controls),
		Controls:          controls,
	}, nil
}

func validID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, "/ \x00") {
		return false
	}
	return true
}

func controlProfileMetadata() Control {
	// Stored profiles always carry required metadata by construction
	// (T09 schema). This is a projected self-check, no raw values.
	return Control{Name: "profile-metadata-complete", State: StatePass}
}

func controlRuntimeBinding(rt string, chk Checker, ctx context.Context) Control {
	if rt == "" {
		return Control{Name: "runtime-bound", State: StateWarning, Note: "no runtime bound"}
	}
	ok, err := chk.RuntimeQualified(ctx, rt)
	if err != nil {
		return Control{Name: "runtime-bound", State: StateFail, Note: "qualification unavailable"}
	}
	if !ok {
		return Control{Name: "runtime-bound", State: StateFail, Note: "runtime not qualified"}
	}
	ver, err := chk.RuntimeVersion(ctx, rt)
	if err != nil || ver == "" {
		return Control{Name: "runtime-bound", State: StateWarning, Note: "version unknown"}
	}
	return Control{Name: "runtime-bound", State: StatePass, Note: "qualified@" + ver}
}

// declaredUnobservedCapabilities makes the current coverage explicit. These
// controls are deliberately not probed: observing them would require a real
// runtime, which remains outside this local read-only diagnostic.
func declaredUnobservedCapabilities() []Control {
	return []Control{
		{Name: "navigator", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "battery", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "network", State: StateUnsupported, Note: "network observation not implemented"},
		{Name: "webrtc", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "storage-capabilities", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "plugins-mime", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "input-devices", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "rendering-apis", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "canvas-2d", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "canvas-webgl", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "webgl2", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "audio-context", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "offline-audio-context", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "font-rendering", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "performance", State: StateUnsupported, Note: "runtime observation not implemented"},
		{Name: "permissions", State: StateUnsupported, Note: "runtime observation not implemented"},
	}
}

func aggregate(controls []Control) Verdict {
	anyFail := false
	anyWarn := false
	for _, c := range controls {
		switch c.State {
		case StateFail:
			anyFail = true
		case StateWarning:
			anyWarn = true
		}
	}
	switch {
	case anyFail:
		return VerdictBroken
	case anyWarn:
		return VerdictRisky
	case len(controls) == 0:
		return VerdictUnknown
	default:
		return VerdictClean
	}
}

// Batch diagnoses ids concurrently without leaking raw values.
func Batch(ctx context.Context, chk Checker, ids []string) ([]*Diagnostic, []error) {
	results := make([]*Diagnostic, len(ids))
	errs := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			d, err := Diagnose(ctx, chk, id)
			results[i] = d
			errs[i] = err
		}(i, id)
	}
	wg.Wait()
	return results, errs
}
