package sessiontrack

import "testing"

func TestLifecycleIsRedactedAndSorted(t *testing.T) {
	r := New()
	if err := r.Start("z-session"); err != nil {
		t.Fatal(err)
	}
	if err := r.Start("a-session"); err != nil {
		t.Fatal(err)
	}
	if err := r.Transition("z-session", Starting, "OBSERVATION_UNSUPPORTED"); err != nil {
		t.Fatal(err)
	}
	if err := r.Transition("z-session", Running, "LIFECYCLE_STARTED"); err != nil {
		t.Fatal(err)
	}
	v := r.Snapshot()
	if len(v) != 2 || v[0].SessionKey != "a-session" || v[1].State != Running {
		t.Fatalf("unexpected snapshot: %#v", v)
	}
	if v[1].ReasonCode != "LIFECYCLE_STARTED" {
		t.Fatalf("unexpected reason: %#v", v[1])
	}
}

func TestInvalidOrRawInputsFailClosed(t *testing.T) {
	r := New()
	for _, key := range []string{"", "../secret", "line\nvalue"} {
		if r.Start(key) == nil {
			t.Fatalf("accepted key %q", key)
		}
	}
	if err := r.Start("ok"); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
	if r.Transition("ok", State("UNKNOWN"), "x") == nil {
		t.Fatal("accepted unsupported state")
	}
	if r.Transition("missing", Running, "x") == nil {
		t.Fatal("accepted unknown session")
	}
	if r.Transition("ok", Failed, "raw-cookie=abc") != nil {
		t.Fatal("accepted raw reason")
	}
}
