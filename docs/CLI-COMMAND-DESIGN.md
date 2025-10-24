# MUXI CLI Command Design Specification

**Version:** 1.0  
**Status:** Draft for Review  
**Last Updated:** 2025-10-24

---

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [Profile Architecture](#profile-architecture)
- [Command Structure](#command-structure)
- [Complete Command Tree](#complete-command-tree)
- [Detailed Command Specifications](#detailed-command-specifications)
- [Common Workflows](#common-workflows)
- [Implementation Notes](#implementation-notes)

---

## Design Philosophy

### Command Structure Pattern

```
muxi [global-flags] <command> <subcommand> [args] [flags]
```

### Global Flags (available on all commands)

- `--profile <name>` - Profile to use (overrides default)
- `--output <format>` - Output format: `text` (default), `json`, `yaml`
- `--no-color` - Disable colored output
- `--debug` - Enable debug logging
- `--help` - Show help for any command

### Key Design Decisions

1. **Flat structure with flags** - `muxi agent list --profile my-bot` (not nested)
   - Shorter commands, easier to type
   - Standard CLI pattern (kubectl, docker, gh, etc.)
   - Better tab completion

2. **Resource-oriented** - Commands map to resources: `agent`, `secret`, `mcp`, `formation`
   - Intuitive CRUD operations
   - Maps cleanly to REST APIs

3. **Profile-aware routing** - Profiles know whether they connect to Server or Formation
   - Server profiles: Use HMAC auth, access formations via `/api/{formation_id}/*`
   - Formation profiles: Use formation auth (API keys), direct access
   - CLI automatically routes commands based on profile type

---

## Profile Architecture

### Two Profile Types

#### 1. Server Profile (connects to MUXI Server)

**Purpose:** Manage multiple formations on a server

**Connection:**
```
CLI → MUXI Server (port 7890)
      ↓
      Server API (/rpc/*)         [Formation lifecycle]
      ↓
      Formation Proxy (/api/*)     [Formation API access]
```

**Authentication:** HMAC-SHA256 (server-level credentials)

**Example profile:**
```yaml
# ~/.muxi/profiles.yaml
profiles:
  production:
    type: server                    # Profile type
    url: https://muxi.company.com:7890
    auth:
      type: hmac
      key_id: MUXI_PROD_KEY
      secret_key: sk_...
    default_formation: my-bot       # Optional: default formation for commands
```

**Supported Operations:**
- Formation lifecycle (deploy, list, stop, restart, delete, etc.)
- Server management (status, logs)
- Formation API access (via proxy) for all formations on the server

#### 2. Formation Profile (connects directly to Formation)

**Purpose:** Interact with a single formation (standalone or remote)

**Connection:**
```
CLI → Formation API (direct)
```

**Authentication:** Formation-specific (admin or client API keys)

**Example profile:**
```yaml
# ~/.muxi/profiles.yaml
profiles:
  my-bot:
    type: formation                 # Profile type
    url: http://localhost:8271      # Direct to formation
    auth:
      type: api_key
      admin_key: fma_...            # For admin operations
      client_key: fmc_...           # For user operations
```

**Supported Operations:**
- Formation API commands only (agent, secret, mcp, chat, etc.)
- No formation lifecycle commands (deploy, stop, etc.)

### Profile Commands Map to Profile Type

| Command | Server Profile | Formation Profile |
|---------|----------------|-------------------|
| `muxi formation deploy` | ✅ Yes | ❌ No |
| `muxi formation list` | ✅ Yes | ❌ No |
| `muxi formation stop` | ✅ Yes | ❌ No |
| `muxi server status` | ✅ Yes | ❌ No |
| `muxi agent list` | ✅ Yes (via proxy) | ✅ Yes (direct) |
| `muxi secret create` | ✅ Yes (via proxy) | ✅ Yes (direct) |
| `muxi chat` | ✅ Yes (via proxy) | ✅ Yes (direct) |

### Auto-Detection on First Run

**Scenario 1: Local MUXI Server detected**
```bash
# CLI detects ~/.muxi/server/credentials.json
→ Auto-creates "localhost" server profile
→ Sets as default
```

**Scenario 2: No server detected**
```bash
# CLI shows setup wizard
→ Offers to:
   1. Install MUXI Server locally
   2. Add remote server profile
   3. Add formation profile (direct connection)
```

---

## Command Structure

### Formation API Commands (work with both profile types)

These commands work with **both** server and formation profiles:

```bash
# With server profile (proxied)
muxi agent list --profile production
→ Connects to: https://server:7890/api/my-bot/v1/agents
→ Auth: HMAC (server credentials)

# With formation profile (direct)
muxi agent list --profile my-bot
→ Connects to: http://localhost:8271/v1/agents
→ Auth: API key (formation admin key)
```

### Server-Only Commands

These commands **only** work with server profiles:

```bash
muxi formation deploy bundle.tar.gz
muxi formation list
muxi server status
```

**Error with formation profile:**
```bash
$ muxi formation list --profile my-bot
Error: Command 'formation list' requires a server profile.
  Profile 'my-bot' is a formation profile (type: formation).
  
Hint: Use a server profile instead:
  muxi formation list --profile production
```

### Context and Defaults

**Server profile with default formation:**
```yaml
profiles:
  production:
    type: server
    url: https://server:7890
    default_formation: my-bot    # Commands default to this formation
```

```bash
# These are equivalent:
muxi agent list --profile production
muxi agent list --profile production --formation my-bot
```

**Formation profile (no formation flag needed):**
```yaml
profiles:
  my-bot:
    type: formation
    url: http://localhost:8271
```

```bash
# Formation is implicit (from profile)
muxi agent list --profile my-bot
```

---

## Complete Command Tree

```
muxi
├── version                    # Show CLI version
├── help                       # Show help
├── whoami                     # Show current profile & registry auth status
│
├── profile                    # Profile management (local config)
│   ├── add <name>            # Add new profile (server or formation)
│   ├── use <name>            # Set default profile
│   ├── list                  # List all profiles
│   ├── remove <name>         # Remove profile
│   └── current               # Show current default profile
│
├── registry                   # Registry operations (push/pull schemas)
│   ├── login <registry>      # Authenticate to registry
│   ├── logout <registry>     # Remove registry credentials
│   ├── push <type> <file>    # Push formation/agent/mcp schema
│   ├── pull <ref>            # Pull schema by reference
│   ├── search <term>         # Search registry
│   ├── list                  # List user's schemas
│   └── delete <ref>          # Delete schema version
│
├── formation                  # Formation lifecycle (Server API only)
│   ├── deploy <bundle>       # Deploy new formation
│   ├── list                  # List formations
│   ├── get <id>              # Get formation details
│   ├── update <id> <bundle>  # Update formation (new version)
│   ├── delete <id>           # Delete formation
│   ├── start <id>            # Start stopped formation
│   ├── stop <id>             # Stop running formation
│   ├── restart <id>          # Restart formation
│   ├── rollback <id>         # Rollback to previous version
│   └── logs <id>             # Stream formation logs
│
├── status                     # Get formation status (Formation API)
│   └── [--profile <name>]    # Server profile requires --formation flag
│
├── config                     # Get formation config (Formation API)
│   └── [--profile <name>]    # Server profile requires --formation flag
│
├── agent                      # Agent management (Formation API)
│   ├── list                  # List agents in formation
│   ├── add                   # Add new agent (API-defined)
│   ├── get <agent_id>        # Get agent details
│   ├── update <agent_id>     # Update agent config
│   └── delete <agent_id>     # Delete agent (API-defined only)
│
├── secret                     # Secret management (Formation API)
│   ├── list                  # List secrets (masked values)
│   ├── create <key>          # Create new secret
│   ├── update <key>          # Update secret value
│   └── delete <key>          # Delete secret
│
├── mcp                        # MCP server management (Formation API)
│   ├── list                  # List MCP servers
│   ├── add                   # Add new MCP server
│   ├── get <server_id>       # Get MCP server details
│   ├── update <server_id>    # Update MCP server config
│   └── delete <server_id>    # Delete MCP server
│
├── chat                       # Interactive chat (Formation API)
│   └── [<formation_id>]      # Formation ID required for server profiles
│
├── session                    # Session management (Formation API)
│   ├── list                  # List user sessions
│   ├── get <session_id>      # Get session details/history
│   └── delete <session_id>   # Delete session
│
├── sop                        # Standard Operating Procedures (Formation API)
│   ├── list                  # List available SOPs
│   └── get <sop_name>        # Get SOP details
│
├── audit                      # Audit logs (Formation API)
│   ├── list                  # Get audit log entries
│   └── clear                 # Clear audit log (requires confirmation)
│
├── server                     # Server management (Server API only)
│   ├── status                # Get server status & statistics
│   └── logs                  # Get server audit logs
│
└── local                      # Local utilities (no API)
    ├── validate <file>       # Validate schema file
    ├── pack <file>           # Pack formation with resolved refs
    ├── list                  # List local cache
    └── prune                 # Prune local cache
```

---

## Detailed Command Specifications

### Profile Management

#### `muxi profile add <name>`

Add a new profile (server or formation type).

**Flags:**
- `--type <type>` - Profile type: `server` or `formation` (auto-detected if not specified)
- `--url <url>` - Server/Formation URL (required)
- `--key-id <id>` - HMAC key ID (for server profiles)
- `--secret-key <key>` - HMAC secret key (for server profiles)
- `--admin-key <key>` - Admin API key (for formation profiles)
- `--client-key <key>` - Client API key (for formation profiles)
- `--default` - Set as default profile
- `--auto-detect` - Auto-detect localhost server credentials
- `--formation <id>` - Default formation ID (for server profiles)

**Examples:**

```bash
# Add server profile (explicit)
muxi profile add production \
  --type server \
  --url https://muxi.company.com:7890 \
  --key-id MUXI_PROD_KEY \
  --secret-key "sk_..."

# Add server profile (auto-detected local)
muxi profile add localhost --auto-detect --default

# Add formation profile
muxi profile add my-bot \
  --type formation \
  --url http://localhost:8271 \
  --admin-key "fma_..." \
  --client-key "fmc_..."

# Add server profile with default formation
muxi profile add staging \
  --url https://staging.company.com:7890 \
  --key-id MUXI_STAGING \
  --secret-key "sk_..." \
  --formation support-bot
```

#### `muxi profile list`

List all configured profiles.

**Output:**
```
NAME         TYPE        URL                              DEFAULT  FORMATION
localhost    server      http://localhost:7890            ✓        my-bot
production   server      https://muxi.company.com:7890             support-bot
my-bot       formation   http://localhost:8271
staging      server      https://staging.company.com:7890
```

#### `muxi profile use <name>`

Set default profile.

```bash
muxi profile use production
# ✓ Default profile set to 'production'
```

#### `muxi whoami`

Show current profile and registry authentication status.

**Output:**
```
Profile: production (server)
  URL: https://muxi.company.com:7890
  Key ID: MUXI_PROD_KEY
  Default Formation: support-bot

Registry Accounts:
  ✓ registry.muxihub.com (user: john@company.com)
  ✗ Not logged in to private.company.com
```

---

### Registry Operations

#### `muxi registry login <registry>`

Authenticate to a registry (interactive token input).

**API Mapping:** Registry API `/auth/login` (TBD)

**Flags:**
- `--token <token>` - Provide token directly (non-interactive)

**Examples:**
```bash
# Interactive (prompts for token)
muxi registry login registry.muxihub.com

# Non-interactive
muxi registry login registry.muxihub.com --token "reg_..."
```

#### `muxi registry push <type> <file>`

Push a schema to the registry.

**API Mapping:** Registry API `/schemas/:type` (TBD)

**Args:**
- `<type>` - Schema type: `formation`, `agent`, `mcp`
- `<file>` - YAML file path

**Flags:**
- `--tag <tag>` - Version tag (default: `latest`)
- `--registry <url>` - Registry URL (default: registry.muxihub.com)
- `--org <org>` - Organization name

**Examples:**
```bash
muxi registry push formation my-bot.yaml --tag 1.0.0
muxi registry push agent summarizer.yaml --tag latest
```

#### `muxi registry pull <ref>`

Pull a schema from registry or GitHub.

**API Mapping:** Registry API `/schemas/:org/:name/:tag` OR GitHub clone (TBD)

**Args:**
- `<ref>` - Schema reference: `org/name:tag` or `github.com/user/repo`

**Flags:**
- `--output <file>` - Output file path (default: `./{name}.yaml`)

**Examples:**
```bash
muxi registry pull myorg/support-bot:1.0.0
muxi registry pull github.com/user/repo
```

---

### Formation Lifecycle (Server API - Server Profiles Only)

#### `muxi formation deploy <bundle>`

Deploy a new formation to the server.

**Profile Requirement:** Server profile only

**API Mapping:** `POST /rpc/formations`

**Args:**
- `<bundle>` - Formation bundle file (.tar.gz)

**Flags:**
- `--id <id>` - Formation ID (default: from bundle metadata)
- `--auto-restart` - Enable auto-restart on crash (default: true)
- `--env KEY=VALUE` - Set environment variables (repeatable)

**Examples:**
```bash
muxi formation deploy my-bot.tar.gz --profile production
muxi formation deploy my-bot.tar.gz --id production-bot --env MODEL=gpt-4
```

#### `muxi formation list`

List all formations on the server.

**Profile Requirement:** Server profile only

**API Mapping:** `GET /rpc/formations`

**Flags:**
- `--status <status>` - Filter by status: `running`, `stopped`, `crashed`

**Output:**
```
ID              STATUS    PORT   VERSION   UPTIME    RESTARTS
chat-api        running   8001   2         3h 24m    0
workflow-engine running   8002   1         1d 5h     2
old-bot         stopped   -      1         -         -
```

#### `muxi formation get <id>`

Get detailed information about a formation.

**Profile Requirement:** Server profile only

**API Mapping:** `GET /rpc/formations/{id}`

**Output (text):**
```
Formation: chat-api
Status: running
Port: 8001
PID: 12345
Version: 2
URL: https://muxi.company.com:7890/api/chat-api
Created: 2025-10-24T10:00:00Z
Uptime: 3h 24m
Restarts: 0
Previous version available: yes
```

#### `muxi formation update <id> <bundle>`

Update formation to new version (keeps previous for rollback).

**Profile Requirement:** Server profile only

**API Mapping:** `PUT /rpc/formations/{id}`

**Examples:**
```bash
muxi formation update my-bot my-bot-v2.tar.gz --profile production
```

#### `muxi formation stop/start/restart <id>`

Control formation lifecycle.

**Profile Requirement:** Server profile only

**API Mapping:** 
- `POST /rpc/formations/{id}/stop`
- `POST /rpc/formations/{id}/restart`

**Examples:**
```bash
muxi formation stop chat-api --profile production
muxi formation restart chat-api --profile production
```

#### `muxi formation rollback <id>`

Rollback formation to previous version.

**Profile Requirement:** Server profile only

**API Mapping:** `POST /rpc/formations/{id}/rollback`

**Examples:**
```bash
muxi formation rollback my-bot --profile production
```

#### `muxi formation delete <id>`

Delete a formation from the server.

**Profile Requirement:** Server profile only

**API Mapping:** `DELETE /rpc/formations/{id}`

**Flags:**
- `--confirm` - Skip confirmation prompt

**Examples:**
```bash
muxi formation delete old-bot --profile production --confirm
```

#### `muxi formation logs <id>`

Stream formation logs (stdout/stderr).

**Profile Requirement:** Server profile only

**API Mapping:** Read from server's log files OR Formation API `GET /v1/logs` (if formation provides it)

**Flags:**
- `--follow, -f` - Stream logs in real-time
- `--lines, -n <num>` - Number of lines (default: 100)
- `--since <duration>` - Show logs since duration (e.g., "5m", "1h")

**Examples:**
```bash
muxi formation logs my-bot --follow --profile production
muxi formation logs my-bot --lines 500 --profile production
```

---

### Formation Status & Config (Formation API - Both Profile Types)

#### `muxi status`

Get formation runtime status.

**Profile Requirements:**
- **Formation profile:** Direct access, no additional flags
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/status`

**Examples:**
```bash
# With formation profile (direct)
muxi status --profile my-bot

# With server profile (proxied)
muxi status --profile production --formation my-bot
```

**Output:**
```
Formation: my-bot
Status: running
Port: 8271
Version: 1.0.0

Agents: 3 total, 2 active
MCP Servers: 2 total, 2 active
Memory Usage: 256.7 MB
CPU: 12.5%
Requests: 100 total, 10 active
Uptime: 3h 24m
```

#### `muxi config`

Get full formation configuration.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/config`

**Examples:**
```bash
# With formation profile
muxi config --profile my-bot --output yaml

# With server profile
muxi config --profile production --formation my-bot --output json
```

---

### Agent Management (Formation API - Both Profile Types)

#### `muxi agent list`

List all agents in formation.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/agents`

**Examples:**
```bash
# With formation profile
muxi agent list --profile my-bot

# With server profile
muxi agent list --profile production --formation my-bot
```

**Output:**
```
ID               NAME              ROLE        ACTIVE  SOURCE
weather-bot      Weather Assistant specialist ✓       formation
code-helper      Code Helper       specialist ✓       api
summarizer       Summarizer        specialist ✗       formation
```

#### `muxi agent add`

Add a new agent via API (interactive wizard).

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `POST /v1/agents`

**Flags:**
- `--file <file>` - Agent definition YAML file (non-interactive)
- `--id <id>` - Agent ID
- `--name <name>` - Agent name
- `--model <model>` - LLM model
- `--system-message <msg>` - System message

**Examples:**
```bash
# Interactive wizard
muxi agent add --profile my-bot

# From file
muxi agent add --profile my-bot --file weather-agent.yaml

# Inline (with server profile)
muxi agent add --profile production --formation my-bot \
  --id weather \
  --name "Weather Bot" \
  --model "openai/gpt-4o" \
  --system-message "You are a weather expert."
```

#### `muxi agent get <agent_id>`

Get agent details.

**API Mapping:** Formation API `GET /v1/agents` (filtered)

**Examples:**
```bash
muxi agent get weather-bot --profile my-bot --output json
```

#### `muxi agent update <agent_id>`

Update agent configuration.

**API Mapping:** Formation API `PATCH /v1/agents/{agent_id}`

**Flags:**
- `--active <bool>` - Enable/disable agent
- `--model <model>` - Update LLM model
- `--system-message <msg>` - Update system message

**Examples:**
```bash
muxi agent update weather-bot --profile my-bot --active false
muxi agent update code-helper --profile my-bot --model "anthropic/claude-3.5-sonnet"
```

#### `muxi agent delete <agent_id>`

Delete an API-defined agent.

**API Mapping:** Formation API `DELETE /v1/agents/{agent_id}`

**Note:** Cannot delete formation-defined agents (from YAML).

**Examples:**
```bash
muxi agent delete old-agent --profile my-bot --confirm
```

---

### Secret Management (Formation API - Both Profile Types)

#### `muxi secret list`

List all secrets (with masked values).

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/secrets`

**Examples:**
```bash
muxi secret list --profile my-bot
muxi secret list --profile production --formation my-bot
```

**Output:**
```
KEY                          VALUE
OPENAI_API_KEY               sk-••••••••••••••••••••••••••••••••Gst
ANTHROPIC_API_KEY            sk-ant-••••••••••••••••••••••••••••
DATABASE_URL                 postgresql://••••••••@localhost:5432/db
```

#### `muxi secret create <key>`

Create a new secret (prompts for value securely).

**API Mapping:** Formation API `POST /v1/secrets`

**Flags:**
- `--value <value>` - Provide value directly (not recommended, visible in shell history)

**Examples:**
```bash
# Interactive (recommended - secure input)
muxi secret create OPENAI_API_KEY --profile my-bot
# Prompts: Enter value: ••••••

# Inline (not recommended)
muxi secret create OPENAI_API_KEY --profile my-bot --value "sk-..."

# With server profile
muxi secret create OPENAI_API_KEY --profile production --formation my-bot
```

#### `muxi secret update <key>`

Update secret value (prompts for new value).

**API Mapping:** Formation API `PUT /v1/secrets/{key}`

**Examples:**
```bash
muxi secret update OPENAI_API_KEY --profile my-bot
```

#### `muxi secret delete <key>`

Delete a secret.

**API Mapping:** Formation API `DELETE /v1/secrets/{key}`

**Flags:**
- `--confirm` - Skip confirmation prompt

**Examples:**
```bash
muxi secret delete OLD_API_KEY --profile my-bot --confirm
```

---

### MCP Server Management (Formation API - Both Profile Types)

#### `muxi mcp list`

List MCP servers in formation.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/mcp/servers`

**Examples:**
```bash
muxi mcp list --profile my-bot
muxi mcp list --profile production --formation my-bot
```

#### `muxi mcp add`

Add new MCP server (interactive wizard or file).

**API Mapping:** Formation API `POST /v1/mcp/servers`

**Flags:**
- `--file <file>` - MCP server definition YAML

**Examples:**
```bash
muxi mcp add --profile my-bot
muxi mcp add --profile my-bot --file postgres-mcp.yaml
```

#### `muxi mcp get <server_id>`

Get MCP server details.

**API Mapping:** Formation API `GET /v1/mcp/servers` (filtered)

#### `muxi mcp update <server_id>`

Update MCP server configuration.

**API Mapping:** Formation API `PATCH /v1/mcp/servers/{server_id}`

#### `muxi mcp delete <server_id>`

Delete MCP server.

**API Mapping:** Formation API `DELETE /v1/mcp/servers/{server_id}`

---

### Chat (Formation API - Both Profile Types)

#### `muxi chat [<formation_id>]`

Start interactive chat session with formation.

**Profile Requirements:**
- **Formation profile:** No formation_id needed (implicit from profile)
- **Server profile:** Requires `<formation_id>` argument

**API Mapping:** Formation API `POST /v1/chat` (streaming)

**Flags:**
- `--user-id <id>` - User ID (default: system user)
- `--session-id <id>` - Continue existing session
- `--output <file>` - Save transcript to file

**Interactive Commands:**
- `/help` - Show help
- `/session` - Show current session ID
- `/new` - Start new session
- `/history` - Show session history
- `/exit` or `/quit` - Exit chat

**Examples:**
```bash
# With formation profile (formation implicit)
muxi chat --profile my-bot

# With server profile (formation explicit)
muxi chat my-bot --profile production

# Continue existing session
muxi chat my-bot --profile production --session-id ses_abc123

# Save transcript
muxi chat --profile my-bot --output transcript.txt
```

**Interactive session:**
```
╭──────────────────────────────────────────────╮
│ MUXI Chat: my-bot                            │
│ Session: ses_abc123                          │
│ Type /help for commands, /exit to quit       │
╰──────────────────────────────────────────────╯

You: What's the weather like today?

Bot: Today's weather is sunny with a high of 72°F and a low of 55°F.
     Winds from the west at 10-15 mph. Perfect day for outdoor activities!

You: Thanks!

Bot: You're welcome! Enjoy the beautiful weather!

You: /exit
Goodbye!
```

---

### Session Management (Formation API - Both Profile Types)

#### `muxi session list`

List user sessions.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/sessions/{user_id}`

**Flags:**
- `--user-id <id>` (default: current user)

**Examples:**
```bash
muxi session list --profile my-bot
muxi session list --profile production --formation my-bot --user-id user-123
```

**Output:**
```
SESSION ID       CREATED              MESSAGES  LAST ACTIVITY
ses_abc123       2025-10-24 10:00:00  15        2 hours ago
ses_xyz789       2025-10-23 14:30:00  42        1 day ago
ses_def456       2025-10-22 09:15:00  8         2 days ago
```

#### `muxi session get <session_id>`

Get session details and history.

**API Mapping:** Formation API `GET /v1/sessions/{user_id}/{session_id}`

**Examples:**
```bash
muxi session get ses_abc123 --profile my-bot
```

**Output:**
```
Session: ses_abc123
Formation: my-bot
User: user-123
Created: 2025-10-24 10:00:00
Messages: 15

History:
1. [10:00:15] User: What's the weather?
2. [10:00:17] Bot: The weather today is sunny...
3. [10:01:22] User: Thanks!
4. [10:01:23] Bot: You're welcome!
...
```

#### `muxi session delete <session_id>`

Delete a session.

**API Mapping:** Formation API `DELETE /v1/sessions/{user_id}/{session_id}`

**Examples:**
```bash
muxi session delete ses_old123 --profile my-bot --confirm
```

---

### Standard Operating Procedures (Formation API - Both Profile Types)

#### `muxi sop list`

List available SOPs.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/sops`

**Examples:**
```bash
muxi sop list --profile my-bot
```

**Output:**
```
NAME                    TITLE                          TYPE      STEPS
customer-onboarding     Customer Onboarding Procedure  template  5
incident-response       Security Incident Response     guide     8
support-escalation      Support Ticket Escalation      template  3
```

#### `muxi sop get <sop_name>`

Get SOP details.

**API Mapping:** Formation API `GET /v1/sops/{sop_name}`

**Examples:**
```bash
muxi sop get customer-onboarding --profile my-bot
muxi sop get incident-response --profile production --formation support-bot --output yaml
```

---

### Audit Logs (Formation API - Both Profile Types)

#### `muxi audit list`

Get audit log entries.

**Profile Requirements:**
- **Formation profile:** Direct access
- **Server profile:** Requires `--formation <id>` flag

**API Mapping:** Formation API `GET /v1/audit`

**Flags:**
- `--limit <num>` - Number of entries (default: 100)
- `--action <action>` - Filter by action (e.g., `agent.created`)
- `--resource-type <type>` - Filter by resource type (agent, secret, mcp_server, etc.)
- `--since <timestamp>` - Entries since timestamp (ISO 8601)

**Examples:**
```bash
muxi audit list --profile my-bot --limit 50
muxi audit list --profile production --formation my-bot --action agent.created
muxi audit list --profile my-bot --resource-type secret --since 2025-10-20T00:00:00Z
```

**Output:**
```
TIMESTAMP            ACTION         RESOURCE      USER   RESULT
2025-10-24 14:23:45  agent.created  weather-bot   admin  success
2025-10-24 14:24:10  agent.deleted  old-bot       admin  success
2025-10-24 14:25:33  secret.created OPENAI_KEY    admin  success
```

#### `muxi audit clear`

Clear the audit log (requires confirmation).

**API Mapping:** Formation API `DELETE /v1/audit?confirm=clear-audit-log`

**Flags:**
- `--confirm` - Skip confirmation prompt

**Examples:**
```bash
muxi audit clear --profile my-bot --confirm
```

---

### Server Management (Server API - Server Profiles Only)

#### `muxi server status`

Get server status and statistics.

**Profile Requirement:** Server profile only

**API Mapping:** `GET /rpc/server/status`

**Examples:**
```bash
muxi server status --profile production
```

**Output:**
```
Server ID: muxi-prod-abc123
Version: 1.0.0
Uptime: 1d 5h 32m

Formations:
  Total: 5
  Running: 4
  Stopped: 1

Port Pool:
  Total: 1000
  Available: 995
  Allocated: 5

Runtime:
  Go Version: go1.21.5
  Goroutines: 42
```

#### `muxi server logs`

Get server audit logs.

**Profile Requirement:** Server profile only

**API Mapping:** `GET /rpc/server/logs`

**Flags:**
- `--lines <num>` - Number of lines (default: 100, max: 10000)

**Examples:**
```bash
muxi server logs --profile production
muxi server logs --profile production --lines 500
```

**Output (JSON lines):**
```
{"time":"2025-10-24T10:30:15Z","level":"info","method":"POST","path":"/rpc/formations"...}
{"time":"2025-10-24T10:31:22Z","level":"info","method":"GET","path":"/rpc/formations"...}
```

---

### Local Utilities (No API)

#### `muxi local validate <file>`

Validate schema file against official specification.

**Examples:**
```bash
muxi local validate my-formation.yaml
muxi local validate agent.yaml
muxi local validate mcp-server.yaml
```

**Output:**
```
✓ Schema validation passed
  Type: formation
  Schema version: 1.0
  ID: my-bot
  Agents: 3
  MCP Servers: 2
```

#### `muxi local pack <file>`

Pack formation with all references resolved (inline agents, MCPs, etc.).

**Output:** Single YAML file with all refs inlined.

**Flags:**
- `--output <file>` - Output file path (default: `{name}-packed.yaml`)

**Examples:**
```bash
muxi local pack my-formation.yaml --output packed.yaml
```

#### `muxi local list`

List local cache contents.

**Examples:**
```bash
muxi local list
```

**Output:**
```
CACHED SCHEMAS:
myorg/support-bot:1.0.0  (formation)  Downloaded: 2025-10-20
myorg/weather-agent:1.2.0 (agent)     Downloaded: 2025-10-21
```

#### `muxi local prune`

Prune/clear local cache.

**Examples:**
```bash
muxi local prune --confirm
```

---

## Common Workflows

### Initial Setup

**Scenario 1: Local server exists**
```bash
# Auto-detect and configure localhost profile
muxi profile add localhost --auto-detect --default

# Verify
muxi whoami
# Output: Profile: localhost (server)
#         URL: http://localhost:7890
#         ...
```

**Scenario 2: Remote server**
```bash
# Add remote server profile
muxi profile add production \
  --type server \
  --url https://muxi.company.com:7890 \
  --key-id MUXI_PROD_KEY \
  --secret-key "sk_..." \
  --default

# Verify server connectivity
muxi server status
```

**Scenario 3: Direct formation connection**
```bash
# Add formation profile (standalone formation)
muxi profile add my-bot \
  --type formation \
  --url http://localhost:8271 \
  --admin-key "fma_..." \
  --client-key "fmc_..." \
  --default

# Check formation status
muxi status
```

---

### Deploy & Manage Formation (Server Profile)

```bash
# Deploy formation
muxi formation deploy my-bot.tar.gz --profile production

# Check deployment
muxi formation list --profile production

# Get formation details
muxi formation get my-bot --profile production

# Check formation runtime status
muxi status --profile production --formation my-bot

# View logs
muxi formation logs my-bot --profile production --follow

# Update to new version
muxi formation update my-bot my-bot-v2.tar.gz --profile production

# Rollback if needed
muxi formation rollback my-bot --profile production

# Restart formation
muxi formation restart my-bot --profile production

# Stop formation
muxi formation stop my-bot --profile production

# Delete formation
muxi formation delete my-bot --profile production --confirm
```

---

### Configure Formation at Runtime

**With server profile (proxied to formation):**
```bash
# Add agent
muxi agent add \
  --profile production \
  --formation my-bot \
  --file weather-agent.yaml

# Create secret
muxi secret create OPENAI_API_KEY \
  --profile production \
  --formation my-bot
# Prompts for value securely

# Add MCP server
muxi mcp add \
  --profile production \
  --formation my-bot \
  --file postgres-mcp.yaml

# List configuration
muxi agent list --profile production --formation my-bot
muxi secret list --profile production --formation my-bot
muxi mcp list --profile production --formation my-bot
```

**With formation profile (direct):**
```bash
# Add agent (no --formation flag needed)
muxi agent add --profile my-bot --file weather-agent.yaml

# Create secret
muxi secret create OPENAI_API_KEY --profile my-bot

# Add MCP server
muxi mcp add --profile my-bot --file postgres-mcp.yaml

# List configuration
muxi agent list --profile my-bot
muxi secret list --profile my-bot
muxi mcp list --profile my-bot
```

---

### Interact with Formation

```bash
# Start interactive chat
muxi chat --profile my-bot

# Or with server profile
muxi chat my-bot --profile production

# Continue existing session
muxi chat my-bot --profile production --session-id ses_abc123

# List sessions
muxi session list --profile my-bot

# View session history
muxi session get ses_abc123 --profile my-bot

# Check available SOPs
muxi sop list --profile my-bot

# View SOP details
muxi sop get customer-onboarding --profile my-bot
```

---

### Monitor & Audit

**Server-level monitoring:**
```bash
# Server status
muxi server status --profile production

# Server audit logs
muxi server logs --profile production --lines 500

# List all formations
muxi formation list --profile production

# Formation logs (stdout/stderr)
muxi formation logs my-bot --profile production --follow
```

**Formation-level monitoring:**
```bash
# Formation runtime status
muxi status --profile production --formation my-bot

# Full configuration
muxi config --profile production --formation my-bot --output yaml

# Audit logs (formation operations)
muxi audit list --profile production --formation my-bot --limit 100

# Filter audit logs
muxi audit list --profile my-bot --action agent.created --since 2025-10-20T00:00:00Z
```

---

### Registry Operations

```bash
# Login to registry
muxi registry login registry.muxihub.com

# Push formation schema
muxi registry push formation my-bot.yaml --tag 1.0.0

# Push agent schema
muxi registry push agent weather-agent.yaml --tag latest

# Search registry
muxi registry search "weather"

# List my schemas
muxi registry list

# Pull schema
muxi registry pull myorg/support-bot:1.0.0

# Pull from GitHub
muxi registry pull github.com/user/formation-repo

# Delete schema version
muxi registry delete myorg/old-bot:0.9.0 --confirm
```

---

### Multi-Profile Management

```bash
# Add multiple profiles
muxi profile add localhost --auto-detect
muxi profile add staging --url https://staging.company.com:7890 --key-id ...
muxi profile add production --url https://prod.company.com:7890 --key-id ...

# Switch between profiles
muxi profile use localhost
muxi formation list

muxi profile use production
muxi formation list

# Or use --profile flag explicitly
muxi formation list --profile staging
muxi formation list --profile production

# Check current profile
muxi whoami
```

---

## Implementation Notes

### URL Routing Logic

The CLI must intelligently route requests based on profile type:

**Server Profile:**
```go
func BuildURL(profile Profile, endpoint string, formation string) string {
    if profile.Type == "server" {
        // Server API endpoints
        if strings.HasPrefix(endpoint, "/rpc/") {
            return profile.URL + endpoint
        }
        
        // Formation API endpoints (proxied)
        if formation == "" && profile.DefaultFormation != "" {
            formation = profile.DefaultFormation
        }
        if formation == "" {
            return "", errors.New("formation ID required for server profiles")
        }
        return profile.URL + "/api/" + formation + endpoint
    }
    
    // Formation profile (direct)
    return profile.URL + endpoint
}

// Examples:
// Server profile + /rpc/formations → http://server:7890/rpc/formations
// Server profile + /v1/agents + formation=my-bot → http://server:7890/api/my-bot/v1/agents
// Formation profile + /v1/agents → http://formation:8271/v1/agents
```

### Authentication Logic

```go
func AddAuthHeaders(req *http.Request, profile Profile) error {
    if profile.Type == "server" {
        // HMAC authentication
        timestamp := time.Now().Unix()
        signature := GenerateHMAC(
            profile.Auth.KeyID,
            profile.Auth.SecretKey,
            req.Method,
            req.URL.Path,
            timestamp,
            req.Body,
        )
        req.Header.Set("X-MUXI-Key-ID", profile.Auth.KeyID)
        req.Header.Set("X-MUXI-Timestamp", fmt.Sprintf("%d", timestamp))
        req.Header.Set("X-MUXI-Signature", signature)
        return nil
    }
    
    if profile.Type == "formation" {
        // Formation API key authentication
        // Use admin key for admin operations, client key for user operations
        if isAdminOperation(req) {
            req.Header.Set("X-MUXI-Admin-Key", profile.Auth.AdminKey)
        } else {
            req.Header.Set("X-MUXI-Client-Key", profile.Auth.ClientKey)
        }
        return nil
    }
    
    return errors.New("unknown profile type")
}
```

### Command Validation

```go
func ValidateCommand(cmd string, profile Profile) error {
    serverOnlyCommands := []string{
        "formation deploy",
        "formation list",
        "formation stop",
        "formation restart",
        "formation rollback",
        "formation delete",
        "server status",
        "server logs",
    }
    
    if profile.Type != "server" && contains(serverOnlyCommands, cmd) {
        return fmt.Errorf(
            "Command '%s' requires a server profile.\n" +
            "  Profile '%s' is a formation profile (type: formation).\n\n" +
            "Hint: Use a server profile instead:\n" +
            "  %s --profile <server-profile>",
            cmd, profile.Name, cmd,
        )
    }
    
    return nil
}
```

### Auto-Detection Logic

```go
func AutoDetectLocalServer() (*Profile, error) {
    // Check for credentials file
    homeDir, _ := os.UserHomeDir()
    credsPath := filepath.Join(homeDir, ".muxi/server/credentials.json")
    
    if !fileExists(credsPath) {
        return nil, errors.New("no local server credentials found")
    }
    
    // Read credentials
    creds, err := readServerCredentials(credsPath)
    if err != nil {
        return nil, err
    }
    
    // Check if server is running
    resp, err := http.Get("http://localhost:7890/health")
    if err != nil || resp.StatusCode != 200 {
        return nil, errors.New("local server not running")
    }
    
    // Create profile
    profile := &Profile{
        Name: "localhost",
        Type: "server",
        URL:  "http://localhost:7890",
        Auth: Auth{
            Type:      "hmac",
            KeyID:     creds.KeyID,
            SecretKey: creds.SecretKey,
        },
    }
    
    return profile, nil
}
```

### HMAC Signature Generation

```go
func GenerateHMAC(keyID, secretKey, method, path string, timestamp int64, body []byte) string {
    // Canonical string: timestamp;method;path;body_hash
    bodyHash := ""
    if body != nil && len(body) > 0 {
        h := sha256.Sum256(body)
        bodyHash = hex.EncodeToString(h[:])
    }
    
    canonical := fmt.Sprintf("%d;%s;%s;%s", timestamp, method, path, bodyHash)
    
    // HMAC-SHA256
    mac := hmac.New(sha256.New, []byte(secretKey))
    mac.Write([]byte(canonical))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    return signature
}
```

### Profile Storage Format

**File:** `~/.muxi/profiles.yaml`

```yaml
version: "1.0"
default_profile: localhost

profiles:
  localhost:
    type: server
    url: http://localhost:7890
    auth:
      type: hmac
      key_id: MUXI_LOCAL_abc123
      secret_key: sk_...
    default_formation: my-bot
  
  production:
    type: server
    url: https://muxi.company.com:7890
    auth:
      type: hmac
      key_id: MUXI_PROD_KEY
      secret_key: sk_...
    default_formation: support-bot
  
  my-bot:
    type: formation
    url: http://localhost:8271
    auth:
      type: api_key
      admin_key: fma_...
      client_key: fmc_...

registries:
  registry.muxihub.com:
    token: reg_...
    user: john@company.com
    created_at: "2025-10-24T10:00:00Z"
```

### Error Handling Examples

**Server profile required:**
```bash
$ muxi formation list --profile my-bot

Error: Command 'formation list' requires a server profile.
  Profile 'my-bot' is a formation profile (type: formation).

Hint: Use a server profile instead:
  muxi formation list --profile <server-profile>

Available server profiles:
  - localhost (http://localhost:7890)
  - production (https://muxi.company.com:7890)
```

**Formation ID required:**
```bash
$ muxi agent list --profile production

Error: Formation ID required for server profiles.

Hint: Specify --formation flag:
  muxi agent list --profile production --formation <formation-id>

Or set default formation in profile:
  muxi profile add production --formation my-bot
```

**Invalid profile:**
```bash
$ muxi status --profile unknown

Error: Profile 'unknown' not found.

Available profiles:
  - localhost (server) [default]
  - production (server)
  - my-bot (formation)

Hint: Create a new profile:
  muxi profile add unknown --url <url> ...
```

---

## Output Formatting

### Text Output (Default)

Human-readable, colorized output:

```
✓ Formation deployed successfully
  ID: my-bot
  Status: running
  Port: 8001
  URL: http://localhost:7890/api/my-bot
```

### JSON Output

Machine-readable JSON for scripting:

```bash
muxi formation get my-bot --output json
```

```json
{
  "id": "my-bot",
  "status": "running",
  "port": 8001,
  "version": 2,
  "created_at": "2025-10-24T10:00:00Z"
}
```

### YAML Output

Configuration-friendly YAML:

```bash
muxi config --profile my-bot --output yaml
```

```yaml
formation_id: my-bot
version: "1.0.0"
agents:
  total: 3
  resource: /v1/agents
secrets:
  total: 5
  resource: /v1/secrets
```

---

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Command-line usage error
- `3` - Authentication error
- `4` - Resource not found
- `5` - Server error
- `130` - Interrupted by user (Ctrl+C)

---

## Shell Completion

The CLI should provide shell completion for:
- Commands and subcommands
- Profile names
- Formation IDs
- Agent IDs
- Secret keys
- MCP server IDs

**Installation:**
```bash
# Bash
muxi completion bash > /etc/bash_completion.d/muxi

# Zsh
muxi completion zsh > /usr/local/share/zsh/site-functions/_muxi

# Fish
muxi completion fish > ~/.config/fish/completions/muxi.fish
```

---

## Environment Variables

### CLI Configuration

- `MUXI_PROFILE` - Default profile name
- `MUXI_CONFIG_DIR` - Config directory (default: `~/.muxi`)
- `MUXI_OUTPUT_FORMAT` - Default output format: `text`, `json`, `yaml`
- `MUXI_NO_COLOR` - Disable colors (set to `1`)
- `MUXI_DEBUG` - Enable debug logging (set to `1`)

**Examples:**
```bash
export MUXI_PROFILE=production
muxi formation list  # Uses production profile

export MUXI_OUTPUT_FORMAT=json
muxi agent list --profile my-bot  # Outputs JSON

export MUXI_NO_COLOR=1
muxi status --profile my-bot  # No colors
```

---

## API Endpoint Mapping Reference

### Server API (`/rpc/*`) - Server Profiles Only

| CLI Command | HTTP Method | API Endpoint |
|-------------|-------------|--------------|
| `muxi formation deploy` | POST | `/rpc/formations` |
| `muxi formation list` | GET | `/rpc/formations` |
| `muxi formation get <id>` | GET | `/rpc/formations/{id}` |
| `muxi formation update <id>` | PUT | `/rpc/formations/{id}` |
| `muxi formation delete <id>` | DELETE | `/rpc/formations/{id}` |
| `muxi formation stop <id>` | POST | `/rpc/formations/{id}/stop` |
| `muxi formation restart <id>` | POST | `/rpc/formations/{id}/restart` |
| `muxi formation rollback <id>` | POST | `/rpc/formations/{id}/rollback` |
| `muxi server status` | GET | `/rpc/server/status` |
| `muxi server logs` | GET | `/rpc/server/logs` |

### Formation API (`/v1/*`) - Both Profile Types

**Direct (Formation Profile):** `http://formation:8271/v1/*`  
**Proxied (Server Profile):** `http://server:7890/api/{formation_id}/v1/*`

| CLI Command | HTTP Method | API Endpoint |
|-------------|-------------|--------------|
| `muxi status` | GET | `/v1/status` |
| `muxi config` | GET | `/v1/config` |
| `muxi agent list` | GET | `/v1/agents` |
| `muxi agent add` | POST | `/v1/agents` |
| `muxi agent get <id>` | GET | `/v1/agents` |
| `muxi agent update <id>` | PATCH | `/v1/agents/{agent_id}` |
| `muxi agent delete <id>` | DELETE | `/v1/agents/{agent_id}` |
| `muxi secret list` | GET | `/v1/secrets` |
| `muxi secret create <key>` | POST | `/v1/secrets` |
| `muxi secret update <key>` | PUT | `/v1/secrets/{key}` |
| `muxi secret delete <key>` | DELETE | `/v1/secrets/{key}` |
| `muxi mcp list` | GET | `/v1/mcp/servers` |
| `muxi mcp add` | POST | `/v1/mcp/servers` |
| `muxi mcp get <id>` | GET | `/v1/mcp/servers` |
| `muxi mcp update <id>` | PATCH | `/v1/mcp/servers/{server_id}` |
| `muxi mcp delete <id>` | DELETE | `/v1/mcp/servers/{server_id}` |
| `muxi chat` | POST | `/v1/chat` (streaming) |
| `muxi session list` | GET | `/v1/sessions/{user_id}` |
| `muxi session get <id>` | GET | `/v1/sessions/{user_id}/{session_id}` |
| `muxi session delete <id>` | DELETE | `/v1/sessions/{user_id}/{session_id}` |
| `muxi sop list` | GET | `/v1/sops` |
| `muxi sop get <name>` | GET | `/v1/sops/{sop_name}` |
| `muxi audit list` | GET | `/v1/audit` |
| `muxi audit clear` | DELETE | `/v1/audit?confirm=clear-audit-log` |

---

## Summary

This CLI design provides:

✅ **Two profile types** - Server and Formation, with intelligent routing  
✅ **Complete Server API coverage** - Formation lifecycle management  
✅ **Complete Formation API coverage** - Runtime configuration and interaction  
✅ **Registry operations** - Push/pull schemas (based on PRD)  
✅ **Profile auto-detection** - Seamless localhost setup  
✅ **Consistent UX** - Flat command structure with intuitive flags  
✅ **Flexible authentication** - HMAC for servers, API keys for formations  
✅ **Error handling** - Clear messages with helpful hints  
✅ **Multiple output formats** - Text, JSON, YAML  
✅ **Shell completion** - Bash, Zsh, Fish support  

---

## Next Steps

1. **Review and approve** this specification
2. **Implement profile management** (`profile add/list/use/remove`)
3. **Implement HMAC client** with URL routing logic
4. **Implement formation lifecycle commands** (Server API)
5. **Implement formation configuration commands** (Formation API)
6. **Add interactive chat** with streaming support
7. **Implement registry operations** (once registry API is ready)
8. **Polish UX** (colors, progress indicators, error messages)
9. **Add shell completions** (bash/zsh/fish)
10. **Write documentation** (README, man pages, etc.)

---

**Version:** 1.0  
**Status:** Draft for Review  
**Last Updated:** 2025-10-24
