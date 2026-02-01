package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
	"github.com/user/blade-agent-runtime/internal/core/apply"
	"github.com/user/blade-agent-runtime/internal/core/config"
	"github.com/user/blade-agent-runtime/internal/core/diff"
	"github.com/user/blade-agent-runtime/internal/core/exec"
	"github.com/user/blade-agent-runtime/internal/core/policy"
	"github.com/user/blade-agent-runtime/internal/core/task"
	"github.com/user/blade-agent-runtime/internal/core/workspace"
	utillog "github.com/user/blade-agent-runtime/internal/util/log"
	utilpath "github.com/user/blade-agent-runtime/internal/util/path"
)

type App struct {
	RepoRoot         string
	BarDir           string
	ConfigPath       string
	Config           *config.Config
	Logger           *utillog.Logger
	Git              *gitadapter.Runner
	TaskManager      *task.Manager
	WorkspaceManager *workspace.Manager
	DiffEngine       *diff.Engine
	ApplyEngine      *apply.Engine
	PolicyEngine     *policy.Engine
	ExecRunner       *exec.Runner
}

var rootCmd = &cobra.Command{
	Use:   "bar",
	Short: "Blade Agent Runtime",
}

func Execute() error {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output")
	rootCmd.PersistentFlags().String("config", "", "config file path")
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(taskCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(wrapCmd())
	rootCmd.AddCommand(diffCmd())
	rootCmd.AddCommand(applyCmd())
	rootCmd.AddCommand(rollbackCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(logCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(uiCmd())
	return rootCmd.Execute()
}

func initApp(requireBar bool) (*App, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repoRoot, err := utilpath.FindRepoRoot(cwd)
	if err != nil {
		return nil, err
	}
	barDir := utilpath.BarDir(repoRoot)
	if requireBar {
		if _, err := os.Stat(barDir); err != nil {
			return nil, errors.New("bar not initialized")
		}
	}
	cfgPath, _ := rootCmd.Flags().GetString("config")
	if cfgPath == "" {
		cfgPath = filepath.Join(barDir, "config.yaml")
	}
	cfgManager := config.NewManager(cfgPath)
	cfg, err := cfgManager.Load()
	if err != nil {
		return nil, err
	}
	verbose, _ := rootCmd.Flags().GetBool("verbose")
	quiet, _ := rootCmd.Flags().GetBool("quiet")
	logger := utillog.New(os.Stdout, os.Stderr, verbose || cfg.Output.Verbose, quiet)
	gitRunner := gitadapter.NewRunner()
	tm := task.NewManager(repoRoot, barDir)
	wm := workspace.NewManager(repoRoot, filepath.Join(barDir, "workspaces"), gitRunner)
	diffEngine := diff.NewEngine(gitRunner)
	applyEngine := apply.NewEngine(gitRunner)
	policyEngine := policy.NewEngine()
	if cfg.Policy.Enabled {
		if err := policyEngine.Load(cfg.Policy.Path); err != nil {
			return nil, err
		}
	}
	return &App{
		RepoRoot:         repoRoot,
		BarDir:           barDir,
		ConfigPath:       cfgPath,
		Config:           cfg,
		Logger:           logger,
		Git:              gitRunner,
		TaskManager:      tm,
		WorkspaceManager: wm,
		DiffEngine:       diffEngine,
		ApplyEngine:      applyEngine,
		PolicyEngine:     policyEngine,
		ExecRunner:       exec.NewRunner(),
	}, nil
}

func requireActiveTask(app *App) (*task.Task, error) {
	t, err := app.TaskManager.GetActive()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func fail(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w", err)
}
