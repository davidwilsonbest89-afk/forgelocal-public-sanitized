package drift

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestT36CompareIsStableWithinExplicitThreshold(t *testing.T) {
	baseline := []Control{{Key: "canvas", State: "UNSUPPORTED"}, {Key: "runtime", State: "PASS"}}
	report := Compare("baseline-sha", "current-sha", baseline, baseline, 0)
	if !report.WithinLimit || len(report.Findings) != 0 || report.Version != Version {
		t.Fatalf("unexpected stable report: %#v", report)
	}
}

func TestT36CompareDetectsRedactedStateDriftDeterministically(t *testing.T) {
	baseline := []Control{{Key: "zeta", State: "PASS"}, {Key: "alpha", State: "PASS"}, {Key: "missing", State: "PASS"}}
	current := []Control{{Key: "zeta", State: "FAIL"}, {Key: "alpha", State: "PASS"}, {Key: "added", State: "UNSUPPORTED"}}
	report := Compare("b", "c", baseline, current, 1)
	if report.WithinLimit || len(report.Findings) != 3 {
		t.Fatalf("threshold/findings mismatch: %#v", report)
	}
	want := []Finding{{Key: "added", State: StateAdded}, {Key: "missing", State: StateMissing}, {Key: "zeta", State: StateChanged}}
	for i := range want {
		if report.Findings[i] != want[i] {
			t.Fatalf("finding %d=%#v, want %#v", i, report.Findings[i], want[i])
		}
	}
	body, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"canvas fingerprint", "UserAgent", "127.0.0.1", "raw_value"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("raw value leaked: %s", body)
		}
	}
}

func TestT36NegativeThresholdFailsClosedToZero(t *testing.T) {
	report := Compare("b", "c", nil, []Control{{Key: "a", State: "PASS"}}, -1)
	if report.MaxChanges != 0 || report.WithinLimit {
		t.Fatalf("negative threshold was not fail-closed: %#v", report)
	}
}
