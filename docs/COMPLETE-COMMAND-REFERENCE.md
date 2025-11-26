# Complete CLI Command Reference

**Date:** 2025-11-25
**Timeline:** 1 Week
**Total Commands:** ~80

---

## Implementation Priority

**Day 1-2: Core** - Formation dev, secrets, profiles
**Day 3: Server** - Lifecycle, management
**Day 4-5: Formation API** - Agents, MCPs, chat, monitoring
**Day 6-7: Polish** - Testing, docs, UX

---

## Complete Command Tree

```
muxi
│
├── ℹ️  GLOBAL COMMANDS
│   ├── --version                      # Show CLI version
│   ├── --help                         # Show help and commands
│   └── <command> --help               # Help for specific command
│
├── 🏗️  FORMATION DEVELOPMENT (Local)
│   ├── new formation my-bot           # Creates my-bot/ directory
│   ├── new agent weather              # Creates agents/weather.yaml
│   ├── new mcp postgres               # Creates mcps/postgres.yaml
│   ├── new sop onboarding             # Creates sops/onboarding.md
│   ├── new trigger webhook            # Creates triggers/webhook.md
│   ├── new a2a external-api           # Creates a2a/external-api.yaml
│   ├── validate                       # Validate formation files
│   └── deploy [--profile <name>]      # Deploy to server(s)
│
├── 🔐 SECRET MANAGEMENT (Local)
│   ├── secrets setup                  # Setup wizard (only prompts for new secrets)
│   ├── secrets sync                   # Sync secrets file with formation YAMLs
│   ├── secrets set <key>              # Set secret (prompt for value)
│   ├── secrets list                   # List secret keys (no values)
│   └── secrets delete <key>           # Delete secret (with validation)
│
├── 🌐 REGISTRY
│   ├── registry login [--registry <url>]          # Authenticate (default: registry.muxi.org)
│   ├── registry logout [--registry <url>]         # Logout
│   ├── registry search <query> [--registry <url>] # Search formations
│   ├── push [--registry <url>]                    # Publish formation
│   └── pull <ref> [--registry <url>]              # Download formation
│
├── 🖥️  PROFILE MANAGEMENT
│   ├── profile add <name>             # Add server profile (wizard)
│   ├── profile list                   # List all profiles
│   ├── profile use <name>             # Set default profile
│   └── profile remove <name>          # Remove profile
│
├── 📦 FORMATION LIFECYCLE (Server API)
│   ├── formation list [--profile <name>]      # List formations on server
│   ├── formation stop <id> [--profile <name>] # Stop formation
│   ├── formation restart <id> [--profile <name>] # Restart formation
│   ├── formation rollback <id> [--profile <name>] # Rollback version
│   └── formation delete <id> [--profile <name>] # Delete formation
│
├── 🖥️  SERVER MANAGEMENT (Server API)
│   ├── server version [--profile <name>] # Server version info
│   ├── server status [--profile <name>]  # Server status & stats
│   ├── server logs [--profile <name>]    # Server audit logs
│   └── server ping [--profile <name>]    # Test reachability
│
├── ⚙️  FORMATION CONFIGURATION (Formation API)
│   ├── config show [--formation <id>] [--profile <name>] # Full config
│   ├── status [--formation <id>] [--profile <name>]      # Runtime status
│   │
│   ├── agent
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── add [--formation <id>] [--profile <name>] [--file <yaml>]
│   │   ├── get <agent_id> [--formation <id>] [--profile <name>]
│   │   ├── update <agent_id> [--formation <id>] [--profile <name>]
│   │   └── delete <agent_id> [--formation <id>] [--profile <name>]
│   │
│   ├── mcp
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── add [--formation <id>] [--profile <name>] [--file <yaml>]
│   │   ├── get <server_id> [--formation <id>] [--profile <name>]
│   │   ├── update <server_id> [--formation <id>] [--profile <name>]
│   │   └── delete <server_id> [--formation <id>] [--profile <name>]
│   │
│   ├── overlord
│   │   ├── show [--formation <id>] [--profile <name>]     # Get overlord config
│   │   └── persona [--formation <id>] [--profile <name>]  # Get persona
│   │
│   ├── llm
│   │   ├── settings [--formation <id>] [--profile <name>] # Get LLM settings
│   │   └── models [--formation <id>] [--profile <name>]   # List configured models
│   │
│   └── memory
│       ├── config [--formation <id>] [--profile <name>]   # Get memory config
│       └── buffer list [--formation <id>] [--profile <name>] # List buffer entries
│
├── 💬 INTERACTION (Formation API)
│   ├── chat [--formation <id>] [--profile <name>]         # Interactive chat
│   ├── avchat [--formation <id>] [--profile <name>] --file <path> # Audio/video chat
│   │
│   ├── session
│   │   ├── list [--formation <id>] [--profile <name>] [--user <user_id>]
│   │   ├── get <session_id> [--formation <id>] [--profile <name>]
│   │   ├── messages <session_id> [--formation <id>] [--profile <name>]
│   │   └── delete <session_id> [--formation <id>] [--profile <name>]
│   │
│   ├── trigger
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── get <name> [--formation <id>] [--profile <name>]
│   │   └── invoke <name> [--formation <id>] [--profile <name>] [--data <json>]
│   │
│   ├── job
│   │   ├── list [--formation <id>] [--profile <name>] [--user <user_id>]
│   │   └── cancel <job_id> [--formation <id>] [--profile <name>]
│   │
│   └── request
│       ├── status <request_id> [--formation <id>] [--profile <name>]
│       └── cancel <request_id> [--formation <id>] [--profile <name>]
│
├── 📊 MONITORING & LOGS (Formation API)
│   ├── logs [--formation <id>] [--profile <name>] [--follow]
│   │
│   ├── logging
│   │   ├── config [--formation <id>] [--profile <name>]
│   │   ├── destinations list [--formation <id>] [--profile <name>]
│   │   ├── destinations add [--formation <id>] [--profile <name>]
│   │   ├── destinations update <id> [--formation <id>] [--profile <name>]
│   │   └── destinations delete <id> [--formation <id>] [--profile <name>]
│   │
│   └── audit
│       ├── list [--formation <id>] [--profile <name>]
│       └── clear [--formation <id>] [--profile <name>] --confirm
│
├── 🗓️  SCHEDULER (Formation API)
│   ├── scheduler config [--formation <id>] [--profile <name>]
│   ├── scheduler list [--formation <id>] [--profile <name>]
│   ├── scheduler add [--formation <id>] [--profile <name>]
│   ├── scheduler update <id> [--formation <id>] [--profile <name>]
│   └── scheduler delete <id> [--formation <id>] [--profile <name>]
│
├── 🤝 A2A CONFIGURATION (Formation API)
│   ├── a2a config [--formation <id>] [--profile <name>]
│   └── a2a outbound list [--formation <id>] [--profile <name>]
│
├── 📚 SOPS & KNOWLEDGE (Formation API - Read-only)
│   ├── sop list [--formation <id>] [--profile <name>]
│   └── sop get <name> [--formation <id>] [--profile <name>]
│
└── 👥 USERS (Formation API)
    ├── user identifiers list [--formation <id>] [--profile <name>]
    ├── user identifiers add [--formation <id>] [--profile <name>]
    ├── user get <identifier> [--formation <id>] [--profile <name>]
    └── user delete <identifier> [--formation <id>] [--profile <name>]
```

---

## Command Details

### Global Commands

#### `muxi --version`

**Purpose:** Display CLI version information

**Output:**
```
muxi version 1.0.0
```

**Example:**
```bash
muxi --version
# muxi version 1.0.0
```

---

#### `muxi --help`

**Purpose:** Display help and list all available commands

**Output:**
```
MUXI CLI - Formation development and server management

Usage:
  muxi [command]

Available Commands:
  new         Create formation or components
  validate    Validate formation configuration
  deploy      Deploy formation to server(s)
  secrets     Manage formation secrets
  profile     Manage server profiles
  formation   Formation lifecycle operations
  server      Server management
  agent       Agent management (Formation API)
  mcp         MCP server management (Formation API)
  chat        Interactive chat
  ...

Flags:
  -h, --help      Help for muxi
  -v, --version   Version information

Use "muxi [command] --help" for more information about a command.
```

---

#### `muxi <command> --help`

**Purpose:** Display help for specific command

**Examples:**
```bash
muxi new --help
# Create formation or components
# 
# Usage:
#   muxi new formation [name] [flags]
#   muxi new agent <name> [flags]
#   muxi new mcp <name> [flags]
#   ...

muxi deploy --help
# Deploy formation to server(s)
# 
# Usage:
#   muxi deploy [flags]
# 
# Flags:
#   --profile string   Server profile to deploy to
#   --no-wizard        Skip interactive prompts
```

---

### Server Commands

#### `muxi server version [--profile <name>]`

**Purpose:** Get server version information

**Flags:**
- `--profile <name>` - Server profile (optional, uses default if not specified)

**API Endpoint:** `GET /health` (Server API)

**Output:**
```
Server: production (https://api.company.com:7890)
Version: 1.0.0
Go: go1.21.0
Build: 2025-11-25T10:30:00Z
Uptime: 5d 12h 34m
```

**Examples:**
```bash
# Use default profile
muxi server version
# Using default profile: localhost
# Server: localhost (http://localhost:7890)
# Version: 1.0.0-dev

# Specific profile
muxi server version --profile production
# Server: production (https://api.company.com:7890)
# Version: 1.0.0
# Go: go1.21.0
# Build: 2025-11-25T10:30:00Z

# Using context
muxi profile use production
muxi server version
# Server: production (https://api.company.com:7890)
# Version: 1.0.0
```

**Error Cases:**
```bash
# No profiles configured
muxi server version
# ✗ No server profiles configured
# 
#   Add one with:
#     muxi profile add production

# Server unreachable
muxi server version --profile production
# ✗ Cannot connect to server
# 
#   Server: production (https://api.company.com:7890)
#   Error: connection refused
```

---

### Secret Management Commands

#### `muxi secrets setup`

**Purpose:** Interactive wizard to configure secrets (only prompts for secrets NOT already in secrets.enc)

**Behavior:**
- Scans all YAML files for `${{ secrets.* }}` references
- Only prompts for secrets that don't exist in `secrets.enc`
- Shows which secrets are already configured
- Encrypts new secrets to `secrets.enc`

**Output:**
```
Configuring secrets...

New secrets (3):
  OPENAI_API_KEY: sk-proj-***
  MCP_WEATHER_API_KEY: ***
  DATABASE_PASSWORD: ***

Already configured (5):
  ✓ ANTHROPIC_API_KEY
  ✓ SLACK_TOKEN
  ✓ MCP_POSTGRES_PASSWORD
  ...

✓ 3 new secrets encrypted
```

**Examples:**
```bash
# First time setup
muxi secrets setup
# Configuring secrets...
# New secrets (8):
#   OPENAI_API_KEY: ***
#   ANTHROPIC_API_KEY: ***
#   ...
# ✓ 8 secrets encrypted

# After some secrets are configured
muxi secrets setup
# Configuring secrets...
# New secrets (2):
#   NEW_SERVICE_KEY: ***
#   MCP_WEATHER_API_KEY: ***
# Already configured (6):
#   ✓ OPENAI_API_KEY
#   ✓ SLACK_TOKEN
#   ...
# ✓ 2 new secrets encrypted

# No new secrets needed
muxi secrets setup
# Configuring secrets...
# All secrets configured (8)
# ✓ No new secrets to configure
```

---

#### `muxi secrets sync`

**Purpose:** Sync secrets file with formation YAML files (scan for `${{ secrets.* }}` and update secrets template)

**Behavior:**
- Scans all YAML files for `${{ secrets.* }}` references
- Removes unused secrets from `secrets` file (not referenced in any YAML)
- Adds new secrets found in YAMLs to `secrets` file
- Does NOT modify `secrets.enc` (only updates the `secrets` template)

**Output:**
```
Syncing secrets...

Added 2 new secrets:
  • MCP_WEATHER_API_KEY
  • OPENAI_API_KEY

Removed 1 unused secret:
  • OLD_SERVICE_TOKEN

✓ Secrets synced! Run 'muxi secrets setup' to configure new secrets.
```

**Examples:**
```bash
# After adding new MCP with auth
muxi new mcp weather-api  # Adds MCP_WEATHER_API_KEY to formation
muxi secrets sync
# Syncing secrets...
# Added 1 new secret:
#   • MCP_WEATHER_API_KEY
# ✓ Secrets synced! Run 'muxi secrets setup' to configure.

# After removing an agent that used secrets
# (Removed agents/slack.yaml which used SLACK_TOKEN)
muxi secrets sync
# Syncing secrets...
# Removed 1 unused secret:
#   • SLACK_TOKEN
# ✓ Secrets synced!

# When everything is in sync
muxi secrets sync
# Syncing secrets...
# ✓ Secrets already in sync
```

---

#### `muxi secrets set <key>`

**Purpose:** Set or update a single secret value

**Examples:**
```bash
muxi secrets set OPENAI_API_KEY
# Value: sk-proj-***
# ✓ Secret 'OPENAI_API_KEY' encrypted

# Update existing secret
muxi secrets set DATABASE_PASSWORD
# Current value: ****** (8 characters)
# New value: ***
# ✓ Secret 'DATABASE_PASSWORD' updated
```

---

#### `muxi secrets list`

**Purpose:** List all secret keys (without values)

**Output:**
```bash
muxi secrets list
# Configured secrets (8):
#   • OPENAI_API_KEY
#   • ANTHROPIC_API_KEY
#   • SLACK_TOKEN
#   • DATABASE_PASSWORD
#   • MCP_WEATHER_API_KEY
#   • MCP_POSTGRES_PGHOST
#   • MCP_POSTGRES_PGPORT
#   • MCP_POSTGRES_PGDATABASE
```

---

#### `muxi secrets delete <key>`

**Purpose:** Delete a secret with validation

**Behavior:**
- Warns if secret is still referenced in YAML files
- Requires confirmation before deleting
- Removes from both `secrets` and `secrets.enc`

**Examples:**
```bash
# Delete unused secret
muxi secrets delete OLD_API_KEY
# ⚠ Secret 'OLD_API_KEY' is not referenced in any YAML files.
# Delete? [y/N]: y
# ✓ Secret 'OLD_API_KEY' deleted

# Try to delete referenced secret
muxi secrets delete OPENAI_API_KEY
# ✗ Cannot delete secret 'OPENAI_API_KEY'
# 
#   Referenced in:
#     • formation.yaml (line 15)
#     • agents/assistant.yaml (line 42)
# 
#   Remove these references first, then run:
#     muxi secrets sync
```

---

## ALL Commands Designed ✅

**Total:** ~80 commands covering all API endpoints

**Implementation:** Priority order (6 weeks), but ALL designed now

---

**Status:** Complete command reference
**Next:** Implement in priority order
