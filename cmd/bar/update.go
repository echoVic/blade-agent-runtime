package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	repoOwner      = "echoVic"
	repoName       = "blade-agent-runtime"
	installURL     = "https://echovic.github.io/blade-agent-runtime/install.sh"
	currentVersion = "0.0.19"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update BAR to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			check, _ := cmd.Flags().GetBool("check")

			latest, err := getLatestVersion()
			if err != nil {
				return fmt.Errorf("failed to check latest version: %w", err)
			}

			fmt.Printf("Current version: v%s\n", currentVersion)
			fmt.Printf("Latest version:  %s\n", latest)

			if "v"+currentVersion == latest {
				fmt.Println("\nYou are already on the latest version!")
				return nil
			}

			if check {
				fmt.Println("\nRun 'bar update' to upgrade.")
				return nil
			}

			fmt.Println("\nUpdating...")

			var updateCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				return fmt.Errorf("auto-update not supported on Windows, please reinstall manually")
			}

			updateCmd = exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL %s | sh", installURL))
			updateCmd.Stdout = os.Stdout
			updateCmd.Stderr = os.Stderr

			if err := updateCmd.Run(); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}

			fmt.Println("\nUpdate complete! Run 'bar --version' to verify.")
			return nil
		},
	}
	cmd.Flags().Bool("check", false, "only check for updates, don't install")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("bar version %s\n", currentVersion)
			fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimSpace(release.TagName), nil
}
