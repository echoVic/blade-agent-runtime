package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/completion"
	"github.com/user/blade-agent-runtime/internal/core/ledger"
	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			clean, err := app.WorkspaceManager.IsClean(task.WorkspacePath)
			if err != nil {
				return err
			}
			taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
			ledgerManager := ledger.NewManager(taskDir)
			steps, err := ledgerManager.List()
			if err != nil {
				return err
			}
			var last *ledger.Step
			if len(steps) > 0 {
				last = steps[len(steps)-1]
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				out := map[string]any{
					"repository":   app.RepoRoot,
					"active_task":  task.ID,
					"task_name":    task.Name,
					"workspace":    task.WorkspacePath,
					"branch":       task.Branch,
					"base":         task.BaseRef,
					"status":       statusString(clean),
					"steps":        len(steps),
					"last_step_id": "",
				}
				if last != nil {
					out["last_step_id"] = last.StepID
				}
				data, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(os.Stdout, string(data))
				return nil
			}
			app.Logger.Info("BAR Status")
			app.Logger.Info("Repository:  %s", app.RepoRoot)
			app.Logger.Info("Active Task: %s (%s)", task.Name, task.ID)
			app.Logger.Info("Workspace:   %s", task.WorkspacePath)
			app.Logger.Info("Branch:      %s", task.Branch)
			app.Logger.Info("Base:        %s", task.BaseRef)
			app.Logger.Info("Status:      %s", statusString(clean))
			app.Logger.Info("Steps:       %d", len(steps))
			if last != nil {
				app.Logger.Info("Last Step:   %s (%s)", last.StepID, last.Kind)
			}
			return nil
		},
	}
	cmd.Flags().String("format", "text", "output format (text/json)")
	return cmd
}
func logCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show task ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
			ledgerManager := ledger.NewManager(taskDir)
			steps, err := ledgerManager.List()
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			limit, _ := cmd.Flags().GetInt("limit")
			stepID, _ := cmd.Flags().GetString("step")
			output, _ := cmd.Flags().GetString("output")
			if stepID != "" {
				step, err := ledgerManager.GetByID(stepID)
				if err != nil {
					return err
				}
				if step == nil {
				return barerrors.StepNotFound(stepID)
			}
				return writeLogOutput(format, output, renderStepDetail(step))
			}
			if limit > 0 && len(steps) > limit {
				steps = steps[len(steps)-limit:]
			}
			if format == "json" {
				data, _ := json.MarshalIndent(steps, "", "  ")
				return writeLogOutput(format, output, string(data))
			}
			if format == "markdown" {
				return writeLogOutput(format, output, renderMarkdown(steps))
			}
			return writeLogOutput(format, output, renderTable(steps))
		},
	}
	cmd.Flags().String("step", "", "show a specific step")
	cmd.Flags().Int("limit", 10, "limit number of steps")
	cmd.Flags().String("format", "table", "output format (table/json/markdown)")
	cmd.Flags().String("output", "", "write output to file")
	_ = cmd.RegisterFlagCompletionFunc("step", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		app, err := initApp(true)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		task, err := requireActiveTask(app)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completions := completion.GetStepCompletions(app.BarDir, task.ID)
		return completion.ToCobraCompletions(completions), cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}
func statusString(clean bool) string {
	if clean {
		return "clean"
	}
	return "dirty"
}

func renderTable(steps []*ledger.Step) string {
	lines := []string{"STEP   KIND      COMMAND                          DURATION  EXIT  FILES"}
	for _, s := range steps {
		cmd := ""
		if len(s.Cmd) > 0 {
			cmd = strings.Join(s.Cmd, " ")
		}
		exit := ""
		if s.ExitCode != nil {
			exit = fmt.Sprintf("%d", *s.ExitCode)
		}
		files := ""
		if s.DiffStat != nil {
			files = fmt.Sprintf("%d (+%d, -%d)", s.DiffStat.Files, s.DiffStat.Additions, s.DiffStat.Deletions)
		}
		lines = append(lines, fmt.Sprintf("%-6s %-9s %-30s %-8s %-4s %s", s.StepID, s.Kind, trim(cmd, 30), formatDuration(s.DurationMs), exit, files))
	}
	return strings.Join(lines, "\n")
}
func renderMarkdown(steps []*ledger.Step) string {
	lines := []string{"| Step | Kind | Command | Duration | Exit | Files |", "|------|------|---------|----------|------|-------|"}
	for _, s := range steps {
		cmd := strings.Join(s.Cmd, " ")
		exit := ""
		if s.ExitCode != nil {
			exit = fmt.Sprintf("%d", *s.ExitCode)
		}
		files := ""
		if s.DiffStat != nil {
			files = fmt.Sprintf("%d (+%d, -%d)", s.DiffStat.Files, s.DiffStat.Additions, s.DiffStat.Deletions)
		}
		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s | %s |", s.StepID, s.Kind, trim(cmd, 40), formatDuration(s.DurationMs), exit, files))
	}
	return strings.Join(lines, "\n")
}
func renderStepDetail(s *ledger.Step) string {
	lines := []string{
		fmt.Sprintf("Step %s", s.StepID),
		"──────────────────────────────",
		fmt.Sprintf("Kind:     %s", s.Kind),
	}
	if len(s.Cmd) > 0 {
		lines = append(lines, fmt.Sprintf("Command:  %s", strings.Join(s.Cmd, " ")))
	}
	lines = append(lines, fmt.Sprintf("Started:  %s", s.StartedAt.Format(time.RFC3339)))
	lines = append(lines, fmt.Sprintf("Ended:    %s", s.EndedAt.Format(time.RFC3339)))
	if s.ExitCode != nil {
		lines = append(lines, fmt.Sprintf("Exit:     %d", *s.ExitCode))
	}
	if s.DiffStat != nil {
		lines = append(lines, fmt.Sprintf("Files:    %d (+%d, -%d)", s.DiffStat.Files, s.DiffStat.Additions, s.DiffStat.Deletions))
	}
	return strings.Join(lines, "\n")
}
func writeLogOutput(format string, output string, content string) error {
	if output == "" {
		fmt.Fprintln(os.Stdout, content)
		return nil
	}
	return os.WriteFile(output, []byte(content+"\n"), 0o644)
}
func formatDuration(ms int64) string {
	if ms == 0 {
		return "-"
	}
	d := time.Duration(ms) * time.Millisecond
	return d.String()
}
