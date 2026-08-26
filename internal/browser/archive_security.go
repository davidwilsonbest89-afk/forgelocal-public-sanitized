package browser

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxArchiveFiles      = 4096
	maxArchiveFileBytes  = 1 << 30
	maxArchiveTotalBytes = 2 << 30
	maxArchivePathDepth  = 32
)

func secureExtractArchive(file, destDir string) error {
	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(parent, ".forgelocal-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)

	switch {
	case strings.HasSuffix(file, ".zip"):
		err = secureExtractZip(file, stage)
	case strings.HasSuffix(file, ".tar.gz"):
		err = secureExtractTarGz(file, stage)
	default:
		err = fmt.Errorf("unknown archive format: %s", file)
	}
	if err != nil {
		return err
	}
	return replaceDirectoryWithStage(stage, destDir)
}

func replaceDirectoryWithStage(stage, destDir string) error {
	backup := ""
	if _, err := os.Stat(destDir); err == nil {
		backup = destDir + ".previous"
		_ = os.RemoveAll(backup)
		if err := os.Rename(destDir, backup); err != nil {
			return fmt.Errorf("stage existing runtime: %w", err)
		}
	}
	if err := os.Rename(stage, destDir); err != nil {
		if backup != "" {
			_ = os.Rename(backup, destDir)
		}
		return fmt.Errorf("activate extracted runtime: %w", err)
	}
	if backup != "" {
		if err := os.RemoveAll(backup); err != nil {
			return fmt.Errorf("remove previous runtime after activation: %w", err)
		}
	}
	return nil
}

func secureArchiveTarget(root, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("archive entry has empty name")
	}
	// Archive names use slash, but treating backslash as a separator too avoids
	// platform-dependent traversal when a Windows archive is processed elsewhere.
	name = strings.ReplaceAll(name, "\\", "/")
	clean := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes extraction root: %q", name)
	}
	depth := len(strings.Split(filepath.ToSlash(clean), "/"))
	if depth > maxArchivePathDepth {
		return "", fmt.Errorf("archive entry exceeds path depth limit: %q", name)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes extraction root: %q", name)
	}
	return targetAbs, nil
}

func rejectArchiveSymlink(mode os.FileMode) error {
	if mode&os.ModeSymlink != 0 {
		return fmt.Errorf("archive symlink entries are not allowed")
	}
	if mode&os.ModeType != 0 && mode&os.ModeDir == 0 {
		return fmt.Errorf("archive special file entries are not allowed")
	}
	return nil
}

func secureExtractZip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	if len(r.File) > maxArchiveFiles {
		return fmt.Errorf("archive contains too many files: %d", len(r.File))
	}
	var total uint64
	for _, f := range r.File {
		target, err := secureArchiveTarget(destDir, f.Name)
		if err != nil {
			return err
		}
		mode := f.FileInfo().Mode()
		if err := rejectArchiveSymlink(mode); err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
			continue
		}
		if f.UncompressedSize64 > maxArchiveFileBytes || total > maxArchiveTotalBytes-f.UncompressedSize64 {
			return fmt.Errorf("archive expanded size exceeds limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_EXCL, normalizedArchiveMode(mode))
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			_ = out.Close()
			return err
		}
		err = copyArchiveBytes(out, rc, f.UncompressedSize64)
		closeErr := rc.Close()
		outErr := out.Close()
		if err == nil {
			err = closeErr
		}
		if err == nil {
			err = outErr
		}
		if err != nil {
			return err
		}
		total += f.UncompressedSize64
	}
	return nil
}

func secureExtractTarGz(tarFile, destDir string) error {
	f, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	var files, total uint64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		files++
		if files > maxArchiveFiles {
			return fmt.Errorf("archive contains too many files: %d", files)
		}
		target, err := secureArchiveTarget(destDir, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			size, err := checkedArchiveSize(hdr.Size)
			if err != nil || size > maxArchiveFileBytes || total > maxArchiveTotalBytes-size {
				return fmt.Errorf("archive expanded size exceeds limit")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_EXCL, normalizedArchiveMode(os.FileMode(hdr.Mode)))
			if err != nil {
				return err
			}
			err = copyArchiveBytes(out, tr, size)
			closeErr := out.Close()
			if err == nil {
				err = closeErr
			}
			if err != nil {
				return err
			}
			total += size
		default:
			return fmt.Errorf("archive entry type is not allowed: %d", hdr.Typeflag)
		}
	}
}

func normalizedArchiveMode(mode os.FileMode) os.FileMode {
	mode &= 0700
	if mode == 0 {
		mode = 0600
	}
	return mode
}

func checkedArchiveSize(size int64) (uint64, error) {
	if size < 0 {
		return 0, fmt.Errorf("negative archive entry size")
	}
	return uint64(size), nil
}

func copyArchiveBytes(dst io.Writer, src io.Reader, size uint64) error {
	if size > maxArchiveFileBytes || size > uint64(1<<63-1) {
		return fmt.Errorf("archive entry exceeds file size limit")
	}
	n, err := io.CopyN(dst, src, int64(size))
	if err != nil {
		return err
	}
	if uint64(n) != size {
		return fmt.Errorf("archive entry size mismatch")
	}
	return nil
}
