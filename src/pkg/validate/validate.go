package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/secrets"
	"gopkg.in/yaml.v3"
)

// Result represents validation results
type Result struct {
	Errors   []Issue
	Warnings []Issue
}

// Issue represents a validation error or warning
type Issue struct {
	Field   string
	Message string
	File    string
}

// IsValid returns true if there are no errors
func (r *Result) IsValid() bool {
	return len(r.Errors) == 0
}

// Formation validates a formation directory
func Formation(rootDir string) (*Result, error) {
	result := &Result{
		Errors:   []Issue{},
		Warnings: []Issue{},
	}

	// Check formation.yaml exists
	formationPath := filepath.Join(rootDir, "formation.yaml")
	if _, err := os.Stat(formationPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, Issue{
			File:    "formation.yaml",
			Message: "formation.yaml not found",
		})
		return result, nil
	}

	// Parse formation.yaml
	data, err := os.ReadFile(formationPath)
	if err != nil {
		result.Errors = append(result.Errors, Issue{
			File:    "formation.yaml",
			Message: fmt.Sprintf("failed to read: %v", err),
		})
		return result, nil
	}

	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		result.Errors = append(result.Errors, Issue{
			File:    "formation.yaml",
			Message: fmt.Sprintf("invalid YAML: %v", err),
		})
		return result, nil
	}

	// Validate required fields
	validateRequiredFields(formation, result)

	// Validate server config
	validateServer(formation, result)

	// Validate LLM config
	validateLLM(formation, result)

	// Collect and validate secret references
	secretRefs := collectSecretRefs(string(data))
	validateSecretRefs(rootDir, secretRefs, result)

	// Validate agent files
	validateAgents(rootDir, result)

	// Validate MCP files
	validateMCPs(rootDir, result)

	return result, nil
}

// validateRequiredFields checks for required top-level fields
func validateRequiredFields(formation map[string]interface{}, result *Result) {
	required := []string{"schema", "id"}

	for _, field := range required {
		if _, ok := formation[field]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    "formation.yaml",
				Field:   field,
				Message: fmt.Sprintf("required field '%s' is missing", field),
			})
		}
	}

	// Check schema version
	if schema, ok := formation["schema"].(string); ok {
		if schema != "1.0.0" {
			result.Warnings = append(result.Warnings, Issue{
				File:    "formation.yaml",
				Field:   "schema",
				Message: fmt.Sprintf("unknown schema version '%s', expected '1.0.0'", schema),
			})
		}
	}

	// Check id format
	if id, ok := formation["id"].(string); ok {
		if !isValidID(id) {
			result.Errors = append(result.Errors, Issue{
				File:    "formation.yaml",
				Field:   "id",
				Message: "id must be lowercase alphanumeric with hyphens (e.g., 'my-formation')",
			})
		}
	}
}

// validateServer checks server configuration
func validateServer(formation map[string]interface{}, result *Result) {
	server, ok := formation["server"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    "formation.yaml",
			Field:   "server",
			Message: "server configuration not found, using defaults",
		})
		return
	}

	// Check port range
	if port, ok := server["port"].(int); ok {
		if port < 1 || port > 65535 {
			result.Errors = append(result.Errors, Issue{
				File:    "formation.yaml",
				Field:   "server.port",
				Message: fmt.Sprintf("invalid port %d (must be 1-65535)", port),
			})
		}
	}

	// Check API keys exist
	apiKeys, ok := server["api_keys"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    "formation.yaml",
			Field:   "server.api_keys",
			Message: "no API keys configured",
		})
	} else {
		if _, ok := apiKeys["admin_key"]; !ok {
			result.Warnings = append(result.Warnings, Issue{
				File:    "formation.yaml",
				Field:   "server.api_keys.admin_key",
				Message: "admin_key not configured",
			})
		}
	}
}

// validateLLM checks LLM configuration
func validateLLM(formation map[string]interface{}, result *Result) {
	llm, ok := formation["llm"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    "formation.yaml",
			Field:   "llm",
			Message: "LLM configuration not found",
		})
		return
	}

	// Check for at least one model or api_key
	hasAPIKeys := false
	hasModels := false

	if apiKeys, ok := llm["api_keys"].(map[string]interface{}); ok && len(apiKeys) > 0 {
		hasAPIKeys = true
	}

	if models, ok := llm["models"].([]interface{}); ok && len(models) > 0 {
		hasModels = true
	}

	if !hasAPIKeys && !hasModels {
		result.Warnings = append(result.Warnings, Issue{
			File:    "formation.yaml",
			Field:   "llm",
			Message: "no API keys or models configured",
		})
	}
}

// collectSecretRefs finds all ${{ secrets.* }} references in content
func collectSecretRefs(content string) []string {
	pattern := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z0-9_]+)\s*\}\}`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	refs := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			refs = append(refs, match[1])
			seen[match[1]] = true
		}
	}

	return refs
}

// validateSecretRefs checks that referenced secrets exist
func validateSecretRefs(rootDir string, refs []string, result *Result) {
	if len(refs) == 0 {
		return
	}

	sm := secrets.NewManager(rootDir)
	existingSecrets, err := sm.List()
	if err != nil {
		// Secrets file might not exist yet
		for _, ref := range refs {
			result.Warnings = append(result.Warnings, Issue{
				Field:   ref,
				Message: fmt.Sprintf("secret '%s' referenced but secrets not configured (run 'muxi secrets setup')", ref),
			})
		}
		return
	}

	// Build set of existing secrets
	existing := make(map[string]bool)
	for _, s := range existingSecrets {
		existing[s] = true
	}

	// Check each reference
	for _, ref := range refs {
		if !existing[ref] {
			result.Errors = append(result.Errors, Issue{
				Field:   ref,
				Message: fmt.Sprintf("Secret `%s` is referenced but not set.\n * Run `muxi secrets set %s <VALUE>` to set it", ref, ref),
			})
		}
	}
}

// validateAgents checks agent files in agents/ directory
func validateAgents(rootDir string, result *Result) {
	agentsDir := filepath.Join(rootDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return // No agents directory is OK
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		agentPath := filepath.Join(agentsDir, entry.Name())
		data, err := os.ReadFile(agentPath)
		if err != nil {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("agents", entry.Name()),
				Message: fmt.Sprintf("failed to read: %v", err),
			})
			continue
		}

		var agent map[string]interface{}
		if err := yaml.Unmarshal(data, &agent); err != nil {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("agents", entry.Name()),
				Message: fmt.Sprintf("invalid YAML: %v", err),
			})
			continue
		}

		// Check required agent fields
		if _, ok := agent["id"]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("agents", entry.Name()),
				Field:   "id",
				Message: "required field 'id' is missing",
			})
		}

		if _, ok := agent["role"]; !ok {
			result.Warnings = append(result.Warnings, Issue{
				File:    filepath.Join("agents", entry.Name()),
				Field:   "role",
				Message: "agent has no role defined",
			})
		}

		// Collect secret refs from agent file
		secretRefs := collectSecretRefs(string(data))
		validateSecretRefs(rootDir, secretRefs, result)
	}
}

// validateMCPs checks MCP files in mcps/ directory
func validateMCPs(rootDir string, result *Result) {
	mcpsDir := filepath.Join(rootDir, "mcps")
	if _, err := os.Stat(mcpsDir); os.IsNotExist(err) {
		return // No MCPs directory is OK
	}

	entries, err := os.ReadDir(mcpsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		mcpPath := filepath.Join(mcpsDir, entry.Name())
		data, err := os.ReadFile(mcpPath)
		if err != nil {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("mcps", entry.Name()),
				Message: fmt.Sprintf("failed to read: %v", err),
			})
			continue
		}

		var mcp map[string]interface{}
		if err := yaml.Unmarshal(data, &mcp); err != nil {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("mcps", entry.Name()),
				Message: fmt.Sprintf("invalid YAML: %v", err),
			})
			continue
		}

		// Check required MCP fields
		if _, ok := mcp["id"]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("mcps", entry.Name()),
				Field:   "id",
				Message: "required field 'id' is missing",
			})
		}

		if _, ok := mcp["type"]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    filepath.Join("mcps", entry.Name()),
				Field:   "type",
				Message: "required field 'type' is missing",
			})
		}

		// Collect secret refs from MCP file
		secretRefs := collectSecretRefs(string(data))
		validateSecretRefs(rootDir, secretRefs, result)
	}
}

// isValidID checks if an ID follows the naming convention
func isValidID(id string) bool {
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$`)
	return pattern.MatchString(id)
}
