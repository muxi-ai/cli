package defaults

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the structure of ~/.muxi/cli/defaults.yaml
type Config struct {
	Version string `yaml:"version"`
	UserID  string `yaml:"user_id,omitempty"`
}

// configPath returns the path to defaults.yaml
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "defaults.yaml")
}

// Load loads the defaults configuration
func Load() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Version: "1.0",
			}, nil
		}
		return nil, fmt.Errorf("failed to read defaults config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse defaults config: %w", err)
	}

	return &config, nil
}

// Save saves the defaults configuration
func Save(config *Config) error {
	path := configPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetUserID returns the default user ID
func GetUserID() string {
	config, err := Load()
	if err != nil {
		return ""
	}
	return config.UserID
}

// SetUserID sets the default user ID
func SetUserID(userID string) error {
	config, err := Load()
	if err != nil {
		return err
	}

	config.UserID = userID
	return Save(config)
}

// GetEffectiveUserID returns the user ID to use, checking:
// 1. Formation-level override (from .muxi file)
// 2. Global default (from ~/.muxi/cli/defaults.yaml)
// Returns empty string if neither is set.
func GetEffectiveUserID(formationUserID string) string {
	// Formation-level takes precedence
	if formationUserID != "" {
		return formationUserID
	}
	// Fall back to global
	return GetUserID()
}
