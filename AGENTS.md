# Agent Workflow Guide

**Project:** MUXI CLI  
**Last Updated:** 2025-12-08

---

## Quick Context

This is the MUXI CLI - a command-line tool for formation scaffolding, secrets management, registry integration, and server management.

**Ecosystem:**
- Part of the larger MUXI ecosystem - see [../AGENTS.md](../AGENTS.md) for the full picture
- Full architecture overview: [../MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md)
- API specs live in [../schemas/api/](../schemas/api/) (server-api-v1-final.yaml, formation-api-v1-final.yaml)

**Key Documentation:**
- [STATUS.md](STATUS.md) - Current status, what's working, what's next
- [docs/plan-formation-api.md](docs/plan-formation-api.md) - **Formation API implementation plan with parallelization strategy**
- [docs/architecture.md](docs/architecture.md) - CLI architecture, key files, patterns
- [docs/api-reference.md](docs/api-reference.md) - HMAC auth, server communication
- [docs/UX-PATTERNS.md](docs/UX-PATTERNS.md) - TUI design system and conventions

---

## Current Priority: Formation API Commands

**See [docs/plan-formation-api.md](docs/plan-formation-api.md) for full details.**

The next major work is implementing Formation API commands. This can be parallelized:

1. **Foundation (blocking):** `pkg/formation/client.go`, `auth.go`, `types.go` - must complete first
2. **Then 6 parallel tracks:** Each track is independent after foundation

**Track assignments:**
| Track | Commands | Est |
|-------|----------|-----|
| A | `info`, `triggers`, `sops` | 2h |
| B | `agents`, `mcp` | 3h |
| C | `secrets --remote`, `config --remote` | 3h |
| D | `sessions`, `history`, `clear` | 2h |
| E | `trigger`, `jobs`, `audit`, `stream` | 4h |
| F | `scheduler`, `users`, `memory` | 3h |

---

## Before You Start

1. **Read STATUS.md** - Understand what's complete and what's in progress
2. **Check AGENT_CHAT.md** - See if other agents are working on related files
3. **Review docs/UX-PATTERNS.md** - Follow established TUI conventions

---

## Agent Workflow Checklist

Follow this checklist **in order** for every task:

### 1. Sync in AGENT_CHAT
- Check [AGENT_CHAT.md](AGENT_CHAT.md) for in-progress work
- Read the entire file before adding anything
- Pick a short, task-themed tag (e.g., `<fix-deploy-streaming>`)
- List files you'll modify inside your tag
- If another agent claimed a file you need, wait or coordinate

### 2. Research the Codebase
- Inspect relevant files before editing
- Trace implementations to understand dependencies
- Use ast-grep for code patterns (see tips below)

### 3. Plan Appropriately
- **Large changes:** Draft a plan, get confirmation before coding
- **Small/straightforward:** Form a quick plan and proceed

### 4. Execute the Plan
- Keep diffs focused
- Follow existing code style
- Add minimal but helpful comments

### 5. Verify
```bash
cd src && go build ./...
cd src && go test ./...
```

### 6. Clean Up
- Remove your tag from AGENT_CHAT.md when done
- Update STATUS.md if you completed a milestone

---

## Key Files Quick Reference

```
src/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command, CLI setup, command groups
│   ├── set.go             # muxi set default (server/registry/user)
│   ├── new.go             # muxi new subcommands
│   ├── config.go          # muxi config subcommands
│   ├── deploy.go          # muxi deploy
│   ├── formation.go       # muxi formation subcommands
│   ├── server.go          # muxi server subcommands
│   ├── secrets.go         # muxi secrets subcommands
│   ├── shortcuts.go       # Formation dir shortcuts (get, stop, start, etc.)
│   └── registry.go        # Registry commands (login, push, pull, etc.)
│
├── pkg/
│   ├── defaults/          # Global defaults (~/.muxi/cli/defaults.yaml)
│   ├── scaffold/          # Scaffolding wizards
│   ├── secrets/           # Fernet encryption
│   ├── wizard/            # Prompt helpers
│   ├── ui/                # TUI design system
│   ├── server/            # Server API client, HMAC auth
│   ├── registry/          # Registry API client
│   ├── context/           # Formation context detection
│   └── validate/          # Formation validation
│
│   # NEW (Formation API - to be created):
│   └── formation/         # Formation API client (TODO)
│       ├── client.go      # HTTP client
│       ├── auth.go        # API key resolution
│       └── types.go       # Response types
```

For full architecture details, see [docs/architecture.md](docs/architecture.md).

---

## Coding Conventions

### General Rules
- Follow existing patterns in the codebase
- Use `ui.*` functions for all terminal output
- Handle Ctrl+C gracefully at all prompts
- Use validation loops (retry on invalid input, don't exit)
- Normalize IDs: spaces → hyphens, auto-suggest names

### Code Search Tips

Prefer **ast-grep** over text grep for code patterns:

```bash
# Find function definitions
ast-grep -p 'func $NAME($$$) error' --lang go

# Find struct definitions  
ast-grep -p 'type $NAME struct { $$$ }' --lang go

# Find specific function calls
ast-grep -p 'ui.PromptSuccess($$$)' --lang go
```

Use text grep for: error messages, comments, strings, config files.

---

## Dependencies

```go
require (
    github.com/spf13/cobra v1.8.0      // CLI framework
    gopkg.in/yaml.v3 v3.0.1            // YAML parsing
    github.com/fatih/color v1.16.0     // Colored output
)
```

---

## Session Checklist

Starting a new session? Quick checklist:

- [ ] Read STATUS.md for current state
- [ ] Check AGENT_CHAT.md for active work
- [ ] Review docs/UX-PATTERNS.md for TUI conventions
- [ ] Check existing code in src/pkg/ for patterns

**Ready to build!**
