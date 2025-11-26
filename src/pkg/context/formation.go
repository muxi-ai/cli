package context

import (
	"fmt"
	"os"
	"path/filepath"
)

// FormationContext represents a detected formation directory
type FormationContext struct {
	RootDir string // Absolute path to formation root
	ID      string // Formation ID (from formation.yaml or directory name)
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
			// Found it!
			ctx := &FormationContext{
				RootDir: searchDir,
				ID:      filepath.Base(searchDir), // Use directory name as ID
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
