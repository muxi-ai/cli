# Overlord Configuration Guide

Configure the Overlord's behavior and settings using `muxi config overlord`.

## Quick Start

```bash
cd my-formation
muxi config overlord
```

## What You Can Configure

### 1. Persona

The persona defines how the Overlord communicates with users.

**Options:**
- **Enter text directly** - Opens `$EDITOR` with an empty document
- **Load from file** - Load persona from a markdown or text file

```yaml
overlord:
  persona: |
    You are a helpful AI assistant for TechCorp.
    You are professional, concise, and accurate.
    You specialize in technical support.
```

### 2. Response Options

Control how the Overlord formats and delivers responses.

```yaml
overlord:
  response:
    format: "markdown"    # markdown, text, or html
    streaming: true       # Stream responses as generated
    progress: true        # Show status updates during operations
```

**Format options:**
- **Markdown** - Rich formatting (default)
- **Plain text** - No formatting
- **HTML** - Web rendering

### 3. Workflow Behavior

Configure how the Overlord routes tasks and manages workflows.

```yaml
overlord:
  workflow:
    routing_strategy: "capability"
    auto_decomposition: true
    plan_approval_threshold: 7
    complexity_method: "hybrid"
    parallel_execution: true
    max_parallel_tasks: 5
    enable_agent_affinity: true
    error_recovery: "retry_with_backoff"
    timeouts:
      task: 300
      workflow: 3600
```

**Routing strategies:**
| Strategy | Description |
|----------|-------------|
| Capability-based | Match task to agent skills (default) |
| Load-balanced | Distribute work evenly |
| Round-robin | Rotate through agents |
| Priority-based | Prefer higher-priority agents |

**Complexity methods:**
| Method | Description |
|--------|-------------|
| Hybrid | Balanced approach (default) |
| Heuristic | Fast, rule-based |
| LLM | Accurate, uses LLM |

**Error recovery:**
| Strategy | Description |
|----------|-------------|
| Retry with backoff | Retry with increasing delays (default) |
| Retry with alternate | Try a different agent |
| Fail fast | Stop immediately |
| Skip and continue | Return partial results |

### 4. Clarification Settings

Configure how the Overlord asks clarifying questions.

```yaml
overlord:
  clarification:
    style: "conversational"
    persist_learned_info: false
    max_rounds:
      direct: 3
      brainstorm: 10
      planning: 7
      execution: 3
```

**Clarification styles:**
- **Conversational** - Friendly, natural (default)
- **Formal** - Professional, structured
- **Brief** - Minimal, direct

**Max rounds per mode:**
| Mode | Default | Purpose |
|------|---------|---------|
| Direct | 3 | Quick disambiguation |
| Brainstorm | 10 | Creative exploration |
| Planning | 7 | Requirements gathering |
| Execution | 3 | Parameter clarification |

## Example: Full Overlord Configuration

```yaml
overlord:
  persona: |
    You are a helpful AI assistant for TechCorp.
    You specialize in cloud infrastructure and DevOps.
    Be concise but thorough in your responses.
  
  response:
    format: "markdown"
    streaming: true
    progress: true
  
  workflow:
    routing_strategy: "capability"
    auto_decomposition: true
    plan_approval_threshold: 7
    complexity_method: "hybrid"
    parallel_execution: true
    max_parallel_tasks: 5
    enable_agent_affinity: true
    error_recovery: "retry_with_backoff"
    timeouts:
      task: 300
      workflow: 3600
  
  clarification:
    style: "conversational"
    persist_learned_info: false
    max_rounds:
      direct: 3
      brainstorm: 10
      planning: 7
      execution: 3
```

## Settings Reference

### Workflow Settings

| Setting | Default | Range | Description |
|---------|---------|-------|-------------|
| `plan_approval_threshold` | 7 | 1-10 | Complexity above this requires user approval |
| `max_parallel_tasks` | 5 | 1-20 | Maximum concurrent tasks |
| `timeouts.task` | 300 | >0 | Task timeout in seconds |
| `timeouts.workflow` | 3600 | >0 | Workflow timeout in seconds |

### Clarification Settings

| Setting | Default | Range | Description |
|---------|---------|-------|-------------|
| `max_rounds.direct` | 3 | 1-32 | Max questions in direct mode |
| `max_rounds.brainstorm` | 10 | 1-32 | Max questions in brainstorm mode |
| `max_rounds.planning` | 7 | 1-32 | Max questions in planning mode |
| `max_rounds.execution` | 3 | 1-32 | Max questions in execution mode |

## Tips

1. **Start with defaults** - The default settings work well for most use cases
2. **Lower threshold for safety** - Reduce `plan_approval_threshold` if you want more user control
3. **Enable parallel execution** - Speeds up complex multi-agent workflows
4. **Use agent affinity** - Improves consistency by routing similar tasks to the same agent
5. **Conversational style** - Best for user-facing applications

## LLM Settings

The Overlord uses the formation's text model by default. To configure a different model or adjust settings like temperature (lower is better for routing), use:

```bash
muxi config llm
```

## Related

- [LLM Guide](llm.md) - Configure the text model used by the Overlord
- [Agents Guide](agents.md) - Configure agents that the Overlord routes to
- [Formations Guide](formations.md) - Formation structure overview
