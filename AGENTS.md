# Agent Workflow Guide

**Project:** MUXI CLI  
**Last Updated:** 2026-02-09

---

## Quick Context

The MUXI CLI is a Go-based command-line tool for creating, deploying, and managing AI agent formations.

**Ecosystem:**
- Part of the larger MUXI ecosystem - see [muxi/AGENTS.md](https://github.com/muxi-ai/muxi/blob/main/AGENTS.md)
- Architecture overview: [muxi/ARCHITECTURE.md](https://github.com/muxi-ai/muxi/blob/main/ARCHITECTURE.md)

**Key Documentation:**
- [contributing/architecture.md](contributing/architecture.md) - CLI architecture, key files, patterns
- [contributing/api-reference.md](contributing/api-reference.md) - HMAC auth, server communication
- [contributing/ux-patterns.md](contributing/ux-patterns.md) - TUI design system and conventions

---

## Before You Start

1. **Check the codebase** - Understand existing patterns in `src/pkg/`
2. **Review UX patterns** - Follow established TUI conventions
3. **Build and test** - Ensure everything compiles before and after changes

---

## Agent Workflow Checklist

### 1. Research the Codebase
- Inspect relevant files before editing
- Trace implementations to understand dependencies
- Check existing patterns in similar commands

### 2. Plan Appropriately
- **Large changes:** Draft a plan, get confirmation before coding
- **Small/straightforward:** Form a quick plan and proceed

### 3. Execute the Plan
- Keep diffs focused
- Follow existing code style
- Add minimal but helpful comments

### 4. Verify
```bash
cd src && go build ./...
cd src && go test ./...
```

---

## Key Files Quick Reference

```
src/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command, CLI setup
│   ├── new.go             # muxi new subcommands
│   ├── config.go          # muxi config subcommands
│   ├── deploy.go          # muxi deploy
│   ├── up.go              # muxi up (local dev)
│   ├── down.go            # muxi down (stop local dev)
│   ├── formation.go       # muxi remote subcommands
│   ├── secrets.go         # muxi secrets subcommands
│   ├── chat.go            # muxi chat
│   ├── logs.go            # muxi logs
│   └── ...                # Other commands
│
├── pkg/
│   ├── defaults/          # Global defaults (~/.muxi/cli/)
│   ├── scaffold/          # Scaffolding wizards
│   ├── secrets/           # Fernet encryption
│   ├── wizard/            # Prompt helpers
│   ├── ui/                # TUI design system
│   ├── server/            # Server API client, HMAC auth
│   ├── registry/          # Registry API client
│   ├── formation/         # Formation API client
│   ├── context/           # Formation context detection
│   ├── validate/          # Formation validation
│   ├── chat/              # Chat session management
│   └── telemetry/         # Usage analytics
```

For full architecture details, see [contributing/architecture.md](contributing/architecture.md).

---

## Coding Conventions

### General Rules
- Follow existing patterns in the codebase
- Use `ui.*` functions for all terminal output
- Handle Ctrl+C gracefully at all prompts
- Use validation loops (retry on invalid input, don't exit)
- Normalize IDs: spaces → hyphens, auto-suggest names

### Code Search Tips

Use ripgrep for searching:

```bash
# Find function definitions
rg "func \w+\(" src/

# Find struct definitions  
rg "type \w+ struct" src/

# Find specific function calls
rg "ui\.PromptSuccess" src/
```

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

- [ ] Review [contributing/ux-patterns.md](contributing/ux-patterns.md) for TUI conventions
- [ ] Check existing code in `src/pkg/` for patterns
- [ ] Build and test before making changes

**Ready to build!**
