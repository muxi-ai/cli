package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDeclarations_AgentDeclaredAndExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create formation with declared agent
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
agents:
  - my-agent
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// Create matching agent file
	agentsDir := filepath.Join(tmpDir, "agents")
	os.Mkdir(agentsDir, 0755)
	agentContent := `id: "my-agent"
role: "generalist"
`
	os.WriteFile(filepath.Join(agentsDir, "my-agent.yaml"), []byte(agentContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have no declaration-related errors or warnings
	for _, e := range result.Errors {
		if e.Field == "agents" {
			t.Errorf("unexpected agents error: %s", e.Message)
		}
	}
	for _, w := range result.Warnings {
		if w.Field == "agents" {
			t.Errorf("unexpected agents warning: %s", w.Message)
		}
	}
}

func TestValidateDeclarations_AgentDeclaredButMissing(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
agents:
  - ghost-agent
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// No agents directory or file
	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "agents" && contains(e.Message, "ghost-agent") && contains(e.Message, "no matching file") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for declared agent with no matching file")
	}
}

func TestValidateDeclarations_AgentFileNotDeclared(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// Create agent file without declaring it
	agentsDir := filepath.Join(tmpDir, "agents")
	os.Mkdir(agentsDir, 0755)
	agentContent := `id: "orphan-agent"
role: "generalist"
`
	os.WriteFile(filepath.Join(agentsDir, "orphan-agent.yaml"), []byte(agentContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range result.Warnings {
		if w.Field == "agents" && contains(w.Message, "orphan-agent") && contains(w.Message, "not declared") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for agent file that is not declared in formation manifest")
	}
}

func TestValidateDeclarations_MCPDeclaredAndExists(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
mcp:
  servers:
    - my-mcp
`
	os.WriteFile(formationPath, []byte(content), 0644)

	mcpsDir := filepath.Join(tmpDir, "mcps")
	os.Mkdir(mcpsDir, 0755)
	mcpContent := `id: "my-mcp"
type: "http"
endpoint: "https://example.com"
`
	os.WriteFile(filepath.Join(mcpsDir, "my-mcp.yaml"), []byte(mcpContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Errors {
		if e.Field == "mcp.servers" {
			t.Errorf("unexpected mcp.servers error: %s", e.Message)
		}
	}
	for _, w := range result.Warnings {
		if w.Field == "mcp.servers" {
			t.Errorf("unexpected mcp.servers warning: %s", w.Message)
		}
	}
}

func TestValidateDeclarations_MCPDeclaredButMissing(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
mcp:
  servers:
    - ghost-mcp
`
	os.WriteFile(formationPath, []byte(content), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "mcp.servers" && contains(e.Message, "ghost-mcp") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for declared MCP with no matching file")
	}
}

func TestValidateDeclarations_MCPFileNotDeclared(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	mcpsDir := filepath.Join(tmpDir, "mcps")
	os.Mkdir(mcpsDir, 0755)
	mcpContent := `id: "orphan-mcp"
type: "command"
`
	os.WriteFile(filepath.Join(mcpsDir, "orphan-mcp.yaml"), []byte(mcpContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range result.Warnings {
		if w.Field == "mcp.servers" && contains(w.Message, "orphan-mcp") && contains(w.Message, "not declared") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for MCP file that is not declared")
	}
}

func TestValidateDeclarations_A2ADeclaredAndExists(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
a2a:
  outbound:
    services:
      - my-service
`
	os.WriteFile(formationPath, []byte(content), 0644)

	a2aDir := filepath.Join(tmpDir, "a2a")
	os.Mkdir(a2aDir, 0755)
	svcContent := `id: "my-service"
url: "https://example.com"
`
	os.WriteFile(filepath.Join(a2aDir, "my-service.yaml"), []byte(svcContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Errors {
		if e.Field == "a2a.outbound.services" {
			t.Errorf("unexpected a2a error: %s", e.Message)
		}
	}
}

func TestValidateDeclarations_A2ADeclaredButMissing(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
a2a:
  outbound:
    services:
      - ghost-service
`
	os.WriteFile(formationPath, []byte(content), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "a2a.outbound.services" && contains(e.Message, "ghost-service") {
			found = true
		}
	}
	if !found {
		t.Error("expected error for declared A2A service with no matching file")
	}
}

func TestValidateDeclarations_FileIDFromFilename(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
agents:
  - no-id-agent
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// Create agent file without an id field - should use filename stem
	agentsDir := filepath.Join(tmpDir, "agents")
	os.Mkdir(agentsDir, 0755)
	agentContent := `role: "generalist"
system_message: "I have no id field"
`
	os.WriteFile(filepath.Join(agentsDir, "no-id-agent.yaml"), []byte(agentContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT have a "declared but no matching file" error since filename stem matches
	for _, e := range result.Errors {
		if e.Field == "agents" && contains(e.Message, "no-id-agent") && contains(e.Message, "no matching file") {
			t.Error("should match agent by filename stem when id field is missing")
		}
	}
}

func TestValidateDeclarations_AFS_Extension(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.afs")
	content := `schema: "1.0.0"
id: "my-formation"
agents:
  - my-agent
`
	os.WriteFile(formationPath, []byte(content), 0644)

	agentsDir := filepath.Join(tmpDir, "agents")
	os.Mkdir(agentsDir, 0755)
	agentContent := `id: "my-agent"
role: "generalist"
`
	os.WriteFile(filepath.Join(agentsDir, "my-agent.afs"), []byte(agentContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Errors {
		if e.Field == "agents" && contains(e.Message, "my-agent") {
			t.Errorf("unexpected agents error with .afs files: %s", e.Message)
		}
	}
}

func TestValidateDeclarations_NoAgentsNoDirNoWarning(t *testing.T) {
	tmpDir := t.TempDir()

	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// No agents dir at all
	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, w := range result.Warnings {
		if w.Field == "agents" {
			t.Errorf("unexpected agents warning when no agents dir exists: %s", w.Message)
		}
	}
}

func TestValidateDeclarations_InlineDictIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Dict entries in agents list should be ignored (only strings are checked)
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
agents:
  - my-agent
  - id: "inline-agent"
    role: "helper"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	agentsDir := filepath.Join(tmpDir, "agents")
	os.Mkdir(agentsDir, 0755)
	agentContent := `id: "my-agent"
role: "generalist"
`
	os.WriteFile(filepath.Join(agentsDir, "my-agent.yaml"), []byte(agentContent), 0644)

	result, err := Formation(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not error about inline-agent not having a file
	for _, e := range result.Errors {
		if e.Field == "agents" && contains(e.Message, "inline-agent") {
			t.Error("inline dict entries should not be checked against files")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
