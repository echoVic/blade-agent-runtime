package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	gitadapter "github.com/user/blade-agent-runtime/internal/adapters/git"
	"github.com/user/blade-agent-runtime/internal/core/apply"
	"github.com/user/blade-agent-runtime/internal/core/config"
	"github.com/user/blade-agent-runtime/internal/core/diff"
	"github.com/user/blade-agent-runtime/internal/core/exec"
	"github.com/user/blade-agent-runtime/internal/core/policy"
	"github.com/user/blade-agent-runtime/internal/core/task"
	"github.com/user/blade-agent-runtime/internal/core/workspace"
	"github.com/user/blade-agent-runtime/internal/guide"
	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
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
	Run: func(cmd *cobra.Command, args []string) {
		if isInteractive() {
			showQuickStart()
		} else {
			_ = cmd.Help()
		}
	},
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
			return nil, barerrors.NotInitialized()
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

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func newGuide() *guide.Guide {
	return guide.New(os.Stdin, os.Stdout)
}

func showQuickStart() {
	g := newGuide()
	g.Print("")
	g.Print("🚀 BAR - Blade Agent Runtime")
	g.Print("")
	g.Print("Quick Start:")
	g.Print("  bar task start <name>    Create a new task and start working")
	g.Print("  bar run -- <command>     Run a command in the isolated workspace")
	g.Print("  bar diff                 View changes made by the agent")
	g.Print("  bar apply                Apply changes to the main branch")
	g.Print("")
	g.Print("Common Commands:")
	g.Print("  bar status               Show current status")
	g.Print("  bar log                  View operation history")
	g.Print("  bar task list            List all tasks")
	g.Print("  bar rollback --base      Reset to initial state")
	g.Print("")
	g.Print("Get Started:")
	g.Print("  bar task start fix-bug   Create your first task")
	g.Print("")
	g.Print("For more information: bar --help")
	g.Print("")
}


