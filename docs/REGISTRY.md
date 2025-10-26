# MUXI CLI - Registry Commands

**Version:** Alpha (v0.1.0)  
**Status:** Specification  
**Last Updated:** 2025-01-15

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Authentication](#authentication)
4. [Publishing Commands](#publishing-commands)
5. [Discovery Commands](#discovery-commands)
6. [Configuration](#configuration)
7. [Error Handling](#error-handling)
8. [Implementation Notes](#implementation-notes)

---

## Overview

### Purpose

Registry commands enable developers to:
- **Publish** formations to the MUXI Registry (backed by GitHub)
- **Discover** formations published by others
- **Install** formations with a single command
- **Search** for formations by name/description

### Design Philosophy

**GitHub-Backed Model:**
- Registry uses GitHub repos as storage backend
- Formations published to `github.com/username/muxi-formation-name`
- Registry is a lightweight UX/discovery layer
- GitHub handles versioning, releases, CDN

**Key Principles:**
1. **Zero friction for users** - No auth needed to pull
2. **Minimal trust burden** - GitHub App with fine-grained permissions
3. **Git-native workflow** - Devs can use git directly on repos
4. **Lazy discovery** - Any `muxi-*` repo is automatically discoverable

---

## Architecture

### Command Flow

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   CLI User   │         │   Registry   │         │    GitHub    │
│              │         │     API      │         │              │
└──────┬───────┘         └──────┬───────┘         └──────┬───────┘
       │                        │                        │
       │ muxi login             │                        │
       │─────────────────────────>                       │
       │                        │                        │
       │                        │ GitHub App Install     │
       │                        │───────────────────────>│
       │                        │<───────────────────────│
       │                        │                        │
       │<─────────────────────────                       │
       │ Token saved            │                        │
       │                        │                        │
       │ muxi push              │                        │
       │─────────────────────────>                       │
       │                        │ Create repo + release  │
       │                        │───────────────────────>│
       │                        │                        │
       │<─────────────────────────                       │
       │ Published!             │                        │
       │                        │                        │
       │ muxi pull @user/form   │                        │
       │─────────────────────────>                       │
       │ Get metadata           │                        │
       │<─────────────────────────                       │
       │                        │                        │
       │ Download bundle.zip                             │
       │────────────────────────────────────────────────>│
       │<────────────────────────────────────────────────│
```

### Naming Conventions

**Formation References:**
```
@username/formation-name              # Latest version
@username/formation-name:1.0.0        # Specific version
@username/formation-name:1.x          # Version range (future)
```

**GitHub Repo Mapping:**
```
@ranaroussi/customer-support  →  github.com/ranaroussi/muxi-customer-support
```

**Repo Prefix:** All formation repos use `muxi-` prefix

---

## Authentication

### `muxi login`

Authenticate with MUXI Registry via GitHub App installation.

**Syntax:**
```bash
muxi login [options]
```

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)
- `--browser <command>` - Custom browser command (default: auto-detect)
- `--no-browser` - Print URL instead of opening browser

**Behavior:**

1. **Check existing authentication:**
   ```bash
   $ muxi login
   
   ✓ Already authenticated as @ranaroussi
     Token expires: 2025-12-31
   
   Continue to refresh? [y/N]
   ```

2. **GitHub App Installation:**
   ```bash
   $ muxi login
   
   Opening browser to install MUXI GitHub App...
   → https://github.com/apps/muxi-registry/installations/new
   
   Waiting for authentication...
   ```

3. **User sees in browser:**
   ```
   GitHub: "Install MUXI Registry"
   
   Permissions:
   ✓ Create repositories
   ✓ Read and write code in selected repositories
   
   Repository access:
   ○ All repositories
   ● Only select repositories (recommended)
   
   [Install]
   ```

4. **After install:**
   ```
   Browser shows:
   ┌─────────────────────────────────────────┐
   │  ✓ MUXI Registry Installed              │
   │                                         │
   │  You're authenticated as @ranaroussi    │
   │                                         │
   │  Return to your terminal to continue.   │
   └─────────────────────────────────────────┘
   
   Terminal shows:
   ✓ Authenticated as @ranaroussi
     Token saved to ~/.muxi/credentials.json
     
   You can now publish formations with: muxi push
   ```

**Credentials Storage:**

Creates/updates `~/.muxi/credentials.json`:
```json
{
  "registry": {
    "url": "https://registry.muxi.org",
    "token": "mxr_abc123def456...",
    "github_installation_id": 12345678,
    "github_username": "ranaroussi",
    "created_at": "2025-01-15T10:30:00Z",
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

**Permissions (600):** Only user-readable

**Exit Codes:**
- `0` - Success
- `1` - Authentication failed
- `2` - Browser failed to open
- `3` - Timeout waiting for callback

---

### `muxi logout`

Clear registry authentication.

**Syntax:**
```bash
muxi logout [options]
```

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)

**Behavior:**

```bash
$ muxi logout

Logged out from registry.muxi.org
Token removed from ~/.muxi/credentials.json

Note: MUXI GitHub App is still installed.
To fully revoke access, visit:
  https://github.com/settings/installations
```

**Exit Codes:**
- `0` - Success (even if not logged in)

---

## Publishing Commands

### `muxi push`

Publish formation to registry (creates/updates GitHub repo).

**Syntax:**
```bash
muxi push [path] [options]
```

**Arguments:**
- `path` - Formation directory (default: current directory)

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)
- `--tag <version>` - Override version from formation.yaml (future)
- `--dry-run` - Show what would be published without pushing
- `--force` - Force push even if version exists (requires confirmation)
- `-y, --yes` - Skip confirmation prompts

**Behavior:**

#### 1. **Validation Phase**

```bash
$ muxi push

Validating formation...
→ Checking formation.yaml exists... ✓
→ Reading formation.yaml... ✓
→ Formation: customer-support v1.0.0
→ Validating schema... ✓
→ Checking required files... ✓
  • formation.yaml ✓
  • README.md ✓
  • agents/ ✓ (3 files)
```

**Validation Checks:**
- `formation.yaml` exists and is valid
- `version` field is present and valid semver
- `id` field matches repo naming conventions
- Required directories exist (agents/, mcps/, etc.)
- No secrets in files (warns about `secrets.enc`)

#### 2. **Authentication Check**

```bash
→ Checking authentication... ✓
→ Authenticated as @ranaroussi
```

**If not authenticated:**
```bash
→ Checking authentication... ✗

You need to authenticate first:
  muxi login

Then try again: muxi push
```

#### 3. **Repository Check**

**Case A: Repo doesn't exist**

```bash
→ Checking github.com/ranaroussi/muxi-customer-support...
→ Repository doesn't exist

MUXI will create a new PUBLIC repository:

  📦 github.com/ranaroussi/muxi-customer-support

This repository will:
  • Be publicly visible on GitHub
  • Contain your formation source code
  • Be managed by MUXI GitHub App
  • Be used for distribution via registry.muxi.org

Continue? [Y/n]
```

**Case B: Repo exists, checking version**

```bash
→ Checking github.com/ranaroussi/muxi-customer-support... ✓
→ Repository exists

→ Checking if v1.0.0 already exists...
→ Tag v1.0.0 found

✗ Error: Version 1.0.0 already published

Options:
  • Increment version in formation.yaml
  • Use --force to overwrite (not recommended)
  
Current versions:
  v1.0.0 (published 2 weeks ago)
  v0.9.0 (published 1 month ago)
```

#### 4. **Bundling Phase**

```bash
→ Bundling formation files...

Including:
  • formation.yaml (2.1 KB)
  • README.md (5.3 KB)
  • agents/ (3 files, 12.4 KB)
  • mcps/ (2 files, 3.2 KB)
  • sops/ (1 file, 1.8 KB)
  • triggers/ (2 files, 4.1 KB)

Excluding:
  • .git/
  • .muxi/
  • secrets.enc (🔒 secrets should not be published)
  • node_modules/
  • __pycache__/
  • .DS_Store

→ Created bundle.zip (28.9 KB) ✓
→ SHA256: abc123def456...
```

**Bundle Contents:**
- All formation files
- Compressed as ZIP
- Excludes: `.git`, secrets, build artifacts

#### 5. **Publishing Phase**

**If creating new repo:**

```bash
→ Creating repository via GitHub API... ✓
→ Repository: github.com/ranaroussi/muxi-customer-support

→ Initializing local git repository... ✓
→ Adding files... ✓
→ Committing... ✓
  Commit: Initial formation: customer-support v1.0.0
  
→ Adding remote 'origin'... ✓
→ Pushing to GitHub... ✓
```

**If updating existing repo:**

```bash
→ Cloning existing repository... ✓
→ Updating files... ✓
→ Committing changes... ✓
  Commit: Update to v1.1.0
  
→ Pushing to GitHub... ✓
```

#### 6. **Release Phase**

```bash
→ Creating tag v1.0.0... ✓
→ Pushing tag... ✓

→ Creating GitHub release v1.0.0... ✓
→ Uploading bundle.zip (28.9 KB)... ✓

Release URL: github.com/ranaroussi/muxi-customer-support/releases/tag/v1.0.0
```

#### 7. **Registry Notification**

```bash
→ Notifying registry.muxi.org... ✓
→ Registry cached metadata ✓
```

**API Call:**
```
POST https://registry.muxi.org/api/formations/publish
Authorization: Bearer <token>

{
  "github_repo": "ranaroussi/muxi-customer-support",
  "version": "1.0.0",
  "formation_id": "customer-support"
}
```

#### 8. **Success Message**

```bash
✓ Published @ranaroussi/customer-support v1.0.0!

View at:
  Registry: https://registry.muxi.org/@ranaroussi/customer-support
  GitHub:   https://github.com/ranaroussi/muxi-customer-support

Share with:
  muxi pull @ranaroussi/customer-support

Install with:
  muxi pull @ranaroussi/customer-support
  cd customer-support/
  muxi deploy
```

**Exit Codes:**
- `0` - Success
- `1` - Validation failed
- `2` - Not authenticated
- `3` - Version already exists
- `4` - GitHub API error
- `5` - Network error

---

### `muxi push --dry-run`

Preview what would be published without actually pushing.

**Example:**

```bash
$ muxi push --dry-run

DRY RUN MODE - No changes will be made

Formation: customer-support v1.0.0

Would create repository:
  github.com/ranaroussi/muxi-customer-support

Would bundle files (28.9 KB):
  • formation.yaml (2.1 KB)
  • README.md (5.3 KB)
  • agents/ (3 files, 12.4 KB)
  • mcps/ (2 files, 3.2 KB)
  • sops/ (1 file, 1.8 KB)
  • triggers/ (2 files, 4.1 KB)

Would create:
  • Git repository
  • Tag v1.0.0
  • GitHub release
  • Registry entry

To publish for real:
  muxi push
```

---

## Discovery Commands

### `muxi pull`

Download and extract a formation from the registry.

**Syntax:**
```bash
muxi pull <formation-ref> [options]
```

**Arguments:**
- `<formation-ref>` - Formation reference (e.g., `@user/name` or `@user/name:version`)

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)
- `--output <path>` - Output directory (default: formation name)
- `--no-extract` - Download bundle.zip without extracting
- `--version <ver>` - Specific version (alternative to `@user/name:ver` syntax)

**Behavior:**

#### 1. **Parse Formation Reference**

```bash
$ muxi pull @ranaroussi/customer-support

Parsing formation reference...
→ User: ranaroussi
→ Formation: customer-support
→ Version: latest (not specified)
```

#### 2. **Query Registry**

```bash
→ Querying registry.muxi.org...
→ Formation found ✓

Formation: @ranaroussi/customer-support
Latest version: 1.2.0
Description: AI-powered customer support with escalation
GitHub: github.com/ranaroussi/muxi-customer-support
Downloads: 1,234
```

**API Call:**
```
GET https://registry.muxi.org/api/formations/@ranaroussi/customer-support
```

**If not in registry cache (lazy discovery):**
```bash
→ Not found in registry, checking GitHub...
→ Found github.com/ranaroussi/muxi-customer-support ✓
→ Registering formation... ✓
```

#### 3. **Download Bundle**

```bash
→ Downloading from GitHub...
→ URL: github.com/ranaroussi/muxi-customer-support/releases/download/v1.2.0/bundle.zip
→ Size: 28.9 KB

Downloading... ████████████████████ 100% (28.9 KB)
→ Download complete ✓
→ SHA256: abc123def456... ✓
```

**Direct GitHub download** (doesn't count against registry rate limits)

#### 4. **Extract**

```bash
→ Extracting to ./customer-support/...

Extracted:
  • formation.yaml
  • README.md
  • agents/ (3 files)
  • mcps/ (2 files)
  • sops/ (1 file)
  • triggers/ (2 files)

✓ Downloaded @ranaroussi/customer-support v1.2.0
```

#### 5. **Record Download**

```bash
→ Recording download... ✓
```

**API Call:**
```
POST https://registry.muxi.org/api/formations/@ranaroussi/customer-support/1.2.0/download
```

Increments download counter (for stats).

#### 6. **Success Message**

```bash
✓ Formation ready!

Location: ./customer-support/

Next steps:
  cd customer-support/
  
  # Review the formation
  cat README.md
  
  # Configure secrets (if needed)
  cp secrets.enc.example secrets.enc
  # Edit secrets.enc with your credentials
  
  # Deploy
  muxi deploy --profile local
```

**Exit Codes:**
- `0` - Success
- `1` - Formation not found
- `2` - Invalid formation reference
- `3` - Download failed
- `4` - Extraction failed
- `5` - Network error

---

### `muxi pull` - Specific Version

```bash
$ muxi pull @ranaroussi/customer-support:1.0.0

Parsing formation reference...
→ User: ranaroussi
→ Formation: customer-support
→ Version: 1.0.0 (explicitly specified)

→ Querying registry.muxi.org...
→ Version 1.0.0 found ✓

→ Downloading from GitHub...
# (same flow as above)

✓ Downloaded @ranaroussi/customer-support v1.0.0
```

---

### `muxi search`

Search for formations in the registry.

**Syntax:**
```bash
muxi search <query> [options]
```

**Arguments:**
- `<query>` - Search query (matches name, description, README)

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)
- `--limit <n>` - Max results (default: 20)
- `--sort <field>` - Sort by: `relevance`, `downloads`, `stars`, `updated` (default: `relevance`)
- `-o, --output <format>` - Output format: `text`, `json`, `yaml`

**Behavior:**

```bash
$ muxi search "customer support"

Searching registry.muxi.org for "customer support"...

Found 3 formations:

┌────────────────────────────────────────────────────────────────┐
│ @ranaroussi/customer-support                          v1.2.0   │
├────────────────────────────────────────────────────────────────┤
│ AI-powered customer support with escalation                    │
│ ⬇ 1.2K pulls  ⭐ 45 stars  📦 28.9 KB                         │
│                                                                │
│ muxi pull @ranaroussi/customer-support                         │
└────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────┐
│ @somedev/support-bot                                  v2.1.0   │
├────────────────────────────────────────────────────────────────┤
│ Simple support automation bot                                  │
│ ⬇ 450 pulls  ⭐ 12 stars  📦 15.2 KB                          │
│                                                                │
│ muxi pull @somedev/support-bot                                 │
└────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────┐
│ @anotherdev/zendesk-integration                       v1.0.0   │
├────────────────────────────────────────────────────────────────┤
│ Zendesk integration for customer support                       │
│ ⬇ 89 pulls  ⭐ 3 stars  📦 8.1 KB                             │
│                                                                │
│ muxi pull @anotherdev/zendesk-integration                      │
└────────────────────────────────────────────────────────────────┘

Showing 3 of 3 results
```

**API Call:**
```
GET https://registry.muxi.org/api/search?q=customer+support&limit=20&sort=relevance
```

**JSON Output:**

```bash
$ muxi search "customer support" --output json

[
  {
    "ref": "@ranaroussi/customer-support",
    "version": "1.2.0",
    "description": "AI-powered customer support with escalation",
    "downloads": 1234,
    "stars": 45,
    "size_bytes": 29593,
    "github_repo": "ranaroussi/muxi-customer-support",
    "registry_url": "https://registry.muxi.org/@ranaroussi/customer-support"
  },
  ...
]
```

**Exit Codes:**
- `0` - Success (even if no results)
- `1` - Invalid query
- `2` - Network error

---

### `muxi show`

Display detailed information about a formation.

**Syntax:**
```bash
muxi show <formation-ref> [options]
```

**Arguments:**
- `<formation-ref>` - Formation reference (e.g., `@user/name`)

**Options:**
- `--registry <url>` - Registry URL (default: `registry.muxi.org`)
- `--version <ver>` - Show specific version info
- `-o, --output <format>` - Output format: `text`, `json`, `yaml`

**Behavior:**

```bash
$ muxi show @ranaroussi/customer-support

Formation: @ranaroussi/customer-support
Version:   1.2.0 (latest)
Author:    @ranaroussi

Description:
  AI-powered customer support with intelligent escalation,
  sentiment analysis, and multi-channel integration.

Stats:
  Downloads:   1,234 total
  Stars:       45
  Size:        28.9 KB
  Published:   2025-01-10 (5 days ago)

Components:
  • 3 agents (escalation, sentiment, router)
  • 2 MCPs (zendesk, slack)
  • 1 SOP (escalation-procedure)
  • 2 triggers (new-ticket, urgent-keyword)

Links:
  Registry:  https://registry.muxi.org/@ranaroussi/customer-support
  GitHub:    https://github.com/ranaroussi/muxi-customer-support
  Issues:    https://github.com/ranaroussi/muxi-customer-support/issues

Versions:
  v1.2.0  (latest, 5 days ago)   - 234 downloads
  v1.1.0  (2 weeks ago)          - 456 downloads
  v1.0.0  (1 month ago)          - 544 downloads

Installation:
  muxi pull @ranaroussi/customer-support
```

**API Call:**
```
GET https://registry.muxi.org/api/formations/@ranaroussi/customer-support
GET https://registry.muxi.org/api/formations/@ranaroussi/customer-support/versions
```

**Exit Codes:**
- `0` - Success
- `1` - Formation not found
- `2` - Network error

---

## Configuration

### Credentials File

**Location:** `~/.muxi/credentials.json`

**Structure:**
```json
{
  "registry": {
    "url": "https://registry.muxi.org",
    "token": "mxr_abc123def456...",
    "github_installation_id": 12345678,
    "github_username": "ranaroussi",
    "created_at": "2025-01-15T10:30:00Z",
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

**Permissions:** `600` (user read/write only)

**Token Format:** `mxr_` prefix + random alphanumeric

---

### Environment Variables

**Override registry URL:**
```bash
export MUXI_REGISTRY_URL=https://registry.example.com
muxi pull @user/formation
```

**Override credentials path:**
```bash
export MUXI_CREDENTIALS_FILE=/path/to/credentials.json
```

---

## Error Handling

### Common Errors

#### Not Authenticated

```bash
$ muxi push

✗ Error: Not authenticated

You need to authenticate with the registry first:
  muxi login

Then try again: muxi push
```

**Exit Code:** `2`

---

#### Formation Not Found

```bash
$ muxi pull @nonexistent/formation

✗ Error: Formation not found

"@nonexistent/formation" doesn't exist in:
  • Registry: registry.muxi.org
  • GitHub: github.com/nonexistent/muxi-formation

Check for typos, or search:
  muxi search "formation"
```

**Exit Code:** `1`

---

#### Version Already Exists

```bash
$ muxi push

✗ Error: Version 1.0.0 already published

The tag v1.0.0 already exists in:
  github.com/ranaroussi/muxi-customer-support

Options:
  1. Increment version in formation.yaml
  2. Use --force to overwrite (not recommended)

Current versions:
  v1.0.0 (published 2 weeks ago)
  v0.9.0 (published 1 month ago)
```

**Exit Code:** `3`

---

#### Invalid formation.yaml

```bash
$ muxi push

✗ Error: Invalid formation.yaml

Missing required field: version

Your formation.yaml must include:
  formation:
    id: my-formation
    version: "1.0.0"    # ← Required for publishing
    ...

See: https://docs.muxi.org/formations/schema
```

**Exit Code:** `1`

---

#### Network Error

```bash
$ muxi pull @user/formation

Querying registry.muxi.org...
✗ Error: Network error

Could not connect to registry.muxi.org
  • Check your internet connection
  • Check if registry is down: https://status.muxi.org

Details: dial tcp: lookup registry.muxi.org: no such host
```

**Exit Code:** `5`

---

#### GitHub API Rate Limit

```bash
$ muxi pull @user/formation

✗ Error: GitHub API rate limit exceeded

You've exceeded GitHub's rate limit.

Options:
  1. Wait 47 minutes for reset
  2. Authenticate with GitHub (increases limit):
       muxi login

Rate limit info:
  Limit:      60 requests/hour (anonymous)
  Used:       60
  Resets:     2025-01-15 11:30:00 (in 47 minutes)
```

**Exit Code:** `4`

---

## Implementation Notes

### GitHub API Integration

**Endpoints Used:**

```go
// Create repository
POST /user/repos
{
  "name": "muxi-formation-name",
  "description": "Formation description from formation.yaml",
  "homepage": "https://registry.muxi.org/@user/formation-name",
  "private": false,
  "auto_init": false
}

// Create tag
POST /repos/:owner/:repo/git/refs
{
  "ref": "refs/tags/v1.0.0",
  "sha": "<commit-sha>"
}

// Create release
POST /repos/:owner/:repo/releases
{
  "tag_name": "v1.0.0",
  "name": "v1.0.0",
  "body": "Release notes from formation changelog",
  "draft": false,
  "prerelease": false
}

// Upload release asset
POST /repos/:owner/:repo/releases/:release_id/assets?name=bundle.zip
Content-Type: application/zip
[binary data]

// Get release info
GET /repos/:owner/:repo/releases/tags/v1.0.0

// Get repo info
GET /repos/:owner/:repo

// List releases
GET /repos/:owner/:repo/releases
```

### Registry API Integration

**Endpoints Used:**

```go
// Publish notification
POST /api/formations/publish
Authorization: Bearer <token>
{
  "github_repo": "user/muxi-formation",
  "version": "1.0.0",
  "formation_id": "formation-name"
}

// Get formation metadata
GET /api/formations/@:user/:name
Response: {
  "user": "ranaroussi",
  "name": "customer-support",
  "latest_version": "1.2.0",
  "description": "...",
  "github_repo": "ranaroussi/muxi-customer-support",
  "downloads": 1234,
  "stars": 45
}

// Get specific version
GET /api/formations/@:user/:name/:version

// Record download
POST /api/formations/@:user/:name/:version/download

// Search
GET /api/search?q=customer+support&limit=20
```

### Bundling Logic

**Files to Include:**
- `formation.yaml` (required)
- `README.md` (recommended)
- `agents/` (all files)
- `mcps/` (all files)
- `sops/` (all files)
- `triggers/` (all files)
- `knowledge/` (future: may need size limits)

**Files to Exclude:**
- `.git/`
- `.muxi/`
- `secrets.enc` (warn if found)
- `node_modules/`
- `__pycache__/`
- `.DS_Store`
- `*.pyc`
- `.env`

**Bundle Format:**
- ZIP compression
- Preserve directory structure
- Include SHA256 hash in metadata

### Version Resolution

**Latest:**
```
@user/formation          → Query registry for latest_version
                          → Download that version from GitHub
```

**Specific:**
```
@user/formation:1.0.0    → Download v1.0.0 from GitHub
```

**Semver Range (Future):**
```
@user/formation:^1.0.0   → Resolve to highest 1.x version
@user/formation:~1.2.0   → Resolve to highest 1.2.x version
```

---

**End of Registry CLI Specification**

See also:
- [ALPHA-PRD.md](../../registry/ALPHA-PRD.md) - Overall registry product requirements
- [CLI-COMMAND-DESIGN.md](./CLI-COMMAND-DESIGN.md) - Main CLI command design
- [IMPLEMENTATION-PLAN.md](./IMPLEMENTATION-PLAN.md) - CLI implementation plan
