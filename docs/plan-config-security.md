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
│ ℹ Configure how the formation handles user credentials when  │
│ MCP tools or services request authentication.                │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: Credential Mode

```
User Credentials

  Controls how the formation handles credential requests from tools.

  Select credential handling mode:
    ◯ redirect (always redirect to external credential management)
    ◯ dynamic (collect credentials inline when safe)
```

---

### Flow 1: Redirect Mode (Production)

```
Redirect Mode

  Users will be redirected to configure credentials externally.
  This is the recommended mode for production deployments.

  Custom message to show when credentials are needed (opens $EDITOR)
  ✓ Redirect message: configured

  Example message:
    "For security, please configure your credentials at https://portal.company.com/credentials"
```

**Output:**
```yaml
user_credentials:
  mode: "redirect"
  redirect_message: |
    For security, credentials must be configured outside of this chat interface.
    Please use your organization's credential management system.
```

---

### Flow 2: Dynamic Mode (Development)

```
Dynamic Mode

  ⚠ Dynamic credential collection should only be used in development.
  Credentials will be collected inline when tools request them.

  Environments where dynamic mode is allowed (comma-separated)
  ✓ Allowed environments: development, staging

  Require HTTPS for credential collection?
  Require HTTPS? [Y/n]: y
  ✓ Require HTTPS: enabled

  How long to cache collected credentials (minutes)
  ✓ Credential TTL: 60 minutes

  Maximum failed auth attempts before lockout
  ✓ Max attempts: 3
```

```
Encryption Settings

  Credentials collected in dynamic mode are encrypted at rest.

  Encryption type
    ◯ fernet (symmetric, recommended)
    ◯ aes-256-gcm
  ✓ Encryption: fernet

  Encryption key (will be stored in secrets)
  ✓ Saved USER_CREDENTIALS_ENCRYPTION_KEY to secrets

  Custom salt for key derivation (optional, adds isolation)
  ✓ Salt: (using default)
```

**Output:**
```yaml
user_credentials:
  mode: "dynamic"
  allowed_environments: ["development", "staging"]
  require_https: true
  credential_ttl_minutes: 60
  max_attempts: 3
  encryption:
    type: "fernet"
    key: "${{ secrets.USER_CREDENTIALS_ENCRYPTION_KEY }}"
    salt: "muxi-user-credentials-salt-v1"
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
- allowed_environments: array of valid environment names
- credential_ttl_minutes: positive integer
- max_attempts: 1-10
- Encryption type: fernet or aes-256-gcm

## Implementation Notes

1. Use radio boxes (◯) for selections, not numbered lists
2. Add dimmed hints above each prompt explaining the setting
3. Indent content under section headers by 2 spaces
4. Show `[current]` on options that match existing configuration
5. Default selection to current value when editing
6. Use $EDITOR for multi-line redirect message
7. Show warning about dynamic mode security implications
8. Generate secure encryption key automatically if not provided

## Security Considerations

- **Redirect mode** is recommended for production - credentials never touch the formation
- **Dynamic mode** should only be used in development/staging environments
- Encryption key should be auto-generated (32 bytes, base64 encoded)
- HTTPS should always be required for dynamic mode in non-local environments
