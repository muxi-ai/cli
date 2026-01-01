# CLI Telemetry Implementation Plan

**Status:** Planning  
**Est. Time:** 4-6 hours  
**Priority:** High

---

## Overview

Implement anonymous telemetry to understand CLI usage patterns. All telemetry is opt-out via `muxi telemetry disable` or `MUXI_TELEMETRY=0`.

## Already Completed

- [x] `muxi telemetry enable/disable/status` command
- [x] Config storage in `~/.muxi/config.yaml`
- [x] Hidden from `--help`

---

## Implementation Tasks

### 1. Machine ID (`pkg/telemetry/machine.go`)

Deterministic ID from OS hardware, cached in config.

```go
func GetMachineID() string
func getOSMachineID() string  // platform-specific
func generateMachineID(osID string) string  // sha256 + format
```

**Platform sources:**
- macOS: `ioreg -rd1 -c IOPlatformExpertDevice` → IOPlatformUUID
- Linux: `/etc/machine-id` or `/var/lib/dbus/machine-id`
- Windows: `wmic csproduct get uuid`

**Algorithm:** `sha256(os_id + "muxi")` → format as UUID (8-4-4-4-12)

### 2. Country Lookup (`pkg/telemetry/geo.go`)

Country code cached permanently in config (one-time fetch).

```go
func GetCountry() string  // returns cached or fetches once
```

**Flow:**
1. Check `~/.muxi/config.yaml` for `country` field
2. If present → use cached value (never refetch)
3. If missing → fetch from `https://ipapi.co/json/` (2s timeout)
4. Save to config.yaml

**Config example:**
```yaml
machine_id: "7f83b165-7ff1-fc53-b92d-c18148a1d65d"
telemetry: true
country: "US"
```

No separate geo.json file needed - everything in one config.

### 3. Local Aggregator (`pkg/telemetry/aggregator.go`)

Stores counters locally, flushes hourly.

```go
type LocalState struct {
    LastFlush    time.Time        `json:"last_flush"`
    Registry     RegistryStats    `json:"registry"`
    Formations   FormationStats   `json:"formations"`
    Scaffolding  ScaffoldStats    `json:"scaffolding"`
    Usage        UsageStats       `json:"usage"`
    Help         map[string]int   `json:"help"`
}

type RegistryStats struct {
    Pulls  int `json:"pulls"`
    Pushes int `json:"pushes"`
}

type FormationStats struct {
    Created  int `json:"created"`
    Deployed int `json:"deployed"`
}

type ScaffoldStats struct {
    Agents   int `json:"agents"`
    MCPs     int `json:"mcps"`
    SOPs     int `json:"sops"`
    Triggers int `json:"triggers"`
}

type UsageStats struct {
    ChatSessions int `json:"chat_sessions"`
    LogsViewed   int `json:"logs_viewed"`
}

func Load() *LocalState              // from ~/.muxi/cli/telemetry.json
func (s *LocalState) Save()          // to ~/.muxi/cli/telemetry.json
func (s *LocalState) Reset()         // clear counters, update last_flush
func (s *LocalState) FlushIfDue()    // check 1h, send if due, reset

// Increment methods
func (s *LocalState) IncrementPull()
func (s *LocalState) IncrementPush()
func (s *LocalState) IncrementFormationCreated()
func (s *LocalState) IncrementDeploy()
func (s *LocalState) IncrementScaffold(kind string)  // "agent", "mcp", "sop", "trigger"
func (s *LocalState) IncrementChat()
func (s *LocalState) IncrementLogs()
func (s *LocalState) IncrementHelp(cmd string)
```

**Storage:** `~/.muxi/cli/telemetry.json`

### 4. Telemetry Client (`pkg/telemetry/client.go`)

```go
func IsEnabled() bool    // check MUXI_TELEMETRY env + config
func Send(event Event)   // POST, fire-and-forget, single retry
```

**Endpoint:** `POST https://capture.muxi.org/v1/telemetry`  
**Timeout:** 2 seconds  
**Retries:** 1 (non-blocking)

### 5. Command Integration

Telemetry hooks in specific commands (not generic PersistentPreRun):

```go
// cmd/registry.go - pull command
func runPull(cmd *cobra.Command, args []string) error {
    state := telemetry.Load()
    defer state.Save()
    
    // ... pull logic ...
    
    state.IncrementPull()
    state.FlushIfDue()  // sends if enabled and 1h passed
    return nil
}

// cmd/deploy.go
func runDeploy(cmd *cobra.Command, args []string) error {
    state := telemetry.Load()
    defer state.Save()
    
    // ... deploy logic ...
    
    state.IncrementDeploy()
    state.FlushIfDue()
    return nil
}

// cmd/new.go - formation subcommand
func runNewFormation(cmd *cobra.Command, args []string) error {
    state := telemetry.Load()
    defer state.Save()
    
    // ... scaffold logic ...
    
    state.IncrementFormationCreated()
    state.FlushIfDue()
    return nil
}

// Any command with --help flag
func init() {
    rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
        if cmd.Flags().Changed("help") {
            state := telemetry.Load()
            state.IncrementHelp(cmd.Name())
            state.FlushIfDue()
            state.Save()
        }
    }
}
```

**Flow:**
1. Load `~/.muxi/cli/telemetry.json` (always)
2. Increment relevant counter (always)
3. Check if flush due (>1h) → send if telemetry enabled, then reset
4. Save state (always)

**Note:** Data is always collected locally. Only the send is conditional on telemetry being enabled.

---

## Data to Collect

Based on PRD, CLI telemetry uses **local aggregation with hourly flush**.

### Payload Structure

Designed to answer key product questions:
1. **Registry adoption** - Are people sharing/consuming formations?
2. **Creation → Deployment funnel** - Do formations get used?
3. **Feature adoption** - What do people add to formations?
4. **Infrastructure scale** - How do people set up?
5. **Engagement mode** - Interactive vs automated?
6. **UX friction** - Where do people need help?

```json
{
  "module": "cli",
  "machine_id": "...",
  "ts": "2025-01-15T10:00:00Z",
  "country": "US",
  "schema_version": 1,
  "payload": {
    "system": {
      "version": "0.20251024.3",
      "os": "darwin",
      "arch": "arm64"
    },
    "registry": {
      "pulls": 7,
      "pushes": 2
    },
    "formations": {
      "created": 3,
      "deployed": 12
    },
    "scaffolding": {
      "agents": 5,
      "mcps": 2,
      "sops": 1,
      "triggers": 0
    },
    "usage": {
      "chat_sessions": 45,
      "logs_viewed": 8
    },
    "infrastructure": {
      "profiles_configured": 2,
      "formations_configured": 5,
      "registries_configured": 1
    },
    "help": {
      "total": 15,
      "deploy": 5,
      "registry": 3,
      "secrets": 2,
      "other": 5
    }
  }
}
```

### What Each Section Tracks

| Section | Fields | Source |
|---------|--------|--------|
| `registry` | `pulls`, `pushes` | `muxi pull`, `muxi push` commands |
| `formations` | `created`, `deployed` | `muxi new formation`, `muxi deploy` |
| `scaffolding` | `agents`, `mcps`, `sops`, `triggers` | `muxi new agent/mcp/sop/trigger` |
| `usage` | `chat_sessions`, `logs_viewed` | `muxi chat`, `muxi logs` |
| `infrastructure` | counts | Read from `~/.muxi/cli/*.yaml` at flush |
| `help` | per-command counts | Explicit `--help` flag only (not default subcommand listing) |

### Help Tracking

Only track explicit `--help` / `-h` flag usage:

```go
// Only track when --help flag is explicitly used
if cmd.Flags().Changed("help") {
    telemetry.IncrementHelp(cmd.Name())
}
```

This captures "I need help" intent, not just exploring subcommands.

---

## Commands to Track

| Metric | Command(s) | Increment Method |
|--------|------------|------------------|
| `registry.pulls` | `muxi pull`, `muxi registry pull` | `IncrementPull()` |
| `registry.pushes` | `muxi push`, `muxi registry push` | `IncrementPush()` |
| `formations.created` | `muxi new formation` | `IncrementFormationCreated()` |
| `formations.deployed` | `muxi deploy` | `IncrementDeploy()` |
| `scaffolding.agents` | `muxi new agent` | `IncrementScaffold("agent")` |
| `scaffolding.mcps` | `muxi new mcp` | `IncrementScaffold("mcp")` |
| `scaffolding.sops` | `muxi new sop` | `IncrementScaffold("sop")` |
| `scaffolding.triggers` | `muxi new trigger` | `IncrementScaffold("trigger")` |
| `usage.chat_sessions` | `muxi chat` | `IncrementChat()` |
| `usage.logs_viewed` | `muxi logs` | `IncrementLogs()` |
| `help.*` | Any `--help` flag | `IncrementHelp(cmdName)` |

**Infrastructure counts** are read from config files at flush time (not incremented).

---

## File Structure

```
pkg/telemetry/
├── machine.go       # Machine ID generation (deterministic from OS)
├── machine_test.go
├── geo.go           # Country lookup (one-time fetch, cached in config)
├── geo_test.go
├── aggregator.go    # Local state storage, hourly flush
├── aggregator_test.go
├── client.go        # HTTP client for sending events
├── client_test.go
└── types.go         # Event types

~/.muxi/
├── config.yaml          # machine_id, telemetry, country
└── cli/
    └── telemetry.json   # Local command counters, last_flush
```

---

## Opt-Out Behavior

Check in order:
1. `MUXI_TELEMETRY=0` env var → disabled
2. `~/.muxi/config.yaml` → `telemetry: false` → disabled
3. Otherwise → enabled (default)

When disabled:
- **Still collects locally** (to `~/.muxi/cli/telemetry.json`)
- No network requests (data not sent)
- No geo lookups
- Machine ID still generated (for other features)

If user later enables telemetry, accumulated data will be sent on next flush.

---

## Testing Strategy

1. **Unit tests** - Mock HTTP, mock OS commands
2. **Integration** - Verify end-to-end with test endpoint
3. **Manual** - `TELEMETRY_URL=http://localhost:8080` override

---

## Privacy Considerations

**We DO collect:**
- Hashed machine ID (not raw hardware IDs)
- Country code (not IP address stored)
- Command usage patterns
- Error categories (not error messages)

**We DO NOT collect:**
- Formation names or IDs
- Secret values
- API keys
- File contents
- User input
- IP addresses (used for geo lookup only)

---

## Rollout Plan

1. Implement behind feature flag initially
2. Add to next minor release
3. Document in README/docs
4. Announce in changelog

---

## Open Questions

1. Should we track `chat` session duration separately?
2. Track registry search queries? (privacy concern)
3. Include formation size (file count, LOC)?
4. Track `--help` usage to improve docs?
