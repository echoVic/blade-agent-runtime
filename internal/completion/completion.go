package completion

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user/blade-agent-runtime/internal/core/ledger"
	"github.com/user/blade-agent-runtime/internal/core/task"
)

type Completion struct {
	Value       string
	Description string
}

func GetTaskCompletions(barDir string, includeAll bool) []Completion {
	tm := task.NewManager("", barDir)
	tasks, err := tm.List()
	if err != nil {
		return nil
	}

	var completions []Completion
	for _, t := range tasks {
		if !includeAll && t.Status == task.TaskStatusClosed {
			continue
		}

		desc := fmt.Sprintf("%s (%s)", t.Name, t.Status)
		completions = append(completions, Completion{
			Value:       t.ID,
			Description: desc,
		})

		completions = append(completions, Completion{
			Value:       t.Name,
			Description: fmt.Sprintf("ID: %s", t.ID),
		})
	}

	return completions
}

func GetStepCompletions(barDir, taskID string) []Completion {
	taskDir := filepath.Join(barDir, "tasks", taskID)
	lm := ledger.NewManager(taskDir)

	steps, err := lm.List()
	if err != nil {
		return nil
	}

	var completions []Completion
	for _, s := range steps {
		desc := string(s.Kind)
		if len(s.Cmd) > 0 {
			cmdStr := strings.Join(s.Cmd, " ")
			if len(cmdStr) > 30 {
				cmdStr = cmdStr[:27] + "..."
			}
			desc = fmt.Sprintf("%s: %s", s.Kind, cmdStr)
		}

		completions = append(completions, Completion{
			Value:       s.StepID,
			Description: desc,
		})
	}

	return completions
}

func ToCobraCompletions(completions []Completion) []string {
	result := make([]string, len(completions))
	for i, c := range completions {
		if c.Description != "" {
			result[i] = fmt.Sprintf("%s\t%s", c.Value, c.Description)
		} else {
			result[i] = c.Value
		}
	}
	return result
}
