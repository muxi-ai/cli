package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// DetectFormation walks up the directory tree to find a formation config file
// Checks for both formation.afs and formation.yaml (afs takes precedence)
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
		// Check if formation.afs or formation.yaml exists
		formationFile, found := FindFormationFile(searchDir)
		if found {
			// Found it! Read the formation file to get ID
			ctx := &FormationContext{
				RootDir: searchDir,
			}

			// Read and parse formation file
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

// FindFormationFile looks for formation.afs or formation.yaml in a directory
// Returns the path and true if found, empty string and false otherwise
// Checks .afs first, then .yaml
func FindFormationFile(dir string) (string, bool) {
	afs := filepath.Join(dir, "formation.afs")
	if _, err := os.Stat(afs); err == nil {
		return afs, true
	}
	yaml := filepath.Join(dir, "formation.yaml")
	if _, err := os.Stat(yaml); err == nil {
		return yaml, true
	}
	return "", false
}

// FindConfigFile looks for a config file with .afs or .yaml extension
// baseName should not include extension (e.g., "weather" not "weather.yaml")
// Returns the path and true if found, empty string and false otherwise
func FindConfigFile(dir, baseName string) (string, bool) {
	afs := filepath.Join(dir, baseName+".afs")
	if _, err := os.Stat(afs); err == nil {
		return afs, true
	}
	yaml := filepath.Join(dir, baseName+".yaml")
	if _, err := os.Stat(yaml); err == nil {
		return yaml, true
	}
	return "", false
}

// HasConfigExtension checks if a filename has .afs or .yaml extension
func HasConfigExtension(name string) bool {
	return strings.HasSuffix(name, ".afs") || strings.HasSuffix(name, ".yaml")
}

// ConfigExtensions returns the list of valid config file extensions
func ConfigExtensions() []string {
	return []string{".afs", ".yaml"}
}
