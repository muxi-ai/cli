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

// DeclaredComponents holds the explicitly declared component IDs from a formation manifest
type DeclaredComponents struct {
	Agents      map[string]bool
	MCPServers  map[string]bool
	A2AServices map[string]bool
}

// GetDeclaredComponents parses a formation file and returns declared component IDs
func GetDeclaredComponents(rootDir string) *DeclaredComponents {
	dc := &DeclaredComponents{
		Agents:      make(map[string]bool),
		MCPServers:  make(map[string]bool),
		A2AServices: make(map[string]bool),
	}

	formationPath, found := FindFormationFile(rootDir)
	if !found {
		return dc
	}

	data, err := os.ReadFile(formationPath)
	if err != nil {
		return dc
	}

	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		return dc
	}

	// Extract agents
	if agents, ok := formation["agents"].([]interface{}); ok {
		for _, a := range agents {
			if id, ok := a.(string); ok {
				dc.Agents[id] = true
			}
		}
	}

	// Extract mcp.servers
	if mcp, ok := formation["mcp"].(map[string]interface{}); ok {
		if servers, ok := mcp["servers"].([]interface{}); ok {
			for _, s := range servers {
				if id, ok := s.(string); ok {
					dc.MCPServers[id] = true
				}
			}
		}
	}

	// Extract a2a.outbound.services
	if a2a, ok := formation["a2a"].(map[string]interface{}); ok {
		if outbound, ok := a2a["outbound"].(map[string]interface{}); ok {
			if services, ok := outbound["services"].([]interface{}); ok {
				for _, s := range services {
					if id, ok := s.(string); ok {
						dc.A2AServices[id] = true
					}
				}
			}
		}
	}

	return dc
}

// IsComponentDeclared checks if a file in a component directory is declared.
// Returns true if the component dir has no declarations (backward compat) or if the file is declared.
func (dc *DeclaredComponents) IsComponentDeclared(relPath string) bool {
	parts := strings.SplitN(filepath.ToSlash(relPath), "/", 2)
	if len(parts) != 2 {
		return true
	}

	dir := parts[0]
	filename := parts[1]
	stem := strings.TrimSuffix(filename, filepath.Ext(filename))

	switch dir {
	case "agents":
		if len(dc.Agents) == 0 {
			return true // no declarations = include all (backward compat)
		}
		return dc.Agents[stem]
	case "mcps":
		if len(dc.MCPServers) == 0 {
			return true
		}
		return dc.MCPServers[stem]
	case "a2a":
		if len(dc.A2AServices) == 0 {
			return true
		}
		return dc.A2AServices[stem]
	default:
		return true // non-component dirs always included
	}
}
