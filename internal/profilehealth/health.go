// Package profilehealth aggregates redacted diagnostic states.
// It has no profile storage, runtime, network, or raw-value dependency.
package profilehealth

import "sort"

const Version = "profile-health-projection-v1"
const ObservationMode = "PROJECTED_METADATA_ONLY"

type CheckState string

const (
	CheckPass        CheckState = "PASS"
	CheckWarning     CheckState = "WARNING"
	CheckFail        CheckState = "FAIL"
	CheckUnsupported CheckState = "UNSUPPORTED"
)

type HealthState string

const (
	HealthHealthy HealthState = "HEALTHY"
	HealthAtRisk  HealthState = "AT_RISK"
	HealthBroken  HealthState = "BROKEN"
	HealthUnknown HealthState = "UNKNOWN"
)

type Check struct {
	Key   string     `json:"key"`
	State CheckState `json:"state"`
}

type Explanation struct {
	Code string `json:"code"`
}

type Health struct {
	Version         string        `json:"version"`
	ObservationMode string        `json:"observation_mode"`
	State           HealthState   `json:"state"`
	Explanations    []Explanation `json:"explanations"`
}

// Aggregate returns only fixed state and explanation codes; it never echoes
// check values, profile data, host data, or runtime identifiers.
func Aggregate(checks []Check) Health {
	ordered := append([]Check(nil), checks...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Key < ordered[j].Key })
	state := HealthHealthy
	explanations := make([]Explanation, 0)
	seen := map[string]bool{}
	for _, check := range ordered {
		if check.Key == "" || seen[check.Key] {
			continue
		}
		seen[check.Key] = true
		switch check.State {
		case CheckFail:
			state = HealthBroken
			if len(explanations) == 0 || explanations[len(explanations)-1].Code != "CHECK_FAILED" {
				explanations = append(explanations, Explanation{Code: "CHECK_FAILED"})
			}
		case CheckWarning:
			if state != HealthBroken {
				state = HealthAtRisk
			}
			if !containsCode(explanations, "CHECK_WARNING") {
				explanations = append(explanations, Explanation{Code: "CHECK_WARNING"})
			}
		case CheckUnsupported:
			if state == HealthHealthy {
				state = HealthUnknown
			}
			if !containsCode(explanations, "OBSERVATION_UNSUPPORTED") {
				explanations = append(explanations, Explanation{Code: "OBSERVATION_UNSUPPORTED"})
			}
		}
	}
	return Health{Version: Version, ObservationMode: ObservationMode, State: state, Explanations: explanations}
}

func containsCode(explanations []Explanation, code string) bool {
	for _, explanation := range explanations {
		if explanation.Code == code {
			return true
		}
	}
	return false
}
