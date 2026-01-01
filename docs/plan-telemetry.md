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

Stores counters locally, flushes daily.

```go
type LocalState struct {
    LastFlush   time.Time         `json:"last_flush"`
    Commands    map[string]int    `json:"commands"`
    Connections ConnectionStats   `json:"connections"`
}

type ConnectionStats struct {
    ProfilesConfigured    int `json:"profiles_configured"`
    FormationsConfigured  int `json:"formations_configured"`
    RegistriesConfigured  int `json:"registries_configured"`
    SuccessfulConnections int `json:"successful_connections"`
    FailedConnections     int `json:"failed_connections"`
}

func Load() *LocalState              // from ~/.muxi/cli/telemetry.json
func (s *LocalState) Save()          // to ~/.muxi/cli/telemetry.json
func (s *LocalState) Reset()         // clear counters, update last_flush
func (s *LocalState) IncrementCommand(cmd string)
func (s *LocalState) RecordConnection(success bool)
func (s *LocalState) FlushIfDue()    // check 1h, send if due, reset
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

### 5. Command Integration (`cmd/root.go`)

```go
// In PersistentPreRun (runs before every command)
func init() {
    rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
        if !telemetry.IsEnabled() {
            return
        }
        
        // Load local state
        state := telemetry.Load()
        
        // Check if flush is due (>1h since last)
        state.FlushIfDue()
        
        // Increment command counter
        state.IncrementCommand(cmd.Name())
        
        // Save state
        state.Save()
    }
}
```

**Flow:**
1. Load `~/.muxi/cli/telemetry.json`
2. If `last_flush` > 1h ago:
   - Send hourly_summary with accumulated counters
   - Clear file (reset all counters to 0, update `last_flush`)
3. Increment command counter
4. Save state

### 6. Connection Tracking (`pkg/server/client.go`)

```go
// After each API call
if err != nil {
    telemetry.RecordConnection(false)
} else {
    telemetry.RecordConnection(true)
}
```

---

## Data to Collect

Based on PRD, CLI telemetry uses **local aggregation with daily flush**.

### Daily Payload (from PRD)

```json
{
  "module": "cli",
  "event": "hourly_summary",
  "machine_id": "...",
  "ts": "2025-01-15T10:00:00Z",
  "country": "US",
  "schema_version": 1,
  "payload": {
    "system": {
      "muxi_version": "0.20251024.3",
      "os": "darwin",
      "arch": "arm64"
    },
    "commands": {
      "deploy": 15,
      "new": 8,
      "chat": 45,
      "agents": 12,
      "mcp": 5,
      "scheduler": 3,
      "logs": 20,
      "registry": 10,
      "sessions": 8,
      "memory": 4,
      "config": 6,
      "secrets": 3,
      "server": 12,
      "profiles": 2,
      "other": 15
    },
    "connections": {
      "profiles_configured": 2,
      "formations_configured": 5,
      "registries_configured": 1,
      "successful_connections": 120,
      "failed_connections": 3
    }
  }
}
```

### Command Categories

| Command | Category | Notes |
|---------|----------|-------|
| `new` | `new` | Scaffolding |
| `deploy` | `deploy` | Deployments |
| `validate` | `validate` | Pre-deploy checks |
| `chat` | `chat` | Interactive sessions |
| `agents`, `mcp`, `sops`, `triggers` | Each tracked | Formation components |
| `config`, `secrets` | Each tracked | Configuration |
| `server`, `profiles`, `formations` | Each tracked | Server management |
| `registry`, `login`, `push`, `pull` | `registry` | Registry operations |
| `logs`, `sessions`, `memory` | Each tracked | Monitoring |
| Everything else | `other` | Catch-all |

### Connection Tracking

| Field | When to Increment |
|-------|-------------------|
| `successful_connections` | Any successful API call to server |
| `failed_connections` | Network error, auth error, server error |

---

## Commands to Track

| Command | Subcommands | Notes |
|---------|-------------|-------|
| `new` | `formation`, `agent`, `mcp`, `sop`, `trigger` | Scaffolding usage |
| `deploy` | - | Most important metric |
| `validate` | - | Pre-deploy checks |
| `chat` | - | Interactive usage |
| `config` | `llm`, `memory`, `a2a`, etc. | Configuration patterns |
| `secrets` | `set`, `get`, `list`, `delete` | Secrets management |
| `server` | `list`, `get`, `start`, `stop`, `delete` | Server management |
| `registry` | `push`, `pull`, `search`, `login` | Registry usage |
| `profiles` | `add`, `list`, `remove`, `show` | Multi-server usage |

---

## File Structure

```
pkg/telemetry/
├── machine.go       # Machine ID generation (deterministic from OS)
├── machine_test.go
├── geo.go           # Country lookup (one-time fetch, cached in config)
├── geo_test.go
├── aggregator.go    # Local state storage, daily flush
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
- No network requests
- No geo lookups
- Machine ID still generated (for other features)

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
