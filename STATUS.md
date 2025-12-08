# MUXI CLI - Current Status

**Last Updated:** 2025-12-08
**Version:** 0.4.0-dev
**Status:** ✅ Ready for Formation API Implementation

---

## 🎯 Current State

### Overview
MUXI CLI is **fully functional** with complete scaffolding system, secrets management, all config commands, full registry integration, and **complete server/formation lifecycle commands** with SSE streaming progress.

### Recent Changes (2025-12-08)
- ✅ **Formation API Foundation COMPLETE:** `pkg/formation/` package ready
  - `client.go` - HTTP client with admin/client key auth
  - `auth.go` - API key resolution from secrets.enc or env vars
  - `types.go` - Response types for all Formation API endpoints
  - `flags.go` - Common flag helpers (-F, -p, -u)
  - Tested against live formation
- ✅ **Unified defaults command:** `muxi set default server|registry|user`
- ✅ **Command groups in help:** Organized into Formation/Registry/Server/Config groups
- ✅ **Default user_id support:** For upcoming Formation API commands
- ✅ **Formation API plan:** Complete parallelization strategy in `docs/plan-formation-api.md`

### What Exists ✅

**Formation Lifecycle Commands (All Complete!):**
- ✅ `muxi deploy` - Deploy formation with SSE streaming progress
- ✅ `muxi formation list` - List all deployed formations
- ✅ `muxi formation get <id>` - Get formation details (-v for verbose)
- ✅ `muxi formation stop <id>` - Stop a running formation
- ✅ `muxi formation start <id>` - Start a stopped formation (SSE streaming)
- ✅ `muxi formation restart <id>` - Restart formation (SSE streaming)
- ✅ `muxi formation rollback <id>` - Rollback to previous version (SSE streaming)
- ✅ `muxi formation delete <id>` - Delete formation from server
- ✅ `muxi formation logs <id>` - View formation logs

**Shortcut Commands (from formation directory):**
- ✅ `muxi get` - Get current formation details
- ✅ `muxi stop` - Stop current formation
- ✅ `muxi start` - Start current formation
- ✅ `muxi restart` - Restart current formation
- ✅ `muxi rollback` - Rollback current formation
- ✅ `muxi delete` - Delete current formation
- ✅ `muxi logs` - View current formation logs

**SSE Streaming Features:**
- ✅ Real-time progress updates from server
- ✅ Stage-by-stage progress (extracting, validating, spawning, health_check, etc.)
- ✅ Download progress percentage for runtime images
- ✅ Health check countdown display
- ✅ Ctrl+C cancellation with cleanup
- ✅ Notification sounds on completion (success/failure)
- ✅ Version validation before update (must be higher than server)

**Server Management:**
- ✅ `muxi server add` - Add server with HMAC credentials
- ✅ `muxi server list` - List servers (shows online/offline)
- ✅ `muxi server default` - Set default server
- ✅ `muxi server remove` - Remove server
- ✅ `muxi server status` - Show server status
- ✅ `muxi server ping` - Continuous ping with latency stats

**Scaffolding Commands:**
- ✅ `muxi new formation` - Full wizard with 21 LLM providers
- ✅ `muxi new agent` - Full wizard with role selection, A2A visibility
- ✅ `muxi new mcp` - Full wizard (HTTP/Stdio, formation/agent-level)
- ✅ `muxi new sop` - Wizard with title, description, mode
- ✅ `muxi new trigger` - Wizard with webhook template
- ✅ `muxi new a2a-service` - Full wizard with auth options

**Configuration Commands:**
- ✅ `muxi config a2a` - Inbound/outbound wizard
- ✅ `muxi config llm` - Full LLM configuration wizard
- ✅ `muxi config memory` - Full memory configuration wizard
- ✅ `muxi config overlord` - Full overlord configuration wizard
- ✅ `muxi config security` - User credential handling wizard
- ✅ `muxi config logging` - Logging streams wizard
- ✅ `muxi config async` - Async response settings wizard
- ✅ `muxi validate` - Validate formation configuration
- ✅ `muxi edit <type>` - Open files in $EDITOR

**Secrets Management:**
- ✅ `muxi secrets list [--with-values]` - List all secrets
- ✅ `muxi secrets set <name> [value]` - Set/update secret
- ✅ `muxi secrets delete <name>` - Delete secret
- ✅ `muxi secrets setup` - Populate secrets.enc from template
- ✅ `muxi secrets sync [-i] [--dry-run]` - Sync with formation files
- ✅ Fernet encryption (Python runtime compatible)

**Registry Commands:**
- ✅ `muxi login` - Authenticate with registry
- ✅ `muxi logout` - Remove registry credentials
- ✅ `muxi push` - Publish formation to registry
- ✅ `muxi pull @user/formation` - Download formation
- ✅ `muxi search "query"` - Search formations
- ✅ `muxi show @user/formation` - Display formation info
- ✅ `muxi registry mine` - List your published formations

**Default Settings (Unified Command):**
- ✅ `muxi set default server [name]` - Set default server (--local/--global)
- ✅ `muxi set default registry [name]` - Set default registry (--local/--global)
- ✅ `muxi set default user [id]` - Set default user ID (--local/--global)

**CLI Help Organization:**
- ✅ Formation Commands (available inside a formation directory)
- ✅ Registry Commands
- ✅ Server Commands
- ✅ Configuration

---

## 📋 Command Reference

### Deploy Command

```bash
muxi deploy [flags]

Flags:
  --profile string   Server profile to use
  --dry-run         Validate and create bundle without deploying
  --no-stream       Disable streaming progress (simpler output)
```

**Features:**
- Auto-detects new deploy vs update
- SSE streaming with stage progress
- Version validation (update requires higher version)
- Ctrl+C cancellation with server cleanup
- Notification sound on completion

### Formation Commands

```bash
# List all formations
muxi formation list [--profile string]

# Get formation details
muxi formation get <id> [flags]
  -v, --verbose      Show internal details (port, pid)
  --profile string   Server profile to use

# Stop a formation
muxi formation stop <id> [flags]
  -f, --force        Skip confirmation prompt
  --profile string   Server profile to use

# Start a stopped formation
muxi formation start <id> [--profile string]

# Restart a formation
muxi formation restart <id> [flags]
  -f, --force        Skip confirmation prompt
  --profile string   Server profile to use

# Rollback to previous version
muxi formation rollback <id> [flags]
  -f, --force        Skip confirmation prompt
  --profile string   Server profile to use

# Delete a formation
muxi formation delete <id> [flags]
  -f, --force        Skip confirmation prompt
  --profile string   Server profile to use

# View logs
muxi formation logs <id> [flags]
  -n, --lines int    Number of lines to show (default 100)
  -f, --follow       Stream new logs (like tail -f)
  --stream string    Filter by stream (stdout, stderr)
  --profile string   Server profile to use
```

### Shortcut Commands

Run from inside a formation directory:

```bash
muxi get [-v]              # Get current formation details
muxi stop [-f]             # Stop current formation
muxi start                 # Start current formation
muxi restart [-f]          # Restart current formation
muxi rollback [-f]         # Rollback current formation
muxi delete [-f]           # Delete current formation
muxi logs [-n 100] [-f]    # View logs (-f to follow)
```

### Server Commands

```bash
muxi server add <name> --url <url> --key-id <id> --secret-key <key>
muxi server list
muxi server default <name>
muxi server remove <name>
muxi server status [--profile string]
muxi server ping [--profile string]
```

---

## 🔄 SSE Streaming Progress Stages

### Deploy (POST)
1. `extracting` - Extracting bundle
2. `validating` - Validating formation.yaml
3. `resolving_runtime` - Resolving runtime version
4. `downloading_sif` - Downloading SIF file (with %)
5. `pulling_runner` - Pulling runtime-runner Docker image
6. `spawning` - Starting formation process
7. `health_check` - Waiting for health check (countdown)

### Update (PUT)
1. `extracting` - Extracting bundle to staging
2. `validating` - Validating formation.yaml
3. `resolving_runtime` - Resolving runtime version
4. `downloading_sif` - Downloading SIF file
5. `pulling_runner` - Pulling runtime-runner
6. `spawning_staging` - Starting staging version
7. `health_check` - Waiting for staging health check
8. `swapping` - Atomic switch (staging → current)
9. `stopping_old` - Stopping old version

### Start
1. `validating` - Loading formation configuration
2. `resolving_runtime` - Resolving runtime version
3. `downloading_sif` - Downloading SIF file
4. `pulling_runner` - Pulling runtime-runner
5. `spawning` - Starting process
6. `health_check` - Waiting for health check

### Restart
1. `stopping` - Stopping current process
2. `resolving_runtime` - Resolving runtime version
3. `downloading_sif` - Downloading SIF file
4. `pulling_runner` - Pulling runtime-runner
5. `spawning` - Starting new process
6. `health_check` - Waiting for health check

### Rollback
1. `validating` - Validating rollback request
2. `stopping` - Stopping current formation
3. `swapping` - Swapping to previous version
4. `resolving_runtime` - Resolving runtime version
5. `downloading_sif` - Downloading SIF file
6. `pulling_runner` - Pulling runtime-runner
7. `spawning` - Starting formation
8. `health_check` - Waiting for health check

---

## ✅ Testing Status

| Command | Status |
|---------|--------|
| Deploy (POST) | ✅ Tested |
| Update (PUT) | ✅ Tested |
| Ctrl+C on deploy | ✅ Tested |
| Ctrl+C on update | ✅ Tested |
| `formation list` | ✅ Tested |
| `formation get` | ✅ Tested |
| `formation delete` | ✅ Tested |
| `formation stop` | ✅ Tested |
| `formation start` | ✅ Tested |
| `formation restart` | ✅ Tested |
| `formation rollback` | ✅ Tested |
| `formation logs` | ✅ Tested |

---

## 📝 Recent Changes

### 2025-12-06
- ✅ Added `formation start` command with SSE streaming
- ✅ Added `muxi start` shortcut command
- ✅ Added SSE streaming to `formation rollback`
- ✅ Added version validation before update (must be higher)
- ✅ All formation lifecycle commands complete and tested

### 2025-12-05
- ✅ Added `formation restart` with SSE streaming
- ✅ Added notification sounds (Glass.aiff/Sosumi.aiff on macOS)
- ✅ Added Ctrl+C cancellation with server cleanup
- ✅ Improved error messages with contextual hints

### 2025-12-04
- ✅ Added SSE streaming to deploy/update
- ✅ Added progress stages with spinners
- ✅ Added health check countdown display

### 2025-12-03
- ✅ Server commands complete (add, list, default, remove, status, ping)
- ✅ Formation list/get/delete commands
- ✅ HMAC authentication implemented

---

## 🎯 What's Next

### Formation API Commands (Priority)
See **[docs/plan-formation-api.md](docs/plan-formation-api.md)** for full implementation plan.

**Foundation: ✅ COMPLETE**
- [x] `pkg/formation/client.go` - Formation API HTTP client
- [x] `pkg/formation/auth.go` - API key resolution
- [x] `pkg/formation/types.go` - Response types
- [x] `pkg/formation/flags.go` - Common flags

**Parallel Tracks: ✅ ALL COMPLETE**
- [x] Track A: `muxi info`, `muxi triggers`, `muxi sops`
- [x] Track B: `muxi agents`, `muxi mcp`
- [x] Track C: `muxi secrets --remote`, `muxi config --remote`
- [x] Track D: `muxi sessions`, `muxi history`, `muxi clear`
- [x] Track E: `muxi trigger`, `muxi jobs`, `muxi audit`, `muxi stream`
- [x] Track F: `muxi scheduler`, `muxi users`, `muxi memory`
- [ ] Phase 6: `muxi chat` (deferred - needs careful UX design)

### Future Enhancements
- [ ] `muxi formation scale <id> --replicas N` - Horizontal scaling
- [ ] `muxi formation exec <id> -- command` - Execute in formation
- [ ] Tab completion for bash/zsh/fish

---

**Last Updated:** 2025-12-08
**Maintained by:** MUXI Team
