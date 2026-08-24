// Package hardware exposes a synthetic, redacted hardware diagnostic contract.
// It intentionally performs no host probing and has no network or runtime side effects.
package hardware

const (
	DiagnosticVersion = "hardware-projection-v1"
	ObservationMode   = "PROJECTED_METADATA_ONLY"
)

type State string

const (
	StateUnsupported State = "UNSUPPORTED"
)

type Control struct {
	Name  string `json:"name"`
	State State  `json:"state"`
	Note  string `json:"note"`
}

type Diagnostic struct {
	DiagnosticVersion string    `json:"diagnostic_version"`
	ObservationMode   string    `json:"observation_mode"`
	Controls          []Control `json:"controls"`
}

// Project returns the closed T34 capability set. No host value is read or
// copied into the response; unsupported is deliberately fail-closed.
func Project() Diagnostic {
	names := []string{"cpu", "memory", "storage", "gpu", "display", "network-adapters", "thermal"}
	controls := make([]Control, 0, len(names))
	for _, name := range names {
		controls = append(controls, Control{Name: name, State: StateUnsupported, Note: "host observation not implemented"})
	}
	return Diagnostic{DiagnosticVersion: DiagnosticVersion, ObservationMode: ObservationMode, Controls: controls}
}
