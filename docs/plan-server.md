# Server Commands Implementation Plan

**Date:** 2025-12-02  
**Status:** Planning  
**Priority:** HIGH  
**API Spec:** `../schemas/api/server-api-v1-final.yaml`

---

## Overview

Implement CLI commands for interacting with MUXI Server. The server manages formation deployment, lifecycle, and monitoring via a REST API with HMAC authentication.

**Server URL:** Default port 7890

---

## Implementation Phases

### Phase 1: Foundation (Server Profiles + Auth)
1. Server profile storage (`~/.muxi/cli/servers.yaml`)
2. HMAC authentication
3. `muxi server add` - Add server profile
4. `muxi server list` - List profiles
5. `muxi server set` - Set default profile
6. `muxi server remove` - Remove profile

### Phase 2: Basic Operations
7. `muxi server status` - Server health/status
8. `muxi formation list` - List deployed formations
9. `muxi formation get <id>` - Get formation details

### Phase 3: Deployment
10. `muxi deploy` - Deploy formation to server
11. `muxi formation update <id>` - Update existing formation

### Phase 4: Lifecycle Management
12. `muxi formation stop <id>` - Stop formation
13. `muxi formation restart <id>` - Restart formation
14. `muxi formation rollback <id>` - Rollback to previous version
15. `muxi formation delete <id>` - Delete formation
16. `muxi logs <id>` - Get formation logs

---

## Phase 1: Foundation

### 1.1 Server Profile Storage

**File:** `~/.muxi/cli/servers.yaml`

```yaml
version: "1.0"
default: localhost

servers:
  localhost:
    url: http://localhost:7890
    key_id: MUXI_abc123
    secret_key: sk_xyz789...
    added_at: "2025-12-02T10:00:00Z"
    
  production:
    url: https://api.company.com:7890
    key_id: MUXI_prod_key
    secret_key: sk_prod_secret...
    added_at: "2025-12-02T10:00:00Z"
```

**Implementation:** `pkg/server/config.go`

```go
type ServerConfig struct {
    Version  string                   `yaml:"version"`
    Default  string                   `yaml:"default"`
    Servers  map[string]ServerEntry   `yaml:"servers"`
}

type ServerEntry struct {
    URL       string    `yaml:"url"`
    KeyID     string    `yaml:"key_id"`
    SecretKey string    `yaml:"secret_key"`
    AddedAt   time.Time `yaml:"added_at"`
}

func LoadServers() (*ServerConfig, error)
func SaveServers(config *ServerConfig) error
func GetDefaultServer() string
func GetServer(name string) (*ServerEntry, error)
```

### 1.2 HMAC Authentication

**Format:**
```
Authorization: MUXI-HMAC key={KEY_ID}, timestamp={UNIX_TIMESTAMP}, signature={BASE64_SIGNATURE}
```

**Signature Generation:**
```
message = "{timestamp};{method};{path}"
signature = base64(HMAC-SHA256(secret_key, message))
```

**Implementation:** `pkg/server/auth.go`

```go
func GenerateHMACSignature(secretKey, method, path string) (string, int64) {
    timestamp := time.Now().Unix()
    message := fmt.Sprintf("%d;%s;%s", timestamp, method, path)
    
    mac := hmac.New(sha256.New, []byte(secretKey))
    mac.Write([]byte(message))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    return signature, timestamp
}

func BuildAuthHeader(keyID, secretKey, method, path string) string {
    signature, timestamp := GenerateHMACSignature(secretKey, method, path)
    return fmt.Sprintf("MUXI-HMAC key=%s, timestamp=%d, signature=%s", 
        keyID, timestamp, signature)
}
```

### 1.3 Server Client

**Implementation:** `pkg/server/client.go`

```go
type Client struct {
    BaseURL    string
    KeyID      string
    SecretKey  string
    HTTPClient *http.Client
}

func NewClient(profile string) (*Client, error)
func NewClientFromEntry(entry *ServerEntry) *Client

func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error)
func (c *Client) Get(path string) (*http.Response, error)
func (c *Client) Post(path string, body io.Reader) (*http.Response, error)
func (c *Client) Put(path string, body io.Reader) (*http.Response, error)
func (c *Client) Delete(path string) (*http.Response, error)
```

### 1.4 Commands: Server Profile Management

**`muxi server add`**
```
muxi server add

Add Server

  Server Name: production
  Server URL: https://api.company.com:7890
  
  Authentication (from ~/.muxi-server/credentials.yaml on server):
  Key ID: MUXI_abc123
  Secret Key: ••••••••••••••••

  Testing connection... OK
  ✓ Server "production" added

  Set as default? [Y/n]: y
  ✓ Default server set to: production
```

**`muxi server list`**
```
muxi server list

  NAME         URL                              STATUS
  localhost    http://localhost:7890            ● online
  production   https://api.company.com:7890     ○ offline
  * = default
```

**`muxi server set`**
```
muxi server set

Select default server:
  ◉ localhost [current]
  ○ production

✓ Default server set to: production
```

**`muxi server remove`**
```
muxi server remove

Select server to remove:
  ○ localhost
  ◉ production

Remove "production"? [y/N]: y
✓ Server "production" removed
```

---

## Phase 2: Basic Operations

### 2.1 `muxi server status`

**API:** `GET /rpc/server/status`

**Output:**
```
muxi server status

Server: localhost (http://localhost:7890)

  Status:     ● healthy
  Version:    1.0.0-dev
  Uptime:     2d 5h 30m
  
  Formations: 3 total (2 running, 1 stopped)
  Port Pool:  8000-9000 (997 available)
  
  Runtime:    docker (muxi-runtime:1.2.0)
  Platform:   darwin/arm64
```

**With `--json` flag:** Output raw JSON response.

### 2.2 `muxi formation list`

**API:** `GET /rpc/formations`

**Output:**
```
muxi formation list

  ID              STATUS     PORT    VERSION   UPTIME
  chat-api        ● running  8001    1.2.0     5d 12h
  workflow-bot    ● running  8002    2.0.1     2d 3h
  test-formation  ○ stopped  -       1.0.0     -
```

**With `--profile` flag:** Use specific server profile.

### 2.3 `muxi formation get <id>`

**API:** `GET /rpc/formations/{id}`

**Output:**
```
muxi formation get chat-api

Formation: chat-api

  Status:     ● running
  Port:       8001
  Version:    1.2.0 (previous: 1.1.0)
  Uptime:     5d 12h 30m
  
  Runtime:    docker (muxi-runtime:1.2.0)
  PID:        12345
  
  Health:     ✓ healthy (last check: 30s ago)
  Restarts:   0
  
  Deployed:   2025-11-27 10:30:00
  Updated:    2025-11-30 15:45:00
```

---

## Phase 3: Deployment

### 3.1 `muxi deploy`

**API:** `POST /rpc/formations` (multipart/gzip)

**Flow:**
1. Validate formation (run `muxi validate`)
2. Create deployment bundle (tar.gz)
3. Upload to server
4. Show deployment progress
5. Wait for health check

**Output:**
```
muxi deploy

Deploying to: localhost

  Validating formation... OK
  Creating bundle... OK (15 files, 24 KB)
  
  ⠹ Uploading to server...
  ✓ Uploaded

  ⠹ Starting formation...
  ✓ Formation started on port 8003

  ⠹ Waiting for health check...
  ✓ Health check passed

✓ Deployed chat-api v1.0.0

  Formation URL: http://localhost:7890/api/chat-api
  Direct URL:    http://localhost:8003
```

**Options:**
- `--profile <name>` - Deploy to specific server
- `--dry-run` - Validate and create bundle without deploying

### 3.2 `muxi formation update <id>`

**API:** `PUT /rpc/formations/{id}` (multipart/gzip)

**Output:**
```
muxi formation update chat-api

Updating: chat-api on localhost

  Current version: 1.1.0
  New version:     1.2.0

  Creating bundle... OK
  
  ⠹ Stopping current version...
  ✓ Stopped
  
  ⠹ Deploying new version...
  ✓ Deployed
  
  ⠹ Health check...
  ✓ Healthy

✓ Updated chat-api to v1.2.0

  Rollback available: muxi formation rollback chat-api
```

---

## Phase 4: Lifecycle Management

### 4.1 `muxi formation stop <id>`

**API:** `POST /rpc/formations/{id}/stop`

```
muxi formation stop chat-api

  ⠹ Stopping chat-api...
  ✓ Formation stopped
```

### 4.2 `muxi formation restart <id>`

**API:** `POST /rpc/formations/{id}/restart`

```
muxi formation restart chat-api

  ⠹ Restarting chat-api...
  ✓ Formation restarted on port 8001
```

### 4.3 `muxi formation rollback <id>`

**API:** `POST /rpc/formations/{id}/rollback`

```
muxi formation rollback chat-api

  Current:  v1.2.0
  Previous: v1.1.0

  Roll back to v1.1.0? [y/N]: y

  ⠹ Rolling back...
  ✓ Rolled back to v1.1.0
```

### 4.4 `muxi formation delete <id>`

**API:** `DELETE /rpc/formations/{id}`

```
muxi formation delete test-formation

  Delete "test-formation"? [y/N]: y

  ⠹ Deleting formation...
  ✓ Formation deleted
```

### 4.5 `muxi logs <id>`

**API:** `GET /rpc/formations/{id}/logs`

```
muxi logs chat-api

[2025-12-02 10:30:15] INFO  Server started on port 8001
[2025-12-02 10:30:16] INFO  Health check endpoint ready
[2025-12-02 10:31:00] INFO  Received request: POST /chat
[2025-12-02 10:31:02] INFO  Response sent (200, 1.2s)
```

**Options:**
- `--lines <n>` - Number of lines (default 100)
- `--follow` - Stream logs in real-time (future)
- `--stream stdout|stderr|all` - Filter by stream

---

## File Structure

```
pkg/server/
├── config.go      # ServerConfig, LoadServers, SaveServers
├── auth.go        # HMAC signature generation
├── client.go      # HTTP client with auth
├── types.go       # API response types
└── bundle.go      # Deployment bundle creation

cmd/
├── server.go      # muxi server add/list/set/remove
├── formation.go   # muxi formation list/get/stop/restart/rollback/delete
├── deploy.go      # muxi deploy
└── logs.go        # muxi logs
```

---

## API Response Types

**Implementation:** `pkg/server/types.go`

```go
// Common response wrapper
type APIResponse struct {
    Success bool            `json:"success"`
    Data    json.RawMessage `json:"data,omitempty"`
    Error   string          `json:"error,omitempty"`
    Message string          `json:"message,omitempty"`
    Code    int             `json:"code,omitempty"`
}

// GET /rpc/server/status
type ServerStatus struct {
    Server struct {
        Status  string `json:"status"`
        Version string `json:"version"`
        Uptime  int64  `json:"uptime"`
    } `json:"server"`
    Formations struct {
        Total   int `json:"total"`
        Running int `json:"running"`
        Stopped int `json:"stopped"`
    } `json:"formations"`
    Runtime struct {
        Type    string `json:"type"`
        Version string `json:"version"`
    } `json:"runtime"`
}

// GET /rpc/formations
type FormationListItem struct {
    ID      string `json:"id"`
    Status  string `json:"status"`
    Port    int    `json:"port"`
    Version string `json:"version"`
    Uptime  int64  `json:"uptime"`
    Healthy bool   `json:"healthy"`
}

// GET /rpc/formations/{id}
type FormationDetail struct {
    ID              string `json:"id"`
    Status          string `json:"status"`
    Port            int    `json:"port"`
    CurrentVersion  string `json:"current_version"`
    PreviousVersion string `json:"previous_version"`
    Uptime          int64  `json:"uptime"`
    PID             int    `json:"pid"`
    Healthy         bool   `json:"healthy"`
    RestartCount    int    `json:"restart_count"`
    DeployedAt      string `json:"deployed_at"`
    UpdatedAt       string `json:"updated_at"`
}

// POST /rpc/formations (deploy response)
type DeployResponse struct {
    ID      string `json:"id"`
    Port    int    `json:"port"`
    Version string `json:"version"`
    Status  string `json:"status"`
}

// GET /rpc/formations/{id}/logs
type LogsResponse struct {
    FormationID string   `json:"formation_id"`
    Lines       []string `json:"lines"`
    Stream      string   `json:"stream"`
}
```

---

## Error Handling

**Server Errors:**
```go
// Common error responses
var (
    ErrUnauthorized    = errors.New("authentication failed - check credentials")
    ErrNotFound        = errors.New("formation not found")
    ErrConflict        = errors.New("formation already exists")
    ErrServerError     = errors.New("server error - check server logs")
    ErrConnectionError = errors.New("cannot connect to server")
)

func handleAPIError(resp *http.Response) error {
    switch resp.StatusCode {
    case 401:
        return ErrUnauthorized
    case 404:
        return ErrNotFound
    case 409:
        return ErrConflict
    case 500:
        return ErrServerError
    default:
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
}
```

---

## Testing

### Unit Tests
- HMAC signature generation
- Config loading/saving
- Bundle creation

### Integration Tests (with mock server)
- Authentication flow
- Deploy/update flow
- Error handling

### Manual Testing
```bash
# Start local server
cd ../server && ./muxi-server serve

# Test commands
muxi server add
muxi server status
muxi deploy
muxi formation list
muxi logs my-formation
```

---

## Timeline

| Phase | Tasks | Estimate |
|-------|-------|----------|
| 1 | Foundation (profiles, auth, client) | 4 hours |
| 2 | Basic operations (status, list, get) | 2 hours |
| 3 | Deployment (deploy, update) | 3 hours |
| 4 | Lifecycle (stop, restart, rollback, delete, logs) | 3 hours |
| - | Testing & polish | 2 hours |
| **Total** | | **14 hours** |

---

## Dependencies

- Server must be running for integration testing
- Server credentials from `~/.muxi-server/credentials.yaml`
- Formation bundle format (tar.gz with formation.yaml)

---

## Notes

1. **Profile Resolution:** Same as registry (flag > .muxi > default)
2. **Bundle Format:** Same as registry push bundle
3. **Health Checks:** Server handles health checks, CLI just reports status
4. **Log Streaming:** `--follow` flag is future enhancement (needs WebSocket/SSE)
