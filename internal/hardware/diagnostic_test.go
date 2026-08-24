package hardware

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestT34ProjectIsDeterministicAndReadOnly(t *testing.T) {
	first := Project()
	second := Project()
	if first.DiagnosticVersion != DiagnosticVersion || first.ObservationMode != ObservationMode {
		t.Fatalf("unexpected header: %#v", first)
	}
	if len(first.Controls) != 7 || len(second.Controls) != 7 {
		t.Fatalf("unexpected control count: %d/%d", len(first.Controls), len(second.Controls))
	}
	for i := range first.Controls {
		if first.Controls[i] != second.Controls[i] {
			t.Fatalf("control %d changed: %#v vs %#v", i, first.Controls[i], second.Controls[i])
		}
		if first.Controls[i].State != StateUnsupported || first.Controls[i].Note != "host observation not implemented" {
			t.Fatalf("control %d is not fail-closed: %#v", i, first.Controls[i])
		}
	}
}

func TestT34ProjectDoesNotExposeHostValues(t *testing.T) {
	body, err := json.Marshal(Project())
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(body)
	for _, forbidden := range []string{"hostname", "machine-id", "serial", "/proc", "/sys", "cpuinfo", "meminfo", "MAC", "127.0.0.1"} {
		if strings.Contains(strings.ToLower(encoded), strings.ToLower(forbidden)) {
			t.Fatalf("host value hint %q leaked: %s", forbidden, encoded)
		}
	}
}
