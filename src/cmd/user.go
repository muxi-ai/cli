package cmd

import (
	"fmt"

	"github.com/muxi-ai/cli/pkg/defaults"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user settings",
	Long:  `Manage default user ID for Formation API commands.`,
}

var userDefaultCmd = &cobra.Command{
	Use:   "default [user_id]",
	Short: "Set global default user ID",
	Long: `Set the global default user ID for Formation API commands.

This is saved to ~/.muxi/cli/defaults.yaml and used by commands like:
chat, sessions, history, clear, jobs, etc.

Per-formation override: Use 'muxi set user' inside a formation directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserDefault,
}

var userShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current default user ID",
	RunE:  runUserShow,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userDefaultCmd)
	userCmd.AddCommand(userShowCmd)
}

// runUserDefault handles muxi user default [user_id]
func runUserDefault(cmd *cobra.Command, args []string) error {
	var userID string
	var err error

	if len(args) > 0 {
		userID = args[0]
	} else {
		// Interactive mode
		ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Set Global Default User ID                          MUXI │
│──────────────────────────────────────────────────────────────│
│ Set the default user ID for all Formation API commands.      │
│ Override per-formation with: muxi set user                   │
╰──────────────────────────────────────────────────────────────╯`)

		fmt.Println()

		// Show current value if set
		currentDefault := defaults.GetUserID()
		defaultValue := currentDefault
		if defaultValue == "" {
			defaultValue = "default-user"
		}

		if currentDefault != "" {
			ui.Dimmed(fmt.Sprintf("  Current default: %s", currentDefault))
			fmt.Println()
		}

		userID, err = wizard.PromptString("Enter user ID", defaultValue, nil)
		if err != nil {
			return err
		}
	}

	// Save
	if err := defaults.SetUserID(userID); err != nil {
		return fmt.Errorf("failed to save default: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Global default user ID set to '%s'", userID))
	fmt.Println()
	ui.Dimmed("  Override per-formation with: muxi set user")
	fmt.Println()

	return nil
}

// runUserShow handles muxi user show
func runUserShow(cmd *cobra.Command, args []string) error {
	globalDefault := defaults.GetUserID()

	fmt.Println()
	fmt.Println("  Default User ID:")
	fmt.Println()

	if globalDefault != "" {
		fmt.Printf("    Global:     %s\n", ui.CyanText(globalDefault))
	} else {
		fmt.Printf("    Global:     %s\n", ui.DimmedText("(not set)"))
	}

	// Check for formation-level override
	dotMuxi, err := loadDotMuxi()
	if err == nil && dotMuxi.UserID != "" {
		fmt.Printf("    Formation:  %s %s\n", ui.CyanText(dotMuxi.UserID), ui.DimmedText("(override)"))
	}

	fmt.Println()
	ui.Dimmed("  Set global:     muxi user default <user_id>")
	ui.Dimmed("  Set formation:  muxi set user <user_id>")
	fmt.Println()

	return nil
}
