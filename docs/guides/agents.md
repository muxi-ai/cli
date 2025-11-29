# Agents Guide

This guide covers creating and configuring agents in MUXI formations.

## What is an Agent?

An agent is an AI persona within your formation. Each agent has:
- A unique ID and name
- A role defining its purpose
- A system prompt guiding its behavior
- Optional tool access (MCPs)
- Optional A2A visibility settings

## Creating an Agent

Must be run inside a formation directory.

### Interactive Mode (Recommended)

```bash
muxi new agent
```

The wizard guides you through:
1. **Agent ID** - Unique identifier (e.g., `research-assistant`)
2. **Display Name** - Human-friendly name
3. **Role** - Choose from predefined roles or custom
4. **System Prompt** - Instructions for the agent
5. **A2A Visibility** - Make agent callable by other formations

### With Name

```bash
muxi new agent research-assistant
```

### Non-Interactive

```bash
muxi new agent research-assistant --no-wizard
```

## Agent Roles

The wizard offers common roles with pre-written system prompts:

| Role | Description |
|------|-------------|
| **assistant** | General-purpose helpful assistant |
| **researcher** | Information gathering and analysis |
| **writer** | Content creation and editing |
| **coder** | Software development tasks |
| **analyst** | Data analysis and insights |
| **support** | Customer service and help |
| **coordinator** | Task delegation and orchestration |
| **custom** | Write your own system prompt |

## File Structure

Agents are stored in `agents/`:

```
my-formation/
└── agents/
    ├── research-assistant.yaml
    ├── content-writer.yaml
    └── coordinator.yaml
```

## Agent Configuration

```yaml
schema: "1.0.0"
id: "research-assistant"
name: "Research Assistant"
description: "Gathers and analyzes information from various sources"

role: researcher

system_prompt: |
  You are a research assistant specialized in gathering and analyzing information.
  
  Your responsibilities:
  - Search for relevant information using available tools
  - Synthesize findings into clear summaries
  - Cite sources when possible
  - Ask clarifying questions when needed

# Optional: Agent-specific LLM override
# llm:
#   provider: anthropic
#   model: claude-sonnet-4-20250514

# Optional: Make visible to other formations via A2A
# a2a:
#   visible: true
#   description: "Research and information gathering"

# Optional: Restrict to specific MCPs
# mcps:
#   - web-search
#   - document-reader
```

## A2A Visibility

To make an agent callable by other formations:

```yaml
a2a:
  visible: true
  description: "Brief description of what this agent can do"
```

When visible:
- Other formations can discover this agent
- They can send tasks to it via A2A protocol
- The agent appears in your formation's A2A service catalog

## Agent-Specific MCPs

By default, agents can access all formation-level MCPs. To restrict:

```yaml
mcps:
  - allowed-mcp-1
  - allowed-mcp-2
```

Or create agent-specific MCPs:

```bash
muxi new mcp private-tool --agent research-assistant
```

This creates `mcps/research-assistant/private-tool.yaml`.

## Multiple Agents

Formations can have multiple agents working together:

```
agents/
├── coordinator.yaml      # Delegates tasks
├── researcher.yaml       # Gathers information
├── writer.yaml           # Creates content
└── reviewer.yaml         # Quality checks
```

Use SOPs to define workflows between agents.

## Best Practices

1. **Single responsibility** - Each agent should have a focused purpose
2. **Clear system prompts** - Be specific about capabilities and constraints
3. **Use roles** - Start with predefined roles, customize as needed
4. **Limit tool access** - Only give agents the MCPs they need
5. **Document A2A agents** - Clear descriptions help other formations
6. **Test individually** - Verify each agent works before combining
