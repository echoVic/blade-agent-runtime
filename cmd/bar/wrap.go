package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/core/ledger"
)

func wrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wrap -- <command> [args...]",
		Short: "Wrap an interactive command and record changes on exit",
		Long: `Wrap an interactive command (like claude, cursor, aider) and automatically
record all file changes when the command exits.

Example:
  bar wrap -- claude
  bar wrap -- aider
  bar wrap -- cursor .

The command runs in the task's isolated workspace with full TTY support.
When the command exits, BAR records a step with the diff of all changes.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}

			startTime := time.Now().UTC()

			childCmd := exec.Command(args[0], args[1:]...)
			childCmd.Dir = task.WorkspacePath
			childCmd.Stdin = os.Stdin
			childCmd.Stdout = os.Stdout
			childCmd.Stderr = os.Stderr
			childCmd.Env = append(os.Environ(),
				"BAR_ACTIVE=true",
				"BAR_TASK_ID="+task.ID,
				"BAR_TASK_NAME="+task.Name,
				"BAR_WORKSPACE="+task.WorkspacePath,
				"BAR_BASE_REF="+task.BaseRef,
				"BAR_REPO_ROOT="+task.RepoRoot,
			)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigChan
				if childCmd.Process != nil {
					childCmd.Process.Signal(sig)
				}
			}()

			app.Logger.Info("Starting wrapped command in %s", task.WorkspacePath)
			app.Logger.Info("Changes will be recorded when the command exits")
			app.Logger.Info("")

			runErr := childCmd.Run()

			endTime := time.Now().UTC()
			duration := endTime.Sub(startTime)

			exitCode := 0
			if runErr != nil {
				if exitErr, ok := runErr.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			app.Logger.Info("")
			app.Logger.Info("Command exited with code %d", exitCode)

			taskDir := filepath.Join(app.BarDir, "tasks", task.ID)
			ledgerManager := ledger.NewManager(taskDir)

			stepID, err := ledgerManager.NextStepID()
			if err != nil {
				return err
			}

			diffResult, err := app.DiffEngine.Generate(task.WorkspacePath, task.BaseRef)
			if err != nil {
				return err
			}

			if diffResult.Files == 0 {
				app.Logger.Info("No changes detected, skipping step record")
				return nil
			}

			artifactsDir := filepath.Join(taskDir, "artifacts")
			if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
				return err
			}

			patchPath := filepath.Join(artifactsDir, stepID+".patch")
			if err := os.WriteFile(patchPath, diffResult.Patch, 0o644); err != nil {
				return err
			}

			step := &ledger.Step{
				StepID:     stepID,
				Kind:       ledger.StepKindRun,
				StartedAt:  startTime,
				EndedAt:    endTime,
				DurationMs: duration.Milliseconds(),
				Cmd:        args,
				Cwd:        task.WorkspacePath,
				ExitCode:   &exitCode,
				DiffStat: &ledger.DiffStat{
					Files:     diffResult.Files,
					Additions: diffResult.Additions,
					Deletions: diffResult.Deletions,
					FileList:  diffResult.FileList,
				},
				Artifacts: &ledger.Artifacts{
					Patch: filepath.Join("artifacts", stepID+".patch"),
				},
			}

			if err := ledgerManager.Append(step); err != nil {
				return err
			}

			app.Logger.Info("Step %s recorded", stepID)
			app.Logger.Info("Files changed: %d (+%d, -%d)", diffResult.Files, diffResult.Additions, diffResult.Deletions)

			return nil
		},
	}
	return cmd
}
