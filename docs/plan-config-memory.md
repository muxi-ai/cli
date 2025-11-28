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
  [1] Working memory (in-memory, always enabled)
  [2] Buffer memory (conversation context)
  [3] Persistent memory (database-backed, long-term)

Select (1-3): _
```

---

### Flow 1: Working Memory
```
Working Memory Configuration
────────────────────────────

Mode:
  [1] Local (default, in-process)
  [2] Remote (FAISSx server)

Select (1-2): _
```

#### If Local:
```
Max memory (MB):
  "auto" = 10% of RAM (min 64MB, max 1GB)
  Or specify integer (e.g., 256)

Max memory ["auto"]: _

Vector dimension [1536]: _

FIFO cleanup interval (minutes) [5]: _
```

#### If Remote:
```
FAISSx server URL: tcp://localhost:8000

API key (will be stored in secrets):
  API Key: _

Tenant ID: _

Max memory (MB) [256]: _
  Note: Remote mode requires explicit integer, not "auto"
```

**Output:**
```yaml
memory:
  working:
    mode: "local"  # or "remote"
    max_memory_mb: "auto"
    vector_dimension: 1536
    fifo_interval_min: 5
    # remote:  # if remote mode
    #   url: "tcp://localhost:8000"
    #   api_key: "${{ secrets.FAISSX_API_KEY }}"
    #   tenant: "tenant-id"
```

---

### Flow 2: Buffer Memory
```
Buffer Memory Configuration
───────────────────────────

Context window size (recent messages) [10]: _

Buffer multiplier (total = size × multiplier) [10]: _

Enable vector similarity search? (Y/n): _
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
Persistent Memory Configuration
───────────────────────────────

ℹ Persistent memory requires a PostgreSQL or SQLite database.

Database type:
  [1] PostgreSQL
  [2] SQLite

Select (1-2): _
```

#### If PostgreSQL:
```
Connection details:
  Host [localhost]: _
  Port [5432]: _
  Database: mydb
  Username: _
  Password (will be stored in secrets): _

  Connection string: postgres://user:***@localhost:5432/mydb
```

#### If SQLite:
```
Database file path [./memory.db]: _

  Connection string: sqlite:///./memory.db
```

#### Common settings:
```
Query timeout (seconds) [30]: _

Enable user synopsis generation? (Y/n): _
  ℹ Generates LLM-synthesized summaries of user context

Synopsis cache TTL (seconds) [3600]: _
```

**Output:**
```yaml
memory:
  persistent:
    connection_string: "postgres://user:${{ secrets.DB_PASSWORD }}@localhost:5432/mydb"
    query_timeout_seconds: 30
    user_synopsis:
      enabled: true
      cache_ttl: 3600
```

## Secrets Created
- `FAISSX_API_KEY` (if remote working memory)
- `DB_PASSWORD` (if PostgreSQL persistent memory)

## Validation
- PostgreSQL connection string format
- SQLite path exists/writable
- Remote URL format (tcp:// or tcps://)

## Questions
1. Should we test database connection before saving?
2. Should embedding_model be configurable here or in `muxi config llm`?
