package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasConfigExtension(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"file.afs", true},
		{"file.yaml", true},
		{"file.AFS", false},  // case sensitive
		{"file.YAML", false}, // case sensitive
		{"file.yml", false},
		{"file.json", false},
		{"file.txt", false},
		{"file", false},
		{".afs", true},
		{".yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasConfigExtension(tt.name)
			if got != tt.want {
				t.Errorf("HasConfigExtension(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestConfigExtensions(t *testing.T) {
	exts := ConfigExtensions()
	if len(exts) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(exts))
	}
	if exts[0] != ".afs" {
		t.Errorf("expected first extension to be .afs, got %s", exts[0])
	}
	if exts[1] != ".yaml" {
		t.Errorf("expected second extension to be .yaml, got %s", exts[1])
	}
}

func TestFindFormationFile(t *testing.T) {
	t.Run("finds formation.afs", func(t *testing.T) {
		tmpDir := t.TempDir()
		afsPath := filepath.Join(tmpDir, "formation.afs")
		os.WriteFile(afsPath, []byte("id: test"), 0644)

		path, found := FindFormationFile(tmpDir)
		if !found {
			t.Error("expected to find formation.afs")
		}
		if path != afsPath {
			t.Errorf("expected %s, got %s", afsPath, path)
		}
	})

	t.Run("finds formation.yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		yamlPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(yamlPath, []byte("id: test"), 0644)

		path, found := FindFormationFile(tmpDir)
		if !found {
			t.Error("expected to find formation.yaml")
		}
		if path != yamlPath {
			t.Errorf("expected %s, got %s", yamlPath, path)
		}
	})

	t.Run("prefers afs over yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		afsPath := filepath.Join(tmpDir, "formation.afs")
		yamlPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(afsPath, []byte("id: afs"), 0644)
		os.WriteFile(yamlPath, []byte("id: yaml"), 0644)

		path, found := FindFormationFile(tmpDir)
		if !found {
			t.Error("expected to find formation file")
		}
		if path != afsPath {
			t.Errorf("expected afs to take precedence, got %s", path)
		}
	})

	t.Run("returns false when not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		path, found := FindFormationFile(tmpDir)
		if found {
			t.Error("expected not to find formation file")
		}
		if path != "" {
			t.Errorf("expected empty path, got %s", path)
		}
	})
}

func TestFindConfigFile(t *testing.T) {
	t.Run("finds afs file", func(t *testing.T) {
		tmpDir := t.TempDir()
		afsPath := filepath.Join(tmpDir, "weather.afs")
		os.WriteFile(afsPath, []byte("id: weather"), 0644)

		path, found := FindConfigFile(tmpDir, "weather")
		if !found {
			t.Error("expected to find weather.afs")
		}
		if path != afsPath {
			t.Errorf("expected %s, got %s", afsPath, path)
		}
	})

	t.Run("finds yaml file", func(t *testing.T) {
		tmpDir := t.TempDir()
		yamlPath := filepath.Join(tmpDir, "weather.yaml")
		os.WriteFile(yamlPath, []byte("id: weather"), 0644)

		path, found := FindConfigFile(tmpDir, "weather")
		if !found {
			t.Error("expected to find weather.yaml")
		}
		if path != yamlPath {
			t.Errorf("expected %s, got %s", yamlPath, path)
		}
	})

	t.Run("prefers afs over yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		afsPath := filepath.Join(tmpDir, "agent.afs")
		yamlPath := filepath.Join(tmpDir, "agent.yaml")
		os.WriteFile(afsPath, []byte("id: afs"), 0644)
		os.WriteFile(yamlPath, []byte("id: yaml"), 0644)

		path, found := FindConfigFile(tmpDir, "agent")
		if !found {
			t.Error("expected to find config file")
		}
		if path != afsPath {
			t.Errorf("expected afs to take precedence, got %s", path)
		}
	})

	t.Run("returns false when not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		path, found := FindConfigFile(tmpDir, "nonexistent")
		if found {
			t.Error("expected not to find config file")
		}
		if path != "" {
			t.Errorf("expected empty path, got %s", path)
		}
	})
}

func TestDetectFormation(t *testing.T) {
	t.Run("detects formation in current dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `id: my-formation
name: My Formation
version: "1.0.0"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		// Change to tmp dir
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		ctx, err := DetectFormation()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ctx.ID != "my-formation" {
			t.Errorf("expected ID 'my-formation', got %s", ctx.ID)
		}
		if ctx.Name != "My Formation" {
			t.Errorf("expected Name 'My Formation', got %s", ctx.Name)
		}
		if ctx.Version != "1.0.0" {
			t.Errorf("expected Version '1.0.0', got %s", ctx.Version)
		}
	})

	t.Run("detects formation in parent dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(formationPath, []byte("id: parent-formation"), 0644)

		// Create subdirectory
		subDir := filepath.Join(tmpDir, "subdir")
		os.Mkdir(subDir, 0755)

		// Change to subdir
		oldWd, _ := os.Getwd()
		os.Chdir(subDir)
		defer os.Chdir(oldWd)

		ctx, err := DetectFormation()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ctx.ID != "parent-formation" {
			t.Errorf("expected ID 'parent-formation', got %s", ctx.ID)
		}
	})

	t.Run("uses directory name as fallback ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(formationPath, []byte("name: test"), 0644) // No ID field

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		ctx, err := DetectFormation()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// ID should fall back to directory name
		if ctx.ID == "" {
			t.Error("expected non-empty ID fallback")
		}
	})

	t.Run("returns error when not in formation", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		_, err := DetectFormation()
		if err == nil {
			t.Error("expected error when not in formation directory")
		}
	})
}

func TestMustDetectFormation(t *testing.T) {
	t.Run("returns friendly error", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		_, err := MustDetectFormation()
		if err == nil {
			t.Error("expected error when not in formation directory")
		}
		// Should contain helpful message
		errStr := err.Error()
		if !contains(errStr, "formation directory") {
			t.Errorf("error should mention formation directory: %s", errStr)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
