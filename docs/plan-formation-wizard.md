# Plan: Formation Wizard Enhancements

## Overview
Add async, streaming, and LLM provider configuration to `muxi new formation` wizard.

## New Flow
1. Formation ID
2. Description
3. Streaming responses (y/N)
4. Async responses (y/N) → webhook URL if yes
5. LLM provider selection (required)

---

## Streaming Responses
```
Enable streaming responses? (y/N): y
```
- Default: No
- If yes: Sets `overlord.response.streaming: true` in formation.yaml

---

## Async Responses
```
Enable async responses for long-running tasks? (y/N): y

Webhook URL:
  Async responses are delivered via webhook for long-running tasks.

  Webhook URL: example.com/webhook
```

### Webhook URL Handling:
- Auto-prepend `https://` if no protocol provided
- If empty URL provided:
  ```
  Disable async responses? (y/N): n

  Webhook URL: _
  ```
  - If yes: Skip async config entirely
  - If no: Ask for webhook URL again (loop)

### Output:
```yaml
async:
  threshold_seconds: 30
  webhook_url: "https://example.com/webhook"
```

---

## LLM Provider Selection (Required)

Based on onellm providers (vendor key matches `vendor/model` syntax):

```
LLM Provider Setup
──────────────────

You need at least one LLM provider for the formation to work.
You can add more later using 'muxi config llm'.

Select provider:
  [1] OpenAI
  [2] Anthropic
  [3] Google
  [4] Mistral
  [5] Groq
  [6] xAI (Grok)
  [7] DeepSeek
  [8] Cohere
  [9] Together
  [10] Fireworks
  [11] Perplexity
  [12] OpenRouter
  [13] Moonshot
  [14] Minimax
  [15] GLM
  [16] Vercel AI
  [17] Anyscale
  [18] Local (Ollama/llama_cpp)
  [19] Azure OpenAI
  [20] AWS Bedrock
  [21] GCP Vertex AI

Select (1-21): _
```

### Cloud Providers with Built-in Base URLs (1-17):

| # | Provider | Vendor Key | Default Model | API Key Prefix |
|---|----------|------------|---------------|----------------|
| 1 | OpenAI | `openai` | gpt-5-mini | sk- |
| 2 | Anthropic | `anthropic` | claude-4.5-sonnet | sk-ant- |
| 3 | Google | `google` | gemini-2.0-flash | AI |
| 4 | Mistral | `mistral` | mistral-large-latest | |
| 5 | Groq | `groq` | llama-3.3-70b-versatile | gsk_ |
| 6 | xAI | `xai` | grok-4 | xai- |
| 7 | DeepSeek | `deepseek` | deepseek-chat | sk- |
| 8 | Cohere | `cohere` | command-r-plus-08-2024 | |
| 9 | Together | `together` | meta-llama/Llama-3.3-70b-Instruct | |
| 10 | Fireworks | `fireworks` | accounts/fireworks/models/llama-v3p3-70b-instruct | |
| 11 | Perplexity | `perplexity` | sonar-pro | pplx- |
| 12 | OpenRouter | `openrouter` | openai/gpt-4o | sk-or- |
| 13 | Moonshot | `moonshot` | kimi-k2-instruct | |
| 14 | Minimax | `minimax` | abab6.5s-chat | |
| 15 | GLM | `glm` | glm-4-plus | |2
| 16 | Vercel AI | `vercel` | openai/gpt-5-mini | |
| 17 | Anyscale | `anyscale` | meta-llama/Meta-Llama-3.1-70B-Instruct | |

### Flow for Cloud Providers (1-17):
```
{Provider} API Key: _

✓ {provider} configured with {default_model} as default model
```
Output:
```yaml
llm:
  api_keys:
    {vendor}: "${{ secrets.{VENDOR}_API_KEY }}"
  models:
    - text: "{vendor}/{default_model}"
```

### Local (18) - Ollama/llama_cpp:
```
Local LLM Setup
───────────────

ℹ For Ollama, llama.cpp, or other local inference servers.

Provider:
  [1] Ollama (default: http://localhost:11434/v1)
  [2] llama_cpp (requires base URL)

Select (1-2): _

Base URL [http://localhost:11434/v1]: _

Model name (e.g., llama3, mistral, phi3): llama3

✓ Local LLM configured
```
Output:
```yaml
llm:
  api_keys:
    ollama: "local"  # placeholder for local
  models:
    - text: "ollama/llama3"
      base_url: "http://localhost:11434/v1"
```

### Enterprise Cloud Providers (19-21):

For Azure, Bedrock, and Vertex AI - add commented template to YAML
and instruct users to configure manually (requires existing cloud setup).

#### Azure OpenAI (19):
```
✓ Azure OpenAI template added to formation.yaml

Next steps:
  1. Edit formation.yaml and uncomment the Azure configuration
  2. Fill in your resource name, deployment name, and API version
  3. Run 'muxi secrets set AZURE_API_KEY' to add your API key
```
Adds to YAML (commented):
```yaml
# Azure OpenAI - uncomment and configure:
# llm:
#   api_keys:
#     azure: "${{ secrets.AZURE_API_KEY }}"
#   models:
#     - text: "azure/<deployment-name>"
#       api_base: "https://<resource-name>.openai.azure.com"
#       api_version: "2024-02-15-preview"
```

#### AWS Bedrock (20):
```
✓ AWS Bedrock template added to formation.yaml

Next steps:
  1. Edit formation.yaml and uncomment the Bedrock configuration
  2. Set your AWS region and model ID
  3. Ensure AWS credentials are configured (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
```
Adds to YAML (commented):
```yaml
# AWS Bedrock - uncomment and configure:
# llm:
#   models:
#     - text: "bedrock/anthropic.claude-3-sonnet"
#       aws_region: "us-east-1"
```

#### GCP Vertex AI (21):
```
✓ GCP Vertex AI template added to formation.yaml

Next steps:
  1. Edit formation.yaml and uncomment the Vertex AI configuration
  2. Fill in your project ID and region
  3. Ensure GOOGLE_APPLICATION_CREDENTIALS is set
```
Adds to YAML (commented):
```yaml
# GCP Vertex AI - uncomment and configure:
# llm:
#   models:
#     - text: "vertexai/gemini-1.5-pro"
#       project_id: "<project-id>"
#       region: "us-central1"
```

## Updated Formation Template

The template is dynamically generated based on user choices:

```yaml
schema: "1.0.0"

id: "my-formation"
description: "My formation"
version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8271
  api_keys:
    admin_key: "${{ secrets.FORMATION_ADMIN_API_KEY }}"
    client_key: "${{ secrets.FORMATION_CLIENT_API_KEY }}"

overlord:
  persona: |
    You are a helpful AI assistant.
  response:
    streaming: true  # Only if user enabled streaming

# Only if user enabled async:
async:
  threshold_seconds: 30
  webhook_url: "https://example.com/webhook"

# LLM config based on provider selection:
llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"  # varies by provider
  models:
    - text: "openai/gpt-4o"  # varies by provider

scheduler:
  timezone: "UTC"

# ─────────────────────────────────────────────────────────────
# Additional configuration (uncomment/edit as needed)
# ─────────────────────────────────────────────────────────────

# Add more LLM providers (use 'muxi config llm' for guided setup):
# llm:
#   api_keys:
#     anthropic: "${{ secrets.ANTHROPIC_API_KEY }}"

# Persistent memory (use 'muxi config memory' for guided setup):
# memory:
#   persistent:
#     connection_string: "postgres://user:pass@host:5432/db"

# Logging streams (use 'muxi config logging' for guided setup):
# logging:
#   streams:
#     - transport: "stdout"
#       level: "info"

# Input limits (defaults shown):
# input_limits:
#   max_message_length: 100000
#   max_file_size_bytes: 52428800

# Runtime settings:
# runtime:
#   built_in_mcps: true

# User credentials mode:
# user_credentials:
#   mode: "redirect"
```

---

## Validation
- Webhook URL: Auto-prepend https:// if missing protocol
- API keys: Basic format validation (sk- for OpenAI, etc.)
- Provider name: Lowercase, alphanumeric

## Secrets File Update
The secrets template should include the selected provider:
```
FORMATION_ADMIN_API_KEY=
FORMATION_CLIENT_API_KEY=
OPENAI_API_KEY=          # or ANTHROPIC_API_KEY, etc.
```

## Open Questions
1. Should we validate API keys by making a test request?
2. For "Local" - should we verify the server is reachable?
