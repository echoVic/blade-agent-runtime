package main

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jaevor/go-nanoid"
	"github.com/spf13/cobra"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
)

func taskStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			name := args[0]
			base, _ := cmd.Flags().GetString("base")
			if base == "" {
				_, branch, err := gitadapter.CurrentHEAD(app.RepoRoot)
				if err != nil {
					return err
				}
				if branch != "" {
					base = branch
				} else {
					head, _, err := gitadapter.CurrentHEAD(app.RepoRoot)
					if err != nil {
						return err
					}
					base = head
				}
			}
			gen, err := nanoid.Standard(8)
			if err != nil {
				return err
			}
			id := gen()
			branchName := app.Config.Git.BranchPrefix + sanitizeName(name) + "-" + id
			workspacePath := filepath.Join(app.BarDir, "workspaces", id)
			if _, err := app.WorkspaceManager.Create(id, branchName, base); err != nil {
				return err
			}
			task, err := app.TaskManager.Create(id, name, base, branchName, workspacePath)
			if err != nil {
				_ = app.WorkspaceManager.Delete(workspacePath)
				return err
			}
			noSwitch, _ := cmd.Flags().GetBool("no-switch")
			if !noSwitch {
				if err := app.TaskManager.SetActive(task.ID); err != nil {
					return err
				}
			}
			app.Logger.Info("Created task: %s (id: %s)", task.Name, task.ID)
			app.Logger.Info("Workspace: %s", task.WorkspacePath)
			app.Logger.Info("Branch: %s", task.Branch)
			if !noSwitch {
				app.Logger.Info("Switched to task: %s", task.Name)
			}
			return nil
		},
	}
	cmd.Flags().String("base", "", "base branch or commit")
	cmd.Flags().Bool("no-switch", false, "do not switch to the new task")
	return cmd
}

func sanitizeName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\-]+`)
	out := strings.ToLower(name)
	out = strings.ReplaceAll(out, " ", "-")
	out = re.ReplaceAllString(out, "-")
	out = strings.Trim(out, "-")
	if out == "" {
		out = "task"
	}
	return out
}
