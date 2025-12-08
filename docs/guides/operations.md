# Operations Commands

Runtime operations for deployed formations: firing triggers, managing async jobs, viewing audit logs, and streaming live logs.

---

## Overview

Operations commands let you interact with running formations:

| Command | Purpose | Auth |
|---------|---------|------|
| `muxi trigger` | Fire a trigger with data | Client Key |
| `muxi jobs` | List/cancel async jobs | Client Key + User ID |
| `muxi audit` | View/clear audit log | Admin Key |
| `muxi stream` | Live log streaming | Admin Key |

---

## Fire Trigger

Fire a trigger on a deployed formation with optional JSON data.

```bash
# Fire trigger with inline JSON
muxi trigger github-issue --data '{"issue": {"number": 123, "title": "Bug"}}'

# Fire trigger with JSON file
muxi trigger webhook --file event.json

# Fire asynchronously (returns job ID)
muxi trigger daily-report --async

# From anywhere (specify formation)
muxi trigger my-trigger -F my-formation -p production
```

**Output (sync):**
```
  ✓ Trigger fired

  Status:     completed
  Request ID: req_abc123

  Response:
  Issue #123 has been categorized as a bug report...
```

**Output (async):**
```
  ✓ Trigger fired

  Status:     async
  Job ID:     job_xyz789

  Track progress: muxi jobs -u <user>
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.yaml) |
| `--profile` | `-p` | Server profile |
| `--data` | | JSON string to send with trigger |
| `--file` | | Path to JSON file |
| `--async` | | Fire asynchronously (returns job ID) |

**Notes:**
- Use `--data` OR `--file`, not both
- If neither provided, sends empty object `{}`
- Async mode is useful for long-running triggers

---

## Manage Async Jobs

List and manage asynchronous jobs for a user.

### List Jobs

```bash
# List jobs for default user
muxi jobs

# List jobs for specific user
muxi jobs -u alice

# From anywhere
muxi jobs -F my-formation -u alice
```

**Output:**
```
  JOB ID         STATUS       PROGRESS   CREATED
  ──────────────  ────────────  ──────────  ────────────
  job_456789      processing   75%        2 minutes ago
  job_123456      completed    100%       1 hour ago
  job_abc123      failed       -          3 hours ago
```

### Cancel Job

```bash
# Cancel a running job
muxi jobs cancel job_456789 -u alice
```

**Output:**
```
  ✓ Job cancelled

✓ Cancelled job job_456789
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (required) |

**Job Statuses:**
- `pending` - Queued, not yet started
- `processing` - Currently running
- `completed` - Finished successfully
- `failed` - Finished with error
- `cancelled` - Cancelled by user

---

## View Audit Log

View administrative actions logged by the formation. Requires admin key.

```bash
# View recent audit entries (default: 50)
muxi audit

# View more entries
muxi audit --lines 100

# Clear audit log (with confirmation)
muxi audit --clear

# From anywhere
muxi audit -F my-formation -p production
```

**Output:**
```
  TIMESTAMP             ACTION              RESOURCE            USER
  ────────────────────  ──────────────────  ──────────────────  ──────────
  2025-12-08 10:15:00   agent.created       weather-bot         admin
  2025-12-08 10:14:30   config.updated      llm.temperature     admin
  2025-12-08 10:10:00   secret.added        OPENAI_API_KEY      system
```

**Clear audit log:**
```bash
muxi audit --clear
```
```
  Clear the audit log?

  This action cannot be undone!
  Type 'clear' to confirm: clear

  ✓ Audit log cleared
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--lines` | `-n` | Number of entries to show (default: 50) |
| `--clear` | | Clear the audit log |

---

## Stream Logs

Stream real-time logs from a formation via Server-Sent Events (SSE). Requires admin key.

```bash
# Stream all logs
muxi stream

# Filter by log level
muxi stream --level error

# Filter by agent
muxi stream --agent weather-bot

# Filter by user
muxi stream -u alice

# Filter by request ID
muxi stream --request req_abc123

# Combine filters
muxi stream --level warn --agent code-reviewer

# From anywhere
muxi stream -F my-formation -p production
```

**Output:**
```
  Streaming logs (Ctrl+C to stop)

  [10:15:00] INFO   chat.started user=alice session=sess_abc
  [10:15:01] INFO   agent.invoked agent=weather-assistant
  [10:15:02] DEBUG  llm.request model=gpt-4 tokens=150
  [10:15:03] INFO   chat.completed user=alice duration=2.5s
  [10:15:10] WARN   rate.limit user=bob remaining=5
  [10:15:15] ERROR  mcp.timeout server=database-tools
```

Press `Ctrl+C` to stop streaming.

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | Filter by user ID |
| `--level` | | Filter by level (debug, info, warn, error) |
| `--agent` | | Filter by agent ID |
| `--request` | | Filter by request ID |

**Log Levels:**
- `DEBUG` - Detailed debugging information
- `INFO` - General operational events
- `WARN` - Warning conditions
- `ERROR` - Error conditions

---

## API Reference

| Command | API Endpoint | Method | Auth |
|---------|--------------|--------|------|
| `muxi trigger <name>` | `/triggers/{name}` | POST | Client Key |
| `muxi jobs` | `/jobs/{user_id}` | GET | Client Key + User ID |
| `muxi jobs cancel` | `/jobs/{user_id}/{job_id}` | DELETE | Client Key + User ID |
| `muxi audit` | `/audit` | GET | Admin Key |
| `muxi audit --clear` | `/audit` | DELETE | Admin Key |
| `muxi stream` | `/logs/stream` | GET (SSE) | Admin Key |

---

## Examples

### Workflow: Fire Trigger and Track Job

```bash
# 1. Fire trigger asynchronously
muxi trigger process-batch --file batch.json --async
# Output: Job ID: job_abc123

# 2. Check job status
muxi jobs -u system

# 3. Monitor logs while job runs
muxi stream --request req_xyz
```

### Debugging: Investigate Errors

```bash
# 1. Stream only errors
muxi stream --level error

# 2. Check audit log for recent changes
muxi audit --lines 20

# 3. Filter logs by specific agent
muxi stream --agent problematic-agent --level debug
```

### Monitoring: Watch User Activity

```bash
# Stream logs for specific user
muxi stream -u alice

# Check their async jobs
muxi jobs -u alice
```

---

## Related Commands

- `muxi triggers` - List available triggers (Track A)
- `muxi sessions` - Manage user sessions
- `muxi info` - Formation status and health
