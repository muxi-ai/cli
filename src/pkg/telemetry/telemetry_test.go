package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMachineID(t *testing.T) {
	// Test that machine ID is deterministic
	id1 := GetMachineID()
	id2 := GetMachineID()

	if id1 == "" {
		t.Error("Machine ID should not be empty")
	}

	if id1 != id2 {
		t.Errorf("Machine ID should be deterministic, got %s and %s", id1, id2)
	}

	// Check UUID format (8-4-4-4-12)
	if len(id1) != 36 {
		t.Errorf("Machine ID should be 36 chars, got %d", len(id1))
	}
}

func TestHashMachineID(t *testing.T) {
	// Test hash function produces consistent output
	hash1 := hashMachineID("test-id")
	hash2 := hashMachineID("test-id")

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic, got %s and %s", hash1, hash2)
	}

	// Different inputs should produce different outputs
	hash3 := hashMachineID("other-id")
	if hash1 == hash3 {
		t.Error("Different inputs should produce different hashes")
	}
}

func TestLocalState(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()

	// Override home dir for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Also set USERPROFILE for Windows
	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	// Create CLI dir
	cliDir := filepath.Join(tmpDir, ".muxi", "cli")
	os.MkdirAll(cliDir, 0755)

	// Test loading non-existent state
	state := Load()
	if state == nil {
		t.Fatal("Load should return new state if file doesn't exist")
	}

	// Test incrementing counters
	state.IncrementDeploy()
	state.IncrementDeploy()
	state.IncrementPull()
	state.IncrementPush()
	state.IncrementFormationCreated()
	state.IncrementScaffold("agent")
	state.IncrementScaffold("mcp")
	state.IncrementChat()
	state.IncrementLogs()
	state.IncrementHelp("deploy")
	state.IncrementHelp("deploy")
	state.IncrementHelp("secrets")

	if state.Formations.Deployed != 2 {
		t.Errorf("Expected 2 deploys, got %d", state.Formations.Deployed)
	}

	if state.Registry.Pulls != 1 {
		t.Errorf("Expected 1 pull, got %d", state.Registry.Pulls)
	}

	if state.Registry.Pushes != 1 {
		t.Errorf("Expected 1 push, got %d", state.Registry.Pushes)
	}

	if state.Scaffolding.Agents != 1 {
		t.Errorf("Expected 1 agent, got %d", state.Scaffolding.Agents)
	}

	if state.Help["deploy"] != 2 {
		t.Errorf("Expected 2 help:deploy, got %d", state.Help["deploy"])
	}

	// Test save and reload
	if err := state.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := Load()
	if loaded.Formations.Deployed != 2 {
		t.Errorf("After reload: expected 2 deploys, got %d", loaded.Formations.Deployed)
	}

	// Test reset
	loaded.Reset()
	if loaded.Formations.Deployed != 0 {
		t.Errorf("After reset: expected 0 deploys, got %d", loaded.Formations.Deployed)
	}
	if len(loaded.Help) != 0 {
		t.Errorf("After reset: expected empty help map, got %v", loaded.Help)
	}
}

func TestIsEnabled(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()

	// Override home dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	// Clear any existing env var
	origTelemetry := os.Getenv("MUXI_TELEMETRY")
	os.Unsetenv("MUXI_TELEMETRY")
	defer func() {
		if origTelemetry != "" {
			os.Setenv("MUXI_TELEMETRY", origTelemetry)
		}
	}()

	// Default should be enabled
	if !IsEnabled() {
		t.Error("Telemetry should be enabled by default")
	}

	// Test env var override
	os.Setenv("MUXI_TELEMETRY", "0")
	if IsEnabled() {
		t.Error("Telemetry should be disabled when MUXI_TELEMETRY=0")
	}
	os.Unsetenv("MUXI_TELEMETRY")

	// Test config file override
	configDir := filepath.Join(tmpDir, ".muxi")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")
	os.WriteFile(configPath, []byte("telemetry: false\n"), 0644)

	if IsEnabled() {
		t.Error("Telemetry should be disabled when config says false")
	}
}

func TestBuildEvent(t *testing.T) {
	state := NewLocalState()
	state.IncrementDeploy()
	state.IncrementPull()
	state.IncrementHelp("deploy")

	event := state.buildEvent()

	if event.Module != "cli" {
		t.Errorf("Expected module 'cli', got %s", event.Module)
	}

	if event.SchemaVersion != 1 {
		t.Errorf("Expected schema version 1, got %d", event.SchemaVersion)
	}

	if event.Payload.Formations.Deployed != 1 {
		t.Errorf("Expected 1 deploy in payload, got %d", event.Payload.Formations.Deployed)
	}

	if event.Payload.Registry.Pulls != 1 {
		t.Errorf("Expected 1 pull in payload, got %d", event.Payload.Registry.Pulls)
	}

	// Verify JSON serialization works
	_, err := json.Marshal(event)
	if err != nil {
		t.Errorf("Failed to marshal event: %v", err)
	}
}

func TestFlushIfDue(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	// Disable telemetry to avoid actual HTTP calls
	os.Setenv("MUXI_TELEMETRY", "0")
	defer os.Unsetenv("MUXI_TELEMETRY")

	cliDir := filepath.Join(tmpDir, ".muxi", "cli")
	os.MkdirAll(cliDir, 0755)

	state := NewLocalState()
	state.IncrementDeploy()

	// Not due yet (just created)
	state.FlushIfDue()
	if state.Formations.Deployed != 1 {
		t.Error("Counters should not be reset when flush not due")
	}

	// Force flush by setting old timestamp
	state.LastFlush = time.Now().Add(-2 * time.Hour)
	state.FlushIfDue()

	if state.Formations.Deployed != 0 {
		t.Error("Counters should be reset after flush")
	}
}
