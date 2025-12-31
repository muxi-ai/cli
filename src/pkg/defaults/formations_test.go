package defaults

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormationsCRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".muxi", "cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Test AddFormation
	entry := FormationEntry{
		DefaultProfile: "my-server",
		AdminKey:       "admin-key",
		ClientKey:      "client-key",
	}

	err = AddFormation("test-formation", entry)
	if err != nil {
		t.Fatalf("AddFormation() error: %v", err)
	}

	// Test FormationExists
	if !FormationExists("test-formation") {
		t.Error("FormationExists() should return true after add")
	}
	if FormationExists("nonexistent") {
		t.Error("FormationExists() should return false for nonexistent")
	}

	// Test GetFormation
	got, err := GetFormation("test-formation")
	if err != nil {
		t.Fatalf("GetFormation() error: %v", err)
	}
	if got.DefaultProfile != entry.DefaultProfile {
		t.Errorf("DefaultProfile = %q, want %q", got.DefaultProfile, entry.DefaultProfile)
	}
	if got.AdminKey != entry.AdminKey {
		t.Errorf("AdminKey = %q, want %q", got.AdminKey, entry.AdminKey)
	}

	// Test RemoveFormation
	err = RemoveFormation("test-formation")
	if err != nil {
		t.Fatalf("RemoveFormation() error: %v", err)
	}

	if FormationExists("test-formation") {
		t.Error("FormationExists() should return false after removal")
	}
}

func TestGetFormation_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".muxi", "cli")
	os.MkdirAll(configDir, 0755)

	_, err = GetFormation("nonexistent")
	if err == nil {
		t.Error("GetFormation() should error for nonexistent formation")
	}
}

func TestLoadFormations_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".muxi", "cli")
	os.MkdirAll(configDir, 0755)

	config, err := LoadFormations()
	if err != nil {
		t.Fatalf("LoadFormations() error: %v", err)
	}
	if config.Formations == nil {
		t.Error("Formations map should be initialized")
	}
}
