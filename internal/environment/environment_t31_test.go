package environment

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestT31RenderingControlsAreExplicitlyProjectedAndUnsupported(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"canvas-2d", "canvas-webgl", "webgl2", "audio-context", "offline-audio-context"}
	seen := map[string]Control{}
	for _, control := range d.Controls {
		seen[control.Name] = control
	}
	for _, name := range want {
		control, ok := seen[name]
		if !ok {
			t.Fatalf("T31 control %q missing", name)
		}
		if control.State != StateUnsupported {
			t.Fatalf("T31 control %q state=%q, want UNSUPPORTED", name, control.State)
		}
		if control.Note != "runtime observation not implemented" {
			t.Fatalf("T31 control %q note=%q", name, control.Note)
		}
	}
}

func TestT31RenderingControlsHaveStableOrderAndNoRawValues(t *testing.T) {
	first, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Controls) != len(second.Controls) {
		t.Fatalf("control length changed: %d vs %d", len(first.Controls), len(second.Controls))
	}
	for i := range first.Controls {
		if first.Controls[i] != second.Controls[i] {
			t.Fatalf("control %d is not deterministic: %#v vs %#v", i, first.Controls[i], second.Controls[i])
		}
	}
	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"canvas fingerprint", "canvas_value", "webgl vendor", "audio hash", "AudioBuffer", "UserAgent", "127.0.0.1", "binary_hash_sha256"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("T31 raw value hint %q leaked: %s", forbidden, body)
		}
	}
}
