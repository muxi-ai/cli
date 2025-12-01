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
  redirect_url: "https://example.com/credentials"
```

**Benefits:**
- Credentials never touch the formation
- Integrates with enterprise credential vaults
- Audit-friendly

**Configuration options:**
- **Redirect URL** - URL where users configure their credentials
- **Redirect message** (optional) - Custom message shown to users (runtime generates default)

---

## Mode 2: Dynamic (Development Only)

In dynamic mode, credentials are collected inline when tools request them. Credentials are encrypted at rest using Fernet encryption.

```yaml
user_credentials:
  mode: "dynamic"
  redirect_url: "https://example.com/credentials"
  encryption:
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
```

**Configuration options:**

| Setting | Description |
|---------|-------------|
| `redirect_url` | Fallback URL when dynamic mode is not available |

**Security features:**
- Fernet encryption (user-provided or auto-generated key stored in secrets)

---

## Secrets Created

| Secret | Mode | Description |
|--------|------|-------------|
| `USER_CREDENTIALS_ENCRYPTION_KEY` | Dynamic | Fernet encryption key (user-provided or auto-generated) |

---

## Examples

### Production Setup (Redirect)

```bash
muxi config security
# Select: Redirect
# Enter redirect URL
# Keep default message or enter custom
```

Result:
```yaml
user_credentials:
  mode: "redirect"
  redirect_url: "https://admin.example.com/credentials"
```

### Development Setup (Dynamic)

```bash
muxi config security
# Select: Dynamic
# Enter fallback redirect URL
# Enter encryption key or press Enter to auto-generate
```

Result:
```yaml
user_credentials:
  mode: "dynamic"
  redirect_url: "https://admin.example.com/credentials"
  encryption:
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
```

---

## Security Best Practices

1. **Always use redirect mode in production** - Keep credentials out of the formation
2. **Use HTTPS redirect URLs** - Ensure credential management happens over secure connections
3. **Dynamic mode for development only** - Never use dynamic mode in production

---

## Related

- [Secrets Guide](secrets.md) - Managing encrypted secrets
- [MCP Guide](mcps.md) - Configuring MCP tools that may request credentials
