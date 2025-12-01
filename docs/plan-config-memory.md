# Plan: `muxi config memory`

## Overview
Configure memory settings in formation.yaml. First asks which type, then flows into specific configuration.

## Command
```bash
muxi config memory
```

## Banner
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Memory                                    MUXI │
│──────────────────────────────────────────────────────────────│
│ ℹ Memory enables agents to remember context and information. │
│ Working memory is always on. Persistent memory needs a DB.   │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: Memory Type Selection
```
What would you like to configure?
  ◯ Working memory (in-memory, always enabled)
  ◯ Buffer memory (conversation context)
  ◯ Persistent memory (database-backed, long-term)
```

---

### Flow 1: Working Memory

```
Working Memory
  
  Short-term vector memory for semantic search within a session.

  Select mode:
    ◯ Local (default, in-process)
    ◯ Remote (FAISSx server)
```

#### If Local:
```
  "auto" uses 10% of RAM (min 64MB, max 1GB)
  ✓ Max memory: auto
  
  Vector dimension from embedding model: openai/text-embedding-3-large
  ✓ Vector dimension: 3072
  
  How often to clean up old vectors (FIFO)
  ✓ Cleanup interval (minutes): 5
```

Note: Vector dimension is auto-detected from the configured embedding model.

#### If Remote:
```
  FAISSx server endpoint (tcp:// or tcps://)
  ✓ Server URL: tcp://localhost:8000

  API key for FAISSx authentication
  ✓ Saved FAISSX_API_KEY to secrets
  
  Tenant identifier for multi-tenant FAISSx
  ✓ Saved FAISSX_TENANT_ID to secrets

  Remote mode requires explicit memory limit
  ✓ Max memory (MB): 256
```

**Output (Local):**
```yaml
memory:
  working:
    mode: "local"
    max_memory_mb: "auto"
    vector_dimension: 1536
    fifo_interval_min: 5
```

**Output (Remote):**
```yaml
memory:
  working:
    mode: "remote"
    max_memory_mb: 256
    vector_dimension: 1536
    fifo_interval_min: 5
    remote:
      url: "tcp://localhost:8000"
      api_key: "${{ secrets.FAISSX_API_KEY }}"
      tenant: "${{ secrets.FAISSX_TENANT_ID }}"
```

---

### Flow 2: Buffer Memory

```
Buffer Memory

  Conversation context that persists across messages in a session.

  Number of recent messages to keep in context
  ✓ Context window size: 10

  Multiplier for total buffer (total = size × multiplier)
  ✓ Buffer multiplier: 10

  Use vector similarity to find relevant past messages
  Enable vector similarity search? [Y/n]: y
  ✓ Vector search: enabled
```

**Output:**
```yaml
memory:
  buffer:
    size: 10
    multiplier: 10
    vector_search: true
```

---

### Flow 3: Persistent Memory

```
Persistent Memory

  Long-term memory stored in a database, persists across sessions.

  Select database type:
    ◯ PostgreSQL (requires PostgreSQL 17+ with pgvector extension)
    ◯ SQLite (local file, good for development)
```

#### If PostgreSQL:
```
  Requires PostgreSQL 17+ with pgvector extension installed.
  
  Enter connection string or hostname:
  > postgres://user:pass@host:5432/db
  
  ✓ Saved PERSISTENT_DB_CONNECTION_STRING to secrets
```

**Or if hostname entered:**
```
  Enter connection string or hostname:
  > db.example.com

  ✓ Host: db.example.com
  
  PostgreSQL port (default: 5432)
  ✓ Port: 5432
  
  Database name
  ✓ Database: muxi_memory
  
  Database username
  ✓ Username: muxi
  
  Database password
  ✓ Password: ********

  Built connection string: postgres://muxi:***@db.example.com:5432/muxi_memory
  ✓ Saved PERSISTENT_DB_CONNECTION_STRING to secrets
```

#### If SQLite:
```
  Path to SQLite database file (will be created if doesn't exist)
  ✓ Database path: ./memory.db
```

#### Common settings (PostgreSQL only):
```
  Maximum time to wait for database queries
  ✓ Query timeout (seconds): 30

  Generate LLM-synthesized summaries of user context
  Enable user synopsis? [Y/n]: y
  ✓ User synopsis: enabled

  How long to cache generated synopses
  ✓ Synopsis cache TTL (seconds): 3600
```

**Output (PostgreSQL):**
```yaml
memory:
  persistent:
    connection_string: "${{ secrets.PERSISTENT_DB_CONNECTION_STRING }}"
    query_timeout_seconds: 30
    user_synopsis:
      enabled: true
      cache_ttl: 3600
```

**Output (SQLite):**
```yaml
memory:
  persistent:
    connection_string: "sqlite:///./memory.db"
```

---

## Secrets Created

| Secret | When Created |
|--------|--------------|
| `FAISSX_API_KEY` | Remote working memory selected |
| `FAISSX_TENANT_ID` | Remote working memory selected |
| `PERSISTENT_DB_CONNECTION_STRING` | PostgreSQL persistent memory selected |

Note: SQLite doesn't use secrets (just a file path).

## Environment Variable Detection

Check for existing env vars and offer to import:
- `FAISSX_API_KEY` → for remote working memory
- `DATABASE_URL` → for PostgreSQL connection string

## Validation

- PostgreSQL connection string format (`postgres://` or `postgresql://`)
- SQLite path is valid and parent directory exists
- Remote URL format (`tcp://` or `tcps://`)
- Vector dimension is positive integer
- Memory limits are positive integers

## Implementation Notes

1. Use radio boxes (◯) for selections, not numbered lists
2. Add dimmed hints above each prompt explaining the setting
3. Indent content under section headers by 2 spaces
4. Store sensitive values in secrets, reference with `${{ secrets.NAME }}`
5. Detect env vars and offer to import (like LLM config)
6. For PostgreSQL, accept either full connection string OR hostname (then prompt for details)
