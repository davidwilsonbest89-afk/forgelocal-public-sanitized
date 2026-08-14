package backup

import (
	"errors"
	"os"
	"testing"
)

func TestVerifyAcceptsPublishedArtifactAndRejectsCorruption(t *testing.T) {
	svc, _, _ := newTestService(t)
	created, err := svc.Create("profile-verify", "key-test-01", []byte(`{"metadata":"migration-preimage"}`))
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if err := svc.Verify(created.ID); err != nil {
		t.Fatalf("verify intact backup: %v", err)
	}

	body, err := os.ReadFile(created.ArtifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	body[len(body)/2] ^= 0x01
	if err := os.WriteFile(created.ArtifactPath, body, 0600); err != nil {
		t.Fatalf("corrupt artifact: %v", err)
	}
	if err := svc.Verify(created.ID); !errors.Is(err, ErrIntegrity) {
		t.Fatalf("verify corrupt backup error = %v, want ErrIntegrity", err)
	}
}
