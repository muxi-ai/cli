# Plan: Logging Configuration Restructure

**Date:** 2025-12-18  
**Status:** Planned  
**Related PRD:** `/runtime/docs/prd/logging-restructure.md`

## Overview

Update CLI to support the new logging configuration structure that separates **system events** (infrastructure/debugging) from **conversation events** (user-facing/streamable).

## New Configuration Structure

```yaml
# Old structure
logging:
  enabled: true
  streams:
    - transport: stdout
      level: debug
      format: jsonl

# New structure
logging:
  system:
    level: debug              # default: debug
    destination: stdout       # default: stdout, or file path
  conversation:
    enabled: true
    streams:
      - transport: stdout
        level: debug
        format: jsonl
```

## API Response Changes

### `GET /logging`

**Old:**
```json
{
  "enabled": true,
  "streams": [...]
}
```

**New:**
```json
{
  "system": {
    "level": "debug",
    "destination": "stdout"
  },
  "conversation": {
    "enabled": true,
    "streams": [...]
  }
}
```

### `GET /logging/destinations`

**Old:**
```json
{
  "destinations": [...],
  "count": 2
}
```

**New:**
```json
{
  "system": {
    "level": "debug",
    "destination": "stdout"
  },
  "conversation": {
    "destinations": [...],
    "count": 2
  }
}
```

## Files to Change

### 1. `pkg/formation/types.go`

Update response types:

```go
// LoggingConfigResponse from GET /logging
type LoggingConfigResponse struct {
    System       LoggingSystemConfig       `json:"system"`
    Conversation LoggingConversationConfig `json:"conversation"`
}

type LoggingSystemConfig struct {
    Level       string `json:"level"`
    Destination string `json:"destination"`
}

type LoggingConversationConfig struct {
    Enabled bool            `json:"enabled"`
    Streams []LoggingStream `json:"streams"`
}

// LoggingDestinationsResponse from GET /logging/destinations
type LoggingDestinationsResponse struct {
    System       LoggingSystemConfig              `json:"system"`
    Conversation LoggingConversationDestinations  `json:"conversation"`
}

type LoggingConversationDestinations struct {
    Destinations []LoggingDestination `json:"destinations"`
    Count        int                  `json:"count"`
}
```

### 2. `cmd/logging.go`

#### `muxi logging status`

**New output:**
```
$ muxi logging status

 ╭──────────────────────────────────╮
 │ ⌬ my-formation │ ⚙︎ production │
 ╰──────────────────────────────────╯

  Logging Configuration

  System Events
    Level:       debug
    Destination: stdout

  Conversation Events
    Status:  ✓ enabled
    Streams: 2
      • stdout (1)
      • file   (1)

  List conversation destinations with:
    muxi logging list -f my-formation -p production
```

#### `muxi logging list`

**New output:**
```
$ muxi logging list

 ╭──────────────────────────────────╮
 │ ⌬ my-formation │ ⚙︎ production │
 ╰──────────────────────────────────╯

  System Events
    Level:       debug
    Destination: stdout

  Conversation Destinations (2)
  ─────────────────────────────────────────────────────────────────────────────────────
  ID              TRANSPORT   DESTINATION               LEVEL    FORMAT   STATUS
  ─────────────────────────────────────────────────────────────────────────────────────
  stdout-1        stdout      -                         debug    jsonl    ● on
  file-1          file        /var/log/conversation.log debug    jsonl    ● on
```

### 3. `pkg/scaffold/logging.go`

#### Update `ConfigureLogging()`

Add initial prompt to choose system vs conversation:

```go
func ConfigureLogging() error {
    // ... banner ...

    // Step 1: Choose what to configure
    options := []wizard.SelectOption{
        {Value: "system", Label: "System events", Description: "startup, errors, infrastructure"},
        {Value: "conversation", Label: "Conversation events", Description: "user-facing, streamable"},
    }

    choice, err := wizard.PromptSelect("What would you like to configure?", options, 0)
    if err != nil {
        return err
    }

    switch choice {
    case "system":
        return configureSystemLogging(ctx.RootDir)
    case "conversation":
        return configureConversationLogging(ctx.RootDir)
    }
    return nil
}
```

#### Add `configureSystemLogging()`

New function for system logging configuration:

```go
func configureSystemLogging(rootDir string) error {
    fmt.Println()
    ui.Bold("System Logging")
    fmt.Println()
    ui.Dimmed("  System events include: startup, connections, errors, infrastructure")
    fmt.Println()

    // Level
    levelOptions := []wizard.SelectOption{
        {Value: "debug", Label: "debug", Description: "all events"},
        {Value: "info", Label: "info", Description: "informational and above"},
        {Value: "warn", Label: "warn", Description: "warnings and errors only"},
        {Value: "error", Label: "error", Description: "errors only"},
    }
    level, err := wizard.PromptSelect("  Level", levelOptions, 0)
    if err != nil {
        return err
    }
    ui.PromptSuccess("  Level", level)

    // Destination
    destOptions := []wizard.SelectOption{
        {Value: "stdout", Label: "stdout", Description: "console output"},
        {Value: "file", Label: "file", Description: "write to file"},
    }
    destType, err := wizard.PromptSelect("  Destination", destOptions, 0)
    if err != nil {
        return err
    }

    destination := "stdout"
    if destType == "file" {
        destination, err = wizard.PromptString("  File path", "/var/log/system.log", nil)
        if err != nil {
            return err
        }
    }
    ui.PromptSuccess("  Destination", destination)

    // Update formation.yaml
    if err := setSystemLogging(rootDir, level, destination); err != nil {
        return err
    }

    fmt.Println()
    ui.Success("System logging configured")
    return nil
}
```

#### Rename current flow to `configureConversationLogging()`

```go
func configureConversationLogging(rootDir string) error {
    fmt.Println()
    ui.Bold("Conversation Logging")
    fmt.Println()
    ui.Dimmed("  Conversation events include: requests, agent activity, responses")
    fmt.Println()

    // Current wizard flow: add/view/remove streams
    options := []wizard.SelectOption{
        {Value: "add", Label: "Add a new log stream"},
        {Value: "view", Label: "View/edit current streams"},
        {Value: "remove", Label: "Remove a stream"},
    }
    // ... rest of current implementation
}
```

#### Add helper functions

```go
func setSystemLogging(rootDir, level, destination string) error {
    // Read formation.yaml
    // Find or create logging.system section
    // Set level and destination
    // Write back
}

func getSystemLogging(rootDir string) (level, destination string, err error) {
    // Read current system logging config
}
```

## Implementation Order

1. **Update types** (`pkg/formation/types.go`)
2. **Update `muxi logging status`** (`cmd/logging.go`)
3. **Update `muxi logging list`** (`cmd/logging.go`)
4. **Add system logging wizard** (`pkg/scaffold/logging.go`)
5. **Test with formation** that has new logging structure
6. **Commit and push**

## Testing

### Manual Tests

1. `muxi logging status` shows both system and conversation config
2. `muxi logging list` shows system config + conversation destinations
3. `muxi config logging` → System → configure level/destination
4. `muxi config logging` → Conversation → add/view/remove streams
5. Verify formation.yaml is updated correctly with new structure

### Edge Cases

- Formation with old logging structure (should still work or show migration hint)
- Formation with only system logging configured
- Formation with only conversation logging configured
- Empty logging config (should show defaults)

## Notes

- No backward compatibility required per PRD
- System logging defaults: `level: debug`, `destination: stdout`
- Conversation logging defaults: `enabled: false` (no streams)
