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
│ ℹ The Overlord orchestrates your agents, routes tasks, and   │
│ manages conversations. Configure its personality and behavior.│
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: What to Configure
```
What would you like to configure?
  [1] Persona (identity and communication style)
  [2] LLM settings (routing model)
  [3] Response format (markdown, streaming, widgets)
  [4] Workflow behavior (decomposition, approval, timeouts)
  [5] Clarification settings (question style, limits)
  [6] Caching (routing decision cache)

Select (1-6): _
```

---

### Flow 1: Persona
```
Overlord Persona
────────────────

The persona defines how the Overlord communicates with users.

Current persona:
  "You are a helpful AI assistant."

How would you like to set the persona?
  [1] Enter text directly
  [2] Load from file
  [3] Use a template

Select (1-3): _
```

#### Enter directly:
```
Enter persona (multi-line, end with empty line):
> You are Aria, a friendly AI assistant for TechCorp.
> You are helpful, professional, and concise.
> You specialize in technical support and product information.
>

✓ Persona updated
```

#### Load from file:
```
File path: ./persona.md

✓ Persona loaded from ./persona.md
```

#### Use template:
```
Select template:
  [1] Professional assistant
  [2] Technical expert
  [3] Creative helper
  [4] Customer support
  [5] Research assistant

Select: 2

Customization:
  Company/product name: TechCorp
  Specialization: cloud infrastructure

✓ Persona generated from template
```

**Output:**
```yaml
overlord:
  persona: |
    You are Aria, a friendly AI assistant for TechCorp.
    You are helpful, professional, and concise.
    You specialize in technical support and product information.
```

---

### Flow 2: LLM Settings
```
Overlord LLM Configuration
──────────────────────────

The Overlord uses its own LLM for routing and task delegation.

Use formation default model? (Y/n): n

Provider:
  [1] OpenAI
  [2] Anthropic
  [3] Local (llama.cpp)
  [4] Other

Select: 1

Model:
  [1] gpt-4o (recommended for complex routing)
  [2] gpt-4o-mini (faster, cheaper)
  [3] Other

Select: 2

Settings:
  Temperature [0.2]: _
    ℹ Lower = more consistent routing
  
  Max tokens [2000]: _
  Timeout (seconds) [45]: _
  Fallback model (optional): openai/gpt-4o

Max extraction tokens [500]: _
  ℹ Tokens used when extracting content from media for routing
```

**Output:**
```yaml
overlord:
  llm:
    model: "openai/gpt-4o-mini"
    settings:
      temperature: 0.2
      max_tokens: 2000
      timeout_seconds: 45
      fallback_model: "openai/gpt-4o"
    max_extraction_tokens: 500
```

---

### Flow 3: Response Format
```
Response Format Configuration
─────────────────────────────

Default format:
  [1] markdown (default)
  [2] text (plain text)
  [3] html
  [4] json

Select [1]: _

Enable interactive widgets? (Y/n): _
  ℹ Buttons, forms, charts in responses

Enable streaming? (Y/n): _
  ℹ Stream responses in real-time

Enable progress events? (Y/n): _
  ℹ Show progress updates during processing
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

### Flow 4: Workflow Behavior
```
Workflow Configuration
──────────────────────

Auto-decomposition:
  Enable automatic task decomposition? (Y/n): _
  ℹ Breaks complex requests into subtasks

Plan approval threshold (1-10) [7]: _
  ℹ Complexity above this requires user approval

Complexity calculation method:
  [1] heuristic (fast, rule-based)
  [2] llm (accurate, slower)
  [3] hybrid (balanced)

Select [3]: _

Routing strategy:
  [1] capability_based (match task to agent skills)
  [2] load_balanced (distribute evenly)
  [3] round_robin (rotate agents)
  [4] priority_based (prefer high-priority agents)

Select [1]: _

Enable agent affinity? (Y/n): _
  ℹ Prefer agents that succeeded on similar tasks

Parallel execution:
  Enable parallel task execution? (Y/n): _
  Max parallel tasks (1-20) [5]: _

Error recovery strategy:
  [1] retry_with_backoff (default)
  [2] retry_with_alternate (try different agent)
  [3] fail_fast (stop immediately)
  [4] skip_and_continue (partial results)

Select [1]: _

Timeouts:
  Task timeout (seconds) [300]: _
  Workflow timeout (seconds) [3600]: _
  Max workflow duration (seconds) [7200]: _
```

**Output:**
```yaml
overlord:
  workflow:
    auto_decomposition: true
    plan_approval_threshold: 7
    complexity_method: "hybrid"
    routing_strategy: "capability_based"
    enable_agent_affinity: true
    parallel_execution: true
    max_parallel_tasks: 5
    error_recovery: "retry_with_backoff"
    timeouts:
      task_timeout: 300
      workflow_timeout: 3600
    max_timeout_seconds: 7200
```

---

### Flow 5: Clarification Settings
```
Clarification Configuration
───────────────────────────

ℹ The Overlord asks clarifying questions for ambiguous requests.

Question style:
  [1] conversational (friendly, natural)
  [2] formal (professional, structured)
  [3] brief (minimal, to the point)

Select [1]: _

Persist learned information across sessions? (y/N): _
  ℹ Privacy: 'no' = session-only, 'yes' = remembers preferences

Max clarification rounds per mode:
  Direct (quick disambiguation) [3]: _
  Brainstorm (creative exploration) [10]: _
  Planning (requirements gathering) [7]: _
  Execution (parameter clarification) [3]: _
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

### Flow 6: Caching
```
Routing Cache Configuration
───────────────────────────

Enable routing decision caching? (Y/n): _
  ℹ Caches routing decisions for similar requests

Cache TTL (seconds) [3600]: _
```

**Output:**
```yaml
overlord:
  caching:
    enabled: true
    ttl: 3600
```

## Validation
- Temperature 0.0-1.0
- Thresholds in valid ranges
- Timeout values positive
- Max rounds 1-32

## Questions
1. Should persona editing open $EDITOR for multi-line?
2. Should we show "current value" for each setting?
3. Templates - should we include actual persona text or just structure?
