# Logs Guide

Stream real-time logs from formations.

## Basic Usage

Stream logs with at least one filter:

```bash
muxi logs -u alice                    # Filter by user
muxi logs -s sess_abc123              # Filter by session
muxi logs --request req_xyz           # Filter by request
muxi logs --agent weather-bot         # Filter by agent
muxi logs --level error               # Filter by log level
muxi logs --event "chat.*"            # Filter by event type
```

Press `Ctrl+C` to stop streaming.

## Filters

### User Filter

Stream all logs for a specific user:

```bash
muxi logs -u alice
muxi logs --user usr_abc123
```

### Session Filter

Stream logs for a specific session:

```bash
muxi logs -s sess_abc123
muxi logs --session sess_abc123
```

### Request Filter

Stream logs for a specific request:

```bash
muxi logs --request req_xyz789
```

### Agent Filter

Stream logs from a specific agent:

```bash
muxi logs --agent weather-bot
muxi logs --agent muxi-generalist
```

### Level Filter

Filter by log level:

```bash
muxi logs --level debug      # All levels
muxi logs --level info       # Info and above
muxi logs --level warn       # Warnings and above
muxi logs --level error      # Errors and critical
muxi logs --level critical   # Critical only
```

### Event Type Filter

Filter by event type (supports wildcards):

```bash
muxi logs --event "chat.*"           # All chat events
muxi logs --event "agent.selected"   # Specific event
muxi logs --event "error.*"          # All error events
```

## Combining Filters

Combine multiple filters for precise log streaming:

```bash
# Errors for a specific user
muxi logs -u alice --level error

# Agent logs for a session
muxi logs -s sess_abc123 --agent weather-bot

# Chat events for a user
muxi logs -u alice --event "chat.*"
```

## Options

```
-f, --formation    Formation ID
-p, --profile      Server profile
-u, --user         Filter by user ID
-s, --session      Filter by session ID
    --request      Filter by request ID
    --agent        Filter by agent ID
    --level        Filter by log level
    --event        Filter by event type (wildcards supported)
```

## Log Format

Logs are displayed with timestamps:

```
[24-Dec-2024 10:30:45 UTC] {"event": "chat.message", "user_id": "alice", ...}
```

## Troubleshooting

### No logs appearing

- Verify the formation is running: `muxi get`
- Check your filter is correct
- Ensure there's activity matching your filter

### Connection drops

- Check network connectivity
- Verify admin API key is valid
- Try reconnecting with `muxi logs`
