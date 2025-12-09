# Formation API Commands Implementation Plan

**Date:** 2025-12-08
**Status:** Complete (Phases 1-5)
**Priority:** HIGH
**API Spec:** `../schemas/api/formation-api-v1-final.yaml`

---

## Overview

Implement CLI commands for interacting with **deployed formations** via the Formation API. This is distinct from the Server API:

| API | Purpose | Auth | Base URL |
|-----|---------|------|----------|
| Server API | Lifecycle management (deploy, stop, start) | HMAC | `/rpc/formations/*` |
| Formation API | Runtime interaction (chat, agents, sessions) | API Keys | `/api/{id}/v1/*` |

**Authentication:**
- Admin Key (`X-MUXI-ADMIN-KEY`) - For management endpoints
- Client Key (`X-MUXI-CLIENT-KEY`) - For user interaction endpoints

---

## Command Design Principles

1. **No `muxi api` prefix** - Commands should feel native
2. **Shortcut from formation dir** - Like existing shortcuts (e.g., `muxi stop` = `muxi formation stop`)
3. **Formation ID flag** - `--formation <id>` or `-F <id>` for anywhere
4. **Server/Profile flag** - `--profile <name>` or `-p <name>` for testing same formation on different servers
5. **User ID** - `--user <id>` or `-u <id>`, with default from config
6. **`--remote` flag** - Distinguishes local vs runtime operations for conflicting commands

---

## Global Configuration

### Unified Defaults Command (Implemented)

All defaults use the same pattern: `muxi set default <item>`

```yaml
# ~/.muxi/cli/defaults.yaml (global user_id)
version: "1.0"
user_id: "alice"

# ~/.muxi/cli/servers.yaml (global server default)
default: "production"
servers: {...}

# ~/.muxi/cli/registries.yaml (global registry default)
default_registry: "registry.muxi.org"
registries: {...}

# .muxi (formation-level overrides)
profile: "local"
registry: "muxihub"
user_id: "dev-user"
```

**Commands:**
```bash
muxi set default server [name]     # Set default server
muxi set default registry [name]   # Set default registry  
muxi set default user [user_id]    # Set default user ID

# Flags
--local, -l    # Save to .muxi (this formation only)
--global, -g   # Save to ~/.muxi/cli/ (all formations)
```

**Behavior:**
- Outside formation dir: always sets global (no prompt)
- Inside formation dir: prompts "Local or Global?" 
- `--local` / `--global` flags bypass prompt

**Resolution order (for each setting):**
1. Command flag (e.g., `--user`, `--profile`)
2. `.muxi` in formation dir (local)
3. `~/.muxi/cli/` config files (global)
4. Prompt if none set

**Helper function:** `defaults.GetEffectiveUserID(formationUserID)` checks formation then global.

---

## Conflict Resolution: `--remote` Flag

For commands with both local and remote variants, use `--remote` flag:

```go
func runSecrets(cmd *cobra.Command, args []string) error {
    remote, _ := cmd.Flags().GetBool("remote")

    if remote {
        return runRemoteSecrets(cmd, args)  // Formation API
    }
    return runLocalSecrets(cmd, args)  // Local secrets.enc
}
```

| Command | Default (local) | `--remote` flag |
|---------|----------------|-----------------|
| `muxi secrets list` | List secrets.enc | GET /secrets |
| `muxi secrets set KEY` | Update secrets.enc | POST/PUT /secrets |
| `muxi secrets delete KEY` | Remove from secrets.enc | DELETE /secrets/{key} |
| `muxi config` | Interactive wizards | GET /config |
| `muxi config llm` | LLM wizard | GET /llm/settings |
| `muxi config memory` | Memory wizard | GET /memory |
| `muxi config overlord` | Overlord wizard | GET /overlord |

**Future-proof:** If we add local `muxi agents` (list from `agents/*.yaml`), same pattern applies.

---

## Implementation Phases

### Phase 1: Foundation & Introspection
1. `muxi info` - Formation runtime status
2. `muxi agents` - List/manage agents
3. `muxi mcp` - List MCP servers
4. `muxi triggers` - List triggers
5. `muxi sops` - List SOPs

### Phase 2: Configuration (--remote flag)
6. `muxi secrets --remote` - Runtime secrets
7. `muxi config --remote` - Runtime config dump

### Phase 3: Session Management
8. `muxi sessions` - List user sessions
9. `muxi history` - Get session messages
10. `muxi clear` - Clear session/buffer

### Phase 4: Operations
11. `muxi trigger` - Execute trigger
12. `muxi jobs` - List/cancel async jobs
13. `muxi audit` - View audit log
14. `muxi stream` - Live log streaming

### Phase 5: Advanced Admin
15. `muxi scheduler` - Scheduler jobs
16. `muxi users` - User identity management
17. `muxi memory` - Memory operations

### Phase 6: Chat (Last - needs careful UX design)
18. `muxi chat` - Interactive chat with formation

---

## Phase 1: Foundation & Introspection

### `muxi info`

**Get formation runtime status.**

```bash
muxi info                           # From formation dir
muxi info -F my-formation           # From anywhere
muxi info -F my-formation -p prod   # Specific server profile
muxi info --full                    # Include full config
```

**API:** `GET /status`, `GET /config`

**Auth:** Admin Key

**Output:**
```
Formation: my-formation

  Status:     ● running
  Version:    1.2.0
  Uptime:     5d 12h 30m

  Agents:     3 (2 active)
  MCP:        2 servers connected
  Memory:     256 MB working, 1.2 GB persistent

  Stats:
    Requests:   1,234 total (12 active)
    CPU:        15%
    Buffer:     850 messages

  URL: http://localhost:7890/api/my-formation
```

---

### `muxi agents`

**List and manage agents in running formation.**

```bash
muxi agents                    # List agents
muxi agents -v                 # Verbose
muxi agents add <config.yaml>  # Add agent (POST /agents)
muxi agents remove <id>        # Remove agent (DELETE /agents/{id})
muxi agents disable <id>       # Disable agent (PATCH /agents/{id})
muxi agents enable <id>        # Enable agent
```

**API:** `GET/POST/PATCH/DELETE /agents`

**Auth:** Admin Key

**Output:**
```
  ID                  ROLE        STATUS    MODEL
  weather-assistant   specialist  ● active  openai/gpt-4o
  code-helper         specialist  ● active  anthropic/claude-3
  general-assistant   generalist  ○ idle    openai/gpt-4o-mini
```

---

### `muxi mcp`

**List MCP servers in running formation.**

```bash
muxi mcp                       # List servers
muxi mcp -v                    # Verbose with tools
muxi mcp add <config.yaml>     # Add server
muxi mcp remove <id>           # Remove server
```

**API:** `GET/POST/PATCH/DELETE /mcp/servers`

**Auth:** Admin Key

**Output:**
```
  ID              TYPE      STATUS       TOOLS
  local-tools     command   ● connected  5 tools
  web-search      http      ● connected  3 tools
  database        sse       ○ error      -
```

---

### `muxi triggers`

**List available triggers.**

```bash
muxi triggers
muxi triggers -F my-formation
```

**API:** `GET /triggers`

**Auth:** Client Key

**Output:**
```
  NAME                 DESCRIPTION
  github-issue         Handle GitHub issue events
  linear-ticket        Process Linear ticket updates
  deployment-notify    Deployment notification handler
```

---

### `muxi sops`

**List Standard Operating Procedures.**

```bash
muxi sops                      # List SOPs
muxi sops show <name>          # Show SOP details
```

**API:** `GET /sops`, `GET /sops/{name}`

**Auth:** Client Key

**Output:**
```
  NAME                 TYPE       STEPS  AGENTS
  customer-onboarding  template   5      identity-verifier, account-manager
  incident-response    guide      8      security-analyst, incident-coordinator
```

---

## Phase 2: Configuration (--remote flag)

### `muxi secrets --remote`

**Manage runtime secrets.**

```bash
# Local (default, existing behavior)
muxi secrets list
muxi secrets set OPENAI_KEY
muxi secrets delete OLD_KEY

# Remote (Formation API)
muxi secrets list --remote
muxi secrets set NEW_KEY --remote
muxi secrets delete OLD_KEY --remote

# From anywhere (always remote)
muxi formation secrets list -F my-formation
muxi formation secrets set KEY -F my-formation
```

**API:** `GET/POST/PUT/DELETE /secrets`

**Auth:** Admin Key

**Output (--remote):**
```
  KEY                    VALUE
  OPENAI_API_KEY         sk-••••••••••••••••Gst
  ANTHROPIC_API_KEY      sk-ant-••••••••••••••
  DATABASE_URL           postgresql://••••@localhost
```

---

### `muxi config --remote`

**View runtime configuration.**

```bash
# Local (default, existing behavior - interactive wizards)
muxi config llm
muxi config memory

# Remote (Formation API - read-only dump)
muxi config --remote              # Full config
muxi config llm --remote          # LLM settings
muxi config memory --remote       # Memory settings
muxi config overlord --remote     # Overlord config
muxi config async --remote        # Async settings
muxi config a2a --remote          # A2A settings
muxi config logging --remote      # Logging config
```

**API:** `GET /config`, `GET /llm/settings`, `GET /memory`, etc.

**Auth:** Admin Key

---

## Phase 3: Session Management ✅ COMPLETE

*See [docs/guides/sessions.md](guides/sessions.md) for full usage guide.*

### `muxi sessions`

**List user sessions.**

```bash
muxi sessions                    # Uses default user
muxi sessions -u alice           # Specific user
muxi sessions -u alice --active  # Active only
```

**API:** `GET /sessions`

**Auth:** Client Key + `X-User-ID`

**Output:**
```
  SESSION ID      MESSAGES   LAST ACTIVITY      STATUS
  sess_abc123     25         2 minutes ago      ● active
  sess_xyz789     12         3 hours ago        ○ inactive
```

---

### `muxi history`

**Get session message history.**

```bash
muxi history -s sess_abc123           # Uses default user
muxi history -s sess_abc123 -u alice  # Specific user
muxi history -s sess_abc123 --lines 50
muxi history -s sess_abc123 --json
```

**API:** `GET /sessions/{session_id}/messages`

**Auth:** Client Key + `X-User-ID`

**Output:**
```
Session: sess_abc123 (25 messages)

[10:00:15] alice: What's the weather like?
[10:00:18] weather-assistant: The weather today is sunny with a high of 72°F.

[10:05:00] alice: Can you help me with Python code?
[10:05:03] code-helper: Of course! What do you need help with?
```

---

### `muxi clear`

**Clear session or buffer memory.**

```bash
muxi clear -s sess_abc123        # Clear session
muxi clear --all                 # Clear all user buffer
muxi clear -s sess_abc123 -f     # Skip confirmation
```

**API:** `DELETE /sessions/{session_id}`, `DELETE /memory/buffer`

**Auth:** Client Key + `X-User-ID`

---

## Phase 4: Operations

### `muxi trigger`

**Execute a trigger.**

```bash
muxi trigger github-issue --data '{"issue": {"number": 123}}'
muxi trigger github-issue --file event.json
muxi trigger github-issue --async   # Don't wait for result
```

**API:** `POST /triggers/{trigger_name}`

**Auth:** Client Key

---

### `muxi jobs`

**List and manage async jobs.**

```bash
muxi jobs                        # List jobs for default user
muxi jobs -u webhook-bot         # List jobs for user
muxi jobs cancel <job_id>        # Cancel job
```

**API:** `GET /jobs/{user_id}`, `DELETE /jobs/{user_id}/{job_id}`

**Auth:** Client Key

**Output:**
```
  JOB ID        STATUS       PROGRESS   CREATED
  job_456789    processing   75%        2 minutes ago
  job_123456    completed    100%       1 hour ago
```

---

### `muxi audit`

**View formation audit log.**

```bash
muxi audit
muxi audit --lines 50
muxi audit --action agent.created
muxi audit --since "2025-12-01"
muxi audit --clear               # Clear audit log
```

**API:** `GET /audit`, `DELETE /audit`

**Auth:** Admin Key

**Output:**
```
  TIMESTAMP            ACTION           RESOURCE         USER
  2025-12-06 10:15:00  agent.created    weather-bot      admin
  2025-12-06 10:14:30  secret.created   OPENAI_KEY       admin
  2025-12-06 09:00:00  scheduler.job    daily-report     system
```

---

### `muxi stream`

**Stream live logs from formation.**

```bash
muxi stream -u alice              # Filter by user
muxi stream --level ERROR         # Filter by level
muxi stream --agent weather-bot   # Filter by agent
muxi stream --request req_abc123  # Filter by request
```

**API:** `GET /logs/stream` (SSE)

**Auth:** Admin Key

**Output:**
```
Streaming logs (Ctrl+C to stop)

[10:15:00] INFO  chat.started     user=alice session=sess_abc
[10:15:01] INFO  agent.invoked    agent=weather-assistant
[10:15:02] INFO  llm.request      model=gpt-4o tokens=150
[10:15:03] INFO  chat.completed   user=alice duration=2.1s
```

---

## Phase 5: Advanced Admin

### `muxi scheduler`

**Manage scheduled jobs.**

```bash
muxi scheduler                   # List scheduled jobs
muxi scheduler add <config.yaml> # Create job
muxi scheduler remove <id>       # Remove job
muxi scheduler show <id>         # Show job details
```

**API:** `GET/POST/DELETE /scheduler/jobs`

**Auth:** Admin Key

---

### `muxi users`

**Manage user identities.**

```bash
muxi users identifiers -u alice           # List user's identifiers
muxi users link -u alice "alice@co.com"   # Link identifier
muxi users unlink "alice@co.com"          # Unlink identifier
muxi users resolve "alice@co.com"         # Resolve to MUXI user
```

**API:** `GET/POST/DELETE /users/identifiers`

**Auth:** Client Key

---

### `muxi memory`

**Memory operations.**

```bash
muxi memory status               # Buffer status
muxi memory list                 # List user memories
muxi memory add <content>        # Add memory
muxi memory delete <id>          # Delete memory
```

**API:** `GET /memory/buffer`, `GET/POST/DELETE /memories`

**Auth:** Client Key + `X-User-ID`

---

## Phase 6: Chat (Last)

### `muxi chat`

**Interactive chat with formation. (Requires careful UX design)**

```bash
# Basic usage
muxi chat "What's the weather?"

# With context
muxi chat -u alice "Help me with my account"
muxi chat -u alice -s sess_abc123 "Continue our conversation"

# Options
muxi chat --no-stream "Quick question"
muxi chat --group analyst "Generate Q1 report"

# Pipe input
echo "Analyze this" | muxi chat
cat doc.txt | muxi chat "Summarize:"

# From anywhere
muxi chat -F my-formation -p prod "Hello"
```

**API:** `POST /chat`

**Auth:** Client Key + `X-User-ID`

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID |
| `--session` | `-s` | Session ID |
| `--group` | `-g` | Group ID for routing |
| `--no-stream` | | Disable streaming |
| `--webhook` | | Webhook URL for async |

---

## Phase 6: Chat (Detailed Design)

### Overview

`muxi chat` provides an interactive chat interface with a formation using Bubble Tea for input and SSE streaming for responses.

### Modes

**Interactive Mode (default):**
```bash
muxi chat                          # New session
muxi chat -s sess_abc123           # Resume session
muxi chat --no-splash              # Skip welcome banner
```

**One-shot Mode (for scripting):**
```bash
muxi chat "What's the weather?"              # Single message, stream response, exit
muxi chat --no-stream "Quick question"       # Wait for full response
echo "Analyze this" | muxi chat              # Pipe input
cat doc.txt | muxi chat "Summarize:"         # Pipe with prompt
```

### Interactive UI

```
$ muxi chat

╭── MUXI Chat ────────────────────────────────────────────────╮
│               │                                             │
│  ███╗   ███╗  │ Chatting with:                              │
│  ████╗ ████║  │  ⌬ Formation: my-formation                  │
│  ██║╚██╔╝██║  │  ⏍ Server: local                            │
│  ██║ ╚═╝ ██║  │  ♕ User: alice                              │
│  ╚═╝     ╚═╝  │                                             │
╰─────────────────────────────────────────────────────────────╯

>  What's the weather in NYC?

⠧  Thinking... 2.1s  (ESC to stop)

𝐌  The weather in New York City is currently sunny with a 
   temperature of 72°F (22°C). Expected high of 78°F today
   with clear skies throughout the afternoon.

────────────────────────────────────────────────────────────────
>  type your message here...                                    
   (auto-grows from 1 line to 50% screen height)
────────────────────────────────────────────────────────────────
alice@local://my-formation              ? help  / commands  ESC
```

**UI Elements:**
- **Header:** ASCII "M" logo with formation/server/user info (brand recognition for screenshots)
- **User messages:** `>` prefix
- **Thinking state:** Spinner with elapsed time, ESC to cancel
- **Response prefix:** `𝐌` (Unicode bold M)
- **Input area:** Borderless (no `│` sides) for easy copy, auto-grows to ~20 lines / 50% screen
- **Status bar:** `user@server://formation` URI notation + keyboard hints (dimmed)

### Implementation Approach: Hybrid TUI

Uses Bubble Tea for input box only, streams responses to stdout above:

1. **Input Component** (Bubble Tea)
   - Multiline support (for pasting code)
   - Up/down arrow for message history
   - Tab completion for `/commands`
   - Enter to send, Shift+Enter for newline

2. **Response Streaming** (stdout)
   - SSE streaming from `POST /chat`
   - Real-time token display
   - Markdown rendering via `glamour` (code blocks, lists, links, etc.)

3. **State Management**
   - Session ID persistence
   - Message history (local cache)
   - Connection status

### Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/exit`, `/quit` | Exit chat |
| `/clear` | Clear screen |
| `/session` | Show current session ID |
| `/sessions` | List all sessions |
| `/switch <id>` | Switch to different session |
| `/new` | Start new session |
| `/history` | Show recent messages |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | New line |
| `Up/Down` | Browse message history |
| `Ctrl+C` | Cancel current response / Exit |
| `Ctrl+L` | Clear screen |
| `Ctrl+D` | Exit (EOF) |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID |
| `--session` | `-s` | Resume session ID |
| `--group` | `-g` | Agent group for routing |
| `--no-stream` | | Disable streaming (wait for full response) |
| `--no-splash` | | Skip welcome banner |

### File Structure

```
cmd/
└── chat.go           # Main command, mode detection

pkg/
└── chat/
    ├── chat.go       # Interactive chat loop
    ├── input.go      # Bubble Tea input component
    ├── stream.go     # SSE response streaming
    ├── commands.go   # Slash command handlers
    └── history.go    # Local message history cache
```

### Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Input component
- `github.com/charmbracelet/lipgloss` - Styling (already used)
- `github.com/charmbracelet/glamour` - Markdown rendering with syntax highlighting

### API

**Endpoint:** `POST /chat`

**Auth:** Client Key + `X-User-ID`

**Request:**
```json
{
  "message": "What's the weather?",
  "user_id": "alice",
  "session_id": "sess_abc123",
  "stream": true
}
```

**Response (SSE):**
```
event: chat_token
data: {"token": "The"}

event: chat_token
data: {"token": " weather"}

event: chat_done
data: {"session_id": "sess_abc123", "message_id": "msg_xyz"}
```

---

## API Coverage Summary

| Endpoint | Command | Phase |
|----------|---------|-------|
| `GET /health` | (internal) | - |
| `GET /status` | `muxi info` | 1 |
| `GET /config` | `muxi config --remote` | 2 |
| `GET /overlord` | `muxi config overlord --remote` | 2 |
| `GET/POST/PATCH/DELETE /agents` | `muxi agents` | 1 |
| `GET/POST/PUT/DELETE /secrets` | `muxi secrets --remote` | 2 |
| `GET/PATCH /llm/settings` | `muxi config llm --remote` | 2 |
| `GET/DELETE /audit` | `muxi audit` | 4 |
| `GET /sops` | `muxi sops` | 1 |
| `POST /chat` | `muxi chat` | 6 |
| `POST /avchat` | `muxi chat --av` (future) | - |
| `GET /events/{user_id}` | (internal/future) | - |
| `GET/DELETE /jobs/{user_id}` | `muxi jobs` | 4 |
| `GET/DELETE /requests/{id}` | `muxi jobs` (combined) | 4 |
| `GET /triggers` | `muxi triggers` | 1 |
| `POST /triggers/{name}` | `muxi trigger` | 4 |
| `GET/POST/PATCH/DELETE /logging` | `muxi config logging --remote` | 2 |
| `GET /logs/stream` | `muxi stream` | 4 |
| `GET/PATCH /memory` | `muxi config memory --remote` | 2 |
| `GET/DELETE /memory/buffer` | `muxi clear` | 3 |
| `GET/POST/DELETE /memories` | `muxi memory` | 5 |
| `GET/PATCH /async` | `muxi config async --remote` | 2 |
| `GET/PATCH /scheduler` | `muxi scheduler` | 5 |
| `GET/POST/DELETE /scheduler/jobs` | `muxi scheduler` | 5 |
| `GET/PATCH /a2a` | `muxi config a2a --remote` | 2 |
| `GET /sessions` | `muxi sessions` | 3 |
| `GET/DELETE /sessions/{id}` | `muxi sessions`, `muxi clear` | 3 |
| `GET /sessions/{id}/messages` | `muxi history` | 3 |
| `GET/POST/DELETE /users/identifiers` | `muxi users` | 5 |
| `GET/PATCH /mcp` | `muxi mcp` | 1 |
| `GET/POST/PATCH/DELETE /mcp/servers` | `muxi mcp` | 1 |

---

## File Structure

```
pkg/
├── formation/
│   ├── client.go       # Formation API client
│   ├── auth.go         # API key management
│   ├── types.go        # API response types
│   └── cache.go        # Key caching
│
└── config/
    └── defaults.go     # Default user_id, etc.

cmd/
├── info.go             # muxi info
├── agents.go           # muxi agents
├── mcp.go              # muxi mcp
├── triggers.go         # muxi triggers, muxi trigger
├── sops.go             # muxi sops
├── secrets.go          # Update: add --remote support
├── config.go           # Update: add --remote support
├── sessions.go         # muxi sessions
├── history.go          # muxi history
├── clear.go            # muxi clear
├── jobs.go             # muxi jobs
├── audit.go            # muxi audit
├── stream.go           # muxi stream
├── scheduler.go        # muxi scheduler
├── users.go            # muxi users
├── memory.go           # muxi memory
└── chat.go             # muxi chat (Phase 6)
```

---

## Timeline (Sequential)

| Phase | Tasks | Estimate |
|-------|-------|----------|
| 1 | Foundation (info, agents, mcp, triggers, sops) | 6 hours |
| 2 | Config --remote (secrets, config) | 4 hours |
| 3 | Sessions (sessions, history, clear) | 3 hours |
| 4 | Operations (trigger, jobs, audit, stream) | 5 hours |
| 5 | Advanced (scheduler, users, memory) | 4 hours |
| 6 | Chat (streaming, UX, edge cases) | 6 hours |
| - | Testing & polish | 4 hours |
| **Total** | | **32 hours** |

---

## Parallelization Strategy

Commands can be developed in parallel after the foundation is complete.

### Foundation (Blocking - 1 Droid) ✅ COMPLETE

All other tracks can now proceed.

```
pkg/formation/
├── client.go    # HTTP client with Get/Post/Patch/Delete + admin/client key variants
├── auth.go      # ResolveKeys() from env vars or secrets.enc, BuildFormationURL()
├── types.go     # Response structs for all endpoints (tested against live formation)
└── flags.go     # AddCommonFlags(), ClientFromFlags(), ResolveUserID()
```

**Delivered:**
- `Client` struct with `Get()`, `GetClient()`, `GetWithUser()`, `Post()`, `PostClient()`, `PostWithUser()`, `Patch()`, `Put()`, `Delete()`, `DeleteWithUser()`
- `ResolveFormationID()` - from flag or formation.yaml
- `ResolveProfile()` - from flag, .muxi, or global default
- `ResolveKeys()` - from env vars (`MUXI_ADMIN_KEY`/`MUXI_CLIENT_KEY`) or secrets.enc
- `ClientFromFlags()`, `ClientAndUserFromFlags()` - one-liner client creation
- All response types tested against live `some-forma` formation

---

### Parallel Tracks (After Foundation)

| Track | Commands | Files | Auth | Estimate | Status |
|-------|----------|-------|------|----------|--------|
| **A: Introspection** | `info`, `triggers`, `sops` | info.go, triggers.go, sops.go | Admin/Client | 2h | **Done** |
| **B: CRUD Resources** | `agents`, `mcp` | agents.go, mcp.go | Admin | 3h | Done |
| **C: Config Extension** | `secrets --remote`, `config --remote` | secrets.go*, config.go* | Admin | 3h | Done |
| **D: Sessions** | `sessions`, `history`, `clear` | sessions.go, history.go, clear.go | Client+User | 2h | **Done** |
| **E: Operations** | `trigger`, `jobs`, `audit`, `stream` | triggers.go*, jobs.go, audit.go, stream.go | Mixed | 4h | Done |
| **F: Advanced** | `scheduler`, `users`, `memory` | scheduler.go, users.go, memory.go | Mixed | 3h | **Done** |

*\* = modifying existing file*

---

### Track Details

**Track A - Introspection (Read-only, simple GETs)**
```
muxi info      → GET /status, GET /config
muxi triggers  → GET /triggers  
muxi sops      → GET /sops, GET /sops/{name}
```
- Simplest track, good starter task
- No mutations, just display formatting

**Track B - CRUD Resources (Similar patterns)**
```
muxi agents    → GET/POST/PATCH/DELETE /agents
muxi mcp       → GET/POST/PATCH/DELETE /mcp/servers
```
- Both follow same CRUD pattern
- Include add/remove/enable/disable subcommands

**Track C - Config Extension (Modify existing commands)**
```
muxi secrets --remote  → GET/POST/PUT/DELETE /secrets
muxi config --remote   → GET /config, GET /llm/settings, etc.
```
- Extends existing local commands with `--remote` flag
- Needs careful integration with existing code

**Track D - Sessions (User-scoped)** ✅ COMPLETE
```
muxi sessions  → GET /sessions
muxi history   → GET /sessions/{id}/messages
muxi clear     → DELETE /sessions/{id}
```
- All require `X-User-ID` header
- Uses default user_id from config
- See [docs/guides/sessions.md](guides/sessions.md) for usage guide

**Track E - Operations (Mixed complexity)**
```
muxi trigger   → POST /triggers/{name}
muxi jobs      → GET/DELETE /jobs/{user_id}
muxi audit     → GET/DELETE /audit
muxi stream    → GET /logs/stream (SSE)
```
- `stream` is most complex (SSE handling)
- Others are straightforward

**Track F - Advanced (Lower priority)**
```
muxi scheduler → GET/POST/DELETE /scheduler/jobs
muxi users     → GET/POST/DELETE /users/identifiers
muxi memory    → GET/POST/DELETE /memories
```
- Can be deferred if needed
- Less commonly used features

---

### Parallelization Timeline

```
Hour 0-3:  [Foundation] ████████████████████████████████
Hour 3-5:  [Track A] ████████  [Track B] ████████████  [Track C] ████████████
Hour 3-5:  [Track D] ████████  [Track E] ████████████████████████████████████
Hour 5-8:  [Track E cont.]     [Track F] ████████████████████████
Hour 8+:   [Testing & Integration] ████████████████████████████████████████
```

**With 4-5 droids:** ~8 hours total (vs 32 hours sequential)

---

### Assignment Recommendations

| Droid | Assignment | Why |
|-------|------------|-----|
| **Droid 0** | Foundation | Most experienced, sets patterns for others |
| **Droid 1** | Track A + D | Simple GETs + user-scoped, related patterns |
| **Droid 2** | Track B | CRUD patterns, self-contained |
| **Droid 3** | Track C | Requires understanding existing secrets/config code |
| **Droid 4** | Track E | Most variety, includes SSE streaming |
| **Droid 5** | Track F | Can start later, lower priority |

---

### Integration Points

After parallel development, integration needed for:

1. **Shared types** - Ensure `pkg/formation/types.go` covers all endpoints
2. **Error handling** - Consistent error messages across commands
3. **Output formatting** - Consistent table/list formatting
4. **Flag behavior** - `-F`, `-p`, `-u` work identically everywhere
5. **Help text** - Consistent style and examples

---

## Notes

1. **Formation Detection:** Same logic as shortcuts - read `formation.yaml`
2. **Server Proxy:** Use server's `/api/{id}/*` proxy, not direct port access
3. **Streaming:** Reuse SSE parsing from deploy command
4. **Default User ID:** Critical for good UX - most commands need it
5. **Key Caching:** Store in `~/.muxi/cli/formation-keys.yaml`, TTL 1 hour
6. **Chat Last:** Needs careful UX design for streaming, interrupts, files, etc.
