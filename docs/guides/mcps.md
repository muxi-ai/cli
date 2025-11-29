# MCPs Guide

This guide covers creating and configuring MCP (Model Context Protocol) servers in MUXI formations.

## What is an MCP?

MCP servers provide tools that agents can use. They expose capabilities like:
- Web search and browsing
- Database access
- File operations
- API integrations
- Custom business logic

## Creating an MCP

Must be run inside a formation directory.

### Interactive Mode (Recommended)

```bash
muxi new mcp
```

The wizard guides you through:
1. **MCP ID** - Unique identifier (e.g., `github-api`)
2. **Display Name** - Human-friendly name
3. **Description** - What this MCP does
4. **Transport Type** - HTTP or Stdio
5. **Connection Details** - URL or command
6. **Authentication** - API key, bearer, basic, or custom

### With Name

```bash
muxi new mcp github-api
```

### Agent-Specific MCP

```bash
muxi new mcp private-tool --agent research-assistant
```

Creates MCP only accessible to the specified agent.

### Non-Interactive

```bash
muxi new mcp github-api --no-wizard
```

## Transport Types

### HTTP Transport

For remote MCP servers accessible via HTTP:

```yaml
transport:
  type: http
  url: "https://api.example.com/mcp"
```

### Stdio Transport

For local MCP servers run as subprocesses:

```yaml
transport:
  type: stdio
  command: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-filesystem"
    - "/path/to/directory"
```

## File Structure

```
my-formation/
└── mcps/
    ├── web-search.yaml           # Formation-level MCP
    ├── database.yaml             # Formation-level MCP
    └── research-assistant/       # Agent-specific MCPs
        └── private-tool.yaml
```

## MCP Configuration

### HTTP MCP with API Key

```yaml
schema: "1.0.0"
id: "weather-api"
name: "Weather API"
description: "Get weather forecasts and conditions"
active: true

transport:
  type: http
  url: "https://api.weather.example.com/mcp"

auth:
  type: api_key
  header: "X-API-Key"
  key: ${{ secrets.WEATHER_API_KEY }}
```

### HTTP MCP with Bearer Token

```yaml
schema: "1.0.0"
id: "github-api"
name: "GitHub API"
description: "Access GitHub repositories and issues"
active: true

transport:
  type: http
  url: "https://api.github.com/mcp"

auth:
  type: bearer
  token: ${{ secrets.GITHUB_TOKEN }}
```

### HTTP MCP with Basic Auth

```yaml
schema: "1.0.0"
id: "internal-api"
name: "Internal API"
description: "Access internal company systems"
active: true

transport:
  type: http
  url: "https://internal.company.com/mcp"

auth:
  type: basic
  username: ${{ secrets.INTERNAL_USERNAME }}
  password: ${{ secrets.INTERNAL_PASSWORD }}
```

### Stdio MCP (Local)

```yaml
schema: "1.0.0"
id: "filesystem"
name: "Filesystem Access"
description: "Read and write local files"
active: true

transport:
  type: stdio
  command: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-filesystem"
    - "/allowed/directory"
  env:
    NODE_ENV: "production"
```

### Custom Headers Auth

```yaml
auth:
  type: custom
  headers:
    Authorization: "Custom ${{ secrets.CUSTOM_TOKEN }}"
    X-Client-ID: ${{ secrets.CLIENT_ID }}
    X-Tenant-ID: "tenant-123"
```

## Secrets Management

MCP auth credentials are stored encrypted:

```bash
# Set during wizard or manually
muxi secrets set WEATHER_API_KEY
muxi secrets set GITHUB_TOKEN

# List configured secrets
muxi secrets list
```

## Agent Access Control

### Formation-Level MCPs (Default)

All agents can use MCPs in `mcps/`:

```
mcps/
├── web-search.yaml    # All agents can use
└── database.yaml      # All agents can use
```

### Agent-Specific MCPs

Only specified agent can use:

```
mcps/
└── research-assistant/
    └── private-api.yaml   # Only research-assistant can use
```

### Restricting Agent Access

In agent config, limit which MCPs it can use:

```yaml
# agents/junior-assistant.yaml
mcps:
  - web-search        # Can only use these
  - documentation
```

## Common MCP Patterns

### Web Search

```yaml
id: "web-search"
transport:
  type: http
  url: "https://search-mcp.example.com"
auth:
  type: api_key
  header: "X-API-Key"
  key: ${{ secrets.SEARCH_API_KEY }}
```

### Database Access

```yaml
id: "database"
transport:
  type: stdio
  command: "mcp-postgres"
  args:
    - "--connection-string"
    - ${{ secrets.DATABASE_URL }}
```

### File System

```yaml
id: "files"
transport:
  type: stdio
  command: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-filesystem"
    - "/data"
```

## Best Practices

1. **Use secrets for credentials** - Never hardcode API keys
2. **Descriptive IDs** - `github-issues` not `api1`
3. **Limit scope** - Only expose necessary capabilities
4. **Agent-specific when needed** - Restrict sensitive tools
5. **Document tools** - Clear descriptions help agents use them correctly
6. **Test connections** - Verify MCP works before deploying
