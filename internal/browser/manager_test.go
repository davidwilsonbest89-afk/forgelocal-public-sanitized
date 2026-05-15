package browser

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestShouldRetryLaunchForProtocolEOF(t *testing.T) {
	cases := []string{
		"target closed: could not read protocol padding: EOF",
		"launch firefox: target closed",
		"unexpected EOF",
	}

	for _, msg := range cases {
		if !shouldRetryLaunch(errors.New(msg)) {
			t.Fatalf("expected retryable launch error for %q", msg)
		}
	}
}

func TestShouldRetryLaunchRejectsRegularErrors(t *testing.T) {
	if shouldRetryLaunch(errors.New("profile appears to be in use")) {
		t.Fatal("profile lock errors should not restart Playwright")
	}
}

func TestCleanProfileLocks(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("lock"), 0644); err != nil {
			t.Fatalf("write lock %s: %v", name, err)
		}
	}

	cleanProfileLocks(dir)

	for _, name := range []string{"SingletonLock", "SingletonCookie", "SingletonSocket"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got err=%v", name, err)
		}
	}
}
