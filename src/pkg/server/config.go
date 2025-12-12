package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ProfilesConfig is the structure of ~/.muxi/cli/profiles.yaml
type ProfilesConfig struct {
	Version        string                  `yaml:"version"`
	DefaultProfile string                  `yaml:"default_profile"`
	Profiles       map[string]ProfileEntry `yaml:"profiles"`
}

// ProfileEntry stores connection info for a server profile
type ProfileEntry struct {
	URL       string    `yaml:"url"`
	KeyID     string    `yaml:"key_id"`
	SecretKey string    `yaml:"secret_key"`
	AddedAt   time.Time `yaml:"added_at"`
}

// profilesPath returns the path to profiles.yaml
func profilesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "profiles.yaml")
}

// LoadProfiles loads the profiles configuration
func LoadProfiles() (*ProfilesConfig, error) {
	path := profilesPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProfilesConfig{
				Version:  "1.0",
				Profiles: make(map[string]ProfileEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read profiles config: %w", err)
	}

	var config ProfilesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse profiles config: %w", err)
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]ProfileEntry)
	}

	return &config, nil
}

// SaveProfiles saves the profiles configuration
func SaveProfiles(config *ProfilesConfig) error {
	path := profilesPath()

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

// GetDefaultProfile returns the default profile name
func GetDefaultProfile() string {
	config, err := LoadProfiles()
	if err != nil {
		return ""
	}
	return config.DefaultProfile
}

// GetProfile returns a profile entry by name
func GetProfile(name string) (*ProfileEntry, error) {
	config, err := LoadProfiles()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = config.DefaultProfile
	}

	if name == "" {
		return nil, fmt.Errorf("no profile specified and no default set")
	}

	entry, ok := config.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return &entry, nil
}

// AddProfile adds a new profile to the config
func AddProfile(name string, entry ProfileEntry) error {
	config, err := LoadProfiles()
	if err != nil {
		return err
	}

	config.Profiles[name] = entry
	return SaveProfiles(config)
}

// RemoveProfile removes a profile from the config
func RemoveProfile(name string) error {
	config, err := LoadProfiles()
	if err != nil {
		return err
	}

	if _, ok := config.Profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(config.Profiles, name)

	// Clear default if it was the removed profile
	if config.DefaultProfile == name {
		config.DefaultProfile = ""
	}

	return SaveProfiles(config)
}

// SetDefaultProfile sets the default profile
func SetDefaultProfile(name string) error {
	config, err := LoadProfiles()
	if err != nil {
		return err
	}

	if _, ok := config.Profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found", name)
	}

	config.DefaultProfile = name
	return SaveProfiles(config)
}
