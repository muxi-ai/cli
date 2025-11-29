# Registry Commands Implementation Plan

**Date:** 2025-11-29  
**Status:** Planning  
**Priority:** High (enables formation sharing)

---

## Overview

Implement CLI commands for interacting with the MUXI Registry:
- `muxi login` - Authenticate with registry
- `muxi logout` - Remove credentials
- `muxi push` - Publish formation to registry
- `muxi pull` - Download formation from registry
- `muxi search` - Search for formations
- `muxi show` - Display formation details

**Key insight:** CLI is simple - registry handles all GitHub operations. CLI just needs to authenticate, zip, and upload/download.

---

## Commands

### 1. `muxi login`

**Purpose:** Authenticate with the registry via GitHub OAuth

**Flow:**
1. Open browser to `https://registry.muxi.org/auth/cli/authorize`
2. User authenticates with GitHub (in browser)
3. Registry returns `mxr_` token
4. CLI saves token to `~/.muxi/credentials.json`

**Implementation Options:**

**Option A: Browser callback (preferred)**
1. Start local HTTP server on random port (e.g., 9876)
2. Open browser to `/auth/cli/authorize?callback=http://localhost:9876`
3. After OAuth, registry redirects to local server with token
4. CLI receives token, saves, stops server

**Option B: Device code flow**
1. CLI requests device code from registry
2. Display code to user: "Enter code XXXX at registry.muxi.org/device"
3. Poll registry for token completion
4. Save token when received

**Option C: Manual copy-paste**
1. Open browser to `/auth/cli/authorize`
2. User copies token from browser
3. CLI prompts: "Paste your token:"
4. Save token

**Recommendation:** Start with Option C (simplest), upgrade to A later.

**Output:**
```
$ muxi login

Opening browser to authenticate...
  https://registry.muxi.org/auth/cli/authorize

Paste your token: mxr_xxxxx

✓ Logged in as ranaroussi
  Token saved to ~/.muxi/credentials.json
```

**Credentials File:** `~/.muxi/credentials.json`
```json
{
  "registry": {
    "url": "https://registry.muxi.org",
    "token": "mxr_...",
    "username": "ranaroussi",
    "created_at": "2025-11-29T10:30:00Z"
  }
}
```

**Registry URLs:**
- Production: `https://registry.muxi.org`
- Development: `https://muxi.registry` (local)

---

### Registry Configuration

**Multiple registries supported** (per DESIGN.md):

**Resolution priority (highest to lowest):**
1. `--registry` flag (explicit)
2. `.muxi` file in formation directory
3. `default_registry` from `~/.muxi/config.yaml`
4. Default: `registry.muxi.org`

**Config file:** `~/.muxi/config.yaml`
```yaml
default_registry: registry.muxi.org
```

**Credentials file:** `~/.muxi/credentials.json`
```json
{
  "registries": {
    "registry.muxi.org": {
      "token": "mxr_...",
      "username": "ranaroussi",
      "created_at": "2025-11-29T10:30:00Z"
    },
    "private.company.com": {
      "token": "mxr_...",
      "username": "ranaroussi",
      "created_at": "2025-11-29T11:00:00Z"
    }
  }
}
```

**Formation-level override:** `.muxi` file
```yaml
registry: private.company.com
```

---

### 2. `muxi logout`

**Purpose:** Remove stored credentials for a registry

**Syntax:**
```bash
muxi logout                    # Logout from current/default registry
muxi logout private.company.com  # Logout from specific registry
```

**Flow:**
1. Remove token for specified registry from credentials.json
2. Confirm logout

**Output:**
```
$ muxi logout

✓ Logged out from registry.muxi.org
```

---

### 3. `muxi push`

**Purpose:** Publish formation to registry

**Prerequisites:** Must be logged in, must be in formation directory

**Flow:**
1. Validate formation.yaml exists and is valid
2. Check version is semver
3. Create ZIP bundle (excluding secrets, .git, etc.)
4. POST to `/api/formations/publish`
5. Display success with URLs

**Flags:**
- `--org <name>` - Publish to organization instead of personal account
- `--dry-run` - Show what would be published without uploading

**ZIP Contents:**
```
Include:
  - formation.yaml
  - README.md
  - agents/*.yaml
  - mcps/*.yaml
  - a2a/*.yaml
  - sops/*.md
  - triggers/*.md
  - knowledge/*.md

Exclude:
  - .git/
  - .muxi/
  - secrets.enc (WARN if present!)
  - .key
  - .env*
  - node_modules/
  - __pycache__/
  - *.pyc
  - .DS_Store
```

**Output:**
```
$ muxi push

Validating formation...
  ✓ formation.yaml valid
  ✓ Version: 1.2.0
  ✓ 3 agents, 2 MCPs, 1 SOP

Creating bundle...
  ✓ 12 files (28.5 KB)

Publishing to registry...
  ✓ Uploaded to registry
  ✓ GitHub repo updated
  ✓ Release v1.2.0 created

✓ Published @ranaroussi/customer-support v1.2.0

View at:
  Registry: https://registry.muxi.org/@ranaroussi/customer-support
  GitHub:   https://github.com/ranaroussi/muxi-customer-support

Share with:
  muxi pull @ranaroussi/customer-support
```

**Warnings:**
```
⚠ Warning: secrets.enc found - this will NOT be included in the bundle.
  Recipients will need to run: muxi secrets setup
```

---

### 4. `muxi pull`

**Purpose:** Download formation from registry

**Syntax:**
```bash
muxi pull @user/formation          # Latest version
muxi pull @user/formation:1.0.0    # Specific version
muxi pull @user/formation -o dir   # Custom output directory
```

**Flow:**
1. GET metadata from `/api/formations/@user/formation?pull=true`
2. Download ZIP from GitHub (direct, not through registry)
3. Extract to `./{formation-name}/` or specified directory
4. Display success and next steps

**Flags:**
- `-o, --output <dir>` - Output directory (default: `./{formation-name}`)
- `--force` - Overwrite existing directory

**Output:**
```
$ muxi pull @ranaroussi/customer-support

Fetching @ranaroussi/customer-support...
  ✓ Found v1.2.0 (28.5 KB)

Downloading...
  ✓ Downloaded from GitHub

Extracting to ./customer-support/...
  ✓ 12 files extracted

✓ Downloaded @ranaroussi/customer-support v1.2.0

Next steps:
  cd customer-support
  muxi secrets setup    # Configure required secrets
```

---

### 5. `muxi search`

**Purpose:** Search for formations in registry

**Syntax:**
```bash
muxi search "customer support"
muxi search "chatbot" --sort downloads
muxi search "ai" --limit 50
```

**Flow:**
1. GET `/api/search?q=query&sort=sort&limit=limit`
2. Display results in table format

**Flags:**
- `--sort <field>` - Sort by: relevance, downloads, stars, updated (default: relevance)
- `--limit <n>` - Max results (default: 20, max: 100)

**Output:**
```
$ muxi search "customer support"

Found 3 formations:

  @ranaroussi/customer-support        ⭐ 45   ↓ 1,234
    AI-powered customer support with intelligent escalation
    v1.2.0 • Updated 2 days ago

  @somedev/support-bot                ⭐ 12   ↓ 450
    Simple support automation bot
    v2.1.0 • Updated 1 week ago

  @muxi/helpdesk                      ⭐ 89   ↓ 2,100
    Official MUXI helpdesk formation
    v3.0.0 • Updated 3 days ago

Pull with: muxi pull @user/formation
```

---

### 6. `muxi show`

**Purpose:** Display detailed formation info

**Syntax:**
```bash
muxi show @user/formation
muxi show @user/formation:1.0.0
muxi show @user/formation --versions
```

**Flow:**
1. GET `/api/formations/@user/formation`
2. Display formatted info

**Flags:**
- `--versions` - Show all versions

**Output:**
```
$ muxi show @ranaroussi/customer-support

@ranaroussi/customer-support v1.2.0

  AI-powered customer support with intelligent escalation

  ⭐ 45 stars   ↓ 1,234 downloads   📦 28.5 KB

Components:
  • 3 agents
  • 2 MCPs
  • 1 SOP
  • 2 triggers

Links:
  Registry: https://registry.muxi.org/@ranaroussi/customer-support
  GitHub:   https://github.com/ranaroussi/muxi-customer-support

Published: Jan 15, 2025
Updated:   2 days ago

Pull with: muxi pull @ranaroussi/customer-support
```

---

## Implementation

### New Package: `pkg/registry`

```go
// pkg/registry/client.go
type Client struct {
    BaseURL string
    Token   string  // mxr_xxx token
}

func NewClient() *Client
func (c *Client) IsAuthenticated() bool
func (c *Client) GetFormation(ref string, trackDownload bool) (*Formation, error)
func (c *Client) GetVersions(ref string) ([]Version, error)
func (c *Client) Search(query string, sort string, limit int) ([]Formation, error)
func (c *Client) Publish(zipPath string, org string) (*PublishResult, error)

// pkg/registry/credentials.go
func LoadCredentials() (*Credentials, error)
func SaveCredentials(creds *Credentials) error
func DeleteCredentials() error

// pkg/registry/bundle.go
func CreateBundle(formationDir string) (string, error)  // Returns path to zip
func ExtractBundle(zipPath, destDir string) error
```

### New Commands

```go
// cmd/registry.go
var loginCmd = &cobra.Command{...}
var logoutCmd = &cobra.Command{...}
var pushCmd = &cobra.Command{...}
var pullCmd = &cobra.Command{...}
var searchCmd = &cobra.Command{...}
var showCmd = &cobra.Command{...}
```

### Files

| File | Purpose |
|------|---------|
| `pkg/registry/client.go` | HTTP client for registry API |
| `pkg/registry/credentials.go` | Load/save `~/.muxi/credentials.json` |
| `pkg/registry/bundle.go` | Create/extract ZIP bundles |
| `pkg/registry/types.go` | Data structures (Formation, Version, etc.) |
| `cmd/registry.go` | CLI commands |

---

## Implementation Order

1. **Credentials management** - Load/save/delete credentials
2. **`muxi login`** - Manual token paste (Option C)
3. **`muxi logout`** - Remove credentials
4. **Registry client** - HTTP client with auth
5. **`muxi show`** - Test API connection
6. **`muxi search`** - Search functionality
7. **`muxi pull`** - Download formations
8. **Bundle creation** - ZIP with exclusions
9. **`muxi push`** - Publish formations
10. **Browser login** - Upgrade to Option A (future)

---

## Error Handling

| Error | Message |
|-------|---------|
| Not logged in | "Not authenticated. Run: muxi login" |
| Token expired | "Token expired. Run: muxi login" |
| Formation not found | "Formation @user/name not found" |
| Version not found | "Version 1.0.0 not found for @user/name" |
| Rate limited | "Rate limit exceeded. Try again in X minutes." |
| Network error | "Failed to connect to registry. Check your internet connection." |
| Invalid formation | "Invalid formation.yaml: {details}" |
| Version exists | "Version 1.0.0 already exists. Bump version in formation.yaml" |

---

## Testing

1. **Unit tests** for bundle creation/extraction
2. **Unit tests** for credentials management
3. **Integration tests** with mock registry server
4. **Manual testing** with real registry

---

## Security Considerations

1. **Credentials file permissions** - Set to 600 (user only)
2. **Never include secrets.enc** - Warn user if present
3. **Never include .key** - Exclude from bundle
4. **Token in memory only** - Don't log tokens
5. **HTTPS only** - Enforce TLS for registry communication

---

## Future Enhancements

- Browser-based login (local callback server)
- `muxi whoami` - Show current user
- `muxi unpublish` - Remove formation from registry
- `muxi star @user/formation` - Star a formation
- Private formations (requires registry update)
- Organization management
