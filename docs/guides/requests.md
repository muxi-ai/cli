# Requests Guide

Manage and track chat requests.

## Overview

Requests represent individual chat messages sent to a formation. Use this to track async requests, view status, and cancel in-progress requests.

## List Requests

List all requests for a user:

```bash
muxi requests list
muxi requests list -u alice
```

Output:
```
Requests for user 'alice' (5)

ID                    STATUS       STARTED              DURATION
req_abc123            completed    Dec 24, 14:30:00    1.2s
req_def456            completed    Dec 24, 14:25:00    0.8s
req_ghi789            processing   Dec 24, 14:35:00    -
req_jkl012            failed       Dec 24, 14:20:00    2.5s
req_mno345            cancelled    Dec 24, 14:15:00    0.5s
```

## Show Request Details

View details of a specific request:

```bash
muxi requests show req_abc123
muxi requests show req_abc123 -u alice
```

Output:
```
Request Details

ID:          req_abc123
Status:      completed
User:        alice
Session:     sess_xyz789
Started:     Dec 24, 2024 14:30:00 UTC
Duration:    1.2s
Agent:       weather-bot
Tokens:      [125, 89, 36, 0, 0, 0]
```

## Cancel Request

Cancel an in-progress request:

```bash
muxi requests cancel req_ghi789
muxi requests cancel req_ghi789 -u alice
```

Only requests with status "processing" can be cancelled.

## Request Statuses

| Status | Description |
|--------|-------------|
| `pending` | Request queued, not yet processing |
| `processing` | Currently being processed |
| `completed` | Successfully completed |
| `failed` | Failed with an error |
| `cancelled` | Cancelled by user |

## Options

```
-f, --formation    Formation ID
-p, --profile      Server profile
-u, --user         User ID
```

## Use Cases

### Track Async Requests

When using async mode, track request progress:

```bash
# Send async request
muxi triggers run my-trigger --async

# Track progress
muxi requests show req_abc123
```

### Cancel Long-Running Requests

If a request is taking too long:

```bash
muxi requests list -u alice
muxi requests cancel req_ghi789
```

### Debug Failed Requests

Check why a request failed:

```bash
muxi requests show req_jkl012
muxi logs --request req_jkl012
```

## Related Commands

- `muxi logs --request <id>` - Stream logs for a specific request
- `muxi sessions` - View and manage sessions
- `muxi triggers run --async` - Run triggers asynchronously
