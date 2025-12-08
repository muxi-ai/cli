package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:     "secrets",
	Short:   "Manage formation secrets",
	GroupID: "formation",
	Long: `Manage encrypted secrets for a formation.

Secrets are stored in secrets.enc (encrypted) and can be referenced
in formation.yaml using ${{ secrets.SECRET_NAME }} syntax.

Must be run inside a formation directory.`,
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Long: `List all secret keys stored in secrets.enc.

By default, only shows secret names. Use --with-values to also display
the actual secret values (use with caution).

Examples:
  # List secret names only
  muxi secrets list

  # List secrets with values
  muxi secrets list --with-values`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		withValues, _ := cmd.Flags().GetBool("with-values")

		ctx, err := context.MustDetectFormation()
		if err != nil {
			ui.ErrorBlock(
				"Not in formation directory",
				"This command must be run inside a formation directory.",
				"Navigate to your formation:\n  cd my-formation",
			)
			return nil
		}

		mgr := secrets.NewManager(ctx.RootDir)
		if err := mgr.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize secrets: %w", err)
		}

		secretsList, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		if len(secretsList) == 0 {
			ui.Dimmed("No secrets stored")
			fmt.Println()
			ui.Dimmed("Add secrets with:")
			fmt.Printf("  %s\n", ui.Command("muxi secrets set SECRET_NAME"))
			return nil
		}

		// Sort alphabetically
		sort.Strings(secretsList)

		fmt.Println()
		if withValues {
			ui.Bold(fmt.Sprintf("Secrets (%d):", len(secretsList)))
			fmt.Println()
			for _, name := range secretsList {
				value, _ := mgr.Get(name)
				fmt.Printf("  %s = %s\n", name, value)
			}
		} else {
			ui.Bold(fmt.Sprintf("Secrets (%d):", len(secretsList)))
			fmt.Println()
			for _, name := range secretsList {
				fmt.Printf("  %s\n", name)
			}
			fmt.Println()
			ui.Dimmed("Use --with-values to show secret values")
		}

		return nil
	},
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <name> [value]",
	Short: "Set a secret value",
	Long: `Set or update a secret value.

If value is not provided, you will be prompted to enter it.
Secret names are automatically normalized to uppercase with underscores.

Examples:
  # Set with prompt (recommended - value not in shell history)
  muxi secrets set OPENAI_API_KEY

  # Set directly (value visible in shell history)
  muxi secrets set OPENAI_API_KEY sk-...`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var value string

		ctx, err := context.MustDetectFormation()
		if err != nil {
			ui.ErrorBlock(
				"Not in formation directory",
				"This command must be run inside a formation directory.",
				"Navigate to your formation:\n  cd my-formation",
			)
			return nil
		}

		mgr := secrets.NewManager(ctx.RootDir)
		if err := mgr.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize secrets: %w", err)
		}

		normalizedName := secrets.NormalizeName(name)

		// Tell user if name was normalized
		if normalizedName != name {
			ui.Dimmed(fmt.Sprintf("  (normalized: %s → %s)", name, normalizedName))
		}

		if len(args) > 1 {
			value = args[1]
		} else {
			// Prompt for value
			fmt.Printf("%s: ", normalizedName)
			fmt.Scanln(&value)
		}

		if value == "" {
			return fmt.Errorf("secret value cannot be empty")
		}

		if err := mgr.Set(normalizedName, value, true); err != nil {
			return fmt.Errorf("failed to set secret: %w", err)
		}

		ui.Success(fmt.Sprintf("Secret '%s' saved", normalizedName))
		return nil
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a secret",
	Long: `Delete a secret from secrets.enc.

Examples:
  muxi secrets delete OPENAI_API_KEY`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		ctx, err := context.MustDetectFormation()
		if err != nil {
			ui.ErrorBlock(
				"Not in formation directory",
				"This command must be run inside a formation directory.",
				"Navigate to your formation:\n  cd my-formation",
			)
			return nil
		}

		mgr := secrets.NewManager(ctx.RootDir)
		if err := mgr.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize secrets: %w", err)
		}

		normalizedName := secrets.NormalizeName(name)

		// Tell user if name was normalized
		if normalizedName != name {
			ui.Dimmed(fmt.Sprintf("  (normalized: %s → %s)", name, normalizedName))
		}

		deleted, err := mgr.Delete(normalizedName)
		if err != nil {
			return fmt.Errorf("failed to delete secret: %w", err)
		}

		if !deleted {
			ui.Warning(fmt.Sprintf("Secret '%s' not found", normalizedName))
			return nil
		}

		ui.Success(fmt.Sprintf("Secret '%s' deleted", normalizedName))
		return nil
	},
}

var secretsSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up secrets from template",
	Long: `Populate secrets.enc from the secrets template file.

This command is used after pulling a formation from a registry.
It prompts for values for each key in the secrets template that
doesn't have a value in secrets.enc yet.

Examples:
  # Interactive setup - prompts for each missing secret
  muxi secrets setup

  # Preview what would be prompted
  muxi secrets setup --dry-run`,
	Args: cobra.NoArgs,
	RunE: runSecretsSetup,
}

var secretsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize secrets with formation files",
	Long: `Scan formation files and synchronize secrets.

This command scans all formation files for ${{ secrets.XXX }} patterns,
updates the secrets template, removes unused secrets, and prompts for
any missing values.

Flow:
1. Scan formation files for secret references
2. Add new secrets to template (will prompt via setup)
3. Delete unused secrets from both template and secrets.enc
4. Run setup to prompt for empty values

Examples:
  # Sync secrets with formation files
  muxi secrets sync

  # Interactive mode - confirm each deletion
  muxi secrets sync -i

  # Preview changes without applying
  muxi secrets sync --dry-run

  # Sync without prompting for values
  muxi secrets sync --no-setup`,
	Args: cobra.NoArgs,
	RunE: runSecretsSync,
}

func runSecretsSetup(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation",
		)
		return nil
	}

	mgr := secrets.NewManager(ctx.RootDir)

	// Get template keys
	templateKeys, err := mgr.GetTemplateKeys()
	if err != nil {
		return fmt.Errorf("failed to read secrets template: %w", err)
	}

	if len(templateKeys) == 0 {
		ui.Dimmed("No secrets in template file")
		fmt.Println()
		ui.Dimmed("Add secrets to the 'secrets' file or use:")
		fmt.Printf("  %s\n", ui.Command("muxi secrets sync"))
		return nil
	}

	// Initialize manager (loads secrets.enc if exists)
	if err := mgr.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Find keys that need values
	var missingKeys []string
	for _, key := range templateKeys {
		if !mgr.Exists(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) == 0 {
		ui.Success("All secrets are configured")
		return nil
	}

	sort.Strings(missingKeys)

	if dryRun {
		fmt.Println()
		ui.Bold(fmt.Sprintf("Would prompt for %d secret(s):", len(missingKeys)))
		fmt.Println()
		for _, key := range missingKeys {
			fmt.Printf("  %s\n", key)
		}
		return nil
	}

	fmt.Println()
	ui.Bold(fmt.Sprintf("Setting up %d secret(s):", len(missingKeys)))
	fmt.Println()

	configured := 0
	for _, key := range missingKeys {
		fmt.Printf("  %s: ", key)
		var value string
		fmt.Scanln(&value)

		if value == "" {
			ui.Dimmed(fmt.Sprintf("    Skipped %s (empty)", key))
			continue
		}

		if err := mgr.Set(key, value, true); err != nil {
			ui.Warning(fmt.Sprintf("    Failed to save %s: %v", key, err))
			continue
		}

		ui.Success(fmt.Sprintf("    %s saved", key))
		configured++
	}

	fmt.Println()
	if configured > 0 {
		ui.Success(fmt.Sprintf("%d secret(s) configured", configured))
	}
	if configured < len(missingKeys) {
		ui.Dimmed(fmt.Sprintf("%d secret(s) skipped", len(missingKeys)-configured))
	}

	return nil
}

func runSecretsSync(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	interactive, _ := cmd.Flags().GetBool("interactive")
	noSetup, _ := cmd.Flags().GetBool("no-setup")

	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation",
		)
		return nil
	}

	mgr := secrets.NewManager(ctx.RootDir)
	if err := mgr.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Scan formation files
	ui.Step("Scanning formation files...")
	foundSecrets, err := secrets.ScanFormationFiles(ctx.RootDir)
	if err != nil {
		return fmt.Errorf("failed to scan formation files: %w", err)
	}

	sort.Strings(foundSecrets)

	if len(foundSecrets) == 0 {
		ui.Dimmed("No secrets referenced in formation files")
		return nil
	}

	fmt.Printf("  Found %d secret(s)\n", len(foundSecrets))
	fmt.Println()

	// Get current secrets
	currentSecrets, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list current secrets: %w", err)
	}

	// Build sets for comparison
	foundSet := make(map[string]bool)
	for _, s := range foundSecrets {
		foundSet[s] = true
	}

	currentSet := make(map[string]bool)
	for _, s := range currentSecrets {
		currentSet[s] = true
	}

	// Find secrets to add (in formation files but not in template)
	var toAdd []string
	for _, s := range foundSecrets {
		if !currentSet[s] {
			toAdd = append(toAdd, s)
		}
	}

	// Find secrets to delete (in secrets.enc but not in formation files)
	var toDelete []string
	for _, s := range currentSecrets {
		if !foundSet[s] {
			toDelete = append(toDelete, s)
		}
	}

	// Show changes
	hasChanges := len(toAdd) > 0 || len(toDelete) > 0

	if !hasChanges {
		ui.Success("Secrets are in sync")
		if !noSetup {
			// Still run setup in case there are empty values
			return runSecretsSetup(cmd, args)
		}
		return nil
	}

	// Process additions (add to template only)
	if len(toAdd) > 0 {
		for _, name := range toAdd {
			if dryRun {
				fmt.Printf("  %s %s (would add to template)\n", ui.GreenText("+"), name)
			} else {
				if err := mgr.AddToTemplate(name); err != nil {
					ui.Warning(fmt.Sprintf("  Failed to add %s to template: %v", name, err))
				} else {
					fmt.Printf("  %s %s (added to template)\n", ui.GreenText("+"), name)
				}
			}
		}
	}

	// Process deletions (delete from both template and secrets.enc)
	if len(toDelete) > 0 {
		if interactive && !dryRun {
			fmt.Println()
			ui.Bold("The following secrets are no longer referenced:")
			for _, name := range toDelete {
				fmt.Printf("  %s %s\n", ui.RedText("-"), name)
			}
			fmt.Println()
			fmt.Print("Delete these secrets? (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				ui.Dimmed("Skipped deletion")
				toDelete = nil
			}
		}

		for _, name := range toDelete {
			if dryRun {
				fmt.Printf("  %s %s (would delete)\n", ui.RedText("-"), name)
			} else {
				// Delete from secrets.enc
				if _, err := mgr.Delete(name); err != nil {
					ui.Warning(fmt.Sprintf("  Failed to delete %s: %v", name, err))
					continue
				}
				// Delete from template
				if err := mgr.DeleteFromTemplate(name); err != nil {
					ui.Warning(fmt.Sprintf("  Failed to remove %s from template: %v", name, err))
				}
				fmt.Printf("  %s %s (deleted)\n", ui.RedText("-"), name)
			}
		}
	}

	if dryRun {
		fmt.Println()
		ui.Dimmed("Dry run - no changes made")
		return nil
	}

	fmt.Println()
	ui.Success("Secrets synchronized")

	// Run setup for missing values (prompt for newly added secrets)
	if !noSetup && len(toAdd) > 0 {
		fmt.Println()
		return runSecretsSetup(cmd, args)
	}

	return nil
}

func init() {
	// Add flags
	secretsListCmd.Flags().Bool("with-values", false, "Show secret values (use with caution)")
	secretsSetupCmd.Flags().Bool("dry-run", false, "Preview what would be prompted")
	secretsSyncCmd.Flags().BoolP("interactive", "i", false, "Confirm deletions interactively")
	secretsSyncCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	secretsSyncCmd.Flags().Bool("no-setup", false, "Skip prompting for values")

	// Add subcommands
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)
	secretsCmd.AddCommand(secretsSetupCmd)
	secretsCmd.AddCommand(secretsSyncCmd)

	// Add to root
	rootCmd.AddCommand(secretsCmd)
}
