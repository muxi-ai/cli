# API Reference

**Last Updated:** 2025-12-08

---

## API Specifications

The full OpenAPI specs are maintained in the schemas repository:

- **Server API:** [../schemas/api/server-api-v1-final.yaml](../../schemas/api/server-api-v1-final.yaml)
- **Formation API:** [../schemas/api/formation-api-v1-final.yaml](../../schemas/api/formation-api-v1-final.yaml)

This document covers CLI-specific implementation details.

---

## Server Connection

### Default Configuration
| Item | Value |
|------|-------|
| Default Port | 7890 |
| Auth Type | HMAC-SHA256 |
| Management API | `/rpc/*` |
| Formation Proxy | `/api/{id}/*` |

### Local Files
| File | Purpose |
|------|---------|
| `~/.muxi/profiles.yaml` | Server profiles |
| `~/.muxi/server/credentials.json` | Auto-generated local credentials |
| `~/.muxi/server/config.yaml` | Server configuration |

---

## HMAC Authentication

All `/rpc/*` endpoints require HMAC authentication.

### Signature Generation

```go
func GenerateHMACSignature(keyID, secretKey, method, path, timestamp, body string) string {
    // Build canonical string
    canonical := fmt.Sprintf("%s\n%s\n%s\n%s", method, path, timestamp, body)
    
    // HMAC-SHA256
    mac := hmac.New(sha256.New, []byte(secretKey))
    mac.Write([]byte(canonical))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    
    return signature
}
```

### Required Headers

```go
headers := map[string]string{
    "X-MUXI-Key-ID":    keyID,
    "X-MUXI-Timestamp": timestamp,  // RFC3339 format
    "X-MUXI-Signature": signature,
    "Content-Type":     "application/json",
}
```

### Example Request

```bash
# Generate timestamp
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Test HMAC auth
curl -X GET http://localhost:7890/rpc/formations \
  -H "X-MUXI-Key-ID: $KEY_ID" \
  -H "X-MUXI-Timestamp: $TIMESTAMP" \
  -H "X-MUXI-Signature: $SIGNATURE"
```

---

## Key Endpoints

### Public (No Auth)
```
GET /health     → {"status": "healthy"}
GET /ping       → {"message": "pong"}
```

### Formation Management (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/rpc/formations/deploy` | Deploy new formation |
| GET | `/rpc/formations` | List formations |
| GET | `/rpc/formations/{id}` | Get formation details |
| PUT | `/rpc/formations/{id}` | Update formation |
| DELETE | `/rpc/formations/{id}` | Delete formation |
| POST | `/rpc/formations/{id}/stop` | Stop formation |
| POST | `/rpc/formations/{id}/start` | Start formation |
| POST | `/rpc/formations/{id}/restart` | Restart formation |
| POST | `/rpc/formations/{id}/rollback` | Rollback version |
| GET | `/rpc/formations/{id}/logs` | Get logs |

### Server Management (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rpc/server/status` | Server status |
| GET | `/rpc/server/logs` | Audit logs |

---

## SSE Streaming

Deploy, start, restart, and rollback operations support SSE streaming for real-time progress.

### Request
```
POST /rpc/formations/deploy
Accept: text/event-stream
```

### Response Format
```
event: progress
data: {"stage":"extracting","message":"Extracting bundle..."}

event: progress
data: {"stage":"validating","message":"Validating formation.yaml"}

event: progress
data: {"stage":"downloading_sif","message":"Downloading runtime","progress":45.5}

event: complete
data: {"formation_id":"my-app","status":"running","port":8001}
```

### Progress Stages

**Deploy (new):**
1. `extracting` - Extracting bundle
2. `validating` - Validating formation.yaml
3. `resolving_runtime` - Resolving runtime version
4. `downloading_sif` - Downloading SIF (with progress %)
5. `pulling_runner` - Pulling Docker image
6. `spawning` - Starting process
7. `health_check` - Health check (with countdown)

**Update (existing):**
1. `extracting` - Extract to staging
2. `validating` - Validate formation
3. `resolving_runtime` - Resolve runtime
4. `downloading_sif` - Download SIF
5. `pulling_runner` - Pull runner
6. `spawning_staging` - Start staging version
7. `health_check` - Staging health check
8. `swapping` - Atomic switch
9. `stopping_old` - Stop old version

---

## Credentials File Format

Local server credentials (`~/.muxi/server/credentials.json`):

```json
{
  "key_id": "MUXI_LOCAL_abc123",
  "secret_key": "base64-encoded-secret-key",
  "server_id": "hostname-sha256hash",
  "created_at": "2025-10-24T12:00:00Z"
}
```

CLI auto-reads this for localhost profile.

---

## Error Responses

```json
{
  "error": {
    "code": "FORMATION_NOT_FOUND",
    "message": "Formation 'my-app' not found",
    "details": {}
  }
}
```

### Common Error Codes
| Code | HTTP | Description |
|------|------|-------------|
| `UNAUTHORIZED` | 401 | Invalid HMAC signature |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `FORMATION_NOT_FOUND` | 404 | Formation doesn't exist |
| `FORMATION_ALREADY_EXISTS` | 409 | Formation ID taken |
| `FORMATION_NOT_RUNNING` | 409 | Can't stop stopped formation |
| `VERSION_NOT_HIGHER` | 409 | Update requires higher version |

---

## Testing Server Connection

```bash
# Check server health (no auth)
curl http://localhost:7890/health

# Test with wrong auth (should return 401)
curl -X GET http://localhost:7890/rpc/formations \
  -H "X-MUXI-Key-ID: wrong" \
  -H "X-MUXI-Timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  -H "X-MUXI-Signature: wrong"
```

---

## Related Documentation

- [../schemas/api/server-api-v1-final.yaml](../../schemas/api/server-api-v1-final.yaml) - Full OpenAPI spec
- [architecture.md](architecture.md) - CLI architecture
- [../STATUS.md](../STATUS.md) - Implementation status
