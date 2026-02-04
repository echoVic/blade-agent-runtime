package main

import (
	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/completion"
	"github.com/user/blade-agent-runtime/internal/core/task"
	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
)

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}
	cmd.AddCommand(taskStartCmd())
	cmd.AddCommand(taskListCmd())
	cmd.AddCommand(taskSwitchCmd())
	cmd.AddCommand(taskCloseCmd())
	return cmd
}

func taskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			tasks, err := app.TaskManager.List()
			if err != nil {
				return err
			}
			all, _ := cmd.Flags().GetBool("all")
			state, err := app.TaskManager.LoadState()
			if err != nil {
				return err
			}
			app.Logger.Info("ID       NAME                  STATUS   CREATED")
			for _, t := range tasks {
				if !all && t.Status == "closed" {
					continue
				}
				mark := " "
				if t.ID == state.ActiveTaskID {
					mark = "*"
				}
				app.Logger.Info("%s %-7s %-20s %-7s %s", mark, t.ID, trim(t.Name, 20), t.Status, t.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	cmd.Flags().Bool("all", false, "show closed tasks")
	return cmd
}

func taskSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <task_id|name>",
		Short: "Switch active task",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			app, err := initApp(true)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			completions := completion.GetTaskCompletions(app.BarDir, false)
			return completion.ToCobraCompletions(completions), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			key := args[0]
			task, err := app.TaskManager.Get(key)
			if err != nil {
				task, err = app.TaskManager.ResolveByName(key)
				if err != nil {
					return err
				}
			}
			if err := app.TaskManager.SetActive(task.ID); err != nil {
				return err
			}
			app.Logger.Info("Switched to task: %s (%s)", task.Name, task.ID)
			return nil
		},
	}
	return cmd
}

func taskCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [task_id]",
		Short: "Close a task",
		Args:  cobra.RangeArgs(0, 1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			app, err := initApp(true)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			completions := completion.GetTaskCompletions(app.BarDir, false)
			return completion.ToCobraCompletions(completions), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(true)
			if err != nil {
				return err
			}
			var t *task.Task
			if len(args) == 0 {
				task, err := app.TaskManager.GetActive()
				if err != nil {
					return err
				}
				t = task
			} else {
				task, err := app.TaskManager.Get(args[0])
				if err != nil {
					task, err = app.TaskManager.ResolveByName(args[0])
					if err != nil {
						return err
					}
				}
				t = task
			}
			force, _ := cmd.Flags().GetBool("force")
			keep, _ := cmd.Flags().GetBool("keep")
			del, _ := cmd.Flags().GetBool("delete")
			if !force {
				clean, err := app.WorkspaceManager.IsClean(t.WorkspacePath)
				if err != nil {
					return err
				}
				if !clean {
				return barerrors.WorkspaceNotClean()
			}
			}
			if !keep {
				if err := app.WorkspaceManager.Delete(t.WorkspacePath); err != nil {
					return err
				}
			}
			if del {
				if err := app.TaskManager.Delete(t.ID); err != nil {
					return err
				}
			} else {
				if err := app.TaskManager.Close(t); err != nil {
					return err
				}
			}
			if state, err := app.TaskManager.LoadState(); err == nil {
				if state.ActiveTaskID == t.ID {
					state.ActiveTaskID = ""
					_ = app.TaskManager.SaveState(state)
				}
			}
			if del {
				app.Logger.Info("Deleted task: %s (%s)", t.Name, t.ID)
			} else {
				app.Logger.Info("Closed task: %s (%s)", t.Name, t.ID)
			}
			if !keep {
				app.Logger.Info("Worktree deleted: %s", t.WorkspacePath)
			} else {
				app.Logger.Info("Worktree kept: %s", t.WorkspacePath)
			}
			return nil
		},
	}
	cmd.Flags().Bool("keep", false, "keep worktree")
	cmd.Flags().Bool("delete", false, "delete task records")
	cmd.Flags().Bool("force", false, "force close")
	return cmd
}

func trim(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}
