package context

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FormationContext represents a detected formation directory
type FormationContext struct {
	RootDir string // Absolute path to formation root
	ID      string // Formation ID (from formation.yaml)
	Name    string // Formation name (from formation.yaml)
	Version string // Formation version (from formation.yaml)
}

// formationYAML represents the fields we read from formation.yaml
type formationYAML struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// DetectFormation walks up the directory tree to find a formation.yaml file
// Returns the formation context if found, error otherwise
// Stops at home directory or root
func DetectFormation() (*FormationContext, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get home directory for boundary check
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // Continue without home check if we can't get it
	}

	// Walk up the directory tree (max 5 levels)
	searchDir := currentDir
	for i := 0; i < 5; i++ {
		// Check if formation.yaml exists in current directory
		formationFile := filepath.Join(searchDir, "formation.yaml")
		if _, err := os.Stat(formationFile); err == nil {
			// Found it! Read the formation.yaml to get ID
			ctx := &FormationContext{
				RootDir: searchDir,
			}

			// Read and parse formation.yaml
			data, err := os.ReadFile(formationFile)
			if err == nil {
				var f formationYAML
				if yaml.Unmarshal(data, &f) == nil {
					ctx.ID = f.ID
					ctx.Name = f.Name
					ctx.Version = f.Version
				}
			}

			// Fallback to directory name if no ID in YAML
			if ctx.ID == "" {
				ctx.ID = filepath.Base(searchDir)
			}

			return ctx, nil
		}

		// Stop if we've reached home or root
		if searchDir == homeDir || searchDir == "/" || searchDir == "." {
			break
		}

		// Move up one directory
		parentDir := filepath.Dir(searchDir)
		if parentDir == searchDir {
			// Reached root (parent == self)
			break
		}
		searchDir = parentDir
	}

	return nil, fmt.Errorf("not in a formation directory")
}

// MustDetectFormation is like DetectFormation but returns a user-friendly error
func MustDetectFormation() (*FormationContext, error) {
	ctx, err := DetectFormation()
	if err != nil {
		return nil, fmt.Errorf("not in a formation directory\n\nRun this command from inside a formation directory:\n  cd my-formation\n  muxi new agent weather\n\nOr create a new formation:\n  muxi new formation")
	}
	return ctx, nil
}
