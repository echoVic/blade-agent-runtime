package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jaevor/go-nanoid"
	"github.com/spf13/cobra"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
	"github.com/user/blade-agent-runtime/internal/core/ledger"
	"github.com/user/blade-agent-runtime/internal/core/task"
	"github.com/user/blade-agent-runtime/internal/web"
)

func wrapCmd() *cobra.Command {
	var noUI bool
	var uiPort int
	var uiServer *web.Server

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
			app, err := initAppWithAutoInit()
			if err != nil {
				return err
			}

			// Start UI by default (unless --no-ui is set)
			if !noUI {
				addr := fmt.Sprintf(":%d", uiPort)
				uiServer = web.NewServer(addr, app.TaskManager, app.BarDir)
				go func() {
					if err := uiServer.Start(); err != nil {
						app.Logger.Error("Web UI failed: %v", err)
					}
				}()

				// Wait a bit for server to start
				time.Sleep(500 * time.Millisecond)
				url := fmt.Sprintf("http://localhost:%d", uiPort)
				app.Logger.Info("Web UI running at %s", url)
				openBrowser(url)
			}

			task, err := getOrCreateTask(app, args[0])
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

			// If UI is running, wait for user to close it
			if uiServer != nil {
				app.Logger.Info("")
				app.Logger.Info("Web UI still running. Press Ctrl+C to exit.")

				waitChan := make(chan os.Signal, 1)
				signal.Notify(waitChan, syscall.SIGINT, syscall.SIGTERM)
				<-waitChan

				uiServer.Stop()
				app.Logger.Info("Web UI stopped.")
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&noUI, "no-ui", false, "Disable Web UI")
	cmd.Flags().IntVarP(&uiPort, "port", "p", 8080, "Port for Web UI")
	return cmd
}

func getOrCreateTask(app *App, cmdName string) (*task.Task, error) {
	activeTask, err := app.TaskManager.GetActive()
	if err == nil && activeTask != nil {
		return activeTask, nil
	}

	name := "wrap-" + sanitizeName(cmdName)
	base := ""
	_, branch, err := gitadapter.CurrentHEAD(app.RepoRoot)
	if err != nil {
		return nil, err
	}
	if branch != "" {
		base = branch
	} else {
		head, _, err := gitadapter.CurrentHEAD(app.RepoRoot)
		if err != nil {
			return nil, err
		}
		base = head
	}

	gen, err := nanoid.Standard(8)
	if err != nil {
		return nil, err
	}
	id := gen()
	branchName := app.Config.Git.BranchPrefix + name + "-" + id
	workspacePath := filepath.Join(app.BarDir, "workspaces", id)

	if _, err := app.WorkspaceManager.Create(id, branchName, base); err != nil {
		return nil, err
	}

	t, err := app.TaskManager.Create(id, name, base, branchName, workspacePath)
	if err != nil {
		_ = app.WorkspaceManager.Delete(workspacePath)
		return nil, err
	}

	if err := app.TaskManager.SetActive(t.ID); err != nil {
		return nil, err
	}

	app.Logger.Info("Created task: %s (id: %s)", t.Name, t.ID)
	app.Logger.Info("Workspace: %s", t.WorkspacePath)

	return t, nil
}
