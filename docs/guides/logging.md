# Logging Configuration Guide

Configure logging streams using `muxi config logging`.

## Quick Start

```bash
cd my-formation
muxi config logging
```

## What You Can Configure

### Transport Types

| Transport | Description |
|-----------|-------------|
| **stdout** | Console output for development |
| **file** | Local file logging |
| **http** | Webhook to logging services (Datadog, Splunk, etc.) |
| **kafka** | Message queue for high-volume logs |

---

## stdout (Console Output)

Basic console logging for development.

```yaml
logging:
  streams:
    - transport: "stdout"
      level: "info"
      format: "jsonl"
```

**Options:**
| Setting | Values | Default |
|---------|--------|---------|
| `level` | debug, info, warn, error | info |
| `format` | jsonl, text | jsonl |

---

## file (Local File)

Log to a local file.

```yaml
logging:
  streams:
    - transport: "file"
      destination: "/var/log/formation.log"
      level: "info"
      format: "jsonl"
```

**Options:**
| Setting | Description |
|---------|-------------|
| `destination` | Path to log file |
| `level` | debug, info, warn, error |
| `format` | jsonl, text |

---

## http (Webhook)

Send logs to HTTP endpoints like Datadog, Splunk, Elastic, etc.

```yaml
logging:
  streams:
    - transport: "http"
      destination: "https://logs.company.com/ingest"
      format: "datadog"
      auth:
        type: "bearer"
        token: "${{ secrets.LOGGING_BEARER_TOKEN }}"
```

**Supported formats:**
- jsonl (generic JSON lines)
- datadog
- splunk (HEC)
- elastic
- loki (Grafana)
- newrelic
- otlp (OpenTelemetry)

**Authentication types:**
| Type | Fields |
|------|--------|
| none | - |
| bearer | token |
| apikey | header, key |
| basic | username, password |

---

## kafka (Message Queue)

Send logs to Kafka for high-volume processing.

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

**Options:**
| Setting | Description |
|---------|-------------|
| `brokers` | Kafka broker addresses (comma, space, or newline-separated) |
| `topic` | Topic name for logs |
| `format` | jsonl, msgpack |
| `auth` | none or sasl (username/password) |

---

## Multiple Streams

You can configure multiple streams to send logs to different destinations:

```yaml
logging:
  streams:
    - transport: "stdout"
      level: "debug"
      format: "text"
    
    - transport: "http"
      destination: "https://logs.datadog.com/v1/input"
      format: "datadog"
      auth:
        type: "bearer"
        token: "${{ secrets.DATADOG_API_KEY }}"
    
    - transport: "file"
      destination: "/var/log/formation.log"
      level: "error"
      format: "jsonl"
```

---

## Secrets Created

| Secret | When Created |
|--------|--------------|
| `LOGGING_BEARER_TOKEN` | HTTP stream with bearer auth |
| `LOGGING_API_KEY` | HTTP stream with API key auth |
| `LOGGING_BASIC_PASSWORD` | HTTP stream with basic auth |
| `KAFKA_PASSWORD` | Kafka stream with SASL auth |

---

## Managing Streams

### Add a Stream
```bash
muxi config logging
# Select: Add a new log stream
```

### View/Edit Streams
```bash
muxi config logging
# Select: View/edit current streams
```

### Remove a Stream
```bash
muxi config logging
# Select: Remove a stream
```

---

## Related

- [Secrets Guide](secrets.md) - Managing encrypted secrets
