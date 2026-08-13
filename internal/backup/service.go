package backup

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"forgelocal/internal/secrets"
)

var (
	magic        = []byte("FLBK")
	trailer      = []byte("FLEND")
	ErrBusy      = errors.New("profile snapshot already in progress")
	ErrFormat    = errors.New("invalid flbackup format")
	ErrIntegrity = errors.New("backup integrity check failed")
	idPattern    = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
)

const (
	formatVersion  byte   = 1
	maxHeaderSize         = 64 * 1024
	maxPayloadSize uint64 = 512 * 1024 * 1024
)

type Manifest struct {
	Version   int    `json:"version"`
	BackupID  string `json:"backup_id"`
	ProfileID string `json:"profile_id"`
	KeyID     string `json:"key_id"`
	CreatedAt string `json:"created_at"`
}

type Header struct {
	Manifest Manifest `json:"manifest"`
	Nonce    string   `json:"nonce"`
}

type Backup struct {
	ID           string
	ProfileID    string
	ArtifactPath string
	KeyID        string
	SHA256       string
	CreatedAt    time.Time
}

type Restore struct {
	ID              string
	BackupID        string
	SourceProfileID string
	TargetProfileID string
	TargetPath      string
	CorrelationID   string
	CreatedAt       time.Time
}

type ProfileLocks struct {
	mu   sync.Mutex
	held map[string]bool
}

func NewProfileLocks() *ProfileLocks { return &ProfileLocks{held: map[string]bool{}} }
func (l *ProfileLocks) Acquire(id string) (func(), error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.held[id] {
		return nil, ErrBusy
	}
	l.held[id] = true
	return func() { l.mu.Lock(); delete(l.held, id); l.mu.Unlock() }, nil
}

type Store interface {
	BeginBackup(Backup, string) error
	MarkPublished(backupID string) error
	CommitBackup(Backup) error
	GetBackup(id string) (Backup, error)
	HasBackup(id string) (bool, error)
	RecoverBackup(Backup, string) error
	Quarantine(backupID, code string) error
	BeginRestore(Restore) error
	CommitRestore(id string) error
	FailRestore(id, code string) error
}

type Service struct {
	Root             string
	Vault            secrets.Vault
	Store            Store
	Locks            *ProfileLocks
	Now              func() time.Time
	AfterPublishHook func(Backup) error // test-only crash injection; production wiring leaves nil.
}

func (s *Service) Create(profileID, keyID string, payload []byte) (Backup, error) {
	if !validID(profileID) || !validID(keyID) {
		return Backup{}, fmt.Errorf("invalid profile or key id")
	}
	if len(payload) == 0 {
		return Backup{}, fmt.Errorf("empty backup payload")
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.Locks == nil {
		s.Locks = NewProfileLocks()
	}
	release, err := s.Locks.Acquire(profileID)
	if err != nil {
		return Backup{}, err
	}
	defer release()
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		return Backup{}, fmt.Errorf("create backup directory: %w", err)
	}
	key, err := s.Vault.Get(keyID)
	if err != nil {
		return Backup{}, fmt.Errorf("read backup key: %w", err)
	}
	backup := Backup{ID: newID("bkp", s.Now()), ProfileID: profileID, KeyID: keyID, CreatedAt: s.Now()}
	backup.ArtifactPath = filepath.Join(s.Root, backup.ID+".flbackup")
	if err := s.Store.BeginBackup(backup, "backup:"+backup.ID); err != nil {
		return Backup{}, err
	}
	body, err := encode(Manifest{Version: int(formatVersion), BackupID: backup.ID, ProfileID: profileID, KeyID: keyID, CreatedAt: backup.CreatedAt.UTC().Format(time.RFC3339Nano)}, key, payload)
	if err != nil {
		return Backup{}, err
	}
	if err := atomicWrite(backup.ArtifactPath, body); err != nil {
		return Backup{}, err
	}
	sum := sha256.Sum256(body)
	backup.SHA256 = hex.EncodeToString(sum[:])
	if err := s.Store.MarkPublished(backup.ID); err != nil {
		return Backup{}, err
	}
	if s.AfterPublishHook != nil {
		if err := s.AfterPublishHook(backup); err != nil {
			return Backup{}, err
		}
	}
	if err := s.Store.CommitBackup(backup); err != nil {
		return Backup{}, err
	}
	return backup, nil
}

func (s *Service) Restore(backupID, targetProfileID, targetPath string) (Restore, error) {
	if !validID(backupID) || !validID(targetProfileID) {
		return Restore{}, fmt.Errorf("invalid backup or target profile id")
	}
	backup, err := s.Store.GetBackup(backupID)
	if err != nil {
		return Restore{}, err
	}
	if backup.ProfileID == targetProfileID {
		return Restore{}, fmt.Errorf("restore target must use a new profile id")
	}
	rootAbs, err := filepath.Abs(s.Root)
	if err != nil {
		return Restore{}, err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return Restore{}, err
	}
	if targetAbs == rootAbs || strings.HasPrefix(targetAbs+string(os.PathSeparator), rootAbs+string(os.PathSeparator)) {
		return Restore{}, fmt.Errorf("restore target cannot be inside backup directory")
	}
	if _, err := os.Stat(targetAbs); err == nil {
		return Restore{}, fmt.Errorf("restore target already exists")
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	key, err := s.Vault.Get(backup.KeyID)
	if err != nil {
		return Restore{}, fmt.Errorf("read restore key: %w", err)
	}
	body, err := os.ReadFile(backup.ArtifactPath)
	if err != nil {
		return Restore{}, err
	}
	manifest, payload, err := decode(body, key)
	if err != nil {
		return Restore{}, err
	}
	if manifest.BackupID != backup.ID || manifest.ProfileID != backup.ProfileID || manifest.KeyID != backup.KeyID {
		return Restore{}, ErrIntegrity
	}
	restore := Restore{ID: newID("rst", s.Now()), BackupID: backup.ID, SourceProfileID: backup.ProfileID, TargetProfileID: targetProfileID, TargetPath: targetAbs, CorrelationID: "restore:" + backup.ID, CreatedAt: s.Now()}
	if err := s.Store.BeginRestore(restore); err != nil {
		return Restore{}, err
	}
	if err := os.MkdirAll(targetAbs, 0700); err != nil {
		_ = s.Store.FailRestore(restore.ID, "TARGET_CREATE_FAILED")
		return Restore{}, err
	}
	if err := atomicWrite(filepath.Join(targetAbs, "payload.bin"), payload); err != nil {
		_ = os.RemoveAll(targetAbs)
		_ = s.Store.FailRestore(restore.ID, "PAYLOAD_WRITE_FAILED")
		return Restore{}, err
	}
	if err := s.Store.CommitRestore(restore.ID); err != nil {
		_ = os.RemoveAll(targetAbs)
		return Restore{}, err
	}
	return restore, nil
}

// Reconcile repairs artifacts left after a crash between rename and SQLite commit.
func (s *Service) Reconcile() (recovered, quarantined int, err error) {
	entries, err := os.ReadDir(s.Root)
	if errors.Is(err, os.ErrNotExist) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".flbackup") {
			continue
		}
		path := filepath.Join(s.Root, entry.Name())
		// #nosec G304 -- entry.Name() originates from os.ReadDir(s.Root), has no caller-controlled directory component, and is never accepted from HTTP input.
		body, e := os.ReadFile(path)
		if e != nil {
			return recovered, quarantined, e
		}
		manifest, e := parseHeader(body)
		if e != nil || manifest.BackupID+".flbackup" != entry.Name() {
			if e = s.quarantine(path, entry.Name(), "INVALID_FORMAT"); e != nil {
				return recovered, quarantined, e
			}
			quarantined++
			continue
		}
		exists, e := s.Store.HasBackup(manifest.BackupID)
		if e != nil {
			return recovered, quarantined, e
		}
		if exists {
			continue
		}
		key, e := s.Vault.Get(manifest.KeyID)
		if e != nil {
			if e = s.quarantine(path, manifest.BackupID, "KEY_UNAVAILABLE"); e != nil {
				return recovered, quarantined, e
			}
			quarantined++
			continue
		}
		if _, _, e = decode(body, key); e != nil {
			if e = s.quarantine(path, manifest.BackupID, "INTEGRITY_FAILED"); e != nil {
				return recovered, quarantined, e
			}
			quarantined++
			continue
		}
		sum := sha256.Sum256(body)
		created, e := time.Parse(time.RFC3339Nano, manifest.CreatedAt)
		if e != nil {
			created = time.Now().UTC()
		}
		if e = s.Store.RecoverBackup(Backup{ID: manifest.BackupID, ProfileID: manifest.ProfileID, ArtifactPath: path, KeyID: manifest.KeyID, SHA256: hex.EncodeToString(sum[:]), CreatedAt: created}, "RECOVERED_PUBLISHED_UNREGISTERED"); e != nil {
			return recovered, quarantined, e
		}
		recovered++
	}
	return recovered, quarantined, nil
}

func (s *Service) quarantine(path, id, code string) error {
	dir := filepath.Join(s.Root, "quarantine")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	target := filepath.Join(dir, filepath.Base(path)+".invalid")
	if err := os.Rename(path, target); err != nil {
		return err
	}
	return s.Store.Quarantine(strings.TrimSuffix(id, ".flbackup"), code)
}

func encode(m Manifest, key, payload []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	header := Header{Manifest: m, Nonce: base64Raw(nonce)}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	if len(headerBytes) == 0 || len(headerBytes) > maxHeaderSize {
		return nil, fmt.Errorf("backup header exceeds %d bytes", maxHeaderSize)
	}
	// #nosec G407 -- nonce is freshly generated with crypto/rand immediately above and is unique per artifact.
	ciphertext := gcm.Seal(nil, nonce, payload, headerBytes)
	if uint64(len(ciphertext)) > maxPayloadSize {
		return nil, fmt.Errorf("backup payload exceeds %d bytes", maxPayloadSize)
	}
	var out bytes.Buffer
	out.Write(magic)
	out.WriteByte(formatVersion)
	// #nosec G115 -- header length was bounded by maxHeaderSize before this conversion.
	headerLen := uint32(len(headerBytes))
	if err := binary.Write(&out, binary.BigEndian, headerLen); err != nil {
		return nil, err
	}
	out.Write(headerBytes)
	if err := binary.Write(&out, binary.BigEndian, uint64(len(ciphertext))); err != nil {
		return nil, err
	}
	out.Write(ciphertext)
	out.Write(trailer)
	return out.Bytes(), nil
}

func decode(body, key []byte) (Manifest, []byte, error) {
	header, ciphertext, err := split(body)
	if err != nil {
		return Manifest{}, nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return Manifest{}, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Manifest{}, nil, err
	}
	nonce, err := base64RawDecode(header.Nonce)
	if err != nil {
		return Manifest{}, nil, ErrFormat
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, mustJSON(header))
	if err != nil {
		return Manifest{}, nil, ErrIntegrity
	}
	return header.Manifest, plain, nil
}

func parseHeader(body []byte) (Manifest, error) { h, _, e := split(body); return h.Manifest, e }
func split(body []byte) (Header, []byte, error) {
	if len(body) < 4+1+4+8+len(trailer) || !bytes.Equal(body[:4], magic) || body[4] != formatVersion || !bytes.Equal(body[len(body)-len(trailer):], trailer) {
		return Header{}, nil, ErrFormat
	}
	off := 5
	headerLen := int(binary.BigEndian.Uint32(body[off : off+4]))
	off += 4
	if headerLen <= 0 || headerLen > maxHeaderSize || off+headerLen+8+len(trailer) > len(body) {
		return Header{}, nil, ErrFormat
	}
	headerBytes := body[off : off+headerLen]
	off += headerLen
	rawCipherLen := binary.BigEndian.Uint64(body[off : off+8])
	off += 8
	remaining := len(body) - off - len(trailer)
	if remaining <= 0 || rawCipherLen == 0 || rawCipherLen > maxPayloadSize || rawCipherLen > uint64(remaining) {
		return Header{}, nil, ErrFormat
	}
	// #nosec G115 -- rawCipherLen was bounded by maxPayloadSize and remaining bytes above.
	cipherLen := int(rawCipherLen)
	if off+cipherLen+len(trailer) != len(body) {
		return Header{}, nil, ErrFormat
	}
	var h Header
	if err := json.Unmarshal(headerBytes, &h); err != nil || h.Manifest.Version != int(formatVersion) || !validID(h.Manifest.BackupID) || !validID(h.Manifest.ProfileID) || !validID(h.Manifest.KeyID) {
		return Header{}, nil, ErrFormat
	}
	return h, body[off : off+cipherLen], nil
}
func mustJSON(h Header) []byte                 { b, _ := json.Marshal(h); return b }
func base64Raw(b []byte) string                { return base64.RawURLEncoding.EncodeToString(b) }
func base64RawDecode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func newID(prefix string, now time.Time) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%d-%x", prefix, now.UnixNano(), b)
}
func validID(v string) bool { return idPattern.MatchString(v) }

func atomicWrite(path string, body []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".flbk-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	// #nosec G304 -- dir is filepath.Dir of the internally generated staging path and is not user-controlled.
	d, err := os.Open(dir)

	if err == nil {
		defer d.Close()
		if err := d.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// BackupIDs returns a deterministic list useful for diagnostics without exposing secrets.
func BackupIDs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".flbackup") {
			ids = append(ids, strings.TrimSuffix(e.Name(), ".flbackup"))
		}
	}
	sort.Strings(ids)
	return ids, nil
}

var _ = io.EOF
