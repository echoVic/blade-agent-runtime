package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
	// Try different commands to open browser
	for _, cmd := range []string{"open", "xdg-open", "start"} {
		if _, err := os.Stat("/usr/bin/" + cmd); err == nil || cmd == "start" {
			syscall.Exec("/usr/bin/"+cmd, []string{cmd, url}, os.Environ())
			return
		}
	}
}
