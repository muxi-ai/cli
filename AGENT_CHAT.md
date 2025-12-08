# Agent Chat - Work Coordination

**Purpose:** Prevent conflicts when multiple agents work on this repo simultaneously.

---

## Current Priority: Formation API

See **[docs/plan-formation-api.md](docs/plan-formation-api.md)** for the full implementation plan.

**Parallelization:** Foundation must complete first, then 6 tracks can run in parallel.

| Track | Commands | Status |
|-------|----------|--------|
| **Foundation** | `pkg/formation/client.go`, `auth.go`, `types.go`, `flags.go` | **DONE** |
| **A** | `info`, `triggers`, `sops` | **DONE** |
| **B** | `agents`, `mcp` | **DONE** |
| **C** | `secrets --remote`, `config --remote` | **DONE** |
| **D** | `sessions`, `history`, `clear` | **DONE** |
| **E** | `trigger`, `jobs`, `audit`, `stream` | **DONE** |
| **F** | `scheduler`, `users`, `memory` | **DONE** |

---

## Rules

1. **Before starting work:** Read this entire file
2. **Claim your work:** Add a tag with files you'll modify
3. **Check for conflicts:** If someone claimed a file you need, wait or coordinate
4. **Clean up:** Remove your tag when done

---

## Tag Format

```xml
<your-task-name>
Track: Foundation | A | B | C | D | E | F
Files:
- src/cmd/example.go
- src/pkg/example/example.go

Plan:
- What you're doing
- Expected changes

Status: in-progress | blocked | done
</your-task-name>
```

---

## Active Work

<!-- Add your work tag below this line -->

<!-- Add your work tag above this line -->

---

## Recently Completed

<!-- Move completed tags here for reference, delete after a few days -->

### 2025-12-08
- **Track A (Introspection)** - `src/cmd/` files:
  - `info.go` - Formation status/config (GET /status, GET /config with --full)
  - `triggers.go` - List triggers (GET /triggers)
  - `sops.go` - List SOPs, show details (GET /sops, GET /sops/{name})
  - All with -F, -p flags, GroupID: formation
  - Also added missing client methods for Track F: GetSchedulerConfig, GetSchedulerJobs, GetUserIdentifiers, GetUserIdentifiersForUser, LinkUserIdentifier, UnlinkUserIdentifier, ResolveUserIdentifier
- **Track F (Advanced Admin)** - `src/cmd/` files:
  - `scheduler.go` - View scheduler config, list scheduled jobs (GET /scheduler, GET /scheduler/jobs)
  - `users.go` - Manage user identifiers (list, link, unlink, resolve)
  - `memory.go` - View memory config, manage user memories (list, add, delete)
  - Added GetMemories, AddMemory, DeleteMemory methods to `pkg/formation/client.go`
  - All with -F, -p flags; user-scoped ops support -u flag, GroupID: formation
- **Track E (Operations)** - `src/cmd/` files:
  - `trigger.go` - Fire triggers with JSON data (POST /triggers/{name})
  - `jobs.go` - List/cancel async jobs (GET/DELETE /jobs/{user_id})
  - `audit.go` - View/clear audit log (GET/DELETE /audit)
  - `stream.go` - Real-time SSE log streaming (GET /logs/stream)
  - Added TriggerTrigger, StreamLogs methods to `pkg/formation/client.go`
  - All with -F, -p flags, GroupID: formation
- **Track B (CRUD Resources)** - `src/cmd/` files:
  - `agents.go` - List/manage agents (GET /agents) with list, add, remove, enable, disable subcommands
  - `mcp.go` - List/manage MCP servers (GET /mcp/servers) with list, add, remove subcommands
  - Both with -F, -p, -v flags, GroupID: formation
- **Track C (Config Extension)** - Modified `src/cmd/` files:
  - `secrets.go` - Added `--remote` flag to fetch secrets from Formation API
  - `config.go` - Added `--remote` flag to config, llm, memory, overlord subcommands
  - Added GetLLMSettings, GetMemoryConfig, GetOverlordConfig methods to `pkg/formation/client.go`
- **Track D (Session Management)** - `src/cmd/` files:
  - `sessions.go` - List user sessions (GET /sessions)
  - `history.go` - View session messages (GET /sessions/{id}/messages)
  - `clear.go` - Delete session (DELETE /sessions/{id})
  - All with -F, -p, -u flags, GroupID: formation
- **Foundation (Formation API)** - `pkg/formation/` package with:
  - `client.go` - HTTP client with admin/client key auth
  - `auth.go` - API key resolution from env vars or secrets.enc
  - `types.go` - Response types for all Formation API endpoints
  - `flags.go` - Common flag helpers (-F, -p, -u)
  - Tested against live `some-forma` formation
- Unified `muxi set default` command (server/registry/user)
- Command groups in CLI help output
- Default user_id support in `pkg/defaults/`
- Formation API plan with parallelization strategy

