package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ServersConfig is the structure of ~/.muxi/cli/servers.yaml
type ServersConfig struct {
	Version string                 `yaml:"version"`
	Default string                 `yaml:"default"`
	Servers map[string]ServerEntry `yaml:"servers"`
}

// ServerEntry stores connection info for a server
type ServerEntry struct {
	URL       string    `yaml:"url"`
	KeyID     string    `yaml:"key_id"`
	SecretKey string    `yaml:"secret_key"`
	AddedAt   time.Time `yaml:"added_at"`
}

// serversPath returns the path to servers.yaml
func serversPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "servers.yaml")
}

// LoadServers loads the servers configuration
func LoadServers() (*ServersConfig, error) {
	path := serversPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ServersConfig{
				Version: "1.0",
				Servers: make(map[string]ServerEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read servers config: %w", err)
	}

	var config ServersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse servers config: %w", err)
	}

	if config.Servers == nil {
		config.Servers = make(map[string]ServerEntry)
	}

	return &config, nil
}

// SaveServers saves the servers configuration
func SaveServers(config *ServersConfig) error {
	path := serversPath()

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

// GetDefaultServer returns the default server name
func GetDefaultServer() string {
	config, err := LoadServers()
	if err != nil {
		return ""
	}
	return config.Default
}

// GetServer returns a server entry by name
func GetServer(name string) (*ServerEntry, error) {
	config, err := LoadServers()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = config.Default
	}

	if name == "" {
		return nil, fmt.Errorf("no server specified and no default set")
	}

	entry, ok := config.Servers[name]
	if !ok {
		return nil, fmt.Errorf("server '%s' not found", name)
	}

	return &entry, nil
}

// AddServer adds a new server to the config
func AddServer(name string, entry ServerEntry) error {
	config, err := LoadServers()
	if err != nil {
		return err
	}

	config.Servers[name] = entry
	return SaveServers(config)
}

// RemoveServer removes a server from the config
func RemoveServer(name string) error {
	config, err := LoadServers()
	if err != nil {
		return err
	}

	if _, ok := config.Servers[name]; !ok {
		return fmt.Errorf("server '%s' not found", name)
	}

	delete(config.Servers, name)

	// Clear default if it was the removed server
	if config.Default == name {
		config.Default = ""
	}

	return SaveServers(config)
}

// SetDefaultServer sets the default server
func SetDefaultServer(name string) error {
	config, err := LoadServers()
	if err != nil {
		return err
	}

	if _, ok := config.Servers[name]; !ok {
		return fmt.Errorf("server '%s' not found", name)
	}

	config.Default = name
	return SaveServers(config)
}
