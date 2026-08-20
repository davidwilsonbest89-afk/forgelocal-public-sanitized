package profile

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	snapshotFormatVersion = 1
	maxSnapshotFiles      = 10_000
	maxSnapshotFileSize   = int64(256 * 1024 * 1024)
	maxSnapshotTotalSize  = int64(480 * 1024 * 1024)
)

var (
	snapshotIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	ErrUnsafeSnapshot = errors.New("unsafe profile snapshot path or entry")
)

type snapshotMetadata struct {
	Version int      `json:"version"`
	Profile *Profile `json:"profile"`
}

// CreateBackupSnapshot creates a bounded tar payload of profile metadata and browser-data.
// It intentionally excludes downloads, artifacts, runtime sockets, and all proxy credentials.
func (s *Store) CreateBackupSnapshot(id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	if !snapshotIDPattern.MatchString(id) {
		return nil, ErrUnsafeSnapshot
	}
	profileRoot, err := s.profileRootLocked(p)
	if err != nil {
		return nil, err
	}
	browserRoot := filepath.Join(profileRoot, "browser-data")
	if err := assertSafeDirectory(browserRoot); err != nil {
		return nil, err
	}

	clean := redactForBackup(p)
	metadata, err := json.Marshal(snapshotMetadata{Version: snapshotFormatVersion, Profile: clean})
	if err != nil {
		return nil, err
	}
	if int64(len(metadata)) > maxSnapshotFileSize {
		return nil, fmt.Errorf("snapshot metadata exceeds limit")
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := writeTarFile(tw, "metadata/profile.json", metadata, 0600, time.Now().UTC()); err != nil {
		return nil, err
	}
	var files int
	var total int64 = int64(len(metadata))
	err = filepath.WalkDir(browserRoot, func(full string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(browserRoot, full)
		if err != nil || rel == "." {
			return err
		}
		archiveName := path.Join("browser-data", filepath.ToSlash(rel))
		if !safeArchiveName(archiveName) {
			return ErrUnsafeSnapshot
		}
		info, err := os.Lstat(full)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symbolic link %s", ErrUnsafeSnapshot, full)
		}
		files++
		if files > maxSnapshotFiles {
			return fmt.Errorf("snapshot contains too many files")
		}
		if entry.IsDir() {
			return tw.WriteHeader(&tar.Header{Name: archiveName + "/", Mode: 0700, Typeflag: tar.TypeDir, ModTime: info.ModTime().UTC()})
		}
		if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxSnapshotFileSize {
			return fmt.Errorf("%w: unsupported or oversized file %s", ErrUnsafeSnapshot, full)
		}
		if total > maxSnapshotTotalSize-info.Size() {
			return fmt.Errorf("snapshot exceeds total size limit")
		}
		data, err := readVerifiedRegularFile(full, info)
		if err != nil {
			return err
		}
		if err := writeTarFile(tw, archiveName, data, 0600, info.ModTime().UTC()); err != nil {
			return err
		}
		total += int64(len(data))
		return nil
	})
	if err != nil {
		_ = tw.Close()
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RestoreBackupSnapshot validates a tar payload, creates a distinct profile, then materializes browser-data beneath that profile only.
func (s *Store) RestoreBackupSnapshot(targetID string, payload []byte) (*Profile, error) {
	if !snapshotIDPattern.MatchString(targetID) || len(payload) == 0 || int64(len(payload)) > maxSnapshotTotalSize {
		return nil, ErrUnsafeSnapshot
	}
	meta, files, err := parseSnapshot(payload)
	if err != nil {
		return nil, err
	}
	if meta.Version != snapshotFormatVersion || meta.Profile == nil || !snapshotIDPattern.MatchString(meta.Profile.ID) {
		return nil, ErrUnsafeSnapshot
	}
	if meta.Profile.ID == targetID {
		return nil, fmt.Errorf("restore target must use a new profile id")
	}
	p := *meta.Profile
	p.ID = targetID
	p.Name = meta.Profile.Name + " (restored)"
	p.ContainerID = ""
	p.CreatedAt = time.Time{}
	p.LastUsed = time.Time{}
	if err := s.Create(&p); err != nil {
		return nil, err
	}
	// s.Create created the target directory. A failure after this point must remove the new profile atomically from the store.
	browserRoot := filepath.Join(p.ProfileDir, "browser-data")
	root, err := os.OpenRoot(browserRoot)
	if err != nil {
		_ = s.Delete(p.ID)
		return nil, err
	}
	defer root.Close()
	for name, data := range files {
		rel := strings.TrimPrefix(name, "browser-data/")
		if rel == "" || !safeArchiveName(rel) {
			_ = s.Delete(p.ID)
			return nil, ErrUnsafeSnapshot
		}
		parent := path.Dir(rel)
		if parent != "." {
			if err := root.MkdirAll(parent, 0700); err != nil {
				_ = s.Delete(p.ID)
				return nil, err
			}
		}
		f, err := root.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			_ = s.Delete(p.ID)
			return nil, err
		}
		if _, err := f.Write(data); err != nil {
			_ = f.Close()
			_ = s.Delete(p.ID)
			return nil, err
		}
		if err := f.Sync(); err != nil {
			_ = f.Close()
			_ = s.Delete(p.ID)
			return nil, err
		}
		if err := f.Close(); err != nil {
			_ = s.Delete(p.ID)
			return nil, err
		}
	}
	return &p, nil
}

func redactForBackup(src *Profile) *Profile {
	clone := *src
	clone.ContainerID = ""
	if src.Proxy != nil {
		proxy := *src.Proxy
		proxy.Username = ""
		proxy.Password = ""
		clone.Proxy = &proxy
	}
	return &clone
}

func parseSnapshot(payload []byte) (snapshotMetadata, map[string][]byte, error) {
	tr := tar.NewReader(bytes.NewReader(payload))
	var meta snapshotMetadata
	files := make(map[string][]byte)
	var count int
	var total int64
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return snapshotMetadata{}, nil, err
		}
		count++
		if count > maxSnapshotFiles || !safeArchiveName(h.Name) || h.Size < 0 || h.Size > maxSnapshotFileSize || total > maxSnapshotTotalSize-h.Size {
			return snapshotMetadata{}, nil, ErrUnsafeSnapshot
		}
		if h.Typeflag == tar.TypeDir {
			if !strings.HasPrefix(h.Name, "browser-data/") {
				return snapshotMetadata{}, nil, ErrUnsafeSnapshot
			}
			continue
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return snapshotMetadata{}, nil, ErrUnsafeSnapshot
		}
		data, err := io.ReadAll(io.LimitReader(tr, h.Size+1))
		if err != nil || int64(len(data)) != h.Size {
			return snapshotMetadata{}, nil, ErrUnsafeSnapshot
		}
		total += int64(len(data))
		switch h.Name {
		case "metadata/profile.json":
			if meta.Profile != nil || json.Unmarshal(data, &meta) != nil {
				return snapshotMetadata{}, nil, ErrUnsafeSnapshot
			}
		default:
			if !strings.HasPrefix(h.Name, "browser-data/") || strings.HasSuffix(h.Name, "/") {
				return snapshotMetadata{}, nil, ErrUnsafeSnapshot
			}
			files[h.Name] = data
		}
	}
	if meta.Profile == nil {
		return snapshotMetadata{}, nil, ErrUnsafeSnapshot
	}
	return meta, files, nil
}

func writeTarFile(tw *tar.Writer, name string, data []byte, mode int64, modified time.Time) error {
	if !safeArchiveName(name) {
		return ErrUnsafeSnapshot
	}
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(data)), Typeflag: tar.TypeReg, ModTime: modified}); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func safeArchiveName(name string) bool {
	clean := path.Clean(name)
	return name != "" && clean == name && !strings.HasPrefix(name, "/") && !strings.Contains(name, "\\") && !strings.HasPrefix(clean, "../") && clean != ".."
}

func (s *Store) profileRootLocked(p *Profile) (string, error) {
	base, err := filepath.Abs(s.dir)
	if err != nil {
		return "", err
	}
	expected := filepath.Join(base, p.ID)
	actual, err := filepath.Abs(p.ProfileDir)
	if err != nil || actual != expected {
		return "", ErrUnsafeSnapshot
	}
	return expected, nil
}

func assertSafeDirectory(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrUnsafeSnapshot
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if resolved != root {
		return ErrUnsafeSnapshot
	}
	return nil
}

func readVerifiedRegularFile(name string, before fs.FileInfo) ([]byte, error) {
	// #nosec G304 -- name was lstat'ed under a validated profile root and SameFile is checked before reading.
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, ErrUnsafeSnapshot
	}
	data, err := io.ReadAll(io.LimitReader(f, maxSnapshotFileSize+1))
	if err != nil || int64(len(data)) > maxSnapshotFileSize {
		return nil, ErrUnsafeSnapshot
	}
	return data, nil
}
