package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/core/config"
	"github.com/user/blade-agent-runtime/internal/core/task"
	utillog "github.com/user/blade-agent-runtime/internal/util/log"
	utilpath "github.com/user/blade-agent-runtime/internal/util/path"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize BAR in current repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initApp(false)
			if err != nil {
				return err
			}
			force, _ := cmd.Flags().GetBool("force")
			if _, err := os.Stat(app.BarDir); err == nil && !force {
				return fail(errAlreadyInitialized)
			}
			if err := utilpath.EnsureDir(app.BarDir); err != nil {
				return err
			}
			if err := utilpath.EnsureDir(filepath.Join(app.BarDir, "tasks")); err != nil {
				return err
			}
			if err := utilpath.EnsureDir(filepath.Join(app.BarDir, "workspaces")); err != nil {
				return err
			}
			cfg := config.DefaultConfig()
			cfgManager := config.NewManager(app.ConfigPath)
			if err := cfgManager.Save(cfg); err != nil {
				return err
			}
			state := task.DefaultState()
			if err := task.SaveState(filepath.Join(app.BarDir, "state.json"), state); err != nil {
				return err
			}
			app.Logger.Info("Initialized BAR in %s", app.BarDir)
			checkGitignore(app.RepoRoot, app.Logger)
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "force reinitialize")
	return cmd
}

var errAlreadyInitialized = failString("bar already initialized (use --force to reinitialize)")

func failString(msg string) error {
	return &stringError{msg: msg}
}

type stringError struct {
	msg string
}

func (e *stringError) Error() string {
	return e.msg
}

func initAppWithAutoInit() (*App, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repoRoot, err := utilpath.FindRepoRoot(cwd)
	if err != nil {
		return nil, err
	}
	barDir := utilpath.BarDir(repoRoot)
	cfgPath := filepath.Join(barDir, "config.yaml")

	if _, err := os.Stat(barDir); err != nil {
		if err := utilpath.EnsureDir(barDir); err != nil {
			return nil, err
		}
		if err := utilpath.EnsureDir(filepath.Join(barDir, "tasks")); err != nil {
			return nil, err
		}
		if err := utilpath.EnsureDir(filepath.Join(barDir, "workspaces")); err != nil {
			return nil, err
		}
		cfg := config.DefaultConfig()
		cfgManager := config.NewManager(cfgPath)
		if err := cfgManager.Save(cfg); err != nil {
			return nil, err
		}
		state := task.DefaultState()
		if err := task.SaveState(filepath.Join(barDir, "state.json"), state); err != nil {
			return nil, err
		}
	}

	app, err := initApp(true)
	if err != nil {
		return nil, err
	}
	checkGitignore(repoRoot, app.Logger)
	return app, nil
}

func ensureBarInit(app *App) error {
	if _, err := os.Stat(app.BarDir); err == nil {
		return nil
	}
	if err := utilpath.EnsureDir(app.BarDir); err != nil {
		return err
	}
	if err := utilpath.EnsureDir(filepath.Join(app.BarDir, "tasks")); err != nil {
		return err
	}
	if err := utilpath.EnsureDir(filepath.Join(app.BarDir, "workspaces")); err != nil {
		return err
	}
	cfg := config.DefaultConfig()
	cfgManager := config.NewManager(app.ConfigPath)
	if err := cfgManager.Save(cfg); err != nil {
		return err
	}
	state := task.DefaultState()
	if err := task.SaveState(filepath.Join(app.BarDir, "state.json"), state); err != nil {
		return err
	}
	app.Logger.Info("Initialized BAR in %s", app.BarDir)
	checkGitignore(app.RepoRoot, app.Logger)
	return nil
}

func checkGitignore(repoRoot string, logger *utillog.Logger) {
	path := filepath.Join(repoRoot, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Info("Tip: Add '.bar/' to your .gitignore")
		return
	}

	content := string(data)
	if !strings.Contains(content, ".bar/") && !strings.Contains(content, ".bar") {
		logger.Info("Tip: Add '.bar/' to your .gitignore")
	}
}
