# MUXI CLI - Scaffolding Guide

**Purpose:** Complete reference for `muxi new` commands - local file generation for formations and components.

**Date:** 2025-11-25

---

## Overview

The `muxi new` command family creates local files and directories for MUXI formations and their components. These are scaffolding commands - they don't interact with servers, they just generate the right file structure.

**Pattern:** `muxi new <type> [name] [flags]`

---

## Formation Scaffolding

### `muxi new formation`

Creates a complete formation directory with all necessary files and structure.

#### Modes

**1. Interactive Mode (Default)**

```bash
# No name specified - wizard asks for details
muxi new formation

Create new formation:
  Formation name: my-bot
  Description (optional): My awesome chatbot

Creating formation 'my-bot'...
  ✓ Directory structure created
  ✓ Formation keys generated (admin & client)

Setup secrets:
  [1/1] OPENAI_API_KEY (optional)
    Enter API key (leave empty to skip): sk-...

  ✓ Secrets configured

✓ Formation 'my-bot' created successfully!

Next steps:
  cd my-bot
  muxi validate
  muxi deploy --profile production
```

**2. Interactive with Name**

```bash
# Name specified - skips name prompt but runs wizard
muxi new formation my-bot

Creating formation 'my-bot'...
  ✓ Directory structure created
  ✓ Formation keys generated

Setup secrets:
  [1/1] OPENAI_API_KEY (optional)
    Enter API key (leave empty to skip): sk-...

✓ Formation 'my-bot' created successfully!
```

**3. Non-Interactive Mode**

```bash
# Skip wizard entirely
muxi new formation my-bot --no-wizard

✓ Formation 'my-bot' created

Files created:
  • .gitignore, .key, .muxi, formation.yaml
  • secrets (template with 3 keys)
  • README.md
  • 6 directories (agents/, mcps/, a2a/, sops/, triggers/, knowledge/)

⚠ Configure secrets before deploying:
    cd my-bot
    muxi secrets setup
```

---

### Files Created

```
my-formation/
├── .gitignore              # Git ignore rules
├── .key                    # 32-byte encryption key (auto-generated)
├── .muxi                   # Project defaults (commented)
├── formation.yaml          # Formation configuration
├── secrets                 # Secret keys template
├── README.md               # Getting started guide
├── agents/
│   └── .gitkeep
├── mcps/
│   └── .gitkeep
├── a2a/
│   └── .gitkeep
├── sops/
│   └── .gitkeep
├── triggers/
│   └── .gitkeep
└── knowledge/
    └── .gitkeep
```

**Note:** `secrets.enc` is only created if user provides OPENAI_API_KEY in wizard

---

### File Templates

#### `.gitignore`

```gitignore
# MUXI - Never commit encryption key
.key

# MUXI - Encrypted secrets (uncomment to ignore)
# secrets.enc

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
```

#### `.key`

```
<32-byte random hex string>
```

Auto-generated using cryptographically secure random generator.

#### `.muxi`

```yaml
# Project defaults for this formation
# Uncomment and set values as needed

# Default profile for deployments
# profile: production

# Default registry for push/pull
# registry: private.company.com
```

#### `formation.yaml`

```yaml
schema: "1.0.0"
id: "my-formation"
description: "A new MUXI formation"
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
```

**Notes:**
- `id` field uses user-provided formation name
- `description` uses user-provided description (wizard mode) or default
- FORMATION_* keys are auto-generated and stored in secrets.enc

#### `secrets`

```
# Secret keys for this formation
# Run 'muxi secrets setup' to configure values

FORMATION_ADMIN_API_KEY=
FORMATION_CLIENT_API_KEY=
OPENAI_API_KEY=
```

**Pattern:** Template file showing required keys (plaintext, safe to commit)

#### `README.md`

```markdown
# <Formation Name>

<Description from wizard or default>

Created with MUXI CLI.

## Structure

- `formation.yaml` - Formation configuration
- `agents/` - Agent definitions
- `mcps/` - MCP server configurations
- `a2a/` - Agent-to-Agent communication configs
- `sops/` - Standard Operating Procedures
- `triggers/` - Event triggers
- `knowledge/` - Knowledge base documents
- `secrets.enc` - Encrypted secrets
- `.key` - Encryption key (**never commit!**)

## Getting Started

1. Configure secrets (if not done during init):
   ```bash
   muxi secrets setup
   ```

2. Validate formation:
   ```bash
   muxi validate
   ```

3. Deploy:
   ```bash
   muxi deploy --profile production
   ```

## Adding Components

```bash
muxi new agent my-agent
muxi new mcp my-mcp
muxi new sop my-procedure
muxi new trigger my-trigger
muxi new a2a external-api
```

## Development

Edit `formation.yaml` and component files, then:

```bash
muxi validate          # Check configuration
muxi deploy --profile localhost  # Test locally
```
```

---

### Wizard Behavior

**1. Formation Name (if not provided as arg)**

```
Formation name: _
```

**Validation:**
- Lowercase alphanumeric + hyphens only
- Must start with letter
- 3-50 characters
- Error if directory already exists

**2. Description (optional)**

```
Description (optional): _
```

**Validation:**
- Max 200 characters
- Optional (can press Enter to skip)

**3. Auto-generate Keys**

No prompt - automatically generates:
- `FORMATION_ADMIN_API_KEY=fma_<32 random chars>`
- `FORMATION_CLIENT_API_KEY=fmc_<32 random chars>`

**4. OPENAI_API_KEY (optional)**

```
Setup secrets:
  [1/1] OPENAI_API_KEY (optional)
    Enter API key (leave empty to skip): _
```

**Validation:**
- Starts with `sk-` (OpenAI format)
- Optional (can be empty)
- If provided: creates `secrets.enc` immediately
- If skipped: shows reminder message

**5. Success Message**

```
✓ Formation 'my-bot' created successfully!

Next steps:
  cd my-bot
  muxi validate
  muxi deploy --profile production
```

---

### Error Handling

**Directory already exists:**

```bash
muxi new formation my-bot

✗ Formation directory 'my-bot' already exists

  Choose a different name or remove the existing directory:
    rm -rf my-bot
```

**Invalid formation name:**

```bash
muxi new formation My-Bot

✗ Invalid formation name 'My-Bot'

  Formation names must:
    • Be lowercase
    • Start with a letter
    • Contain only letters, numbers, and hyphens
    • Be 3-50 characters long
  
  Example: my-bot
```

---

## Component Scaffolding

All component commands follow the same pattern and must be run **inside a formation directory**.

### Context Detection

All `muxi new <component>` commands (except `formation`) require being in a formation directory:

```bash
# Outside formation directory
cd ~/projects
muxi new agent weather

✗ Not in a formation directory

  Run this command from inside a formation directory:
    cd my-formation
    muxi new agent weather
  
  Or create a new formation:
    muxi new formation
```

**Detection:** Walk up directory tree (max 5 levels) looking for `formation.yaml`

---

### `muxi new agent <name>`

Creates an agent configuration file.

**Interactive Mode (Default):**

```bash
muxi new agent weather

Create new agent 'weather':
  Description: Provides weather information
  Model [gpt-4o]: gpt-4o-mini
  Max tokens [4000]: 2000

✓ Created agents/weather.yaml

Edit the file to:
  • Add tools and capabilities
  • Configure prompts and persona
  • Set up workflows
```

**Non-Interactive Mode:**

```bash
muxi new agent weather --no-wizard

✓ Created agents/weather.yaml
```

**File Template (`agents/weather.yaml`):**

```yaml
id: weather
description: "Provides weather information"
active: true

model:
  name: "openai/gpt-4o-mini"
  max_tokens: 2000
  temperature: 0.7

persona: |
  You are a weather information assistant.

tools: []
workflows: []
```

**Wizard Fields:**

1. **Description** (optional)
   - Max 200 characters
   - Used in `description` field

2. **Model** (default: gpt-4o)
   - Common options shown
   - Custom input allowed

3. **Max tokens** (default: 4000)
   - Number validation
   - Range: 100-128000

---

### `muxi new mcp <name>`

Creates an MCP server configuration file.

**Interactive Mode:**

```bash
muxi new mcp postgres

Create new MCP server 'postgres':
  Type: [stdio] / sse
  Command: psql
  Arguments (optional): -h localhost
  Environment variables (optional): DATABASE_URL=${{ secrets.DATABASE_URL }}

✓ Created mcps/postgres.yaml

Edit the file to configure connection details.
```

**Non-Interactive Mode:**

```bash
muxi new mcp postgres --no-wizard

✓ Created mcps/postgres.yaml
```

**File Template (`mcps/postgres.yaml`):**

```yaml
id: postgres
description: "PostgreSQL MCP server"
type: stdio

connection:
  command: psql
  args:
    - "-h"
    - "localhost"
  env:
    DATABASE_URL: "${{ secrets.DATABASE_URL }}"

active: true
```

**Wizard Fields:**

1. **Type** (default: stdio)
   - Options: stdio, sse
   - Arrow keys to select

2. **Command**
   - Required
   - Executable name or path

3. **Arguments** (optional)
   - Space-separated
   - Parsed into array

4. **Environment variables** (optional)
   - KEY=value format
   - Can reference secrets

---

### `muxi new a2a <name>`

Creates an Agent-to-Agent communication configuration.

**Interactive Mode:**

```bash
muxi new a2a external-api

Create new A2A config 'external-api':
  Type: [rest] / grpc / websocket
  Base URL: https://api.example.com
  Authentication: [bearer] / basic / apikey / none

✓ Created a2a/external-api.yaml
```

**Non-Interactive Mode:**

```bash
muxi new a2a external-api --no-wizard

✓ Created a2a/external-api.yaml
```

**File Template (`a2a/external-api.yaml`):**

```yaml
id: external-api
description: "External API integration"
type: rest

connection:
  base_url: "https://api.example.com"
  auth:
    type: bearer
    token: "${{ secrets.EXTERNAL_API_TOKEN }}"

endpoints: []
active: true
```

---

### `muxi new sop <name>`

Creates a Standard Operating Procedure document.

**Interactive Mode:**

```bash
muxi new sop customer-onboarding

Create new SOP 'customer-onboarding':
  Title: Customer Onboarding Process
  Description (optional): Steps for onboarding new customers

✓ Created sops/customer-onboarding.md
```

**Non-Interactive Mode:**

```bash
muxi new sop customer-onboarding --no-wizard

✓ Created sops/customer-onboarding.md
```

**File Template (`sops/customer-onboarding.md`):**

```markdown
# Customer Onboarding Process

**Created:** 2025-11-25  
**Status:** Draft

## Overview

Steps for onboarding new customers

## Prerequisites

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
```

**Wizard Fields:**

1. **Title**
   - Required
   - Used as H1 heading

2. **Description** (optional)
   - Added to Overview section

---

### `muxi new trigger <name>`

Creates a trigger configuration file.

**Interactive Mode:**

```bash
muxi new trigger webhook

Create new trigger 'webhook':
  Type: [webhook] / schedule / event
  Description (optional): Webhook for external events

✓ Created triggers/webhook.yaml
```

**Non-Interactive Mode:**

```bash
muxi new trigger webhook --no-wizard

✓ Created triggers/webhook.yaml
```

**File Template (`triggers/webhook.yaml`):**

```yaml
id: webhook
description: "Webhook for external events"
type: webhook

config:
  path: "/webhook"
  method: POST
  auth:
    type: bearer
    token: "${{ secrets.WEBHOOK_TOKEN }}"

handler:
  agent: overlord
  workflow: process_webhook

active: true
```

**Wizard Fields:**

1. **Type** (default: webhook)
   - Options: webhook, schedule, event
   - Arrow keys to select

2. **Description** (optional)
   - Max 200 characters

---

## Global Flags

All `muxi new` commands support:

**`--no-wizard`**
- Skip interactive prompts
- Use defaults for all values
- Useful for CI/CD, quick scaffolding, power users

**Example:**
```bash
muxi new agent weather --no-wizard
muxi new mcp postgres --no-wizard
```

---

## Validation Rules

### Formation Names

- **Pattern:** `^[a-z][a-z0-9-]{2,49}$`
- **Must:** Start with lowercase letter
- **Can:** Contain lowercase letters, numbers, hyphens
- **Length:** 3-50 characters
- **Invalid:** `My-Bot`, `123bot`, `-mybot`, `a`, `my_bot`
- **Valid:** `my-bot`, `chatbot-v2`, `support-agent`

### Component Names (agent, mcp, a2a, trigger)

- **Pattern:** `^[a-z][a-z0-9-_]{2,49}$`
- **Must:** Start with lowercase letter
- **Can:** Contain lowercase letters, numbers, hyphens, underscores
- **Length:** 3-50 characters
- **Invalid:** `My-Agent`, `123agent`, `-agent`
- **Valid:** `weather`, `postgres-db`, `slack_webhook`

### SOP Names

- **Pattern:** `^[a-z][a-z0-9-_]{2,49}$`
- **Output:** Creates kebab-case markdown file
- **Example:** `customer-onboarding` → `customer-onboarding.md`

---

## Workflows

### Complete Formation Setup

```bash
# 1. Create formation (interactive)
muxi new formation my-bot
# Wizard asks for name, description, OPENAI_API_KEY

cd my-bot/

# 2. Add components
muxi new agent weather
muxi new agent research
muxi new mcp postgres
muxi new sop customer-support

# 3. Configure
vim formation.yaml
vim agents/weather.yaml

# 4. Validate
muxi validate

# 5. Deploy
muxi deploy --profile production
```

### Quick Scaffold (CI/CD)

```bash
# Non-interactive formation creation
muxi new formation my-bot --no-wizard
cd my-bot/

# Add components without wizard
muxi new agent weather --no-wizard
muxi new mcp postgres --no-wizard

# Set secrets programmatically
muxi secrets set OPENAI_API_KEY --value "$OPENAI_KEY"
muxi secrets set DATABASE_URL --value "$DB_URL"

# Deploy
muxi deploy --profile production
```

---

## Best Practices

### 1. Use Interactive Mode for First-Time Setup

Wizard helps you configure correctly:
```bash
muxi new formation          # Learn the questions
muxi new agent weather      # Understand options
```

### 2. Use --no-wizard for Repetitive Tasks

Once you know the patterns:
```bash
muxi new agent agent-{1..5} --no-wizard
```

### 3. Commit .muxi to Git

It's a project setting, not a secret:
```yaml
# .muxi
profile: production
registry: private.company.com
```

### 4. Keep secrets Template

The `secrets` file is documentation:
```bash
# Don't delete it!
# Add new keys as you add them to secrets.enc
```

### 5. Use Meaningful Names

Component names become IDs:
```bash
muxi new agent weather      # ✓ Clear purpose
muxi new agent agent1       # ✗ Unclear
```

---

## Error Reference

### Common Errors

**Not in formation directory:**
```
✗ Not in a formation directory

  Run this command from inside a formation directory or create one:
    muxi new formation
```

**File already exists:**
```
✗ File 'agents/weather.yaml' already exists

  Choose a different name or remove the file:
    rm agents/weather.yaml
```

**Invalid name:**
```
✗ Invalid agent name 'My-Agent'

  Component names must:
    • Be lowercase
    • Start with a letter
    • Contain only letters, numbers, hyphens, and underscores
    • Be 3-50 characters long
```

**Formation directory exists:**
```
✗ Formation directory 'my-bot' already exists

  Choose a different name or remove the directory:
    rm -rf my-bot
```

---

## Implementation Notes

### File Generation

**All scaffolding commands:**
1. Validate inputs (name, flags)
2. Check context (formation directory for components)
3. Check file/directory doesn't exist
4. Run wizard (if not --no-wizard)
5. Generate files from templates
6. Auto-generate keys/IDs as needed
7. Show success message + next steps

### Template System

**Use Go text/template:**
- Templates in `pkg/scaffold/templates/`
- Variables: `{{.Name}}`, `{{.Description}}`, etc.
- Conditional sections based on wizard answers

### Key Generation

**Formation keys:**
```go
// Generate FORMATION_ADMIN_API_KEY
adminKey := "fma_" + generateRandomString(32)

// Generate FORMATION_CLIENT_API_KEY
clientKey := "fmc_" + generateRandomString(32)
```

**Encryption key (.key file):**
```go
// 32-byte random for AES-256-GCM
key := make([]byte, 32)
rand.Read(key)
```

---

## Testing Strategy

### Unit Tests

Test each component:
```go
func TestNewFormation(t *testing.T) {
    // Test directory creation
    // Test file templates
    // Test key generation
    // Test wizard logic
}

func TestNewAgent(t *testing.T) {
    // Test in formation context
    // Test outside formation (error)
    // Test file already exists (error)
}
```

### Integration Tests

End-to-end workflows:
```bash
# Test complete formation setup
muxi new formation test-bot --no-wizard
cd test-bot
muxi new agent test-agent --no-wizard
muxi validate  # Should pass
```

---

## Future Enhancements

**Ideas for later:**

1. **Extended wizard questions:**
   - LLM provider selection
   - Model configuration
   - Persistent memory toggle
   - Logging level

2. **Templates:**
   ```bash
   muxi new formation --template chatbot
   muxi new formation --template workflow-automation
   ```

3. **Component from examples:**
   ```bash
   muxi new agent --from examples/weather-agent.yaml
   ```

4. **Bulk creation:**
   ```bash
   muxi new agents weather research support --no-wizard
   ```

---

**Ready for implementation!** 🚀

**Next:** Start building Day 1 packages tomorrow (config, secrets, scaffold)
