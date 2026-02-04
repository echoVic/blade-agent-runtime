package completion

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/blade-agent-runtime/internal/core/ledger"
	"github.com/user/blade-agent-runtime/internal/core/task"
	utiljson "github.com/user/blade-agent-runtime/internal/util/json"
)

func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "bar-completion-test")
	if err != nil {
		t.Fatal(err)
	}

	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

func createTestTask(t *testing.T, barDir, id, name, status string) {
	t.Helper()
	taskDir := filepath.Join(barDir, "tasks", id)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	tk := &task.Task{
		ID:        id,
		Name:      name,
		Status:    task.TaskStatus(status),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := utiljson.WriteFile(filepath.Join(taskDir, "task.json"), tk); err != nil {
		t.Fatal(err)
	}

	ledgerPath := filepath.Join(taskDir, "ledger.jsonl")
	if err := os.WriteFile(ledgerPath, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
}

func createTestStep(t *testing.T, barDir, taskID, stepID string) {
	t.Helper()
	taskDir := filepath.Join(barDir, "tasks", taskID)

	step := &ledger.Step{
		StepID:    stepID,
		Kind:      ledger.StepKindRun,
		StartedAt: time.Now().UTC(),
		EndedAt:   time.Now().UTC(),
		Cmd:       []string{"echo", "test"},
	}

	lm := ledger.NewManager(taskDir)
	if err := lm.Append(step); err != nil {
		t.Fatal(err)
	}
}

func TestGetTaskCompletions(t *testing.T) {
	barDir, cleanup := setupTestEnv(t)
	defer cleanup()

	createTestTask(t, barDir, "abc123", "fix-bug", "active")
	createTestTask(t, barDir, "def456", "add-feature", "active")
	createTestTask(t, barDir, "ghi789", "old-task", "closed")

	tests := []struct {
		name        string
		includeAll  bool
		wantCount   int
		wantContain []string
		wantExclude []string
	}{
		{
			name:        "active tasks only",
			includeAll:  false,
			wantCount:   4,
			wantContain: []string{"abc123", "def456", "fix-bug", "add-feature"},
			wantExclude: []string{"ghi789", "old-task"},
		},
		{
			name:        "all tasks",
			includeAll:  true,
			wantCount:   6,
			wantContain: []string{"abc123", "def456", "ghi789"},
			wantExclude: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions := GetTaskCompletions(barDir, tt.includeAll)

			if len(completions) != tt.wantCount {
				t.Errorf("got %d completions, want %d", len(completions), tt.wantCount)
			}

			for _, want := range tt.wantContain {
				found := false
				for _, c := range completions {
					if c.Value == want || containsInDescription(c, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("completions should contain %q", want)
				}
			}

			for _, exclude := range tt.wantExclude {
				for _, c := range completions {
					if c.Value == exclude {
						t.Errorf("completions should not contain %q", exclude)
					}
				}
			}
		})
	}
}

func TestGetStepCompletions(t *testing.T) {
	barDir, cleanup := setupTestEnv(t)
	defer cleanup()

	createTestTask(t, barDir, "task1", "test-task", "active")
	createTestStep(t, barDir, "task1", "0001")
	createTestStep(t, barDir, "task1", "0002")
	createTestStep(t, barDir, "task1", "0003")

	completions := GetStepCompletions(barDir, "task1")

	if len(completions) != 3 {
		t.Errorf("got %d completions, want 3", len(completions))
	}

	wantSteps := []string{"0001", "0002", "0003"}
	for _, want := range wantSteps {
		found := false
		for _, c := range completions {
			if c.Value == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("completions should contain step %q", want)
		}
	}
}

func TestGetStepCompletions_NoTask(t *testing.T) {
	barDir, cleanup := setupTestEnv(t)
	defer cleanup()

	completions := GetStepCompletions(barDir, "nonexistent")

	if len(completions) != 0 {
		t.Errorf("got %d completions, want 0 for nonexistent task", len(completions))
	}
}

func TestCompletion_HasDescription(t *testing.T) {
	barDir, cleanup := setupTestEnv(t)
	defer cleanup()

	createTestTask(t, barDir, "abc123", "fix-bug", "active")

	completions := GetTaskCompletions(barDir, false)

	if len(completions) == 0 {
		t.Fatal("expected at least one completion")
	}

	for _, c := range completions {
		if c.Description == "" {
			t.Errorf("completion %q should have a description", c.Value)
		}
	}
}

func containsInDescription(c Completion, s string) bool {
	return len(c.Description) > 0 && (c.Description == s || c.Value == s)
}

func TestToCobraCompletions(t *testing.T) {
	completions := []Completion{
		{Value: "abc123", Description: "fix-bug (active)"},
		{Value: "def456", Description: ""},
	}

	result := ToCobraCompletions(completions)

	if len(result) != 2 {
		t.Errorf("got %d results, want 2", len(result))
	}

	if result[0] != "abc123\tfix-bug (active)" {
		t.Errorf("result[0] = %q, want %q", result[0], "abc123\tfix-bug (active)")
	}

	if result[1] != "def456" {
		t.Errorf("result[1] = %q, want %q", result[1], "def456")
	}
}
