package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/muxi-ai/cli/pkg/defaults"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
)

var formationsCmd = &cobra.Command{
	Use:   "formations",
	Short: "Manage saved formation configs",
	GroupID: "server",
	Long: `Add, list, and manage saved formation configurations.

Formations are saved to ~/.muxi/cli/formations.yaml and can be referenced
using the -f (--formation) flag from anywhere, without being in a formation directory.

Each saved formation stores:
- Default server profile to use
- Default user ID for user-scoped operations  
- Client key for Formation API access
- Admin key (optional) for admin operations`,
}

var formationsAddCmd = &cobra.Command{
	Use:   "add [id]",
	Short: "Add a formation config",
	Long: `Add a new formation configuration interactively.

Prompts for server profile, user ID, and API keys.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFormationsAdd,
}

var formationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List saved formations",
	Long: `List all saved formation configurations.

Shows formation name, profile, and user ID.`,
	RunE: runFormationsList,
}

var formationsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show formation config",
	Long: `Show detailed configuration for a saved formation.

Displays profile, user ID, and masked API keys.`,
	Args: RequireArgs(1),
	RunE: runFormationsShow,
}

var formationsRemoveCmd = &cobra.Command{
	Use:     "remove <id>",
	Aliases: []string{"rm"},
	Short:   "Remove a formation config",
	Long:    `Remove a saved formation configuration.`,
	Args:    RequireArgs(1),
	RunE:    runFormationsRemove,
}

func init() {
	rootCmd.AddCommand(formationsCmd)

	formationsCmd.AddCommand(formationsAddCmd)
	formationsCmd.AddCommand(formationsListCmd)
	formationsCmd.AddCommand(formationsShowCmd)
	formationsCmd.AddCommand(formationsRemoveCmd)
}

// runFormationsAdd handles muxi formations add [name]
func runFormationsAdd(cmd *cobra.Command, args []string) error {
	fmt.Println()

	// Get name from arg or prompt
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		var err error
		name, err = wizard.PromptString("Formation ID", "", nil)
		if err != nil {
			return err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("formation ID cannot be empty")
		}
	}

	// Normalize: lowercase, spaces to hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	// Check if exists
	if defaults.FormationExists(name) {
		return fmt.Errorf("formation '%s' already exists\n\nUse %s to update or %s to remove first",
			name, ui.Command("muxi formations show "+name), ui.Command("muxi formations remove "+name))
	}

	// Get available profiles
	profilesConfig, err := server.LoadProfiles()
	if err != nil {
		return err
	}

	if len(profilesConfig.Profiles) == 0 {
		ui.ErrorBlock(
			"No server profiles configured",
			"You need at least one server profile to add a formation.",
			ui.Command("muxi profiles add"),
		)
		return nil
	}

	// Build profile options
	var profileOptions []wizard.SelectOption
	defaultIdx := 0
	i := 0
	for pname := range profilesConfig.Profiles {
		label := pname
		if pname == profilesConfig.DefaultProfile {
			label = pname + " (default)"
			defaultIdx = i
		}
		profileOptions = append(profileOptions, wizard.SelectOption{
			Value: pname,
			Label: label,
		})
		i++
	}

	// Select profile
	selectedProfile, err := wizard.PromptSelect("Server profile", profileOptions, defaultIdx)
	if err != nil {
		return err
	}

	ui.Dimmed("  Enter API keys for this formation")
	fmt.Println()

	// Client key (required)
	clientKey, err := wizard.PromptPassword("Client key", false)
	if err != nil {
		return err
	}
	clientKey = strings.TrimSpace(clientKey)
	if clientKey == "" {
		return fmt.Errorf("client key is required")
	}

	// Admin key (optional)
	adminKey, err := wizard.PromptPassword("Admin key (optional, press Enter to skip)", true)
	if err != nil {
		return err
	}
	adminKey = strings.TrimSpace(adminKey)

	// User ID (optional) - asked last
	defaultUserID := defaults.GetUserID()
	userID, err := wizard.PromptString("Default user ID", defaultUserID, nil)
	if err != nil {
		return err
	}
	userID = strings.TrimSpace(userID)

	// Save
	entry := defaults.FormationEntry{
		DefaultProfile: selectedProfile,
		DefaultUserID:  userID,
		ClientKey:      clientKey,
		AdminKey:       adminKey,
		AddedAt:        time.Now(),
	}

	if err := defaults.AddFormation(name, entry); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Formation \"%s\" added", name))
	fmt.Println()
	fmt.Printf("  You can now use: %s\n", ui.Command("muxi sessions list -f "+name))
	fmt.Println()

	return nil
}

// runFormationsList handles muxi formations list
func runFormationsList(cmd *cobra.Command, args []string) error {
	config, err := defaults.LoadFormations()
	if err != nil {
		return err
	}

	if len(config.Formations) == 0 {
		fmt.Println()
		ui.Dimmed("  No formations saved")
		fmt.Println()
		fmt.Printf("  Add a formation: %s\n", ui.Command("muxi formations add"))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Printf("  %-20s %-15s %-15s\n", "NAME", "PROFILE", "USER")
	fmt.Printf("  %-20s %-15s %-15s\n", "────────────────────", "───────────────", "───────────────")

	for name, entry := range config.Formations {
		userDisplay := entry.DefaultUserID
		if userDisplay == "" {
			userDisplay = ui.DimmedText("-")
		}

		fmt.Printf("  %-20s %-15s %-15s\n", name, entry.DefaultProfile, userDisplay)
	}

	fmt.Println()

	return nil
}

// runFormationsShow handles muxi formations show <name>
func runFormationsShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	entry, err := defaults.GetFormation(name)
	if err != nil {
		ui.ErrorBlock(
			"Formation not found",
			fmt.Sprintf("Formation '%s' is not saved.", name),
			ui.Command("muxi formations list"),
		)
		return nil
	}

	fmt.Println()
	fmt.Printf("  Formation: %s\n", name)
	fmt.Println()
	fmt.Printf("  Profile:    %s\n", entry.DefaultProfile)

	if entry.DefaultUserID != "" {
		fmt.Printf("  User ID:    %s\n", entry.DefaultUserID)
	} else {
		fmt.Printf("  User ID:    %s\n", ui.DimmedText("-"))
	}

	fmt.Println()

	// Masked keys
	fmt.Printf("  Client Key: %s\n", maskKey(entry.ClientKey))
	if entry.AdminKey != "" {
		fmt.Printf("  Admin Key:  %s\n", maskKey(entry.AdminKey))
	} else {
		fmt.Printf("  Admin Key:  %s\n", ui.DimmedText("-"))
	}

	fmt.Println()
	fmt.Printf("  Added:      %s\n", entry.AddedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	return nil
}

// runFormationsRemove handles muxi formations remove <name>
func runFormationsRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	if !defaults.FormationExists(name) {
		ui.ErrorBlock(
			"Formation not found",
			fmt.Sprintf("Formation '%s' is not saved.", name),
			ui.Command("muxi formations list"),
		)
		return nil
	}

	// Confirm
	fmt.Println()
	confirm, err := wizard.PromptConfirm(fmt.Sprintf("Remove formation \"%s\"?", name), false)
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println()
		ui.Dimmed("  Cancelled")
		fmt.Println()
		return nil
	}

	if err := defaults.RemoveFormation(name); err != nil {
		return err
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Formation \"%s\" removed", name))
	fmt.Println()

	return nil
}

// maskKey masks an API key for display (shows first 4 and last 4 chars)
func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
