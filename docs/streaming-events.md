# MUXI Runtime Streaming Events

Reference for all streaming events emitted by the MUXI Runtime.

## Event Structure

```json
{
  "request_id": "req_xxx",
  "user_id": "0",
  "session_id": null,
  "type": "thinking|planning|progress|content|completed|error",
  "content": "Human-readable message",
  "timestamp": 1234567890.123,
  "stage": "stage_identifier",
  // ... additional metadata
}
```

## Event Types

| Type | Purpose |
|------|---------|
| `thinking` | Processing/analyzing |
| `planning` | Decision making |
| `progress` | Status updates |
| `content` | Intermediate content |
| `completed` | Final response |
| `error` | Error occurred |

---

## Normal Agent Flow

| Order | Type | Stage | Content |
|-------|------|-------|---------|
| 1 | `progress` | `init` | Random: "On it...", "Let me check...", etc. |
| 2 | `thinking` | `process_sync_start` | "Understanding the user's request..." |
| 3 | `planning` | `agent_selection` | "Determining the best agent..." |
| 4 | `progress` | `agent_selected` | "I'll use an agent with the right capabilities..." |
| 5 | `progress` | `agent_processing` | "Processing request..." |
| 6 | `planning` | `agent_planning` | "Planning approach..." |
| 7 | `progress` | `tool_call` | "Using the {Tool} tool..." (repeats per tool) |
| 8 | `progress` | `response_preparation` | "Preparing response..." |
| 9 | `completed` | - | *final response* |

---

## Fast Path Flow

For greetings: "hi", "hello", "hey", "thanks", "thank you", "ok", "okay", "got it".

| Order | Type | Stage | Content |
|-------|------|-------|---------|
| 1 | `progress` | `init` | Random: "On it...", etc. |
| 2 | `thinking` | `process_sync_start` | "Understanding the user's request..." |
| 3 | `progress` | `response_preparation` | "Preparing response..." |
| 4 | `completed` | - | *final response* (`fast_path: true`) |

---

## Workflow Path Flow

For complex requests (complexity >= threshold).

| Order | Type | Stage | Content |
|-------|------|-------|---------|
| 1 | `progress` | `init` | Random: "On it...", etc. |
| 2 | `thinking` | `process_sync_start` | "Understanding the user's request..." |
| 3 | `planning` | `workflow_decomposition` | "This is a complex request..." |
| 4 | `progress` | `workflow_start` | "Executing workflow with N tasks..." |
| 5 | `progress` | `workflow_execution` | "Processing workflow tasks... (X/Y - Z%)" (repeats) |
| 6 | `thinking` | `workflow_synthesis` | "Synthesizing results..." |
| 7 | `content` | `final_response` | *workflow result* |
| 8 | `completed` | `final` | *final response* |

---

## Special Events

| Type | Stage | Content | When |
|------|-------|---------|------|
| `thinking` | `clarification_needed` | "I need to clarify..." | Clarification required |
| `thinking` | `tool_retry` | "That didn't work, trying another approach..." | Tool call failed, retrying |
| `planning` | `credential_request` | "I need credentials for {service}..." | Credential needed |
| `error` | `security_blocked` | "I can't process that request." | Security threat |

---

## Key Metadata Fields

| Field | Type | Description |
|-------|------|-------------|
| `fast_path` | bool | True if used fast path |
| `processing_time_ms` | int | Total time in ms |
| `agent_name` / `agent_used` | string | Agent handling request |
| `tool_name` | string | Tool being called |
| `server_id` | string | MCP server for tool |
| `workflow_id` | string | Workflow UUID |
| `failed_tools` | array | List of failed tool names (on retry) |
| `skip_rephrase` | bool | Content not LLM-rephrased |

---

*Last updated: 2024-12-22*
