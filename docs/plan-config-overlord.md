# Plan: `muxi config overlord`

## Overview
Configure the Overlord's behavior, persona, and settings.

## Command
```bash
muxi config overlord
```

## Banner
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Overlord                                  MUXI │
│──────────────────────────────────────────────────────────────│
│ The Overlord orchestrates agents, creates and routes tasks,  │
│ manages conversations, and handles clarifications.           │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: What to Configure
```
What would you like to configure?
  ◯ Persona (identity and communication style)
  ◯ Response options (format, streaming, progress)
  ◯ Workflow behavior (routing, decomposition, timeouts)
  ◯ Clarification settings (question style, limits)
```

---

### Flow 1: Persona

```
Overlord Persona

  The persona defines how the Overlord communicates with users.

  How would you like to set the persona?
    ◯ Enter text directly (opens $EDITOR)
    ◯ Load from file
```

#### Enter directly:
```
  Opens $EDITOR with current persona for editing

  ✓ Persona updated
```

#### Load from file:
```
  Path to persona file (markdown or text)
  ✓ Path: ./persona.md

  ✓ Persona loaded from ./persona.md
```

**Output:**
```yaml
overlord:
  persona: |
    You are a helpful AI assistant.
    You are professional, concise, and accurate.
```

---

### Flow 2: Response Options

```
Response Format

  Default output format for responses
    ◯ markdown (rich formatting, default)
    ◯ text (plain text only)
    ◯ html (web rendering)
  ✓ Format: markdown

  Stream responses as they're generated
  Enable streaming? [Y/n]: y
  ✓ Streaming: enabled

  Show status updates during long operations
  Enable progress events? [Y/n]: y
  ✓ Progress: enabled
```

**Output:**
```yaml
overlord:
  response:
    format: "markdown"
    widgets: true
    streaming: true
    progress: true
```

---

### Flow 3: Workflow Behavior

```
Workflow Configuration

  How to select agents for tasks
    ◯ capability (match task to agent skills)
    ◯ load_balanced (distribute work evenly)
    ◯ round_robin (rotate through agents)
    ◯ priority (prefer higher-priority agents)
  ✓ Routing: capability

  Automatically break complex requests into subtasks
  Enable auto-decomposition? [Y/n]: y
  ✓ Auto-decomposition: enabled

  Complexity score (1-10) above which user approval is required
  ✓ Plan approval threshold: 7

  How to calculate task complexity
    ◯ heuristic (fast, rule-based)
    ◯ llm (accurate, uses LLM)
    ◯ hybrid (balanced approach)
  ✓ Complexity method: hybrid

  Run independent tasks simultaneously
  Enable parallel execution? [Y/n]: y
  ✓ Parallel execution: enabled

  Maximum tasks to run at once (1-20)
  ✓ Max parallel tasks: 5

  Prefer agents that succeeded on similar tasks before
  Enable agent affinity? [Y/n]: y
  ✓ Agent affinity: enabled
```

```
Timeouts & Error Handling

  Maximum time for a single task (seconds)
  ✓ Task timeout: 300

  Maximum time for an entire workflow (seconds)
  ✓ Workflow timeout: 3600

  What to do when a task fails
    ◯ retry_with_backoff (retry with increasing delays)
    ◯ retry_with_alternate (try a different agent)
    ◯ fail_fast (stop immediately)
    ◯ skip_and_continue (return partial results)
  ✓ Error recovery: retry_with_backoff
```

**Output:**
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

---

### Flow 4: Clarification Settings

```
Clarification Settings

  The Overlord asks clarifying questions for ambiguous requests.

  Communication style for clarifying questions
    ◯ conversational (friendly, natural)
    ◯ formal (professional, structured)
    ◯ brief (minimal, direct)
  ✓ Style: conversational

  Remember user preferences across sessions (privacy consideration)
  Persist learned info? [y/N]: n
  ✓ Persist: disabled
```

```
Clarification Limits

  Maximum questions per conversation mode:

  Direct mode (quick disambiguation)
  ✓ Max rounds: 3

  Brainstorm mode (creative exploration)
  ✓ Max rounds: 10

  Planning mode (requirements gathering)
  ✓ Max rounds: 7

  Execution mode (parameter clarification)
  ✓ Max rounds: 3
```

**Output:**
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

---

## Current Value Detection

For each setting, detect current value from formation.yaml and:
1. Show `[current]` next to the matching option
2. Use current value as default selection
3. Display current value in dimmed text before prompt

Example:
```
  How to select agents for tasks
    ◯ capability (match task to agent skills) [current]
    ◯ load_balanced (distribute work evenly)
    ◯ round_robin (rotate through agents)
    ◯ priority (prefer higher-priority agents)
```

## Validation

- Temperature: 0.0-1.0
- Thresholds: 1-10
- Timeout values: positive integers
- Max rounds: 1-32
- Max parallel tasks: 1-20

## Implementation Notes

1. Use radio boxes (◯) for selections, not numbered lists
2. Add dimmed hints above each prompt explaining the setting
3. Indent content under section headers by 2 spaces
4. Show `[current]` on options that match existing configuration
5. Default selection to current value when editing
6. Use $EDITOR for multi-line persona editing
7. Group related settings together (e.g., all timeouts in one section)

## Note on LLM Settings

The Overlord uses the formation's text model by default. If users need a different model for routing decisions, they should use `muxi config llm` to configure the text capability with appropriate settings (lower temperature recommended for routing).
