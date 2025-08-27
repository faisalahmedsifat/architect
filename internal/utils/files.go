package utils

import (
	"os"
	"path/filepath"
)

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func FindProjectRoot() (string, error) {
	// Look for .architect directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		if FileExists(filepath.Join(current, ".architect")) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			return cwd, nil
		}
		current = parent
	}
}
