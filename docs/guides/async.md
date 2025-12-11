# Async Response Configuration Guide

Configure async response settings using `muxi config async`.

## Quick Start

```bash
cd my-formation
muxi config async
```

## What is Async Mode?

When a task takes longer than the configured threshold, the formation switches to async mode:

1. Returns immediately with a task ID
2. Continues processing in the background
3. Delivers the result via webhook when complete

This prevents HTTP timeouts for long-running tasks.

---

## Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `threshold_seconds` | Time before switching to async | 30 |
| `webhook_url` | URL for async response delivery | required |
| `retry.count` | Number of delivery retries | 3 |
| `retry.delay_seconds` | Delay between retries | 5 |

---

## Example Configuration

```yaml
async:
  threshold_seconds: 30
  webhook_url: "https://api.example.com/muxi/callback"
  retry:
    count: 3
    delay_seconds: 5
```

---

## Webhook Payload

When a task completes, the webhook receives:

```json
{
  "task_id": "task_abc123",
  "status": "completed",
  "result": {
    "message": "...",
    "metadata": {}
  },
  "completed_at": "2025-12-01T12:00:00Z"
}
```

---

## Enabling Async

```bash
muxi config async
# Select: Enable and configure async responses
# Set threshold, webhook URL, retry settings
```

---

## Disabling Async

```bash
muxi config async
# Select: Disable async responses
```

This removes the `async:` section from your formation config.

---

## Threshold Guidelines

| Use Case | Recommended Threshold |
|----------|----------------------|
| Simple queries | 15 seconds |
| Standard tasks | 30 seconds |
| Complex workflows | 60 seconds |
| Heavy processing | 120 seconds |

---

## Retry Configuration

**count**: Number of retry attempts (0-10)
- Set to 0 to disable retries
- Recommended: 3 for most cases

**delay_seconds**: Wait time between retries (1-60)
- Recommended: 5 seconds for most cases
- Increase for rate-limited endpoints

---

## Related

- [Formations Guide](formations.md) - Creating formations
- [Logging Guide](logging.md) - Configure logging
