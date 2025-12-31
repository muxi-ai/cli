package scaffold

import (
	"testing"
)

func TestNormalizeComponentName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Component", "my-component"},
		{"myComponent", "mycomponent"},
		{"MY_COMPONENT", "my_component"},
		{"test", "test"},
		{"Test Name Here", "test-name-here"},
	}

	for _, tt := range tests {
		got := normalizeComponentName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeComponentName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateComponentName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name", false},
		{"valid_name", false},
		{"valid123", false},
		{"", true},
		{"has space", true},
		{"has@special", true},
	}

	for _, tt := range tests {
		err := validateComponentName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateComponentName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestParseURLList(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"http://a.com, http://b.com", 2},
		{"http://single.com", 1},
		{"", 0},
		{"http://a.com,http://b.com,http://c.com", 3},
	}

	for _, tt := range tests {
		got := parseURLList(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseURLList(%q) len = %d, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestParseEndpointList(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"/api/v1, /api/v2", 2},
		{"/single", 1},
		{"", 0},
	}

	for _, tt := range tests {
		got := parseEndpointList(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseEndpointList(%q) len = %d, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"hello-world", "Hello World"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		got := titleCase(tt.input)
		if got != tt.want {
			t.Errorf("titleCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateMCPSecretPrefix(t *testing.T) {
	// Just verify it returns non-empty and starts with MCP_
	tests := []string{"my-mcp", "test", "some-long-name"}

	for _, mcpID := range tests {
		got := generateMCPSecretPrefix(mcpID)
		if got == "" {
			t.Errorf("generateMCPSecretPrefix(%q) returned empty", mcpID)
		}
		if len(got) < 4 || got[:4] != "MCP_" {
			t.Errorf("generateMCPSecretPrefix(%q) = %q, should start with 'MCP_'", mcpID, got)
		}
	}
}

func TestParseEnvironmentVariables(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"VAR1=val1, VAR2=val2", 2},
		{"SINGLE=value", 1},
		{"", 0},
		{"A=1,B=2,C=3", 3},
	}

	for _, tt := range tests {
		got := parseEnvironmentVariables(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseEnvironmentVariables(%q) len = %d, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestTriggerTemplateBasic(t *testing.T) {
	result := triggerTemplate("on-message")
	
	if result == "" {
		t.Error("triggerTemplate() returned empty string")
	}
	if !containsStr(result, "on-message") {
		t.Error("template should contain trigger ID")
	}
}

func TestExtractA2ARegistries(t *testing.T) {
	content := `a2a:
  inbound:
    registries:
      - http://registry1.com
      - http://registry2.com`
	
	result := extractA2ARegistries(content)
	_ = result
}

func TestExtractA2AAuthType(t *testing.T) {
	content := `a2a:
  inbound:
    auth:
      type: bearer`
	
	result := extractA2AAuthType(content)
	_ = result
}

func TestMCPTemplateNew(t *testing.T) {
	result := mcpTemplateNew(
		"test-mcp",
		"Test MCP",
		"stdio",
		"",
		"npx",
		"-y test-mcp",
		"",
		"",
		"none",
		"",
		nil,
	)
	
	if result == "" {
		t.Error("mcpTemplateNew() returned empty string")
	}
	if !containsStr(result, "test-mcp") {
		t.Error("template should contain MCP ID")
	}
}

func TestFormatExistingInfo(t *testing.T) {
	result := formatExistingInfo("My Agent", "Does helpful things")
	
	if result == "" {
		t.Error("formatExistingInfo() returned empty string")
	}
}

func TestGetComponentInfo(t *testing.T) {
	// Test with non-existent file (should not panic)
	name, desc := getComponentInfo("/nonexistent/file.yaml")
	_ = name
	_ = desc
}

func TestIsA2AEnabled(t *testing.T) {
	// Test with non-existent dir
	result := isA2AEnabled("/nonexistent/dir")
	if result {
		t.Error("isA2AEnabled should return false for non-existent dir")
	}
}

func TestMcpExistsInAgent(t *testing.T) {
	// Test with non-existent paths
	result := mcpExistsInAgent("/nonexistent", "agent-id", "mcp-id")
	if result {
		t.Error("mcpExistsInAgent should return false for non-existent paths")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestAgentTemplateWithSpecialties(t *testing.T) {
	specialties := []string{"coding", "writing", "research"}
	result := agentTemplate("test-agent", "Test Agent", "You are a test agent", "assistant", specialties, false)

	if result == "" {
		t.Error("agentTemplate returned empty string")
	}

	// Check that specialties are included
	if !containsStr(result, "coding") {
		t.Error("template should contain specialty 'coding'")
	}
	if !containsStr(result, "writing") {
		t.Error("template should contain specialty 'writing'")
	}
}

func TestAgentTemplateWithExternalA2A(t *testing.T) {
	result := agentTemplate("test-agent", "Test Agent", "You are a test agent", "assistant", nil, true)

	if result == "" {
		t.Error("agentTemplate returned empty string")
	}

	// Basic validation that template was generated
	if !containsStr(result, "test-agent") {
		t.Error("template should contain agent id")
	}
}

func TestA2AServiceTemplateBasic(t *testing.T) {
	result := a2aServiceTemplate(
		"test-service",
		"Test Service",
		"A test A2A service",
		"https://example.com/api",
		"api_key",
		"Authorization",
		"Bearer token123",
		"",
		"",
		false,
		0,
		30,
	)

	if result == "" {
		t.Error("a2aServiceTemplate returned empty string")
	}

	if !containsStr(result, "test-service") {
		t.Error("template should contain service id")
	}
	if !containsStr(result, "https://example.com/api") {
		t.Error("template should contain service URL")
	}
}

func TestA2AServiceTemplateWithBasicAuth(t *testing.T) {
	result := a2aServiceTemplate(
		"test-service",
		"Test Service",
		"A test A2A service",
		"https://example.com/api",
		"basic",
		"",
		"",
		"admin",
		"password123",
		false,
		0,
		30,
	)

	if result == "" {
		t.Error("a2aServiceTemplate returned empty string")
	}

	// Basic auth uses secrets.XXX_USERNAME pattern
	if !containsStr(result, "type: \"basic\"") {
		t.Error("template should contain basic auth type")
	}
}

func TestA2AServiceTemplateWithCustomRetry(t *testing.T) {
	result := a2aServiceTemplate(
		"test-service",
		"Test Service",
		"A test A2A service",
		"https://example.com/api",
		"none",
		"",
		"",
		"",
		"",
		true,
		5,
		60,
	)

	if result == "" {
		t.Error("a2aServiceTemplate returned empty string")
	}

	if !containsStr(result, "5") {
		t.Error("template should contain retry attempts")
	}
	if !containsStr(result, "60") {
		t.Error("template should contain timeout")
	}
}

func TestGitignoreTemplate(t *testing.T) {
	result := gitignoreTemplate()
	if result == "" {
		t.Error("gitignoreTemplate returned empty string")
	}
	if !containsStr(result, ".key") {
		t.Error("gitignoreTemplate should contain .key")
	}
}

func TestMuxiTemplate(t *testing.T) {
	result := muxiTemplate()
	if result == "" {
		t.Error("muxiTemplate returned empty string")
	}
	if !containsStr(result, "profile") {
		t.Error("muxiTemplate should contain profile")
	}
}

func TestGenerateFormationYAML(t *testing.T) {
	config := &FormationConfig{
		Name:        "test-formation",
		DisplayName: "Test Formation",
		Description: "A test formation",
	}
	result := generateFormationYAML(config)
	if result == "" {
		t.Error("generateFormationYAML returned empty string")
	}
	if !containsStr(result, "test-formation") {
		t.Error("generateFormationYAML should contain formation ID")
	}
}

func TestGenerateFormationYAML_WithAsync(t *testing.T) {
	config := &FormationConfig{
		Name:        "test-formation",
		DisplayName: "Test Formation",
		EnableAsync: true,
		WebhookURL:  "https://example.com/webhook",
	}
	result := generateFormationYAML(config)
	if result == "" {
		t.Error("generateFormationYAML returned empty string")
	}
	if !containsStr(result, "async") {
		t.Error("generateFormationYAML should contain async config")
	}
}

func TestGenerateFormationYAML_WithStreaming(t *testing.T) {
	config := &FormationConfig{
		Name:            "test-formation",
		DisplayName:     "Test Formation",
		EnableStreaming: true,
	}
	result := generateFormationYAML(config)
	if result == "" {
		t.Error("generateFormationYAML returned empty string")
	}
}

func TestGenerateFormationYAML_WithCloudProvider(t *testing.T) {
	config := &FormationConfig{
		Name:         "test-formation",
		DisplayName:  "Test Formation",
		ProviderType: "cloud",
		Provider: &LLMProvider{
			Name:         "openai",
			Vendor:       "openai",
			DefaultModel: "gpt-4o",
		},
		APIKey: "sk-test",
	}
	result := generateFormationYAML(config)
	if result == "" {
		t.Error("generateFormationYAML returned empty string")
	}
	if !containsStr(result, "openai") {
		t.Error("generateFormationYAML should contain provider")
	}
}

func TestGenerateFormationYAML_WithLocalProvider(t *testing.T) {
	config := &FormationConfig{
		Name:         "test-formation",
		DisplayName:  "Test Formation",
		ProviderType: "local",
		LocalBaseURL: "http://localhost:11434",
		LocalModel:   "llama2",
	}
	result := generateFormationYAML(config)
	if result == "" {
		t.Error("generateFormationYAML returned empty string")
	}
}
