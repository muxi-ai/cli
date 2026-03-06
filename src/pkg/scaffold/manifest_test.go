package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddToTopLevelList_NewSection(t *testing.T) {
	content := `schema: "1.0.0"
id: "my-formation"
`
	result := addToTopLevelList(content, "agents", "my-agent")

	if !strings.Contains(result, "agents:\n  - my-agent") {
		t.Errorf("expected agents list with my-agent, got:\n%s", result)
	}
}

func TestAddToTopLevelList_ExistingSection(t *testing.T) {
	content := `schema: "1.0.0"
id: "my-formation"

agents:
  - existing-agent
`
	result := addToTopLevelList(content, "agents", "new-agent")

	if !strings.Contains(result, "- existing-agent") {
		t.Errorf("expected existing-agent to remain, got:\n%s", result)
	}
	if !strings.Contains(result, "- new-agent") {
		t.Errorf("expected new-agent to be added, got:\n%s", result)
	}
}

func TestAddToTopLevelList_Dedup(t *testing.T) {
	content := `schema: "1.0.0"

agents:
  - my-agent
`
	result := addToTopLevelList(content, "agents", "my-agent")

	count := strings.Count(result, "- my-agent")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of my-agent, got %d", count)
	}
}

func TestAddToTopLevelList_DedupQuoted(t *testing.T) {
	content := `agents:
  - "my-agent"
`
	result := addToTopLevelList(content, "agents", "my-agent")

	if strings.Count(result, "my-agent") != 1 {
		t.Errorf("expected dedup with quoted ID, got:\n%s", result)
	}
}

func TestAddToTopLevelList_MultipleAgents(t *testing.T) {
	content := `agents:
  - agent-one
`
	result := addToTopLevelList(content, "agents", "agent-two")
	result = addToTopLevelList(result, "agents", "agent-three")

	if !strings.Contains(result, "- agent-one") ||
		!strings.Contains(result, "- agent-two") ||
		!strings.Contains(result, "- agent-three") {
		t.Errorf("expected all three agents, got:\n%s", result)
	}
}

func TestAddToNestedList_NewSection(t *testing.T) {
	content := `schema: "1.0.0"
id: "my-formation"
`
	result := addToNestedList(content, "mcp", "servers", "my-mcp")

	if !strings.Contains(result, "mcp:\n  servers:\n    - my-mcp") {
		t.Errorf("expected mcp.servers with my-mcp, got:\n%s", result)
	}
}

func TestAddToNestedList_ExistingParent(t *testing.T) {
	content := `mcp:
  timeout: 30
`
	result := addToNestedList(content, "mcp", "servers", "my-mcp")

	if !strings.Contains(result, "servers:\n    - my-mcp") {
		t.Errorf("expected servers list added under mcp, got:\n%s", result)
	}
	if !strings.Contains(result, "timeout: 30") {
		t.Errorf("expected existing timeout to remain, got:\n%s", result)
	}
}

func TestAddToNestedList_ExistingList(t *testing.T) {
	content := `mcp:
  servers:
    - existing-mcp
`
	result := addToNestedList(content, "mcp", "servers", "new-mcp")

	if !strings.Contains(result, "- existing-mcp") {
		t.Errorf("expected existing-mcp to remain, got:\n%s", result)
	}
	if !strings.Contains(result, "- new-mcp") {
		t.Errorf("expected new-mcp to be added, got:\n%s", result)
	}
}

func TestAddToNestedList_Dedup(t *testing.T) {
	content := `mcp:
  servers:
    - my-mcp
`
	result := addToNestedList(content, "mcp", "servers", "my-mcp")

	count := strings.Count(result, "- my-mcp")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of my-mcp, got %d", count)
	}
}

func TestAddToDeepNestedList_NewSection(t *testing.T) {
	content := `schema: "1.0.0"
id: "my-formation"
`
	result := addToDeepNestedList(content, "a2a", "outbound", "services", "my-service")

	if !strings.Contains(result, "a2a:\n  outbound:\n    services:\n      - my-service") {
		t.Errorf("expected a2a.outbound.services with my-service, got:\n%s", result)
	}
}

func TestAddToDeepNestedList_ExistingA2A(t *testing.T) {
	content := `a2a:
  inbound:
    enabled: true
`
	result := addToDeepNestedList(content, "a2a", "outbound", "services", "my-service")

	if !strings.Contains(result, "outbound:\n    services:\n      - my-service") {
		t.Errorf("expected outbound.services added under a2a, got:\n%s", result)
	}
	if !strings.Contains(result, "inbound:") {
		t.Errorf("expected existing inbound to remain, got:\n%s", result)
	}
}

func TestAddToDeepNestedList_ExistingOutbound(t *testing.T) {
	content := `a2a:
  outbound:
    enabled: true
`
	result := addToDeepNestedList(content, "a2a", "outbound", "services", "my-service")

	if !strings.Contains(result, "services:\n      - my-service") {
		t.Errorf("expected services list added under outbound, got:\n%s", result)
	}
}

func TestAddToDeepNestedList_ExistingServices(t *testing.T) {
	content := `a2a:
  outbound:
    services:
      - existing-service
`
	result := addToDeepNestedList(content, "a2a", "outbound", "services", "new-service")

	if !strings.Contains(result, "- existing-service") {
		t.Errorf("expected existing-service to remain, got:\n%s", result)
	}
	if !strings.Contains(result, "- new-service") {
		t.Errorf("expected new-service to be added, got:\n%s", result)
	}
}

func TestAddToDeepNestedList_Dedup(t *testing.T) {
	content := `a2a:
  outbound:
    services:
      - my-service
`
	result := addToDeepNestedList(content, "a2a", "outbound", "services", "my-service")

	count := strings.Count(result, "- my-service")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of my-service, got %d", count)
	}
}

func TestAddComponentToFormation_Agents(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "test"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	err := AddComponentToFormation(tmpDir, "agents", "my-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(formationPath)
	if !strings.Contains(string(data), "agents:\n  - my-agent") {
		t.Errorf("expected agents list in file, got:\n%s", string(data))
	}
}

func TestAddComponentToFormation_MCPServers(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "test"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	err := AddComponentToFormation(tmpDir, "mcp.servers", "my-mcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(formationPath)
	if !strings.Contains(string(data), "mcp:\n  servers:\n    - my-mcp") {
		t.Errorf("expected mcp.servers list in file, got:\n%s", string(data))
	}
}

func TestAddComponentToFormation_A2AServices(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "test"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	err := AddComponentToFormation(tmpDir, "a2a.outbound.services", "my-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(formationPath)
	if !strings.Contains(string(data), "a2a:\n  outbound:\n    services:\n      - my-service") {
		t.Errorf("expected a2a.outbound.services list in file, got:\n%s", string(data))
	}
}

func TestAddComponentToFormation_AFS(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.afs")
	content := `schema: "1.0.0"
id: "test"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	err := AddComponentToFormation(tmpDir, "agents", "my-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(formationPath)
	if !strings.Contains(string(data), "agents:\n  - my-agent") {
		t.Errorf("expected agents list in .afs file, got:\n%s", string(data))
	}
}

func TestAddComponentToFormation_InvalidSection(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	os.WriteFile(formationPath, []byte("id: test\n"), 0644)

	err := AddComponentToFormation(tmpDir, "invalid.section", "my-id")
	if err == nil {
		t.Error("expected error for invalid section")
	}
}

func TestAddComponentToFormation_NoFormationFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := AddComponentToFormation(tmpDir, "agents", "my-agent")
	if err == nil {
		t.Error("expected error when formation file not found")
	}
}

func TestAddComponentToFormation_PreservesExistingContent(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "my-formation"
name: "My Formation"

overlord:
  response:
    streaming: true

llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"
  models:
    - text: "openai/gpt-4o"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	err := AddComponentToFormation(tmpDir, "agents", "my-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(formationPath)
	result := string(data)

	// Check existing content preserved
	if !strings.Contains(result, "name: \"My Formation\"") {
		t.Error("name field was lost")
	}
	if !strings.Contains(result, "streaming: true") {
		t.Error("overlord config was lost")
	}
	if !strings.Contains(result, "openai/gpt-4o") {
		t.Error("LLM config was lost")
	}
	if !strings.Contains(result, "agents:\n  - my-agent") {
		t.Errorf("agents list not added correctly, got:\n%s", result)
	}
}

func TestAddComponentToFormation_SequentialAdds(t *testing.T) {
	tmpDir := t.TempDir()
	formationPath := filepath.Join(tmpDir, "formation.yaml")
	content := `schema: "1.0.0"
id: "test"
`
	os.WriteFile(formationPath, []byte(content), 0644)

	// Add agent
	AddComponentToFormation(tmpDir, "agents", "agent-one")
	AddComponentToFormation(tmpDir, "agents", "agent-two")
	AddComponentToFormation(tmpDir, "mcp.servers", "mcp-one")

	data, _ := os.ReadFile(formationPath)
	result := string(data)

	if !strings.Contains(result, "- agent-one") {
		t.Error("agent-one missing")
	}
	if !strings.Contains(result, "- agent-two") {
		t.Error("agent-two missing")
	}
	if !strings.Contains(result, "- mcp-one") {
		t.Error("mcp-one missing")
	}
}

func TestCountIndent(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"no indent", 0},
		{"  two spaces", 2},
		{"    four spaces", 4},
		{"\ttab", 1},
		{"", 0},
	}
	for _, tt := range tests {
		got := countIndent(tt.line)
		if got != tt.want {
			t.Errorf("countIndent(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestInsertLine(t *testing.T) {
	lines := []string{"a", "b", "c"}

	result := insertLine(lines, 1, "x")
	if len(result) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(result))
	}
	if result[1] != "x" {
		t.Errorf("expected 'x' at position 1, got %q", result[1])
	}
}

func TestInsertLine_AtEnd(t *testing.T) {
	lines := []string{"a", "b"}

	result := insertLine(lines, 5, "x")
	if result[len(result)-1] != "x" {
		t.Errorf("expected 'x' at end, got %q", result[len(result)-1])
	}
}

func TestAddToTopLevelList_WithComments(t *testing.T) {
	content := `schema: "1.0.0"

# Agents section
agents:
  - existing-agent
  # This is a comment

scheduler:
  timezone: "UTC"
`
	result := addToTopLevelList(content, "agents", "new-agent")

	if !strings.Contains(result, "- existing-agent") {
		t.Error("existing-agent lost")
	}
	if !strings.Contains(result, "- new-agent") {
		t.Errorf("new-agent not added, got:\n%s", result)
	}
	if !strings.Contains(result, "scheduler:") {
		t.Error("scheduler section lost")
	}
}

func TestAddToNestedList_WithOtherSections(t *testing.T) {
	content := `mcp:
  servers:
    - existing-mcp
  timeout: 30

scheduler:
  timezone: "UTC"
`
	result := addToNestedList(content, "mcp", "servers", "new-mcp")

	if !strings.Contains(result, "- existing-mcp") {
		t.Error("existing-mcp lost")
	}
	if !strings.Contains(result, "- new-mcp") {
		t.Errorf("new-mcp not added, got:\n%s", result)
	}
	if !strings.Contains(result, "scheduler:") {
		t.Error("scheduler section lost")
	}
}
