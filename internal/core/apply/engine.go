package apply

import (
	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
)

type Engine struct {
	Git *gitadapter.Runner
}

func NewEngine(git *gitadapter.Runner) *Engine {
	return &Engine{Git: git}
}

func (e *Engine) Commit(workspacePath string, repoRoot string, baseRef string, message string) (string, error) {
	if message == "" {
		message = "bar: apply changes"
	}
	if _, err := e.Git.Run(workspacePath, "add", "-A"); err != nil {
		return "", err
	}
	if _, err := e.Git.Run(workspacePath, "commit", "-m", message); err != nil {
		return "", err
	}
	sha, err := e.Git.Run(workspacePath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	branch, err := e.Git.Run(workspacePath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	if _, err := e.Git.Run(repoRoot, "fetch", ".", branch+":"+baseRef); err != nil {
		if _, err := e.Git.Run(repoRoot, "checkout", baseRef); err != nil {
			return "", err
		}
		if _, err := e.Git.Run(repoRoot, "cherry-pick", sha); err != nil {
			return "", err
		}
	}
	return sha, nil
}
