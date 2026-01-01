package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func globalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "config.yaml")
}

func loadGlobalConfigRaw() (map[string]interface{}, error) {
	path := globalConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config == nil {
		config = make(map[string]interface{})
	}

	return config, nil
}

func saveGlobalConfigRaw(config map[string]interface{}) error {
	path := globalConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func getTelemetryStatus() (*bool, error) {
	config, err := loadGlobalConfigRaw()
	if err != nil {
		return nil, err
	}

	val, exists := config["telemetry"]
	if !exists {
		return nil, nil
	}

	if b, ok := val.(bool); ok {
		return &b, nil
	}

	return nil, nil
}

func setTelemetryStatus(enabled bool) error {
	config, err := loadGlobalConfigRaw()
	if err != nil {
		return err
	}

	config["telemetry"] = enabled

	return saveGlobalConfigRaw(config)
}

var telemetryCmd = &cobra.Command{
	Use:    "telemetry",
	Short:  "Manage telemetry settings",
	Hidden: true,
}

var telemetryEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable telemetry",
	Args:  cobra.NoArgs,
	RunE:  runTelemetryEnable,
}

var telemetryDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable telemetry",
	Args:  cobra.NoArgs,
	RunE:  runTelemetryDisable,
}

var telemetryStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show telemetry status",
	Args:  cobra.NoArgs,
	RunE:  runTelemetryStatus,
}

func runTelemetryEnable(cmd *cobra.Command, args []string) error {
	if err := setTelemetryStatus(true); err != nil {
		ui.Error(err.Error())
		return nil
	}

	fmt.Println("Telemetry enabled successfully.")
	fmt.Println()
	fmt.Println("You can check the telemetry status at any time using:")
	fmt.Println("  muxi telemetry status")

	return nil
}

func runTelemetryDisable(cmd *cobra.Command, args []string) error {
	if err := setTelemetryStatus(false); err != nil {
		ui.Error(err.Error())
		return nil
	}

	fmt.Println("Telemetry disabled successfully.")
	fmt.Println()
	fmt.Println("You can check the telemetry status at any time using:")
	fmt.Println("  muxi telemetry status")

	return nil
}

func runTelemetryStatus(cmd *cobra.Command, args []string) error {
	status, err := getTelemetryStatus()
	if err != nil {
		ui.Error(err.Error())
		return nil
	}

	if status == nil {
		fmt.Println("Telemetry: not configured (default: enabled)")
	} else if *status {
		fmt.Println("Telemetry: enabled")
	} else {
		fmt.Println("Telemetry: disabled")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(telemetryCmd)
	telemetryCmd.AddCommand(telemetryEnableCmd)
	telemetryCmd.AddCommand(telemetryDisableCmd)
	telemetryCmd.AddCommand(telemetryStatusCmd)
}
