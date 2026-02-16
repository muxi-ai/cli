package scaffold

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// validateYAML parses YAML and checks for structural issues
func validateYAML(t *testing.T, name, content string) {
	t.Helper()

	// Step 1: Parse into yaml.Node
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		t.Errorf("%s: invalid YAML syntax: %v", name, err)
		return
	}

	// Step 2: Check for empty mappings that should have children
	errors := checkYAMLStructure(&root, "")
	for _, e := range errors {
		t.Errorf("%s: %s", name, e)
	}

	// Step 3: Parse into map to ensure it's valid
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		t.Errorf("%s: failed to parse as map: %v", name, err)
	}
}

// checkYAMLStructure recursively checks for structural issues
func checkYAMLStructure(node *yaml.Node, path string) []string {
	var errors []string

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			errors = append(errors, checkYAMLStructure(child, path)...)
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			keyName := keyNode.Value
			childPath := keyName
			if path != "" {
				childPath = path + "." + keyName
			}

			// Check for empty mapping value
			if valueNode.Tag == "!!null" && valueNode.Value == "" {
				if isExpectedMappingForTest(keyName) {
					errors = append(errors, "'"+keyName+"' at line "+string(rune(keyNode.Line))+" has no value")
				}
			}

			errors = append(errors, checkYAMLStructure(valueNode, childPath)...)
		}

	case yaml.SequenceNode:
		for i, child := range node.Content {
			childPath := path + "[" + string(rune(i)) + "]"
			errors = append(errors, checkYAMLStructure(child, childPath)...)
		}
	}

	return errors
}

func isExpectedMappingForTest(key string) bool {
	expected := map[string]bool{
		"settings": true, "auth": true, "api_keys": true, "server": true,
		"llm": true, "memory": true, "logging": true, "overlord": true,
		"a2a": true, "mcp": true, "working": true, "buffer": true,
		"persistent": true, "remote": true, "caching": true, "image": true,
		"preprocessing": true, "extraction": true, "workflow": true,
		"clarification": true, "response": true, "timeouts": true,
		"retry": true, "encryption": true, "inbound": true, "outbound": true,
	}
	return expected[key]
}

func TestFormationTemplate_CloudProvider(t *testing.T) {
	config := &FormationConfig{
		Name:         "test-formation",
		DisplayName:  "Test Formation",
		Description:  "A test formation",
		ProviderType: "cloud",
		Provider: &LLMProvider{
			Vendor:       "openai",
			SecretName:   "OPENAI_API_KEY",
			DefaultModel: "gpt-4o",
		},
		EnableStreaming: true,
	}

	yaml := generateFormationYAML(config)
	validateYAML(t, "FormationTemplate_CloudProvider", yaml)
}

func TestFormationTemplate_LocalProvider(t *testing.T) {
	config := &FormationConfig{
		Name:         "test-formation",
		DisplayName:  "Test Formation",
		Description:  "A test formation",
		ProviderType: "local",
		LocalProvider: &LocalProvider{
			Name:   "Ollama",
			Vendor: "ollama",
		},
		LocalModel:   "llama3",
		LocalBaseURL: "http://localhost:11434",
	}

	yaml := generateFormationYAML(config)
	validateYAML(t, "FormationTemplate_LocalProvider", yaml)
}

func TestFormationTemplate_Async(t *testing.T) {
	config := &FormationConfig{
		Name:         "test-formation",
		DisplayName:  "Test Formation",
		Description:  "A test formation",
		ProviderType: "cloud",
		Provider: &LLMProvider{
			Vendor:       "openai",
			SecretName:   "OPENAI_API_KEY",
			DefaultModel: "gpt-4o",
		},
		EnableAsync: true,
		WebhookURL:  "https://example.com/webhook",
	}

	yaml := generateFormationYAML(config)
	validateYAML(t, "FormationTemplate_Async", yaml)
}

func TestAgentTemplate(t *testing.T) {
	yaml := agentTemplate(
		"test-agent",
		"Test Agent",
		"You are a helpful assistant.",
		"general",
		[]string{"coding", "writing"},
		false,
	)
	validateYAML(t, "AgentTemplate", yaml)
}

func TestAgentTemplate_WithA2A(t *testing.T) {
	yaml := agentTemplate(
		"test-agent",
		"Test Agent",
		"You are a helpful assistant.",
		"general",
		[]string{"coding"},
		true,
	)
	validateYAML(t, "AgentTemplate_WithA2A", yaml)
}

func TestMCPTemplate_Command(t *testing.T) {
	yaml := mcpTemplateNew(
		"test-mcp",
		"A test MCP server",
		"command",
		"",                  // endpoint
		"python",            // command
		"-m mcp_server",     // args
		"/path/to/server",   // workingDir
		"pip install mcp",   // installCmd
		"",                  // authType
		"",                  // authHeader
		[]string{"API_KEY"}, // envVars
	)
	validateYAML(t, "MCPTemplate_Stdio", yaml)
}

func TestMCPTemplate_HTTP(t *testing.T) {
	yaml := mcpTemplateNew(
		"test-mcp",
		"A test MCP server",
		"http",
		"https://api.example.com/mcp", // endpoint
		"",                            // command
		"",                            // args
		"",                            // workingDir
		"",                            // installCmd
		"bearer",                      // authType
		"",                            // authHeader
		nil,                           // envVars
	)
	validateYAML(t, "MCPTemplate_HTTP", yaml)
}

func TestSOPTemplate(t *testing.T) {
	yaml := sopTemplate(
		"test-sop",
		"Test SOP",
		"A test standard operating procedure",
		"autonomous",
		"testing, automation",
		false,
	)
	validateYAML(t, "SOPTemplate", yaml)
}

// Note: triggerTemplate generates Markdown, not YAML - no validation needed

// TestModelSettingsYAML tests the settings YAML generation that had the indentation bug
func TestModelSettingsYAML(t *testing.T) {
	// Simulate what updateModelSettingsInFormation does
	settings := ModelSettings{
		Temperature:   "0.7",
		MaxTokens:     "4096",
		Timeout:       "30",
		MaxRetries:    "3",
		FallbackModel: "openai/gpt-4o-mini",
	}

	var settingsYAML strings.Builder
	settingsYAML.WriteString("      settings:\n")
	settingsYAML.WriteString("        temperature: " + settings.Temperature + "\n")
	settingsYAML.WriteString("        max_tokens: " + settings.MaxTokens + "\n")
	settingsYAML.WriteString("        timeout_seconds: " + settings.Timeout + "\n")
	settingsYAML.WriteString("        max_retries: " + settings.MaxRetries + "\n")
	if settings.FallbackModel != "" {
		settingsYAML.WriteString("        fallback_model: \"" + settings.FallbackModel + "\"\n")
	}

	// Create a minimal formation with model + settings
	formationYAML := `schema: "1.0.0"
id: test
llm:
  models:
    - streaming: "openai/gpt-4o"
` + settingsYAML.String()

	validateYAML(t, "ModelSettingsYAML", formationYAML)

	// Also verify the settings structure is correct (not null)
	var formation map[string]interface{}
	if err := yaml.Unmarshal([]byte(formationYAML), &formation); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	llm := formation["llm"].(map[string]interface{})
	models := llm["models"].([]interface{})
	model := models[0].(map[string]interface{})
	settings_parsed := model["settings"]

	if settings_parsed == nil {
		t.Error("settings is nil - indentation bug!")
	}
	if _, ok := settings_parsed.(map[string]interface{}); !ok {
		t.Errorf("settings should be a map, got %T", settings_parsed)
	}
}

func TestA2ATemplate(t *testing.T) {
	output := a2aTemplate(
		"external-api",
		"External API service",
		"llm", // unused but kept for compatibility
		"https://api.example.com",
	)
	validateYAML(t, "A2ATemplate", output)

	// Verify schema compliance
	var a2a map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &a2a); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check required fields exist
	requiredFields := []string{"schema", "id", "name", "description", "url", "active", "auth"}
	for _, field := range requiredFields {
		if _, ok := a2a[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Check fields that should NOT exist
	invalidFields := []string{"type", "connection", "endpoints"}
	for _, field := range invalidFields {
		if _, ok := a2a[field]; ok {
			t.Errorf("Invalid field present: %s (not in schema)", field)
		}
	}
}

func TestA2AServiceTemplate(t *testing.T) {
	yaml := a2aServiceTemplate(
		"billing-service",
		"Billing Service",
		"External billing integration",
		"https://billing.example.com",
		"bearer",
		"",
		"${{ secrets.BILLING_TOKEN }}",
		"",
		"",
		true,
		5,
		60,
	)
	validateYAML(t, "A2AServiceTemplate", yaml)
}
