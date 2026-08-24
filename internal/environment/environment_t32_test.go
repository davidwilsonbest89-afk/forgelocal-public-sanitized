package environment

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestT32ClientRectsIsExplicitlyUnsupported(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, control := range d.Controls {
		if control.Name != "client-rects" {
			continue
		}
		if control.State != StateUnsupported {
			t.Fatalf("client-rects state=%q, want UNSUPPORTED", control.State)
		}
		if control.Note != "runtime observation not implemented" {
			t.Fatalf("client-rects note=%q", control.Note)
		}
		return
	}
	t.Fatal("client-rects control is missing")
}

func TestT32ClientRectsDoesNotExposeGeometry(t *testing.T) {
	d, err := Diagnose(context.Background(), clean(t), "profile-1")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"clientRect", "DOMRect", "boundingClientRect", "x:", "y:", "width:", "height:", "getBoundingClientRect"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("T32 geometry hint %q leaked: %s", forbidden, body)
		}
	}
}
