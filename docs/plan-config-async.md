# Plan: `muxi config async`

Configure async response settings for long-running tasks.

## Current State

In `muxi new formation`, async is configured with:
- Enable/disable toggle
- Webhook URL (required when enabled)

Generated YAML:
```yaml
async:
  threshold_seconds: 30
  webhook_url: "https://example.com/webhook"
```

## Proposed Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `enabled` | Enable async responses | false |
| `threshold_seconds` | Time before switching to async | 30 |
| `webhook_url` | URL for async response delivery | required |
| `retry_count` | Number of delivery retries | 3 |
| `retry_delay_seconds` | Delay between retries | 5 |

## Wizard Flow

```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Async Responses                           MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure how long-running tasks are handled. When a task    │
│ exceeds the threshold, the response is delivered via webhook.│
╰──────────────────────────────────────────────────────────────╯

What would you like to do?
  ◯ Configure async settings
  ◯ Disable async responses
```

### Flow 1: Configure Async Settings

```
Async Response Configuration

  Enable async responses for long-running tasks
  ✓ Enabled: yes

  Time threshold before switching to async mode
  ✓ Threshold: 30 seconds

  Webhook URL for async response delivery
  ✓ Webhook URL: https://api.example.com/muxi/callback

  Number of delivery retry attempts
  ✓ Retry count: 3

  Delay between retry attempts
  ✓ Retry delay: 5 seconds

✓ Async configuration saved to formation.yaml
```

### Flow 2: Disable Async

```
Disable Async Responses

  ⚠ This will remove the async configuration from formation.yaml

  Disable async responses? (y/N): y

✓ Async responses disabled
```

## Output YAML

```yaml
async:
  threshold_seconds: 30
  webhook_url: "https://api.example.com/muxi/callback"
  retry:
    count: 3
    delay_seconds: 5
```

## Validation

- Webhook URL required when enabled
- URL validation (auto-add https://, validate host)
- Threshold must be positive integer (1-3600)
- Retry count must be 0-10
- Retry delay must be 1-60

## Current Value Detection

Show `[current]` for existing values:
```
  Threshold: 30 seconds [current]
    ◯ 15 seconds
    ◯ 30 seconds [current]
    ◯ 60 seconds
    ◯ 120 seconds
    ◯ Custom
```

## Implementation Notes

1. Check if `async:` section exists in formation.yaml
2. If exists, show current values with `[current]` markers
3. Use URL validation helper from security.go
4. Apply YAML formatting (2-space indent, blank lines)
5. Remove commented async section if present

## Files to Create/Modify

- `src/pkg/scaffold/async.go` - New wizard (~300 lines)
- `src/cmd/config.go` - Add `configAsyncCmd`
- `docs/guides/async.md` - User guide
