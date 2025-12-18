# User Credentials Management

## Why These Endpoints Exist

MUXI formations can include MCP (Model Context Protocol) servers that require user-specific authentication - for example, a GitHub MCP that needs each user's personal access token, or a Gmail MCP that needs OAuth tokens.

In the formation YAML, these are configured with placeholders:

```yaml
mcp_servers:
  github-mcp:
    auth:
      token: "${{ user.credentials.github }}"
```

When a user interacts with the formation and triggers a tool that requires their credential, the runtime needs to:
1. Check if the user has stored a credential for that service
2. If not, prompt them to provide one (via the chat interface or redirect URL)
3. Store it encrypted in the database for future use

## The Problem

Previously, credential management was only possible through the chat interface - the user would be prompted mid-conversation to provide their API key. There was no way for developers to:
- Build custom UIs for credential management
- Pre-populate credentials during user onboarding
- Show users which services they've connected
- Allow users to revoke/delete credentials

## The Solution: `/credentials` API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /credentials/services` | List which MCP servers in this formation use user credentials (so devs know what to ask for) |
| `GET /credentials` | List user's stored credentials (metadata + redacted preview, never the actual secret) |
| `POST /credentials` | Store a new credential for a service |
| `GET /credentials/{id}` | Get metadata for a specific credential |
| `DELETE /credentials/{id}` | Revoke/delete a credential |

## Key Design Decisions

1. **Secrets are NEVER returned** - only a redacted preview like `ghp_abc*****xyz`
2. **Service validation is flexible** - we accept any service name (lowercased, no spaces) but provide `/credentials/services` so devs can check what's actually configured
3. **DualKeyAuth** - both ClientKey and AdminKey work, but `X-Muxi-User-ID` is required (except for `/credentials/services` which returns formation-level config)
4. **Credentials are encrypted at rest** using per-user key derivation

## Typical Dev Flow

1. Call `GET /credentials/services` to see what services need credentials
2. Build UI showing "Connect your GitHub", "Connect your Gmail", etc.
3. When user provides token, call `POST /credentials` with `{service: "github", credential: {token: "..."}}`
4. Show connected services via `GET /credentials`
5. Allow disconnection via `DELETE /credentials/{id}`

## CLI Commands

### List Available Services

```bash
muxi credentials services -f <formation>
```

Shows which MCP servers in the formation are configured to use user credentials:

```
  SERVICE         SERVER ID            DESCRIPTION
  ───────────────────────────────────────────────────────────────────────
  github          github-mcp           MCP server requiring user authentication
  gmail           gmail-mcp            MCP server requiring user authentication
```

### List User's Credentials

```bash
muxi credentials list -f <formation> -u <user>
```

Shows credentials stored for a user (metadata only, secrets are redacted):

```
  ID              SERVICE      NAME                 PREVIEW
  ─────────────────────────────────────────────────────────────────────
  cred_abc123     github       Work GitHub          ghp_abc*****xyz
  cred_def456     gmail        Personal Gmail       ya29.a0*****Qx8
```

### Show Credential Details

```bash
muxi credentials show <credential-id> -f <formation> -u <user>
```

### Add a Credential

```bash
muxi credentials add -f <formation> -u <user>
muxi credentials add -f <formation> -u <user> github  # pre-select service
```

Interactive wizard:

```
  Service (↑↓ to select, Enter to confirm):
  ◉ github - MCP server requiring user authentication
  ◯ gmail - MCP server requiring user authentication
  ◯ Other (enter custom service)

  Name (e.g., company account, myusername): Work GitHub
  Token: ****************************

✓ Credential added

  ID:       cred_abc123
  Service:  github
  Name:     Work GitHub
  Preview:  ghp_abc*****xyz
```

### Delete a Credential

```bash
muxi credentials delete <credential-id> -f <formation> -u <user>
muxi credentials rm <credential-id> -f <formation> -u <user>  # alias
```

## See Also

- [Security Configuration](security.md) - Configure credential handling modes
- [MCP Guide](mcps.md) - Configure MCP servers that use credentials
