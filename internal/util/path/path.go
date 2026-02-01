package path

import (
	"errors"
	"os"
	"path/filepath"
)

func FindRepoRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("not a git repository")
		}
		current = parent
	}
}

func BarDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".bar")
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
