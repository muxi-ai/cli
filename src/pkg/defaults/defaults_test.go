package defaults

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

func TestGetEffectiveUserID(t *testing.T) {
	t.Run("formation user takes precedence", func(t *testing.T) {
		got := GetEffectiveUserID("formation-user")
		if got != "formation-user" {
			t.Errorf("GetEffectiveUserID() = %q, want 'formation-user'", got)
		}
	})

	t.Run("empty formation falls back to global", func(t *testing.T) {
		// This will return whatever is in the global config (or empty)
		// Just verify it doesn't panic
		got := GetEffectiveUserID("")
		_ = got // Result depends on actual config state
	})
}

func TestConfigLoadSave(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setTestHome(t, tmpHome)
	defer cleanup()

	t.Run("load returns defaults when file doesn't exist", func(t *testing.T) {
		config, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config.Version != "1.0" {
			t.Errorf("expected version 1.0, got %s", config.Version)
		}
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		config := &Config{
			Version:       "1.0",
			UserID:        "test-user",
			FileExtension: "yaml",
		}

		err := Save(config)
		if err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if loaded.UserID != "test-user" {
			t.Errorf("expected UserID 'test-user', got %s", loaded.UserID)
		}
		if loaded.FileExtension != "yaml" {
			t.Errorf("expected FileExtension 'yaml', got %s", loaded.FileExtension)
		}
	})
}

func TestSetUserID(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setTestHome(t, tmpHome)
	defer cleanup()

	err := SetUserID("alice")
	if err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}

	got := GetUserID()
	if got != "alice" {
		t.Errorf("GetUserID() = %q, want 'alice'", got)
	}
}

func TestFileExtension(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setTestHome(t, tmpHome)
	defer cleanup()

	t.Run("defaults to afs", func(t *testing.T) {
		got := GetFileExtension()
		if got != "afs" {
			t.Errorf("GetFileExtension() = %q, want 'afs'", got)
		}
	})

	t.Run("can set to yaml", func(t *testing.T) {
		err := SetFileExtension("yaml")
		if err != nil {
			t.Fatalf("SetFileExtension failed: %v", err)
		}

		got := GetFileExtension()
		if got != "yaml" {
			t.Errorf("GetFileExtension() = %q, want 'yaml'", got)
		}
	})

	t.Run("can set back to afs", func(t *testing.T) {
		err := SetFileExtension("afs")
		if err != nil {
			t.Fatalf("SetFileExtension failed: %v", err)
		}

		got := GetFileExtension()
		if got != "afs" {
			t.Errorf("GetFileExtension() = %q, want 'afs'", got)
		}
	})

	t.Run("rejects invalid extension", func(t *testing.T) {
		err := SetFileExtension("json")
		if err == nil {
			t.Error("expected error for invalid extension")
		}
	})
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpHome := t.TempDir()
	cleanup := setTestHome(t, tmpHome)
	defer cleanup()

	// Create invalid YAML file
	configDir := filepath.Join(tmpHome, ".muxi", "cli")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "defaults.yaml")
	os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
