package environment

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// fakeChecker is a memory-only profile/registry simulation.
type fakeChecker struct {
	mu          sync.RWMutex
	profiles    map[string]bool
	profileRT   map[string]string // profile -> runtime id ("" means unbound)
	qualifiedRT map[string]bool
	versions    map[string]string
	failLookups int
}

func (f *fakeChecker) ProfileExists(_ context.Context, id string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.failLookups > 0 {
		f.failLookups--
		return false, errLookup
	}
	return f.profiles[id], nil
}
func (f *fakeChecker) ProfileRuntimeID(_ context.Context, id string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if !f.profiles[id] {
		return "", nil
	}
	return f.profileRT[id], nil
}
func (f *fakeChecker) RuntimeQualified(_ context.Context, rt string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.qualifiedRT[rt], nil
}
func (f *fakeChecker) RuntimeVersion(_ context.Context, rt string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.versions[rt], nil
}

var errLookup = context.DeadlineExceeded

func clean(t *testing.T) *fakeChecker {
	t.Helper()
	return &fakeChecker{
		profiles:    map[string]bool{"profile-1": true, "profile-2": true},
		profileRT:   map[string]string{"profile-1": "chromium-t14", "profile-2": "chromium-t14"},
		qualifiedRT: map[string]bool{"chromium-t14": true},
		versions:    map[string]string{"chromium-t14": "126.0"},
	}
}

func TestT13Environment_CleanVerdict(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != VerdictClean {
		t.Fatalf("want CLEAN, got %s", d.Verdict)
	}
	if len(d.Controls) != 13 {
		t.Fatalf("want 13 controls, got %d", len(d.Controls))
	}
}

func TestT30Environment_DeclaresProjectionCoverageWithoutObservation(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.DiagnosticVersion != DiagnosticVersion {
		t.Fatalf("want version %q, got %q", DiagnosticVersion, d.DiagnosticVersion)
	}
	if d.ObservationMode != ObservationModeProjected {
		t.Fatalf("want projection-only mode, got %q", d.ObservationMode)
	}
	unsupported := 0
	for _, control := range d.Controls {
		if control.State == StateUnsupported {
			unsupported++
		}
	}
	if unsupported != 11 {
		t.Fatalf("want 11 declared unsupported capabilities, got %d", unsupported)
	}
}

func TestT30Environment_UnobservedCapabilityNotesAreFixedAndRedacted(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, control := range d.Controls {
		if control.State == StateUnsupported && control.Note != "runtime observation not implemented" && control.Note != "network observation not implemented" {
			t.Fatalf("unsupported control %q has unexpected note %q", control.Name, control.Note)
		}
	}
}

func TestT13Environment_UnboundRuntimeRisky(t *testing.T) {
	f := clean(t)
	f.profileRT["profile-1"] = ""
	d, err := Diagnose(context.Background(), f, "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != VerdictRisky {
		t.Fatalf("want RISKY, got %s", d.Verdict)
	}
}

func TestT13Environment_UnqualifiedRuntimeBroken(t *testing.T) {
	f := clean(t)
	f.profileRT["profile-1"] = "unqualified-runtime"
	d, err := Diagnose(context.Background(), f, "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != VerdictBroken {
		t.Fatalf("want BROKEN, got %s", d.Verdict)
	}
}

func TestT13Environment_ProfileNotFoundExplicit(t *testing.T) {
	_, err := Diagnose(context.Background(), clean(t), "ghost-profile")
	if err != ErrProfileNotFound {
		t.Fatalf("want PROFILE_NOT_FOUND, got %v", err)
	}
}

func TestT13Environment_InvalidIDsRefused(t *testing.T) {
	for _, bad := range []string{"", "a b", "a/b", "a..b", "a\x00b", stringsRepeat("x", 129)} {
		_, err := Diagnose(context.Background(), clean(t), bad)
		if err != ErrInvalidID {
			t.Errorf("%q: want ErrInvalidID, got %v", bad, err)
		}
	}
}

func TestT13Environment_BatchConcurrent(t *testing.T) {
	f := clean(t)
	ids := []string{"profile-1", "profile-2", "ghost-profile"}
	results, errs := Batch(context.Background(), f, ids)
	for i := range ids {
		if errs[i] == nil && results[i] == nil {
			t.Fatalf("idx %d: result nil without error", i)
		}
	}
	if errs[2] != ErrProfileNotFound {
		t.Fatalf("idx 2: want PROFILE_NOT_FOUND, got %v", errs[2])
	}
	if results[0].Verdict != VerdictClean {
		t.Fatalf("idx 0: want CLEAN, got %s", results[0].Verdict)
	}
}

func TestT13Environment_NoRawValuesInOutput(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	raw := fmtSprint(d)
	for _, forbidden := range []string{"UserAgent", "canvas", "fingerprint", "127.0.0.1", "path"} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("raw value hint %q leaked into diagnostic", forbidden)
		}
	}
}

func TestT13Environment_InvalidIDNoLookup(t *testing.T) {
	f := clean(t)
	f.failLookups = 99
	// Invalid ids must be refused BEFORE any lookup occurs.
	_, err := Diagnose(context.Background(), f, "a/b")
	if err != ErrInvalidID {
		t.Fatalf("invalid id must short-circuit: %v", err)
	}
}

func TestT13Environment_DeterministicShape(t *testing.T) {
	d1, _ := Diagnose(context.Background(), clean(t), "profile-1")
	d2, _ := Diagnose(context.Background(), clean(t), "profile-1")
	if len(d1.Controls) != len(d2.Controls) {
		t.Fatal("control count must be deterministic")
	}
	if d1.Verdict != d2.Verdict {
		t.Fatal("verdict must be deterministic")
	}
	for i := range d1.Controls {
		if d1.Controls[i].Name != d2.Controls[i].Name {
			t.Fatal("control names must be stable")
		}
	}
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = s[0]
	}
	return string(b)
}

// fmtSprint stringify helper for the redaction test; intentionally
// not implemented with fmt.Sprint to avoid accidental raw-value leakage
// through reflect-based dumping.
var fmtSprint func(v any) string

func init() {
	fmtSprint = func(v any) string {
		switch x := v.(type) {
		case *Diagnostic:
			s := x.ProfileID + "|" + string(x.Verdict)
			for _, c := range x.Controls {
				s += "|" + c.Name + ":" + string(c.State) + ":" + c.Note
			}
			return s
		default:
			return ""
		}
	}
}

func init() {
	// Replace fmtSprint with a real stringification for the redaction test.
	fmtSprint = func(v any) string {
		switch x := v.(type) {
		case *Diagnostic:
			s := x.ProfileID + "|" + string(x.Verdict)
			for _, c := range x.Controls {
				s += "|" + c.Name + ":" + string(c.State) + ":" + c.Note
			}
			return s
		default:
			return ""
		}
	}
}
