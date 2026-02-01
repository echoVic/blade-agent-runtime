package git

import (
	"bytes"
	"os/exec"
	"strings"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if errBuf.Len() > 0 {
			return "", errorWithOutput(err, errBuf.String())
		}
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func errorWithOutput(err error, out string) error {
	return &GitError{Err: err, Output: out}
}

type GitError struct {
	Err    error
	Output string
}

func (e *GitError) Error() string {
	if e.Output == "" {
		return e.Err.Error()
	}
	return e.Err.Error() + ": " + e.Output
}
