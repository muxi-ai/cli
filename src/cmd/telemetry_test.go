package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// setTestHome sets the home directory for testing, handling both Unix (HOME) and Windows (USERPROFILE)
func setTestHome(t *testing.T, tmpHome string) func() {
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", tmpHome)
	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tmpHome)
	}

	return func() {
		os.Setenv("HOME", oldHome)
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldUserProfile)
		}
	}
}

func TestGlobalConfigPath(t *testing.T) {
	path := globalConfigPath()
	if path == "" {
		t.Error("globalConfigPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Error("globalConfigPath() should return absolute path")
	}
}

func TestLoadGlobalConfigRaw_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	config, err := loadGlobalConfigRaw()
	if err != nil {
		t.Fatalf("loadGlobalConfigRaw() error: %v", err)
	}
	if config == nil {
		t.Fatal("loadGlobalConfigRaw() returned nil")
	}
	if _, exists := config["telemetry"]; exists {
		t.Error("telemetry should not exist for new config")
	}
}

func TestSetAndGetTelemetryStatus(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	// Set telemetry enabled
	err := setTelemetryStatus(true)
	if err != nil {
		t.Fatalf("setTelemetryStatus() error: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".muxi", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config.yaml was not created")
	}

	// Get and verify
	status, err := getTelemetryStatus()
	if err != nil {
		t.Fatalf("getTelemetryStatus() error: %v", err)
	}
	if status == nil {
		t.Fatal("status should not be nil")
	}
	if !*status {
		t.Error("telemetry should be true")
	}
}

func TestTelemetryPreservesExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	// Create config with other fields
	configDir := filepath.Join(tmpDir, ".muxi")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")

	initialConfig := map[string]interface{}{
		"other_setting": "should_persist",
		"nested": map[string]interface{}{
			"key": "value",
		},
	}
	data, _ := yaml.Marshal(initialConfig)
	os.WriteFile(configPath, data, 0644)

	// Set telemetry
	err := setTelemetryStatus(false)
	if err != nil {
		t.Fatalf("setTelemetryStatus() error: %v", err)
	}

	// Read raw file and verify other fields preserved
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Check telemetry was set
	telemetry, ok := raw["telemetry"].(bool)
	if !ok {
		t.Fatal("telemetry field not found or not bool")
	}
	if telemetry != false {
		t.Error("telemetry should be false")
	}

	// Check other fields preserved
	if raw["other_setting"] != "should_persist" {
		t.Error("other_setting was not preserved")
	}
	nested, ok := raw["nested"].(map[string]interface{})
	if !ok || nested["key"] != "value" {
		t.Error("nested config was not preserved")
	}
}

func TestRunTelemetryEnable(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	err := runTelemetryEnable(nil, nil)
	if err != nil {
		t.Fatalf("runTelemetryEnable() error: %v", err)
	}

	status, _ := getTelemetryStatus()
	if status == nil || !*status {
		t.Error("Telemetry should be enabled")
	}
}

func TestRunTelemetryDisable(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	err := runTelemetryDisable(nil, nil)
	if err != nil {
		t.Fatalf("runTelemetryDisable() error: %v", err)
	}

	status, _ := getTelemetryStatus()
	if status == nil || *status {
		t.Error("Telemetry should be disabled")
	}
}

func TestRunTelemetryStatus(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := setTestHome(t, tmpDir)
	defer cleanup()

	// Test not configured
	err := runTelemetryStatus(nil, nil)
	if err != nil {
		t.Fatalf("runTelemetryStatus() error: %v", err)
	}

	// Test enabled
	runTelemetryEnable(nil, nil)
	err = runTelemetryStatus(nil, nil)
	if err != nil {
		t.Fatalf("runTelemetryStatus() error: %v", err)
	}

	// Test disabled
	runTelemetryDisable(nil, nil)
	err = runTelemetryStatus(nil, nil)
	if err != nil {
		t.Fatalf("runTelemetryStatus() error: %v", err)
	}
}

func TestTelemetryCommandHidden(t *testing.T) {
	if !telemetryCmd.Hidden {
		t.Error("telemetryCmd should be hidden")
	}
}
