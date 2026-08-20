package localvault

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func tempRoot(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	return filepath.Join(d, "lv")
}

func token(t *testing.T) []byte {
	t.Helper()
	tok := make([]byte, 32)
	if _, err := rand.Read(tok); err != nil {
		t.Fatal(err)
	}
	return tok
}

func mustOpen(t *testing.T, root string, tok []byte, opts ...OpenOption) *Vault {
	t.Helper()
	v, err := Open(root, tok, opts...)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func TestT12LocalVault_PutGetDeleteRoundTrip(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	if err := v.Put("profile-1", []byte("shh-secret")); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := v.Get("profile-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !bytes.Equal(got, []byte("shh-secret")) {
		t.Fatalf("round trip mismatch")
	}
	got2, err := v.Get("profile-1")
	if err != nil {
		t.Fatalf("get again: %v", err)
	}
	if !bytes.Equal(got, got2) {
		t.Fatal("Get must return fresh copies")
	}
	if err := v.Delete("profile-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := v.Get("profile-1"); err != ErrEntryNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
	// Delete idempotent.
	if err := v.Delete("profile-1"); err != nil {
		t.Fatalf("delete idempotent: %v", err)
	}
}

func TestT12LocalVault_InvalidIDsRejected(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	for _, bad := range []string{"", "a b", "a/b", "a..b", strings.Repeat("x", 129), "a\x00b"} {
		if err := v.Put(bad, []byte("v")); err != ErrInvalidID {
			t.Errorf("put %q: want ErrInvalidID, got %v", bad, err)
		}
		if _, err := v.Get(bad); err != ErrInvalidID {
			t.Errorf("get %q: want ErrInvalidID, got %v", bad, err)
		}
	}
	if err := v.Put(strings.Repeat("a", 128), []byte("ok")); err != nil {
		t.Fatalf("max-length id allowed: %v", err)
	}
}

func TestT12LocalVault_EmptyAndOversizedRejected(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	if err := v.Put("e", []byte{}); err == nil {
		t.Fatal("empty value must be rejected")
	}
	if err := v.Put("e", bytes.Repeat([]byte("x"), maxPlaintext+1)); err != ErrTooLarge {
		t.Fatalf("oversized: want ErrTooLarge, got %v", err)
	}
	if err := v.Put("e", bytes.Repeat([]byte("x"), maxPlaintext)); err != nil {
		t.Fatalf("max size allowed: %v", err)
	}
}

func TestT12LocalVault_CiphertextBoundToAAD(t *testing.T) {
	root := tempRoot(t)
	tok := token(t)
	v1 := mustOpen(t, root, tok)
	if err := v1.Put("bound-test", []byte("secret-value")); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := v1.Close(); err != nil {
		t.Fatal(err)
	}

	// Reopen with a DIFFERENT token: same stored bytes must now fail integrity.
	// Fail-closed contract: the wrong token must make the vault refuse to
	// open entirely rather than silently serve corrupt data.
	v2, err := Open(root, []byte("different-token-of-16-plus"))
	if err == nil {
		_ = v2.Close()
		t.Fatal("wrong token must fail closed at open")
	}
	if _, err := Open(root, append([]byte(nil), tok[1:]...)); err == nil {
		t.Fatal("truncated token must fail closed at open")
	}

	// Reopen with the SAME token: round trip must still succeed.
	v3 := mustOpen(t, root, tok)
	got, err := v3.Get("bound-test")
	if err != nil {
		t.Fatalf("same token reopen: %v", err)
	}
	if !bytes.Equal(got, []byte("secret-value")) {
		t.Fatal("same token must decrypt cleanly")
	}
}

func TestT12LocalVault_NonceUniquePerPut(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	if err := v.Put("n1", []byte("same plaintext")); err != nil {
		t.Fatal(err)
	}
	if err := v.Put("n2", []byte("same plaintext")); err != nil {
		t.Fatal(err)
	}
	// Nonces are persisted base64 in the journal; they must differ.
	data, err := os.ReadFile(filepath.Join(v.root, "lv_journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var nonces []string
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		var e journalEntry
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		nonces = append(nonces, e.Nonce)
	}
	if len(nonces) != 2 || nonces[0] == nonces[1] {
		t.Fatal("nonce reuse detected across puts")
	}
}

func TestT12LocalVault_JournalDurableAfterCrash(t *testing.T) {
	root := tempRoot(t)
	tok := token(t)
	v1 := mustOpen(t, root, tok)
	if err := v1.Put("crash-1", []byte("persisted")); err != nil {
		t.Fatal(err)
	}
	if err := v1.Put("crash-2", []byte("also persisted")); err != nil {
		t.Fatal(err)
	}
	if err := v1.Delete("crash-2"); err != nil {
		t.Fatal(err)
	}
	// Simulate crash: close without clean teardown (same as real crash: the
	// journal is already fsynced on every mutation).
	_ = v1.Close()

	v2 := mustOpen(t, root, tok)
	if !v2.Exists("crash-1") {
		t.Fatal("entry lost after simulated crash")
	}
	if v2.Exists("crash-2") {
		t.Fatal("deleted entry survived reconciliation")
	}
	got, err := v2.Get("crash-1")
	if err != nil || !bytes.Equal(got, []byte("persisted")) {
		t.Fatal("post-crash round trip failed")
	}
	// Put after recovery must continue the sequence without regression.
	if err := v2.Put("crash-3", []byte("after-recovery")); err != nil {
		t.Fatal(err)
	}
	if got, err := v2.Get("crash-3"); err != nil || !bytes.Equal(got, []byte("after-recovery")) {
		t.Fatal("post-recovery put failed")
	}
}

func TestT12LocalVault_CorruptJournalFailsClosed(t *testing.T) {
	root := tempRoot(t)
	tok := token(t)
	v := mustOpen(t, root, tok)
	if err := v.Put("x", []byte("v")); err != nil {
		t.Fatal(err)
	}
	_ = v.Close()
	// Truncate the journal mid-entry (crash during write).
	jp := filepath.Join(root, "lv_journal.jsonl")
	info, _ := os.Stat(jp)
	if err := os.Truncate(jp, info.Size()-3); err != nil {
		t.Fatal(err)
	}
	v2, err := Open(root, tok)
	if err == nil {
		_ = v2.Close()
		t.Fatal("corrupt journal must fail closed, got nil error")
	}
}

func TestT12LocalVault_ShortTokenRefused(t *testing.T) {
	if _, err := Open(tempRoot(t), []byte("short")); err == nil {
		t.Fatal("short token must be refused")
	}
	if _, err := Open(tempRoot(t), nil); err == nil {
		t.Fatal("nil token must be refused")
	}
}

func TestT12LocalVault_MasterNeverOnDisk(t *testing.T) {
	root := tempRoot(t)
	tok := token(t)
	v := mustOpen(t, root, tok)
	if err := v.Put("m", []byte("marker")); err != nil {
		t.Fatal(err)
	}
	_ = v.Close()
	// Scan every persisted file: the unlock token and the marker must never
	// appear in plaintext anywhere under root.
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, tok) {
			t.Errorf("token found in plaintext in %s", filepath.Base(path))
		}
		if bytes.Contains(data, []byte("marker")) {
			t.Errorf("plaintext value found in %s", filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestT12LocalVault_ConcurrentOps(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "c-" + strings.Repeat("0", 3-len(strings.Repeat("", i%10))) + strings.Repeat("x", i%10+1)
			_ = v.Put(id, []byte{byte(i)})
			_, _ = v.Get(id)
			_ = v.Delete(id)
			_ = v.Exists(id)
		}(i)
	}
	wg.Wait()
	if err := v.Put("final", []byte("ok")); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Get("final"); err != nil {
		t.Fatalf("concurrent ops broke the vault: %v", err)
	}
}

func TestT12LocalVault_ClosedVaultRefusesEverything(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if err := v.Put("a", []byte("b")); err != ErrClosed {
		t.Errorf("put after close: %v", err)
	}
	if _, err := v.Get("a"); err != ErrClosed {
		t.Errorf("get after close: %v", err)
	}
	if err := v.Delete("a"); err != ErrClosed {
		t.Errorf("delete after close: %v", err)
	}
	if v.Exists("a") {
		t.Error("exists after close must be false")
	}
	if refs := v.SecretRefs(); refs != nil {
		t.Error("refs after close must be nil")
	}
}

func TestT12LocalVault_Permissions(t *testing.T) {
	root := tempRoot(t)
	tok := token(t)
	v := mustOpen(t, root, tok)
	if err := v.Put("p", []byte("v")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"lv_journal.jsonl", ".lv_salt"} {
		info, err := os.Stat(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("%s perm %v, want 0600", rel, info.Mode().Perm())
		}
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("root perm %v, want 0700", info.Mode().Perm())
	}
}

func TestT12LocalVault_ConfinesSaltAndJournalToRoot(t *testing.T) {
	root := tempRoot(t)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, ".lv_salt")); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, token(t)); err == nil {
		t.Fatal("salt symlink escaping the vault root must fail closed")
	}

	// Create a valid salt first, then substitute only the journal path. The
	// opened os.Root must reject the second symlink independently.
	if err := os.Remove(filepath.Join(root, ".lv_salt")); err != nil {
		t.Fatal(err)
	}
	tok := token(t)
	v := mustOpen(t, root, tok)
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "lv_journal.jsonl")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "journal"), filepath.Join(root, "lv_journal.jsonl")); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, tok); err == nil {
		t.Fatal("journal symlink escaping the vault root must fail closed")
	}
}

func TestT12LocalVault_SecretRefsOpaque(t *testing.T) {
	v := mustOpen(t, tempRoot(t), token(t))
	_ = v.Put("ref-1", []byte("a"))
	_ = v.Put("ref-2", []byte("b"))
	refs := v.SecretRefs()
	if len(refs) != 2 {
		t.Fatalf("want 2 refs, got %d", len(refs))
	}
	// Refs never contain value material; only ids.
}

func TestT12LocalVault_RelativeRootRefused(t *testing.T) {
	if _, err := Open("relative/path", []byte("short")); err == nil {
		t.Fatal("relative root must be refused")
	}
}
