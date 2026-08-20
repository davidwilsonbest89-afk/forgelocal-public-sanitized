package browser

import "testing"

func TestAcquireSnapshotRejectsConcurrentReservationAndReleases(t *testing.T) {
	m := &Manager{sessions: make(map[string]*Session)}
	release, err := m.AcquireSnapshot("profile-one")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.AcquireSnapshot("profile-one"); err == nil {
		t.Fatal("expected concurrent snapshot reservation to be rejected")
	}
	release()
	if _, err := m.AcquireSnapshot("profile-one"); err != nil {
		t.Fatalf("snapshot should be acquirable after release: %v", err)
	}
}

func TestAcquireSnapshotRejectsLiveBrowserSession(t *testing.T) {
	m := &Manager{sessions: map[string]*Session{
		"session-one": {ID: "session-one", ProfileID: "profile-live"},
	}}
	if _, err := m.AcquireSnapshot("profile-live"); err == nil {
		t.Fatal("expected active profile session to reject snapshot")
	}
}
