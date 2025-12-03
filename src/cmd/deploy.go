package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy formation to server",
	Long: `Deploy a formation to a MUXI server.

If the formation already exists on the server, it will be updated.
If it's a new formation, it will be created.`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("profile", "", "Server profile to use")
	deployCmd.Flags().Bool("dry-run", false, "Validate and create bundle without deploying")
}

// ExcludedPatterns are patterns to exclude from deploy bundles
var deployExcludedPatterns = []string{
	".git",
	".muxi",
	"secrets.enc",
	".key",
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

	// Read formation.yaml to get ID and version
	formationPath := filepath.Join(ctx.RootDir, "formation.yaml")
	formationData, err := os.ReadFile(formationPath)
	if err != nil {
		return fmt.Errorf("failed to read formation.yaml: %w", err)
	}

	var metadata FormationMetadata
	if err := yaml.Unmarshal(formationData, &metadata); err != nil {
		return fmt.Errorf("failed to parse formation.yaml: %w", err)
	}

	if metadata.ID == "" {
		ui.ErrorBlock(
			"Missing formation ID",
			"formation.yaml must have an 'id' field.",
			"",
		)
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

	// For updates, version is required
	if isUpdate && metadata.Version == "" {
		ui.ErrorBlock(
			"Version required for update",
			fmt.Sprintf("Formation '%s' already exists on server.\nAdd a 'version' field to formation.yaml before updating.", metadata.ID),
			"version: \"1.1.0\"",
		)
		os.Exit(1)
	}

	// Display deploy info
	fmt.Println()
	if isUpdate {
		fmt.Printf("  Updating: %s on %s\n", metadata.ID, serverName)
		if metadata.Version != "" {
			fmt.Printf("  Version:  %s\n", metadata.Version)
		}
	} else {
		fmt.Printf("  Deploying: %s to %s\n", metadata.ID, serverName)
		if metadata.Version != "" {
			fmt.Printf("  Version:   %s\n", metadata.Version)
		}
	}
	fmt.Println()

	// Create bundle
	spinner := ui.NewSpinner("Creating bundle")
	spinner.Start()

	bundlePath, fileCount, err := createTarGzBundle(ctx.RootDir, metadata.ID)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to create bundle: %w", err)
	}
	defer os.Remove(bundlePath)

	// Get bundle size
	bundleInfo, _ := os.Stat(bundlePath)
	bundleSize := bundleInfo.Size()

	spinner.Stop()
	fmt.Printf("  %s Created bundle (%d files, %s)\n", ui.GreenText("✓"), fileCount, formatBytes(bundleSize))

	if dryRun {
		fmt.Println()
		ui.Success("Dry run complete - bundle created but not uploaded")
		fmt.Println()
		return nil
	}

	// Upload and deploy
	var spinnerMsg string
	if isUpdate {
		spinnerMsg = "Deploying new version"
	} else {
		spinnerMsg = "Deploying formation"
	}
	spinner = ui.NewSpinner(spinnerMsg)
	spinner.Start()

	var deployErr error
	if isUpdate {
		deployErr = client.UpdateFormation(metadata.ID, bundlePath, metadata.Version)
	} else {
		deployErr = client.DeployFormation(metadata.ID, bundlePath, metadata.Version)
	}

	spinner.Stop()

	if deployErr != nil {
		fmt.Printf("  %s Deploy failed\n", ui.RedText("✗"))
		return deployErr
	}

	fmt.Printf("  %s Deployed\n", ui.GreenText("✓"))

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

	return nil
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
