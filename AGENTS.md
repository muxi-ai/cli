# MUXI CLI Development Guide

**Project:** MUXI CLI  
**Status:** Early WIP - Blocked on Runtime API Finalization  
**Last Updated:** 2025-10-24

---

## MUXI Ecosystem

This repository is part of the larger MUXI ecosystem.

**📋 Complete architectural overview:** See [MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md) - explains how all 9 repositories fit together, dependencies, status, and roadmap.

**🎯 This repo (cli):** Command-line tool for formation management - remote server management, local server auto-detection, HMAC authentication.

---

## Quick Start for New Sessions

### Current Status
- ⏸️ **BLOCKED:** Waiting for runtime API to finalize
- 🎯 **When unblocked:** Can build full formation management
- 📋 **Server API:** Ready and documented below
- 🔐 **Auth:** HMAC implementation strategy documented

### What's Ready
1. ✅ Server API is production-ready (see below)
2. ✅ HMAC authentication design documented
3. ✅ Schemas for formation validation
4. ✅ Installation flow designed (auto-detection)

### What's Needed
1. ⏳ Runtime API contract (formation endpoints)
2. ⏳ Formation bundle structure finalized
3. ⏳ Formation YAML validation complete

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

## CLI Command Structure (Planned)

### Profile Management
```bash
# Add server
muxi server add <profile-name> --url <url>
muxi server add production --url https://muxi.company.com:7890

# Set default
muxi server use <profile-name>
muxi server use production

# List servers
muxi server list

# Remove server
muxi server remove <profile-name>

# Show current
muxi server current
```

### Formation Management
```bash
# Deploy formation
muxi formation deploy <bundle.tar.gz>
muxi formation deploy my-formation.tar.gz

# List formations
muxi formation list
muxi formation list --profile production

# Get details
muxi formation get <id>
muxi formation get my-formation

# Update formation
muxi formation update <id> <bundle.tar.gz>
muxi formation update my-formation my-formation-v2.tar.gz

# Stop/Start/Restart
muxi formation stop <id>
muxi formation start <id>  # If stopped
muxi formation restart <id>

# Rollback
muxi formation rollback <id>

# Delete
muxi formation delete <id>

# Logs
muxi formation logs <id>
muxi formation logs <id> --follow
muxi formation logs <id> --lines 1000
```

### Server Management
```bash
# Server status
muxi server status
muxi server status --profile production

# Server logs
muxi server logs
muxi server logs --lines 100
```

### Utility Commands
```bash
# Version
muxi version

# Help
muxi help
muxi formation help
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
├── cmd/
│   ├── root.go              # Root command setup
│   ├── server.go            # Server management commands
│   ├── formation.go         # Formation management commands
│   └── version.go           # Version command
│
├── pkg/
│   ├── client/
│   │   ├── client.go        # HTTP client with HMAC
│   │   └── auth.go          # HMAC signature generation
│   │
│   ├── profile/
│   │   ├── profile.go       # Profile struct & management
│   │   └── detect.go        # Local server auto-detection
│   │
│   ├── api/
│   │   ├── formations.go    # Formation API calls
│   │   └── server.go        # Server API calls
│   │
│   └── config/
│       └── config.go        # Load ~/.muxi/profiles.yaml
│
├── README.md                # Distribution guide
├── AGENTS.md               # This file
└── go.mod                  # Dependencies
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
- **[docs/UX-PATTERNS.md](docs/UX-PATTERNS.md)** - Complete UX design patterns guide ⭐
  - Validation loops, error handling
  - URL normalization, ID normalization
  - Menu selections, multi-line prompts
  - Natural language patterns
  - Secret management, state management
- **[docs/CLI-COMMAND-DESIGN.md](docs/CLI-COMMAND-DESIGN.md)** - Command structure and semantics
- **[docs/COMMAND-SEMANTICS.md](docs/COMMAND-SEMANTICS.md)** - `new` vs `config` command rationale
- **[docs/TUI-DESIGN.md](docs/TUI-DESIGN.md)** - Visual patterns (banners, colors, symbols)
- **[docs/A2A-WIZARD.md](docs/A2A-WIZARD.md)** - A2A configuration wizard guide
- **[docs/BANNERS.md](docs/BANNERS.md)** - Banner reference with MUXI branding

### Key Implementation Files
- **src/pkg/scaffold/components.go** - All wizards (agent, MCP, A2A)
- **src/pkg/wizard/wizard.go** - Prompt implementation
- **src/pkg/ui/ui.go** - TUI design system

### External Repos
- **Server Repo:** [github.com/muxi-ai/server](https://github.com/muxi-ai/server)
- **Runtime Repo:** [github.com/muxi-ai/runtime](https://github.com/muxi-ai/runtime)
- **Schemas Repo:** [github.com/muxi-ai/schemas](https://github.com/muxi-ai/schemas)
- **Install Repo:** [github.com/muxi-ai/install](https://github.com/muxi-ai/install)
- **Architecture:** [MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md)

---

## Session Checklist

Starting a new session? Check these:

- [ ] Read this document (AGENTS.md)
- [ ] Review [docs/UX-PATTERNS.md](docs/UX-PATTERNS.md) for established patterns
- [ ] Review [MUXI-ARCHITECTURE.md](../MUXI-ARCHITECTURE.md) for ecosystem context
- [ ] Check existing wizards in `src/pkg/scaffold/components.go`
- [ ] Follow ID normalization patterns (spaces → hyphens, auto-suggest name)
- [ ] Use validation loops (not exits on invalid input)
- [ ] Handle Ctrl+C gracefully at all prompts

**Ready to build!** 🚀

---

**Last Updated:** 2025-11-28  
**Status:** A2A Scaffolding Complete ✅  
**Next Step:** More config commands (LLM, observability, security)
