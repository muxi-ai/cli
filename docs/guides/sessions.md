# Session Management

Manage user sessions and message history in deployed formations.

---

## Overview

Sessions track conversations between users and formations. Each session contains a sequence of messages (user prompts and agent responses) and maintains state across interactions.

**Commands:**
- `muxi sessions` - List sessions for a user
- `muxi history` - View message history for a session
- `muxi clear` - Delete a session

**Authentication:** All session commands require Client Key + User ID.

---

## Prerequisites

1. **Running formation** - Deploy with `muxi deploy`
2. **Client key** - Set `MUXI_CLIENT_KEY` env var or in `secrets.enc`
3. **User ID** - Via `-u` flag, `.muxi` file, or `muxi set default user`

---

## Commands

### List Sessions

```bash
# List all sessions for default user
muxi sessions

# List sessions for specific user
muxi sessions -u alice

# Show only active sessions
muxi sessions --active

# From anywhere (specify formation)
muxi sessions -F my-formation -u alice
```

**Output:**
```
  SESSION ID           MESSAGES   LAST ACTIVITY      STATUS
  sess_abc123          25         2 minutes ago      ● active
  sess_xyz789          12         3 hours ago        ○ inactive
  sess_def456          8          1 day ago          ○ inactive

  3 session(s) for user alice
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.yaml) |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (required) |
| `--active` | | Show only active sessions |

---

### View History

```bash
# View all messages in session
muxi history -s sess_abc123 -u alice

# Limit to last 50 messages
muxi history -s sess_abc123 -u alice --lines 50

# Output as JSON
muxi history -s sess_abc123 -u alice --json
```

**Output:**
```
  Session: sess_abc123 (25 messages)

  [10:00:15] user: What's the weather like?
  [10:00:18] weather-assistant: The weather today is sunny with a high of 72°F.

  [10:05:00] user: Can you help me with Python code?
  [10:05:03] code-helper: Of course! What do you need help with?
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--session` | `-s` | Session ID (required) |
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (required) |
| `--lines` | `-n` | Limit number of messages (0 = all) |
| `--json` | | Output as JSON |

---

### Clear Session

```bash
# Delete session (with confirmation)
muxi clear -s sess_abc123 -u alice

# Skip confirmation
muxi clear -s sess_abc123 -u alice -f
```

**Output:**
```
  Clear session 'sess_abc123'? This will delete all messages. [y/N]: y

  ✓ Cleared session

✓ Cleared session sess_abc123
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--session` | `-s` | Session ID (required) |
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (required) |
| `--force` | `-f` | Skip confirmation prompt |

---

## User ID Resolution

Session commands require a user ID. Resolution order:

1. **Flag:** `-u alice` (highest priority)
2. **Formation config:** `.muxi` file in formation directory
3. **Global default:** `~/.muxi/cli/defaults.yaml`

**Set default user:**
```bash
# Global default
muxi set default user alice

# Formation-specific (in formation dir)
muxi set default user dev-user --local
```

If no user ID is found, commands will error with:
```
user ID required - use -u flag or set default with: muxi set default user
```

---

## API Reference

| Command | API Endpoint | Method |
|---------|--------------|--------|
| `muxi sessions` | `/sessions` | GET |
| `muxi history` | `/sessions/{id}/messages` | GET |
| `muxi clear` | `/sessions/{id}` | DELETE |

All endpoints require:
- `X-MUXI-CLIENT-KEY` header
- `X-User-ID` header

---

## Examples

### Workflow: Review and Clean Up Sessions

```bash
# 1. Check user's sessions
muxi sessions -u alice

# 2. Review a specific session
muxi history -s sess_abc123 -u alice --lines 20

# 3. Export session to file
muxi history -s sess_abc123 -u alice --json > session-backup.json

# 4. Clear old session
muxi clear -s sess_old -u alice -f
```

### Debugging: Find Active Conversations

```bash
# Show only active sessions
muxi sessions -u webhook-bot --active

# Check what's happening in active session
muxi history -s sess_xyz -u webhook-bot --lines 10
```

---

## Related Commands

- `muxi memory` - Manage persistent user memories
- `muxi jobs` - View async jobs for a user
- `muxi stream` - Live log streaming (filter by user)
