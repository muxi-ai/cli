# Agent Chat - Work Coordination

**Purpose:** Prevent conflicts when multiple agents work on this repo simultaneously.

---

## Current Priority: Formation API

See **[docs/plan-formation-api.md](docs/plan-formation-api.md)** for the full implementation plan.

**Parallelization:** Foundation must complete first, then 6 tracks can run in parallel.

| Track | Commands | Status |
|-------|----------|--------|
| **Foundation** | `pkg/formation/client.go`, `auth.go`, `types.go`, `flags.go` | **DONE** |
| **A** | `info`, `triggers`, `sops` | Ready |
| **B** | `agents`, `mcp` | Ready |
| **C** | `secrets --remote`, `config --remote` | Ready |
| **D** | `sessions`, `history`, `clear` | Ready |
| **E** | `trigger`, `jobs`, `audit`, `stream` | Ready |
| **F** | `scheduler`, `users`, `memory` | Ready |

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

