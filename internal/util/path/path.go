package path

import (
	"crypto/sha256"
	"errors"
	"fmt"
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

func GlobalBarDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bar")
}

func ProjectID(repoRoot string) string {
	name := filepath.Base(repoRoot)
	hash := sha256.Sum256([]byte(repoRoot))
	return fmt.Sprintf("%s-%x", name, hash[:2])
}

func BarDir(repoRoot string) string {
	return filepath.Join(GlobalBarDir(), "projects", ProjectID(repoRoot))
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
