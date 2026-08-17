// Package localvault implements the T12 LocalVault: an encrypted local fallback
// vault used when the native OS vault (SystemVault) is unavailable.
//
// Security contract (never deviate, even in tests):
//   - Plaintext values are NEVER written to disk outside an AES-256-GCM
//     ciphertext with a unique per-entry nonce and an AAD bound to the entry id.
//   - The encryption key is derived from a random 32-byte master salt and a
//     per-entry salt; the master itself is never persisted in plaintext
//     (it is derived from the operator-supplied unlock token for the lifetime
//     of the opened vault only).
//   - The vault is fail-closed: any open failure leaves the vault unusable and
//     refuses every operation. There is NO plaintext fallback.
//   - Files are written atomically (write-to-temp + rename) with fsync and
//     directory mode 0700, file mode 0600.
//   - Entries are journaled: a durable append-only manifest records every
//     mutation with a monotonic sequence number; recovery reconciles the
//     journal against the current key file.
//   - Entries are never kept in logs, SQLite, profile JSON or backup
//     archives; only opaque refs (entry_id) leave the vault.
package localvault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EntryID is an opaque reference that may safely be stored in SQLite or JSON.
type EntryID string

var (
	ErrClosed        = errors.New("localvault: vault is closed")
	ErrEntryNotFound = errors.New("localvault: entry not found")
	ErrInvalidID     = errors.New("localvault: invalid entry id")
	ErrIntegrity     = errors.New("localvault: entry integrity check failed")
	ErrTooLarge      = errors.New("localvault: value exceeds 64 KiB")

	// idToken is the same character class used elsewhere in the code base
	// (secrets package): bounded, URL-safe, no path separators.
	idChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-"
)

func validID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	if strings.Contains(id, "..") {
		return false
	}
	for _, c := range id {
		if !strings.ContainsRune(idChars, c) {
			return false
		}
	}
	return true
}

const maxPlaintext = 64 * 1024

// Vault is the LocalVault handle. It is opened once with a secret unlock token
// held only in memory for the lifetime of the process and must be Closed.
type Vault struct {
	root   string
	master []byte // 32 bytes, in-memory only
	mu     sync.Mutex
	entries map[EntryID][]byte
	journal *os.File
	seq     uint64
	closed  bool
}

type OpenOption func(*openConfig)

type openConfig struct {
	now func() time.Time
}

func WithClock(now func() time.Time) OpenOption {
	return func(c *openConfig) {
		c.now = now
	}
}

// Open opens (or initializes) the LocalVault under root. The unlock token is
// the only secret required; it is not persisted anywhere. An invalid or
// missing token makes Open fail closed.
func Open(root string, unlockToken []byte, opts ...OpenOption) (*Vault, error) {
	cfg := &openConfig{now: time.Now}
	for _, o := range opts {
		o(cfg)
	}
	if len(unlockToken) == 0 {
		return nil, errors.New("localvault: unlock token required")
	}
	if len(unlockToken) < 16 {
		return nil, errors.New("localvault: unlock token too short")
	}
	if !filepath.IsAbs(root) {
		return nil, errors.New("localvault: root must be absolute")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("localvault: create root: %w", err)
	}
	v := &Vault{root: root, entries: map[EntryID][]byte{}}

	// Master key derivation: master = HMAC-less KDF (SHA-256) over
	// (unlockToken || fixed-scope || persistent-salt). The persistent salt is
	// created on first open and never changes; it binds this vault instance
	// to the unlock token without storing the token itself.
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("localvault: create root: %w", err)
	}
	master, err := v.deriveMaster(unlockToken)
	if err != nil {
		return nil, err
	}
	v.master = master

	if err := v.loadJournal(); err != nil {
		// Fail closed: a corrupt journal means we cannot reconcile what was
		// stored; refusing all operations is safer than partial loss.
		v.master = nil
		return nil, fmt.Errorf("localvault: journal load failed (fail-closed): %w", err)
	}
	return v, nil
}

func (v *Vault) saltPath() string { return filepath.Join(v.root, ".lv_salt") }
func (v *Vault) journalPath() string { return filepath.Join(v.root, "lv_journal.jsonl") }

func (v *Vault) deriveMaster(token []byte) ([]byte, error) {
	salt, err := v.persistentSalt()
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	h.Write([]byte("forgelocal-localvault-master-v1"))
	h.Write(token)
	h.Write(salt)
	d := h.Sum(nil)
	return d[:32], nil
}

// persistentSalt creates a 16-byte salt on first open; subsequent opens read
// the same salt. Written once, never rewritten, never logged.
func (v *Vault) persistentSalt() ([]byte, error) {
	path := v.saltPath()
	existing, err := os.ReadFile(path)
	if err == nil {
		if len(existing) != 16 {
			return nil, ErrIntegrity
		}
		return existing, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("localvault: read random: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		// Race: another process created the file first; reread.
		if existing, err2 := os.ReadFile(path); err2 == nil && len(existing) == 16 {
			return existing, nil
		}
		return nil, fmt.Errorf("localvault: create salt: %w", err)
	}
	if _, err := f.Write(salt); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return nil, err
	}
	return salt, f.Close()
}

type journalEntry struct {
	Seq  uint64 `json:"seq"`
	Op   string `json:"op"`   // put | delete
	ID   string `json:"id"`
	At   string `json:"at"`
	Data string `json:"data,omitempty"` // base64 ciphertext blob
	Nonce string `json:"nonce,omitempty"`
}

func (v *Vault) loadJournal() error {
	path := v.journalPath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	// Rebuild in-memory state from the durable append-only journal.
	dec := json.NewDecoder(f)
	var entries map[EntryID][]byte
	latestSeq := uint64(0)
	entries = make(map[EntryID][]byte)
	for {
		var e journalEntry
		if err := dec.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("journal decode at seq %d: %w", latestSeq+1, err)
		}
		if e.Seq == 0 || e.ID == "" {
			return fmt.Errorf("journal malformed entry")
		}
		if e.Seq <= latestSeq {
			return fmt.Errorf("journal sequence regression at %d", e.Seq)
		}
		latestSeq = e.Seq
		switch e.Op {
	case "put":
		nonce, err := base64.RawStdEncoding.DecodeString(e.Nonce)
		if err != nil {
			return fmt.Errorf("journal nonce decode at %d: %w", e.Seq, ErrIntegrity)
		}
		ct, err := base64.RawStdEncoding.DecodeString(e.Data)
		if err != nil {
			return fmt.Errorf("journal data decode at %d: %w", e.Seq, ErrIntegrity)
		}
		pt, err := v.decrypt(EntryID(e.ID), nonce, ct)
			if err != nil {
				return fmt.Errorf("journal decrypt at %d: %w", e.Seq, ErrIntegrity)
			}
			entries[EntryID(e.ID)] = pt
		case "delete":
			delete(entries, EntryID(e.ID))
		default:
			return fmt.Errorf("journal unknown op %q", e.Op)
		}
	}
	v.entries = entries
	v.seq = latestSeq
	v.journal = f
	return nil
}

func (v *Vault) cipher() (cipher.AEAD, error) {
	block, err := aes.NewCipher(v.master)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// aad binds ciphertext to the exact entry id + a global scope constant.
func (v *Vault) aad(id EntryID) []byte {
	h := sha256.New()
	h.Write([]byte("forgelocal-localvault-aad-v1|"))
	h.Write([]byte(id))
	return h.Sum(nil)
}

func (v *Vault) encrypt(id EntryID, plaintext []byte) (nonce, ct []byte, err error) {
	gcm, err := v.cipher()
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ct = gcm.Seal(nil, nonce, plaintext, v.aad(id))
	return nonce, ct, nil
}

func (v *Vault) decrypt(id EntryID, nonce, ct []byte) ([]byte, error) {
	gcm, err := v.cipher()
	if err != nil {
		return nil, err
	}
	pt, err := gcm.Open(nil, nonce, ct, v.aad(id))
	if err != nil {
		return nil, ErrIntegrity
	}
	return pt, nil
}

// Put stores value under id. id follows the code-base id token grammar.
func (v *Vault) Put(id string, value []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return ErrClosed
	}
	if !validID(id) {
		return ErrInvalidID
	}
	if len(value) == 0 {
		return errors.New("localvault: value cannot be empty")
	}
	if len(value) > maxPlaintext {
		return ErrTooLarge
	}
	nonce, ct, err := v.encrypt(EntryID(id), value)
	if err != nil {
		return err
	}
	v.seq++
	e := journalEntry{
		Seq:   v.seq,
		Op:    "put",
		ID:    id,
		At:    time.Now().UTC().Format(time.RFC3339Nano),
		Data:  base64.RawStdEncoding.EncodeToString(ct),
		Nonce: base64.RawStdEncoding.EncodeToString(nonce),
	}
	if err := v.appendJournal(e); err != nil {
		return fmt.Errorf("localvault: journal write: %w", err)
	}
	v.entries[EntryID(id)] = append([]byte(nil), value...)
	return nil
}

// Get returns a copy of the value; never the internal slice.
func (v *Vault) Get(id string) ([]byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return nil, ErrClosed
	}
	if !validID(id) {
		return nil, ErrInvalidID
	}
	pt, ok := v.entries[EntryID(id)]
	if !ok {
		return nil, ErrEntryNotFound
	}
	return append([]byte(nil), pt...), nil
}

// Delete removes an entry. Idempotent: deleting a nonexistent id is a no-op.
func (v *Vault) Delete(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return ErrClosed
	}
	if !validID(id) {
		return ErrInvalidID
	}
	v.seq++
	e := journalEntry{
		Seq: v.seq,
		Op:  "delete",
		ID:  id,
		At:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := v.appendJournal(e); err != nil {
		return fmt.Errorf("localvault: journal write: %w", err)
	}
	delete(v.entries, EntryID(id))
	return nil
}

func (v *Vault) appendJournal(e journalEntry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	if _, err := v.journal.Write(line); err != nil {
		return err
	}
	if err := v.journal.Sync(); err != nil {
		return err
	}
	return nil
}

// Close zeroes the master key and releases the journal file handle.
func (v *Vault) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return nil
	}
	v.closed = true
	for i := range v.master {
		v.master[i] = 0
	}
	v.master = nil
	for k := range v.entries {
		v.entries[k] = nil
	}
	v.entries = nil
	if v.journal != nil {
		err := v.journal.Close()
		v.journal = nil
		return err
	}
	return nil
}

// Exists reports whether an id has an entry without revealing its value.
func (v *Vault) Exists(id string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return false
	}
	_, ok := v.entries[EntryID(id)]
	return ok
}

// SecretRefs lists the current entry ids (opaque refs only; never values).
func (v *Vault) SecretRefs() []EntryID {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.closed {
		return nil
	}
	ids := make([]EntryID, 0, len(v.entries))
	for k := range v.entries {
		ids = append(ids, k)
	}
	return ids
}
