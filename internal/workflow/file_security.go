package workflow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readWorkflowFile(path string) ([]byte, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	parent, base := filepath.Split(filepath.Clean(abs))
	root, err := os.OpenRoot(filepath.Clean(parent))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(base)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("workflow file is not a regular non-symlink file: %s", path)
	}
	file, err := root.Open(base)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}
