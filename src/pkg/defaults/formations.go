package defaults

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FormationsConfig is the structure of ~/.muxi/cli/formations.yaml
type FormationsConfig struct {
	Version    string                    `yaml:"version"`
	Formations map[string]FormationEntry `yaml:"formations"`
}

// FormationEntry stores connection info for a remote formation
type FormationEntry struct {
	DefaultProfile string    `yaml:"default_profile"`
	DefaultUserID  string    `yaml:"default_user_id,omitempty"`
	ClientKey      string    `yaml:"client_key"`
	AdminKey       string    `yaml:"admin_key,omitempty"`
	AddedAt        time.Time `yaml:"added_at"`
}

// formationsPath returns the path to formations.yaml
func formationsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "cli", "formations.yaml")
}

// LoadFormations loads the formations configuration
func LoadFormations() (*FormationsConfig, error) {
	path := formationsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FormationsConfig{
				Version:    "1.0",
				Formations: make(map[string]FormationEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read formations config: %w", err)
	}

	var config FormationsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse formations config: %w", err)
	}

	if config.Formations == nil {
		config.Formations = make(map[string]FormationEntry)
	}

	return &config, nil
}

// SaveFormations saves the formations configuration
func SaveFormations(config *FormationsConfig) error {
	path := formationsPath()

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

// GetFormation returns a formation entry by name
func GetFormation(name string) (*FormationEntry, error) {
	config, err := LoadFormations()
	if err != nil {
		return nil, err
	}

	entry, ok := config.Formations[name]
	if !ok {
		return nil, fmt.Errorf("formation '%s' not found in saved formations", name)
	}

	return &entry, nil
}

// AddFormation adds a new formation to the config
func AddFormation(name string, entry FormationEntry) error {
	config, err := LoadFormations()
	if err != nil {
		return err
	}

	config.Formations[name] = entry
	return SaveFormations(config)
}

// RemoveFormation removes a formation from the config
func RemoveFormation(name string) error {
	config, err := LoadFormations()
	if err != nil {
		return err
	}

	if _, ok := config.Formations[name]; !ok {
		return fmt.Errorf("formation '%s' not found", name)
	}

	delete(config.Formations, name)
	return SaveFormations(config)
}

// FormationExists checks if a formation exists in the config
func FormationExists(name string) bool {
	config, err := LoadFormations()
	if err != nil {
		return false
	}
	_, ok := config.Formations[name]
	return ok
}
