# Formation Info Guide

View runtime status and configuration for a deployed formation.

## Commands

### View Formation Status

```bash
muxi info
```

Shows formation status including:
- Status (running/stopped)
- Version and uptime
- Agent count and active agents
- MCP server connections
- Memory usage
- Request stats and CPU usage

**Example output:**

```
  Formation: my-formation

    Status:     ● running
    Version:    1.2.0
    Uptime:     5d 12h 30m

    Agents:     3 (2 active)
    MCP:        2 servers connected
    Memory:     256 MB working, 1.2 GB usage

    Stats:
      Requests:   1,234 total (12 active)
      CPU:        15%
```

### View Full Configuration

```bash
muxi info --full
```

Adds detailed configuration including:
- Schema version
- Total agents and secrets count
- MCP server configuration (timeout, retries)

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.yaml) |
| `--profile` | `-p` | Server profile (default: from .muxi or global) |
| `--full` | | Include detailed configuration |

## Authentication

Requires **admin API key**. The key is resolved from:
1. `MUXI_ADMIN_KEY` environment variable
2. `FORMATION_ADMIN_API_KEY` in `secrets.enc`

## Usage Patterns

### Check Formation Health

```bash
# Quick status check
muxi info

# Detailed config check
muxi info --full
```

### Targeting Different Formations

```bash
# Use formation in current directory
muxi info

# Specify a different formation
muxi info -F other-formation

# Use a specific server profile
muxi info -p production
```

### Automation/Scripting

The info command is useful for:
- Health checks in CI/CD pipelines
- Monitoring formation status
- Debugging deployment issues

## Related Commands

- `muxi agents` - List agents in detail
- `muxi mcp` - List MCP servers in detail
- `muxi config --remote` - Fetch full configuration as YAML
- `muxi sessions` - List user sessions
