# Formations Guide

This guide covers creating and configuring MUXI formations.

## What is a Formation?

A formation is a complete AI agent system - it contains agents, tools (MCPs), workflows (SOPs), triggers, and configuration. Think of it as a self-contained project that can be deployed and run.

## Creating a Formation

### Interactive Mode (Recommended)

```bash
muxi new formation
```

The wizard guides you through:
1. **Formation ID** - Unique identifier (e.g., `my-chatbot`)
2. **Description** - What this formation does
3. **Streaming** - Enable real-time response streaming
4. **Async Mode** - Enable webhook callbacks for long-running tasks
5. **LLM Provider** - Choose from 21 providers (OpenAI, Anthropic, etc.)
6. **API Key** - Optional, can add later via `muxi secrets set`

### With Name

```bash
muxi new formation my-chatbot
```

Skips the ID prompt but runs the full wizard.

### Non-Interactive

```bash
muxi new formation my-chatbot --no-wizard
```

Creates minimal scaffolding - configure manually or use `muxi config` commands.

## Directory Structure

```
my-formation/
├── formation.yaml      # Main configuration
├── secrets             # Secret keys template
├── secrets.enc         # Encrypted secrets (after setup)
├── .key                # Encryption key (keep secure!)
├── .gitignore          # Git ignore rules
├── .muxi               # Project defaults
├── README.md           # Getting started guide
├── agents/             # Agent configurations
├── mcps/               # MCP server configs
├── a2a/                # A2A service configs
├── sops/               # Standard Operating Procedures
├── triggers/           # Webhook trigger templates
└── knowledge/          # Knowledge base files
```

## formation.yaml

The main configuration file:

```yaml
schema: "1.0.0"
id: "my-chatbot"
name: "My Chatbot"
description: "A helpful assistant"
version: "0.1.0"

llm:
  provider: anthropic
  model: claude-sonnet-4-20250514
  api_key: ${{ secrets.ANTHROPIC_API_KEY }}

streaming: true
async: false

# Optional sections (add as needed):
# memory: ...
# overlord: ...
# a2a: ...
# logging: ...
```

## LLM Providers

The wizard offers 21 providers in three categories:

**Cloud Providers:**
- OpenAI, Anthropic, Google (Gemini), Mistral, Cohere
- Groq, Together AI, Fireworks AI, Perplexity, DeepSeek
- xAI (Grok), AI21, Aleph Alpha, Anyscale, Replicate
- OpenRouter, Hugging Face

**Local Providers:**
- Ollama / llama.cpp (requires base URL)

**Enterprise Providers:**
- Azure OpenAI, AWS Bedrock, GCP Vertex AI
- These add provider-specific template sections

## Configuration Commands

After creating a formation, use `muxi config` to add features:

```bash
muxi config a2a          # Enable A2A communication
muxi config llm          # Change LLM provider (coming soon)
muxi config memory       # Configure memory (coming soon)
muxi config overlord     # Configure overlord (coming soon)
```

## Next Steps After Creation

```bash
cd my-formation

# Add components
muxi new agent assistant
muxi new mcp weather-api
muxi new sop customer-support

# Set up secrets
muxi secrets setup       # Or: muxi secrets set OPENAI_API_KEY

# Validate configuration
muxi validate            # Coming soon

# Deploy
muxi deploy              # Coming soon
```

## Best Practices

1. **Use descriptive IDs** - `customer-support-bot` not `bot1`
2. **Set up secrets early** - Don't commit API keys
3. **Add .key to .gitignore** - Already done by default
4. **Commit secrets template** - Share required keys with team
5. **Use streaming** - Better UX for chat applications
6. **Start simple** - Add features incrementally
