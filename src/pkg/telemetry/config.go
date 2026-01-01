package telemetry

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// globalConfigPath returns the path to ~/.muxi/config.yaml
func globalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".muxi", "config.yaml")
}

// loadGlobalConfig loads the global config as a raw map
func loadGlobalConfig() (map[string]interface{}, error) {
	path := globalConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config == nil {
		config = make(map[string]interface{})
	}

	return config, nil
}

// saveGlobalConfig saves the global config
func saveGlobalConfig(config map[string]interface{}) error {
	path := globalConfigPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// getCachedMachineID returns the cached machine_id from config
func getCachedMachineID() string {
	config, err := loadGlobalConfig()
	if err != nil {
		return ""
	}
	if id, ok := config["machine_id"].(string); ok {
		return id
	}
	return ""
}

// cacheMachineID stores the machine_id in config
func cacheMachineID(id string) {
	config, err := loadGlobalConfig()
	if err != nil {
		config = make(map[string]interface{})
	}
	config["machine_id"] = id
	saveGlobalConfig(config)
}

// getCachedCountry returns the cached country from config
func getCachedCountry() string {
	config, err := loadGlobalConfig()
	if err != nil {
		return ""
	}
	if country, ok := config["country"].(string); ok {
		return country
	}
	return ""
}

// cacheCountry stores the country in config
func cacheCountry(country string) {
	config, err := loadGlobalConfig()
	if err != nil {
		config = make(map[string]interface{})
	}
	config["country"] = country
	saveGlobalConfig(config)
}

// IsEnabled checks if telemetry is enabled
func IsEnabled() bool {
	// Env var takes precedence
	if os.Getenv("MUXI_TELEMETRY") == "0" {
		return false
	}

	config, err := loadGlobalConfig()
	if err != nil {
		return true // default enabled
	}

	if enabled, ok := config["telemetry"].(bool); ok {
		return enabled
	}

	return true // default enabled
}
