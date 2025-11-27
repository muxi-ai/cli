package scaffold

import (
	"fmt"
	"net/url"
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
	// Must be in formation directory - check FIRST
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"Run this command from inside a formation directory:\n  cd my-formation\n  muxi new agent weather\n\nOr create a new formation:\n  muxi new formation",
			"",
		)
		os.Exit(1)
	}

	// If no name provided, handle based on mode
	if name == "" {
		if noWizard {
			return fmt.Errorf("agent ID required (provide as argument or run without --no-wizard)")
		}
		
		// Interactive mode - prompt for ID with validation loop
		for {
			var err error
			inputName, err := wizard.PromptString("Agent ID", "", nil)
			if err != nil {
				return err
			}
			
			// Normalize the input
			name = normalizeComponentName(inputName)
			
			// Validate normalized name
			if err := validateComponentName(name); err != nil {
				ui.PromptError("Agent ID", inputName, err)
				continue
			}
			
			// Check if file already exists
			agentFile := filepath.Join(ctx.RootDir, "agents", name+".yaml")
			if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
				// Show error and re-prompt
				ui.PromptError("Agent ID", inputName, fmt.Errorf("file already exists\n\nChoose a different ID or remove:\n  rm agents/%s.yaml", name))
				continue
			}
			
			// All good - show success with normalized name
			ui.PromptSuccess("Agent ID", name)
			break
		}
	} else {
		// Name provided as argument - normalize it
		name = normalizeComponentName(name)
		
		// Validate normalized name
		if err := validateComponentName(name); err != nil {
			ui.ErrorBlock("Invalid agent ID", err.Error(), "Example: weather-assistant")
			os.Exit(1)
		}
		
		// Check if file already exists
		agentFile := filepath.Join(ctx.RootDir, "agents", name+".yaml")
		if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
			ui.ErrorBlock(
				"Agent file exists",
				fmt.Sprintf("File 'agents/%s.yaml' already exists", name),
				fmt.Sprintf("Choose a different name or remove:\n  rm agents/%s.yaml", name),
			)
			os.Exit(1)
		}
		
		// If interactive mode, show ID success
		if !noWizard {
			ui.PromptSuccess("Agent ID", name)
		}
	}

	// Interactive mode - full wizard
	var agentName, systemMessage, role string
	var specialties []string

	if !noWizard {
		// Agent name (inferred from ID)
		inferredName := titleCase(name)
		agentName, _ = wizard.PromptString("Name", inferredName, nil)
		ui.PromptSuccess("Name", agentName)

		// System message (optional)
		systemMessage, _ = wizard.PromptString("System message (Enter to skip, you can add it later)", "", nil)
		if systemMessage != "" {
			ui.PromptSuccess("System message", "configured")
		} else {
			ui.PromptSkipped("System message")
		}

		// Role selection (interactive)
		roleOptions := []wizard.SelectOption{
			{Value: "generalist", Label: "Generalist", Description: "General-purpose assistant"},
			{Value: "specialist", Label: "Specialist", Description: "Domain expert with specific skills"},
			{Value: "assistant", Label: "Assistant", Description: "Supports other agents"},
			{Value: "custom", Label: "Custom", Description: "Specify your own role"},
		}
		
		fmt.Println()
		selectedRole, err := wizard.PromptSelect("Role", roleOptions, 0)
		if err != nil {
			return fmt.Errorf("failed to select role: %w", err)
		}
		
		// If custom, prompt for custom role name
		if selectedRole == "custom" {
			customRole, _ := wizard.PromptString("Custom role name", "", nil)
			role = customRole
			ui.PromptSuccess("Role", customRole)
		} else {
			role = selectedRole
			// Find the label for display
			for _, opt := range roleOptions {
				if opt.Value == selectedRole {
					ui.PromptSuccess("Role", opt.Label)
					break
				}
			}
		}

		// Specialties (comma-separated, optional)
		specialtiesStr, _ := wizard.PromptString("Specialties (comma-separated, optional)", "", nil)
		if specialtiesStr != "" {
			// Split and trim
			parts := strings.Split(specialtiesStr, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					specialties = append(specialties, trimmed)
				}
			}
			ui.PromptSuccess("Specialties", strings.Join(specialties, ", "))
		} else {
			ui.PromptSkipped("Specialties")
		}
	} else {
		// Non-interactive defaults
		agentName = titleCase(name)
		systemMessage = ""
		role = "generalist"
		specialties = []string{}
	}

	// Create agent file
	agentFile := filepath.Join(ctx.RootDir, "agents", name+".yaml")
	content := agentTemplate(name, agentName, systemMessage, role, specialties)
	if err := os.WriteFile(agentFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create agent file: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Created agents/%s.yaml", name))

	if !noWizard {
		fmt.Println()
		ui.Dimmed("Next steps:")
		ui.Dimmed("  • Add knowledge: Edit the 'knowledge' section")
		ui.Dimmed("  • Add MCP servers: Edit the 'mcp_servers' section")
	}

	return nil
}

// normalizeComponentName converts user input to valid component name format
// Converts spaces to hyphens, lowercases, removes extra hyphens
func normalizeComponentName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	
	// Replace multiple hyphens with single hyphen
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	
	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")
	
	return name
}

// validateComponentName validates component name format
func validateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Pattern: lowercase letter, then lowercase letters/numbers/hyphens/underscores, 3-50 chars
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-_]{2,49}$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("component names must:\n  • Start with a letter\n  • Contain only letters, numbers, hyphens, and underscores\n  • Be 3-50 characters long")
	}

	return nil
}

func agentTemplate(id, name, systemMessage, role string, specialties []string) string {
	description := fmt.Sprintf("%s agent", name)
	
	// System message
	systemMsg := systemMessage
	if systemMsg == "" {
		systemMsg = "You are a helpful AI assistant."
	}
	
	// Specialties
	specialtiesYAML := ""
	if len(specialties) > 0 {
		specialtiesYAML = "specialties:\n"
		for _, s := range specialties {
			specialtiesYAML += fmt.Sprintf("  - %s\n", s)
		}
	} else {
		specialtiesYAML = "specialties: []\n"
	}
	
	return fmt.Sprintf(`schema: "1.0.0"

id: %s
name: %s
description: "%s"
active: true

role: %s
%s
system_message: |
  %s

# Add knowledge sources by editing below:
# knowledge:
#   - source: knowledge/weather-data.md
#     description: "Historical weather patterns and climate trends"
#     enabled: true
knowledge: []

# Add MCP servers by editing below (agent-specific MCP servers):
# mcp_servers:
#   - id: "weather_service"
#     description: "External weather service"
#     active: true
#     type: "http"
#     endpoint: "http://localhost:3000"
#     auth:
#       type: "api_key"
#       header: "X-API-Key"
#       key: "${{ secrets.WEATHER_API_KEY }}"
mcp_servers: []
`, id, name, description, role, specialtiesYAML, systemMsg)
}

// CreateMCP creates a new MCP server configuration file
func CreateMCP(name, agentID string, noWizard bool) error {
	// Must be in formation directory - check FIRST
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"Run this command from inside a formation directory:\n  cd my-formation\n  muxi new mcp weather-api\n\nOr create a new formation:\n  muxi new formation",
			"",
		)
		os.Exit(1)
	}

	// Show banner (formation-level or agent-specific)
	if agentID != "" {
		// Agent-specific banner
		agentFile := filepath.Join(ctx.RootDir, "agents", agentID+".yaml")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			ui.ErrorBlock(
				"Agent not found",
				fmt.Sprintf("Agent '%s' does not exist", agentID),
				"Create the agent first:\n  muxi new agent "+agentID,
			)
			os.Exit(1)
		}
		
		// Load agent to get name
		agentName := titleCase(agentID) // Fallback to ID
		ui.InfoBanner(fmt.Sprintf("[i] MCP for: %s", agentName))
	} else {
		// Formation-level banner
		ui.InfoBanner("[i] Formation-level MCPs can be used by all agents.\n\nFor tools that are going to be used primarily by a specific\nagent, we recommend adding the MCP on the agent-level:\n  $ muxi new mcp --agent <agent-id>")
	}

	// If no name provided, handle based on mode
	if name == "" {
		if noWizard {
			return fmt.Errorf("MCP ID required (provide as argument or run without --no-wizard)")
		}
		
		// Interactive mode - prompt for ID
		for {
			var err error
			inputName, err := wizard.PromptString("MCP ID", "", nil)
			if err != nil {
				return err
			}
			
			// Normalize the input
			name = normalizeComponentName(inputName)
			
			// Validate normalized name
			if err := validateComponentName(name); err != nil {
				ui.PromptError("MCP ID", inputName, err)
				continue
			}
			
			// Check if MCP already exists (different check for agent-level vs formation-level)
			if agentID != "" {
				// Agent-level: check if MCP ID already exists in agent's YAML
				if mcpExistsInAgent(ctx.RootDir, agentID, name) {
					ui.PromptError("MCP ID", inputName, fmt.Errorf("MCP with this ID already exists in the formation\n\nChoose a different ID or edit:\n  agents/%s.yaml", agentID))
					continue
				}
			} else {
				// Formation-level: check if file exists
				mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
				if _, err := os.Stat(mcpFile); !os.IsNotExist(err) {
					ui.PromptError("MCP ID", inputName, fmt.Errorf("file already exists\n\nChoose a different ID or remove:\n  rm mcps/%s.yaml", name))
					continue
				}
			}
			
			ui.PromptSuccess("MCP ID", name)
			break
		}
	} else {
		// Name provided as argument - normalize it
		name = normalizeComponentName(name)
		
		// Validate normalized name
		if err := validateComponentName(name); err != nil {
			ui.ErrorBlock("Invalid MCP ID", err.Error(), "Example: weather-api")
			os.Exit(1)
		}
		
		// Check if MCP already exists (different check for agent-level vs formation-level)
		if agentID != "" {
			// Agent-level: check if MCP ID already exists in agent's YAML
			if mcpExistsInAgent(ctx.RootDir, agentID, name) {
				ui.ErrorBlock(
					"MCP already exists",
					fmt.Sprintf("MCP with ID '%s' already exists in the formation", name),
					fmt.Sprintf("Choose a different ID or edit:\n  agents/%s.yaml", agentID),
				)
				os.Exit(1)
			}
		} else {
			// Formation-level: check if file exists
			mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
			if _, err := os.Stat(mcpFile); !os.IsNotExist(err) {
				ui.ErrorBlock(
					"MCP file exists",
					fmt.Sprintf("File 'mcps/%s.yaml' already exists", name),
					fmt.Sprintf("Choose a different name or remove:\n  rm mcps/%s.yaml", name),
				)
				os.Exit(1)
			}
		}
		
		if !noWizard {
			ui.PromptSuccess("MCP ID", name)
		}
	}

	// Interactive wizard
	var description, transport, endpoint, command, args, workingDir, installCmd string
	var authType, authHeader string
	var envVars []string
	var secrets []string // Secrets to add to secrets file

	if !noWizard {
		// Description
		description, _ = wizard.PromptString("Description", "", nil)
		ui.PromptSuccess("Description", description)

		// Transport type selection
		transportOptions := []wizard.SelectOption{
			{Value: "http", Label: "HTTP", Description: "Streamable HTTP server"},
			{Value: "stdio", Label: "Stdio", Description: "Local command-line tool"},
		}
		
		fmt.Println()
		transport, err = wizard.PromptSelect("Transport", transportOptions, 0)
		if err != nil {
			return fmt.Errorf("failed to select transport: %w", err)
		}
		
		// Find the label for display
		for _, opt := range transportOptions {
			if opt.Value == transport {
				ui.PromptSuccess("Transport", opt.Label)
				break
			}
		}

		if transport == "http" {
			// HTTP-specific prompts
			// Endpoint URL with validation
			for {
				endpoint, _ = wizard.PromptString("Endpoint URL", "", nil)
				
				// Validate URL structure
				if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("must start with http:// or https://"))
					continue
				}
				
				// Parse and validate URL
				parsedURL, err := url.Parse(endpoint)
				if err != nil || parsedURL.Host == "" {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("invalid URL format"))
					continue
				}
				
				ui.PromptSuccess("Endpoint URL", endpoint)
				break
			}

			// Auth selection
			authOptions := []wizard.SelectOption{
				{Value: "none", Label: "None", Description: "No authentication"},
				{Value: "api_key", Label: "API Key", Description: "API key in header"},
				{Value: "bearer", Label: "Bearer Token", Description: "Bearer token authentication"},
				{Value: "basic", Label: "Basic Auth", Description: "Username and password"},
			}
			
			authType, err = wizard.PromptSelect("Authentication", authOptions, 0)
			if err != nil {
				return fmt.Errorf("failed to select auth: %w", err)
			}
			
			// Find the label for display
			for _, opt := range authOptions {
				if opt.Value == authType {
					ui.PromptSuccess("Authentication", opt.Label)
					break
				}
			}

			// Auth-specific prompts (only for non-"none" auth types)
			if authType != "none" {
				fmt.Println() // Line break before auth prompts (not for "none")
				
				secretPrefix := generateMCPSecretPrefix(name)
				
				switch authType {
				case "api_key":
					authHeader, _ = wizard.PromptString("API Key header", "X-API-Key", nil)
					ui.PromptSuccess("API Key", authHeader)
					secrets = append(secrets, secretPrefix+"_API_KEY")
					
				case "bearer":
					// Prompt for bearer token value
					bearerToken, _ := wizard.PromptString("Bearer Token", "", nil)
					ui.PromptSuccess("Bearer Token", bearerToken)
					secrets = append(secrets, secretPrefix+"_BEARER_TOKEN")
					
				case "basic":
					// Prompt for username and password separately
					username, _ := wizard.PromptString("Username", "", nil)
					ui.PromptSuccess("Username", username)
					
					password, _ := wizard.PromptString("Password", "", nil)
					ui.PromptSuccess("Password", password)
					
					secrets = append(secrets, secretPrefix+"_USERNAME", secretPrefix+"_PASSWORD")
				}
			}

		} else {
			// Stdio-specific prompts
			command, _ = wizard.PromptString("Command", "", nil)
			ui.PromptSuccess("Command", command)

			args, _ = wizard.PromptString("Arguments (space-separated, or Enter to skip)", "", nil)
			if args != "" {
				ui.PromptSuccess("Arguments", args)
			} else {
				ui.PromptSkipped("Arguments")
			}

			// Working directory (optional)
			workingDir, _ = wizard.PromptString("Working directory (Enter to skip)", "", nil)
			workingDir = strings.TrimSpace(workingDir) // Ensure no whitespace issues
			if workingDir != "" {
				ui.PromptSuccess("Working directory", workingDir)
			} else {
				ui.PromptSkipped("Working directory")
			}

			// Environment variables
			envInput, _ := wizard.PromptString("Environment variables (comma/space/newline separated, or Enter to skip)", "", nil)
			if envInput != "" {
				envVars = parseEnvironmentVariables(envInput)
				ui.PromptSuccess("Environment", strings.Join(envVars, ", "))
				
				// Add each env var as a secret
				secretPrefix := generateMCPSecretPrefix(name)
				for _, envVar := range envVars {
					secrets = append(secrets, secretPrefix+"_"+envVar)
				}
			} else {
				ui.PromptSkipped("Environment")
			}

			// Linebreak before install command
			fmt.Println()

			// Install command (optional) - at the end
			installCmd, _ = wizard.PromptString("Auto-install command (optional)", "", nil)
			if installCmd != "" {
				ui.PromptSuccess("Install", installCmd)
			} else {
				ui.PromptSkipped("Install")
			}
		}
	} else {
		// Non-interactive defaults
		description = fmt.Sprintf("%s MCP server", titleCase(name))
		transport = "stdio"
		command = "mcp-server"
		args = ""
		workingDir = ""
		installCmd = ""
		authType = "none"
		envVars = []string{}
	}

	// Handle agent-level vs formation-level MCP
	if agentID != "" {
		// Agent-level MCP - append to agent's YAML file
		err := appendMCPToAgent(ctx.RootDir, agentID, name, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader, envVars)
		if err != nil {
			return fmt.Errorf("failed to add MCP to agent: %w", err)
		}
	} else {
		// Formation-level MCP - create separate file
		content := mcpTemplateNew(name, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader, envVars)
		
		// Write MCP file
		mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
		if err := os.WriteFile(mcpFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create MCP file: %w", err)
		}
	}

	// Append secrets to secrets file
	if len(secrets) > 0 {
		secretsFile := filepath.Join(ctx.RootDir, "secrets")
		secretsContent := "\n"
		for _, secret := range secrets {
			secretsContent += secret + "=\n"
		}
		
		f, err := os.OpenFile(secretsFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(secretsContent)
			f.Close()
		}
	}

	fmt.Println()
	if agentID != "" {
		ui.Success(fmt.Sprintf("Added MCP '%s' to agents/%s.yaml", name, agentID))
	} else {
		ui.Success(fmt.Sprintf("Created mcps/%s.yaml", name))
	}
	
	if len(secrets) > 0 {
		secretsList := strings.Join(secrets, ", ")
		ui.Success(fmt.Sprintf("Added %d secret(s) to configure: %s", len(secrets), secretsList))
	}

	if !noWizard {
		fmt.Println()
		ui.Dimmed("Next steps:")
		ui.Dimmed("  • Configure secrets: muxi secrets setup")
		if agentID != "" {
			ui.Dimmed(fmt.Sprintf("  • Edit MCP: agents/%s.yaml (under mcp_servers)", agentID))
		} else {
			ui.Dimmed(fmt.Sprintf("  • Edit MCP: mcps/%s.yaml", name))
		}
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

// generateMCPSecretPrefix generates a secret prefix from MCP ID
// Strips common words: mcp, server, api, tool(s), key
// Example: weather-api -> MCP_WEATHER
// Example: postgres-mcp-server -> MCP_POSTGRES
func generateMCPSecretPrefix(mcpID string) string {
	// Remove common suffixes/prefixes
	stripped := mcpID
	commonWords := []string{"-mcp", "mcp-", "-server", "server-", "-api", "api-", "-tools", "-tool", "tool-", "tools-", "-key", "key-"}
	
	for _, word := range commonWords {
		stripped = strings.ReplaceAll(stripped, word, "")
	}
	
	// Clean up multiple hyphens or leading/trailing hyphens
	stripped = strings.Trim(stripped, "-")
	for strings.Contains(stripped, "--") {
		stripped = strings.ReplaceAll(stripped, "--", "-")
	}
	
	// If nothing left after stripping, use original ID
	if stripped == "" || stripped == "-" {
		stripped = mcpID
	}
	
	// Convert to uppercase and replace - with _
	stripped = strings.ToUpper(strings.ReplaceAll(stripped, "-", "_"))
	
	return "MCP_" + stripped
}

// parseEnvironmentVariables parses comma/space/newline separated environment variable names
// Accepts: "VAR1, VAR2, VAR3" or "VAR1 VAR2 VAR3" or "VAR1,\nVAR2,\nVAR3" or mixed
// Converts all variable names to uppercase
func parseEnvironmentVariables(input string) []string {
	if input == "" {
		return []string{}
	}
	
	// Replace newlines and commas with spaces
	normalized := strings.ReplaceAll(input, "\n", " ")
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, "\\", " ")
	
	// Split by whitespace and filter empties
	parts := strings.Fields(normalized)
	
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			// Convert to uppercase (env vars are conventionally uppercase)
			result = append(result, strings.ToUpper(trimmed))
		}
	}
	
	return result
}

// mcpExistsInAgent checks if an MCP with the given ID already exists in the agent's YAML
func mcpExistsInAgent(rootDir, agentID, mcpID string) bool {
	agentFile := filepath.Join(rootDir, "agents", agentID+".yaml")
	content, err := os.ReadFile(agentFile)
	if err != nil {
		return false
	}
	
	// Simple check: look for 'id: "mcpID"' or 'id: mcpID' in the mcp_servers section
	contentStr := string(content)
	
	// Check if we're in the mcp_servers section and find the ID
	lines := strings.Split(contentStr, "\n")
	inMCPServers := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Track if we're in mcp_servers section
		if strings.HasPrefix(trimmed, "mcp_servers:") {
			inMCPServers = true
			continue
		}
		
		// If we hit another top-level key, we're out of mcp_servers
		if inMCPServers && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "#") {
			inMCPServers = false
		}
		
		// Check for ID match in mcp_servers section
		if inMCPServers && (strings.HasPrefix(trimmed, "id:") || strings.HasPrefix(trimmed, "- id:")) {
			// Extract the ID value
			idPart := strings.TrimPrefix(trimmed, "- ")
			idPart = strings.TrimPrefix(idPart, "id:")
			idPart = strings.TrimSpace(idPart)
			idPart = strings.Trim(idPart, "\"'")
			
			if idPart == mcpID {
				return true
			}
		}
	}
	
	return false
}

// appendMCPToAgent adds an MCP server configuration to an agent's YAML file
func appendMCPToAgent(rootDir, agentID, mcpID, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader string, envVars []string) error {
	agentFile := filepath.Join(rootDir, "agents", agentID+".yaml")
	
	// Read existing agent YAML
	content, err := os.ReadFile(agentFile)
	if err != nil {
		return fmt.Errorf("failed to read agent file: %w", err)
	}
	
	// Generate MCP entry for agent (indented YAML)
	mcpEntry := generateAgentMCPEntry(mcpID, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader, envVars)
	
	// Replace "mcp_servers: []" with the full array
	contentStr := string(content)
	
	// Check if mcp_servers is empty array
	if strings.Contains(contentStr, "mcp_servers: []") {
		// Replace empty array with full array
		replacement := "mcp_servers:\n" + mcpEntry
		contentStr = strings.Replace(contentStr, "mcp_servers: []", replacement, 1)
	} else if strings.Contains(contentStr, "mcp_servers:") {
		// Array already exists - append new entry
		// Find the mcp_servers: line and add after it
		lines := strings.Split(contentStr, "\n")
		var newLines []string
		foundMCPServers := false
		
		for i, line := range lines {
			newLines = append(newLines, line)
			if !foundMCPServers && strings.HasPrefix(strings.TrimSpace(line), "mcp_servers:") {
				foundMCPServers = true
				// Insert new MCP entry after this line
				// But first, we need to find where to insert (after existing entries)
				// For now, insert right after the mcp_servers: line
				// This is a simplified approach - ideally we'd find the end of the array
				
				// Find the end of current mcp_servers array
				j := i + 1
				for j < len(lines) {
					trimmed := strings.TrimSpace(lines[j])
					// Check if this line is still part of mcp_servers array
					if trimmed == "" || (!strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "  ") && !strings.HasPrefix(trimmed, "#")) {
						// Found end of array
						break
					}
					if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(lines[j], "  ") {
						// Found next top-level field
						break
					}
					newLines = append(newLines, lines[j])
					j++
				}
				
				// Insert new MCP entry
				newLines = append(newLines, mcpEntry)
				
				// Skip the lines we already added
				for k := i + 1; k < j; k++ {
					lines[k] = "" // Mark as processed
				}
			}
		}
		
		// Filter out empty processed lines and rebuild
		var finalLines []string
		for _, line := range newLines {
			if line != "" || !foundMCPServers {
				finalLines = append(finalLines, line)
			}
		}
		
		contentStr = strings.Join(finalLines, "\n")
	} else {
		// mcp_servers field doesn't exist - add it at the end
		contentStr += "\n\nmcp_servers:\n" + mcpEntry
	}
	
	// Write back to file
	if err := os.WriteFile(agentFile, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}
	
	return nil
}

// generateAgentMCPEntry creates a properly formatted MCP entry for agent YAML (indented)
func generateAgentMCPEntry(mcpID, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader string, envVars []string) string {
	var entry strings.Builder
	
	// Start with the list item
	entry.WriteString("  - id: \"" + mcpID + "\"\n")
	entry.WriteString("    description: \"" + description + "\"\n")
	entry.WriteString("    active: true\n")
	entry.WriteString("    type: \"" + transport + "\"\n")
	
	if transport == "http" {
		// HTTP transport
		entry.WriteString("    endpoint: \"" + endpoint + "\"\n")
		
		// Auth
		if authType != "none" {
			entry.WriteString("    auth:\n")
			entry.WriteString("      type: \"" + authType + "\"\n")
			
			secretPrefix := generateMCPSecretPrefix(mcpID)
			
			switch authType {
			case "api_key":
				entry.WriteString("      header: \"" + authHeader + "\"\n")
				entry.WriteString("      key: \"${{ secrets." + secretPrefix + "_API_KEY }}\"\n")
			case "bearer":
				entry.WriteString("      token: \"${{ secrets." + secretPrefix + "_BEARER_TOKEN }}\"\n")
			case "basic":
				entry.WriteString("      username: \"${{ secrets." + secretPrefix + "_USERNAME }}\"\n")
				entry.WriteString("      password: \"${{ secrets." + secretPrefix + "_PASSWORD }}\"\n")
			}
		}
	} else {
		// Stdio transport
		if installCmd != "" {
			entry.WriteString("    install: \"" + installCmd + "\"\n")
		}
		
		entry.WriteString("    command: \"" + command + "\"\n")
		
		if args != "" {
			argsList := strings.Fields(args)
			if len(argsList) > 0 {
				entry.WriteString("    args: [")
				for i, arg := range argsList {
					if i > 0 {
						entry.WriteString(", ")
					}
					entry.WriteString("\"" + arg + "\"")
				}
				entry.WriteString("]\n")
			}
		}
		
		if workingDir != "" {
			entry.WriteString("    working_directory: \"" + workingDir + "\"\n")
		}
		
		// Env vars
		if len(envVars) > 0 {
			secretPrefix := generateMCPSecretPrefix(mcpID)
			entry.WriteString("    auth:\n")
			entry.WriteString("      type: \"env\"\n")
			for _, envVar := range envVars {
				entry.WriteString("      " + envVar + ": \"${{ secrets." + secretPrefix + "_" + envVar + " }}\"\n")
			}
		}
	}
	
	return entry.String()
}

func mcpTemplateNew(id, description, transport, endpoint, command, args, workingDir, installCmd, authType, authHeader string, envVars []string) string {
	if description == "" {
		description = fmt.Sprintf("%s MCP server", titleCase(id))
	}

	secretPrefix := generateMCPSecretPrefix(id)
	
	// Build the template
	var tmpl strings.Builder
	
	// Header
	tmpl.WriteString(fmt.Sprintf(`schema: "1.0.0"

id: %s
description: "%s"
active: true

type: %s
`, id, description, transport))

	if transport == "http" {
		// HTTP transport
		tmpl.WriteString(fmt.Sprintf(`endpoint: "%s"
`, endpoint))

		// Optional: retry/timeout comments
		tmpl.WriteString(`
# Optional: Override default retry/timeout settings
# retry_attempts: 3
# timeout_seconds: 30
`)

		// Auth section
		if authType != "none" {
			tmpl.WriteString("\nauth:\n")
			
			switch authType {
			case "api_key":
				tmpl.WriteString(fmt.Sprintf(`  type: api_key
  header: "%s"
  key: "${{ secrets.%s_API_KEY }}"
`, authHeader, secretPrefix))
				
			case "bearer":
				tmpl.WriteString(fmt.Sprintf(`  type: bearer
  token: "${{ secrets.%s_BEARER_TOKEN }}"
`, secretPrefix))
				
			case "basic":
				tmpl.WriteString(fmt.Sprintf(`  type: basic
  username: "${{ secrets.%s_USERNAME }}"
  password: "${{ secrets.%s_PASSWORD }}"
`, secretPrefix, secretPrefix))
			}
		}

	} else {
		// Stdio transport
		// Install command (optional)
		if installCmd != "" {
			tmpl.WriteString(fmt.Sprintf(`install: "%s"

`, installCmd))
		}

		tmpl.WriteString(fmt.Sprintf(`command: "%s"
`, command))

		// Args (inline array format)
		if args != "" {
			// Parse args into inline array
			argsList := strings.Fields(args)
			if len(argsList) > 0 {
				tmpl.WriteString("args: [")
				for i, arg := range argsList {
					if i > 0 {
						tmpl.WriteString(", ")
					}
					tmpl.WriteString(fmt.Sprintf("\"%s\"", arg))
				}
				tmpl.WriteString("]\n")
			}
		} else {
			tmpl.WriteString("args: []\n")
		}

		// Working directory
		if workingDir != "" {
			tmpl.WriteString(fmt.Sprintf("working_directory: \"%s\"\n", workingDir))
		}

		// Optional: retry/timeout comments
		tmpl.WriteString(`
# Optional: Override default retry/timeout settings
# max_retries: 3
# timeout_seconds: 30
`)

		// Env vars via auth.type="env" (as per spec)
		if len(envVars) > 0 {
			tmpl.WriteString("\nauth:\n")
			tmpl.WriteString("  type: \"env\"\n")
			for _, envVar := range envVars {
				tmpl.WriteString(fmt.Sprintf("  %s: \"${{ secrets.%s_%s }}\"\n", envVar, secretPrefix, envVar))
			}
		} else {
			// Show example
			tmpl.WriteString(`
# Optional: Environment variables (passed to command)
# auth:
#   type: "env"
#   NODE_ENV: "${{ secrets.MCP_<ID>_NODE_ENV }}"
#   API_KEY: "${{ secrets.MCP_<ID>_API_KEY }}"
`)
		}

		// Optional: install comment (if not already set)
		if installCmd == "" {
			tmpl.WriteString(`
# Optional: Auto-install command (server will run this before starting)
# install: "npm install -g @modelcontextprotocol/server-json-rpc"
`)
		}

		// Optional: working directory comment (if not already set)
		if workingDir == "" {
			tmpl.WriteString(`
# Optional: Working directory
# working_directory: "/path/to/working/dir"
`)
		}
	}

	return tmpl.String()
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

	return fmt.Sprintf(`schema: "1.0.0"

id: %s
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

	return fmt.Sprintf(`schema: "1.0.0"

id: %s
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
