package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/completion"
	"github.com/user/blade-agent-runtime/internal/core/ledger"
	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
)

func diffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			stepID, _ := cmd.Flags().GetString("step")
			statOnly, _ := cmd.Flags().GetBool("stat")
			output, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")
			if stepID != "" {
				taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
				ledgerManager := ledger.NewManager(taskDir)
				step, err := ledgerManager.GetByID(stepID)
				if err != nil {
					return err
				}
				if step == nil {
					return barerrors.StepNotFound(stepID)
				}
				if statOnly || format == "stat" {
					if step.DiffStat != nil {
						app.Logger.Info("%d files changed, %d insertions(+), %d deletions(-)", step.DiffStat.Files, step.DiffStat.Additions, step.DiffStat.Deletions)
						return nil
					}
					app.Logger.Info("0 files changed")
					return nil
				}
				if format == "json" {
					return outputDiffJSON(app, step.DiffStat, nil, output)
				}
				if step.Artifacts != nil && step.Artifacts.Patch != "" {
					patchPath := filepath.Join(taskDir, step.Artifacts.Patch)
					data, err := os.ReadFile(patchPath)
					if err != nil {
						return err
					}
					if output != "" {
						return os.WriteFile(output, data, 0o644)
					}
					app.Logger.Info("%s", string(data))
					return nil
				}
				return barerrors.PatchNotFound(stepID)
			}
			result, err := app.DiffEngine.Generate(task.WorkspacePath, task.BaseRef)
			if err != nil {
				return err
			}
			if statOnly || format == "stat" {
				app.Logger.Info("%d files changed, %d insertions(+), %d deletions(-)", result.Files, result.Additions, result.Deletions)
				return nil
			}
			if format == "json" {
				stat := &ledger.DiffStat{
					Files:     result.Files,
					Additions: result.Additions,
					Deletions: result.Deletions,
					FileList:  result.FileList,
				}
				return outputDiffJSON(app, stat, result.Patch, output)
			}
			if output != "" {
				return os.WriteFile(output, result.Patch, 0o644)
			}
			app.Logger.Info("%s", string(result.Patch))
			return nil
		},
	}
	cmd.Flags().String("step", "", "show diff for a specific step")
	cmd.Flags().Bool("stat", false, "show stat only")
	cmd.Flags().String("output", "", "write diff to file")
	cmd.Flags().String("format", "patch", "output format: patch, stat, json")
	_ = cmd.RegisterFlagCompletionFunc("step", stepCompletionFunc)
	return cmd
}

func stepCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
}

func outputDiffJSON(app *App, stat *ledger.DiffStat, patch []byte, output string) error {
	type jsonOutput struct {
		Files     int      `json:"files"`
		Additions int      `json:"additions"`
		Deletions int      `json:"deletions"`
		FileList  []string `json:"file_list,omitempty"`
		Patch     string   `json:"patch,omitempty"`
	}
	out := jsonOutput{}
	if stat != nil {
		out.Files = stat.Files
		out.Additions = stat.Additions
		out.Deletions = stat.Deletions
		out.FileList = stat.FileList
	}
	if patch != nil {
		out.Patch = string(patch)
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if output != "" {
		return os.WriteFile(output, data, 0o644)
	}
	app.Logger.Info("%s", string(data))
	return nil
}

func applyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply changes to base branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			message, _ := cmd.Flags().GetString("message")
			noClose, _ := cmd.Flags().GetBool("no-close")
			sha, err := app.ApplyEngine.Commit(task.WorkspacePath, app.RepoRoot, task.BaseRef, message)
			if err != nil {
				return err
			}
			taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
			ledgerManager := ledger.NewManager(taskDir)
			stepID, err := ledgerManager.NextStepID()
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			step := &ledger.Step{
				StepID:        stepID,
				Kind:          ledger.StepKindApply,
				StartedAt:     now,
				EndedAt:       now,
				Mode:          "commit",
				CommitSHA:     sha,
				CommitMessage: message,
				TargetBranch:  task.BaseRef,
			}
			if err := ledgerManager.Append(step); err != nil {
				return err
			}
			if !noClose {
				if err := app.WorkspaceManager.Delete(task.WorkspacePath); err != nil {
					return err
				}
				if err := app.TaskManager.Close(task); err != nil {
					return err
				}
				if state, err := app.TaskManager.LoadState(); err == nil {
					if state.ActiveTaskID == task.ID {
						state.ActiveTaskID = ""
						_ = app.TaskManager.SaveState(state)
					}
				}
			}
			app.Logger.Info("Committed: %s", sha)
			if !noClose {
				app.Logger.Info("Task closed: %s", task.Name)
			}
			return nil
		},
	}
	cmd.Flags().String("message", "", "commit message")
	cmd.Flags().Bool("no-close", false, "do not close task after apply")
	return cmd
}

func rollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			stepID, _ := cmd.Flags().GetString("step")
			base, _ := cmd.Flags().GetBool("base")
			hard, _ := cmd.Flags().GetBool("hard")
			if stepID != "" && !base {
				return barerrors.RollbackNotSupported()
			}
			if !base {
				return barerrors.RollbackRequiresBase()
			}
			if err := app.WorkspaceManager.Reset(task.WorkspacePath, task.BaseRef, hard); err != nil {
				return err
			}
			taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
			ledgerManager := ledger.NewManager(taskDir)
			nextID, err := ledgerManager.NextStepID()
			if err != nil {
				return err
			}
			h := hard
			now := time.Now().UTC()
			step := &ledger.Step{
				StepID:     nextID,
				Kind:       ledger.StepKindRollback,
				StartedAt:  now,
				EndedAt:    now,
				Target:     "base",
				TargetStep: stepID,
				Hard:       &h,
			}
			if err := ledgerManager.Append(step); err != nil {
				return err
			}
			app.Logger.Info("Rolled back to base")
			return nil
		},
	}
	cmd.Flags().String("step", "", "rollback to a specific step")
	cmd.Flags().Bool("base", false, "rollback to base state")
	cmd.Flags().Bool("hard", false, "discard uncommitted changes")
	_ = cmd.RegisterFlagCompletionFunc("step", stepCompletionFunc)
	return cmd
}
