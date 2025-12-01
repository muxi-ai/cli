# Plan: `muxi config security`

## Overview
Configure user credentials handling and security settings in formation.yaml.

## Command
```bash
muxi config security
```

## Banner
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Security                                  MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure how the formation handles user credentials when    │
│ MCP tools or services request authentication.                │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: Credential Mode

```
User Credentials

  Controls how the formation handles credential requests from tools.

  Select credential handling mode:
    ◉ redirect (always redirect to external credential management)
    ◯ dynamic (collect credentials inline when safe)
```

---

### Flow 1: Redirect Mode (Production)

```
Redirect Mode

  Users will be redirected to configure credentials on an external system,
  where you can collect credentials and store them securely using the SDKs.
  This is the recommended mode for production deployments.

  URL where users configure their credentials
  ✓ Redirect URL: https://example.com/credentials

  Custom message to show when credentials are needed (optional)

  Current (default):
    Runtime generates: "For security, please configure your credentials at <URL>."

  Redirect message [keep current]:
  ✓ Redirect message: using default
```

**Output (with default message):**
```yaml
user_credentials:
  mode: "redirect"
  redirect_url: "https://example.com/credentials"
```

**Output (with custom message):**
```yaml
user_credentials:
  mode: "redirect"
  redirect_url: "https://example.com/credentials"
  redirect_message: "Please configure your API keys in your account settings."
```

---

### Flow 2: Dynamic Mode (Development)

```
Dynamic Mode

  ⚠ Dynamic credential collection should only be used in development.
  User credentials will be collected inline when tools request them.

  Fallback URL when dynamic mode is not available
  ✓ Redirect URL: https://example.com/credentials

Encryption Settings

  User credentials collected in dynamic mode are encrypted at rest using Fernet.
  An encryption key will be auto-generated and stored in secrets.

  ✓ Generated USER_CREDENTIALS_ENCRYPTION_KEY
```

**Output:**
```yaml
user_credentials:
  mode: "dynamic"
  redirect_url: "https://example.com/credentials"
  encryption:
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
```

---

## Current Value Detection

For each setting, detect current value from formation.yaml and:
1. Show `[current]` next to the matching option
2. Use current value as default selection

Example:
```
  Select credential handling mode:
    ◯ redirect (always redirect to external credential management) [current]
    ◯ dynamic (collect credentials inline when safe)
```

---

## Secrets Created

| Secret | When Created |
|--------|--------------|
| `USER_CREDENTIALS_ENCRYPTION_KEY` | Dynamic mode with encryption |

## Validation

- Mode must be "redirect" or "dynamic"
- redirect_url: valid URL (https:// auto-added if missing)

## Implementation Notes

1. Use radio boxes (◯) for selections, not numbered lists
2. Add dimmed hints above each prompt explaining the setting
3. Indent content under section headers by 2 spaces
4. Show `[current]` on options that match existing configuration
5. Default selection to current value when editing
6. Use $EDITOR for multi-line redirect message
7. Show warning about dynamic mode security implications
8. Auto-generate Fernet encryption key (32 bytes, base64 encoded) for dynamic mode

## Security Considerations

- **Redirect mode** is recommended for production - credentials never touch the formation
- **Dynamic mode** should only be used in development/staging environments
- Encryption key should be auto-generated (32 bytes, base64 encoded)
- HTTPS should always be required for dynamic mode in non-local environments
