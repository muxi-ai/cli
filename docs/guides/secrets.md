# Secrets Management Guide

This guide covers how to manage secrets in MUXI formations using the `muxi secrets` commands.

## Overview

MUXI supports two modes for secrets management:

| Mode | Command | Description |
|------|---------|-------------|
| **Local** | `muxi secrets list` | Manage secrets in local `secrets.enc` file |
| **Remote** | `muxi secrets list --remote` | View secrets on a running Formation |

### Local Secrets (Default)

MUXI uses encrypted secrets storage to protect sensitive values like API keys, tokens, and credentials. Secrets are:

- **Encrypted at rest** using Fernet encryption (compatible with Python runtime)
- **Stored in `secrets.enc`** in your formation directory
- **Referenced in YAML files** using `${{ secrets.SECRET_NAME }}` syntax

## Files

| File | Purpose |
|------|---------|
| `secrets.enc` | Encrypted secrets storage (do not edit manually) |
| `secrets` | Template file listing all secret keys (KEY=) |
| `.key` | Encryption key (auto-generated, keep secure) |

> **Important:** Add `.key` and `secrets.enc` to `.gitignore`. The `secrets` template can be committed to share required keys with your team.

## Commands

### List Secrets

```bash
# List all secret keys
muxi secrets list

# List with values (use with caution)
muxi secrets list --with-values
```

### Set a Secret

```bash
# Prompt for value (recommended - keeps value out of shell history)
muxi secrets set OPENAI_API_KEY

# Set directly
muxi secrets set OPENAI_API_KEY sk-...
```

Secret names are automatically normalized:
- `my-api-key` → `MY_API_KEY`
- `openai_key` → `OPENAI_KEY`

### Delete a Secret

```bash
muxi secrets delete OPENAI_API_KEY
```

### Setup Secrets (After Pulling from Registry)

When you pull a formation from a registry, it includes a `secrets` template but no `secrets.enc`. Use `setup` to populate values:

```bash
# Interactive setup - prompts for each missing secret
muxi secrets setup

# Preview what would be prompted
muxi secrets setup --dry-run
```

### Sync Secrets (During Development)

When you modify formation files and add/remove secret references, use `sync` to keep everything in sync:

```bash
# Scan files, update template, delete unused, prompt for new values
muxi secrets sync

# Interactive mode - confirm before deleting unused secrets
muxi secrets sync -i

# Preview changes without applying
muxi secrets sync --dry-run

# Sync without prompting for values
muxi secrets sync --no-setup
```

**What sync does:**
1. Scans all formation files for `${{ secrets.* }}` patterns
2. Adds new secrets to the `secrets` template
3. Deletes unused secrets from both template and `secrets.enc`
4. Runs `setup` to prompt for any empty values

**Files scanned:**
- `formation.afs` / `formation.yaml`
- `agents/*.afs` / `agents/*.yaml`
- `mcps/*.afs` / `mcps/*.yaml`
- `a2a/*.afs` / `a2a/*.yaml`
- `sops/*.md`
- `triggers/*.md`

## Referencing Secrets

Use the `${{ secrets.NAME }}` syntax in your YAML files:

```yaml
# formation.afs
llm:
  provider: openai
  api_key: ${{ secrets.OPENAI_API_KEY }}

# mcps/github.afs
auth:
  type: bearer
  token: ${{ secrets.GITHUB_TOKEN }}

# a2a/partner-service.afs
auth:
  type: api_key
  header: X-API-Key
  key: ${{ secrets.PARTNER_API_KEY }}
```

Spacing is flexible - both `${{ secrets.NAME }}` and `${{secrets.NAME}}` work.

## Typical Workflows

### Starting a New Formation

```bash
# Create formation (wizard prompts for API key)
muxi new formation my-app

# Add more secrets as needed
muxi secrets set GITHUB_TOKEN
muxi secrets set DATABASE_URL
```

### Pulling a Formation from Registry

```bash
# Pull formation
muxi pull my-team/my-formation

# Setup secrets (prompts for all required values)
cd my-formation
muxi secrets setup
```

### After Modifying Configuration

```bash
# Added new MCP with auth? Sync to update secrets
muxi secrets sync

# Preview what would change
muxi secrets sync --dry-run
```

### Cleaning Up Unused Secrets

```bash
# See what's unused
muxi secrets sync --dry-run

# Delete with confirmation
muxi secrets sync -i

# Auto-delete without prompting
muxi secrets sync
```

## Security Best Practices

1. **Never commit secrets.enc or .key** - Add to `.gitignore`
2. **Commit the secrets template** - Helps team know what's needed
3. **Use `muxi secrets set` without value** - Avoids shell history exposure
4. **Rotate keys regularly** - Use `muxi secrets set` to update
5. **Use specific secret names** - `GITHUB_TOKEN` not `TOKEN`

## Troubleshooting

### "No secrets in template file"

The `secrets` file is empty or doesn't exist. Either:
- Run `muxi secrets sync` to scan formation files
- Manually add keys to the `secrets` file

### "Failed to decrypt secrets"

The `.key` file may be corrupted or missing. If you have no valuable secrets stored:
1. Delete `.key` and `secrets.enc`
2. Run any secrets command to regenerate

### Secret not being interpolated at runtime

Check that:
1. The secret exists: `muxi secrets list`
2. The name matches exactly (case-sensitive)
3. The syntax is correct: `${{ secrets.NAME }}`

---

## Remote Secrets (Formation API)

When a formation is deployed and running, you can view its configured secrets using the `--remote` flag. This queries the Formation API to show what secrets are loaded in the runtime environment.

### List Remote Secrets

```bash
# View secrets on running formation (values are masked)
muxi secrets list --remote

# Specify formation ID explicitly
muxi secrets list --remote -F my-formation

# Use a specific server profile
muxi secrets list --remote -p production
```

**Example output:**

```
Secrets (3):
  ANTHROPIC_API_KEY    sk-ant-••••••••••
  GITHUB_TOKEN         ghp_••••••••••••
  OPENAI_API_KEY       sk-••••••••••Gst
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--remote` | | Fetch secrets from Formation API |
| `--formation` | `-F` | Formation ID (default: from `formation.afs`) |
| `--profile` | `-p` | Server profile (default: from `.muxi` or global) |

### Authentication

Remote secrets require admin API key access. The key is resolved in order:
1. `MUXI_ADMIN_KEY` environment variable
2. `admin_key` in local `secrets.enc`
3. Server profile configuration

### Use Cases

- **Verify deployment:** Confirm secrets were properly loaded after deploy
- **Debug issues:** Check if a secret exists in the running formation
- **Audit:** Review what secrets are configured in production

> **Note:** Remote secrets are always masked for security. To see full values, use `muxi secrets list --with-values` on the local `secrets.enc` file.
