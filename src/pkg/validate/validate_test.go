package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"a", true},
		{"my-formation", true},
		{"my-app-123", true},
		{"test", true},
		{"a1", true},
		{"abc123", true},
		{"my-cool-app", true},
		{"", false},
		{"-invalid", false},
		{"invalid-", false},
		{"UPPERCASE", false},
		{"has_underscore", false},
		{"has.dot", false},
		{"has space", false},
		{"123start", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := isValidID(tt.id)
			if got != tt.valid {
				t.Errorf("isValidID(%q) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}

func TestCollectSecretRefs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no secrets",
			content: "just some text",
			want:    []string{},
		},
		{
			name:    "single secret",
			content: "api_key: ${{ secrets.OPENAI_KEY }}",
			want:    []string{"OPENAI_KEY"},
		},
		{
			name:    "multiple secrets",
			content: "key1: ${{ secrets.KEY_A }}\nkey2: ${{ secrets.KEY_B }}",
			want:    []string{"KEY_A", "KEY_B"},
		},
		{
			name:    "duplicate secrets",
			content: "${{ secrets.SAME }} and ${{ secrets.SAME }}",
			want:    []string{"SAME"},
		},
		{
			name:    "with spaces",
			content: "${{secrets.NO_SPACE}} and ${{  secrets.WITH_SPACE  }}",
			want:    []string{"NO_SPACE", "WITH_SPACE"},
		},
		{
			name:    "underscore in name",
			content: "${{ secrets.MY_API_KEY_123 }}",
			want:    []string{"MY_API_KEY_123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectSecretRefs(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("collectSecretRefs() got %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("collectSecretRefs()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestIsExpectedMapping(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"settings", true},
		{"server", true},
		{"llm", true},
		{"api_keys", true},
		{"memory", true},
		{"random_key", false},
		{"foo", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := isExpectedMapping(tt.key)
			if got != tt.expected {
				t.Errorf("isExpectedMapping(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestResultIsValid(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "empty result is valid",
			result: Result{Errors: []Issue{}, Warnings: []Issue{}},
			want:   true,
		},
		{
			name: "warnings only is valid",
			result: Result{
				Errors:   []Issue{},
				Warnings: []Issue{{Message: "warning"}},
			},
			want: true,
		},
		{
			name: "errors makes invalid",
			result: Result{
				Errors:   []Issue{{Message: "error"}},
				Warnings: []Issue{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsValid()
			if got != tt.want {
				t.Errorf("Result.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormationValidation(t *testing.T) {
	t.Run("missing formation file", func(t *testing.T) {
		tmpDir := t.TempDir()
		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsValid() {
			t.Error("expected invalid result for missing formation file")
		}
		if len(result.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(result.Errors))
		}
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(formationPath, []byte("invalid: yaml: content:"), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsValid() {
			t.Error("expected invalid result for bad YAML")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		os.WriteFile(formationPath, []byte("name: test\n"), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsValid() {
			t.Error("expected invalid result for missing required fields")
		}

		// Should have errors for missing schema and id
		hasSchemaError := false
		hasIDError := false
		for _, e := range result.Errors {
			if e.Field == "schema" {
				hasSchemaError = true
			}
			if e.Field == "id" {
				hasIDError = true
			}
		}
		if !hasSchemaError {
			t.Error("expected error for missing schema field")
		}
		if !hasIDError {
			t.Error("expected error for missing id field")
		}
	})

	t.Run("invalid id format", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "1.0.0"
id: "INVALID_ID"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsValid() {
			t.Error("expected invalid result for bad id format")
		}
	})

	t.Run("valid minimal formation", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "1.0.0"
id: "my-formation"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsValid() {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "1.0.0"
id: "my-formation"
server:
  port: 99999
`
		os.WriteFile(formationPath, []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasPortError := false
		for _, e := range result.Errors {
			if e.Field == "server.port" {
				hasPortError = true
			}
		}
		if !hasPortError {
			t.Error("expected error for invalid port")
		}
	})

	t.Run("unknown schema version warning", func(t *testing.T) {
		tmpDir := t.TempDir()
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "2.0.0"
id: "my-formation"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasSchemaWarning := false
		for _, w := range result.Warnings {
			if w.Field == "schema" {
				hasSchemaWarning = true
			}
		}
		if !hasSchemaWarning {
			t.Error("expected warning for unknown schema version")
		}
	})
}

func TestValidateAgents(t *testing.T) {
	t.Run("agent missing id", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create formation.yaml
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "1.0.0"
id: "my-formation"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		// Create agents directory with invalid agent
		agentsDir := filepath.Join(tmpDir, "agents")
		os.Mkdir(agentsDir, 0755)
		agentPath := filepath.Join(agentsDir, "test-agent.yaml")
		os.WriteFile(agentPath, []byte("role: helper\n"), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasAgentError := false
		for _, e := range result.Errors {
			if e.File == "agents/test-agent.yaml" && e.Field == "id" {
				hasAgentError = true
			}
		}
		if !hasAgentError {
			t.Error("expected error for agent missing id")
		}
	})
}

func TestValidateMCPs(t *testing.T) {
	t.Run("mcp missing required fields", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create formation.yaml
		formationPath := filepath.Join(tmpDir, "formation.yaml")
		content := `schema: "1.0.0"
id: "my-formation"
`
		os.WriteFile(formationPath, []byte(content), 0644)

		// Create mcps directory with invalid mcp
		mcpsDir := filepath.Join(tmpDir, "mcps")
		os.Mkdir(mcpsDir, 0755)
		mcpPath := filepath.Join(mcpsDir, "test-mcp.yaml")
		os.WriteFile(mcpPath, []byte("name: test\n"), 0644)

		result, err := Formation(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasMCPIDError := false
		hasMCPTypeError := false
		for _, e := range result.Errors {
			if e.File == "mcps/test-mcp.yaml" {
				if e.Field == "id" {
					hasMCPIDError = true
				}
				if e.Field == "type" {
					hasMCPTypeError = true
				}
			}
		}
		if !hasMCPIDError {
			t.Error("expected error for mcp missing id")
		}
		if !hasMCPTypeError {
			t.Error("expected error for mcp missing type")
		}
	})
}
