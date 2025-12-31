package formation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProfile(t *testing.T) {
	// Test with explicit flag value
	result := ResolveProfile("my-server")
	if result != "my-server" {
		t.Errorf("ResolveProfile('my-server') = %q, want 'my-server'", result)
	}

	// Test with empty (should return default or empty)
	result = ResolveProfile("")
	// Just verify it doesn't panic - actual behavior depends on config
	_ = result
}

func TestLoadDotMuxi(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .muxi file with actual struct fields
	dotMuxiContent := `profile: my-server
registry: registry.example.com
user_id: alice
`
	dotMuxiPath := filepath.Join(tmpDir, ".muxi")
	if err := os.WriteFile(dotMuxiPath, []byte(dotMuxiContent), 0644); err != nil {
		t.Fatalf("failed to write .muxi: %v", err)
	}

	config, err := LoadDotMuxi(tmpDir)
	if err != nil {
		t.Fatalf("LoadDotMuxi() error: %v", err)
	}

	if config.Profile != "my-server" {
		t.Errorf("Profile = %q, want 'my-server'", config.Profile)
	}
	if config.UserID != "alice" {
		t.Errorf("UserID = %q, want 'alice'", config.UserID)
	}
}

func TestLoadDotMuxi_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Returns empty config when file doesn't exist (no error)
	config, err := LoadDotMuxi(tmpDir)
	if err != nil {
		t.Fatalf("LoadDotMuxi() error: %v", err)
	}
	if config.Profile != "" {
		t.Errorf("Profile should be empty when file missing, got %q", config.Profile)
	}
}

func TestResolveFormationID(t *testing.T) {
	// With explicit value
	id, err := ResolveFormationID("explicit-id")
	if err != nil {
		t.Fatalf("ResolveFormationID() error: %v", err)
	}
	if id != "explicit-id" {
		t.Errorf("ID = %q, want 'explicit-id'", id)
	}
}
