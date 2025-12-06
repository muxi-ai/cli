# MUXI CLI Development Guide

**Project:** MUXI CLI  
**Status:** ✅ Complete - All Commands Implemented  
**Last Updated:** 2025-12-06

---

## MUXI Ecosystem

This repository is part of the larger MUXI ecosystem.

**📋 Complete architectural overview:** See [MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md) - explains how all 9 repositories fit together, dependencies, status, and roadmap.

**🎯 This repo (cli):** Command-line tool for formation scaffolding, secrets management, registry integration, and server management.

---

## Quick Start for New Sessions

### Current Status
- ✅ **Scaffolding:** All `muxi new` and `muxi config` commands complete
- ✅ **Secrets:** Full encryption/sync system working
- ✅ **Registry:** Full registry integration (login, push, pull, search, show)
- ✅ **Server:** Complete server/formation lifecycle commands
- ✅ **SSE Streaming:** Real-time progress for deploy, update, start, restart, rollback
- ✅ **Docs:** User guides for all features

### What's Working
1. ✅ All scaffolding wizards (formation, agent, mcp, sop, trigger, a2a-service)
2. ✅ All config commands (a2a, llm, memory, overlord, security, logging, async)
3. ✅ Secrets management (list, set, delete, setup, sync)
4. ✅ Registry commands (login, logout, push, pull, search, show, mine)
5. ✅ Server management (add, list, default, remove, status, ping)
6. ✅ Formation lifecycle (deploy, list, get, stop, start, restart, rollback, delete, logs)
7. ✅ SSE streaming with progress stages
8. ✅ Ctrl+C cancellation with cleanup
9. ✅ Notification sounds on completion
10. ✅ Version validation before update

---

## Architecture Overview

### CLI → Server Communication

```
User Command
    ↓
muxi formation deploy bundle.tar.gz
    ↓
CLI (Go binary)
    ↓
HMAC Request Signing
    ↓
HTTP POST → Server :7890/rpc/formations/deploy
    ↓
Server validates, deploys, responds
    ↓
CLI shows progress/result
```

### Profile Management

```yaml
# ~/.muxi/profiles.yaml
profiles:
  localhost:
    url: http://localhost:7890
    auth:
      key_id: "auto-detected"
      secret_key: "auto-detected"  # Read from ~/.muxi/server/credentials.json
    default: true
    
  production:
    url: https://muxi.company.com:7890
    auth:
      key_id: "MUXI_PROD_KEY"
      secret_key: "MUXI_PROD_SECRET"
    default: false
```

### Auto-Detection Flow

```go
// On first run
func EnsureProfile() error {
    if ProfilesExist() {
        return nil
    }
    
    // Detect local server
    server := DetectLocalServer()
    
    if server.HasCredentials {
        // Auto-populate from ~/.muxi/server/credentials.json
        creds := ReadServerCredentials()
        CreateLocalProfile(creds)
    }
}
```

---

## Server API Reference

The CLI talks to the MUXI Server API. Here's what you need to know:

### Base URL
```
http://localhost:7890  (local)
https://your-server.com:7890  (remote)
```

### Authentication

**HMAC Authentication Required for `/rpc/*` endpoints**

```go
// HMAC Signature Generation
func GenerateHMACSignature(keyID, secretKey, method, path, timestamp, body string) string {
    // Canonical string
    canonical := fmt.Sprintf("%s\n%s\n%s\n%s", method, path, timestamp, body)
    
    // HMAC-SHA256
    mac := hmac.New(sha256.New, []byte(secretKey))
    mac.Write([]byte(canonical))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    return signature
}

// Request headers
headers := map[string]string{
    "X-MUXI-Key-ID": keyID,
    "X-MUXI-Timestamp": timestamp,  // RFC3339 format
    "X-MUXI-Signature": signature,
    "Content-Type": "application/json",  // or multipart/form-data for uploads
}
```

### Endpoints

#### 1. Deploy Formation
```
POST /rpc/formations/deploy
Content-Type: multipart/form-data

Fields:
  - formation_bundle: [file] .tar.gz file
  - metadata: [json] {"id": "my-formation", "name": "My Formation"}

Response 201:
{
  "formation_id": "my-formation",
  "version": "1.0.0",
  "status": "running",
  "port": 8001,
  "endpoints": {
    "health": "/api/my-formation/health",
    "chat": "/api/my-formation/chat"
  }
}
```

#### 2. List Formations
```
GET /rpc/formations

Response 200:
{
  "formations": [
    {
      "id": "my-formation",
      "name": "My Formation",
      "status": "running",
      "version": "1.0.0",
      "port": 8001,
      "uptime": 3600,
      "restarts": 0
    }
  ]
}
```

#### 3. Get Formation Details
```
GET /rpc/formations/{id}

Response 200:
{
  "id": "my-formation",
  "name": "My Formation",
  "status": "running",
  "version": "1.0.0",
  "port": 8001,
  "pid": 12345,
  "created_at": "2025-10-24T12:00:00Z",
  "updated_at": "2025-10-24T12:00:00Z",
  "uptime": 3600,
  "restarts": 0,
  "endpoints": [
    {"path": "/health", "method": "GET"},
    {"path": "/chat", "method": "POST"}
  ]
}
```

#### 4. Update Formation (Deploy New Version)
```
PUT /rpc/formations/{id}
Content-Type: multipart/form-data

Fields:
  - formation_bundle: [file] .tar.gz file

Response 200:
{
  "formation_id": "my-formation",
  "version": "1.0.1",
  "previous_version": "1.0.0",
  "status": "running",
  "message": "Formation updated successfully"
}
```

#### 5. Stop Formation
```
POST /rpc/formations/{id}/stop

Response 200:
{
  "formation_id": "my-formation",
  "status": "stopped",
  "message": "Formation stopped successfully"
}
```

#### 6. Restart Formation
```
POST /rpc/formations/{id}/restart

Response 200:
{
  "formation_id": "my-formation",
  "status": "running",
  "message": "Formation restarted successfully"
}
```

#### 7. Rollback Formation
```
POST /rpc/formations/{id}/rollback

Response 200:
{
  "formation_id": "my-formation",
  "version": "1.0.0",
  "previous_version": "1.0.1",
  "status": "running",
  "message": "Rolled back to previous version"
}
```

#### 8. Delete Formation
```
DELETE /rpc/formations/{id}

Response 200:
{
  "formation_id": "my-formation",
  "message": "Formation deleted successfully"
}
```

#### 9. Get Formation Logs
```
GET /rpc/formations/{id}/logs?lines=100&follow=false

Response 200:
{
  "formation_id": "my-formation",
  "logs": [
    "2025-10-24 12:00:00 INFO Starting formation...",
    "2025-10-24 12:00:01 INFO Formation ready"
  ]
}
```

#### 10. Server Status
```
GET /rpc/server/status

Response 200:
{
  "version": "0.20251024.3",
  "uptime": 86400,
  "formations": {
    "total": 5,
    "running": 4,
    "stopped": 1
  },
  "resources": {
    "cpu_percent": 15.5,
    "memory_used_mb": 512,
    "ports_allocated": 5
  }
}
```

#### 11. Server Audit Logs
```
GET /rpc/server/logs?lines=50

Response 200:
{
  "logs": [
    {
      "timestamp": "2025-10-24T12:00:00Z",
      "method": "POST",
      "path": "/rpc/formations/deploy",
      "key_id": "MUXI_LOCAL_...",
      "status": 201,
      "duration_ms": 150
    }
  ]
}
```

#### Public Endpoints (No Auth)
```
GET /health              → {"status": "healthy"}
GET /ping                → {"message": "pong"}
```

---

## CLI Command Reference (Implemented)

### Deploy Command
```bash
# Deploy from formation directory (auto-detects create vs update)
muxi deploy [--profile <name>] [--dry-run] [--no-stream]

# Flags:
#   --profile string   Server profile to use
#   --dry-run         Validate and create bundle without deploying
#   --no-stream       Disable SSE streaming progress
```

### Server Management
```bash
# Add server profile
muxi server add <name> --url <url> --key-id <id> --secret-key <key>

# List all servers (shows online/offline status)
muxi server list

# Set default server
muxi server default <name>

# Remove server profile
muxi server remove <name>

# Show server status
muxi server status [--profile <name>]

# Continuous ping with latency stats
muxi server ping [--profile <name>]
```

### Formation Lifecycle Commands
```bash
# List all formations
muxi formation list [--profile <name>]

# Get formation details
muxi formation get <id> [-v|--verbose] [--profile <name>]

# Stop a running formation
muxi formation stop <id> [-f|--force] [--profile <name>]

# Start a stopped formation (SSE streaming)
muxi formation start <id> [--profile <name>]

# Restart a formation (SSE streaming)
muxi formation restart <id> [-f|--force] [--profile <name>]

# Rollback to previous version (SSE streaming)
muxi formation rollback <id> [-f|--force] [--profile <name>]

# Delete a formation
muxi formation delete <id> [-f|--force] [--profile <name>]

# View formation logs
muxi formation logs <id> [-n|--lines <num>] [-f|--follow] [--stream stdout|stderr] [--profile <name>]
```

### Shortcut Commands (from formation directory)
```bash
muxi get [-v]              # Get current formation details
muxi stop [-f]             # Stop current formation
muxi start                 # Start current formation
muxi restart [-f]          # Restart current formation
muxi rollback [-f]         # Rollback current formation
muxi delete [-f]           # Delete current formation
muxi logs [-n 100] [-f]    # View logs (-f to follow)
```

### Scaffolding Commands
```bash
muxi new formation         # Create new formation with wizard
muxi new agent             # Add agent to formation
muxi new mcp               # Add MCP server
muxi new sop               # Add Standard Operating Procedure
muxi new trigger           # Add webhook trigger
muxi new a2a-service       # Add A2A service
```

### Configuration Commands
```bash
muxi config a2a            # Configure A2A communication
muxi config llm            # Configure LLM provider
muxi config memory         # Configure memory settings
muxi config overlord       # Configure overlord persona
muxi config security       # Configure security/credentials
muxi config logging        # Configure logging streams
muxi config async          # Configure async responses
muxi validate              # Validate formation
muxi edit <type>           # Open file in $EDITOR
```

### Secrets Commands
```bash
muxi secrets list [--with-values]   # List all secrets
muxi secrets set <name> [value]     # Set/update secret
muxi secrets delete <name>          # Delete secret
muxi secrets setup                  # Populate from template
muxi secrets sync [-i] [--dry-run]  # Sync with formation
```

### Registry Commands
```bash
muxi login [--registry <url>]       # Authenticate with registry
muxi logout [--registry <url>]      # Remove credentials
muxi push                           # Publish formation
muxi pull @user/formation           # Download formation
muxi search "query"                 # Search formations
muxi show @user/formation           # Show formation details
muxi registry mine                  # List your formations
muxi registry list                  # List configured registries
muxi registry add <name> <url>      # Add registry
muxi registry remove <name>         # Remove registry
muxi registry default <name>        # Set default registry
```

### Formation-Level Settings
```bash
muxi set server            # Set default server for formation
muxi set registry          # Set default registry for formation
```

### Utility Commands
```bash
muxi version               # Show CLI version
muxi help                  # Show help
muxi <command> --help      # Show command help
```

---

## Local Server Auto-Detection

### Detection Strategy

```go
type ServerInfo struct {
    Installed       bool
    ConfigPath      string
    CredentialsPath string
    Running         bool
    Port            int
}

func DetectLocalServer() *ServerInfo {
    info := &ServerInfo{Port: 7890}
    
    // 1. Check if binary exists
    if path, err := exec.LookPath("muxi-server"); err == nil {
        info.Installed = true
    }
    
    // 2. Check for config
    homeDir, _ := os.UserHomeDir()
    configPath := filepath.Join(homeDir, ".muxi/server/config.yaml")
    if fileExists(configPath) {
        info.ConfigPath = configPath
    }
    
    // 3. Check for credentials (BEST SIGNAL!)
    credsPath := filepath.Join(homeDir, ".muxi/server/credentials.json")
    if fileExists(credsPath) {
        info.CredentialsPath = credsPath
    }
    
    // 4. Check if running
    resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", info.Port))
    if err == nil && resp.StatusCode == 200 {
        info.Running = true
    }
    
    return info
}
```

### Credentials Auto-Read

```go
// Read server credentials for localhost profile
type Credentials struct {
    KeyID     string `json:"key_id"`
    SecretKey string `json:"secret_key"`
    ServerID  string `json:"server_id"`
    CreatedAt string `json:"created_at"`
}

func ReadServerCredentials() (*Credentials, error) {
    homeDir, _ := os.UserHomeDir()
    credsPath := filepath.Join(homeDir, ".muxi/server/credentials.json")
    
    data, err := os.ReadFile(credsPath)
    if err != nil {
        return nil, err
    }
    
    var creds Credentials
    if err := json.Unmarshal(data, &creds); err != nil {
        return nil, err
    }
    
    return &creds, nil
}
```

### First-Run Prompts

```go
func EnsureProfile() error {
    if ProfilesExist() {
        return nil
    }
    
    server := DetectLocalServer()
    
    if !server.Installed {
        fmt.Println("→ No MUXI Server detected.")
        fmt.Print("  Install locally? [Y/n]: ")
        // Offer to run: curl -sSL https://install.muxi.org | bash
        return nil
    }
    
    if !server.Running {
        fmt.Println("→ MUXI Server detected but not running.")
        fmt.Print("  Start server? [Y/n]: ")
        // Offer to run: muxi-server start
        return nil
    }
    
    // Server running - auto-configure
    creds, err := ReadServerCredentials()
    if err != nil {
        fmt.Println("→ MUXI Server running on port 7890")
        fmt.Print("  Add as default profile? [Y/n]: ")
        return nil
    }
    
    // Create localhost profile automatically
    CreateProfile("localhost", "http://localhost:7890", creds, true)
    fmt.Println("✓ Added localhost profile as default")
    
    return nil
}
```

---

## Development Workflow

### Phase 1: Foundation (Current - Blocked)
1. ⏸️ Wait for runtime API to finalize
2. ⏸️ Wait for formation bundle structure
3. ⏸️ Wait for formation YAML validation

### Phase 2: Core Implementation (Next)
1. Profile management (server add/list/use/remove)
2. HMAC authentication client
3. Formation list/get commands
4. Local server auto-detection

### Phase 3: Advanced Features
1. Formation deploy (with bundle upload)
2. Formation logs with --follow
3. Formation update/rollback
4. Server status/logs commands

### Phase 4: Polish
1. Pretty output formatting
2. Progress indicators
3. Error handling
4. Shell completions (bash/zsh/fish)

---

## Development Tips

### Code Search with ast-grep

For code pattern searches, prefer **ast-grep** over text-based grep when searching for:
- Function definitions and calls
- Struct fields and methods
- Import statements
- Variable declarations

**ast-grep** understands Go's AST (Abstract Syntax Tree), making searches more precise:

```bash
# Find all functions that return error
ast-grep -p 'func $NAME($$$) error' --lang go

# Find all struct definitions
ast-grep -p 'type $NAME struct { $$$ }' --lang go

# Find all calls to ui.PromptSuccess
ast-grep -p 'ui.PromptSuccess($$$)' --lang go

# Find all wizard.PromptString calls with 3 args
ast-grep -p 'wizard.PromptString($A, $B, $C)' --lang go
```

**When to use ast-grep vs text grep:**
- Use **ast-grep** for: code patterns, refactoring, finding implementations
- Use **text grep** for: error messages, comments, strings, config files

---

## Key Files & Structure

```
cli/
├── src/
│   ├── cmd/
│   │   ├── root.go              # Root command setup
│   │   ├── new.go               # `muxi new` subcommands
│   │   ├── config.go            # `muxi config` subcommands
│   │   ├── edit.go              # `muxi edit` command
│   │   ├── secrets.go           # `muxi secrets` subcommands
│   │   └── version.go           # Version command
│   │
│   └── pkg/
│       ├── scaffold/
│       │   ├── formation.go     # Formation wizard
│       │   ├── components.go    # Agent, MCP, A2A wizards
│       │   ├── sop.go           # SOP wizard
│       │   └── trigger.go       # Trigger wizard
│       │
│       ├── secrets/
│       │   └── secrets.go       # Fernet encryption, sync logic
│       │
│       ├── wizard/
│       │   └── wizard.go        # Prompt helpers (string, bool, select)
│       │
│       ├── ui/
│       │   └── ui.go            # TUI design system (colors, symbols)
│       │
│       └── config/
│           └── config.go        # Config loading/saving
│
├── docs/
│   ├── guides/                  # User guides
│   │   ├── formations.md
│   │   ├── agents.md
│   │   ├── mcps.md
│   │   ├── sops.md
│   │   ├── triggers.md
│   │   ├── a2a.md
│   │   └── secrets.md
│   │
│   ├── plan-registry.md         # Registry commands plan
│   ├── plan-config-llm.md       # LLM config plan
│   ├── plan-config-memory.md    # Memory config plan
│   ├── plan-config-overlord.md  # Overlord config plan
│   ├── plan-config-logging.md   # Logging config plan
│   ├── UX-PATTERNS.md           # TUI/UX conventions
│   └── BANNERS.md               # Banner design reference
│
├── AGENTS.md                    # This file
├── DESIGN.md                    # CLI architecture & config
├── STATUS.md                    # Current status
├── README.md                    # Project overview
└── go.mod                       # Dependencies
```

---

## Dependencies (Go Modules)

```go
// Minimal dependencies
require (
    github.com/spf13/cobra v1.8.0      // CLI framework
    gopkg.in/yaml.v3 v3.0.1            // YAML parsing
    github.com/fatih/color v1.16.0     // Colored output
)
```

---

## Server Credentials File Format

**Location:** `~/.muxi/server/credentials.json`

```json
{
  "key_id": "MUXI_LOCAL_abc123",
  "secret_key": "base64-encoded-secret-key-here",
  "server_id": "hostname-sha256hash",
  "created_at": "2025-10-24T12:00:00Z"
}
```

**CLI reads this automatically for localhost profile!**

---

## Error Handling Strategy

### HTTP Errors
```go
type APIError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

func HandleAPIError(resp *http.Response) error {
    var apiErr APIError
    json.NewDecoder(resp.Body).Decode(&apiErr)
    
    switch resp.StatusCode {
    case 401:
        return fmt.Errorf("authentication failed: %s", apiErr.Message)
    case 404:
        return fmt.Errorf("formation not found: %s", apiErr.Message)
    case 409:
        return fmt.Errorf("conflict: %s", apiErr.Message)
    default:
        return fmt.Errorf("server error: %s", apiErr.Message)
    }
}
```

### User-Friendly Messages
```go
// Bad
fmt.Println("Error: formation not found")

// Good
fmt.Printf("✗ Formation '%s' not found\n", formationID)
fmt.Println("  Use 'muxi formation list' to see available formations")
```

---

## Testing Strategy

### Unit Tests
```go
// Test HMAC signature generation
func TestHMACSignature(t *testing.T) {
    keyID := "test-key"
    secret := "test-secret"
    method := "POST"
    path := "/rpc/formations/deploy"
    timestamp := "2025-10-24T12:00:00Z"
    body := `{"id":"test"}`
    
    sig := GenerateHMACSignature(keyID, secret, method, path, timestamp, body)
    
    // Verify signature format
    assert.NotEmpty(t, sig)
}
```

### Integration Tests
```bash
# Test against local server
go test -v -tags=integration ./...

# Requires:
# - Local server running
# - Test formations available
```

---

## Blockers & Dependencies

### Current Blockers
1. **Runtime API** - Formation endpoints (/health, /chat, etc.)
2. **Formation Bundle** - Structure and contents (.tar.gz format)
3. **YAML Validation** - Complete schema for formation.yaml

### What We Know
- ✅ Server API is ready (documented above)
- ✅ HMAC auth is documented
- ✅ Profile management strategy is clear
- ✅ Auto-detection logic is designed

### When Unblocked
1. Implement profile management (1-2 days)
2. Implement HMAC client (1 day)
3. Implement formation list/get (1 day)
4. Implement formation deploy (2-3 days)
5. Polish and test (1-2 days)

**Total: ~1 week of focused development once unblocked**

---

## Quick Reference: Server Info

| Item | Value |
|------|-------|
| **Default Port** | 7890 |
| **Auth Type** | HMAC-SHA256 |
| **Management API** | `/rpc/*` |
| **Formation Proxy** | `/api/{id}/*` |
| **Config Dir** | `~/.muxi/server/` |
| **Credentials** | `~/.muxi/server/credentials.json` |
| **Registry** | `~/.muxi/server/registry.json` |

---

## Common Development Tasks

### Test HMAC Authentication
```bash
# Generate HMAC signature manually
./test-hmac.sh

# Test against local server
curl -X GET http://localhost:7890/rpc/formations \
  -H "X-MUXI-Key-ID: $KEY_ID" \
  -H "X-MUXI-Timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  -H "X-MUXI-Signature: $SIGNATURE"
```

### Debug Server Communication
```bash
# Check server health
curl http://localhost:7890/health

# Check if authenticated
curl -X GET http://localhost:7890/rpc/formations \
  -H "X-MUXI-Key-ID: wrong" \
  -H "X-MUXI-Timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  -H "X-MUXI-Signature: wrong"
# Should return 401
```

### Test Profile Detection
```bash
# Check if server is installed
which muxi-server

# Check for credentials
ls -la ~/.muxi/server/credentials.json

# Check if server is running
curl http://localhost:7890/health
```

---

## Useful Links

### Local Documentation (This Repo)
- **[DESIGN.md](DESIGN.md)** - CLI architecture & config structure ⭐
- **[STATUS.md](STATUS.md)** - Current development status
- **[docs/UX-PATTERNS.md](docs/UX-PATTERNS.md)** - Complete UX design patterns guide
- **[docs/BANNERS.md](docs/BANNERS.md)** - Banner reference with MUXI branding
- **[docs/plan-registry.md](docs/plan-registry.md)** - Registry commands implementation plan
- **[docs/plan-config-*.md](docs/)** - Config command implementation plans

### User Guides (docs/guides/)
- **formations.md** - Creating formations, LLM providers
- **agents.md** - Agent roles, system prompts
- **mcps.md** - MCP server configuration
- **sops.md** - Standard Operating Procedures
- **triggers.md** - Webhook triggers
- **a2a.md** - Agent-to-Agent communication
- **secrets.md** - Secrets management

### Key Implementation Files
- **src/pkg/scaffold/** - All scaffolding wizards
- **src/pkg/secrets/secrets.go** - Fernet encryption, sync logic
- **src/pkg/wizard/wizard.go** - Prompt implementation
- **src/pkg/ui/ui.go** - TUI design system

### External Repos
- **Server Repo:** [github.com/muxi-ai/server](https://github.com/muxi-ai/server)
- **Runtime Repo:** [github.com/muxi-ai/runtime](https://github.com/muxi-ai/runtime)
- **Registry Repo:** [github.com/muxi-ai/registry](https://github.com/muxi-ai/registry)
- **Architecture:** [MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md)

---

## Session Checklist

Starting a new session? Check these:

- [ ] Read this document (AGENTS.md)
- [ ] Review [DESIGN.md](DESIGN.md) for CLI architecture & config structure
- [ ] Review [docs/UX-PATTERNS.md](docs/UX-PATTERNS.md) for established patterns
- [ ] Check existing wizards in `src/pkg/scaffold/`
- [ ] Follow ID normalization patterns (spaces → hyphens, auto-suggest name)
- [ ] Use validation loops (not exits on invalid input)
- [ ] Handle Ctrl+C gracefully at all prompts

**Ready to build!** 🚀

---

**Last Updated:** 2025-11-30  
**Status:** Scaffolding & Secrets Complete ✅  
**Next Step:** Config commands (LLM, memory, overlord, logging) and Registry commands
