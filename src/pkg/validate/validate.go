package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
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

	// Check formation.afs or formation.yaml exists
	formationPath, found := context.FindFormationFile(rootDir)
	if !found {
		result.Errors = append(result.Errors, Issue{
			File:    "formation.afs/formation.yaml",
			Message: "formation config file not found (expected formation.afs or formation.yaml)",
		})
		return result, nil
	}
	formationFileName := filepath.Base(formationPath)

	// Parse formation file
	data, err := os.ReadFile(formationPath)
	if err != nil {
		result.Errors = append(result.Errors, Issue{
			File:    formationFileName,
			Message: fmt.Sprintf("failed to read: %v", err),
		})
		return result, nil
	}

	// Step 1: Parse into yaml.Node for structural validation
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		result.Errors = append(result.Errors, Issue{
			File:    formationFileName,
			Message: fmt.Sprintf("invalid syntax: %v", cleanYAMLError(err)),
		})
		return result, nil
	}

	// Step 2: Check for structural issues (empty mappings that should have children)
	validateYAMLStructure(&root, formationFileName, "", result)

	// Step 3: Parse into map for semantic validation
	var formation map[string]interface{}
	if err := yaml.Unmarshal(data, &formation); err != nil {
		result.Errors = append(result.Errors, Issue{
			File:    formationFileName,
			Message: fmt.Sprintf("parse error: %v", cleanYAMLError(err)),
		})
		return result, nil
	}

	// Validate required fields
	validateRequiredFields(formation, formationFileName, result)

	// Validate server config
	validateServer(formation, formationFileName, result)

	// Validate LLM config
	validateLLM(formation, formationFileName, result)

	// Collect and validate secret references
	secretRefs := collectSecretRefs(string(data))
	validateSecretRefs(rootDir, secretRefs, result)

	// Validate agent files
	validateAgents(rootDir, result)

	// Validate MCP files
	validateMCPs(rootDir, result)

	// Validate component declarations vs files
	validateDeclarations(rootDir, formation, result)

	return result, nil
}

// validateRequiredFields checks for required top-level fields
func validateRequiredFields(formation map[string]interface{}, fileName string, result *Result) {
	required := []string{"schema", "id"}

	for _, field := range required {
		if _, ok := formation[field]; !ok {
			result.Errors = append(result.Errors, Issue{
				File:    fileName,
				Field:   field,
				Message: fmt.Sprintf("required field '%s' is missing", field),
			})
		}
	}

	// Check schema version
	if schema, ok := formation["schema"].(string); ok {
		if schema != "1.0.0" {
			result.Warnings = append(result.Warnings, Issue{
				File:    fileName,
				Field:   "schema",
				Message: fmt.Sprintf("unknown schema version '%s', expected '1.0.0'", schema),
			})
		}
	}

	// Check id format
	if id, ok := formation["id"].(string); ok {
		if !isValidID(id) {
			result.Errors = append(result.Errors, Issue{
				File:    fileName,
				Field:   "id",
				Message: "id must be lowercase alphanumeric with hyphens (e.g., 'my-formation')",
			})
		}
	}
}

// validateServer checks server configuration
func validateServer(formation map[string]interface{}, fileName string, result *Result) {
	server, ok := formation["server"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    fileName,
			Field:   "server",
			Message: "server configuration not found, using defaults",
		})
		return
	}

	// Check port range
	if port, ok := server["port"].(int); ok {
		if port < 1 || port > 65535 {
			result.Errors = append(result.Errors, Issue{
				File:    fileName,
				Field:   "server.port",
				Message: fmt.Sprintf("invalid port %d (must be 1-65535)", port),
			})
		}
	}

	// Check API keys exist
	apiKeys, ok := server["api_keys"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    fileName,
			Field:   "server.api_keys",
			Message: "no API keys configured",
		})
	} else {
		if _, ok := apiKeys["admin_key"]; !ok {
			result.Warnings = append(result.Warnings, Issue{
				File:    fileName,
				Field:   "server.api_keys.admin_key",
				Message: "admin_key not configured",
			})
		}
	}
}

// validateLLM checks LLM configuration
func validateLLM(formation map[string]interface{}, fileName string, result *Result) {
	llm, ok := formation["llm"].(map[string]interface{})
	if !ok {
		result.Warnings = append(result.Warnings, Issue{
			File:    fileName,
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
			File:    fileName,
			Field:   "llm",
			Message: "no API keys or models configured",
		})
	}
}

// collectSecretRefs finds all ${{ secrets.* }} references in content
// It ignores YAML comment lines (lines starting with #)
func collectSecretRefs(content string) []string {
	// Filter out comment lines before scanning
	var activeLines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			activeLines = append(activeLines, line)
		}
	}
	filteredContent := strings.Join(activeLines, "\n")

	pattern := regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z0-9_]+)\s*\}\}`)
	matches := pattern.FindAllStringSubmatch(filteredContent, -1)

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
		if entry.IsDir() || !context.HasConfigExtension(entry.Name()) {
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
			for _, e := range splitYAMLErrors(err) {
				result.Errors = append(result.Errors, Issue{
					File:    filepath.Join("agents", entry.Name()),
					Message: e,
				})
			}
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
		if entry.IsDir() || !context.HasConfigExtension(entry.Name()) {
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
			for _, e := range splitYAMLErrors(err) {
				result.Errors = append(result.Errors, Issue{
					File:    filepath.Join("mcps", entry.Name()),
					Message: e,
				})
			}
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

// validateYAMLStructure walks the YAML node tree and checks for structural issues
func validateYAMLStructure(node *yaml.Node, file, path string, result *Result) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			validateYAMLStructure(child, file, path, result)
		}

	case yaml.MappingNode:
		// MappingNode.Content is pairs: [key, value, key, value, ...]
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			keyName := keyNode.Value
			childPath := keyName
			if path != "" {
				childPath = path + "." + keyName
			}

			// Check for empty mapping value that should have children
			// This catches: "settings:" followed by sibling keys instead of children
			if valueNode.Tag == "!!null" && valueNode.Value == "" {
				// Check if this is a key that typically expects a mapping value
				if isExpectedMapping(keyName) {
					result.Errors = append(result.Errors, Issue{
						File:    file,
						Field:   childPath,
						Message: fmt.Sprintf("'%s:' at line %d has no value - expected indented content below it", keyName, keyNode.Line),
					})
				}
			}

			// Recurse into value
			validateYAMLStructure(valueNode, file, childPath, result)
		}

	case yaml.SequenceNode:
		for i, child := range node.Content {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			validateYAMLStructure(child, file, childPath, result)
		}
	}
}

// isExpectedMapping returns true if a key typically expects a mapping/dict value
func isExpectedMapping(key string) bool {
	expectedMappings := map[string]bool{
		"settings":      true,
		"auth":          true,
		"api_keys":      true,
		"server":        true,
		"llm":           true,
		"memory":        true,
		"logging":       true,
		"overlord":      true,
		"a2a":           true,
		"mcp":           true,
		"working":       true,
		"buffer":        true,
		"persistent":    true,
		"remote":        true,
		"caching":       true,
		"image":         true,
		"preprocessing": true,
		"extraction":    true,
		"workflow":      true,
		"clarification": true,
		"response":      true,
		"timeouts":      true,
		"retry":         true,
		"encryption":    true,
		"inbound":       true,
		"outbound":      true,
	}
	return expectedMappings[key]
}

// validateDeclarations checks that declared component IDs match files and vice versa
func validateDeclarations(rootDir string, formation map[string]interface{}, result *Result) {
	formationFileName := "formation"

	// Validate agents
	declaredAgents := extractStringList(formation, "agents")
	fileAgents := getComponentIDs(rootDir, "agents")
	validateComponentDeclarations(declaredAgents, fileAgents, "agent", "agents", formationFileName, result)

	// Validate MCP servers
	var declaredMCPs []string
	if mcp, ok := formation["mcp"].(map[string]interface{}); ok {
		declaredMCPs = extractStringList(mcp, "servers")
	}
	fileMCPs := getComponentIDs(rootDir, "mcps")
	validateComponentDeclarations(declaredMCPs, fileMCPs, "MCP server", "mcp.servers", formationFileName, result)

	// Validate A2A services
	var declaredA2A []string
	if a2a, ok := formation["a2a"].(map[string]interface{}); ok {
		if outbound, ok := a2a["outbound"].(map[string]interface{}); ok {
			declaredA2A = extractStringList(outbound, "services")
		}
	}
	fileA2A := getComponentIDs(rootDir, "a2a")
	validateComponentDeclarations(declaredA2A, fileA2A, "A2A service", "a2a.outbound.services", formationFileName, result)
}

// extractStringList extracts string entries from a list field (ignoring dict entries)
func extractStringList(m map[string]interface{}, key string) []string {
	list, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, item := range list {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// getComponentIDs reads component files from a directory and returns their IDs
func getComponentIDs(rootDir, dirName string) map[string]bool {
	dir := filepath.Join(rootDir, dirName)
	ids := make(map[string]bool)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ids
	}

	for _, entry := range entries {
		if entry.IsDir() || !context.HasConfigExtension(entry.Name()) {
			continue
		}

		// Try to read the id field from the file
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var config map[string]interface{}
		if err := yaml.Unmarshal(data, &config); err != nil {
			continue
		}

		if id, ok := config["id"].(string); ok {
			ids[id] = true
		} else {
			// Fallback to filename stem
			name := entry.Name()
			ext := filepath.Ext(name)
			ids[name[:len(name)-len(ext)]] = true
		}
	}

	return ids
}

// validateComponentDeclarations checks declared IDs against file IDs
func validateComponentDeclarations(declared []string, fileIDs map[string]bool, componentType, section, formationFile string, result *Result) {
	declaredSet := make(map[string]bool)
	for _, id := range declared {
		declaredSet[id] = true
	}

	// Warn about declared IDs that have no matching file
	for _, id := range declared {
		if !fileIDs[id] {
			result.Errors = append(result.Errors, Issue{
				File:    formationFile,
				Field:   section,
				Message: fmt.Sprintf("%s '%s' is declared but no matching file was found", componentType, id),
			})
		}
	}

	// Warn about files that are not declared
	for id := range fileIDs {
		if !declaredSet[id] {
			result.Warnings = append(result.Warnings, Issue{
				Field:   section,
				Message: fmt.Sprintf("%s '%s' exists as a file but is not declared in %s (it will not be loaded)", componentType, id, section),
			})
		}
	}
}

// cleanYAMLError strips the "yaml: " prefix from error messages
func cleanYAMLError(err error) string {
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "yaml: ")
	return msg
}

// splitYAMLErrors splits a multi-line yaml unmarshal error into individual error strings
func splitYAMLErrors(err error) []string {
	msg := err.Error()
	if strings.Contains(msg, "unmarshal errors:") {
		parts := strings.Split(msg, "\n")
		var errors []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" || p == "yaml: unmarshal errors:" {
				continue
			}
			errors = append(errors, p)
		}
		if len(errors) > 0 {
			return errors
		}
	}
	return []string{cleanYAMLError(err)}
}
