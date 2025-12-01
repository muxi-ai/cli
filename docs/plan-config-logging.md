# Plan: `muxi config logging`

## Overview
Configure logging/observability streams in formation.yaml.

## Command
```bash
muxi config logging
```

## Banner
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Logging                                   MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure where logs and events are sent. Supports multiple. │
│ destinations with different formats and filters.             │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: Action Selection
```
What would you like to do?
  ◯ Add a new log stream
  ◯ View/edit current streams
  ◯ Remove a stream
```

---

### Flow 1: Add New Stream

```
Add Log Stream

  Select transport type:
    ◯ stdout (console output)
    ◯ file (local file)
    ◯ http (webhook to logging service)
    ◯ kafka (message queue)
    ◯ trail (MUXI Trail observability)
```

#### stdout:
```
Console Output

  Log level (debug shows everything, error shows only errors)
  ✓ Level: info

  Output format for log entries
    ◯ jsonl (structured, machine-readable)
    ◯ text (human-readable)
  ✓ Format: jsonl

  ✓ Added stdout stream
```

**Output:**
```yaml
logging:
  streams:
    - transport: "stdout"
      level: "info"
      format: "jsonl"
```

#### file:
```
File Output

  Path to log file (will be created if doesn't exist)
  ✓ Path: /var/log/formation.log

  Log level (debug shows everything, error shows only errors)
    ◯ debug
    ◯ info
    ◯ warn
    ◯ error
  ✓ Level: info

  Output format for log entries
    ◯ jsonl (structured, machine-readable)
    ◯ text (human-readable)
  ✓ Format: jsonl

  ✓ Added file stream
```

**Output:**
```yaml
logging:
  streams:
    - transport: "file"
      destination: "/var/log/formation.log"
      level: "info"
      format: "jsonl"
```

#### http:
```
HTTP Stream

  Endpoint URL for log ingestion
  ✓ URL: https://logs.company.com/ingest

  Output format (choose based on your logging service)
    ◯ jsonl (generic JSON lines)
    ◯ datadog
    ◯ splunk_hec
    ◯ elastic
    ◯ grafana_loki
    ◯ newrelic
    ◯ opentelemetry
  ✓ Format: datadog

  Authentication method
    ◯ None
    ◯ Bearer token
    ◯ API key header
    ◯ Basic auth
  ✓ Auth: Bearer token

  Bearer token for authentication
  ✓ Saved LOGGING_BEARER_TOKEN to secrets

  Filter which events to send (glob patterns, comma-separated)
  Example: request.*, error.*, agent.*
  ✓ Events: * (all events)

  ✓ Added http stream
```

**Output:**
```yaml
logging:
  streams:
    - transport: "http"
      destination: "https://logs.company.com/ingest"
      format: "datadog"
      auth:
        type: "bearer"
        token: "${{ secrets.LOGGING_BEARER_TOKEN }}"
      events: ["*"]
```

#### kafka:
```
Kafka Stream

  Broker addresses (comma-separated)
  ✓ Brokers: broker1:9092,broker2:9092

  Topic name for log messages
  ✓ Topic: formation-logs

  Output format for messages
    ◯ jsonl
    ◯ msgpack
  ✓ Format: jsonl

  Authentication method
    ◯ None
    ◯ SASL (username/password)
  ✓ Auth: SASL

  SASL username
  ✓ Username: kafka-user

  SASL password
  ✓ Saved KAFKA_PASSWORD to secrets

  ✓ Added kafka stream
```

**Output:**
```yaml
logging:
  streams:
    - transport: "kafka"
      brokers: ["broker1:9092", "broker2:9092"]
      topic: "formation-logs"
      format: "jsonl"
      auth:
        type: "sasl"
        username: "kafka-user"
        password: "${{ secrets.KAFKA_PASSWORD }}"
```

#### trail:
```
MUXI Trail

  MUXI Trail is our hosted observability service.
  Sign up at: https://trail.muxi.ai

  Trail API token
  ✓ Saved TRAIL_TOKEN to secrets

  ✓ Added trail stream
```

**Output:**
```yaml
logging:
  streams:
    - transport: "trail"
      token: "${{ secrets.TRAIL_TOKEN }}"
```

---

### Flow 2: View/Edit Current Streams

```
Current Streams

  [1] stdout (info, jsonl)
  [2] file → /var/log/formation.log (info, jsonl)
  [3] trail (configured)

  Select a stream to edit, or press Enter to go back: _
```

If a stream is selected, show its current settings with `[current]` indicators and allow modification.

---

### Flow 3: Remove Stream

```
Remove Stream

  Select stream to remove:
    ◯ stdout (info, jsonl)
    ◯ file → /var/log/formation.log (info, jsonl)
    ◯ trail (configured)

  ✓ Selected: file

  Remove this stream? [y/N]: y

  ✓ Stream removed
```

---

## Secrets Created

| Secret | When Created |
|--------|--------------|
| `LOGGING_BEARER_TOKEN` | HTTP stream with bearer auth |
| `LOGGING_API_KEY` | HTTP stream with API key auth |
| `KAFKA_PASSWORD` | Kafka stream with SASL auth |
| `TRAIL_TOKEN` | MUXI Trail stream |

## Environment Variable Detection

Check for existing env vars and offer to import:
- `DATADOG_API_KEY` → for Datadog format
- `SPLUNK_HEC_TOKEN` → for Splunk HEC format
- `TRAIL_TOKEN` → for MUXI Trail

## Validation

- Valid URL format for HTTP destinations
- Valid broker format for Kafka (host:port)
- Log level is one of: debug, info, warn, error
- Events are valid glob patterns

## Implementation Notes

1. Use radio boxes (◯) for selections, not numbered lists
2. Add dimmed hints above each prompt explaining the setting
3. Indent content under section headers by 2 spaces
4. Store sensitive values in secrets, reference with `${{ secrets.NAME }}`
5. Detect env vars and offer to import (like LLM config)
6. Show `[current]` on options that match existing configuration
7. Default selection to current value when editing
