package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
)

// CreateAgent creates a new agent configuration file
func CreateAgent(name string, noWizard bool) error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock("Not in formation directory", err.Error(), "")
		return fmt.Errorf("not in formation directory")
	}

	// Validate name
	if err := validateComponentName(name); err != nil {
		ui.ErrorBlock("Invalid agent name", err.Error(), "Example: weather-agent")
		return fmt.Errorf("invalid name")
	}

	// Check if file already exists
	agentFile := filepath.Join(ctx.RootDir, "agents", name+".yaml")
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		ui.ErrorBlock(
			"Agent file exists",
			fmt.Sprintf("File 'agents/%s.yaml' already exists", name),
			fmt.Sprintf("Choose a different name or remove:\n  rm agents/%s.yaml", name),
		)
		return fmt.Errorf("file exists")
	}

	// Interactive mode - get optional fields
	var description, model string
	var maxTokens int

	if !noWizard {
		description, _ = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}

		model, _ = wizard.PromptString("Model", "gpt-4o", nil)
		ui.PromptSuccess("Model", model)

		maxTokensStr, _ := wizard.PromptString("Max tokens", "4000", nil)
		fmt.Sscanf(maxTokensStr, "%d", &maxTokens)
		if maxTokens == 0 {
			maxTokens = 4000
		}
		ui.PromptSuccess("Max tokens", fmt.Sprintf("%d", maxTokens))
	} else {
		// Non-interactive defaults
		description = ""
		model = "gpt-4o"
		maxTokens = 4000
	}

	// Create agent file
	content := agentTemplate(name, description, model, maxTokens)
	if err := os.WriteFile(agentFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create agent file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created agents/%s.yaml", name))

	if !noWizard {
		fmt.Println()
		ui.Dimmed("Edit the file to:")
		ui.Dimmed("  • Add tools and capabilities")
		ui.Dimmed("  • Configure prompts and persona")
		ui.Dimmed("  • Set up workflows")
	}

	return nil
}

// validateComponentName validates component name format
func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Pattern: lowercase letter, then lowercase letters/numbers/hyphens/underscores, 3-50 chars
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-_]{2,49}$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("component names must:\n  • Be lowercase\n  • Start with a letter\n  • Contain only letters, numbers, hyphens, and underscores\n  • Be 3-50 characters long")
	}

	return nil
}

func agentTemplate(name, description, model string, maxTokens int) string {
	if description == "" {
		description = fmt.Sprintf("%s agent", strings.ReplaceAll(name, "-", " "))
	}

	return fmt.Sprintf(`id: %s
description: "%s"
active: true

model:
  name: "openai/%s"
  max_tokens: %d
  temperature: 0.7

persona: |
  You are a helpful AI assistant.

tools: []
workflows: []
`, name, description, model, maxTokens)
}

// CreateMCP creates a new MCP server configuration file
func CreateMCP(name string, noWizard bool) error {
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock("Not in formation directory", err.Error(), "")
		return fmt.Errorf("not in formation directory")
	}

	if err := validateComponentName(name); err != nil {
		ui.ErrorBlock("Invalid MCP server name", err.Error(), "Example: postgres-db")
		return fmt.Errorf("invalid name")
	}

	mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
	if _, err := os.Stat(mcpFile); !os.IsNotExist(err) {
		ui.ErrorBlock(
			"MCP file exists",
			fmt.Sprintf("File 'mcps/%s.yaml' already exists", name),
			fmt.Sprintf("Choose a different name or remove:\n  rm mcps/%s.yaml", name),
		)
		return fmt.Errorf("file exists")
	}

	var description, mcpType, command string
	if !noWizard {
		description, _ = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}

		mcpType, _ = wizard.PromptString("Type", "stdio", nil)
		ui.PromptSuccess("Type", mcpType)

		command, _ = wizard.PromptString("Command", "", nil)
		ui.PromptSuccess("Command", command)
	} else {
		description = ""
		mcpType = "stdio"
		command = "mcp-server"
	}

	content := mcpTemplate(name, description, mcpType, command)
	if err := os.WriteFile(mcpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create MCP file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created mcps/%s.yaml", name))

	if !noWizard {
		fmt.Println()
		ui.Dimmed("Edit the file to configure connection details")
	}

	return nil
}

// CreateSOP creates a new SOP document
func CreateSOP(name string, noWizard bool) error {
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock("Not in formation directory", err.Error(), "")
		return fmt.Errorf("not in formation directory")
	}

	if err := validateComponentName(name); err != nil {
		ui.ErrorBlock("Invalid SOP name", err.Error(), "Example: customer-onboarding")
		return fmt.Errorf("invalid name")
	}

	sopFile := filepath.Join(ctx.RootDir, "sops", name+".md")
	if _, err := os.Stat(sopFile); !os.IsNotExist(err) {
		ui.ErrorBlock(
			"SOP file exists",
			fmt.Sprintf("File 'sops/%s.md' already exists", name),
			fmt.Sprintf("Choose a different name or remove:\n  rm sops/%s.md", name),
		)
		return fmt.Errorf("file exists")
	}

	var title, description string
	if !noWizard {
		title, _ = wizard.PromptString("Title", titleCase(name), nil)
		ui.PromptSuccess("Title", title)

		description, _ = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}
	} else {
		title = titleCase(name)
		description = ""
	}

	content := sopTemplate(title, description)
	if err := os.WriteFile(sopFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create SOP file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created sops/%s.md", name))

	return nil
}

// CreateTrigger creates a new trigger configuration file
func CreateTrigger(name string, noWizard bool) error {
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock("Not in formation directory", err.Error(), "")
		return fmt.Errorf("not in formation directory")
	}

	if err := validateComponentName(name); err != nil {
		ui.ErrorBlock("Invalid trigger name", err.Error(), "Example: webhook-handler")
		return fmt.Errorf("invalid name")
	}

	triggerFile := filepath.Join(ctx.RootDir, "triggers", name+".yaml")
	if _, err := os.Stat(triggerFile); !os.IsNotExist(err) {
		ui.ErrorBlock(
			"Trigger file exists",
			fmt.Sprintf("File 'triggers/%s.yaml' already exists", name),
			fmt.Sprintf("Choose a different name or remove:\n  rm triggers/%s.yaml", name),
		)
		return fmt.Errorf("file exists")
	}

	var description, triggerType string
	if !noWizard {
		description, _ = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}

		triggerType, _ = wizard.PromptString("Type", "webhook", nil)
		ui.PromptSuccess("Type", triggerType)
	} else {
		description = ""
		triggerType = "webhook"
	}

	content := triggerTemplate(name, description, triggerType)
	if err := os.WriteFile(triggerFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create trigger file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created triggers/%s.yaml", name))

	return nil
}

// CreateA2A creates a new A2A configuration file
func CreateA2A(name string, noWizard bool) error {
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock("Not in formation directory", err.Error(), "")
		return fmt.Errorf("not in formation directory")
	}

	if err := validateComponentName(name); err != nil {
		ui.ErrorBlock("Invalid A2A name", err.Error(), "Example: external-api")
		return fmt.Errorf("invalid name")
	}

	a2aFile := filepath.Join(ctx.RootDir, "a2a", name+".yaml")
	if _, err := os.Stat(a2aFile); !os.IsNotExist(err) {
		ui.ErrorBlock(
			"A2A file exists",
			fmt.Sprintf("File 'a2a/%s.yaml' already exists", name),
			fmt.Sprintf("Choose a different name or remove:\n  rm a2a/%s.yaml", name),
		)
		return fmt.Errorf("file exists")
	}

	var description, a2aType, baseURL string
	if !noWizard {
		description, _ = wizard.PromptString("Description (optional, press Enter to skip)", "", nil)
		if description != "" {
			ui.PromptSuccess("Description", description)
		} else {
			ui.PromptSkipped("Description")
		}

		a2aType, _ = wizard.PromptString("Type", "rest", nil)
		ui.PromptSuccess("Type", a2aType)

		baseURL, _ = wizard.PromptString("Base URL", "", nil)
		ui.PromptSuccess("Base URL", baseURL)
	} else {
		description = ""
		a2aType = "rest"
		baseURL = "https://api.example.com"
	}

	content := a2aTemplate(name, description, a2aType, baseURL)
	if err := os.WriteFile(a2aFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create A2A file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created a2a/%s.yaml", name))

	return nil
}

// Helper to convert kebab-case to Title Case
func titleCase(s string) string {
	words := strings.Split(s, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func mcpTemplate(name, description, mcpType, command string) string {
	if description == "" {
		description = fmt.Sprintf("%s MCP server", titleCase(name))
	}

	return fmt.Sprintf(`id: %s
description: "%s"
type: %s

connection:
  command: %s
  args: []
  env: {}

active: true
`, name, description, mcpType, command)
}

func sopTemplate(title, description string) string {
	date := "2025-11-26" // TODO: Use actual date
	
	content := fmt.Sprintf(`# %s

**Created:** %s  
**Status:** Draft

`, title, date)

	if description != "" {
		content += fmt.Sprintf(`## Overview

%s

`, description)
	} else {
		content += `## Overview

(Add overview here)

`
	}

	content += `## Prerequisites

- List prerequisites here

## Steps

1. First step
2. Second step
3. Third step

## Expected Outcomes

- Outcome 1
- Outcome 2

## Notes

- Additional notes
`
	return content
}

func triggerTemplate(name, description, triggerType string) string {
	if description == "" {
		description = fmt.Sprintf("%s trigger", titleCase(name))
	}

	return fmt.Sprintf(`id: %s
description: "%s"
type: %s

config:
  path: "/%s"
  method: POST

handler:
  agent: overlord

active: true
`, name, description, triggerType, name)
}

func a2aTemplate(name, description, a2aType, baseURL string) string {
	if description == "" {
		description = fmt.Sprintf("%s integration", titleCase(name))
	}

	return fmt.Sprintf(`id: %s
description: "%s"
type: %s

connection:
  base_url: "%s"
  auth:
    type: bearer
    token: "${{ secrets.%s_TOKEN }}"

endpoints: []
active: true
`, name, description, a2aType, baseURL, strings.ToUpper(strings.ReplaceAll(name, "-", "_")))
}
