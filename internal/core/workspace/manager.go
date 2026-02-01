package workspace

import (
	"path/filepath"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
)

type Manager struct {
	RepoRoot      string
	WorkspacesDir string
	Git           *gitadapter.Runner
}

func NewManager(repoRoot string, workspacesDir string, git *gitadapter.Runner) *Manager {
	return &Manager{
		RepoRoot:      repoRoot,
		WorkspacesDir: workspacesDir,
		Git:           git,
	}
}

func (m *Manager) Create(taskID string, branch string, baseRef string) (string, error) {
	path := filepath.Join(m.WorkspacesDir, taskID)
	_, err := m.Git.Run(m.RepoRoot, "worktree", "add", "-b", branch, path, baseRef)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (m *Manager) Delete(path string) error {
	_, err := m.Git.Run(m.RepoRoot, "worktree", "remove", "--force", path)
	return err
}

func (m *Manager) IsClean(path string) (bool, error) {
	out, err := m.Git.Run(path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

func (m *Manager) Reset(path string, baseRef string, hard bool) error {
	if hard {
		if _, err := m.Git.Run(path, "reset", "--hard", baseRef); err != nil {
			return err
		}
		if _, err := m.Git.Run(path, "clean", "-fd"); err != nil {
			return err
		}
		return nil
	}
	_, err := m.Git.Run(path, "reset", "--hard", baseRef)
	return err
}
