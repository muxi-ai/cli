package scaffold

import "fmt"

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

func formationYAMLTemplate(name, description string) string {
	if description == "" {
		description = "A new MUXI formation"
	}

	return fmt.Sprintf(`schema: "1.0.0"

id: "%s"
description: "%s"
version: "1.0.0"

# Server configuration
server:
  host: "0.0.0.0"
  port: 8271
  api_keys:
    admin_key: "${{ secrets.FORMATION_ADMIN_API_KEY }}"
    client_key: "${{ secrets.FORMATION_CLIENT_API_KEY }}"

# Overlord configuration
overlord:
  persona: |
    You are a helpful AI assistant.

# LLM configuration
llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"
  
  models:
    - text: "openai/gpt-4o"
`, name, description)
}

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
