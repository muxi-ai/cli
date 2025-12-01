package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// GetConfigDir returns the CLI config directory (~/.muxi/cli)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".muxi", "cli"), nil
}

// GetRegistriesPath returns the path to registries.yaml
func GetRegistriesPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "registries.yaml"), nil
}

// LoadRegistries loads the registries configuration
func LoadRegistries() (*RegistriesConfig, error) {
	path, err := GetRegistriesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config
			return &RegistriesConfig{
				Version:         "1.0",
				DefaultRegistry: DefaultRegistryURL,
				Registries:      make(map[string]RegistryEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read registries config: %w", err)
	}

	var config RegistriesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse registries config: %w", err)
	}

	if config.Registries == nil {
		config.Registries = make(map[string]RegistryEntry)
	}

	return &config, nil
}

// SaveRegistries saves the registries configuration
func SaveRegistries(config *RegistriesConfig) error {
	path, err := GetRegistriesPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions (600)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write registries config: %w", err)
	}

	return nil
}

// GetToken returns the token for a registry
func GetToken(registry string) (string, error) {
	config, err := LoadRegistries()
	if err != nil {
		return "", err
	}

	if registry == "" {
		registry = config.DefaultRegistry
	}
	if registry == "" {
		registry = DefaultRegistryURL
	}

	entry, ok := config.Registries[registry]
	if !ok {
		return "", nil
	}

	return entry.Token, nil
}

// SetToken saves a token for a registry
func SetToken(registry, token, username string) error {
	config, err := LoadRegistries()
	if err != nil {
		return err
	}

	if registry == "" {
		registry = DefaultRegistryURL
	}

	config.Registries[registry] = RegistryEntry{
		Token:     token,
		Username:  username,
		CreatedAt: time.Now(),
	}

	return SaveRegistries(config)
}

// RemoveToken removes the token for a registry
func RemoveToken(registry string) error {
	config, err := LoadRegistries()
	if err != nil {
		return err
	}

	if registry == "" {
		registry = config.DefaultRegistry
	}
	if registry == "" {
		registry = DefaultRegistryURL
	}

	delete(config.Registries, registry)

	return SaveRegistries(config)
}

// GetUsername returns the username for a registry
func GetUsername(registry string) (string, error) {
	config, err := LoadRegistries()
	if err != nil {
		return "", err
	}

	if registry == "" {
		registry = config.DefaultRegistry
	}
	if registry == "" {
		registry = DefaultRegistryURL
	}

	entry, ok := config.Registries[registry]
	if !ok {
		return "", nil
	}

	return entry.Username, nil
}

// GetDefaultRegistry returns the default registry
func GetDefaultRegistry() string {
	config, err := LoadRegistries()
	if err != nil {
		return DefaultRegistryURL
	}

	if config.DefaultRegistry != "" {
		return config.DefaultRegistry
	}

	return DefaultRegistryURL
}

// IsLoggedIn checks if user is logged in to a registry
func IsLoggedIn(registry string) bool {
	token, err := GetToken(registry)
	if err != nil {
		return false
	}
	return token != ""
}
