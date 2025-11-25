# 1-Week Implementation Plan

**Timeline:** 7 Days  
**Goal:** Ship production-ready CLI with full API coverage

---

## Day 1-2: Foundation & Core (Monday-Tuesday)

### Configuration System (4 hours)
- [ ] `pkg/config/` - Load/save all 4 config files
- [ ] Config structs (servers.yaml, formations.yaml, registries.yaml, config.yaml)
- [ ] Default values (registry.muxi.org)
- [ ] File permissions (600 for sensitive files)
- [ ] Tests

### Secrets Management (4 hours)
- [ ] `pkg/secrets/` - AES-256-GCM encryption (port from runtime)
- [ ] Encrypt/decrypt
- [ ] Load/save secrets.enc
- [ ] Generate encryption key
- [ ] Tests

### Formation Context (2 hours)
- [ ] `pkg/context/` - Formation detection
- [ ] Walk up directory tree (max 5 levels)
- [ ] Parse formation.yaml
- [ ] Load secrets
- [ ] Tests

### Basic Commands (6 hours)
- [ ] `muxi init` - Formation scaffolding
  - Create directory structure
  - Generate formation.yaml
  - Auto-generate ADMIN_KEY + CLIENT_KEY
  - Create secrets.enc
  - Generate .key
  - Create .gitignore
- [ ] `muxi validate` - Validation
  - Parse formation.yaml
  - Validate schema
  - Check secret references
  - Show errors/warnings
- [ ] `muxi secrets set/list/delete` - Secret management
  - Set secret (prompt for value)
  - List keys (no values)
  - Delete with validation
- [ ] Tests

**Day 1-2 Deliverable:** Can create formations, manage secrets locally

---

## Day 3: Server Operations (Wednesday)

### API Clients (4 hours)
- [ ] `pkg/client/server.go` - Server API client
  - HMAC signature generation
  - HTTP client wrapper
  - Error handling
- [ ] `pkg/client/resolver.go` - Connection resolver
  - Profile resolution (3-tier)
  - Multi-server support
- [ ] Tests

### Profile Management (3 hours)
- [ ] `muxi profile add` - Add profile wizard
  - Interactive prompts
  - Test connection
  - Save to servers.yaml
- [ ] `muxi profile list/use/remove`
- [ ] Tests

### Deployment (4 hours)
- [ ] `muxi deploy` - Deploy to server(s)
  - Create bundle (tar.gz)
  - Upload to all servers in profile
  - Show progress per server
  - Handle errors
- [ ] Tests

### Server Management (3 hours)
- [ ] `muxi formation list/stop/restart/rollback/delete`
- [ ] `muxi server status/logs/ping`
- [ ] Tests

**Day 3 Deliverable:** Can deploy to servers, manage formations

---

## Day 4: Registry & Formation API (Thursday)

### Registry Integration (4 hours)
- [ ] `pkg/client/registry.go` - Registry API client
  - OAuth flow (open browser)
  - Bearer token auth
  - Save to registries.yaml
- [ ] `muxi login/logout` - Authentication
- [ ] `muxi push` - Publish formation
  - Create bundle
  - Upload to registry
  - Show progress
- [ ] `muxi pull` - Download formation
  - Fetch metadata
  - Download from GitHub
  - Extract
  - Handle public/private
- [ ] `muxi search` - Search registry
- [ ] Tests

### Formation API Client (4 hours)
- [ ] `pkg/client/formation.go` - Formation API client
  - API key auth (admin/client)
  - Connection via server proxy
  - Error handling
- [ ] Tests

### Basic Formation Commands (4 hours)
- [ ] `muxi status` - Runtime status
- [ ] `muxi config show` - Full configuration
- [ ] Tests

**Day 4 Deliverable:** Registry working, can connect to formations

---

## Day 5: Formation Configuration (Friday)

### Agent Management (3 hours)
- [ ] `muxi agent list/add/get/update/delete`
- [ ] File upload support (`--file`)
- [ ] Tests

### MCP Management (3 hours)
- [ ] `muxi mcp list/add/get/update/delete`
- [ ] File upload support
- [ ] Tests

### Chat & Sessions (4 hours)
- [ ] `muxi chat` - Interactive chat
  - SSE stream handling
  - Interactive prompt
  - Exit commands
- [ ] `muxi session list/get/messages/delete`
- [ ] Tests

### Triggers & Jobs (2 hours)
- [ ] `muxi trigger list/get/invoke`
- [ ] `muxi job list/cancel`
- [ ] Tests

**Day 5 Deliverable:** Formation configuration working, chat functional

---

## Day 6: Advanced Features (Saturday)

### Secrets Setup Wizard (3 hours)
- [ ] `muxi secrets setup` - Setup wizard
  - Read secrets.example
  - Find all secret references
  - Smart prompts with context
  - Handle optional secrets
- [ ] Secret reference validation
  - Find all ${{ secrets.XXX }} references
  - Prevent deletion of referenced secrets
  - Sync with secrets.example
- [ ] Tests

### Logs & Monitoring (3 hours)
- [ ] `muxi logs` - Log streaming
  - Server logs (from server)
  - Formation logs (via Formation API)
  - `--follow` flag (SSE)
- [ ] `muxi audit list/clear` - Audit logs
- [ ] Tests

### Remaining Commands (4 hours)
- [ ] `muxi sop list/get` - SOPs (read-only)
- [ ] `muxi logging` commands - Logging config
- [ ] `muxi scheduler` commands - Scheduler jobs
- [ ] `muxi overlord/llm/memory` commands - Advanced config
- [ ] `muxi user` commands - User management
- [ ] Tests

**Day 6 Deliverable:** All ~80 commands implemented

---

## Day 7: Polish & Ship (Sunday)

### UX Improvements (3 hours)
- [ ] Animated spinners (github.com/briandowns/spinner)
- [ ] Progress bars (github.com/schollz/progressbar/v3)
- [ ] Nice error messages
  - "Add one now?" prompts
  - Helpful suggestions
  - Color coding
- [ ] Interactive wizards polished

### Testing (4 hours)
- [ ] Run all unit tests
- [ ] Integration tests
  - Mock server/formation/registry APIs
  - Test end-to-end flows
- [ ] Manual testing
  - Test with real server
  - Test multi-server deployment
  - Test registry integration
- [ ] Fix bugs

### Documentation (2 hours)
- [ ] README.md - Getting started
- [ ] Command examples
- [ ] Configuration guide
- [ ] Troubleshooting

### Release (1 hour)
- [ ] Version 1.0.0
- [ ] Build for platforms (macOS, Linux, Windows)
- [ ] Tag release
- [ ] Announce

**Day 7 Deliverable:** Production-ready CLI shipped! 🚀

---

## Daily Checklist

### Every Day:
- [ ] Write tests for new code
- [ ] Run tests before committing
- [ ] Keep code clean and documented
- [ ] Update DESIGN.md if architecture changes

### Every Commit:
- [ ] Tests pass
- [ ] Code formatted (go fmt)
- [ ] No linter errors (go vet)

---

## Success Metrics

**By End of Week:**
- [ ] ~80 commands implemented
- [ ] Multi-server deployment working
- [ ] Registry integration working
- [ ] Secrets management with validation
- [ ] Formation API fully covered
- [ ] Tests passing (>70% coverage)
- [ ] Nice UX (spinners, progress, errors)
- [ ] Production ready

---

## Risk Mitigation

### If Behind Schedule:
**Day 3:** Skip server ping, focus on core deployment  
**Day 5:** Skip advanced config commands (overlord, llm, memory)  
**Day 6:** Skip scheduler, logging config - focus on core features  

**Must-Have by Day 7:**
- ✅ init, validate, deploy, secrets
- ✅ profile management
- ✅ registry (push/pull)
- ✅ formation lifecycle (list, stop, restart)
- ✅ agents, mcps (basic CRUD)
- ✅ chat

**Nice-to-Have (can defer):**
- ⏸️ Advanced config commands
- ⏸️ Scheduler jobs
- ⏸️ Logging configuration
- ⏸️ Full user management

---

## Development Tips

### Fast Iteration:
```bash
# Run specific tests
go test ./pkg/config -v

# Run with race detector
go test -race ./...

# Build quickly
go build -o muxi ./src

# Test command
./muxi init test-formation
cd test-formation
./muxi validate
```

### Debug Mode:
```bash
export MUXI_DEBUG=1
./muxi deploy --profile local
```

### Mock APIs:
```bash
# Use httptest for testing
server := httptest.NewServer(handler)
defer server.Close()
```

---

## Package Implementation Order

### Day 1-2:
1. `pkg/config` - Configuration
2. `pkg/secrets` - Encryption
3. `pkg/context` - Formation detection
4. `cmd/init` - muxi init
5. `cmd/validate` - muxi validate
6. `cmd/secrets` - muxi secrets

### Day 3:
7. `pkg/client/server` - Server API
8. `pkg/client/resolver` - Connection resolution
9. `cmd/profile` - muxi profile
10. `cmd/deploy` - muxi deploy
11. `cmd/formation` - muxi formation
12. `cmd/server` - muxi server

### Day 4:
13. `pkg/client/registry` - Registry API
14. `pkg/client/formation` - Formation API
15. `cmd/registry` - muxi login/push/pull/search
16. `cmd/status` - muxi status
17. `cmd/config` - muxi config show

### Day 5:
18. `cmd/agent` - muxi agent
19. `cmd/mcp` - muxi mcp
20. `cmd/chat` - muxi chat
21. `cmd/session` - muxi session
22. `cmd/trigger` - muxi trigger
23. `cmd/job` - muxi job

### Day 6:
24. Enhanced `cmd/secrets` - Setup wizard, validation
25. `cmd/logs` - muxi logs
26. `cmd/audit` - muxi audit
27. All remaining commands

### Day 7:
28. `pkg/ui` - Spinners, progress bars
29. Error message improvements
30. Tests, docs, polish

---

## Definition of Done

**Each Command:**
- [ ] Implemented
- [ ] Unit tests written
- [ ] Integration test (if applicable)
- [ ] Error handling
- [ ] Help text
- [ ] Examples in docs

**Each API Client:**
- [ ] Implemented
- [ ] Authentication working
- [ ] Error handling
- [ ] Unit tests
- [ ] Integration tests (mocked)

**Final:**
- [ ] All commands working
- [ ] Tests passing
- [ ] Docs complete
- [ ] Ready for production

---

**LET'S SHIP IT IN 7 DAYS! 🚀**
