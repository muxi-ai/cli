package cmd

import (
	"fmt"
	"os"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/defaults"
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
	Use:     "set",
	Short:   "Set defaults",
	GroupID: "config",
	Long:    `Set default server, registry, or user ID (local or global).`,
}

var setDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Set default values",
	Long: `Set default server, registry, or user ID.

In a formation directory: prompts for local (this formation) or global.
Outside formation directory: sets global default.

Use --local or --global flags to skip the prompt.`,
}

var setDefaultProfileCmd = &cobra.Command{
	Use:     "profile [name]",
	Aliases: []string{"server"},
	Short:   "Set default profile",
	Long:    `Set the default server profile for deployments and commands.`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runSetDefaultProfile,
}

var setDefaultRegistryCmd = &cobra.Command{
	Use:   "registry [name]",
	Short: "Set default registry",
	Long:  `Set the default registry for push/pull commands.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSetDefaultRegistry,
}

var setDefaultUserCmd = &cobra.Command{
	Use:   "user [user_id]",
	Short: "Set default user ID",
	Long:  `Set the default user ID for Formation API commands (chat, sessions, etc.).`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSetDefaultUser,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setDefaultCmd)

	// Add subcommands to 'set default'
	setDefaultCmd.AddCommand(setDefaultProfileCmd)
	setDefaultCmd.AddCommand(setDefaultRegistryCmd)
	setDefaultCmd.AddCommand(setDefaultUserCmd)

	// Add --local and --global flags to each
	for _, cmd := range []*cobra.Command{setDefaultProfileCmd, setDefaultRegistryCmd, setDefaultUserCmd} {
		cmd.Flags().BoolP("local", "l", false, "Set for this formation only (requires formation directory)")
		cmd.Flags().BoolP("global", "g", false, "Set as global default")
	}
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

// isInFormationDir checks if we're in a formation directory
func isInFormationDir() bool {
	_, err := context.MustDetectFormation()
	return err == nil
}

// determineScope returns "local" or "global" based on flags and context
func determineScope(cmd *cobra.Command) (string, error) {
	local, _ := cmd.Flags().GetBool("local")
	global, _ := cmd.Flags().GetBool("global")

	if local && global {
		return "", fmt.Errorf("cannot specify both --local and --global")
	}

	if local {
		if !isInFormationDir() {
			return "", fmt.Errorf("--local requires being in a formation directory")
		}
		return "local", nil
	}

	if global {
		return "global", nil
	}

	// No flag specified - check context
	if !isInFormationDir() {
		// Not in formation dir, default to global
		return "global", nil
	}

	// In formation dir, prompt
	fmt.Println()
	options := []wizard.SelectOption{
		{Value: "local", Label: "Local (this formation only, saved to .muxi)"},
		{Value: "global", Label: "Global (all formations, saved to ~/.muxi/cli/)"},
	}
	scope, err := wizard.PromptSelect("Apply to", options, 0)
	if err != nil {
		return "", err
	}
	return scope, nil
}

// runSetDefaultProfile handles muxi set default profile
func runSetDefaultProfile(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd)
	if err != nil {
		return err
	}

	// Load profiles
	config, err := server.LoadProfiles()
	if err != nil {
		return err
	}

	if len(config.Profiles) == 0 {
		fmt.Println()
		ui.Dimmed("  No profiles configured")
		fmt.Println()
		fmt.Printf("  Add a profile first: %s\n", ui.Command("muxi profiles add"))
		fmt.Println()
		return nil
	}

	var selected string

	if len(args) > 0 {
		// Profile name provided as argument
		selected = args[0]
		if _, ok := config.Profiles[selected]; !ok {
			return fmt.Errorf("profile '%s' not found", selected)
		}
	} else {
		// Interactive selection
		var currentDefault string
		if scope == "local" {
			dotMuxi, _ := loadDotMuxi()
			currentDefault = dotMuxi.Profile
		} else {
			currentDefault = config.DefaultProfile
		}

		var options []wizard.SelectOption
		currentIndex := 0
		i := 0
		for name := range config.Profiles {
			label := name
			if name == currentDefault {
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
		selected, err = wizard.PromptSelect("Select profile", options, currentIndex)
		if err != nil {
			return err
		}
	}

	// Save based on scope
	if scope == "local" {
		dotMuxi, err := loadDotMuxi()
		if err != nil {
			return fmt.Errorf("failed to load .muxi: %w", err)
		}
		dotMuxi.Profile = selected
		if err := saveDotMuxi(dotMuxi); err != nil {
			return fmt.Errorf("failed to save .muxi: %w", err)
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default profile set to '%s' (local)", selected))
	} else {
		if err := server.SetDefaultProfile(selected); err != nil {
			return err
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default profile set to '%s' (global)", selected))
	}

	fmt.Println()
	return nil
}

// runSetDefaultRegistry handles muxi set default registry
func runSetDefaultRegistry(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd)
	if err != nil {
		return err
	}

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

	var selected string

	if len(args) > 0 {
		// Registry name provided as argument
		selected = args[0]
		if _, ok := config.Registries[selected]; !ok {
			return fmt.Errorf("registry '%s' not found", selected)
		}
	} else {
		// Interactive selection
		var currentDefault string
		if scope == "local" {
			dotMuxi, _ := loadDotMuxi()
			currentDefault = dotMuxi.Registry
		} else {
			currentDefault = config.DefaultRegistry
		}

		var options []wizard.SelectOption
		currentIndex := 0
		i := 0
		for name := range config.Registries {
			label := name
			if name == currentDefault {
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
		selected, err = wizard.PromptSelect("Select registry", options, currentIndex)
		if err != nil {
			return err
		}
	}

	// Save based on scope
	if scope == "local" {
		dotMuxi, err := loadDotMuxi()
		if err != nil {
			return fmt.Errorf("failed to load .muxi: %w", err)
		}
		dotMuxi.Registry = selected
		if err := saveDotMuxi(dotMuxi); err != nil {
			return fmt.Errorf("failed to save .muxi: %w", err)
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default registry set to '%s' (local)", selected))
	} else {
		config.DefaultRegistry = selected
		if err := registry.SaveRegistries(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default registry set to '%s' (global)", selected))
	}

	fmt.Println()
	return nil
}

// runSetDefaultUser handles muxi set default user
func runSetDefaultUser(cmd *cobra.Command, args []string) error {
	scope, err := determineScope(cmd)
	if err != nil {
		return err
	}

	var userID string

	if len(args) > 0 {
		userID = args[0]
	} else {
		// Interactive input
		var currentDefault string
		if scope == "local" {
			dotMuxi, _ := loadDotMuxi()
			currentDefault = dotMuxi.UserID
		} else {
			currentDefault = defaults.GetUserID()
		}

		defaultValue := currentDefault
		if defaultValue == "" {
			defaultValue = "default-user"
		}

		fmt.Println()
		userID, err = wizard.PromptString("Enter user ID", defaultValue, nil)
		if err != nil {
			return err
		}
	}

	// Save based on scope
	if scope == "local" {
		dotMuxi, err := loadDotMuxi()
		if err != nil {
			return fmt.Errorf("failed to load .muxi: %w", err)
		}
		dotMuxi.UserID = userID
		if err := saveDotMuxi(dotMuxi); err != nil {
			return fmt.Errorf("failed to save .muxi: %w", err)
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default user ID set to '%s' (local)", userID))
	} else {
		if err := defaults.SetUserID(userID); err != nil {
			return err
		}
		fmt.Println()
		ui.Success(fmt.Sprintf("Default user ID set to '%s' (global)", userID))
	}

	fmt.Println()
	return nil
}
