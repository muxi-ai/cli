package cmd

import (
	"fmt"
	"sort"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/secrets"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage formation secrets",
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
			ui.Dimmed("  muxi secrets set SECRET_NAME")
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

func init() {
	// Add flags
	secretsListCmd.Flags().Bool("with-values", false, "Show secret values (use with caution)")

	// Add subcommands
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)

	// Add to root
	rootCmd.AddCommand(secretsCmd)
}
