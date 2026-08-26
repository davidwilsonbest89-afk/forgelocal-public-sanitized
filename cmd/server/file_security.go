package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func openServerRegularFile(path string, flags int, mode os.FileMode) (*os.Root, *os.File, error) {
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
			return nil, nil, fmt.Errorf("server file is not a regular non-symlink file: %s", path)
		}
	} else if !os.IsNotExist(statErr) {
		_ = root.Close()
		return nil, nil, statErr
	}
	file, err := root.OpenFile(base, flags, mode)
	if err != nil {
		_ = root.Close()
		return nil, nil, err
	}
	return root, file, nil
}

func readServerRegularFile(path string) ([]byte, error) {
	root, file, err := openServerRegularFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer root.Close()
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
