package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/core/exec"
	"github.com/user/blade-agent-runtime/internal/core/ledger"
	"github.com/user/blade-agent-runtime/internal/core/policy"
	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "Run a command in current task workspace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			task, err := requireActiveTask(app)
			if err != nil {
				return err
			}
			timeout, _ := cmd.Flags().GetDuration("timeout")
			noRecord, _ := cmd.Flags().GetBool("no-record")
			envFlags, _ := cmd.Flags().GetStringArray("env")
			cwdFlag, _ := cmd.Flags().GetString("cwd")
			env := map[string]string{}
			for _, kv := range envFlags {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				}
			}
			env["BAR_ACTIVE"] = "true"
			env["BAR_TASK_ID"] = task.ID
			env["BAR_TASK_NAME"] = task.Name
			env["BAR_WORKSPACE"] = task.WorkspacePath
			env["BAR_BASE_REF"] = task.BaseRef
			env["BAR_REPO_ROOT"] = task.RepoRoot
			if app.Config.Policy.Enabled {
				res, err := app.PolicyEngine.Check(args)
				if err != nil {
					return err
				}
				if !res.Allowed {
				rule := ""
				reason := ""
				if len(res.Events) > 0 {
					rule = res.Events[0].Rule
					reason = res.Events[0].Reason
				}
				return barerrors.PolicyViolation(rule, reason)
			}
				for _, ev := range res.Events {
					if ev.Action == "warn" {
						app.Logger.Info("Policy warning: %s", ev.Reason)
					}
				}
			}
			cwd := task.WorkspacePath
			if cwdFlag != "" {
				cwd = filepath.Join(cwd, cwdFlag)
			}
			ctx := context.Background()
			opts := execOptions(timeout, cwd, env)
			result, err := app.ExecRunner.Run(ctx, args, &opts)
			if err != nil {
				return err
			}
			if noRecord {
				app.Logger.Info("Exit code: %d", result.ExitCode)
				return nil
			}
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
			artifactsDir := filepath.Join(taskDir, "artifacts")
			if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
				return err
			}
			patchPath := filepath.Join(artifactsDir, stepID+".patch")
			if err := os.WriteFile(patchPath, diffResult.Patch, 0o644); err != nil {
				return err
			}
			outputPath := filepath.Join(artifactsDir, stepID+".output")
			if err := writeOutput(outputPath, result.Stdout, result.Stderr); err != nil {
				return err
			}
			exit := result.ExitCode
			step := &ledger.Step{
				StepID:     stepID,
				Kind:       ledger.StepKindRun,
				StartedAt:  time.Now().Add(-result.Duration).UTC(),
				EndedAt:    time.Now().UTC(),
				DurationMs: result.Duration.Milliseconds(),
				Cmd:        args,
				Cwd:        cwd,
				ExitCode:   &exit,
				DiffStat: &ledger.DiffStat{
					Files:     diffResult.Files,
					Additions: diffResult.Additions,
					Deletions: diffResult.Deletions,
				},
				Artifacts: &ledger.Artifacts{
					Patch:  filepath.Join("artifacts", stepID+".patch"),
					Output: filepath.Join("artifacts", stepID+".output"),
				},
			}
			if app.Config.Policy.Enabled {
				res, _ := app.PolicyEngine.Check(args)
				if res != nil && len(res.Events) > 0 {
					step.PolicyEvents = policyEvents(res.Events)
				}
			}
			if err := ledgerManager.Append(step); err != nil {
				return err
			}
			app.Logger.Info("Step %s completed (exit code: %d)", stepID, result.ExitCode)
			app.Logger.Info("Files changed: %d (+%d, -%d)", diffResult.Files, diffResult.Additions, diffResult.Deletions)
			return nil
		},
	}
	cmd.Flags().Duration("timeout", 0, "timeout")
	cmd.Flags().Bool("no-record", false, "do not record to ledger")
	cmd.Flags().StringArray("env", []string{}, "environment variables")
	cmd.Flags().String("cwd", "", "working directory inside workspace")
	return cmd
}

func execOptions(timeout time.Duration, cwd string, env map[string]string) exec.Options {
	return exec.Options{
		Cwd:     cwd,
		Env:     env,
		Timeout: timeout,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
	}
}

func writeOutput(path string, stdout []byte, stderr []byte) error {
	content := []byte("=== STDOUT ===\n")
	content = append(content, stdout...)
	content = append(content, []byte("\n\n=== STDERR ===\n")...)
	content = append(content, stderr...)
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func policyEvents(events []policy.Event) []ledger.PolicyEvent {
	out := []ledger.PolicyEvent{}
	for _, e := range events {
		out = append(out, ledger.PolicyEvent{
			Rule:    e.Rule,
			Action:  e.Action,
			Matched: e.Matched,
		})
	}
	return out
}
