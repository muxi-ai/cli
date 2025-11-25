# MUXI CLI - Complete Design

**Status:** Ready to Build  
**Timeline:** 1 Week  
**Date:** 2025-11-25

---

## Overview

Command-line tool for MUXI formation development and server management. Supports local development, multi-server deployment, and complete formation lifecycle management.

**Key Features:**
- 🏗️ Local formation development (init, validate, secrets)
- 🚀 Multi-server deployment (deploy to all servers in profile)
- 🔐 Smart secrets management (validation, sync, wizard)
- 🌐 Registry integration (push/pull formations)
- ⚙️ Complete Formation API coverage (~80 commands)
- 🖥️ Server operations (lifecycle, monitoring)

---

## Configuration Architecture

### 4 Configuration Files

```
~/.muxi/cli/
├── config.yaml          # Global CLI settings
├── servers.yaml         # Server profiles (HMAC auth)
├── formations.yaml      # Formation credentials (API keys)
└── registries.yaml      # Registry auth tokens
```

### 1. config.yaml - Global Settings

```yaml
default_profile: localhost
default_registry: registry.muxi.org
output_format: text
no_color: false
debug: false
```

### 2. servers.yaml - Server Profiles

**Multi-server support built-in:**

```yaml
version: "1.0"
default_profile: production

profiles:
  production:
    servers:
      - id: us-east-1
        url: https://east.company.com:7890
        auth:
          key_id: MUXI_EAST_KEY
          secret_key: sk_east_...
      
      - id: us-west-1
        url: https://west.company.com:7890
        auth:
          key_id: MUXI_WEST_KEY
          secret_key: sk_west_...
  
  localhost:
    servers:
      - id: default
        url: http://localhost:7890
        auth:
          key_id: MUXI_LOCAL_abc123
          secret_key: sk_local_...
```

### 3. formations.yaml - Formation Credentials

**NO URLs** (all via server):

```yaml
version: "1.0"

formations:
  my-bot:
    admin_key: fma_abc123...
    client_key: fmc_xyz789...
    added_at: "2025-11-25T12:00:00Z"
    notes: "Production chat bot"
```

### 4. registries.yaml - Registry Auth

```yaml
version: "1.0"
default_registry: registry.muxi.org

registries:
  registry.muxi.org:
    token: mxr_5Jw9k2Lp8Nm4Qr6Ts1Uv3Wx7Yz0...
    username: ranaroussi
    created_at: "2025-11-25T10:30:00Z"
```

### 5. .muxi - Per-Formation Defaults (Optional)

```yaml
# Only 2 fields!
profile: production
registry: private.company.com
```

---

## Profile Resolution

### 3-Tier Hierarchy

**Priority (highest to lowest):**
1. `--profile` flag (explicit)
2. `.muxi` file (project default, if in formation dir)
3. Global `default_profile` (from servers.yaml)

**Example:**
```bash
# 1. Explicit flag (highest)
muxi deploy --profile staging  # Uses staging

# 2. .muxi file
cd my-formation/
cat .muxi
# profile: production
muxi deploy  # Uses production

# 3. Global default
cd my-formation/
rm .muxi
muxi deploy  # Uses global default from servers.yaml
```

---

## Connection Model

### CLI is a SERVER Tool

**Key Decision:** No direct formation connections. All communication via server.

```
CLI → Server :7890/rpc/*              (Server API - lifecycle)
CLI → Server :7890/api/{id}/v1/*      (Formation API - proxied)
```

**Benefits:**
- ✅ Simpler code (one connection model)
- ✅ Better security (formations not exposed)
- ✅ Consistent authentication

---

## Secrets Management

### Local Secrets (Runtime-Compatible)

**Port encryption from runtime:**
- Algorithm: AES-256-GCM
- Key: 32-byte random (`.key` file)
- Format: Encrypted JSON (`secrets.enc`)

**Smart Validation:**
```bash
muxi secrets delete OPENAI_API_KEY

✗ Cannot delete secret 'OPENAI_API_KEY'

  This secret is referenced in:
    • formation.yaml (line 52)
    • agents/research-agent.yaml (line 18)
  
  Remove references first, then delete.
```

**Auto-sync with secrets.example:**
```bash
muxi secrets set NEW_KEY

  Enter value: ••••••••
  ✓ Secret 'NEW_KEY' saved

⚠ secrets.example is out of sync

  Add 'NEW_KEY' to secrets.example? [Y/n]: y
  ✓ Updated secrets.example
```

**Setup Wizard:**
```bash
muxi secrets setup

Analyzing formation...

Found 5 secret references:
  • FORMATION_ADMIN_API_KEY ✓ (set)
  • FORMATION_CLIENT_API_KEY ✓ (set)
  • OPENAI_API_KEY ✗ (not set)
  • DATABASE_URL ✗ (not set)
  • SLACK_WEBHOOK ✗ (optional)

Setting up 2 required secrets...

[1/2] OPENAI_API_KEY
  Used in: formation.yaml (line 52)
  Enter value: ••••••••
  ✓ Set

[2/2] DATABASE_URL
  Used in: mcps/postgres.yaml (line 12)
  Enter value: ••••••••
  ✓ Set

✓ Secrets configured
```

**Validation Rules:**
- ✅ Cannot delete secrets referenced in formation files
- ✅ Validates all references during `muxi validate`
- ✅ Keeps secrets.example in sync
- ✅ Auto-generates ADMIN_KEY and CLIENT_KEY on `muxi init`

---

## Multi-Server Deployment

**Deploy to ALL servers in profile:**

```bash
muxi deploy --profile production

Deploying to profile 'production' (3 servers)...

[1/3] us-east-1 (https://east.company.com:7890)
  ⠋ Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

[2/3] us-west-1 (https://west.company.com:7890)
  ⠋ Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

[3/3] eu-central-1 (https://eu.company.com:7890)
  ⠋ Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

✓ Deployed to 3/3 servers successfully
```

---

## Registry Integration

### Default Registry: registry.muxi.org

**Public formations:** No auth required (like GitHub)

```bash
muxi pull @user/public-formation
⠋ Fetching formation...
✓ Downloaded @user/public-formation v1.0.0
```

**Private formations:** Registry returns auth error → CLI prompts

```bash
muxi pull @user/private-formation
⠋ Fetching formation...
✗ Authentication required

  Login to registry.muxi.org? [Y/n]: y
  ⠋ Opening browser...
  ✓ Authenticated as ranaroussi

⠋ Fetching formation...
✓ Downloaded @user/private-formation v1.0.0
```

---

## Error Messages & UX

### Interactive Prompts

**No profiles configured:**
```bash
muxi deploy

✗ No server profiles configured

  Add one now? [Y/n]: y

Server profile setup:
  Profile name: production
  Server URL: https://muxi.company.com:7890
  HMAC Key ID: MUXI_PROD_KEY
  HMAC Secret: ••••••••
  
  ⠋ Testing connection...
  ✓ Connection successful
  
  Set as default? [Y/n]: y
  ✓ Profile 'production' added

Continue with deployment? [Y/n]: y
⠋ Deploying formation...
```

### Animated Indicators

**Spinners:**
```bash
⠋ Validating formation...
⠙ Creating bundle...
⠹ Uploading to production...
✓ Formation 'my-bot' deployed
```

**Progress bars:**
```bash
Uploading to registry...
████████████████████████████████ 100% | 2.3 MB | 1.2 MB/s
```

---

## Command Structure

### Core Commands (Priority 1 - Day 1-2)

```bash
# Formation Development
muxi init <name>                    # Create formation
muxi validate                       # Validate files
muxi deploy [--profile <name>]      # Deploy to server(s)

# Secrets Management (Local)
muxi secrets set <key>              # Set secret (prompt)
muxi secrets list                   # List keys only
muxi secrets delete <key>           # Delete with validation
muxi secrets setup                  # Setup wizard

# Profile Management
muxi profile add <name>             # Add profile (wizard)
muxi profile list                   # List all profiles
muxi profile use <name>             # Set default
muxi profile remove <name>          # Remove profile

# Registry
muxi login [registry]               # Auth (default: registry.muxi.org)
muxi logout [registry]              # Logout
muxi push [--registry <url>]        # Publish formation
muxi pull <ref> [--registry <url>]  # Download formation
muxi search <query> [--registry <url>] # Search formations
```

### Server Operations (Priority 2 - Day 3)

```bash
# Formation Lifecycle (Server API)
muxi formation list [--profile <name>]
muxi formation stop <id> [--profile <name>]
muxi formation restart <id> [--profile <name>]
muxi formation rollback <id> [--profile <name>]
muxi formation delete <id> [--profile <name>]

# Server Management
muxi server status [--profile <name>]
muxi server logs [--profile <name>]
muxi server ping [--profile <name>]
```

### Formation Configuration (Priority 3 - Day 4-5)

```bash
# Status & Config
muxi status [--formation <id>] [--profile <name>]
muxi config show [--formation <id>] [--profile <name>]

# Agents
muxi agent list [--formation <id>] [--profile <name>]
muxi agent add [--formation <id>] [--profile <name>] [--file <yaml>]
muxi agent get <id> [--formation <id>] [--profile <name>]
muxi agent update <id> [--formation <id>] [--profile <name>]
muxi agent delete <id> [--formation <id>] [--profile <name>]

# MCPs
muxi mcp list [--formation <id>] [--profile <name>]
muxi mcp add [--formation <id>] [--profile <name>] [--file <yaml>]
muxi mcp get <id> [--formation <id>] [--profile <name>]
muxi mcp update <id> [--formation <id>] [--profile <name>]
muxi mcp delete <id> [--formation <id>] [--profile <name>]

# Chat & Sessions
muxi chat [--formation <id>] [--profile <name>]
muxi session list [--formation <id>] [--profile <name>]
muxi session get <id> [--formation <id>] [--profile <name>]
muxi session messages <id> [--formation <id>] [--profile <name>]
muxi session delete <id> [--formation <id>] [--profile <name>]

# Triggers & Jobs
muxi trigger list [--formation <id>] [--profile <name>]
muxi trigger get <name> [--formation <id>] [--profile <name>]
muxi trigger invoke <name> [--formation <id>] [--profile <name>]
muxi job list [--formation <id>] [--profile <name>]
muxi job cancel <id> [--formation <id>] [--profile <name>]

# Monitoring
muxi logs [--formation <id>] [--profile <name>] [--follow]
muxi audit list [--formation <id>] [--profile <name>]
muxi audit clear [--formation <id>] [--profile <name>]

# SOPs (Read-only)
muxi sop list [--formation <id>] [--profile <name>]
muxi sop get <name> [--formation <id>] [--profile <name>]
```

**See [COMPLETE-COMMAND-REFERENCE.md](./COMPLETE-COMMAND-REFERENCE.md) for all ~80 commands**

---

## Implementation Details

### Package Structure

```
pkg/
├── config/             # Configuration management
│   ├── config.go       # Load/save all config files
│   └── defaults.go     # Default values
│
├── client/             # API clients
│   ├── server.go       # Server API (HMAC auth)
│   ├── formation.go    # Formation API (API keys)
│   ├── registry.go     # Registry API (Bearer token)
│   └── resolver.go     # Connection resolution
│
├── secrets/            # Secrets management
│   ├── crypto.go       # AES-256-GCM encryption
│   ├── store.go        # Load/save secrets
│   └── validate.go     # Reference validation
│
├── context/            # Formation context detection
│   └── formation.go    # Walk up directory tree
│
├── wizard/             # Interactive prompts
│   └── wizard.go       # Confirm, string, password, select
│
└── ui/                 # User interface
    ├── spinner.go      # Animated spinners
    ├── progress.go     # Progress bars
    └── format.go       # Output formatting

cmd/
├── root.go             # Root command + global flags
├── init.go             # muxi init
├── validate.go         # muxi validate
├── deploy.go           # muxi deploy
├── secrets.go          # muxi secrets *
├── profile.go          # muxi profile *
├── formation.go        # muxi formation *
├── server.go           # muxi server *
├── registry.go         # muxi login/push/pull/search
├── agent.go            # muxi agent *
├── mcp.go              # muxi mcp *
├── chat.go             # muxi chat
├── session.go          # muxi session *
├── trigger.go          # muxi trigger *
├── job.go              # muxi job *
├── logs.go             # muxi logs
└── ... (other commands)
```

### Key Technologies

- **CLI Framework:** Cobra
- **HTTP Client:** Go stdlib
- **Encryption:** Go crypto (AES-256-GCM)
- **YAML:** gopkg.in/yaml.v3
- **Spinners:** github.com/briandowns/spinner
- **Progress:** github.com/schollz/progressbar/v3
- **Terminal:** golang.org/x/term (password input)

### Authentication

**1. Server API (HMAC):**
```go
func signRequest(req *http.Request, keyID, secret string) {
    timestamp := time.Now().Unix()
    bodyHash := sha256Hash(req.Body)
    message := fmt.Sprintf("%d;%s;%s;%s", timestamp, req.Method, req.URL.Path, bodyHash)
    
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(message))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    req.Header.Set("X-MUXI-Key-ID", keyID)
    req.Header.Set("X-MUXI-Timestamp", fmt.Sprintf("%d", timestamp))
    req.Header.Set("X-MUXI-Signature", signature)
}
```

**2. Formation API (API Keys):**
```go
func addFormationAuth(req *http.Request, adminKey string, isAdmin bool) {
    if isAdmin {
        req.Header.Set("X-MUXI-Admin-Key", adminKey)
    } else {
        req.Header.Set("X-MUXI-Client-Key", clientKey)
    }
}
```

**3. Registry API (Bearer Token):**
```go
func addRegistryAuth(req *http.Request, token string) {
    req.Header.Set("Authorization", "Bearer "+token)
}
```

---

## Context-Aware Behavior

**Git-style:** Formation commands only work in formation directory

```bash
# In formation directory
cd my-formation/
muxi validate           # ✓ Works
muxi deploy             # ✓ Works
muxi secrets set KEY    # ✓ Works

# Outside formation directory
cd ~/projects
muxi validate           # ✗ Error: Not in formation directory
muxi deploy             # ✗ Error: Not in formation directory

# Server operations work anywhere
muxi formation list     # ✓ Works (server operation)
muxi server status      # ✓ Works (server operation)
```

**Formation detection:**
- Walks up directory tree (max 5 levels)
- Stops at home directory or root
- Looks for `formation.yaml`

---

## Testing Strategy

### Unit Tests
- Config loading/saving
- Secrets encryption/decryption
- HMAC signing
- Profile resolution
- Secret reference validation

### Integration Tests
- Server API calls (mocked)
- Formation API calls (mocked)
- Registry API calls (mocked)
- End-to-end flows (deploy, pull, etc.)

### Manual Testing
- Real server deployment
- Multi-server deployment
- Registry publishing
- Interactive wizards

---

## Success Criteria

**Week 1 Complete When:**
- [ ] All core commands working (init, validate, deploy, secrets)
- [ ] Multi-server deployment working
- [ ] Registry integration working (push/pull)
- [ ] Server operations working (list, stop, restart)
- [ ] Formation API commands working (agents, mcps, chat)
- [ ] Tests passing (>70% coverage)
- [ ] Nice error messages and UX
- [ ] Ready for production use

**Production Ready:**
- ~80 commands implemented
- Full Formation API coverage
- Multi-server deployment
- Smart secrets management
- Interactive wizards
- Beautiful UX

---

**Status:** Design complete, ready to build in 1 week! 🚀
