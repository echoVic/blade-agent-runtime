package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/user/blade-agent-runtime/internal/web"
)

func uiCmd() *cobra.Command {
	var port int
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Start Web UI server",
		Long:  `Start a web server for viewing tasks, ledger, and diffs in the browser.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := initAppWithAutoInit()
			if err != nil {
				return err
			}

			addr := fmt.Sprintf(":%d", port)
			server := web.NewServer(addr, app.TaskManager, app.BarDir)

			// Handle graceful shutdown
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			// Start server in goroutine
			errChan := make(chan error, 1)
			go func() {
				errChan <- server.Start()
			}()

			// Open browser if requested
			if !noOpen {
				url := fmt.Sprintf("http://localhost%d", port)
				app.Logger.Info("Opening %s in browser...", url)
				openBrowser(url)
			}

			app.Logger.Info("Web UI running. Press Ctrl+C to stop.")

			select {
			case <-ctx.Done():
				app.Logger.Info("Shutting down...")
				return server.Stop()
			case err := <-errChan:
				return err
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't open browser automatically")

	return cmd
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}
