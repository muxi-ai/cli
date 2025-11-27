# A2A Configuration Wizard

**Date:** 2025-11-27  
**Status:** ✅ Inbound Complete, ⏳ Outbound Pending  
**Command:** `muxi config a2a`

---

## Overview

The A2A (Agent-to-Agent) configuration wizard helps users set up inbound and outbound A2A communication in `formation.yaml`. This is a **formation-level configuration** (not a file creation command).

**Note:** This uses `muxi config` (not `muxi new`) because it **modifies** the existing `formation.yaml` file. See [COMMAND-SEMANTICS.md](./COMMAND-SEMANTICS.md) for the reasoning.

---

## Command Structure

```bash
# Interactive wizard (asks for direction)
muxi config a2a

# Direct to inbound (skip direction question)
muxi config a2a --inbound

# Direct to outbound (skip direction question)
muxi config a2a --outbound

# Non-interactive mode (not yet implemented)
muxi config a2a --inbound --no-wizard
```

---

## Inbound Wizard Flow

### 1. Direction Selection (if no flag)
```
A2A Direction (↑↓ to select, Enter to confirm):
  ◉ Inbound - Accept connections from external agents
  ◯ Outbound - Connect to external agent services

✓ A2A Direction: Inbound
```

**Skipped if:** `--inbound` or `--outbound` flag is provided

---

### 2. Registry URLs
```
Registry URLs (comma or line-separated, must start with https://): 
https://private.company.com, https://registry.partner.com

✓ Registries: 2 added
```

**Accepts:**
- Comma-separated: `https://a.com, https://b.com`
- Line-separated (multi-line input)
- Single URL: `https://registry.com`

**Validates:**
- Must start with `https://` (no http://)
- Valid URL format (protocol, hostname, no trailing dots)
- At least one registry required

**Error examples:**
```
✗ invalid URL: http://registry.com (must start with https://)
✗ invalid URL: https://asdasd. (malformed hostname)
```

---

### 3. Authentication
```
Authentication (↑↓ to select, Enter to confirm):
  ◯ None (not recommended)
  ◉ API Key
  ◯ Bearer Token
  ◯ Basic Auth

✓ Authentication: API Key
```

**Options:**
1. **None** - No authentication (shows warning)
2. **API Key** - Custom header with API key
3. **Bearer Token** - Bearer authentication
4. **Basic Auth** - Username and password

---

### 4. Auth Details (based on type)

#### API Key
```
API Key header [X-API-Key]: X-A2A-Key
✓ API Key: X-A2A-Key
```

**Generated secret:** `A2A_INBOUND_API_KEY`

#### Bearer Token
```
Bearer Token: my-secret-token
✓ Bearer Token: my-secret-token
```

**Generated secret:** `A2A_INBOUND_BEARER_TOKEN`

#### Basic Auth
```
Username: admin
✓ Username: admin

Password: secret123
✓ Password: secret123
```

**Generated secrets:** `A2A_INBOUND_USERNAME`, `A2A_INBOUND_PASSWORD`

#### None
```
⚠ Warning: Inbound A2A without authentication is not recommended
```

**No secrets generated**

---

### 5. Trusted Endpoints (Optional)
```
Trusted endpoints (optional, comma or line-separated): 
muxi.partner.com, api.external.com

✓ Trusted endpoints: 2 added
```

**OR skip:**
```
Trusted endpoints (optional, comma or line-separated): ⏎
⊘ Trusted endpoints: skipped
```

**Accepts:**
- Comma-separated: `api1.com, api2.com`
- Line-separated (multi-line input)
- Domain names only (no protocol)

---

### 6. Final Output
```
✓ A2A inbound configuration added to formation.yaml
✓ Added 1 secret(s) to configure: A2A_INBOUND_API_KEY
```

---

## Generated Configuration

### Formation.yaml (New a2a Section)
```yaml
# Agent-to-Agent communication
a2a:
  enabled: true

  inbound:
    enabled: true
    port: 8181  # Hardcoded, no prompt
    registries:
      - "https://private.company.com"
      - "https://registry.partner.com"
    trusted_endpoints:
      - "muxi.partner.com"
      - "api.external.com"
    auth:
      type: "api_key"
      header: "X-A2A-Key"
      key: "${{ secrets.A2A_INBOUND_API_KEY }}"

  outbound:
    enabled: false
```

### Secrets File (Auto-Appended)
```bash
# Appended to secrets file
A2A_INBOUND_API_KEY=
```

**Or for bearer:**
```bash
A2A_INBOUND_BEARER_TOKEN=
```

**Or for basic:**
```bash
A2A_INBOUND_USERNAME=
A2A_INBOUND_PASSWORD=
```

---

## Auth Type Examples

### API Key
```yaml
auth:
  type: "api_key"
  header: "X-A2A-Key"
  key: "${{ secrets.A2A_INBOUND_API_KEY }}"
```

### Bearer Token
```yaml
auth:
  type: "bearer"
  token: "${{ secrets.A2A_INBOUND_BEARER_TOKEN }}"
```

### Basic Auth
```yaml
auth:
  type: "basic"
  username: "${{ secrets.A2A_INBOUND_USERNAME }}"
  password: "${{ secrets.A2A_INBOUND_PASSWORD }}"
```

### None
```yaml
# No auth section added
```

---

## Implementation Details

### Files
- **Command:** `src/cmd/config.go` - `configA2ACmd`
- **Logic:** `src/pkg/scaffold/components.go` - `ConfigureA2A()`, `configureInboundA2A()`
- **Helpers:** `parseURLList()`, `parseEndpointList()`, `updateFormationA2AInbound()`

### Key Functions

#### `ConfigureA2A(inbound, outbound, noWizard bool)`
- Entry point
- Determines direction (from flags or wizard)
- Routes to `configureInboundA2A()` or `configureOutboundA2A()`

#### `configureInboundA2A(rootDir string, noWizard bool)`
- Prompts for registries, auth, endpoints
- Validates inputs
- Calls `updateFormationA2AInbound()` to modify YAML
- Appends secrets to secrets file

#### `updateFormationA2AInbound(...)`
- Reads `formation.yaml`
- Checks if `a2a:` section exists
- Adds new section or errors if exists (for now)
- Writes back to file

---

## Current Limitations

### 1. Existing a2a Section
**Issue:** If `a2a:` section already exists in `formation.yaml`, the wizard errors:
```
Error: A2A section already exists in formation.yaml - manual editing required for now
```

**Future:** Implement YAML parsing/merging to update existing sections

---

### 2. No-Wizard Mode
**Issue:** `--no-wizard` flag not yet supported:
```bash
muxi config a2a --inbound --no-wizard
# Error: --no-wizard is not yet supported for A2A configuration
```

**Future:** Add non-interactive defaults

---

### 3. Outbound Not Implemented
**Issue:** Outbound wizard not yet built:
```bash
muxi config a2a --outbound
# Error: outbound A2A configuration is not yet implemented
```

**Status:** Next priority (see below)

---

### 4. Port Not Configurable
**Design:** Port is hardcoded to `8181` (standard A2A port)

**Rationale:** Users rarely need to change this. Can add `--port` flag if needed.

---

## Testing

### Manual Test
```bash
# Create test formation
cd /tmp
muxi new formation test-a2a --no-wizard
cd test-a2a

# Run inbound wizard
muxi config a2a --inbound
# Enter: https://registry.com
# Select: API Key
# Header: X-A2A-Key (or press Enter)
# Endpoints: (press Enter to skip)

# Verify output
cat formation.yaml  # Should have a2a section
cat secrets         # Should have A2A_INBOUND_API_KEY=
```

### Validation Test
```bash
# Invalid URL (no https)
echo "http://invalid.com" | muxi config a2a --inbound
# Expected: ✗ invalid URL: http://invalid.com (must start with https://)

# Invalid URL (malformed)
echo "https://bad..domain" | muxi config a2a --inbound
# Expected: ✗ invalid URL: https://bad..domain (malformed hostname)
```

---

## Next Steps

### 1. Outbound Wizard (Priority: HIGH)
**Similar flow to inbound:**
- Registries (https:// URLs)
- Auth (same options)
- Service-specific auth (optional)
- Default retry/timeout settings

**Generated config:**
```yaml
a2a:
  outbound:
    enabled: true
    registries:
      - "https://a2a.muxihub.com"
    default_retry_attempts: 3
    default_timeout_seconds: 30
    services:
      - service_id: "partner@api.external.com"
        auth:
          type: "bearer"
          token: "${{ secrets.PARTNER_BEARER_TOKEN }}"
```

---

### 2. YAML Merging (Priority: MEDIUM)
**Problem:** Can't update existing a2a sections

**Solution:**
- Parse YAML using `gopkg.in/yaml.v3`
- Merge new values with existing
- Preserve comments and formatting
- Handle both adding and updating

**Example:**
```yaml
# Before
a2a:
  enabled: true
  outbound:
    enabled: true

# After (add inbound)
a2a:
  enabled: true
  inbound:
    enabled: true
    # ... new config
  outbound:
    enabled: true  # preserved
```

---

### 3. No-Wizard Mode (Priority: LOW)
**Support:**
```bash
muxi config a2a --inbound \
  --registries="https://r1.com,https://r2.com" \
  --auth-type=bearer \
  --auth-token="$MY_TOKEN" \
  --endpoints="api.partner.com"
```

---

### 4. Additional Config Commands (Future)
Following the `muxi config` pattern:

```bash
muxi config llm             # Configure LLM providers
muxi config logging         # Configure logging streams
muxi config scheduler       # Configure scheduler
muxi config filtering       # Configure A2A filtering
muxi config webhooks        # Configure webhooks
```

---

## Related Documentation

- **Command Semantics:** [COMMAND-SEMANTICS.md](./COMMAND-SEMANTICS.md) - Why `config` vs `new`
- **Schema Reference:** `/Users/ran/Projects/muxi/code/schemas/formation/README.md`
- **TUI Design:** [TUI-DESIGN.md](./TUI-DESIGN.md) - Prompts, colors, symbols
- **Input History:** [INPUT-HISTORY.md](./INPUT-HISTORY.md) - Arrow keys for history
