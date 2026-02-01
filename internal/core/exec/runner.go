package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

type Runner struct{}

type Options struct {
	Cwd     string
	Env     map[string]string
	Timeout time.Duration
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader
}

type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, cmd []string, opts *Options) (*Result, error) {
	if len(cmd) == 0 {
		return nil, os.ErrInvalid
	}
	if opts == nil {
		opts = &Options{}
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	start := time.Now()
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	c.Dir = opts.Cwd
	env := os.Environ()
	for k, v := range opts.Env {
		env = append(env, k+"="+v)
	}
	c.Env = env
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	if opts.Stdout != nil {
		c.Stdout = io.MultiWriter(opts.Stdout, &stdoutBuf)
	} else {
		c.Stdout = &stdoutBuf
	}
	if opts.Stderr != nil {
		c.Stderr = io.MultiWriter(opts.Stderr, &stderrBuf)
	} else {
		c.Stderr = &stderrBuf
	}
	if opts.Stdin != nil {
		c.Stdin = opts.Stdin
	}
	err := c.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}
	duration := time.Since(start)
	return &Result{
		ExitCode: exitCode,
		Stdout:   stdoutBuf.Bytes(),
		Stderr:   stderrBuf.Bytes(),
		Duration: duration,
	}, nil
}
