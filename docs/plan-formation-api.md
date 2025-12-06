# Formation API Commands Implementation Plan

**Date:** 2025-12-06  
**Status:** Planning  
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
2. **Shortcut from formation dir** - Like existing shortcuts
3. **Formation ID flag** - `--formation <id>` or `-F <id>` for anywhere
4. **User ID required** - Most client endpoints need `--user <id>`
5. **No conflicts** - Don't shadow existing commands

---

## Implementation Phases

### Phase 1: Chat (Core Interaction)
1. `muxi chat` - Send message to formation (streaming)

### Phase 2: Introspection
2. `muxi info` - Formation runtime status/info
3. `muxi agents list` - List agents in formation (conflicts with `muxi new agent`, so use subcommand)
4. `muxi triggers` - List available triggers

### Phase 3: Session Management
5. `muxi sessions` - List user sessions
6. `muxi history` - Get session message history
7. `muxi clear` - Clear session/buffer

### Phase 4: Advanced
8. `muxi trigger` - Execute a trigger
9. `muxi audit` - View audit log
10. `muxi stream` - Stream live logs (SSE)

---

## Phase 1: Chat

### `muxi chat`

**The primary way to interact with a deployed formation.**

```bash
# From inside formation directory
muxi chat "What's the weather like today?"

# From anywhere
muxi chat --formation my-formation "What's the weather like today?"
muxi chat -F my-formation "Hello!"

# With user context
muxi chat --user alice "Help me with my account"

# With session persistence
muxi chat --user alice --session sess_abc123 "Continue our conversation"

# Non-streaming mode
muxi chat --no-stream "Quick question"

# Pipe input
echo "Analyze this data" | muxi chat --user alice
cat document.txt | muxi chat --user alice "Summarize this:"
```

**API:** `POST /api/{formation}/v1/chat`

**Auth:** Client Key (`X-MUXI-CLIENT-KEY`)

**Headers:**
- `X-User-ID: {user_id}` (required)
- `Accept: text/event-stream` (for streaming)

**Request:**
```json
{
  "message": "What's the weather like?",
  "session_id": "sess_abc123"  // optional
}
```

**Output (streaming):**
```
muxi chat --user alice "What's the weather?"

◐ Thinking...

The weather today is sunny with a high of 72°F (22°C). 
Perfect day to be outside!

---
Tokens: 45 prompt, 28 completion | 1.2s
```

**Output (non-streaming):**
```
muxi chat --user alice --no-stream "Quick question"

The answer is 42.
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (optional if in formation dir) |
| `--user` | `-u` | User ID (required) |
| `--session` | `-s` | Session ID for conversation continuity |
| `--no-stream` | | Disable streaming output |
| `--profile` | | Server profile to use |

**Implementation Notes:**
- Detect formation from `formation.yaml` if in formation dir
- Get formation URL from server: `GET /rpc/formations/{id}` returns port
- Connect to formation at `http://localhost:{port}/v1/chat`
- Or via server proxy: `http://server:7890/api/{formation}/v1/chat`
- Stream tokens in real-time using SSE
- Show usage stats at the end

---

## Phase 2: Introspection

### `muxi info`

**Get formation runtime status and configuration.**

```bash
# From formation directory
muxi info

# From anywhere
muxi info --formation my-formation

# Verbose output
muxi info -v
```

**API:** `GET /api/{formation}/v1/status`

**Auth:** Admin Key

**Output:**
```
muxi info

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

### `muxi agents list`

**List agents in the running formation.**

```bash
muxi agents list
muxi agents list --formation my-formation
muxi agents list -v  # verbose
```

**API:** `GET /api/{formation}/v1/agents`

**Auth:** Admin Key

**Output:**
```
muxi agents list

  ID                  ROLE        STATUS    MODEL
  weather-assistant   specialist  ● active  openai/gpt-4o
  code-helper         specialist  ● active  anthropic/claude-3
  general-assistant   generalist  ○ idle    openai/gpt-4o-mini
```

**Verbose Output:**
```
muxi agents list -v

weather-assistant
  Role:        specialist
  Status:      ● active
  Model:       openai/gpt-4o
  Specialties: weather_forecasting, meteorology
  A2A:         internal=yes, external=yes
  Source:      formation

code-helper
  Role:        specialist
  Status:      ● active
  Model:       anthropic/claude-3
  Specialties: programming, debugging
  A2A:         internal=yes, external=no
  Source:      api
```

---

### `muxi triggers`

**List available triggers.**

```bash
muxi triggers
muxi triggers --formation my-formation
```

**API:** `GET /api/{formation}/v1/triggers`

**Auth:** Client Key

**Output:**
```
muxi triggers

  NAME                    DESCRIPTION
  github-issue           Handle GitHub issue events
  linear-ticket          Process Linear ticket updates
  deployment-notify      Deployment notification handler
```

---

## Phase 3: Session Management

### `muxi sessions`

**List user sessions.**

```bash
muxi sessions --user alice
muxi sessions --user alice --active  # only active sessions
muxi sessions -F my-formation --user alice
```

**API:** `GET /api/{formation}/v1/sessions`

**Auth:** Client Key + `X-User-ID` header

**Output:**
```
muxi sessions --user alice

  SESSION ID      MESSAGES   LAST ACTIVITY         STATUS
  sess_abc123     25         2 minutes ago         ● active
  sess_xyz789     12         3 hours ago           ○ inactive
  sess_def456     8          2 days ago            ○ inactive
```

---

### `muxi history`

**Get session message history.**

```bash
muxi history --session sess_abc123 --user alice
muxi history -s sess_abc123 -u alice --lines 20
muxi history -s sess_abc123 -u alice --json
```

**API:** `GET /api/{formation}/v1/sessions/{session_id}/messages`

**Auth:** Client Key + `X-User-ID` header

**Output:**
```
muxi history --session sess_abc123 --user alice

Session: sess_abc123 (25 messages)

[10:00:15] alice: What's the weather like?
[10:00:18] weather-assistant: The weather today is sunny with a high of 72°F.

[10:05:00] alice: Can you help me with Python code?
[10:05:03] code-helper: Of course! What do you need help with?

... (showing last 10 messages, use --lines for more)
```

---

### `muxi clear`

**Clear session or buffer memory.**

```bash
# Clear specific session
muxi clear --session sess_abc123 --user alice

# Clear all user buffer memory
muxi clear --user alice --all

# With confirmation skip
muxi clear --session sess_abc123 --user alice -f
```

**API:** 
- `DELETE /api/{formation}/v1/sessions/{session_id}`
- `DELETE /api/{formation}/v1/memory/buffer`

**Auth:** Client Key + `X-User-ID` header

**Output:**
```
muxi clear --session sess_abc123 --user alice

Clear session "sess_abc123"? (25 messages will be deleted) [y/N]: y

✓ Session cleared (25 messages removed)
```

---

## Phase 4: Advanced

### `muxi trigger`

**Execute a trigger with data.**

```bash
muxi trigger github-issue --data '{"issue": {"number": 123, "title": "Bug"}}'
muxi trigger github-issue --file event.json
muxi trigger github-issue --user webhook-bot --data '...'
```

**API:** `POST /api/{formation}/v1/triggers/{trigger_name}`

**Auth:** Client Key

**Output:**
```
muxi trigger github-issue --data '{"issue": {"number": 123}}'

  ⠹ Executing trigger...
  ✓ Trigger queued

  Job ID: job_abc123
  Status: processing
  
  Use 'muxi jobs --user webhook-bot' to check status
```

---

### `muxi audit`

**View formation audit log.**

```bash
muxi audit
muxi audit --lines 50
muxi audit --action agent.created
muxi audit --since "2025-12-01"
```

**API:** `GET /api/{formation}/v1/audit`

**Auth:** Admin Key

**Output:**
```
muxi audit --lines 10

  TIMESTAMP            ACTION           RESOURCE         USER
  2025-12-06 10:15:00  agent.created    weather-bot      admin
  2025-12-06 10:14:30  secret.created   OPENAI_KEY       admin
  2025-12-06 09:00:00  scheduler.job    daily-report     system
```

---

### `muxi stream`

**Stream live logs from formation.**

```bash
muxi stream --user alice
muxi stream --level ERROR
muxi stream --agent weather-assistant
muxi stream --request req_abc123
```

**API:** `GET /api/{formation}/v1/logs/stream` (SSE)

**Auth:** Admin Key

**Output:**
```
muxi stream --user alice

Streaming logs for user: alice (Ctrl+C to stop)

[10:15:00] INFO  chat.started     user=alice session=sess_abc
[10:15:01] INFO  agent.invoked    agent=weather-assistant
[10:15:02] INFO  llm.request      model=gpt-4o tokens=150
[10:15:03] INFO  chat.completed   user=alice duration=2.1s
^C
Stopped.
```

---

## API Key Management

Formations use API keys for authentication. The CLI needs to:

1. **Get keys from server** - Server knows formation's keys
2. **Cache keys locally** - Avoid repeated lookups
3. **Support key override** - `--admin-key` / `--client-key` flags

**Key Discovery:**
```bash
# Server returns formation details including API endpoint
GET /rpc/formations/{id}
{
  "id": "my-formation",
  "port": 8001,
  "api_url": "http://localhost:7890/api/my-formation",
  "keys": {
    "admin": "fma-abc...",  # Only returned with server HMAC auth
    "client": "fmc-xyz..."
  }
}
```

**Key Caching:**
- Store in `~/.muxi/cli/formation-keys.yaml`
- TTL: 1 hour
- Refresh on 401 errors

---

## Command Summary

| Command | Description | Auth | Phase |
|---------|-------------|------|-------|
| `muxi chat` | Send message to formation | Client | 1 |
| `muxi info` | Formation runtime status | Admin | 2 |
| `muxi agents list` | List formation agents | Admin | 2 |
| `muxi triggers` | List available triggers | Client | 2 |
| `muxi sessions` | List user sessions | Client | 3 |
| `muxi history` | Get message history | Client | 3 |
| `muxi clear` | Clear session/buffer | Client | 3 |
| `muxi trigger` | Execute a trigger | Client | 4 |
| `muxi audit` | View audit log | Admin | 4 |
| `muxi stream` | Stream live logs | Admin | 4 |

---

## File Structure

```
pkg/formation/
├── client.go      # Formation API client
├── auth.go        # API key management
├── types.go       # API response types
└── cache.go       # Key caching

cmd/
├── chat.go        # muxi chat
├── info.go        # muxi info
├── agents.go      # muxi agents list (add subcommand)
├── triggers.go    # muxi triggers, muxi trigger
├── sessions.go    # muxi sessions, muxi history
├── clear.go       # muxi clear
├── audit.go       # muxi audit
└── stream.go      # muxi stream
```

---

## Existing Commands (No Conflicts)

| Existing Command | Purpose | Conflict? |
|-----------------|---------|-----------|
| `muxi deploy` | Deploy formation | No |
| `muxi formation *` | Server lifecycle | No |
| `muxi get/stop/start/...` | Server shortcuts | No - different purpose |
| `muxi new agent` | Scaffold agent | Use `muxi agents list` to avoid |
| `muxi config *` | Configure formation | No |
| `muxi secrets *` | Local secrets | No - Formation API is runtime |
| `muxi server *` | Server profiles | No |
| `muxi login/push/...` | Registry | No |

---

## Timeline

| Phase | Tasks | Estimate |
|-------|-------|----------|
| 1 | Chat command (streaming, sessions) | 4 hours |
| 2 | Info, agents list, triggers | 3 hours |
| 3 | Sessions, history, clear | 3 hours |
| 4 | Trigger execution, audit, stream | 4 hours |
| - | Testing & polish | 2 hours |
| **Total** | | **16 hours** |

---

## Notes

1. **Formation Detection**: Same logic as shortcuts - read `formation.yaml`
2. **Server Proxy**: Use server's `/api/{id}/*` proxy, not direct port access
3. **Streaming**: Reuse SSE parsing from deploy command
4. **User ID**: Most client endpoints require it - make it a persistent setting?
5. **Key Caching**: Important for performance - avoid key lookup on every request
