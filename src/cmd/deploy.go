package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/validate"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var deployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy formation to server",
	GroupID: "formation",
	Long: `Deploy a formation to a MUXI server.

If the formation already exists on the server, it will be updated.
If it's a new formation, it will be created.

By default, deployment progress is streamed from the server.
Use --no-stream for a simpler non-streaming mode.`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("profile", "", "Server profile to use")
	deployCmd.Flags().Bool("dry-run", false, "Validate and create bundle without deploying")
	deployCmd.Flags().Bool("no-stream", false, "Disable streaming progress (simpler output)")
}

// ExcludedPatterns are patterns to exclude from deploy bundles
var deployExcludedPatterns = []string{
	".git",
	".muxi",
	"secrets", // Exclude secrets template, but NOT secrets.enc or .key
	".env",
	".env.*",
	"node_modules",
	"__pycache__",
	"*.pyc",
	".DS_Store",
	"*.log",
	"*.tmp",
	".vscode",
	".idea",
}

// FormationMetadata represents minimal formation.yaml fields we need
type FormationMetadata struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func runDeploy(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	noStream, _ := cmd.Flags().GetBool("no-stream")

	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"Run this command from inside a formation directory.",
			"cd my-formation && muxi deploy",
		)
		os.Exit(1)
	}

	// Read formation config to get ID and version
	formationPath, found := context.FindFormationFile(ctx.RootDir)
	if !found {
		return fmt.Errorf("formation config file not found (formation.afs or formation.yaml)")
	}
	formationData, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation config: %w", err)
	}

	var metadata FormationMetadata
	if err := yaml.Unmarshal(formationData, &metadata); err != nil {
		return fmt.Errorf("failed to parse formation config: %w", err)
	}

	if metadata.ID == "" {
		ui.ErrorBlock(
			"Missing formation ID",
			"Formation config must have an 'id' field.",
			"",
		)
		os.Exit(1)
	}

	// Validate formation before deploying
	validationResult, err := validate.Formation(ctx.RootDir)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !validationResult.IsValid() {
		fmt.Println()
		ui.Error(fmt.Sprintf("Formation '%s' has %d error(s)", metadata.ID, len(validationResult.Errors)))
		fmt.Println()
		fmt.Printf("  Run %s to see all issues\n", ui.CyanText("muxi validate"))
		fmt.Println()
		os.Exit(1)
	}

	// Get server client
	client, err := server.NewClient(profile)
	if err != nil {
		return err
	}

	serverName := profile
	if serverName == "" {
		serverName = server.GetDefaultServer()
	}

	// Check if formation already exists
	existingFormation, _ := client.GetFormation(metadata.ID)
	isUpdate := existingFormation != nil

	// For updates, version is required and must be higher than server version
	if isUpdate {
		if metadata.Version == "" {
			ui.ErrorBlock(
				"Version required for update",
				fmt.Sprintf("Formation '%s' already exists on server.\nAdd a 'version' field to formation.yaml before updating.", metadata.ID),
				"version: \"1.1.0\"",
			)
			os.Exit(1)
		}

		// Check version is higher than server version
		serverVersion := ""
		if existingFormation.Version != nil {
			serverVersion = existingFormation.Version.Semantic
		}

		if serverVersion != "" && !isVersionHigher(metadata.Version, serverVersion) {
			ui.ErrorBlock(
				"Version conflict",
				fmt.Sprintf("Cannot update '%s' to version %s\nServer already has version %s", metadata.ID, metadata.Version, serverVersion),
				"Bump the version in formation.yaml and try again.",
			)
			os.Exit(1)
		}
	}

	// Display deploy info
	fmt.Println()
	if isUpdate {
		fmt.Printf("  Updating: %s on %s\n", ui.BoldText(metadata.ID), serverName)
		if metadata.Version != "" {
			fmt.Printf("  Version:  %s\n", metadata.Version)
		}
	} else {
		fmt.Printf("  Deploying: %s to %s\n", ui.BoldText(metadata.ID), serverName)
		if metadata.Version != "" {
			fmt.Printf("  Version:   %s\n", metadata.Version)
		}
	}
	fmt.Println()

	// Create bundle
	spinner := ui.NewSpinner("Creating bundle...")
	spinner.Start()

	bundlePath, fileCount, err := createTarGzBundle(ctx.RootDir, metadata.ID)
	if err != nil {
		spinner.StopWithError("Failed to create bundle")
		return fmt.Errorf("failed to create bundle: %w", err)
	}
	defer os.Remove(bundlePath)

	// Get bundle size
	bundleInfo, _ := os.Stat(bundlePath)
	bundleSize := bundleInfo.Size()

	spinner.StopWithSuccess(fmt.Sprintf("Created bundle (%d files, %s)", fileCount, formatBytes(bundleSize)))

	if dryRun {
		fmt.Println()
		ui.Success("Dry run complete - bundle created but not uploaded")
		fmt.Println()
		return nil
	}

	// Deploy with streaming or non-streaming
	if noStream {
		return deployNonStreaming(client, metadata, bundlePath, isUpdate)
	}
	return deployStreaming(client, metadata, bundlePath, isUpdate)
}

// deployStreaming deploys with SSE progress streaming
func deployStreaming(client *server.Client, metadata FormationMetadata, bundlePath string, isUpdate bool) error {
	// Start with "Pushing to server" spinner
	spinner := ui.NewSpinner("Pushing to server...")
	spinner.Start()

	// Set up signal handling for cleanup on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan

		// Stop spinner and show cancelled state
		spinner.Stop()
		fmt.Printf("\r\033[K  %s %s\n", ui.DimmedText("—"), ui.DimmedText("Cancelled"))

		cleanupSpinner := ui.NewSpinner("Cleaning up...")
		cleanupSpinner.Start()

		if isUpdate {
			// For updates, cancel the running update (original keeps running)
			err := client.CancelUpdate(metadata.ID)
			if err != nil && strings.Contains(err.Error(), "not found") {
				// Server hasn't registered yet, wait and retry
				time.Sleep(5 * time.Second)
				err = client.CancelUpdate(metadata.ID)
			}
			if err != nil && !strings.Contains(err.Error(), "not found") {
				cleanupSpinner.StopWithError(fmt.Sprintf("Cleanup failed: %v", err))
			} else {
				cleanupSpinner.StopWithSuccess("Original version still running")
			}
			fmt.Println()
			ui.Error("Update aborted")
		} else {
			// For new deployments, delete the partial formation
			err := client.DeleteFormation(metadata.ID)
			if err != nil && strings.Contains(err.Error(), "not found") {
				// Server hasn't registered yet, wait and retry
				time.Sleep(5 * time.Second)
				err = client.DeleteFormation(metadata.ID)
			}
			if err != nil && !strings.Contains(err.Error(), "not found") {
				cleanupSpinner.StopWithError(fmt.Sprintf("Cleanup failed: %v", err))
			} else {
				cleanupSpinner.StopWithSuccess("Cleaned up partial deployment")
			}
			fmt.Println()
			ui.Error("Deployment aborted")
		}
		fmt.Println()
		os.Exit(130) // Standard exit code for Ctrl+C
	}()

	var lastStage string
	var lastProgress *server.DeployProgressEvent
	var complete *server.DeployCompleteEvent
	var deployErr error

	// SSE callback - called for each progress event
	callback := func(event server.SSEEvent) error {
		if event.Event == "progress" {
			var progress server.DeployProgressEvent
			if err := json.Unmarshal([]byte(event.Data), &progress); err != nil {
				return nil // Skip malformed events
			}

			// Same stage with updated progress - just update the message
			if lastStage == progress.Stage {
				spinner.UpdateMessage(formatServerStageMessage(progress))
				lastProgress = &progress
				return nil
			}

			// First server event: finish "Pushing" and show it succeeded
			if lastStage == "" {
				spinner.StopWithSuccess("Pushed to server")
			} else {
				// Complete previous stage
				spinner.StopWithSuccess(formatStageComplete(lastStage, lastProgress))
			}

			// Start new spinner for this stage
			lastStage = progress.Stage
			lastProgress = &progress
			// Use padded spinner for health_check to give margin from terminal bottom
			if progress.Stage == "health_check" {
				spinner = ui.NewSpinnerWithPadding(formatServerStageMessage(progress), 1)
			} else {
				spinner = ui.NewSpinner(formatServerStageMessage(progress))
			}
			spinner.Start()
		}
		return nil
	}

	// Execute deploy
	if isUpdate {
		complete, deployErr = client.UpdateFormationStreaming(metadata.ID, bundlePath, metadata.Version, callback)
	} else {
		complete, deployErr = client.DeployFormationStreaming(metadata.ID, bundlePath, metadata.Version, callback)
	}

	// Handle final state
	if deployErr != nil {
		if lastStage != "" {
			spinner.StopWithError(ui.DimmedText("[SERVER]") + " Failed")
		} else {
			spinner.StopWithError("Push failed")
		}
		playNotificationSound(false)
		fmt.Println()
		return deployErr
	}

	// Complete final stage
	if lastStage != "" {
		spinner.StopWithSuccess(formatStageComplete(lastStage, lastProgress))
	} else {
		spinner.StopWithSuccess("Pushed to server")
	}

	// Success message
	fmt.Println()
	ui.Success(fmt.Sprintf("Formation `%s` running!", metadata.ID))

	// Show URL
	fmt.Println()
	if complete != nil && complete.URL != "" {
		fmt.Printf("  Formation URL: %s\n", complete.URL)
	} else {
		fmt.Printf("  Formation URL: %s/api/%s\n", client.BaseURL, metadata.ID)
	}
	fmt.Println()

	// Play notification sound
	playNotificationSound(true)

	return nil
}

// deployNonStreaming deploys without streaming (simpler output)
func deployNonStreaming(client *server.Client, metadata FormationMetadata, bundlePath string, isUpdate bool) error {
	var spinnerMsg string
	if isUpdate {
		spinnerMsg = "Deploying new version..."
	} else {
		spinnerMsg = "Deploying formation..."
	}
	spinner := ui.NewSpinner(spinnerMsg)
	spinner.Start()

	var deployErr error
	if isUpdate {
		deployErr = client.UpdateFormation(metadata.ID, bundlePath, metadata.Version)
	} else {
		deployErr = client.DeployFormation(metadata.ID, bundlePath, metadata.Version)
	}

	if deployErr != nil {
		spinner.StopWithError("Deploy failed")
		playNotificationSound(false)
		return deployErr
	}

	spinner.StopWithSuccess("Deployed")

	// Success message
	fmt.Println()
	if isUpdate {
		ui.Success(fmt.Sprintf("Updated %s", metadata.ID))
	} else {
		ui.Success(fmt.Sprintf("Deployed %s", metadata.ID))
	}

	// Show URL
	fmt.Println()
	fmt.Printf("  URL: %s/api/%s\n", client.BaseURL, metadata.ID)
	fmt.Println()

	// Play notification sound
	playNotificationSound(true)

	return nil
}

// playNotificationSound plays a sound to notify the user of completion
func playNotificationSound(success bool) {
	if runtime.GOOS == "darwin" {
		// macOS: use different system sounds for success/failure
		sound := "/System/Library/Sounds/Glass.aiff"
		if !success {
			sound = "/System/Library/Sounds/Sosumi.aiff"
		}
		exec.Command("afplay", sound).Run()
	} else {
		// Other platforms: ASCII bell
		fmt.Print("\a")
	}
}

// formatServerStageMessage formats a progress event into a spinner message with [SERVER] prefix
func formatServerStageMessage(p server.DeployProgressEvent) string {
	prefix := ui.DimmedText("[SERVER]") + " "

	switch p.Stage {
	case "extracting":
		return prefix + "Extracting formation bundle..."
	case "validating":
		return prefix + "Validating formation files..."
	case "resolving_runtime":
		return prefix + "Resolving runtime version..."
	case "downloading_sif":
		if p.Progress > 0 {
			return prefix + fmt.Sprintf("Downloading runtime image... %d%%", p.Progress)
		}
		return prefix + "Downloading runtime image..."
	case "pulling_runner":
		return prefix + "Pulling runtime runner..."
	case "spawning":
		return prefix + "Starting formation..."
	case "spawning_staging":
		return prefix + "Starting staging version..."
	case "health_check":
		if p.Attempt > 0 && p.MaxAttempts > 0 {
			remaining := p.MaxAttempts - p.Attempt
			return prefix + fmt.Sprintf("Waiting for formation to start (timeout: %s)", formatTimeout(remaining))
		}
		return prefix + "Waiting for formation to start..."
	case "swapping":
		return prefix + "Switching to new version..."
	case "stopping_old":
		return prefix + "Stopping old version..."
	default:
		if p.Message != "" {
			return prefix + p.Message
		}
		return prefix + p.Stage + "..."
	}
}

// formatStageComplete formats a completed stage message with [SERVER] prefix
func formatStageComplete(stage string, p *server.DeployProgressEvent) string {
	prefix := ui.DimmedText("[SERVER]") + " "

	switch stage {
	case "extracting":
		return prefix + "Extracted formation bundle"
	case "validating":
		return prefix + "Validated formation files"
	case "resolving_runtime":
		if p != nil && p.Version != "" {
			return prefix + "Resolved runtime version " + p.Version
		}
		return prefix + "Resolved runtime version"
	case "downloading_sif":
		return prefix + "Downloaded runtime image"
	case "pulling_runner":
		return prefix + "Pulled runtime runner"
	case "spawning":
		return prefix + "Started formation"
	case "spawning_staging":
		return prefix + "Started staging version"
	case "health_check":
		return prefix + "Formation started"
	case "swapping":
		return prefix + "Switched to new version"
	case "stopping_old":
		return prefix + "Stopped old version"
	default:
		return prefix + stage + " complete"
	}
}

// createTarGzBundle creates a tar.gz bundle of the formation directory
func createTarGzBundle(formationDir, formationID string) (string, int, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "muxi-deploy-*.tar.gz")
	if err != nil {
		return "", 0, err
	}
	tmpPath := tmpFile.Name()

	// Create gzip writer
	gzWriter := gzip.NewWriter(tmpFile)
	tarWriter := tar.NewWriter(gzWriter)

	fileCount := 0

	// Walk the directory
	err = filepath.Walk(formationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(formationDir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if excluded
		if shouldExclude(relPath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Use path inside formation ID directory
		header.Name = filepath.Join(formationID, relPath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
			fileCount++
		}

		return nil
	})

	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", 0, err
	}

	// Close writers
	if err := tarWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", 0, err
	}
	if err := gzWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", 0, err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", 0, err
	}

	return tmpPath, fileCount, nil
}

// shouldExclude checks if a path should be excluded
func shouldExclude(path string, isDir bool) bool {
	name := filepath.Base(path)

	for _, pattern := range deployExcludedPatterns {
		// Check exact match
		if name == pattern {
			return true
		}

		// Check glob pattern
		if strings.Contains(pattern, "*") {
			matched, _ := filepath.Match(pattern, name)
			if matched {
				return true
			}
		}

		// Check if path starts with pattern (for directories)
		if isDir && strings.HasPrefix(path, pattern) {
			return true
		}
	}

	return false
}

// formatTimeout formats seconds as m:ss for countdown display
func formatTimeout(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// isVersionHigher returns true if newVersion is strictly higher than oldVersion
// Supports semver format: major.minor.patch with optional suffix
func isVersionHigher(newVersion, oldVersion string) bool {
	newParts := parseVersion(newVersion)
	oldParts := parseVersion(oldVersion)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if newParts[i] > oldParts[i] {
			return true
		}
		if newParts[i] < oldParts[i] {
			return false
		}
	}

	// All parts equal - not higher
	return false
}

// parseVersion extracts major, minor, patch from version string
func parseVersion(version string) [3]int {
	var parts [3]int

	// Strip leading 'v' if present
	version = strings.TrimPrefix(version, "v")

	// Split by dots and parse each part
	segments := strings.Split(version, ".")
	for i := 0; i < 3 && i < len(segments); i++ {
		// Extract numeric part (stop at first non-digit)
		numStr := ""
		for _, c := range segments[i] {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if numStr != "" {
			fmt.Sscanf(numStr, "%d", &parts[i])
		}
	}

	return parts
}
