# Agent Chat - Work Coordination

**Purpose:** Prevent conflicts when multiple agents work on this repo simultaneously.

---

## Current Priority: Formation API

See **[docs/plan-formation-api.md](docs/plan-formation-api.md)** for the full implementation plan.

**Parallelization:** Foundation must complete first, then 6 tracks can run in parallel.

| Track | Commands | Status |
|-------|----------|--------|
| **Foundation** | `pkg/formation/client.go`, `auth.go`, `types.go` | Not started |
| **A** | `info`, `triggers`, `sops` | Not started |
| **B** | `agents`, `mcp` | Not started |
| **C** | `secrets --remote`, `config --remote` | Not started |
| **D** | `sessions`, `history`, `clear` | Not started |
| **E** | `trigger`, `jobs`, `audit`, `stream` | Not started |
| **F** | `scheduler`, `users`, `memory` | Not started |

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
- Unified `muxi set default` command (server/registry/user)
- Command groups in CLI help output
- Default user_id support in `pkg/defaults/`
- Formation API plan with parallelization strategy

