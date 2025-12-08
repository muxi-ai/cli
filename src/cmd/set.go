package cmd

import (
	"fmt"
	"os"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/registry"
	"github.com/muxi-ai/cli/pkg/server"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// DotMuxi represents the .muxi file structure
type DotMuxi struct {
	Profile  string `yaml:"profile,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	UserID   string `yaml:"user_id,omitempty"`
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set formation-level defaults",
	Long:  `Set default server or registry for this formation (saved to .muxi file).`,
}

var setServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Set default server for this formation",
	RunE:  runSetServer,
}

var setRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Set default registry for this formation",
	RunE:  runSetRegistry,
}

var setUserCmd = &cobra.Command{
	Use:   "user [user_id]",
	Short: "Set default user ID for this formation",
	Long:  `Set the default user ID for Formation API commands (chat, sessions, etc.).`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSetUser,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setServerCmd)
	setCmd.AddCommand(setRegistryCmd)
	setCmd.AddCommand(setUserCmd)
}

// loadDotMuxi loads the .muxi file from the current directory
func loadDotMuxi() (*DotMuxi, error) {
	data, err := os.ReadFile(".muxi")
	if err != nil {
		if os.IsNotExist(err) {
			return &DotMuxi{}, nil
		}
		return nil, err
	}

	var config DotMuxi
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveDotMuxi saves the .muxi file to the current directory
func saveDotMuxi(config *DotMuxi) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(".muxi", data, 0644)
}

// checkFormationDir checks if we're in a formation directory
func checkFormationDir() bool {
	_, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"Run this command from inside a formation directory.",
			"cd my-formation && muxi set server",
		)
		return false
	}
	return true
}

// runSetServer handles muxi set server
func runSetServer(cmd *cobra.Command, args []string) error {
	if !checkFormationDir() {
		os.Exit(1)
	}

	// Show banner
	ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Set Default Server                                  MUXI │
│──────────────────────────────────────────────────────────────│
│ Set the default server profile for this formation.           │
│ This is saved to .muxi and overrides the global default.     │
╰──────────────────────────────────────────────────────────────╯`)

	// Load servers
	config, err := server.LoadServers()
	if err != nil {
		return err
	}

	if len(config.Servers) == 0 {
		fmt.Println()
		ui.Dimmed("  No servers configured")
		fmt.Println()
		fmt.Printf("  Add a server first: %s\n", ui.Command("muxi server add"))
		fmt.Println()
		return nil
	}

	// Load current .muxi
	dotMuxi, err := loadDotMuxi()
	if err != nil {
		return fmt.Errorf("failed to load .muxi: %w", err)
	}

	// Build options
	var options []wizard.SelectOption
	currentIndex := 0
	i := 0
	for name := range config.Servers {
		label := name
		if name == dotMuxi.Profile {
			label += " [current]"
			currentIndex = i
		}
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: label,
		})
		i++
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select server for this formation", options, currentIndex)
	if err != nil {
		return err
	}

	// Save
	dotMuxi.Profile = selected
	if err := saveDotMuxi(dotMuxi); err != nil {
		return fmt.Errorf("failed to save .muxi: %w", err)
	}

	fmt.Println()
	ui.Success("Saved to .muxi")
	fmt.Println()

	return nil
}

// runSetRegistry handles muxi set registry
func runSetRegistry(cmd *cobra.Command, args []string) error {
	if !checkFormationDir() {
		os.Exit(1)
	}

	// Show banner
	ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Set Default Registry                                MUXI │
│──────────────────────────────────────────────────────────────│
│ Set the default registry for this formation.                 │
│ This is saved to .muxi and overrides the global default.     │
╰──────────────────────────────────────────────────────────────╯`)

	// Load registries
	config, err := registry.LoadRegistries()
	if err != nil {
		return err
	}

	if len(config.Registries) == 0 {
		fmt.Println()
		ui.Dimmed("  No registries configured")
		fmt.Println()
		fmt.Printf("  Add a registry first: %s\n", ui.Command("muxi registry add"))
		fmt.Println()
		return nil
	}

	// Load current .muxi
	dotMuxi, err := loadDotMuxi()
	if err != nil {
		return fmt.Errorf("failed to load .muxi: %w", err)
	}

	// Build options
	var options []wizard.SelectOption
	currentIndex := 0
	i := 0
	for name := range config.Registries {
		label := name
		if name == dotMuxi.Registry {
			label += " [current]"
			currentIndex = i
		}
		options = append(options, wizard.SelectOption{
			Value: name,
			Label: label,
		})
		i++
	}

	fmt.Println()
	selected, err := wizard.PromptSelect("Select registry for this formation", options, currentIndex)
	if err != nil {
		return err
	}

	// Save
	dotMuxi.Registry = selected
	if err := saveDotMuxi(dotMuxi); err != nil {
		return fmt.Errorf("failed to save .muxi: %w", err)
	}

	fmt.Println()
	ui.Success("Saved to .muxi")
	fmt.Println()

	return nil
}

// runSetUser handles muxi set user
func runSetUser(cmd *cobra.Command, args []string) error {
	if !checkFormationDir() {
		os.Exit(1)
	}

	// Load current .muxi
	dotMuxi, err := loadDotMuxi()
	if err != nil {
		return fmt.Errorf("failed to load .muxi: %w", err)
	}

	var userID string

	if len(args) > 0 {
		// User ID provided as argument
		userID = args[0]
	} else {
		// Show banner for interactive mode
		ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Set Default User ID                                 MUXI │
│──────────────────────────────────────────────────────────────│
│ Set the default user ID for Formation API commands.          │
│ Used by: chat, sessions, history, clear, jobs, etc.          │
╰──────────────────────────────────────────────────────────────╯`)

		fmt.Println()
		
		// Show current value if set
		defaultValue := dotMuxi.UserID
		if defaultValue == "" {
			defaultValue = "default-user"
		}
		
		userID, err = wizard.PromptString("Enter user ID", defaultValue, nil)
		if err != nil {
			return err
		}
	}

	// Save
	dotMuxi.UserID = userID
	if err := saveDotMuxi(dotMuxi); err != nil {
		return fmt.Errorf("failed to save .muxi: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Default user ID set to '%s'", userID))
	fmt.Println()

	return nil
}
