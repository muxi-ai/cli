# CLI Pre-Launch Cleanup Plan

Based on [prelaunch-cleanup.md](../../architecture/prelaunch-cleanup.md) from runtime repo cleanup.

**Last Updated:** 2026-01-26

---

## Summary

The CLI repo is relatively clean. Main tasks:
1. Consolidate docs with muxi/docs/cli (avoid duplication)
2. Move planning docs to `../architecture/cli/`
3. Rename `docs/` to `contributing/` (contributor-focused docs only)
4. Delete redundant files
5. Add OpenSSF badge and .vscode/settings.json

---

## 1. Documentation Comparison

### muxi/docs/cli/ (user docs site) - 10 files
```
chat.md, cheatsheet.md, configuration.md, deploy.md, new.md
README.md, registry.md, secrets.md, server.md, setup.md
```

### cli/docs/guides/ - 27 files
Many overlap with docs site, many are missing from docs site.

---

## 2. Documentation Actions

### A. Move to `muxi/docs/cli/` (user-facing, missing from docs site)
These guides should be on the public docs site:
- [ ] `agents.md` - Agent management
- [ ] `mcps.md` - MCP server management
- [ ] `triggers.md` - Trigger management
- [ ] `sops.md` - SOP viewing
- [ ] `sessions.md` - Session management
- [ ] `memory.md` - Memory management
- [ ] `logs.md` - Log streaming
- [ ] `bump.md` - Version bumping
- [ ] `info.md` - Formation info
- [ ] `formations.md` - Formation management

### B. Delete from cli/docs/guides/ (duplicates of docs site)
- [ ] `chat.md` → already in muxi/docs/cli/
- [ ] `registry.md` → already in muxi/docs/cli/
- [ ] `secrets.md` → already in muxi/docs/cli/
- [ ] `server.md` → already in muxi/docs/cli/
- [ ] `config.md` → already in muxi/docs/cli/configuration.md

### C. Move to `../architecture/cli/` (historical/planning)
- [ ] `docs/plan-server.md`
- [ ] `docs/plan-formation-api.md`
- [ ] `docs/plan-logging-restructure.md`
- [ ] `docs/plan-afs-extension.md`
- [ ] `docs/plan-telemetry.md`
- [ ] `docs/DESIGN.md`
- [ ] `docs/COMMAND-SEMANTICS.md`
- [ ] `docs/REGISTRY.md`

### D. Delete entirely
- [ ] `docs/GIT-WORKFLOW.md` → reference muxi repo instead
- [ ] `docs/guides/` directory (after moving relevant files)
- [ ] `STATUS.md` - outdated task tracking
- [ ] `AGENT_CHAT.md` - not actively used
- [ ] `playground/` directory - test data, already gitignored

### E. Keep in cli/contributing/ (rename from docs/)
Contributor-focused documentation:
```
contributing/
├── README.md           # Contributing overview
├── architecture.md     # CLI architecture
├── api-reference.md    # Server/Formation API reference
├── streaming-events.md # SSE event types
├── UX-PATTERNS.md      # TUI patterns
├── TUI-DESIGN.md       # Design system
└── BANNERS.md          # Banner/logo specs
```

---

## 3. Target Directory Structure

```
cli/
├── .github/workflows/
├── .vscode/settings.json    # ADD
├── contributing/            # RENAME from docs/
│   ├── README.md
│   ├── architecture.md
│   ├── api-reference.md
│   ├── streaming-events.md
│   ├── UX-PATTERNS.md
│   ├── TUI-DESIGN.md
│   └── BANNERS.md
├── src/
│   ├── cmd/
│   └── pkg/
├── AGENTS.md
├── CHANGELOG.md
├── LICENSE
└── README.md
```

---

## 4. Configuration

### Add .vscode/settings.json
```json
{
    "files.associations": {
        "*.afs": "yaml"
    },
    "editor.formatOnSave": true,
    "editor.rulers": [100],
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package"
}
```

### Update .gitignore
- [ ] Remove `.vscode/` from gitignore (commit settings for consistency)

---

## 5. CI/CD & README

### Add OpenSSF badge to README.md
```markdown
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/muxi-ai/cli/badge)](https://scorecard.dev/viewer/?uri=github.com/muxi-ai/cli)
```

### README updates
- [ ] Ensure installation instructions are current
- [ ] Link to muxi repo for CONTRIBUTING, CODE_OF_CONDUCT, SECURITY

---

## 6. Security

- [x] No Dependabot alerts (checked 2026-01-26)
- [x] golang.org/x/net already updated to v0.48.0

---

## Execution Checklist

### Phase 1: Move docs to muxi/docs/cli/
- [ ] Copy 10 guides to muxi/docs/cli/ (with frontmatter)
- [ ] Commit and push to muxi repo

### Phase 2: Move planning docs to architecture/
- [ ] Create `../architecture/cli/` directory
- [ ] Move 8 planning/design docs there
- [ ] Commit and push to architecture repo

### Phase 3: CLI repo cleanup
- [ ] Delete duplicated guides (5 files)
- [ ] Delete remaining guides/ directory
- [ ] Rename `docs/` → `contributing/`
- [ ] Delete STATUS.md, AGENT_CHAT.md
- [ ] Delete playground/ directory
- [ ] Add .vscode/settings.json
- [ ] Update .gitignore
- [ ] Add OpenSSF badge to README
- [ ] Commit and push

---

## Decisions Needed

1. **Which guides to move to docs site?** (Listed 10 above - confirm)
2. **Any guides to keep as contributor docs?** (e.g., admin.md, security.md)
3. **Delete playground/?** (Already gitignored, but physically present)
