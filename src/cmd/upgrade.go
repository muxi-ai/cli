package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/updates"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade MUXI CLI to the latest version",
	Long: `Upgrade MUXI CLI to the latest version.

Detects your installation method and provides the appropriate upgrade path:
  - Homebrew: Shows brew upgrade command
  - Go install: Shows go install command  
  - Binary: Downloads and replaces the current binary`,
	RunE: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	// Force refresh cache to get latest version
	updates.RefreshCache()
	
	updateInfo := updates.CheckCachedUpdate()
	
	if updateInfo == nil {
		fmt.Println()
		ui.Success("MUXI CLI is up to date")
		fmt.Printf("  Current version: %s\n", version)
		return nil
	}

	fmt.Println()
	fmt.Printf("Current version: %s\n", updateInfo.CurrentVersion)
	fmt.Printf("Latest version:  %s\n", updateInfo.LatestVersion)
	fmt.Println()

	method := detectInstallMethod()

	switch method {
	case "homebrew":
		return upgradeHomebrew()
	case "go":
		return upgradeGo()
	default:
		return upgradeBinary(updateInfo.LatestVersion)
	}
}

// detectInstallMethod determines how the CLI was installed
func detectInstallMethod() string {
	exe, err := os.Executable()
	if err != nil {
		return "binary"
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "binary"
	}

	exeLower := strings.ToLower(exe)

	// Check for Homebrew (macOS/Linux)
	if strings.Contains(exeLower, "homebrew") || 
	   strings.Contains(exeLower, "cellar") ||
	   strings.Contains(exeLower, "linuxbrew") {
		return "homebrew"
	}

	// Check for Go install
	if strings.Contains(exeLower, "go/bin") ||
	   strings.Contains(exeLower, "gopath") {
		return "go"
	}

	return "binary"
}

func upgradeHomebrew() error {
	ui.Info("Detected: Homebrew installation")
	fmt.Println()
	fmt.Println("Run the following command to upgrade:")
	fmt.Println()
	fmt.Println("  brew upgrade muxi")
	fmt.Println()
	return nil
}

func upgradeGo() error {
	ui.Info("Detected: Go installation")
	fmt.Println()
	fmt.Println("Run the following command to upgrade:")
	fmt.Println()
	fmt.Println("  go install github.com/muxi-ai/cli@latest")
	fmt.Println()
	return nil
}

func upgradeBinary(latestVersion string) error {
	ui.Info("Detected: Binary installation")
	fmt.Println()

	// Determine platform
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build download URL
	binaryName := fmt.Sprintf("muxi-%s-%s", goos, goarch)
	if goos == "windows" {
		binaryName += ".exe"
	}

	downloadURL := fmt.Sprintf(
		"https://github.com/muxi-ai/cli/releases/download/v%s/%s",
		latestVersion,
		binaryName,
	)

	fmt.Printf("Downloading %s...\n", binaryName)

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Download to temp file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Create temp file in same directory as executable
	tmpFile, err := os.CreateTemp(filepath.Dir(exePath), "muxi-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Copy downloaded content
	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to download: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace old binary
	// On Windows, we need to rename the old one first
	if goos == "windows" {
		oldPath := exePath + ".old"
		os.Remove(oldPath) // Remove any previous .old file
		if err := os.Rename(exePath, oldPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to backup old binary: %w", err)
		}
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Upgraded to MUXI CLI %s", latestVersion))
	return nil
}
