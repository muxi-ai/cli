package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultEndpoint = "https://capture.muxi.org/v1/telemetry"
	sendTimeout     = 2 * time.Second
)

// Version is set at build time
var Version = "dev"

// Send posts the telemetry event to the server (fire-and-forget)
func Send(event Event) {
	endpoint := os.Getenv("TELEMETRY_URL")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: sendTimeout}

	// Fire and forget with single retry
	for i := 0; i < 2; i++ {
		resp, err := client.Post(endpoint, "application/json", bytes.NewReader(data))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return
			}
		}
	}
}

// getSystemInfo returns CLI version and platform info
func getSystemInfo() SystemInfo {
	return SystemInfo{
		Version: Version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}

// getInfrastructureInfo reads config files to count configured resources
func getInfrastructureInfo() InfrastructureInfo {
	home, _ := os.UserHomeDir()
	cliDir := filepath.Join(home, ".muxi", "cli")

	return InfrastructureInfo{
		ProfilesConfigured:   countProfiles(cliDir),
		FormationsConfigured: countFormations(cliDir),
		RegistriesConfigured: countRegistries(cliDir),
	}
}

func countProfiles(cliDir string) int {
	path := filepath.Join(cliDir, "profiles.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var profiles map[string]interface{}
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return 0
	}

	if items, ok := profiles["profiles"].([]interface{}); ok {
		return len(items)
	}

	return 0
}

func countFormations(cliDir string) int {
	path := filepath.Join(cliDir, "formations.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var formations map[string]interface{}
	if err := yaml.Unmarshal(data, &formations); err != nil {
		return 0
	}

	if items, ok := formations["formations"].([]interface{}); ok {
		return len(items)
	}

	return 0
}

func countRegistries(cliDir string) int {
	path := filepath.Join(cliDir, "registries.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var registries map[string]interface{}
	if err := yaml.Unmarshal(data, &registries); err != nil {
		return 0
	}

	if items, ok := registries["registries"].([]interface{}); ok {
		return len(items)
	}

	return 0
}
