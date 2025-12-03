package cmd

import (
	"fmt"
	"os"

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

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setServerCmd)
	setCmd.AddCommand(setRegistryCmd)
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
func checkFormationDir() error {
	if _, err := os.Stat("formation.yaml"); os.IsNotExist(err) {
		return fmt.Errorf("not in a formation directory (no formation.yaml found)")
	}
	return nil
}

// runSetServer handles muxi set server
func runSetServer(cmd *cobra.Command, args []string) error {
	if err := checkFormationDir(); err != nil {
		return err
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
		fmt.Println("  Add a server first: muxi server add")
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
	if err := checkFormationDir(); err != nil {
		return err
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
		fmt.Println("  Add a registry first: muxi registry add")
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
