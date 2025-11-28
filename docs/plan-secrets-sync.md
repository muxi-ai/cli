# Plan: `muxi secrets sync` Command

## Overview

A command to synchronize secrets between formation files and the encrypted secrets store.

## Command

```bash
muxi secrets sync
```

## Flow

### Step 1: Scan Formation Files
Scan all YAML files for `${{ secrets.XXX }}` patterns:
- `formation.yaml`
- `agents/*.yaml`
- `mcps/*.yaml`
- `a2a/*.yaml`
- `sops/*.md` (frontmatter)
- `triggers/*.md` (frontmatter)

Extract all referenced secret names.

### Step 2: Compare with `secrets` Template
The `secrets` file lists all keys that should exist:
```
OPENAI_API_KEY=
GITHUB_TOKEN=
DATABASE_URL=
```

### Step 3: Reconciliation

**Add missing keys to `secrets` template:**
- If a `${{ secrets.XXX }}` is found in YAML but XXX is not in `secrets` template
- Add `XXX=` to the `secrets` file

**Remove unused keys from `secrets` and `secrets.enc`:**
- If a key exists in `secrets`/`secrets.enc` but is NOT referenced anywhere in formation files
- Prompt user: "SECRET_NAME is not used in any formation files. Delete? (y/N)"
- If yes, remove from both `secrets` template and `secrets.enc`

**Prompt for missing values:**
- If a key exists in `secrets` template (KEY=) but has no value in `secrets.enc`
- Start interactive wizard to prompt for each missing value
- Store in `secrets.enc` as user provides them

## Example Session

```
$ muxi secrets sync

Scanning formation files...
  ✓ formation.yaml (3 secrets)
  ✓ agents/assistant.yaml (0 secrets)
  ✓ mcps/github.yaml (1 secret)

Found 4 unique secrets referenced:
  - OPENAI_API_KEY
  - ANTHROPIC_API_KEY
  - GITHUB_TOKEN
  - SLACK_WEBHOOK_URL

Checking secrets store...
  ✓ OPENAI_API_KEY: stored
  ✓ ANTHROPIC_API_KEY: stored
  ⊘ GITHUB_TOKEN: missing value
  ⊘ SLACK_WEBHOOK_URL: missing value

Unused secrets in store:
  - OLD_API_KEY (not referenced in any files)

Delete unused secret OLD_API_KEY? (y/N): y
✓ Deleted OLD_API_KEY

Enter missing secrets:

GITHUB_TOKEN: ghp_xxxxxxxxxxxx
✓ GITHUB_TOKEN: ***

SLACK_WEBHOOK_URL: https://hooks.slack.com/xxx
✓ SLACK_WEBHOOK_URL: ***

✓ Secrets synchronized!
  - 4 secrets active
  - 1 secret deleted
  - 2 secrets added
```

## Flags

```bash
muxi secrets sync --dry-run    # Show what would change without making changes
muxi secrets sync --no-delete  # Don't prompt to delete unused secrets
muxi secrets sync --no-prompt  # Don't prompt for missing values (just report)
```

## Implementation Notes

1. Use `regexp.MustCompile(`\$\{\{\s*secrets\.([A-Z0-9_]+)\s*\}\}`)` to find patterns
2. Parse `secrets` template with `secrets.ParseSecretsFile()`
3. Use `secrets.Manager` for encrypted store operations
4. Normalize all names with `secrets.NormalizeName()`

## Related Commands

- `muxi secrets list` - List all secrets in store
- `muxi secrets set KEY` - Set a single secret
- `muxi secrets get KEY` - Get a secret value (masked by default)
- `muxi secrets delete KEY` - Delete a secret
- `muxi secrets sync` - Synchronize secrets with formation files

## Priority

Medium - useful for maintenance but not blocking initial release.

---

**Created:** 2025-11-28
**Status:** Planned
