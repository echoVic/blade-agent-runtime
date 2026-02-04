package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestBarError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *BarError
		contains []string
	}{
		{
			name: "message only",
			err: &BarError{
				Code:    ErrNotInitialized,
				Message: "Test message",
			},
			contains: []string{"❌", "Test message"},
		},
		{
			name: "message with hint",
			err: &BarError{
				Code:    ErrNotInitialized,
				Message: "Test message",
				Hint:    "Test hint",
			},
			contains: []string{"❌", "Test message", "💡", "Test hint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Error() = %q, want to contain %q", result, s)
				}
			}
		})
	}
}

func TestBarError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &BarError{
		Code:    ErrGitOperation,
		Message: "Git failed",
		Cause:   cause,
	}

	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}

	errNoCause := &BarError{
		Code:    ErrNotInitialized,
		Message: "No cause",
	}
	if errNoCause.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", errNoCause.Unwrap())
	}
}

func TestNotInitialized(t *testing.T) {
	err := NotInitialized()
	if err.Code != ErrNotInitialized {
		t.Errorf("Code = %v, want %v", err.Code, ErrNotInitialized)
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Error() should contain 'not initialized'")
	}
	if !strings.Contains(err.Error(), "bar init") {
		t.Errorf("Error() should contain hint about 'bar init'")
	}
}

func TestNoActiveTask(t *testing.T) {
	err := NoActiveTask()
	if err.Code != ErrNoActiveTask {
		t.Errorf("Code = %v, want %v", err.Code, ErrNoActiveTask)
	}
	if !strings.Contains(err.Error(), "No active task") {
		t.Errorf("Error() should contain 'No active task'")
	}
	if !strings.Contains(err.Error(), "bar task start") {
		t.Errorf("Error() should contain hint about 'bar task start'")
	}
}

func TestTaskNotFound(t *testing.T) {
	err := TaskNotFound("my-task")
	if err.Code != ErrTaskNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrTaskNotFound)
	}
	if !strings.Contains(err.Error(), "my-task") {
		t.Errorf("Error() should contain task name 'my-task'")
	}
	if !strings.Contains(err.Error(), "bar task list") {
		t.Errorf("Error() should contain hint about 'bar task list'")
	}
}

func TestStepNotFound(t *testing.T) {
	err := StepNotFound("0001")
	if err.Code != ErrStepNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrStepNotFound)
	}
	if !strings.Contains(err.Error(), "0001") {
		t.Errorf("Error() should contain step ID '0001'")
	}
	if !strings.Contains(err.Error(), "bar log") {
		t.Errorf("Error() should contain hint about 'bar log'")
	}
}

func TestPatchNotFound(t *testing.T) {
	err := PatchNotFound("0002")
	if err.Code != ErrPatchNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrPatchNotFound)
	}
	if !strings.Contains(err.Error(), "0002") {
		t.Errorf("Error() should contain step ID '0002'")
	}
}

func TestWorkspaceNotClean(t *testing.T) {
	err := WorkspaceNotClean()
	if err.Code != ErrWorkspaceNotClean {
		t.Errorf("Code = %v, want %v", err.Code, ErrWorkspaceNotClean)
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Errorf("Error() should contain 'uncommitted changes'")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("Error() should contain hint about '--force'")
	}
}

func TestPolicyViolation(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		reason   string
		contains []string
	}{
		{
			name:     "with rule and reason",
			rule:     "no-rm-rf",
			reason:   "Dangerous command",
			contains: []string{"no-rm-rf", "Dangerous command"},
		},
		{
			name:     "without rule and reason",
			rule:     "",
			reason:   "",
			contains: []string{"blocked by policy", "safety reasons"},
		},
		{
			name:     "with rule only",
			rule:     "test-rule",
			reason:   "",
			contains: []string{"test-rule", "safety reasons"},
		},
		{
			name:     "with reason only",
			rule:     "",
			reason:   "Custom reason",
			contains: []string{"blocked by policy", "Custom reason"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PolicyViolation(tt.rule, tt.reason)
			if err.Code != ErrPolicyViolation {
				t.Errorf("Code = %v, want %v", err.Code, ErrPolicyViolation)
			}
			result := err.Error()
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Error() = %q, want to contain %q", result, s)
				}
			}
		})
	}
}

func TestNotGitRepo(t *testing.T) {
	err := NotGitRepo()
	if err.Code != ErrNotGitRepo {
		t.Errorf("Code = %v, want %v", err.Code, ErrNotGitRepo)
	}
	if !strings.Contains(err.Error(), "git repository") {
		t.Errorf("Error() should contain 'git repository'")
	}
	if !strings.Contains(err.Error(), "git init") {
		t.Errorf("Error() should contain hint about 'git init'")
	}
}

func TestGitOperation(t *testing.T) {
	cause := errors.New("permission denied")
	err := GitOperation("checkout", cause)
	if err.Code != ErrGitOperation {
		t.Errorf("Code = %v, want %v", err.Code, ErrGitOperation)
	}
	if !strings.Contains(err.Error(), "checkout") {
		t.Errorf("Error() should contain operation 'checkout'")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestCommandFailed(t *testing.T) {
	cause := errors.New("not found")
	err := CommandFailed("npm install", cause)
	if err.Code != ErrCommandFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrCommandFailed)
	}
	if !strings.Contains(err.Error(), "npm install") {
		t.Errorf("Error() should contain command 'npm install'")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestRollbackNotSupported(t *testing.T) {
	err := RollbackNotSupported()
	if err.Code != ErrRollbackFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrRollbackFailed)
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("Error() should contain 'not supported'")
	}
	if !strings.Contains(err.Error(), "--base") {
		t.Errorf("Error() should contain hint about '--base'")
	}
}

func TestRollbackRequiresBase(t *testing.T) {
	err := RollbackRequiresBase()
	if err.Code != ErrRollbackFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrRollbackFailed)
	}
	if !strings.Contains(err.Error(), "not specified") {
		t.Errorf("Error() should contain 'not specified'")
	}
	if !strings.Contains(err.Error(), "--base") {
		t.Errorf("Error() should contain hint about '--base'")
	}
}

func TestUpdateFailed(t *testing.T) {
	cause := errors.New("network error")
	err := UpdateFailed(cause)
	if err.Code != ErrUpdateFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrUpdateFailed)
	}
	if !strings.Contains(err.Error(), "update") {
		t.Errorf("Error() should contain 'update'")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestUpdateNotSupportedOnWindows(t *testing.T) {
	err := UpdateNotSupportedOnWindows()
	if err.Code != ErrUpdateFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrUpdateFailed)
	}
	if !strings.Contains(err.Error(), "Windows") {
		t.Errorf("Error() should contain 'Windows'")
	}
	if !strings.Contains(err.Error(), "go install") {
		t.Errorf("Error() should contain hint about 'go install'")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("original")
	err := Wrap(cause, "wrapped message")
	if err.Code != "" {
		t.Errorf("Code = %v, want empty", err.Code)
	}
	if !strings.Contains(err.Error(), "wrapped message") {
		t.Errorf("Error() should contain 'wrapped message'")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestWrapWithHint(t *testing.T) {
	cause := errors.New("original")
	err := WrapWithHint(cause, "wrapped message", "helpful hint")
	if err.Code != "" {
		t.Errorf("Code = %v, want empty", err.Code)
	}
	if !strings.Contains(err.Error(), "wrapped message") {
		t.Errorf("Error() should contain 'wrapped message'")
	}
	if !strings.Contains(err.Error(), "helpful hint") {
		t.Errorf("Error() should contain 'helpful hint'")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestErrorsUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := GitOperation("push", cause)

	if !errors.Is(err, cause) {
		t.Errorf("errors.Is() should return true for the cause")
	}

	var barErr *BarError
	if !errors.As(err, &barErr) {
		t.Errorf("errors.As() should work with *BarError")
	}
	if barErr.Code != ErrGitOperation {
		t.Errorf("Code = %v, want %v", barErr.Code, ErrGitOperation)
	}
}
