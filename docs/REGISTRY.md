# MUXI CLI - Registry API Reference

**Version:** Alpha (v0.1.0)  
**Registry Status:** ✅ Production Ready  
**Last Updated:** 2025-10-29

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Authentication](#authentication)
4. [API Endpoints](#api-endpoints)
5. [Publishing Flow](#publishing-flow)
6. [Discovery Flow](#discovery-flow)
7. [Data Structures](#data-structures)
8. [Error Handling](#error-handling)
9. [Rate Limiting](#rate-limiting)
10. [Implementation Notes](#implementation-notes)

---

## Overview

### Registry Status

**✅ Production Ready** - All core registry features are complete and tested.

The MUXI Registry provides a complete backend API for:
- ✅ Formation publishing (with GitHub integration)
- ✅ Formation discovery (lazy loading from GitHub)
- ✅ Organization support
- ✅ Stats collection and tracking
- ✅ Version management
- ✅ Download tracking

### Architecture Shift (Important!)

**The registry is now the gatekeeper** for all GitHub operations:

**OLD Approach** (Previously):
- ❌ CLI creates GitHub repo
- ❌ CLI creates releases
- ❌ CLI uploads assets
- ❌ Problem: CLI needs GitHub OAuth token

**NEW Approach** (Current - Production):
- ✅ CLI only needs `mxr_` token (simple authentication)
- ✅ CLI zips formation and uploads to registry
- ✅ **Registry** has users' GitHub OAuth tokens (stored encrypted)
- ✅ **Registry** handles all GitHub operations
- ✅ **Registry** creates repos, pushes files, creates releases

**This means the CLI is simpler**: Just authenticate, zip, and upload!

---

## Architecture

### System Overview

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   CLI User   │         │   Registry   │         │    GitHub    │
│              │         │     API      │         │              │
└──────┬───────┘         └──────┬───────┘         └──────┬───────┘
       │                        │                        │
       │ 1. muxi login          │                        │
       │───────────────────────>│                        │
       │                        │ GitHub OAuth flow      │
       │                        │<──────────────────────>│
       │<───────────────────────│                        │
       │ Token: mxr_xxx         │                        │
       │                        │                        │
       │ 2. muxi push           │                        │
       │ (sends formation.zip)  │                        │
       │───────────────────────>│                        │
       │                        │ Create repo            │
       │                        │───────────────────────>│
       │                        │ Push files             │
       │                        │───────────────────────>│
       │                        │ Create release         │
       │                        │───────────────────────>│
       │                        │ Upload asset           │
       │                        │───────────────────────>│
       │<───────────────────────│                        │
       │ Success!               │                        │
       │                        │                        │
       │ 3. muxi pull @user/f   │                        │
       │───────────────────────>│                        │
       │ Get metadata           │                        │
       │<───────────────────────│                        │
       │                        │                        │
       │ 4. Download bundle.zip directly from GitHub    │
       │────────────────────────────────────────────────>│
       │<────────────────────────────────────────────────│
       │                        │                        │
       │ 5. Track download      │                        │
       │───────────────────────>│                        │
       │<───────────────────────│                        │
```

### Naming Conventions

**Formation References:**
```
@username/formation-name              # Latest version
@username/formation-name:1.0.0        # Specific version
```

**GitHub Repo Mapping:**
```
@ranaroussi/customer-support  →  github.com/ranaroussi/muxi-customer-support
@muxi/customer-support        →  github.com/muxi-ai/muxi-customer-support
```

**Reserved Usernames:**
Some organizations have reserved/shortened registry usernames:

| GitHub Account | Registry Username | Example Formation |
|----------------|-------------------|-------------------|
| `muxi-ai` | `@muxi` | `@muxi/customer-support` |
| `ranaroussi` | `@ranaroussi` | `@ranaroussi/my-formation` |

**Repo Prefix:** All formation repos use `muxi-` prefix on GitHub.

---

## Authentication

### Flow Overview

The CLI authenticates with the registry via GitHub OAuth, but the user experience is streamlined:

1. **User runs** `muxi login`
2. **CLI opens browser** to registry OAuth page
3. **User authenticates** with GitHub
4. **Registry** stores GitHub OAuth token (encrypted) and returns CLI token
5. **CLI** saves `mxr_` token locally

### CLI Token Format

```
mxr_{60_random_alphanumeric_chars}
```

Example: `mxr_5Jw9k2Lp8Nm4Qr6Ts1Uv3Wx7Yz0Ab2Cd4Ef6Gh8Ij0Kl2Mn4Op6Qr8St0`

### Storage

**Registry Auth File:** `~/.muxi/cli/registries.yaml`

```yaml
version: "1.0"
default_registry: registry.muxi.org

registries:
  registry.muxi.org:
    token: mxr_5Jw9k2Lp8Nm4Qr6Ts1Uv3Wx7Yz0Ab2Cd4Ef6Gh8Ij0Kl2Mn4Op6Qr8St0
    username: ranaroussi
    created_at: "2025-10-29T10:30:00Z"
```

**Permissions:** `600` (user read/write only)

**Token Expiration:** Tokens currently don't expire, but this may change.

---

## API Endpoints

### Base URL

```
Production: https://registry.muxi.org
Development: http://localhost:8080 (or https://muxi.registry)
```

### Authentication Header

```http
Authorization: Bearer mxr_5Jw9k2Lp8Nm4Qr6Ts1Uv3Wx7Yz0Ab2Cd4Ef6Gh8Ij0Kl2Mn4Op6Qr8St0
```

---

### 1. Authentication Endpoints

#### `GET /auth/cli/authorize`

Start CLI authentication flow.

**Request:**
```http
GET /auth/cli/authorize HTTP/1.1
Host: registry.muxi.org
```

**Response:** HTML page with GitHub OAuth flow

**CLI Usage:**
1. Open browser to this URL
2. User logs in with GitHub
3. User redirects to callback URL
4. CLI receives token via callback or polling

---

### 2. Formation Discovery (Public)

#### `GET /api/formations/@:user/:name`

Get formation metadata (info only, no download tracking).

**Authentication:** Optional (higher rate limits if authenticated)

**Request:**
```http
GET /api/formations/@ranaroussi/customer-support HTTP/1.1
Host: registry.muxi.org
Authorization: Bearer mxr_xxx (optional)
```

**Response:**
```json
{
  "success": true,
  "formation": {
    "user": "ranaroussi",
    "name": "customer-support",
    "description": "AI-powered customer support with intelligent escalation",
    "latest_version": "1.2.0",
    "github_repo": "ranaroussi/muxi-customer-support",
    "github_stars": 45,
    "total_downloads": 1234,
    "created_at": "2025-01-10T14:30:00Z",
    "updated_at": "2025-01-15T10:20:00Z"
  },
  "stats": {
    "agents_count": 3,
    "mcps_count": 2,
    "sops_count": 1,
    "triggers_count": 2,
    "knowledge_count": 0
  },
  "latest_version": {
    "version": "1.2.0",
    "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.2.0/bundle.zip",
    "size_bytes": 29593,
    "published_at": "2025-01-15T10:20:00Z"
  }
}
```

**Use Case:** Check formation info before pulling, display metadata.

**Note:** This does NOT increment download counter. Use `?pull=true` for that.

---

#### `GET /api/formations/@:user/:name?pull=true`

Get formation metadata AND track download.

**Authentication:** Optional (but recommended for higher rate limits)

**Request:**
```http
GET /api/formations/@ranaroussi/customer-support?pull=true HTTP/1.1
Host: registry.muxi.org
Authorization: Bearer mxr_xxx (optional)
```

**Response:** Same as above, but increments download counter.

**Use Case:** When user actually pulls/downloads formation.

---

#### `GET /api/formations/@:user/:name:version`

Get specific version metadata.

**Authentication:** Optional

**Request:**
```http
GET /api/formations/@ranaroussi/customer-support:1.0.0 HTTP/1.1
Host: registry.muxi.org
```

**Response:**
```json
{
  "success": true,
  "formation": {
    "user": "ranaroussi",
    "name": "customer-support",
    "description": "...",
    "latest_version": "1.2.0",
    "github_repo": "ranaroussi/muxi-customer-support",
    "github_stars": 45,
    "total_downloads": 1234
  },
  "version": {
    "version": "1.0.0",
    "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.0.0/bundle.zip",
    "size_bytes": 28500,
    "published_at": "2024-12-10T09:15:00Z"
  },
  "stats": {
    "agents_count": 2,
    "mcps_count": 1,
    "sops_count": 1,
    "triggers_count": 1,
    "knowledge_count": 0
  }
}
```

**With Pull Tracking:**
```http
GET /api/formations/@ranaroussi/customer-support:1.0.0?pull=true
```

---

#### `GET /api/formations/@:user/:name/versions`

List all versions of a formation.

**Authentication:** Optional

**Request:**
```http
GET /api/formations/@ranaroussi/customer-support/versions HTTP/1.1
Host: registry.muxi.org
```

**Response:**
```json
{
  "success": true,
  "formation": {
    "user": "ranaroussi",
    "name": "customer-support",
    "latest_version": "1.2.0"
  },
  "versions": [
    {
      "version": "1.2.0",
      "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.2.0/bundle.zip",
      "size_bytes": 29593,
      "download_count": 234,
      "published_at": "2025-01-15T10:20:00Z"
    },
    {
      "version": "1.1.0",
      "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.1.0/bundle.zip",
      "size_bytes": 28900,
      "download_count": 456,
      "published_at": "2025-01-01T08:00:00Z"
    },
    {
      "version": "1.0.0",
      "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.0.0/bundle.zip",
      "size_bytes": 28500,
      "download_count": 544,
      "published_at": "2024-12-10T09:15:00Z"
    }
  ]
}
```

---

#### `GET /api/search`

Search for formations.

**Authentication:** Optional

**Request:**
```http
GET /api/search?q=customer+support&limit=20&sort=downloads HTTP/1.1
Host: registry.muxi.org
```

**Query Parameters:**
- `q` - Search query (required)
- `limit` - Max results (default: 20, max: 100)
- `sort` - Sort by: `relevance`, `downloads`, `stars`, `updated` (default: `relevance`)

**Response:**
```json
{
  "success": true,
  "query": "customer support",
  "total": 3,
  "results": [
    {
      "user": "ranaroussi",
      "name": "customer-support",
      "description": "AI-powered customer support with intelligent escalation",
      "latest_version": "1.2.0",
      "github_repo": "ranaroussi/muxi-customer-support",
      "github_stars": 45,
      "total_downloads": 1234,
      "size_bytes": 29593
    },
    {
      "user": "somedev",
      "name": "support-bot",
      "description": "Simple support automation bot",
      "latest_version": "2.1.0",
      "github_repo": "somedev/muxi-support-bot",
      "github_stars": 12,
      "total_downloads": 450,
      "size_bytes": 15200
    }
  ]
}
```

---

### 3. Formation Publishing (Authenticated)

#### `POST /api/formations/publish`

Publish formation to registry (creates/updates GitHub repo).

**Authentication:** Required

**Request:**
```http
POST /api/formations/publish HTTP/1.1
Host: registry.muxi.org
Authorization: Bearer mxr_xxx
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="formation.zip"
Content-Type: application/zip

[binary zip data]
------WebKitFormBoundary
Content-Disposition: form-data; name="org"

muxi-ai
------WebKitFormBoundary--
```

**Form Fields:**
- `file` - Formation ZIP file (required)
- `org` - GitHub organization name (optional, for org publishing)

**ZIP Contents Requirements:**
- `formation.yaml` - Must exist and be valid
- `formation.version` - Must be valid semver (e.g., "1.0.0")
- Other files: agents/, mcps/, sops/, triggers/, knowledge/, README.md

**Response (Success):**
```json
{
  "success": true,
  "message": "Formation published successfully",
  "formation": {
    "user": "ranaroussi",
    "name": "customer-support",
    "version": "1.2.0",
    "github_repo": "ranaroussi/muxi-customer-support",
    "registry_url": "https://registry.muxi.org/@ranaroussi/customer-support",
    "github_url": "https://github.com/ranaroussi/muxi-customer-support",
    "release_url": "https://github.com/ranaroussi/muxi-customer-support/releases/tag/v1.2.0"
  },
  "stats": {
    "agents_count": 3,
    "mcps_count": 2,
    "sops_count": 1,
    "triggers_count": 2,
    "knowledge_count": 0
  }
}
```

**Response (Organization Publishing):**
```json
{
  "success": true,
  "message": "Formation published to organization",
  "formation": {
    "user": "muxi",
    "name": "customer-support",
    "version": "1.0.0",
    "github_repo": "muxi-ai/muxi-customer-support",
    "registry_url": "https://registry.muxi.org/@muxi/customer-support",
    "github_url": "https://github.com/muxi-ai/muxi-customer-support",
    "release_url": "https://github.com/muxi-ai/muxi-customer-support/releases/tag/v1.0.0",
    "published_by": "ranaroussi"
  }
}
```

**What the Registry Does:**
1. ✅ Validates ZIP and formation.yaml
2. ✅ Checks version doesn't already exist
3. ✅ Creates/verifies GitHub repository
4. ✅ Pushes all formation files to GitHub
5. ✅ Creates git tag (e.g., `v1.0.0`)
6. ✅ Creates GitHub release
7. ✅ Uploads formation.zip as release asset
8. ✅ Analyzes formation structure
9. ✅ Stores metadata in registry database
10. ✅ Returns success with URLs

**Error Responses:**

**Not Authenticated:**
```json
{
  "success": false,
  "error": "Authentication required",
  "message": "Please provide a valid Bearer token"
}
```
Status: `401 Unauthorized`

**Invalid formation.yaml:**
```json
{
  "success": false,
  "error": "Invalid formation.yaml",
  "message": "Missing required field: version"
}
```
Status: `400 Bad Request`

**Version Already Exists:**
```json
{
  "success": false,
  "error": "Version already exists",
  "message": "Version 1.0.0 already exists for @ranaroussi/customer-support",
  "existing_version": {
    "version": "1.0.0",
    "published_at": "2024-12-10T09:15:00Z",
    "release_url": "https://github.com/ranaroussi/muxi-customer-support/releases/tag/v1.0.0"
  }
}
```
Status: `409 Conflict`

**GitHub API Error:**
```json
{
  "success": false,
  "error": "GitHub API error",
  "message": "Failed to create repository: API rate limit exceeded"
}
```
Status: `502 Bad Gateway`

---

## Publishing Flow

### Step-by-Step CLI Flow

**1. Validate Formation Locally**
```
Check formation.yaml exists
Parse formation.yaml
Validate version is semver
Check required files exist
```

**2. Create ZIP Bundle**
```
Include:
  - formation.yaml
  - README.md
  - agents/
  - mcps/
  - sops/
  - triggers/
  - knowledge/

Exclude:
  - .git/
  - .muxi/
  - secrets.enc (warn if present!)
  - node_modules/
  - __pycache__/
  - .DS_Store
  - *.pyc
```

**3. Upload to Registry**
```http
POST /api/formations/publish
Authorization: Bearer mxr_xxx
Content-Type: multipart/form-data

[formation.zip]
```

**4. Registry Processes (Behind the Scenes)**
```
Registry:
  → Validates ZIP
  → Parses formation.yaml
  → Checks user's GitHub OAuth token
  → Creates/updates GitHub repo
  → Pushes files via GitHub Contents API
  → Creates tag v{version}
  → Creates GitHub release
  → Uploads ZIP as release asset
  → Analyzes structure (counts agents, MCPs, etc.)
  → Stores in database
  → Returns success
```

**5. Display Success to User**
```
✓ Published @ranaroussi/customer-support v1.2.0!

View at:
  Registry: https://registry.muxi.org/@ranaroussi/customer-support
  GitHub:   https://github.com/ranaroussi/muxi-customer-support

Share with:
  muxi pull @ranaroussi/customer-support
```

---

### Organization Publishing

To publish to an organization, include the `org` form field:

```http
POST /api/formations/publish
Authorization: Bearer mxr_xxx
Content-Type: multipart/form-data

file=formation.zip
org=muxi-ai
```

**Requirements:**
- User must have access to the organization on GitHub
- Organization must exist in registry (or be mappable via reserved_usernames)

**Example:**
```bash
# Publishes to github.com/muxi-ai/muxi-customer-support
# Shows as @muxi/customer-support on registry
```

---

## Discovery Flow

### Pulling a Formation

**1. CLI Requests Metadata**
```http
GET /api/formations/@ranaroussi/customer-support?pull=true
Authorization: Bearer mxr_xxx (optional but recommended)
```

**2. Registry Returns Metadata**
```json
{
  "success": true,
  "formation": {...},
  "latest_version": {
    "version": "1.2.0",
    "download_url": "https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.2.0/bundle.zip",
    "size_bytes": 29593
  }
}
```

**3. CLI Downloads ZIP Directly from GitHub**
```http
GET https://github.com/ranaroussi/muxi-customer-support/releases/download/v1.2.0/bundle.zip
```

This is a direct GitHub download, doesn't go through registry bandwidth.

**4. CLI Extracts ZIP**
```
Extract to ./customer-support/
  - formation.yaml
  - README.md
  - agents/
  - mcps/
  - sops/
  - triggers/
```

**5. Done!**
```
✓ Downloaded @ranaroussi/customer-support v1.2.0
  Location: ./customer-support/
```

---

### Lazy Discovery

If formation not in registry database:

**1. Registry Checks GitHub**
```
User visits: @ranaroussi/formation-name
  ↓
Registry: Not in database
  ↓
Check: github.com/ranaroussi/muxi-formation-name
  ↓
Found! Fetch metadata
  ↓
Cache in database
  ↓
Return to user
```

**2. What Registry Fetches**
- Repo description
- Latest release info
- README.md
- Star count
- Downloads release asset (or main branch if no releases)
- Analyzes formation structure
- Stores stats

**3. Next Request**
```
Served instantly from cache!
```

---

## Data Structures

### Formation Object

```typescript
interface Formation {
  user: string;              // Registry username (e.g., "ranaroussi", "muxi")
  name: string;              // Formation name (e.g., "customer-support")
  description: string;       // From formation.yaml
  latest_version: string;    // Semver (e.g., "1.2.0")
  github_repo: string;       // Full repo (e.g., "ranaroussi/muxi-customer-support")
  github_stars: number;      // Star count from GitHub
  total_downloads: number;   // All-time download count
  created_at: string;        // ISO 8601 timestamp
  updated_at: string;        // ISO 8601 timestamp
}
```

### Version Object

```typescript
interface Version {
  version: string;           // Semver (e.g., "1.0.0")
  download_url: string;      // GitHub release asset URL
  size_bytes: number;        // Bundle size
  download_count: number;    // Downloads of this version
  published_at: string;      // ISO 8601 timestamp
}
```

### Formation Stats

```typescript
interface FormationStats {
  agents_count: number;      // Number of agent files in agents/
  mcps_count: number;        // Number of MCP files in mcps/
  sops_count: number;        // Number of SOP files in sops/
  triggers_count: number;    // Number of trigger files in triggers/
  knowledge_count: number;   // Number of .md files in knowledge/
}
```

**Pattern Matching:**
- Agents: `.py` files in `agents/` directory
- MCPs: `.py` files in `mcps/` directory  
- SOPs: `.md` files in `sops/` directory
- Triggers: `.py` files in `triggers/` directory
- Knowledge: `.md` files in `knowledge/` directory (excluding README.md)

---

## Error Handling

### HTTP Status Codes

| Status | Meaning | Example |
|--------|---------|---------|
| `200` | Success | Formation found and returned |
| `201` | Created | Formation published successfully |
| `400` | Bad Request | Invalid formation.yaml |
| `401` | Unauthorized | Missing or invalid token |
| `403` | Forbidden | Token valid but insufficient permissions |
| `404` | Not Found | Formation doesn't exist |
| `409` | Conflict | Version already exists |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Server error |
| `502` | Bad Gateway | GitHub API error |

### Error Response Format

```json
{
  "success": false,
  "error": "Error type",
  "message": "Human-readable error message",
  "details": {
    // Optional additional context
  }
}
```

### Common Errors

**Not Authenticated:**
```json
{
  "success": false,
  "error": "Authentication required",
  "message": "Please authenticate with: muxi login"
}
```

**Formation Not Found:**
```json
{
  "success": false,
  "error": "Formation not found",
  "message": "Formation @ranaroussi/nonexistent not found in registry or GitHub"
}
```

**Rate Limited:**
```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "message": "You have exceeded the rate limit. Try again in 45 minutes.",
  "details": {
    "limit": 60,
    "remaining": 0,
    "reset_at": "2025-10-29T12:00:00Z"
  }
}
```

---

## Rate Limiting

### Anonymous Requests (by IP)

**Limits:**
- 5 requests per second
- 100 requests per 10 minutes

**Applies to:**
- GET /api/formations/...
- GET /api/search

**Headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1698765432
```

### Authenticated Requests (by user)

**Limits:**
- 10 requests per second
- 1000 requests per 10 minutes

**Applies to:**
- All endpoints with valid Bearer token

**Headers:**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 945
X-RateLimit-Reset: 1698765432
X-RateLimit-User: ranaroussi
```

### Recommendation

**Always send Bearer token**, even for public endpoints, to get higher rate limits:

```http
GET /api/formations/@user/name
Authorization: Bearer mxr_xxx
```

---

## Implementation Notes

### Registry Auth Storage

**File:** `~/.muxi/cli/registries.yaml`

```yaml
version: "1.0"
default_registry: registry.muxi.org

registries:
  registry.muxi.org:
    token: mxr_...
    username: ranaroussi
    created_at: "2025-10-29T10:30:00Z"
```

**Permissions:** `600` (user read/write only)

### Bundle Creation

**Files to Include:**
```
formation.yaml          # Required
README.md              # Recommended
agents/*.py            # All Python files
mcps/*.py              # All Python files
sops/*.md              # All Markdown files
triggers/*.py          # All Python files
knowledge/*.md         # All Markdown files
```

**Files to Exclude:**
```
.git/                  # Git directory
.muxi/                 # CLI cache
secrets.enc            # ⚠️ Warn user if present!
.env                   # Environment files
node_modules/          # Dependencies
__pycache__/           # Python cache
*.pyc                  # Compiled Python
.DS_Store              # macOS files
```

**Compression:** Standard ZIP format

### Version Resolution

```
@user/formation              → latest_version from registry
@user/formation:1.0.0        → specific version
@user/formation:1.x          → (future) resolve to latest 1.x
```

### Error Recovery

**Network Errors:**
```
Retry with exponential backoff:
  - Wait 1s, retry
  - Wait 2s, retry
  - Wait 4s, retry
  - Give up after 3 attempts
```

**GitHub Rate Limits:**
```
If GitHub rate limited:
  - Display remaining time to reset
  - Suggest authentication (higher limits)
  - Exit gracefully
```

---

## Quick Reference

### CLI Commands (Recommended)

```bash
# Authenticate
muxi login

# Publish
muxi push                          # Personal account
muxi push --org muxi-ai            # Organization

# Discovery
muxi pull @user/formation          # Latest version
muxi pull @user/formation:1.0.0    # Specific version
muxi search "customer support"     # Search

# Info
muxi show @user/formation          # Formation details
```

### API Endpoints (Summary)

```
Authentication:
  GET  /auth/cli/authorize

Discovery (Public):
  GET  /api/formations/@:user/:name
  GET  /api/formations/@:user/:name?pull=true
  GET  /api/formations/@:user/:name:version
  GET  /api/formations/@:user/:name/versions
  GET  /api/search?q=query

Publishing (Authenticated):
  POST /api/formations/publish
```

---

## See Also

- [CLI Command Design](./CLI-COMMAND-DESIGN.md) - Complete CLI specification
- [Implementation Plan](./IMPLEMENTATION-PLAN.md) - CLI development roadmap
- [Registry Documentation](../../registry/docs/) - Full registry documentation

---

**End of Registry API Reference**

**Status:** ✅ Registry is production-ready  
**Last Updated:** 2025-10-29  
**Registry Version:** v1.0 (Phase 2 Complete)
