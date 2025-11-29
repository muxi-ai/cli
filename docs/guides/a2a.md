# A2A (Agent-to-Agent) Guide

This guide covers A2A communication in MUXI formations - both exposing your agents to others and connecting to external agent services.

## What is A2A?

A2A (Agent-to-Agent) protocol enables formations to communicate with each other. Your formation can:
- **Expose agents** (inbound) - Let other formations call your agents
- **Call external agents** (outbound) - Use agents from other formations

## Two Directions

| Direction | Description | Command |
|-----------|-------------|---------|
| **Inbound** | Other formations call YOUR agents | `muxi config a2a --inbound` |
| **Outbound** | Your formation calls EXTERNAL agents | `muxi new a2a-service` |

## Inbound A2A

### Enabling Inbound A2A

```bash
muxi config a2a --inbound
```

The wizard guides you through:
1. **Registry URLs** - Where to publish your agent catalog
2. **Authentication** - How other formations authenticate

### Configuration in formation.yaml

```yaml
a2a:
  inbound:
    enabled: true
    registries:
      - https://registry.muxi.cloud
      - https://internal.company.com/registry
    auth:
      type: api_key
      header: X-A2A-Key
      key: ${{ secrets.A2A_INBOUND_API_KEY }}
```

### Authentication Options

**API Key:**
```yaml
auth:
  type: api_key
  header: X-A2A-Key
  key: ${{ secrets.A2A_INBOUND_API_KEY }}
```

**Bearer Token:**
```yaml
auth:
  type: bearer
  token: ${{ secrets.A2A_INBOUND_BEARER_TOKEN }}
```

**Basic Auth:**
```yaml
auth:
  type: basic
  username: ${{ secrets.A2A_INBOUND_USERNAME }}
  password: ${{ secrets.A2A_INBOUND_PASSWORD }}
```

### Making Agents Visible

Enable A2A visibility in agent config:

```yaml
# agents/research-assistant.yaml
a2a:
  visible: true
  description: "Research and information gathering specialist"
```

Only agents with `visible: true` appear in your formation's A2A catalog.

## Outbound A2A (External Services)

### Creating an A2A Service

```bash
muxi new a2a-service
```

The wizard guides you through:
1. **Service ID** - Unique identifier (e.g., `billing-service`)
2. **Display Name** - Human-friendly name
3. **Description** - What this service provides
4. **Service URL** - The external formation's A2A endpoint
5. **Authentication** - How to authenticate with the service

### File Structure

```
my-formation/
└── a2a/
    ├── billing-service.yaml
    ├── notification-hub.yaml
    └── analytics-engine.yaml
```

### A2A Service Configuration

```yaml
schema: "1.0.0"
id: "billing-service"
name: "Billing Service"
description: "External billing and invoicing agents"
url: "https://billing.partner.com/a2a"
active: true

auth:
  type: api_key
  header: X-API-Key
  key: ${{ secrets.A2A_SERVICE_BILLING_SERVICE_API_KEY }}

# Optional rate limiting
retry_attempts: 3
timeout_seconds: 30
```

### Authentication Types

**API Key:**
```yaml
auth:
  type: api_key
  header: X-API-Key
  key: ${{ secrets.A2A_SERVICE_BILLING_API_KEY }}
```

**Bearer Token:**
```yaml
auth:
  type: bearer
  token: ${{ secrets.A2A_SERVICE_BILLING_TOKEN }}
```

**Basic Auth:**
```yaml
auth:
  type: basic
  username: ${{ secrets.A2A_SERVICE_BILLING_USERNAME }}
  password: ${{ secrets.A2A_SERVICE_BILLING_PASSWORD }}
```

**Custom Headers:**
```yaml
auth:
  type: custom
  headers:
    Authorization: "Custom ${{ secrets.CUSTOM_TOKEN }}"
    X-Client-ID: ${{ secrets.CLIENT_ID }}
    X-Tenant-ID: "tenant-123"
```

## Enabling Outbound A2A

After creating services, enable outbound in formation.yaml:

```bash
muxi config a2a --outbound
```

Or manually add:

```yaml
a2a:
  outbound:
    enabled: true
    services:
      - billing-service
      - notification-hub
```

## Complete A2A Setup

A formation with both inbound and outbound:

```yaml
# formation.yaml
a2a:
  inbound:
    enabled: true
    registries:
      - https://registry.muxi.cloud
    auth:
      type: api_key
      header: X-A2A-Key
      key: ${{ secrets.A2A_INBOUND_API_KEY }}
  
  outbound:
    enabled: true
    services:
      - billing-service
      - analytics-engine
```

```yaml
# a2a/billing-service.yaml
schema: "1.0.0"
id: "billing-service"
name: "Billing Service"
url: "https://billing.partner.com/a2a"
auth:
  type: bearer
  token: ${{ secrets.A2A_SERVICE_BILLING_SERVICE_BEARER_TOKEN }}
```

```yaml
# agents/support-agent.yaml
a2a:
  visible: true
  description: "Customer support specialist"
```

## Using A2A in Agents

Agents can call external services in their prompts:

```markdown
When handling billing questions:
1. Use @billing-service to look up invoice details
2. Use @billing-service to check payment status
```

Or in SOPs:

```markdown
## Step 3: Process Payment

**Agent:** @finance-bot
**A2A Services:** @billing-service

1. Verify payment details
2. Call billing service to process
3. Confirm transaction
```

## Registry Discovery

When outbound is enabled, agents can discover available services:

```yaml
# formation.yaml
a2a:
  outbound:
    enabled: true
    registries:
      - https://registry.muxi.cloud
    services:
      - billing-service  # Explicitly configured
      # Plus any discovered from registries
```

## Secrets Management

A2A credentials are stored encrypted:

```bash
# Inbound auth
muxi secrets set A2A_INBOUND_API_KEY

# Service auth (auto-named by wizard)
muxi secrets set A2A_SERVICE_BILLING_SERVICE_API_KEY

# List all secrets
muxi secrets list
```

## Best Practices

1. **Use specific service names** - `billing-service` not `service1`
2. **Document visible agents** - Clear descriptions help callers
3. **Secure all endpoints** - Always require authentication
4. **Limit agent visibility** - Only expose what's needed
5. **Set timeouts** - Prevent hanging on slow services
6. **Monitor A2A calls** - Track usage and errors
7. **Version your agents** - Breaking changes affect callers
8. **Test connectivity** - Verify services before deploying
