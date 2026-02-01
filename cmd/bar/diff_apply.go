package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/core/ledger"
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
			if stepID != "" {
				taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
				ledgerManager := ledger.NewManager(taskDir)
				step, err := ledgerManager.GetByID(stepID)
				if err != nil {
					return err
				}
				if step == nil {
					return errors.New("step not found")
				}
				if statOnly {
					if step.DiffStat != nil {
						app.Logger.Info("%d files changed, %d insertions(+), %d deletions(-)", step.DiffStat.Files, step.DiffStat.Additions, step.DiffStat.Deletions)
						return nil
					}
					app.Logger.Info("0 files changed")
					return nil
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
				return errors.New("patch not found")
			}
			result, err := app.DiffEngine.Generate(task.WorkspacePath, task.BaseRef)
			if err != nil {
				return err
			}
			if statOnly {
				app.Logger.Info("%d files changed, %d insertions(+), %d deletions(-)", result.Files, result.Additions, result.Deletions)
				return nil
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
	return cmd
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
				return errors.New("step rollback not supported in v0")
			}
			if !base {
				return errors.New("use --base to rollback")
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
	return cmd
}
