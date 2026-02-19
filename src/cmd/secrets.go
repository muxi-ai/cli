package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/formation"
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
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all secrets",
	Long: `List all secret keys stored in secrets.enc.

By default, only shows secret names. Use --with-values to also display
the actual secret values (use with caution).

Use --remote to fetch secrets from a running Formation via the API
(values are masked for security).

Examples:
  # List secret names only (local)
  muxi secrets list

  # List secrets with values (local)
  muxi secrets list --with-values

  # List secrets from remote Formation
  muxi secrets list --remote

  # List secrets from specific formation
  muxi secrets list --remote -F my-formation`,
	Args: cobra.NoArgs,
	RunE: runSecretsList,
}

func runSecretsList(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")

	if remote {
		return runRemoteSecretsList(cmd, args)
	}
	return runLocalSecretsList(cmd, args)
}

func runLocalSecretsList(cmd *cobra.Command, args []string) error {
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
}

func runRemoteSecretsList(cmd *cobra.Command, args []string) error {
	client, err := formation.ClientFromFlags(cmd)
	if err != nil {
		return err
	}

	formation.PrintBadgeFromFlags(cmd)

	secretsResp, err := client.GetSecrets()
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	if secretsResp.Count == 0 {
		ui.Dimmed("No secrets configured")
		return nil
	}

	// Get sorted keys
	keys := make([]string, 0, len(secretsResp.Secrets))
	for k := range secretsResp.Secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find max key length for alignment
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	fmt.Println()
	ui.Bold(fmt.Sprintf("Secrets (%d):", secretsResp.Count))
	fmt.Println()
	for _, key := range keys {
		value := secretsResp.Secrets[key]
		fmt.Printf("  %-*s  %s\n", maxKeyLen, key, value)
	}

	return nil
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a secret value",
	Long: `Get a secret value by name.

Use --remote to fetch from a running Formation (value will be masked).

Examples:
  # Get local secret
  muxi secrets get OPENAI_API_KEY

  # Get remote secret (masked)
  muxi secrets get OPENAI_API_KEY --remote`,
	Args: RequireArgs(1),
	RunE: runSecretsGet,
}

func runSecretsGet(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")
	name := args[0]

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		formation.PrintBadgeFromFlags(cmd)

		secret, err := client.GetSecret(name)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		fmt.Println()
		fmt.Printf("  %s = %s\n", secret.Key, secret.Value)
		fmt.Println()
		return nil
	}

	// Local
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
	value, found := mgr.Get(normalizedName)
	if !found {
		return fmt.Errorf("secret '%s' not found", normalizedName)
	}

	fmt.Println(value)
	return nil
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <name> [value]",
	Short: "Set a secret value",
	Long: `Set or update a secret value.

If value is not provided, you will be prompted to enter it.
Secret names are automatically normalized to uppercase with underscores.

Use --remote to set on a running Formation.

Examples:
  # Set with prompt (recommended - value not in shell history)
  muxi secrets set OPENAI_API_KEY

  # Set directly (value visible in shell history)
  muxi secrets set OPENAI_API_KEY sk-...
  
  # Set on remote formation
  muxi secrets set OPENAI_API_KEY --remote`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSecretsSet,
}

func runSecretsSet(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")
	name := args[0]
	var value string

	if len(args) > 1 {
		value = args[1]
	} else {
		fmt.Printf("%s: ", name)
		fmt.Scanln(&value)
	}

	if value == "" {
		return fmt.Errorf("secret value cannot be empty")
	}

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		if err := client.SetSecret(name, value); err != nil {
			return fmt.Errorf("failed to set secret: %w", err)
		}

		formation.PrintBadgeFromFlags(cmd)
		ui.Success(fmt.Sprintf("Secret '%s' saved", name))
		return nil
	}

	// Local
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

	if normalizedName != name {
		ui.Dimmed(fmt.Sprintf("  (normalized: %s → %s)", name, normalizedName))
	}

	if err := mgr.Set(normalizedName, value, true); err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}

	ui.Success(fmt.Sprintf("Secret '%s' saved", normalizedName))
	return nil
}

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a secret",
	Long: `Delete a secret.

Use --remote to delete from a running Formation.

Examples:
  muxi secrets delete OPENAI_API_KEY
  muxi secrets delete OPENAI_API_KEY --remote`,
	Args: RequireArgs(1),
	RunE: runSecretsDelete,
}

func runSecretsDelete(cmd *cobra.Command, args []string) error {
	remote, _ := cmd.Flags().GetBool("remote")
	name := args[0]

	if remote {
		client, err := formation.ClientFromFlags(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteSecret(name); err != nil {
			return fmt.Errorf("failed to delete secret: %w", err)
		}

		formation.PrintBadgeFromFlags(cmd)
		ui.Success(fmt.Sprintf("Secret '%s' deleted", name))
		return nil
	}

	// Local
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

	// Scan formation files for actually referenced secrets
	referencedSecrets, scanErr := secrets.ScanFormationFiles(ctx.RootDir)
	referencedSet := make(map[string]bool)
	if scanErr == nil {
		for _, s := range referencedSecrets {
			referencedSet[s] = true
		}
	}

	// Get template keys
	templateKeys, err := mgr.GetTemplateKeys()
	if err != nil {
		return fmt.Errorf("failed to read secrets template: %w", err)
	}

	if len(templateKeys) == 0 {
		if len(referencedSecrets) > 0 {
			sort.Strings(referencedSecrets)
			ui.Warning("Secrets referenced in formation files but not in template:")
			for _, s := range referencedSecrets {
				fmt.Printf("  - %s\n", s)
			}
			fmt.Println()
			ui.Dimmed("Run 'muxi secrets sync' to update the template")
			return nil
		}

		ui.Dimmed("No secrets in template file")
		fmt.Println()
		ui.Dimmed("Add secrets to the 'secrets' file or use:")
		fmt.Printf("  %s\n", ui.Command("muxi secrets sync"))
		return nil
	}

	// Filter template keys to only those actually referenced in formation files
	// (removes stale entries like commented-out secrets)
	var activeKeys []string
	var staleKeys []string
	for _, key := range templateKeys {
		if len(referencedSet) == 0 || referencedSet[key] {
			activeKeys = append(activeKeys, key)
		} else {
			staleKeys = append(staleKeys, key)
		}
	}

	if len(staleKeys) > 0 {
		ui.Warning(fmt.Sprintf("%d secret(s) in template but not referenced in formation files:", len(staleKeys)))
		for _, s := range staleKeys {
			fmt.Printf("  - %s\n", s)
		}
		fmt.Println()
		ui.Dimmed("Run 'muxi secrets sync' to clean up the template")
		fmt.Println()
	}

	// Initialize manager (loads secrets.enc if exists)
	if err := mgr.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize secrets: %w", err)
	}

	// Find keys that need values (only active keys)
	var missingKeys []string
	for _, key := range activeKeys {
		if !mgr.Exists(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) == 0 {
		// Check for secrets referenced but not in template
		var notInTemplate []string
		templateSet := make(map[string]bool)
		for _, k := range templateKeys {
			templateSet[k] = true
		}
		for _, ref := range referencedSecrets {
			if !templateSet[ref] {
				notInTemplate = append(notInTemplate, ref)
			}
		}

		if len(notInTemplate) > 0 {
			sort.Strings(notInTemplate)
			ui.Warning("Some secrets are referenced but not in template:")
			for _, s := range notInTemplate {
				fmt.Printf("  - %s\n", s)
			}
			fmt.Println()
			ui.Dimmed("Run 'muxi secrets sync' to update the template")
			return nil
		}

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

	// Get current secrets from secrets.enc
	currentSecrets, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list current secrets: %w", err)
	}

	// Get current template keys
	templateKeys, err := mgr.GetTemplateKeys()
	if err != nil {
		return fmt.Errorf("failed to read secrets template: %w", err)
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

	templateSet := make(map[string]bool)
	for _, s := range templateKeys {
		templateSet[s] = true
	}

	// Find secrets to add (in formation files but not in template)
	var toAdd []string
	for _, s := range foundSecrets {
		if !templateSet[s] {
			toAdd = append(toAdd, s)
		}
	}

	// Find secrets to delete (in secrets.enc or template but not in formation files)
	deleteSet := make(map[string]bool)
	for _, s := range currentSecrets {
		if !foundSet[s] {
			deleteSet[s] = true
		}
	}
	for _, s := range templateKeys {
		if !foundSet[s] {
			deleteSet[s] = true
		}
	}
	var toDelete []string
	for s := range deleteSet {
		toDelete = append(toDelete, s)
	}
	sort.Strings(toDelete)

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
				// Delete from secrets.enc (may not exist there)
				if currentSet[name] {
					if _, err := mgr.Delete(name); err != nil {
						ui.Warning(fmt.Sprintf("  Failed to delete %s: %v", name, err))
					}
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
	secretsListCmd.Flags().Bool("remote", false, "Fetch secrets from remote Formation API")
	formation.AddCommonFlags(secretsListCmd)

	secretsGetCmd.Flags().Bool("remote", false, "Fetch from remote Formation API")
	formation.AddCommonFlags(secretsGetCmd)

	secretsSetCmd.Flags().Bool("remote", false, "Set on remote Formation")
	formation.AddCommonFlags(secretsSetCmd)

	secretsDeleteCmd.Flags().Bool("remote", false, "Delete from remote Formation")
	formation.AddCommonFlags(secretsDeleteCmd)

	secretsSetupCmd.Flags().Bool("dry-run", false, "Preview what would be prompted")
	secretsSyncCmd.Flags().BoolP("interactive", "i", false, "Confirm deletions interactively")
	secretsSyncCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	secretsSyncCmd.Flags().Bool("no-setup", false, "Skip prompting for values")

	// Add subcommands
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsGetCmd)
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)
	secretsCmd.AddCommand(secretsSetupCmd)
	secretsCmd.AddCommand(secretsSyncCmd)

	// Add to root
	rootCmd.AddCommand(secretsCmd)
}
