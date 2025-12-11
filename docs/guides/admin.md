# Advanced Admin Commands

Manage scheduler jobs, user identifiers, and memories in deployed formations.

---

## Overview

These commands provide administrative control over advanced formation features:

- **Scheduler** - Manage automated scheduled jobs
- **Users** - Map external identifiers to MUXI user IDs
- **Memory** - View memory config and manage user memories

---

## Scheduler

View scheduler configuration and manage scheduled jobs.

**Authentication:** Admin Key

### View Scheduler Config

```bash
# Show scheduler configuration
muxi scheduler

# From anywhere
muxi scheduler -F my-formation
```

**Output:**
```
Scheduler Configuration

  Status:              enabled
  Timezone:            UTC
  Check Interval:      5 minutes
  Max Concurrent Jobs: 10
  Max Failures:        3
```

### List Scheduled Jobs

```bash
# List all scheduled jobs
muxi scheduler list

# From anywhere
muxi scheduler list -F my-formation -p production
```

**Output:**
```
  ID                   TYPE         SCHEDULE        NEXT RUN              STATUS
  ────────────────────────────────────────────────────────────────────────────────
  daily-report         recurring    0 9 * * *       Tomorrow 9:00am       ● enabled
  weekly-cleanup       recurring    0 0 * * 0       Sunday 12:00am        ● enabled
  one-time-deploy      one_time     -               Dec 10 2:00pm         ● pending
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.afs) |
| `--profile` | `-p` | Server profile |

---

## Users

Manage user identifier mappings. Identifiers map external IDs (email, phone, etc.) to MUXI user IDs.

**Authentication:** Client Key

### List User Identifiers

```bash
# List identifiers for a user
muxi users identifiers -u alice

# From anywhere
muxi users identifiers -F my-formation -u alice
```

**Output:**
```
  Identifiers for user 'alice':

  IDENTIFIER                           TYPE            CREATED
  ──────────────────────────────────────────────────────────────────────
  alice@company.com                    email           Dec 1, 2025
  +1-555-123-4567                      phone           Dec 5, 2025
  slack:U12345                         external_id     Dec 6, 2025
```

### Link Identifier

```bash
# Link email to user
muxi users link -u alice "alice@company.com"

# Link with explicit type
muxi users link -u alice "+1-555-123-4567" --type phone

# Link external ID
muxi users link -u alice "slack:U12345" --type external_id
```

**Output:**
```
✓ Linked 'alice@company.com' to user 'alice'
```

### Unlink Identifier

```bash
# Remove identifier mapping
muxi users unlink "alice@company.com"
```

**Output:**
```
✓ Unlinked 'alice@company.com'
```

### Resolve Identifier

```bash
# Look up which user an identifier maps to
muxi users resolve "alice@company.com"
```

**Output:**
```
  Identifier: alice@company.com
  User ID:    alice
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (for identifiers/link) |
| `--type` | | Identifier type (for link) |

---

## Memory

View memory configuration and manage user memories.

**Authentication:** Admin Key (config), Client Key + User ID (memories)

### View Memory Config

```bash
# Show memory configuration
muxi memory

# Or explicitly
muxi memory status
```

**Output:**
```
Memory Configuration

  Buffer:
    Size:          50 messages
    Multiplier:    1.5
    Vector Search: enabled

  Working Memory:
    Max Size:      256 MB
    FIFO Interval: 30 minutes
```

### List User Memories

```bash
# List memories for a user
muxi memory list -u alice

# From anywhere
muxi memory list -F my-formation -u alice
```

**Output:**
```
  Memories for user 'alice' (3):

  ID                   CONTENT                                       CREATED
  ────────────────────────────────────────────────────────────────────────────────
  mem_abc123           Prefers dark mode                             Dec 1, 2025
  mem_def456           Uses TypeScript for all projects              Dec 3, 2025
  mem_ghi789           Timezone: PST, prefers morning meetings       Dec 5, 2025
```

### Add Memory

```bash
# Add a memory for a user
muxi memory add -u alice "Prefers dark mode"

# Add longer memory
muxi memory add -u alice "Expert in Python and Go, prefers functional programming patterns"
```

**Output:**
```
✓ Added memory 'mem_xyz123' for user 'alice'
```

### Delete Memory

```bash
# Delete a specific memory
muxi memory delete -u alice mem_abc123
```

**Output:**
```
✓ Deleted memory 'mem_abc123'
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |
| `--user` | `-u` | User ID (for list/add/delete) |

---

## API Reference

| Command | API Endpoint | Method | Auth |
|---------|--------------|--------|------|
| `muxi scheduler` | `/scheduler` | GET | Admin |
| `muxi scheduler list` | `/scheduler/jobs` | GET | Admin |
| `muxi users identifiers` | `/users/{id}/identifiers` | GET | Client |
| `muxi users link` | `/users/identifiers` | POST | Admin |
| `muxi users unlink` | `/users/identifiers/{id}` | DELETE | Admin |
| `muxi users resolve` | `/users/identifiers/{id}` | GET | Admin |
| `muxi memory` | `/memory` | GET | Admin |
| `muxi memory list` | `/memories` | GET | Client+User |
| `muxi memory add` | `/memories` | POST | Client+User |
| `muxi memory delete` | `/memories/{id}` | DELETE | Client+User |

---

## Examples

### User Onboarding: Set Up Identifiers

```bash
# Link all known identifiers for new user
muxi users link -u alice "alice@company.com" --type email
muxi users link -u alice "+1-555-123-4567" --type phone
muxi users link -u alice "slack:U12345" --type external_id

# Verify
muxi users identifiers -u alice
```

### Memory Management: Track User Preferences

```bash
# Add user preferences as memories
muxi memory add -u alice "Prefers concise responses"
muxi memory add -u alice "Expert in Python, familiar with Go"
muxi memory add -u alice "Working on e-commerce project 'ShopApp'"

# Review current memories
muxi memory list -u alice

# Clean up outdated memory
muxi memory delete -u alice mem_old123
```

### Scheduler Monitoring

```bash
# Check scheduler health
muxi scheduler

# Review scheduled jobs
muxi scheduler list

# Check specific formation
muxi scheduler list -F production-bot -p prod
```

---

## Related Commands

- `muxi sessions` - Manage user sessions
- `muxi history` - View session messages
- `muxi audit` - View formation audit log
- `muxi stream` - Live log streaming
