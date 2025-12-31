package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfilesCRUD(t *testing.T) {
	// Create temp dir for test
	tmpDir, err := os.MkdirTemp("", "muxi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override config path for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create .muxi/cli directory
	configDir := filepath.Join(tmpDir, ".muxi", "cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Test AddProfile
	entry := ProfileEntry{
		URL:       "http://localhost:7890",
		KeyID:     "test-key",
		SecretKey: "test-secret",
	}

	err = AddProfile("test-server", entry)
	if err != nil {
		t.Fatalf("AddProfile() error: %v", err)
	}

	// Test GetProfile
	got, err := GetProfile("test-server")
	if err != nil {
		t.Fatalf("GetProfile() error: %v", err)
	}
	if got.URL != entry.URL {
		t.Errorf("URL = %q, want %q", got.URL, entry.URL)
	}

	// Test SetDefaultProfile
	err = SetDefaultProfile("test-server")
	if err != nil {
		t.Fatalf("SetDefaultProfile() error: %v", err)
	}

	defaultName := GetDefaultProfile()
	if defaultName != "test-server" {
		t.Errorf("GetDefaultProfile() = %q, want 'test-server'", defaultName)
	}

	// Test RemoveProfile
	err = RemoveProfile("test-server")
	if err != nil {
		t.Fatalf("RemoveProfile() error: %v", err)
	}

	_, err = GetProfile("test-server")
	if err == nil {
		t.Error("GetProfile() should error after removal")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
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

	_, err = GetProfile("nonexistent")
	if err == nil {
		t.Error("GetProfile() should error for nonexistent profile")
	}
}
