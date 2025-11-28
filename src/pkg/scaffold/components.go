package scaffold

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/muxi-ai/cli/pkg/context"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/muxi-ai/cli/pkg/wizard"
	"gopkg.in/yaml.v3"
)

// CreateAgent creates a new agent configuration file
func CreateAgent(name string, noWizard bool) error {
	// Must be in formation directory - check FIRST
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show banner in interactive mode
	if !noWizard {
		// Check if A2A is enabled
		if isA2AEnabled(ctx.RootDir) {
			ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Adding new agent                                    MUXI │\n│──────────────────────────────────────────────────────────────│\n│ ℹ This formation has A2A enabled. You can make this agent    │\n│ visible externally for Agent-to-Agent communication.         │\n╰──────────────────────────────────────────────────────────────╯")
		} else {
			ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Adding new agent                                    MUXI │\n╰──────────────────────────────────────────────────────────────╯")
		}
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

	// Check if formation has A2A enabled
	a2aEnabled := isA2AEnabled(ctx.RootDir)
	externalA2A := false

	if !noWizard && a2aEnabled {
		// Ask about external A2A visibility
		fmt.Println()
		externalA2A, _ = wizard.PromptConfirm("Make this agent visible externally (via A2A)", false)
		if externalA2A {
			ui.PromptSuccess("A2A", "External visibility enabled")
		} else {
			ui.PromptSuccess("A2A", "Internal only")
		}
	}

	// Create agent file
	agentFile := filepath.Join(ctx.RootDir, "agents", name+".yaml")
	content := agentTemplate(name, agentName, systemMessage, role, specialties, externalA2A)
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

// isA2AEnabled checks if the formation has A2A enabled
func isA2AEnabled(rootDir string) bool {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return false
	}

	// Simple check: look for "a2a:" section with "enabled: true"
	lines := strings.Split(string(content), "\n")
	inA2ASection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're in the a2a section
		if trimmed == "a2a:" {
			inA2ASection = true
			continue
		}

		// If we hit another top-level key, we're out of a2a section
		if inA2ASection && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(trimmed, "#") {
			inA2ASection = false
		}

		// Check for enabled: true in a2a section
		if inA2ASection && strings.HasPrefix(trimmed, "enabled:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:"))
			return value == "true"
		}
	}

	return false
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

func agentTemplate(id, name, systemMessage, role string, specialties []string, externalA2A bool) string {
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

# Agent-to-Agent communication configuration
a2a:
  internal: true   # Participates in formation A2A
  external: %t  # Participates in external A2A
`, id, name, description, role, specialtiesYAML, systemMsg, externalA2A)
}

// CreateMCP creates a new MCP server configuration file
func CreateMCP(name, agentID string, noWizard bool) error {
	// Must be in formation directory - check FIRST
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show banner in interactive mode (formation-level or agent-specific)
	if !noWizard {
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

			// Load agent to get name (use titleCase of ID for now)
			agentName := titleCase(agentID)
			// Build banner with MUXI on right - total content width is 62 chars
			header := fmt.Sprintf("[+] Adding new MCP to: %s", agentName)
			padding := 57 - len(header) // 62 - 5 (for " MUXI") = 57
			if padding < 0 {
				padding = 0
			}
			bannerLine := fmt.Sprintf("│ %s%s MUXI │", header, strings.Repeat(" ", padding))
			ui.Banner(fmt.Sprintf("╭──────────────────────────────────────────────────────────────╮\n%s\n╰──────────────────────────────────────────────────────────────╯", bannerLine))
		} else {
			// Formation-level banner with red warning
			ui.FormationMCPBanner()
		}
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

			// Check if MCP already exists in formation (formation-level) OR agent (agent-level)
			mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
			formationLevelExists := false
			if _, err := os.Stat(mcpFile); !os.IsNotExist(err) {
				formationLevelExists = true
			}

			agentLevelExists := false
			if agentID != "" {
				agentLevelExists = mcpExistsInAgent(ctx.RootDir, agentID, name)
			}

			if formationLevelExists || agentLevelExists {
				if agentID != "" {
					// Agent-level MCP - clearer message
					ui.PromptError("MCP ID", inputName, fmt.Errorf("MCP with this ID already exists in the formation\n\nChoose a different ID or edit:\n  agents/%s.yaml", agentID))
				} else {
					// Formation-level MCP
					ui.PromptError("MCP ID", inputName, fmt.Errorf("file already exists\n\nChoose a different ID or remove:\n  rm mcps/%s.yaml", name))
				}
				continue
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

		// Check if MCP already exists in formation (formation-level) OR agent (agent-level)
		mcpFile := filepath.Join(ctx.RootDir, "mcps", name+".yaml")
		formationLevelExists := false
		if _, err := os.Stat(mcpFile); !os.IsNotExist(err) {
			formationLevelExists = true
		}

		agentLevelExists := false
		if agentID != "" {
			agentLevelExists = mcpExistsInAgent(ctx.RootDir, agentID, name)
		}

		if formationLevelExists || agentLevelExists {
			if agentID != "" {
				// Agent-level MCP - clearer message
				ui.ErrorBlock(
					"MCP already exists",
					fmt.Sprintf("MCP '%s' already exists in this formation.", name),
					fmt.Sprintf("Choose a different ID or edit:\n  agents/%s.yaml", agentID),
				)
			} else {
				// Formation-level MCP
				ui.ErrorBlock(
					"MCP file exists",
					fmt.Sprintf("File 'mcps/%s.yaml' already exists", name),
					fmt.Sprintf("Choose a different name or remove:\n  rm mcps/%s.yaml", name),
				)
			}
			os.Exit(1)
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

	// TODO(secrets): When secrets management is implemented, use these collected values:
	// - Encrypt and store in secrets.enc
	// - Add to secrets file as placeholders (SECRET_NAME=)
	// - Provide `muxi secrets set SECRET_NAME value` command for updating
	var secretValues = make(map[string]string) // Store actual secret values (unused until secrets management exists)

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

				// Auto-add https:// if no protocol specified
				if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
					endpoint = "https://" + endpoint
				}

				// Parse and validate URL
				parsedURL, err := url.Parse(endpoint)
				if err != nil || parsedURL.Host == "" {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("invalid URL format"))
					continue
				}

				// Additional host validation
				host := parsedURL.Hostname()
				if host == "" {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("invalid hostname"))
					continue
				}

				// Check for trailing dots
				if strings.HasSuffix(host, ".") {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("hostname cannot end with a dot"))
					continue
				}

				// Check for consecutive dots
				if strings.Contains(host, "..") {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("hostname cannot contain consecutive dots"))
					continue
				}

				// Check for empty parts (e.g., "http://example..com")
				if strings.HasPrefix(host, ".") {
					ui.PromptError("Endpoint URL", endpoint, fmt.Errorf("hostname cannot start with a dot"))
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
				secretPrefix := generateMCPSecretPrefix(name)

				switch authType {
				case "api_key":
					authHeader, _ = wizard.PromptString("API Key header name", "X-API-Key", nil)
					ui.PromptSuccess("Header name", authHeader)

					apiKey, _ := wizard.PromptString("API Key value", "", nil)
					if len(apiKey) > 4 {
						ui.PromptSuccess("API Key", "***"+apiKey[len(apiKey)-4:])
					} else {
						ui.PromptSuccess("API Key", "***")
					}

					secretName := secretPrefix + "_API_KEY"
					secrets = append(secrets, secretName)
					secretValues[secretName] = apiKey

				case "bearer":
					bearerToken, _ := wizard.PromptString("Bearer token value", "", nil)
					if len(bearerToken) > 4 {
						ui.PromptSuccess("Bearer token", "***"+bearerToken[len(bearerToken)-4:])
					} else {
						ui.PromptSuccess("Bearer token", "***")
					}

					secretName := secretPrefix + "_BEARER_TOKEN"
					secrets = append(secrets, secretName)
					secretValues[secretName] = bearerToken

				case "basic":
					username, _ := wizard.PromptString("Username", "", nil)
					ui.PromptSuccess("Username", username)

					password, _ := wizard.PromptString("Password", "", nil)
					ui.PromptSuccess("Password", "***")

					usernameSecret := secretPrefix + "_USERNAME"
					passwordSecret := secretPrefix + "_PASSWORD"
					secrets = append(secrets, usernameSecret, passwordSecret)
					secretValues[usernameSecret] = username
					secretValues[passwordSecret] = password
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

	// Append secrets to secrets file (placeholders only)
	// TODO(secrets): When secrets management is implemented:
	// - Write actual values to secrets.enc (encrypted)
	// - Keep secrets file as template with placeholders only
	if len(secrets) > 0 {
		secretsFile := filepath.Join(ctx.RootDir, "secrets")
		secretsContent := "\n"
		for _, secret := range secrets {
			// Write placeholder only (actual values stored in secretValues for future use)
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
		ui.Success(fmt.Sprintf("Added MCP '%s' to agent '%s'", name, agentID))
	} else {
		ui.Success(fmt.Sprintf("Created mcps/%s.yaml", name))
	}

	if len(secrets) > 0 {
		secretsList := strings.Join(secrets, ", ")
		ui.Success(fmt.Sprintf("Added %d secret(s): %s", len(secrets), secretsList))
	}

	if !noWizard && len(secrets) == 0 {
		fmt.Println()
		ui.Dimmed("Next steps:")
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

	// Show banner in interactive mode
	if !noWizard {
		ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Adding new SOP document                             MUXI │\n╰──────────────────────────────────────────────────────────────╯")
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

	// Show banner in interactive mode
	if !noWizard {
		ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Adding new trigger                                  MUXI │\n│──────────────────────────────────────────────────────────────│\n│ ℹ Triggers allow external systems to invoke your agents      │\n│ via webhooks or scheduled events.                            │\n╰──────────────────────────────────────────────────────────────╯")
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

// CreateA2AService creates a new A2A service configuration file in services/
func CreateA2AService(name string, noWizard bool) error {
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Show banner
	fmt.Println("╭──────────────────────────────────────────────────────────────╮")
	fmt.Println("│ [+] Adding new A2A service                              MUXI │")
	fmt.Println("╰──────────────────────────────────────────────────────────────╯")
	fmt.Println()

	// Ensure a2a directory exists (A2A services go in a2a/*.yaml per schema)
	a2aDir := filepath.Join(ctx.RootDir, "a2a")
	if err := os.MkdirAll(a2aDir, 0755); err != nil {
		return fmt.Errorf("failed to create a2a directory: %w", err)
	}

	var serviceID, serviceName, description, serviceURL string
	var authType, authHeader, authKey, authToken, authUsername, authPassword string
	var retryAttempts, timeoutSeconds int
	var customRetry bool
	var secrets []string
	secretValues := make(map[string]string)

	if !noWizard {
		// === REQUIRED FIELDS ===
		
		// Service ID - validation loop with normalization (accepts spaces, converts to kebab-case)
		for {
			var inputID string
			var err error
			if name != "" {
				inputID = name
				name = "" // Only use once
			} else {
				inputID, err = wizard.PromptString("Service ID", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
			}

			// Validate not empty
			if inputID == "" {
				ui.PromptError("Service ID", inputID, fmt.Errorf("service ID is required"))
				continue
			}

			// Normalize: spaces to hyphens, lowercase, etc.
			serviceID = normalizeComponentName(inputID)

			// Validate normalized format
			if err := validateComponentName(serviceID); err != nil {
				ui.PromptError("Service ID", inputID, err)
				continue
			}

			// Check for duplicates
			serviceFile := filepath.Join(a2aDir, serviceID+".yaml")
			if _, err := os.Stat(serviceFile); !os.IsNotExist(err) {
				ui.PromptError("Service ID", inputID, fmt.Errorf("service '%s' already exists", serviceID))
				continue
			}

			ui.PromptSuccess("Service ID", serviceID)
			break
		}

		// Service Name - suggest based on ID (title case)
		inferredName := titleCase(serviceID)
		serviceName, err = wizard.PromptString("Service name", inferredName, nil)
		if err != nil {
			fmt.Println()
			ui.Dimmed("A2A service creation cancelled")
			return nil
		}
		if serviceName == "" {
			serviceName = inferredName
		}
		ui.PromptSuccess("Service name", serviceName)

		// Description
		description, err = wizard.PromptString("Description", "", nil)
		if err != nil {
			fmt.Println()
			ui.Dimmed("A2A service creation cancelled")
			return nil
		}
		if description == "" {
			description = fmt.Sprintf("A2A service: %s", serviceName)
		}
		ui.PromptSuccess("Description", description)

		// Service URL - validation loop
		for {
			serviceURL, err = wizard.PromptString("Service URL", "", nil)
			if err != nil {
				fmt.Println()
				ui.Dimmed("A2A service creation cancelled")
				return nil
			}

			if serviceURL == "" {
				ui.PromptError("Service URL", serviceURL, fmt.Errorf("service URL is required"))
				continue
			}

			// Auto-add https:// if missing
			if !strings.HasPrefix(serviceURL, "https://") && !strings.HasPrefix(serviceURL, "http://") {
				serviceURL = "https://" + serviceURL
			}

			// Reject http://
			if strings.HasPrefix(serviceURL, "http://") {
				ui.PromptError("Service URL", serviceURL, fmt.Errorf("must use https:// (http:// is not secure)"))
				continue
			}

			// Validate URL format
			parsed, parseErr := url.Parse(serviceURL)
			if parseErr != nil || parsed.Host == "" {
				ui.PromptError("Service URL", serviceURL, fmt.Errorf("invalid URL format"))
				continue
			}

			ui.PromptSuccess("Service URL", serviceURL)
			break
		}

		// === AUTHENTICATION ===
		fmt.Println()
		authOptions := []wizard.SelectOption{
			{Value: "none", Label: "None"},
			{Value: "api_key", Label: "API Key"},
			{Value: "bearer", Label: "Bearer Token"},
			{Value: "basic", Label: "Basic Auth"},
		}

		authType, err = wizard.PromptSelect("Authentication", authOptions, 0)
		if err != nil {
			fmt.Println()
			ui.Dimmed("A2A service creation cancelled")
			return nil
		}

		// Show selected auth type
		for _, opt := range authOptions {
			if opt.Value == authType {
				ui.PromptSuccess("Authentication", opt.Label)
				break
			}
		}

		// Auth-specific prompts
		secretPrefix := "A2A_SERVICE_" + strings.ToUpper(strings.ReplaceAll(serviceID, "-", "_"))

		if authType != "none" {
			switch authType {
			case "api_key":
				defaultHeader := "X-API-Key"
				authHeader, err = wizard.PromptString("API Key header name", defaultHeader, nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
				ui.PromptSuccess("Header name", authHeader)

				authKey, err = wizard.PromptString("API Key value", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
				if len(authKey) > 4 {
					ui.PromptSuccess("API Key", "***"+authKey[len(authKey)-4:])
				} else {
					ui.PromptSuccess("API Key", "***")
				}

				secrets = append(secrets, secretPrefix+"_API_KEY")
				secretValues[secretPrefix+"_API_KEY"] = authKey

			case "bearer":
				authToken, err = wizard.PromptString("Bearer token value", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
				if len(authToken) > 4 {
					ui.PromptSuccess("Bearer token", "***"+authToken[len(authToken)-4:])
				} else {
					ui.PromptSuccess("Bearer token", "***")
				}
				secrets = append(secrets, secretPrefix+"_BEARER_TOKEN")
				secretValues[secretPrefix+"_BEARER_TOKEN"] = authToken

			case "basic":
				authUsername, err = wizard.PromptString("Username", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
				ui.PromptSuccess("Username", authUsername)

				authPassword, err = wizard.PromptString("Password", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A service creation cancelled")
					return nil
				}
				ui.PromptSuccess("Password", "***")

				secrets = append(secrets, secretPrefix+"_USERNAME", secretPrefix+"_PASSWORD")
				secretValues[secretPrefix+"_USERNAME"] = authUsername
				secretValues[secretPrefix+"_PASSWORD"] = authPassword
			}
		}

		// === RATE LIMITING ===
		fmt.Println()
		customRetryStr, err := wizard.PromptString("Configure custom retry/timeout? (y/N)", "", nil)
		if err != nil {
			fmt.Println()
			ui.Dimmed("A2A service creation cancelled")
			return nil
		}

		customRetry = strings.ToLower(strings.TrimSpace(customRetryStr)) == "y"
		if customRetry {
			// Retry attempts
			retryStr, err := wizard.PromptString("Retry attempts", "3", nil)
			if err != nil {
				fmt.Println()
				ui.Dimmed("A2A service creation cancelled")
				return nil
			}
			retryAttempts, _ = strconv.Atoi(retryStr)
			if retryAttempts <= 0 {
				retryAttempts = 3
			}
			ui.PromptSuccess("Retry attempts", strconv.Itoa(retryAttempts))

			// Timeout seconds
			timeoutStr, err := wizard.PromptString("Timeout (seconds)", "30", nil)
			if err != nil {
				fmt.Println()
				ui.Dimmed("A2A service creation cancelled")
				return nil
			}
			timeoutSeconds, _ = strconv.Atoi(timeoutStr)
			if timeoutSeconds <= 0 {
				timeoutSeconds = 30
			}
			ui.PromptSuccess("Timeout", fmt.Sprintf("%d seconds", timeoutSeconds))
		} else {
			ui.PromptSuccess("Retry/timeout", "Using defaults (3 retries, 30s timeout)")
		}

	} else {
		// Non-interactive mode
		if name == "" {
			return fmt.Errorf("service ID is required")
		}
		// Normalize the ID (spaces to hyphens, lowercase)
		serviceID = normalizeComponentName(name)
		
		// Validate normalized ID
		if err := validateComponentName(serviceID); err != nil {
			ui.ErrorBlock("Invalid service ID", err.Error(), "Example: external-billing")
			os.Exit(1)
		}
		
		// Check for duplicates
		serviceFile := filepath.Join(a2aDir, serviceID+".yaml")
		if _, err := os.Stat(serviceFile); !os.IsNotExist(err) {
			ui.ErrorBlock(
				"A2A service exists",
				fmt.Sprintf("File 'a2a/%s.yaml' already exists", serviceID),
				fmt.Sprintf("Choose a different ID or remove:\n  rm a2a/%s.yaml", serviceID),
			)
			os.Exit(1)
		}
		
		serviceName = titleCase(serviceID)
		description = fmt.Sprintf("A2A service: %s", serviceName)
		serviceURL = "https://api.example.com/a2a"
		authType = "none"
	}

	// Generate YAML content
	content := a2aServiceTemplate(serviceID, serviceName, description, serviceURL, authType, authHeader, authToken, authUsername, authPassword, customRetry, retryAttempts, timeoutSeconds, secrets)

	// Write file
	serviceFile := filepath.Join(a2aDir, serviceID+".yaml")
	if err := os.WriteFile(serviceFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
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

	// Success messages
	fmt.Println()
	ui.Success(fmt.Sprintf("Created a2a/%s.yaml", serviceID))

	if len(secrets) > 0 {
		secretsList := strings.Join(secrets, ", ")
		ui.Success(fmt.Sprintf("Added %d secret(s): %s", len(secrets), secretsList))
	}

	return nil
}

// a2aServiceTemplate generates the YAML content for an A2A service
func a2aServiceTemplate(id, name, description, serviceURL, authType, authHeader, authToken, authUsername, authPassword string, customRetry bool, retryAttempts, timeoutSeconds int, secrets []string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`schema: "1.0.0"
id: "%s"
name: "%s"
description: "%s"
url: "%s"
active: true

# Optional: Service provider metadata
# author: "Partner Name <email@partner.com>"
# version: "1.0.0"
# documentation: "https://docs.partner.com/a2a"
# support_contact: "support@partner.com"
`, id, name, description, serviceURL))

	// Auth section
	if authType != "none" && authType != "" {
		b.WriteString("\nauth:\n")
		
		secretPrefix := "A2A_SERVICE_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_"))
		
		switch authType {
		case "api_key":
			b.WriteString(fmt.Sprintf(`  type: "api_key"
  header: "%s"
  key: "${{ secrets.%s_API_KEY }}"
`, authHeader, secretPrefix))

		case "bearer":
			b.WriteString(fmt.Sprintf(`  type: "bearer"
  token: "${{ secrets.%s_BEARER_TOKEN }}"
`, secretPrefix))

		case "basic":
			b.WriteString(fmt.Sprintf(`  type: "basic"
  username: "${{ secrets.%s_USERNAME }}"
  password: "${{ secrets.%s_PASSWORD }}"
`, secretPrefix, secretPrefix))
		}
	}

	// Rate limiting section
	if customRetry {
		b.WriteString(fmt.Sprintf(`
retry_attempts: %d
timeout_seconds: %d
`, retryAttempts, timeoutSeconds))
	} else {
		b.WriteString(`
# Optional: Override default retry/timeout settings
# retry_attempts: 3
# timeout_seconds: 30
`)
	}

	return b.String()
}

// ConfigureA2A configures A2A in the formation (inbound or outbound)
func ConfigureA2A(inbound, outbound, noWizard bool) error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation\n\nOr create a new one:\n  muxi new formation",
		)
		os.Exit(1)
	}

	// Determine direction
	direction := ""
	if inbound && outbound {
		return fmt.Errorf("cannot specify both --inbound and --outbound")
	}

	if inbound {
		direction = "inbound"
		// Show inbound banner in interactive mode
		if !noWizard {
			ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [⚙] A2A inbound configuration                           MUXI │\n│──────────────────────────────────────────────────────────────│\n│ ℹ Inbound A2A allows external formations to discover and     │\n│ connect to your agents.                                      │\n╰──────────────────────────────────────────────────────────────╯")
		}
	} else if outbound {
		direction = "outbound"
		// Show outbound banner in interactive mode
		if !noWizard {
			ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [⚙] A2A outbound configuration                          MUXI │\n│──────────────────────────────────────────────────────────────│\n│ ℹ Outbound A2A allows your agents to discover and connect    │\n│ to external formations.                                      │\n╰──────────────────────────────────────────────────────────────╯")
		}
	} else if !noWizard {
		// Show generic banner when asking for direction
		ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [⚙] A2A configuration                                   MUXI │\n╰──────────────────────────────────────────────────────────────╯")

		// Ask for direction
		directionOptions := []wizard.SelectOption{
			{Value: "inbound", Label: "Inbound", Description: "Accept connections from external agents"},
			{Value: "outbound", Label: "Outbound", Description: "Connect to external agent services"},
		}

		direction, err = wizard.PromptSelect("A2A Direction", directionOptions, 0)
		if err != nil {
			return fmt.Errorf("failed to select direction: %w", err)
		}

		// Find the label for display
		for _, opt := range directionOptions {
			if opt.Value == direction {
				ui.PromptSuccess("A2A Direction", opt.Label)
				break
			}
		}
	} else {
		return fmt.Errorf("must specify --inbound or --outbound when using --no-wizard")
	}

	// Route to appropriate wizard
	if direction == "outbound" {
		return configureOutboundA2A(ctx.RootDir, noWizard)
	}

	// Inbound wizard
	return configureInboundA2A(ctx.RootDir, noWizard)
}

// configureInboundA2A configures inbound A2A in the formation
func configureInboundA2A(rootDir string, noWizard bool) error {
	// Check if A2A inbound is already configured
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}
	
	alreadyConfigured := strings.Contains(string(content), "a2a:") && strings.Contains(string(content), "inbound:")
	
	// Extract existing values to use as defaults
	var existingRegistries, existingAuthType, existingAuthHeader, existingTrustedEndpoints string
	var isEnabled bool
	if alreadyConfigured {
		existingRegistries = extractA2ARegistries(string(content))
		existingAuthType = extractA2AAuthType(string(content))
		existingAuthHeader = extractA2AAuthHeader(string(content))
		existingTrustedEndpoints = extractA2ATrustedEndpoints(string(content))
		isEnabled = extractA2AInboundEnabled(string(content))
	}
	
	if alreadyConfigured && !noWizard {
		fmt.Println()
		red := color.New(color.FgRed, color.Bold)
		
		if isEnabled {
			red.Println("  ⚠ A2A inbound is already enabled in the formation")
		} else {
			red.Println("  ⚠ A2A inbound is already configured in the formation")
		}
		fmt.Println()
		
		// Ask if they want to enable/disable
		var togglePrompt string
		if isEnabled {
			togglePrompt = "Disable inbound A2A? (y/N)"
		} else {
			togglePrompt = "Enable inbound A2A? (y/N)"
		}
		
		toggle, err := wizard.PromptString(togglePrompt, "", nil)
		if err != nil {
			// User cancelled (Ctrl+C)
			fmt.Println()
			ui.Dimmed("A2A configuration cancelled")
			return nil
		}
		if strings.ToLower(strings.TrimSpace(toggle)) == "y" {
			if isEnabled {
				// Disable and exit (user just wants to turn it off)
				if err := disableA2AInbound(rootDir); err != nil {
					return fmt.Errorf("failed to disable A2A inbound: %w", err)
				}
				fmt.Println()
				ui.Success("A2A inbound disabled in the formation")
				return nil
			} else {
				// Enable and continue to wizard (user wants to turn on + configure)
				if err := enableA2AInbound(rootDir); err != nil {
					return fmt.Errorf("failed to enable A2A inbound: %w", err)
				}
				fmt.Println()
				ui.Success("A2A inbound enabled in the formation")
				fmt.Println()
				// Fall through to wizard below
			}
		}
		
		// Only ask about replacement if user didn't just enable
		// (if they enabled, they want to configure it now)
		if !(strings.ToLower(strings.TrimSpace(toggle)) == "y" && !isEnabled) {
			fmt.Println()
			ui.Dimmed("This will replace the entire A2A inbound configuration.")
			ui.Dimmed("Existing values will be shown as defaults - press Enter to keep them.")
			fmt.Println()
			
			confirm, err := wizard.PromptString("Continue and replace? (y/N)", "", nil)
			if err != nil {
				// User cancelled (Ctrl+C)
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			if confirm == "" {
				confirm = "n" // Default to no
			}
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			fmt.Println()
		}
	}
	
	var registries []string
	var authType string
	var authHeader, authKey, authToken, authUsername, authPassword string
	var trustedEndpoints []string
	var secrets []string

	// TODO(secrets): When secrets management is implemented, use these collected values:
	// - Encrypt and store in secrets.enc
	// - Add to secrets file as placeholders (SECRET_NAME=)
	// - Provide `muxi secrets set SECRET_NAME value` command for updating
	var secretValues = make(map[string]string) // Store actual secret values (unused until secrets management exists)

	if !noWizard {
		// Registry URLs - loop until valid
		for {
			registriesStr, err := wizard.PromptString("Registry URLs (comma or line-separated)", existingRegistries, nil)
			if err != nil {
				// User cancelled (Ctrl+C)
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			if registriesStr == "" {
				ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("at least one registry URL is required"))
				continue
			}

			// Parse and validate registries
			registries = parseURLList(registriesStr)
			if len(registries) == 0 {
				ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("at least one valid registry URL is required"))
				continue
			}

			// Validate and normalize each registry URL
			valid := true
			for i, registry := range registries {
				// Auto-add https:// if no protocol specified
				if !strings.HasPrefix(registry, "http://") && !strings.HasPrefix(registry, "https://") {
					registry = "https://" + registry
					registries[i] = registry
				}

				// Ensure it's https (not http)
				if strings.HasPrefix(registry, "http://") {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (must use https://)", registry))
					valid = false
					break
				}

				// Validate URL format
				parsedURL, err := url.Parse(registry)
				if err != nil || parsedURL.Host == "" {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (invalid format)", registry))
					valid = false
					break
				}

				// Additional host validation
				host := parsedURL.Hostname()
				if strings.HasSuffix(host, ".") || strings.Contains(host, "..") || strings.HasPrefix(host, ".") {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (malformed hostname)", registry))
					valid = false
					break
				}
			}

			if !valid {
				continue
			}

			ui.PromptSuccess("Registries", fmt.Sprintf("%d added", len(registries)))
			break
		}

		// Authentication
		authOptions := []wizard.SelectOption{
			{Value: "none", Label: "None (not recommended)", Description: "No authentication"},
			{Value: "api_key", Label: "API Key", Description: "API key in header"},
			{Value: "bearer", Label: "Bearer Token", Description: "Bearer token authentication"},
			{Value: "basic", Label: "Basic Auth", Description: "Username and password"},
		}

		// Find default index based on existing auth type
		defaultAuthIndex := 0
		for i, opt := range authOptions {
			if opt.Value == existingAuthType {
				defaultAuthIndex = i
				break
			}
		}

		fmt.Println()
		var err error
		authType, err = wizard.PromptSelect("Authentication", authOptions, defaultAuthIndex)
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

		// Auth-specific prompts
		if authType != "none" {
			switch authType {
			case "api_key":
				defaultHeader := existingAuthHeader
				if defaultHeader == "" {
					defaultHeader = "X-API-Key"
				}
				authHeader, err = wizard.PromptString("API Key header name", defaultHeader, nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A configuration cancelled")
					return nil
				}
				ui.PromptSuccess("Header name", authHeader)

				authKey, err = wizard.PromptString("API Key value", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A configuration cancelled")
					return nil
				}
				if len(authKey) > 4 {
					ui.PromptSuccess("API Key", "***"+authKey[len(authKey)-4:])
				} else {
					ui.PromptSuccess("API Key", "***")
				}

				secrets = append(secrets, "A2A_INBOUND_API_KEY")
				secretValues["A2A_INBOUND_API_KEY"] = authKey

			case "bearer":
				authToken, err = wizard.PromptString("Bearer token value", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A configuration cancelled")
					return nil
				}
				if len(authToken) > 4 {
					ui.PromptSuccess("Bearer token", "***"+authToken[len(authToken)-4:])
				} else {
					ui.PromptSuccess("Bearer token", "***")
				}
				secrets = append(secrets, "A2A_INBOUND_BEARER_TOKEN")
				secretValues["A2A_INBOUND_BEARER_TOKEN"] = authToken

			case "basic":
				authUsername, err = wizard.PromptString("Username", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A configuration cancelled")
					return nil
				}
				ui.PromptSuccess("Username", authUsername)

				authPassword, err = wizard.PromptString("Password", "", nil)
				if err != nil {
					fmt.Println()
					ui.Dimmed("A2A configuration cancelled")
					return nil
				}
				ui.PromptSuccess("Password", "***")

				secrets = append(secrets, "A2A_INBOUND_USERNAME", "A2A_INBOUND_PASSWORD")
				secretValues["A2A_INBOUND_USERNAME"] = authUsername
				secretValues["A2A_INBOUND_PASSWORD"] = authPassword
			}
		} else {
			// Warning for no auth
			fmt.Println()
			ui.Warning("Inbound A2A without authentication is not recommended")
		}

		// Trusted endpoints (optional)
		fmt.Println()
		endpointsStr, err := wizard.PromptString("Trusted endpoints (optional, comma or line-separated)", existingTrustedEndpoints, nil)
		if err != nil {
			// User cancelled (Ctrl+C)
			fmt.Println()
			ui.Dimmed("A2A configuration cancelled")
			return nil
		}
		if endpointsStr != "" {
			trustedEndpoints = parseEndpointList(endpointsStr)
			ui.PromptSuccess("Trusted endpoints", fmt.Sprintf("%d added", len(trustedEndpoints)))
		} else {
			ui.PromptSkipped("Trusted endpoints")
		}
	} else {
		// Non-interactive defaults
		return fmt.Errorf("--no-wizard is not yet supported for A2A configuration")
	}

	// Modify the formation
	wasUpdated, err := updateFormationA2AInbound(rootDir, registries, authType, authHeader, authToken, authUsername, authPassword, trustedEndpoints)
	if err != nil {
		return fmt.Errorf("failed to update the formation: %w", err)
	}

	// Append secrets to secrets file (placeholders only)
	// TODO(secrets): When secrets management is implemented:
	// - Write actual values to secrets.enc (encrypted)
	// - Keep secrets file as template with placeholders only
	if len(secrets) > 0 {
		secretsFile := filepath.Join(rootDir, "secrets")
		secretsContent := "\n"
		for _, secret := range secrets {
			// Write placeholder only (actual values stored in secretValues for future use)
			secretsContent += secret + "=\n"
		}

		f, err := os.OpenFile(secretsFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(secretsContent)
			f.Close()
		}
	}

	fmt.Println()
	if wasUpdated {
		ui.Success("A2A inbound configuration updated in the formation")
	} else {
		ui.Success("A2A inbound configuration added to the formation")
	}

	if len(secrets) > 0 {
		secretsList := strings.Join(secrets, ", ")
		ui.Success(fmt.Sprintf("Added %d secret(s): %s", len(secrets), secretsList))
	}

	return nil
}

// configureOutboundA2A configures outbound A2A in the formation
func configureOutboundA2A(rootDir string, noWizard bool) error {
	// Check if A2A outbound is already configured
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}
	
	alreadyConfigured := strings.Contains(string(content), "a2a:") && strings.Contains(string(content), "outbound:")
	
	// Extract existing values to use as defaults
	var existingRegistries string
	var isEnabled bool
	if alreadyConfigured {
		existingRegistries = extractA2AOutboundRegistries(string(content))
		isEnabled = extractA2AOutboundEnabled(string(content))
	}
	
	if alreadyConfigured && !noWizard {
		fmt.Println()
		red := color.New(color.FgRed, color.Bold)
		
		if isEnabled {
			red.Println("  ⚠ A2A outbound is already enabled in the formation")
		} else {
			red.Println("  ⚠ A2A outbound is already configured in the formation")
		}
		fmt.Println()
		
		// Ask if they want to enable/disable
		var togglePrompt string
		if isEnabled {
			togglePrompt = "Disable outbound A2A? (y/N)"
		} else {
			togglePrompt = "Enable outbound A2A? (y/N)"
		}
		
		toggle, err := wizard.PromptString(togglePrompt, "", nil)
		if err != nil {
			// User cancelled (Ctrl+C)
			fmt.Println()
			ui.Dimmed("A2A configuration cancelled")
			return nil
		}
		if strings.ToLower(strings.TrimSpace(toggle)) == "y" {
			if isEnabled {
				// Disable and exit (user just wants to turn it off)
				if err := disableA2AOutbound(rootDir); err != nil {
					return fmt.Errorf("failed to disable A2A outbound: %w", err)
				}
				fmt.Println()
				ui.Success("A2A outbound disabled in the formation")
				return nil
			} else {
				// Enable and continue to wizard (user wants to turn on + configure)
				if err := enableA2AOutbound(rootDir); err != nil {
					return fmt.Errorf("failed to enable A2A outbound: %w", err)
				}
				fmt.Println()
				ui.Success("A2A outbound enabled in the formation")
				fmt.Println()
				// Fall through to wizard below
			}
		}
		
		// Only ask about replacement if user didn't just enable
		// (if they enabled, they want to configure it now)
		if !(strings.ToLower(strings.TrimSpace(toggle)) == "y" && !isEnabled) {
			fmt.Println()
			ui.Dimmed("This will replace the entire A2A outbound configuration.")
			ui.Dimmed("Existing values will be shown as defaults - press Enter to keep them.")
			fmt.Println()
			
			confirm, err := wizard.PromptString("Continue and replace? (y/N)", "", nil)
			if err != nil {
				// User cancelled (Ctrl+C)
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			if confirm == "" {
				confirm = "n" // Default to no
			}
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			fmt.Println()
		}
	}
	
	var registries []string

	if !noWizard {
		// Registry URLs - loop until valid
		for {
			registriesStr, err := wizard.PromptString("Registry URLs (comma or line-separated)", existingRegistries, nil)
			if err != nil {
				// User cancelled (Ctrl+C)
				fmt.Println()
				ui.Dimmed("A2A configuration cancelled")
				return nil
			}
			if registriesStr == "" {
				ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("at least one registry URL is required"))
				continue
			}

			// Parse and validate registries
			registries = parseURLList(registriesStr)
			if len(registries) == 0 {
				ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("at least one valid registry URL is required"))
				continue
			}

			// Validate and normalize each registry URL
			valid := true
			for i, registry := range registries {
				// Auto-add https:// if no protocol specified
				if !strings.HasPrefix(registry, "http://") && !strings.HasPrefix(registry, "https://") {
					registry = "https://" + registry
					registries[i] = registry
				}

				// Ensure it's https (not http)
				if strings.HasPrefix(registry, "http://") {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (must use https://)", registry))
					valid = false
					break
				}

				// Validate URL format
				parsedURL, err := url.Parse(registry)
				if err != nil || parsedURL.Host == "" {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (invalid format)", registry))
					valid = false
					break
				}

				// Additional host validation
				host := parsedURL.Hostname()
				if strings.HasSuffix(host, ".") || strings.Contains(host, "..") || strings.HasPrefix(host, ".") {
					ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid URL: %s (malformed hostname)", registry))
					valid = false
					break
				}
			}

			if !valid {
				continue
			}

			ui.PromptSuccess("Registries", fmt.Sprintf("%d added", len(registries)))
			break
		}
	} else {
		// Non-interactive defaults
		return fmt.Errorf("--no-wizard is not yet supported for A2A configuration")
	}

	// Modify the formation
	wasUpdated, err := updateFormationA2AOutbound(rootDir, registries)
	if err != nil {
		return fmt.Errorf("failed to update the formation: %w", err)
	}

	fmt.Println()
	if wasUpdated {
		ui.Success("A2A outbound configuration updated in the formation")
	} else {
		ui.Success("A2A outbound configuration added to the formation")
	}
	
	fmt.Println()
	ui.Dimmed("To configure remote A2A services (auth, endpoints, etc):")
	ui.Dimmed("  muxi new a2a-service")

	return nil
}

// parseURLList parses comma, space, or line-separated URLs
func parseURLList(input string) []string {
	if input == "" {
		return []string{}
	}

	// Split by comma, space, or newline (flexible input)
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == '\n' || r == ' '
	})

	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseEndpointList parses comma, space, or line-separated domain names
func parseEndpointList(input string) []string {
	if input == "" {
		return []string{}
	}

	// Split by comma, space, or newline (flexible input)
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == '\n' || r == ' '
	})

	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// updateFormationA2AInbound updates the a2a.inbound section in the formation
// Returns (wasUpdated, error) where wasUpdated is true if an existing config was replaced
func updateFormationA2AInbound(rootDir string, registries []string, authType, authHeader, authToken, authUsername, authPassword string, trustedEndpoints []string) (bool, error) {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return false, fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	wasUpdated := false

	// Build inbound configuration
	var inboundConfig strings.Builder
	inboundConfig.WriteString("  inbound:\n")
	inboundConfig.WriteString("    enabled: true\n")
	inboundConfig.WriteString("    port: 8181\n")
	inboundConfig.WriteString("    registries:\n")
	for _, registry := range registries {
		inboundConfig.WriteString(fmt.Sprintf("      - \"%s\"\n", registry))
	}

	if len(trustedEndpoints) > 0 {
		inboundConfig.WriteString("    trusted_endpoints:\n")
		for _, endpoint := range trustedEndpoints {
			inboundConfig.WriteString(fmt.Sprintf("      - \"%s\"\n", endpoint))
		}
	}

	if authType != "none" {
		inboundConfig.WriteString("    auth:\n")
		inboundConfig.WriteString(fmt.Sprintf("      type: \"%s\"\n", authType))

		switch authType {
		case "api_key":
			inboundConfig.WriteString(fmt.Sprintf("      header: \"%s\"\n", authHeader))
			inboundConfig.WriteString("      key: \"${{ secrets.A2A_INBOUND_API_KEY }}\"\n")
		case "bearer":
			inboundConfig.WriteString("      token: \"${{ secrets.A2A_INBOUND_BEARER_TOKEN }}\"\n")
		case "basic":
			inboundConfig.WriteString("      username: \"${{ secrets.A2A_INBOUND_USERNAME }}\"\n")
			inboundConfig.WriteString("      password: \"${{ secrets.A2A_INBOUND_PASSWORD }}\"\n")
		}
	}

	// Check if a2a section exists
	if strings.Contains(contentStr, "a2a:") {
		// A2A section exists - override it
		// TODO: Implement proper YAML parsing and merging
		// For now, remove old a2a section and add new one
		wasUpdated = true
		
		// Simple approach: find "a2a:" and remove everything until next top-level key or EOF
		lines := strings.Split(contentStr, "\n")
		var newLines []string
		inA2ASection := false
		
		for _, line := range lines {
			if strings.HasPrefix(line, "a2a:") {
				inA2ASection = true
				
				// Also remove the comment line before "a2a:" if it exists
				// Check last line in newLines
				if len(newLines) > 0 {
					lastLine := newLines[len(newLines)-1]
					if strings.Contains(lastLine, "# Agent-to-Agent") || strings.Contains(lastLine, "#Agent-to-Agent") {
						newLines = newLines[:len(newLines)-1] // Remove the comment
					}
				}
				continue
			}
			
			// Check if we're hitting a new top-level key (no leading spaces)
			if inA2ASection && len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && line != "" {
				inA2ASection = false
			}
			
			if !inA2ASection {
				newLines = append(newLines, line)
			}
		}
		
		contentStr = strings.Join(newLines, "\n")
		
		// Add new a2a section
		var a2aSection strings.Builder
		a2aSection.WriteString("\n# Agent-to-Agent communication\n")
		a2aSection.WriteString("a2a:\n")
		a2aSection.WriteString("  enabled: true\n\n")
		a2aSection.WriteString(inboundConfig.String())
		
		contentStr += a2aSection.String()
	} else {
		// No a2a section - add complete section at end
		var a2aSection strings.Builder
		a2aSection.WriteString("\n# Agent-to-Agent communication\n")
		a2aSection.WriteString("a2a:\n")
		a2aSection.WriteString("  enabled: true\n\n")
		a2aSection.WriteString(inboundConfig.String())

		contentStr += a2aSection.String()
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(contentStr), 0644); err != nil {
		return false, fmt.Errorf("failed to write the formation: %w", err)
	}

	return wasUpdated, nil
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

// extractA2ARegistries extracts registry URLs from the formation and returns as comma-separated string
func extractA2ARegistries(content string) string {
	lines := strings.Split(content, "\n")
	var registries []string
	inRegistries := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "registries:") {
			inRegistries = true
			continue
		}
		
		if inRegistries {
			// Check if we're still in registries section (indented with -)
			if strings.HasPrefix(trimmed, "-") {
				// Extract URL from "- "https://registry.com""
				url := strings.TrimPrefix(trimmed, "-")
				url = strings.TrimSpace(url)
				url = strings.Trim(url, `"`)
				registries = append(registries, url)
			} else if !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
				// No longer in registries section
				break
			}
		}
	}
	
	return strings.Join(registries, ", ")
}

// extractA2AAuthType extracts auth type from the formation
func extractA2AAuthType(content string) string {
	lines := strings.Split(content, "\n")
	inAuth := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Look for "auth:" under inbound section
		if strings.Contains(content, "inbound:") && strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
			continue
		}
		
		if inAuth && strings.HasPrefix(trimmed, "type:") {
			authType := strings.TrimPrefix(trimmed, "type:")
			authType = strings.TrimSpace(authType)
			authType = strings.Trim(authType, `"`)
			return authType
		}
	}
	
	return ""
}

// extractA2AAuthHeader extracts API key header name from the formation
func extractA2AAuthHeader(content string) string {
	lines := strings.Split(content, "\n")
	inAuth := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Look for "auth:" under inbound section
		if strings.Contains(content, "inbound:") && strings.HasPrefix(trimmed, "auth:") {
			inAuth = true
			continue
		}
		
		if inAuth && strings.HasPrefix(trimmed, "header:") {
			header := strings.TrimPrefix(trimmed, "header:")
			header = strings.TrimSpace(header)
			header = strings.Trim(header, `"`)
			return header
		}
	}
	
	return ""
}

// extractA2ATrustedEndpoints extracts trusted endpoints from the formation and returns as comma-separated string
func extractA2ATrustedEndpoints(content string) string {
	lines := strings.Split(content, "\n")
	var endpoints []string
	inEndpoints := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "trusted_endpoints:") {
			inEndpoints = true
			continue
		}
		
		if inEndpoints {
			// Check if we're still in trusted_endpoints section (indented with -)
			if strings.HasPrefix(trimmed, "-") {
				// Extract endpoint from "- "example.com""
				endpoint := strings.TrimPrefix(trimmed, "-")
				endpoint = strings.TrimSpace(endpoint)
				endpoint = strings.Trim(endpoint, `"`)
				endpoints = append(endpoints, endpoint)
			} else if !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
				// No longer in endpoints section
				break
			}
		}
	}
	
	return strings.Join(endpoints, ", ")
}

// disableA2AInbound disables inbound A2A in the formation
func disableA2AInbound(rootDir string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var result []string
	
	inInbound := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Find inbound section
		if strings.HasPrefix(trimmed, "inbound:") {
			inInbound = true
			result = append(result, line)
			continue
		}
		
		// If we're in inbound and find enabled line, change it to false
		if inInbound && strings.HasPrefix(trimmed, "enabled:") {
			// Replace enabled: true with enabled: false
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " \t")))
			result = append(result, indent+"enabled: false")
			inInbound = false
			continue
		}
		
		result = append(result, line)
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write the formation: %w", err)
	}

	return nil
}

// disableA2AOutbound disables outbound A2A in the formation
func disableA2AOutbound(rootDir string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var result []string
	
	inOutbound := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Find outbound section
		if strings.HasPrefix(trimmed, "outbound:") {
			inOutbound = true
			result = append(result, line)
			continue
		}
		
		// If we're in outbound and find enabled line, change it to false
		if inOutbound && strings.HasPrefix(trimmed, "enabled:") {
			// Replace enabled: true with enabled: false
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " \t")))
			result = append(result, indent+"enabled: false")
			inOutbound = false
			continue
		}
		
		result = append(result, line)
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write the formation: %w", err)
	}

	return nil
}

// enableA2AInbound enables inbound A2A in the formation
func enableA2AInbound(rootDir string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var result []string
	
	inInbound := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Find inbound section
		if strings.HasPrefix(trimmed, "inbound:") {
			inInbound = true
			result = append(result, line)
			continue
		}
		
		// If we're in inbound and find enabled line, change it to true
		if inInbound && strings.HasPrefix(trimmed, "enabled:") {
			// Replace enabled: false with enabled: true
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " \t")))
			result = append(result, indent+"enabled: true")
			inInbound = false
			continue
		}
		
		result = append(result, line)
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write the formation: %w", err)
	}

	return nil
}

// enableA2AOutbound enables outbound A2A in the formation
func enableA2AOutbound(rootDir string) error {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	var result []string
	
	inOutbound := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Find outbound section
		if strings.HasPrefix(trimmed, "outbound:") {
			inOutbound = true
			result = append(result, line)
			continue
		}
		
		// If we're in outbound and find enabled line, change it to true
		if inOutbound && strings.HasPrefix(trimmed, "enabled:") {
			// Replace enabled: false with enabled: true
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " \t")))
			result = append(result, indent+"enabled: true")
			inOutbound = false
			continue
		}
		
		result = append(result, line)
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write the formation: %w", err)
	}

	return nil
}

// extractA2AInboundEnabled checks if inbound A2A is enabled
func extractA2AInboundEnabled(content string) bool {
	lines := strings.Split(content, "\n")
	inInbound := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "inbound:") {
			inInbound = true
			continue
		}
		
		if inInbound && strings.HasPrefix(trimmed, "enabled:") {
			return strings.Contains(trimmed, "true")
		}
	}
	
	return false
}

// extractA2AOutboundEnabled checks if outbound A2A is enabled
func extractA2AOutboundEnabled(content string) bool {
	lines := strings.Split(content, "\n")
	inOutbound := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "outbound:") {
			inOutbound = true
			continue
		}
		
		if inOutbound && strings.HasPrefix(trimmed, "enabled:") {
			return strings.Contains(trimmed, "true")
		}
	}
	
	return false
}

// extractA2AOutboundRegistries extracts outbound registry URLs from the formation and returns as comma-separated string
func extractA2AOutboundRegistries(content string) string {
	lines := strings.Split(content, "\n")
	var registries []string
	inOutbound := false
	inRegistries := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "outbound:") {
			inOutbound = true
			continue
		}
		
		if inOutbound && strings.HasPrefix(trimmed, "registries:") {
			inRegistries = true
			continue
		}
		
		if inRegistries {
			// Check if we're still in registries section (indented with -)
			if strings.HasPrefix(trimmed, "-") {
				// Extract URL from "- "https://registry.com""
				url := strings.TrimPrefix(trimmed, "-")
				url = strings.TrimSpace(url)
				url = strings.Trim(url, `"`)
				registries = append(registries, url)
			} else if !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t") {
				// No longer in registries section
				break
			}
		}
	}
	
	return strings.Join(registries, ", ")
}

// updateFormationA2AOutbound updates the a2a.outbound section in the formation
// Returns (wasUpdated, error) where wasUpdated is true if an existing config was replaced
func updateFormationA2AOutbound(rootDir string, registries []string) (bool, error) {
	formationFile := filepath.Join(rootDir, "formation.yaml")
	content, err := os.ReadFile(formationFile)
	if err != nil {
		return false, fmt.Errorf("failed to read the formation: %w", err)
	}

	contentStr := string(content)
	wasUpdated := false

	// Build outbound configuration
	var outboundConfig strings.Builder
	outboundConfig.WriteString("  outbound:\n")
	outboundConfig.WriteString("    enabled: true\n")
	outboundConfig.WriteString("    # default_retry_attempts: 3\n")
	outboundConfig.WriteString("    # default_timeout_seconds: 30\n")
	outboundConfig.WriteString("    registries:\n")
	for _, registry := range registries {
		outboundConfig.WriteString(fmt.Sprintf("      - \"%s\"\n", registry))
	}

	// Check if a2a section exists
	if strings.Contains(contentStr, "a2a:") {
		// A2A section exists - check if outbound exists
		if strings.Contains(contentStr, "outbound:") {
			wasUpdated = true
			// Remove old outbound section
			lines := strings.Split(contentStr, "\n")
			var newLines []string
			inOutbound := false
			
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				
				if strings.HasPrefix(trimmed, "outbound:") {
					inOutbound = true
					continue
				}
				
				// Check if we're hitting a new top-level key (no leading spaces at a2a level)
				if inOutbound && len(line) > 0 && !strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "\t\t") {
					// Check if it's not empty and not indented (meaning next section)
					if trimmed != "" && (strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ")) {
						inOutbound = false
					}
				}
				
				if !inOutbound {
					newLines = append(newLines, line)
				}
			}
			
			contentStr = strings.Join(newLines, "\n")
		}
		
		// Add outbound section to existing a2a section
		// Find where to insert (after inbound or after a2a:)
		lines := strings.Split(contentStr, "\n")
		var result []string
		inserted := false
		
		for i, line := range lines {
			result = append(result, line)
			
			// Insert after the last line of inbound section OR after "a2a:" if no inbound
			if !inserted {
				trimmed := strings.TrimSpace(line)
				
				// Look ahead to see if next line is less indented (end of inbound section)
				if i+1 < len(lines) {
					nextLine := lines[i+1]
					nextTrimmed := strings.TrimSpace(nextLine)
					
					// If current line is in a2a section and next line is not indented as much or is a comment
					if strings.Contains(contentStr, "inbound:") && strings.Contains(line, "enabled:") && i > 0 {
						// We might be at the end of inbound, check if next is less indented
						if nextTrimmed == "" || (!strings.HasPrefix(nextLine, "    ") && strings.HasPrefix(line, "    ")) {
							result = append(result, "")
							result = append(result, strings.Split(strings.TrimRight(outboundConfig.String(), "\n"), "\n")...)
							inserted = true
						}
					}
				}
				
				// If we have a2a: but no inbound, insert after "enabled: true"
				if !strings.Contains(contentStr, "inbound:") && trimmed == "enabled: true" && strings.Contains(contentStr, "a2a:") {
					result = append(result, "")
					result = append(result, strings.Split(strings.TrimRight(outboundConfig.String(), "\n"), "\n")...)
					inserted = true
				}
			}
		}
		
		if !inserted {
			// Fallback: append at end
			result = append(result, "")
			result = append(result, strings.Split(strings.TrimRight(outboundConfig.String(), "\n"), "\n")...)
		}
		
		contentStr = strings.Join(result, "\n")
	} else {
		// No a2a section - create complete section with outbound only
		var a2aSection strings.Builder
		a2aSection.WriteString("\n# Agent-to-Agent communication\n")
		a2aSection.WriteString("a2a:\n")
		a2aSection.WriteString("  enabled: true\n\n")
		a2aSection.WriteString(outboundConfig.String())

		contentStr += a2aSection.String()
	}

	// Write back to file
	if err := os.WriteFile(formationFile, []byte(contentStr), 0644); err != nil {
		return false, fmt.Errorf("failed to write the formation: %w", err)
	}

	return wasUpdated, nil
}

// EditComponent opens a component file in the user's preferred editor
func EditComponent(component, id string) error {
	// Must be in formation directory
	ctx, err := context.MustDetectFormation()
	if err != nil {
		ui.ErrorBlock(
			"Not in formation directory",
			"This command must be run inside a formation directory.",
			"Navigate to your formation:\n  cd my-formation",
		)
		return err
	}

	var filePath string
	
	switch component {
	case "formation":
		filePath = filepath.Join(ctx.RootDir, "formation.yaml")
		
	case "agent":
		if id == "" {
			return fmt.Errorf("agent ID required: muxi edit agent <id>")
		}
		filePath = filepath.Join(ctx.RootDir, "agents", id+".yaml")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ui.ErrorBlock(
				"Agent not found",
				fmt.Sprintf("Agent '%s' does not exist.", id),
				fmt.Sprintf("Create it first:\n  muxi new agent %s", id),
			)
			return fmt.Errorf("agent not found: %s", id)
		}
		
	case "mcp":
		if id == "" {
			return fmt.Errorf("MCP ID required: muxi edit mcp <id>")
		}
		
		// Check formation-level first
		formationMCPFile := filepath.Join(ctx.RootDir, "mcps", id+".yaml")
		if _, err := os.Stat(formationMCPFile); err == nil {
			filePath = formationMCPFile
		} else {
			// Check if it's an agent-level MCP
			agentFile, found := findAgentWithMCP(ctx.RootDir, id)
			if found {
				filePath = agentFile
				fmt.Println()
				ui.Info(fmt.Sprintf("MCP '%s' is defined in agent file: %s", id, filepath.Base(agentFile)))
				fmt.Println()
			} else {
				ui.ErrorBlock(
					"MCP not found",
					fmt.Sprintf("MCP '%s' does not exist.", id),
					fmt.Sprintf("Create it first:\n  muxi new mcp %s", id),
				)
				return fmt.Errorf("MCP not found: %s", id)
			}
		}
		
	case "sop":
		if id == "" {
			return fmt.Errorf("SOP ID required: muxi edit sop <id>")
		}
		filePath = filepath.Join(ctx.RootDir, "sops", id+".md")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ui.ErrorBlock(
				"SOP not found",
				fmt.Sprintf("SOP '%s' does not exist.", id),
				fmt.Sprintf("Create it first:\n  muxi new sop %s", id),
			)
			return fmt.Errorf("SOP not found: %s", id)
		}
		
	case "trigger":
		if id == "" {
			return fmt.Errorf("trigger ID required: muxi edit trigger <id>")
		}
		filePath = filepath.Join(ctx.RootDir, "triggers", id+".yaml")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ui.ErrorBlock(
				"Trigger not found",
				fmt.Sprintf("Trigger '%s' does not exist.", id),
				fmt.Sprintf("Create it first:\n  muxi new trigger %s", id),
			)
			return fmt.Errorf("trigger not found: %s", id)
		}

	case "a2a-service":
		if id == "" {
			return fmt.Errorf("A2A service ID required: muxi edit a2a-service <id>")
		}
		filePath = filepath.Join(ctx.RootDir, "a2a", id+".yaml")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ui.ErrorBlock(
				"A2A service not found",
				fmt.Sprintf("A2A service '%s' does not exist.", id),
				fmt.Sprintf("Create it first:\n  muxi new a2a-service %s", id),
			)
			return fmt.Errorf("A2A service not found: %s", id)
		}
		
	default:
		return fmt.Errorf("unknown component type: %s\nSupported: formation, agent, mcp, sop, trigger, a2a-service", component)
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback to platform defaults
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vim"
		}
	}

	// Open file in editor
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening %s in %s...\n", filePath, editor)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

// findAgentWithMCP searches for an agent file that contains the given MCP ID
// Returns the agent file path and true if found
func findAgentWithMCP(rootDir, mcpID string) (string, bool) {
	agentsDir := filepath.Join(rootDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return "", false
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		agentFile := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(agentFile)
		if err != nil {
			continue
		}

		// Parse YAML to check mcp_servers
		var agentData struct {
			MCPServers []struct {
				ID string `yaml:"id"`
			} `yaml:"mcp_servers"`
		}

		if err := yaml.Unmarshal(content, &agentData); err != nil {
			continue
		}

		// Check if this agent has the MCP
		for _, mcp := range agentData.MCPServers {
			if mcp.ID == mcpID {
				return agentFile, true
			}
		}
	}

	return "", false
}
