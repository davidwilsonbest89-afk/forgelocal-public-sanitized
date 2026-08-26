package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func openBrowserRegularFile(path string, flags int, mode os.FileMode) (*os.Root, *os.File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}
	parent, base := filepath.Split(filepath.Clean(abs))
	root, err := os.OpenRoot(filepath.Clean(parent))
	if err != nil {
		return nil, nil, err
	}
	if info, statErr := root.Lstat(base); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			_ = root.Close()
			return nil, nil, fmt.Errorf("browser file is not a regular non-symlink file: %s", path)
		}
	} else if !os.IsNotExist(statErr) {
		_ = root.Close()
		return nil, nil, statErr
	}
	var file *os.File
	if flags == 0 {
		file, err = root.Open(base)
	} else {
		file, err = root.OpenFile(base, flags, mode)
	}
	if err != nil {
		_ = root.Close()
		return nil, nil, err
	}
	return root, file, nil
}

func openBrowserArchiveOutput(root *os.Root, rootDir, target string, flags int, mode os.FileMode) (*os.File, error) {
	rel, err := filepath.Rel(rootDir, target)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) || rel == "" || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("archive output escapes root: %s", target)
	}
	return root.OpenFile(filepath.Clean(rel), flags, mode)
}
