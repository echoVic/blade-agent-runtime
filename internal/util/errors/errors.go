package errors

import (
	"fmt"
	"strings"
)

type BarError struct {
	Code    ErrorCode
	Message string
	Hint    string
	Cause   error
}

type ErrorCode string

const (
	ErrNotInitialized    ErrorCode = "NOT_INITIALIZED"
	ErrNoActiveTask      ErrorCode = "NO_ACTIVE_TASK"
	ErrTaskNotFound      ErrorCode = "TASK_NOT_FOUND"
	ErrStepNotFound      ErrorCode = "STEP_NOT_FOUND"
	ErrPatchNotFound     ErrorCode = "PATCH_NOT_FOUND"
	ErrWorkspaceNotClean ErrorCode = "WORKSPACE_NOT_CLEAN"
	ErrPolicyViolation   ErrorCode = "POLICY_VIOLATION"
	ErrNotGitRepo        ErrorCode = "NOT_GIT_REPO"
	ErrGitOperation      ErrorCode = "GIT_OPERATION"
	ErrCommandFailed     ErrorCode = "COMMAND_FAILED"
	ErrRollbackFailed    ErrorCode = "ROLLBACK_FAILED"
	ErrUpdateFailed      ErrorCode = "UPDATE_FAILED"
)

func (e *BarError) Error() string {
	var sb strings.Builder
	sb.WriteString("❌ ")
	sb.WriteString(e.Message)
	if e.Hint != "" {
		sb.WriteString("\n\n💡 ")
		sb.WriteString(e.Hint)
	}
	return sb.String()
}

func (e *BarError) Unwrap() error {
	return e.Cause
}

func NotInitialized() *BarError {
	return &BarError{
		Code:    ErrNotInitialized,
		Message: "BAR is not initialized in this repository.",
		Hint:    "Run 'bar init' or 'bar task start <name>' to get started.",
	}
}

func NoActiveTask() *BarError {
	return &BarError{
		Code:    ErrNoActiveTask,
		Message: "No active task.",
		Hint:    "Run 'bar task start <name>' to create a new task,\n   or 'bar task list' to see existing tasks,\n   or 'bar task switch <id>' to switch to an existing task.",
	}
}

func TaskNotFound(nameOrID string) *BarError {
	return &BarError{
		Code:    ErrTaskNotFound,
		Message: fmt.Sprintf("Task '%s' not found.", nameOrID),
		Hint:    "Run 'bar task list' to see available tasks.",
	}
}

func StepNotFound(stepID string) *BarError {
	return &BarError{
		Code:    ErrStepNotFound,
		Message: fmt.Sprintf("Step '%s' not found.", stepID),
		Hint:    "Run 'bar log' to see available steps.",
	}
}

func PatchNotFound(stepID string) *BarError {
	return &BarError{
		Code:    ErrPatchNotFound,
		Message: fmt.Sprintf("Patch for step '%s' not found.", stepID),
		Hint:    "The patch file may have been deleted or corrupted.",
	}
}

func WorkspaceNotClean() *BarError {
	return &BarError{
		Code:    ErrWorkspaceNotClean,
		Message: "Workspace has uncommitted changes.",
		Hint:    "Use '--force' to discard changes and proceed,\n   or 'bar apply' to commit changes first,\n   or 'bar rollback --base' to discard all changes.",
	}
}

func PolicyViolation(rule, reason string) *BarError {
	msg := "Command blocked by policy."
	if rule != "" {
		msg = fmt.Sprintf("Command blocked by policy rule: %s", rule)
	}
	hint := "This command has been blocked for safety reasons."
	if reason != "" {
		hint = fmt.Sprintf("Reason: %s", reason)
	}
	return &BarError{
		Code:    ErrPolicyViolation,
		Message: msg,
		Hint:    hint,
	}
}

func NotGitRepo() *BarError {
	return &BarError{
		Code:    ErrNotGitRepo,
		Message: "Not a git repository.",
		Hint:    "BAR requires a git repository. Run 'git init' first.",
	}
}

func GitOperation(operation string, cause error) *BarError {
	return &BarError{
		Code:    ErrGitOperation,
		Message: fmt.Sprintf("Git operation failed: %s", operation),
		Hint:    "Check your git configuration and try again.",
		Cause:   cause,
	}
}

func CommandFailed(cmd string, cause error) *BarError {
	return &BarError{
		Code:    ErrCommandFailed,
		Message: fmt.Sprintf("Failed to execute command: %s", cmd),
		Hint:    "Check if the command exists and is executable.",
		Cause:   cause,
	}
}

func RollbackNotSupported() *BarError {
	return &BarError{
		Code:    ErrRollbackFailed,
		Message: "Step-level rollback is not supported in v0.",
		Hint:    "Use 'bar rollback --base' to rollback to the initial state.",
	}
}

func RollbackRequiresBase() *BarError {
	return &BarError{
		Code:    ErrRollbackFailed,
		Message: "Rollback target not specified.",
		Hint:    "Use 'bar rollback --base' to rollback to the initial state.",
	}
}

func UpdateFailed(cause error) *BarError {
	return &BarError{
		Code:    ErrUpdateFailed,
		Message: "Failed to update BAR.",
		Hint:    "Try reinstalling manually:\n   curl -fsSL https://echovic.github.io/blade-agent-runtime/install.sh | sh",
		Cause:   cause,
	}
}

func UpdateNotSupportedOnWindows() *BarError {
	return &BarError{
		Code:    ErrUpdateFailed,
		Message: "Auto-update is not supported on Windows.",
		Hint:    "Please reinstall manually:\n   go install github.com/echoVic/blade-agent-runtime/cmd/bar@latest",
	}
}

func Wrap(cause error, message string) *BarError {
	return &BarError{
		Code:    "",
		Message: message,
		Cause:   cause,
	}
}

func WrapWithHint(cause error, message, hint string) *BarError {
	return &BarError{
		Code:    "",
		Message: message,
		Hint:    hint,
		Cause:   cause,
	}
}
