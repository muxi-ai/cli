# Memory Configuration Guide

Configure memory settings for your MUXI formation using `muxi config memory`.

## Quick Start

```bash
cd my-formation
muxi config memory
```

## Memory Types

### 1. Working Memory

Short-term vector memory for semantic search within a session.

**Modes:**
- **Local** - In-process memory (default, no external dependencies)
- **Remote** - FAISSx server (for distributed/persistent vector storage)

#### Local Mode
```yaml
memory:
  working:
    mode: "local"
    max_memory_mb: "auto"    # 10% of RAM (min 64MB, max 1GB)
    vector_dimension: 1536   # Must match embedding model
    fifo_interval_min: 5     # Cleanup interval
```

#### Remote Mode (FAISSx)
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

### 2. Buffer Memory

Conversation context that persists across messages in a session.

```yaml
memory:
  buffer:
    size: 10              # Recent messages to keep
    multiplier: 10        # Total buffer = size × multiplier
    vector_search: true   # Find relevant past messages
```

### 3. Persistent Memory

Long-term memory stored in a database, persists across sessions.

#### PostgreSQL (Production)

**Requirements:** PostgreSQL 17+ with pgvector extension

```yaml
memory:
  persistent:
    connection_string: "${{ secrets.PERSISTENT_DB_CONNECTION_STRING }}"
    query_timeout_seconds: 30
    user_synopsis:
      enabled: true
      cache_ttl: 3600
```

#### SQLite (Development)

```yaml
memory:
  persistent:
    connection_string: "sqlite:///./memory.db"
    query_timeout_seconds: 30
    user_synopsis:
      enabled: true
      cache_ttl: 3600
```

## Connection String Options

For PostgreSQL, you can provide either:

1. **Full connection string:**
   ```
   postgres://user:password@host:5432/database
   ```

2. **Just a hostname** - the wizard will prompt for:
   - Port (default: 5432)
   - Database name
   - Username
   - Password

The wizard builds the connection string and stores it securely.

## Secrets

| Secret | When Created | Purpose |
|--------|--------------|---------|
| `FAISSX_API_KEY` | Remote working memory | FAISSx authentication |
| `FAISSX_TENANT_ID` | Remote working memory | Multi-tenant isolation |
| `PERSISTENT_DB_CONNECTION_STRING` | PostgreSQL | Full database connection string |

**Note:** SQLite doesn't use secrets (just a file path).

## Environment Variable Detection

The wizard detects and offers to import:
- `FAISSX_API_KEY` - For remote working memory
- `DATABASE_URL` - For PostgreSQL connection string

## Settings Reference

### Working Memory Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `mode` | `local` | `local` or `remote` |
| `max_memory_mb` | `auto` | Memory limit (`auto` = 10% of RAM) |
| `vector_dimension` | (auto) | Auto-detected from embedding model |
| `fifo_interval_min` | `5` | How often to clean up old vectors |

**Note:** Vector dimension is automatically detected from your configured embedding model. If no embedding model is configured, it defaults to 384 (local all-MiniLM-L6-v2 model).

### Buffer Memory Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `size` | `10` | Recent messages to keep in context |
| `multiplier` | `10` | Total buffer = size × multiplier |
| `vector_search` | `true` | Use similarity search for relevance |

### Persistent Memory Settings

These settings apply to both PostgreSQL and SQLite:

| Setting | Default | Description |
|---------|---------|-------------|
| `query_timeout_seconds` | `30` | Max wait time for queries |
| `user_synopsis.enabled` | `true` | Generate LLM summaries of user context |
| `user_synopsis.cache_ttl` | `3600` | Cache duration for synopses (seconds) |

## Example: Full Memory Configuration

```yaml
memory:
  working:
    mode: "local"
    max_memory_mb: "auto"
    vector_dimension: 1536
    fifo_interval_min: 5
  
  buffer:
    size: 10
    multiplier: 10
    vector_search: true
  
  persistent:
    connection_string: "${{ secrets.PERSISTENT_DB_CONNECTION_STRING }}"
    query_timeout_seconds: 30
    user_synopsis:
      enabled: true
      cache_ttl: 3600
```

## Tips

1. **Start with local working memory** - No setup required
2. **Use SQLite for development** - Switch to PostgreSQL for production
3. **Vector dimensions are auto-detected** - From your configured embedding model
4. **Enable user synopsis** - Improves context awareness with LLM-generated summaries
5. **Set appropriate timeouts** - Prevent hanging queries in production
6. **Look for [current]** - The wizard shows which options are already configured

## PostgreSQL Setup

Install pgvector extension:

```sql
-- PostgreSQL 17+
CREATE EXTENSION IF NOT EXISTS vector;
```

Create the database:

```sql
CREATE DATABASE muxi_memory;
```

## Runtime Memory Management

Once your formation is deployed, use CLI commands to manage memories:

```bash
# View runtime memory configuration
muxi memory status

# List memories for a user
muxi memory list -u alice

# Add a memory
muxi memory add -u alice "Prefers dark mode"

# Delete a memory
muxi memory delete -u alice mem_abc123
```

See [Admin Guide](admin.md) for full details on runtime memory commands.

## Related

- [Admin Guide](admin.md) - Runtime memory management
- [LLM Guide](llm.md) - Configure embedding models
- [Secrets Guide](secrets.md) - Managing sensitive values
- [Formations Guide](formations.md) - Formation structure
