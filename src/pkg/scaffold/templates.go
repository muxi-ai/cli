package scaffold

import (
	"fmt"
	"strings"
)

func gitignoreTemplate() string {
	return `# MUXI - Never commit encryption key
.key

# MUXI - Encrypted secrets (uncomment to ignore)
# secrets.enc

# AI Assistants
.cursor/
.claude/
.factory/
.windsurf/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
`
}

func muxiTemplate() string {
	return `# Project defaults for this formation
# Uncomment and set values as needed

# Default profile for deployments
# profile: production

# Default registry for push/pull
# registry: private.company.com
`
}

// generateFormationYAML creates a dynamic formation.yaml based on wizard config
func generateFormationYAML(config *FormationConfig) string {
	description := config.Description
	if description == "" {
		description = "A new MUXI formation"
	}

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf(`schema: "1.0.0"

id: "%s"
description: "%s"
version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8271
  api_keys:
    admin_key: "${{ secrets.FORMATION_ADMIN_API_KEY }}"
    client_key: "${{ secrets.FORMATION_CLIENT_API_KEY }}"

overlord:
  persona: |
    You are a helpful AI assistant.
`, config.Name, description))

	// Streaming (only if enabled)
	if config.EnableStreaming {
		b.WriteString(`  response:
    streaming: true

`)
	} else {
		b.WriteString("\n")
	}

	// Async (only if enabled with webhook)
	if config.EnableAsync && config.WebhookURL != "" {
		b.WriteString(fmt.Sprintf(`async:
  threshold_seconds: 30
  webhook_url: "%s"

`, config.WebhookURL))
	}

	// LLM configuration based on provider type
	switch config.ProviderType {
	case "cloud":
		if config.Provider != nil {
			b.WriteString(fmt.Sprintf(`llm:
  api_keys:
    %s: "${{ secrets.%s }}"
  models:
    - text: "%s/%s"

`, config.Provider.Vendor, config.Provider.SecretName, config.Provider.Vendor, config.Provider.DefaultModel))
		}
	case "local":
		if config.LocalProvider != nil {
			b.WriteString(fmt.Sprintf(`llm:
  api_keys:
    %s: "local"
  models:
    - text: "%s/%s"
      base_url: "%s"

`, config.LocalProvider.Vendor, config.LocalProvider.Vendor, config.LocalModel, config.LocalBaseURL))
		}
	case "enterprise":
		// For enterprise, we add commented template
		if config.EnterpriseProvider != nil {
			b.WriteString(config.EnterpriseProvider.YAMLTemplate)
			b.WriteString("\n\n")
		}
	default:
		// Default to OpenAI if no provider set (non-wizard mode)
		b.WriteString(`llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"
  models:
    - text: "openai/gpt-4o"

`)
	}

	// Scheduler with UTC timezone
	b.WriteString(`scheduler:
  timezone: "UTC"

`)

	// Additional configuration section
	b.WriteString(`# ─────────────────────────────────────────────────────────────
# Additional configuration (uncomment/edit as needed)
# ─────────────────────────────────────────────────────────────

# Add more LLM providers (use 'muxi config llm' for guided setup):
# llm:
#   api_keys:
#     anthropic: "${{ secrets.ANTHROPIC_API_KEY }}"

# Persistent memory (use 'muxi config memory' for guided setup):
# memory:
#   persistent:
#     connection_string: "postgres://user:pass@host:5432/db"

# Logging streams (use 'muxi config logging' for guided setup):
# logging:
#   streams:
#     - transport: "stdout"
#       level: "info"

# Input limits (defaults shown):
# input_limits:
#   max_message_length: 100000
#   max_file_size_bytes: 52428800

# Runtime settings:
# runtime:
#   built_in_mcps: true

# User credentials mode:
# user_credentials:
#   mode: "redirect"
`)

	return b.String()
}

// generateSecretsTemplate creates a secrets template based on provider
func generateSecretsTemplate(config *FormationConfig) string {
	var b strings.Builder

	b.WriteString(`# Secret keys for this formation
# Run 'muxi secrets setup' to configure values

FORMATION_ADMIN_API_KEY=
FORMATION_CLIENT_API_KEY=
`)

	// Add provider-specific secret
	switch config.ProviderType {
	case "cloud":
		if config.Provider != nil {
			b.WriteString(fmt.Sprintf("%s=\n", config.Provider.SecretName))
		}
	case "enterprise":
		if config.EnterpriseProvider != nil {
			switch config.EnterpriseProvider.Vendor {
			case "azure":
				b.WriteString("AZURE_API_KEY=\n")
			}
		}
	default:
		b.WriteString("OPENAI_API_KEY=\n")
	}

	return b.String()
}

// secretsTemplate returns the default secrets template (for backward compatibility)
func secretsTemplate() string {
	return `# Secret keys for this formation
# Run 'muxi secrets setup' to configure values

FORMATION_ADMIN_API_KEY=
FORMATION_CLIENT_API_KEY=
OPENAI_API_KEY=
`
}

func readmeTemplate(name, description string) string {
	if description == "" {
		description = "A new MUXI formation"
	}

	return fmt.Sprintf(`# %s

%s

Created with MUXI CLI.

## Structure

- ` + "`formation.yaml`" + ` - Formation configuration
- ` + "`agents/`" + ` - Agent definitions
- ` + "`mcps/`" + ` - MCP server configurations
- ` + "`a2a/`" + ` - Agent-to-Agent communication configs
- ` + "`sops/`" + ` - Standard Operating Procedures
- ` + "`triggers/`" + ` - Event triggers
- ` + "`knowledge/`" + ` - Knowledge base documents
- ` + "`secrets.enc`" + ` - Encrypted secrets
- ` + "`.key`" + ` - Encryption key (**never commit!**)

## Getting Started

1. Configure secrets (if not done during init):
   ` + "```bash\n" + `   muxi secrets setup
   ` + "```\n" + `

2. Validate formation:
   ` + "```bash\n" + `   muxi validate
   ` + "```\n" + `

3. Deploy:
   ` + "```bash\n" + `   muxi deploy --profile production
   ` + "```\n" + `

## Adding Components

` + "```bash\n" + `muxi new agent my-agent
muxi new mcp my-mcp
muxi new sop my-procedure
muxi new trigger my-trigger
muxi new a2a external-api
` + "```\n" + `

## Development

Edit ` + "`formation.yaml`" + ` and component files, then:

` + "```bash\n" + `muxi validate          # Check configuration
muxi deploy --profile localhost  # Test locally
` + "```\n" + `
`, name, description)
}
