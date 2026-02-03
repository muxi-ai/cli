package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:     "download [formation-id]",
	Short:   "Download formation from server",
	GroupID: "formation",
	Long: `Download a formation from a MUXI server to your local machine.

When inside a formation directory:
  muxi download           Download and replace current formation
  muxi download <other>   Error - go to that formation's directory

When NOT in a formation directory:
  muxi download <id>      Download to new ./<id>/ directory`,
	Example: `  # Inside formation directory - replace with server version
  cd my-bot
  muxi download

  # Outside formation directory - create new local copy
  muxi download my-bot
  muxi download my-bot --profile production`,
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("profile", "p", "", "Server profile to use")
	downloadCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	downloadCmd.Flags().Bool("include-db", false, "Include SQLite database files in download")
}

func runDownload(cmd *cobra.Command, args []string) error {
	profileFlag, _ := cmd.Flags().GetString("profile")
	force, _ := cmd.Flags().GetBool("force")
	includeDB, _ := cmd.Flags().GetBool("include-db")

	// Check if we're in a formation directory
	ctx, ctxErr := context.DetectFormation()
	inFormationDir := ctxErr == nil

	if inFormationDir {
		return downloadInFormationDir(ctx, profileFlag, force, includeDB, args)
	}
	return downloadOutsideFormationDir(profileFlag, force, includeDB, args)
}

func downloadInFormationDir(ctx *context.FormationContext, profileFlag string, force bool, includeDB bool, args []string) error {
	// If user specified a different formation ID, show error
	if len(args) > 0 && args[0] != ctx.ID {
		ui.ErrorBlock(
			"Cannot download different formation here",
			fmt.Sprintf("You're in the '%s' formation directory.", ctx.ID),
			fmt.Sprintf("To download '%s', either:\n  cd ../%s && muxi download\n  cd /path/to/empty/dir && muxi download %s", args[0], args[0], args[0]),
		)
		return nil
	}

	// Resolve profile
	profile := formation.ResolveProfile(profileFlag)
	if profile == "" {
		return fmt.Errorf("no server profile configured - use --profile or 'muxi set default profile'")
	}

	// Create server client
	client, err := server.NewClient(profile)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// Confirm overwrite
	if !force {
		fmt.Println()
		ui.Warning(fmt.Sprintf("This will replace the entire '%s' directory with the server version.", ctx.ID))
		fmt.Println()
		
		confirmed, err := wizard.PromptConfirm("Download and replace?", false)
		if err != nil || !confirmed {
			ui.Dimmed("Cancelled")
			return nil
		}
	}

	// Download
	fmt.Println()
	ui.Step("Downloading from server...")

	zipPath, err := client.DownloadFormation(ctx.ID, includeDB)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(zipPath)

	// Clear directory (except .git, .muxi, and DB files if not including them)
	if err := clearFormationDir(ctx.RootDir, !includeDB); err != nil {
		return fmt.Errorf("failed to clear directory: %w", err)
	}

	// Extract
	ui.Step("Extracting...")
	if err := extractZip(zipPath, ctx.RootDir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Downloaded '%s' from server", ctx.ID))
	return nil
}

func downloadOutsideFormationDir(profileFlag string, force bool, includeDB bool, args []string) error {
	// Require formation ID
	if len(args) == 0 {
		ui.ErrorBlock(
			"Formation ID required",
			"Specify the formation to download.",
			"muxi download my-bot",
		)
		return nil
	}

	formationID := args[0]

	// Resolve profile (flag > global default)
	profile := profileFlag
	if profile == "" {
		profile = server.GetDefaultProfile()
	}
	if profile == "" {
		return fmt.Errorf("no server profile configured - use --profile or 'muxi profiles add'")
	}

	// Check if target directory already exists
	targetDir := filepath.Join(".", formationID)
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		// Directory exists - check if it's a formation
		if _, found := context.FindFormationFile(targetDir); found {
			ui.ErrorBlock(
				"Formation directory already exists",
				fmt.Sprintf("'%s' already exists and contains a formation.", formationID),
				fmt.Sprintf("cd %s && muxi download", formationID),
			)
			return nil
		}
		// Directory exists but not a formation - still warn
		if !force {
			fmt.Println()
			ui.Warning(fmt.Sprintf("Directory '%s' already exists and will be overwritten.", formationID))
			fmt.Println()
			
			confirmed, err := wizard.PromptConfirm("Continue?", false)
			if err != nil || !confirmed {
				ui.Dimmed("Cancelled")
				return nil
			}
		}
	}

	// Create server client
	client, err := server.NewClient(profile)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// Show confirmation
	if !force {
		fmt.Println()
		fmt.Printf("This will download '%s' from server '%s' to ./%s/\n", formationID, profile, formationID)
		fmt.Println()
		
		confirmed, err := wizard.PromptConfirm("Continue?", true)
		if err != nil || !confirmed {
			ui.Dimmed("Cancelled")
			return nil
		}
	}

	// Download
	fmt.Println()
	ui.Step("Downloading from server...")

	zipPath, err := client.DownloadFormation(formationID, includeDB)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(zipPath)

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract
	ui.Step("Extracting...")
	if err := extractZip(zipPath, targetDir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Downloaded '%s' to ./%s/", formationID, formationID))
	ui.Dimmed(fmt.Sprintf("  cd %s && muxi dev", formationID))
	return nil
}

// clearFormationDir removes all files except .git, .muxi, and optionally memory.db
func clearFormationDir(dir string, preserveDB bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	preserve := map[string]bool{
		".git":  true,
		".muxi": true,
	}
	
	if preserveDB {
		preserve["memory.db"] = true
	}

	for _, entry := range entries {
		if preserve[entry.Name()] {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

// extractZip extracts a zip file to a directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Security: prevent zip slip
		destPath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
