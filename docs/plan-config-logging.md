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
│ ℹ Configure where logs and events are sent. Supports multiple│
│ destinations with different formats and filters.             │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: Action Selection
```
What would you like to do?
  [1] Add a new log stream
  [2] View current streams
  [3] Remove a stream
  [4] Enable/disable logging

Select (1-4): _
```

---

### Flow 1: Add New Stream
```
Select transport type:
  [1] stdout (console output)
  [2] file (local file)
  [3] stream (HTTP/Kafka/ZMQ)
  [4] trail (MUXI Trail service)

Select (1-4): _
```

#### stdout:
```
Console Output Configuration
────────────────────────────

Log level:
  [1] debug (verbose)
  [2] info (default)
  [3] warn
  [4] error

Select [2]: _

Format:
  [1] jsonl (structured, default)
  [2] text (human-readable)

Select [1]: _

✓ stdout stream configured
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
File Output Configuration
─────────────────────────

File path: /var/log/formation.log

Log level:
  [1] debug
  [2] info
  [3] warn
  [4] error

Select [2]: _

Format:
  [1] jsonl
  [2] text

Select [1]: _

✓ file stream configured
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

#### stream (HTTP/Kafka/ZMQ):
```
Stream Output Configuration
───────────────────────────

Protocol:
  [1] HTTP/HTTPS (webhook)
  [2] Kafka
  [3] ZMQ (tcp/ipc)

Select (1-3): _
```

##### HTTP:
```
HTTP Stream Configuration
─────────────────────────

Endpoint URL: https://logs.company.com/ingest

Format:
  [1] jsonl (generic)
  [2] datadog_json
  [3] splunk_hec
  [4] elastic_bulk
  [5] grafana_loki
  [6] newrelic_json
  [7] opentelemetry

Select [1]: _

Authentication:
  [1] None
  [2] Bearer token
  [3] API key
  [4] Basic auth

Select: 2

Bearer token (will be stored in secrets): _

Filter events (comma-separated, or * for all) [*]: request.*,error.*
```

**Output:**
```yaml
logging:
  streams:
    - transport: "stream"
      destination: "https://logs.company.com/ingest"
      protocol: "http"
      format: "datadog_json"
      auth:
        type: "bearer"
        token: "${{ secrets.LOG_BEARER_TOKEN }}"
      events: ["request.*", "error.*"]
```

##### Kafka:
```
Kafka Stream Configuration
──────────────────────────

Broker(s) (comma-separated): broker1:9092,broker2:9092

Topic: application-logs

Format:
  [1] jsonl
  [2] msgpack

Select [1]: _

Authentication:
  [1] None
  [2] SASL (username/password)

Select: 2

Username: _
Password (will be stored in secrets): _
```

**Output:**
```yaml
logging:
  streams:
    - transport: "stream"
      destination: "kafka://broker1:9092,broker2:9092"
      protocol: "kafka"
      topic: "application-logs"
      format: "jsonl"
      auth:
        type: "sasl"
        username: "kafka-user"
        password: "${{ secrets.KAFKA_PASSWORD }}"
```

##### ZMQ:
```
ZMQ Stream Configuration
────────────────────────

URL: tcp://localhost:5555
  ℹ Use tcps:// for encrypted connections

Format:
  [1] jsonl
  [2] msgpack

Select [1]: _
```

#### trail:
```
MUXI Trail Configuration
────────────────────────

ℹ MUXI Trail is our hosted observability service.
  Sign up at: https://trail.muxi.ai

Trail token (will be stored in secrets): _

✓ Trail stream configured
```

**Output:**
```yaml
logging:
  streams:
    - transport: "trail"
      token: "${{ secrets.TRAIL_TOKEN }}"
```

---

### Flow 2: View Current Streams
```
Current Logging Configuration
─────────────────────────────

Logging: enabled

Streams:
  [1] stdout (info, jsonl)
  [2] file → /var/log/formation.log (debug, text)
  [3] trail → MUXI Trail (configured)

Press Enter to continue...
```

---

### Flow 3: Remove Stream
```
Select stream to remove:
  [1] stdout (info, jsonl)
  [2] file → /var/log/formation.log (debug, text)
  [3] trail → MUXI Trail

Select (1-3): 2

Remove file stream? (y/N): y

✓ Stream removed
```

---

### Flow 4: Enable/Disable Logging
```
Logging is currently: enabled

Disable logging? (y/N): _
```

## Secrets Created
- `LOG_BEARER_TOKEN` (HTTP bearer auth)
- `LOG_API_KEY` (HTTP API key auth)
- `KAFKA_PASSWORD` (Kafka SASL)
- `TRAIL_TOKEN` (MUXI Trail)

## Validation
- Valid URL format for destinations
- Required fields per transport type
- Valid log levels
- Valid event patterns (glob syntax)

## Questions
1. Should we support editing existing streams, or just add/remove?
2. Should we test connectivity for HTTP endpoints?
3. Default stream if none configured - stdout info jsonl?
