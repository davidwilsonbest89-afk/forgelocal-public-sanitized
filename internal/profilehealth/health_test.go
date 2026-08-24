package profilehealth

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestT37AggregateStatesAreFailClosed(t *testing.T) {
	cases := []struct {
		name   string
		checks []Check
		want   HealthState
	}{
		{"healthy", []Check{{Key: "runtime", State: CheckPass}}, HealthHealthy},
		{"warning", []Check{{Key: "runtime", State: CheckPass}, {Key: "drift", State: CheckWarning}}, HealthAtRisk},
		{"broken", []Check{{Key: "runtime", State: CheckFail}, {Key: "drift", State: CheckWarning}}, HealthBroken},
		{"unsupported", []Check{{Key: "canvas", State: CheckUnsupported}}, HealthUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Aggregate(tc.checks)
			if got.State != tc.want || got.Version != Version || got.ObservationMode != ObservationMode {
				t.Fatalf("got %#v, want state %q", got, tc.want)
			}
		})
	}
}

func TestT37AggregateIsDeterministicAndRedacted(t *testing.T) {
	checks := []Check{{Key: "z", State: CheckFail}, {Key: "a", State: CheckUnsupported}, {Key: "z", State: CheckFail}}
	first := Aggregate(checks)
	second := Aggregate(checks)
	body1, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	body2, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(body1) != string(body2) {
		t.Fatalf("non-deterministic JSON: %s vs %s", body1, body2)
	}
	for _, forbidden := range []string{"UserAgent", "canvas fingerprint", "127.0.0.1", "raw_value", "runtime-id"} {
		if strings.Contains(string(body1), forbidden) {
			t.Fatalf("raw detail leaked %q: %s", forbidden, body1)
		}
	}
	if len(first.Explanations) != 2 {
		t.Fatalf("want two unique explanations, got %#v", first.Explanations)
	}
}
