# Security Configuration Guide

Configure how your formation handles user credentials using `muxi config security`.

## Quick Start

```bash
cd my-formation
muxi config security
```

## What You Can Configure

### Credential Handling Modes

| Mode | Use Case | Description |
|------|----------|-------------|
| **Redirect** | Production | Users redirected to external credential management |
| **Dynamic** | Development | Credentials collected inline with encryption |

---

## Mode 1: Redirect (Recommended for Production)

In redirect mode, when MCP tools or services request credentials, users are shown a message directing them to configure credentials externally.

```yaml
user_credentials:
  mode: "redirect"
  redirect_message: |
    For security, credentials must be configured outside of this chat interface.
    Please use your organization's credential management system.
```

**Benefits:**
- Credentials never touch the formation
- Integrates with enterprise credential vaults
- Audit-friendly

**Configuration options:**
- **Redirect message** - Custom message shown to users (supports multi-line)

---

## Mode 2: Dynamic (Development Only)

In dynamic mode, credentials are collected inline when tools request them. Credentials are encrypted at rest using Fernet encryption.

```yaml
user_credentials:
  mode: "dynamic"
  allowed_environments: ["development", "staging"]
  require_https: true
  credential_ttl_minutes: 60
  max_attempts: 3
  encryption:
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
```

**Configuration options:**

| Setting | Default | Description |
|---------|---------|-------------|
| `allowed_environments` | development, staging | Environments where dynamic mode works |
| `require_https` | true | Require HTTPS for credential collection |
| `credential_ttl_minutes` | 60 | How long to cache credentials |
| `max_attempts` | 3 | Failed attempts before lockout (1-10) |

**Security features:**
- Fernet encryption (auto-generated key stored in secrets)
- Environment restrictions
- HTTPS requirement
- Attempt limiting

---

## Secrets Created

| Secret | Mode | Description |
|--------|------|-------------|
| `USER_CREDENTIALS_ENCRYPTION_KEY` | Dynamic | Auto-generated Fernet encryption key |

---

## Examples

### Production Setup (Redirect)

```bash
muxi config security
# Select: Redirect
# Enter custom message or use default
```

Result:
```yaml
user_credentials:
  mode: "redirect"
  redirect_message: |
    Please configure your API credentials at https://admin.example.com/credentials
    Contact IT support if you need assistance.
```

### Development Setup (Dynamic)

```bash
muxi config security
# Select: Dynamic
# Configure environments, HTTPS, TTL, max attempts
```

Result:
```yaml
user_credentials:
  mode: "dynamic"
  allowed_environments: ["development", "staging"]
  require_https: true
  credential_ttl_minutes: 60
  max_attempts: 3
  encryption:
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
```

---

## Security Best Practices

1. **Always use redirect mode in production** - Keep credentials out of the formation
2. **Require HTTPS** - Never collect credentials over unencrypted connections
3. **Limit environments** - Only allow dynamic mode in development/staging
4. **Short TTL** - Use shorter credential TTL for sensitive operations
5. **Monitor attempts** - Low max_attempts helps detect brute force attacks

---

## Related

- [Secrets Guide](secrets.md) - Managing encrypted secrets
- [MCP Guide](mcps.md) - Configuring MCP tools that may request credentials
