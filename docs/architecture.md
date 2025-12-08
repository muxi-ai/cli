# CLI Architecture

**Last Updated:** 2025-12-08

---

## Overview

The MUXI CLI is a Go-based command-line tool that communicates with MUXI Server instances to manage formations.

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

---

## Directory Structure

```
cli/
├── AGENTS.md                # Agent workflow guide
├── AGENT_CHAT.md            # Multi-agent coordination
├── STATUS.md                # Current status, what's working
├── README.md                # User-facing documentation
│
├── src/
│   ├── cmd/                 # Command implementations
│   │   ├── root.go         # Root command, global flags
│   │   ├── new.go          # muxi new <type>
│   │   ├── config.go       # muxi config <type>
│   │   ├── edit.go         # muxi edit <type>
│   │   ├── secrets.go      # muxi secrets <action>
│   │   ├── deploy.go       # muxi deploy
│   │   ├── formation.go    # muxi formation <action>
│   │   ├── server.go       # muxi server <action>
│   │   ├── registry.go     # muxi registry <action>
│   │   ├── version.go      # muxi version
│   │   │
│   │   │   # Formation API commands
│   │   ├── agents.go       # muxi agents (list/manage agents)
│   │   ├── mcp.go          # muxi mcp (list/manage MCP servers)
│   │   ├── info.go         # muxi info (formation status)
│   │   ├── triggers.go     # muxi triggers (list triggers)
│   │   ├── sops.go         # muxi sops (list SOPs)
│   │   ├── sessions.go     # muxi sessions (list sessions)
│   │   ├── history.go      # muxi history (session messages)
│   │   ├── clear.go        # muxi clear (clear session)
│   │   ├── trigger.go      # muxi trigger (execute trigger)
│   │   ├── jobs.go         # muxi jobs (async jobs)
│   │   ├── audit.go        # muxi audit (audit log)
│   │   ├── stream.go       # muxi stream (log streaming)
│   │   ├── scheduler.go    # muxi scheduler (scheduled jobs)
│   │   ├── users.go        # muxi users (identity management)
│   │   └── memory.go       # muxi memory (memory ops)
│   │
│   ├── pkg/
│   │   ├── scaffold/       # Scaffolding wizards
│   │   │   ├── formation.go
│   │   │   ├── components.go
│   │   │   ├── sop.go
│   │   │   └── trigger.go
│   │   │
│   │   ├── secrets/        # Fernet encryption, sync
│   │   │   └── secrets.go
│   │   │
│   │   ├── wizard/         # Interactive prompts
│   │   │   └── wizard.go
│   │   │
│   │   ├── ui/             # TUI design system
│   │   │   └── ui.go
│   │   │
│   │   ├── client/         # Server API client
│   │   │   ├── client.go   # HTTP client, HMAC auth
│   │   │   └── sse.go      # SSE streaming
│   │   │
│   │   ├── formation/      # Formation API client
│   │   │   ├── client.go   # HTTP client with admin/client key auth
│   │   │   ├── auth.go     # API key resolution (env vars, secrets.enc)
│   │   │   ├── types.go    # Response types for all endpoints
│   │   │   └── flags.go    # Common flag helpers (-F, -p, -u)
│   │   │
│   │   ├── config/         # Config management
│   │   │   └── config.go
│   │   │
│   │   └── bundle/         # Formation bundling
│   │       └── bundle.go
│   │
│   ├── go.mod
│   └── main.go
│
└── docs/
    ├── architecture.md      # This file
    ├── api-reference.md     # Server API, HMAC auth
    ├── UX-PATTERNS.md       # TUI design patterns
    ├── BANNERS.md           # ASCII banner reference
    ├── DESIGN.md            # Detailed design docs
    └── guides/              # User guides
        ├── formations.md
        ├── agents.md
        ├── mcps.md
        └── ...
```

---

## Command Categories

### Scaffolding (`muxi new`)
Create new components with interactive wizards:
- `muxi new formation` - Full formation wizard
- `muxi new agent` - Add agent to formation
- `muxi new mcp` - Add MCP server
- `muxi new sop` - Add SOP
- `muxi new trigger` - Add webhook trigger
- `muxi new a2a-service` - Add A2A service

### Configuration (`muxi config`)
Configure formation settings:
- `muxi config llm` - LLM provider settings
- `muxi config memory` - Memory configuration
- `muxi config overlord` - Overlord persona
- `muxi config security` - Credentials
- `muxi config logging` - Log streams
- `muxi config async` - Async responses
- `muxi config a2a` - A2A communication

### Secrets (`muxi secrets`)
Encrypted secrets management:
- `muxi secrets list` - List secrets
- `muxi secrets set` - Set/update secret
- `muxi secrets delete` - Remove secret
- `muxi secrets setup` - Populate from template
- `muxi secrets sync` - Sync with formation

### Server Management (`muxi server`)
Manage server profiles:
- `muxi server add` - Add server profile
- `muxi server list` - List servers
- `muxi server default` - Set default
- `muxi server remove` - Remove profile
- `muxi server status` - Show status
- `muxi server ping` - Continuous ping

### Formation Lifecycle (`muxi formation`)
Manage deployed formations:
- `muxi deploy` - Deploy formation
- `muxi formation list` - List formations
- `muxi formation get` - Get details
- `muxi formation stop` - Stop formation
- `muxi formation start` - Start formation
- `muxi formation restart` - Restart formation
- `muxi formation rollback` - Rollback version
- `muxi formation delete` - Delete formation
- `muxi formation logs` - View logs

### Formation API Commands
Interact with running formations via Formation API:
- `muxi agents` - List/manage agents in formation
- `muxi mcp` - List/manage MCP servers
- `muxi info` - Formation runtime status
- `muxi triggers` - List available triggers
- `muxi sops` - List SOPs
- `muxi sessions` - List user sessions
- `muxi history` - View session messages
- `muxi clear` - Clear session/buffer
- `muxi trigger` - Execute a trigger
- `muxi jobs` - List/cancel async jobs
- `muxi audit` - View audit log
- `muxi stream` - Live log streaming
- `muxi scheduler` - Manage scheduled jobs
- `muxi users` - User identity management
- `muxi memory` - Memory operations

### Registry (`muxi registry`)
Formation distribution:
- `muxi login/logout` - Registry auth
- `muxi push/pull` - Publish/download
- `muxi search/show` - Discovery

---

## Profile Management

Profiles are stored in `~/.muxi/profiles.yaml`:

```yaml
profiles:
  localhost:
    url: http://localhost:7890
    auth:
      key_id: "auto-detected"
      secret_key: "auto-detected"
    default: true
    
  production:
    url: https://muxi.company.com:7890
    auth:
      key_id: "MUXI_PROD_KEY"
      secret_key: "MUXI_PROD_SECRET"
    default: false
```

### Auto-Detection

On first run, CLI auto-detects local server credentials from `~/.muxi/server/credentials.json`:

```go
func EnsureProfile() error {
    if ProfilesExist() {
        return nil
    }
    
    server := DetectLocalServer()
    if server.HasCredentials {
        creds := ReadServerCredentials()
        CreateLocalProfile(creds)
    }
}
```

---

## SSE Streaming

Long-running operations (deploy, start, restart, rollback) use Server-Sent Events for real-time progress:

```go
// SSE event structure
type ProgressEvent struct {
    Stage    string  `json:"stage"`
    Message  string  `json:"message"`
    Progress float64 `json:"progress,omitempty"`
}

// Example stages for deploy
stages := []string{
    "extracting",
    "validating", 
    "resolving_runtime",
    "downloading_sif",
    "pulling_runner",
    "spawning",
    "health_check",
}
```

---

## Error Handling

### API Errors
```go
type APIError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

// Handle with user-friendly messages
switch resp.StatusCode {
case 401:
    return fmt.Errorf("authentication failed: %s", apiErr.Message)
case 404:
    return fmt.Errorf("formation not found: %s", apiErr.Message)
case 409:
    return fmt.Errorf("conflict: %s", apiErr.Message)
}
```

### User-Friendly Output
```go
// Bad
fmt.Println("Error: formation not found")

// Good
fmt.Printf("✗ Formation '%s' not found\n", formationID)
fmt.Println("  Use 'muxi formation list' to see available formations")
```

---

## Testing

```bash
# Build
cd src && go build ./...

# Unit tests
cd src && go test ./...

# Integration tests (requires running server)
cd src && go test -v -tags=integration ./...
```

---

## Key Patterns

### Wizard Prompts
All wizards use `pkg/wizard/wizard.go`:
```go
name := wizard.PromptString("Formation name", "my-formation", nil)
useLLM := wizard.PromptBool("Configure LLM?", true)
provider := wizard.PromptSelect("Provider", providers, 0)
```

### UI Output
All output uses `pkg/ui/ui.go`:
```go
ui.PromptSuccess("Formation created")
ui.PromptError("Failed to deploy")
ui.PromptWarning("No secrets configured")
ui.PromptInfo("Using default server")
```

### Ctrl+C Handling
All operations support graceful cancellation:
```go
ctx, cancel := context.WithCancel(context.Background())
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt)

go func() {
    <-sigChan
    cancel()
}()
```

---

## Related Documentation

- [api-reference.md](api-reference.md) - HMAC auth, endpoints
- [UX-PATTERNS.md](UX-PATTERNS.md) - TUI conventions
- [DESIGN.md](DESIGN.md) - Detailed design decisions
- [../schemas/api/](../../schemas/api/) - OpenAPI specs
