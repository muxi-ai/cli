# Plan: `muxi config llm`

## Overview
Configure LLM providers, models, and settings in formation.yaml.

## Command
```bash
muxi config llm
```

## Banner
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure LLM                                       MUXI │
│──────────────────────────────────────────────────────────────│
│ ℹ Configure language models for different capabilities:      │
│ text, vision, audio, documents, embeddings, and streaming.   │
╰──────────────────────────────────────────────────────────────╯
```

## Flow

### Step 1: What to Configure
```
What would you like to configure?
  ○ Add/update API key for a provider
  ○ Configure model for a capability
  ○ Global LLM settings (temperature, tokens, caching)
```

---

### Flow 1: Add/Update Provider API Key
```
Select provider:
  ○ OpenAI [sk-...7x3Q]
  ○ Anthropic [sk-...2nK1]

  ○ Google (Gemini)
  ○ Cohere
  ○ Mistral
  ○ Groq
  ○ xAI (Grok)
  ...
```

- Providers with existing keys shown at top with masked preview
- Line break separates configured from unconfigured
- Selecting a configured provider allows updating the key
- Full provider list matches `muxi new formation` command


**Output:**
```yaml
llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"
    anthropic: "${{ secrets.ANTHROPIC_API_KEY }}"
    mistral: "${{ secrets.MISTRAL_API_KEY }}"
```

---

### Flow 2: Configure Model Capability
```
ℹ Capabilities without explicit models will use the text model.

Select capability to configure:
  ○ Text (main reasoning model) [openai/gpt-4o]
  ○ Vision (image understanding) [using text]
  ○ Audio (speech-to-text) [using text]
  ○ Documents (PDF/doc processing) [using text]
  ○ Embedding (vector search)
  ○ Streaming (progress updates) [using text]
```

- Configured capabilities show current model
- Unconfigured capabilities show "[using text]" to indicate default behavior

#### Text Model:
```
Text Model
──────────

Common models:
  ○ openai/gpt-5
  ○ openai/gpt-5-mini
  ○ anthropic/claude-sonnet-4-5
  ○ anthropic/claude-haiku-4.5
  ○ google/gemini-2.5-flash
  ○ Other (enter)

Or type model: _
```

- Select from list OR type model string directly (e.g., "mistral/mistral-large")

```
Configure settings for this model? (y/N): y

  Temperature (0.0-1.0) [0.7]: _
  Max tokens [1000]: _
  Timeout (seconds) [30]: _
  Max retries [3]: _
  Fallback model (optional): anthropic/claude-sonnet-4-5
```

#### Vision Model:
```
Vision Model
────────────

Common models:
  ○ google/gemini-2.5-flash
  ○ anthropic/claude-sonnet-4-5
  ○ openai/gpt-5
  ○ Other (enter)

Or type model: _

Image settings:
  Max image size (MB) [5]: _
  Enable preprocessing? (Y/n): _
  Max width [1024]: _
  Max height [1024]: _
```

#### Audio Model:
```
Audio Model
───────────

Common models:
  ○ openai/whisper-1
  ○ Other (enter)

Or type model: _

Settings:
  Max audio size (MB) [10]: _
  Language ["auto"]: _
```

#### Documents Model:
```
Documents Model
───────────────

ℹ Used for text extraction from PDFs and documents.

Common models:
  ○ openai/gpt-5
  ○ anthropic/claude-sonnet-4-5
  ○ google/gemini-2.5-flash
  ○ Other (enter)

Or type model: _

Settings:
  Max document size (MB) [20]: _
  Chunk size [1000]: _
  Chunk overlap [100]: _
  Extraction strategy:
    ○ adaptive (recommended)
    ○ fixed
    ○ semantic

  Cache TTL (seconds) [3600]: _
```

#### Embedding Model:
```
Embedding Model
───────────────

Common models:
  ○ openai/text-embedding-3-large
  ○ openai/text-embedding-3-small
  ○ cohere/embed-english-v3.0
  ○ Other (enter)

Or type model: _

Fallback model (optional): _
```

#### Streaming Model:
```
Streaming Model
───────────────

ℹ Used for real-time progress updates. Recommend a fast, cheap model.

Common models:
  ○ openai/gpt-5-mini
  ○ anthropic/claude-sonnet-4-5
  ○ google/gemini-2.5-flash
  ○ Other (enter)

Or type model: _

Settings:
  Temperature [0.7]: _
  Max tokens [100]: _
  Timeout (seconds) [10]: _
```

**Output example:**
```yaml
llm:
  models:
    - text: "openai/gpt-5"
      settings:
        temperature: 0.7
        max_tokens: 1000
        fallback_model: "anthropic/claude-sonnet-4-5"
    - vision: "openai/gpt-5"
      settings:
        image:
          max_size_mb: 5
          preprocessing:
            resize: true
            max_width: 1024
    - embedding: "openai/text-embedding-3-large"
    - streaming: "openai/gpt-4o-mini"
```

---

### Flow 3: Global LLM Settings
```
Global LLM Settings
───────────────────

These are defaults applied to all models unless overridden.

Temperature (0.0-1.0) [0.7]: _
Max tokens [1000]: _
Timeout (seconds) [30]: _
Max retries [3]: _
Default fallback model: _

Response Caching
────────────────
Enable response caching? (Y/n): _
Max cache entries [10000]: _
Similarity threshold (0.0-1.0) [0.95]: _
Cache TTL (seconds) [86400]: _
```

**Output:**
```yaml
llm:
  settings:
    temperature: 0.7
    max_tokens: 1000
    timeout_seconds: 30
    max_retries: 3
    caching:
      enabled: true
      max_entries: 10000
      p: 0.95
      ttl: 86400
```

## Secrets Created
- `{PROVIDER}_API_KEY` for each configured provider

## Validation
- API key format validation (sk- for OpenAI, etc.)
- Model exists for provider
- Temperature in range 0-1
- Positive integers for tokens/timeout/retries

## Questions
1. Should we offer to test API key validity?
2. Should we show current config before asking for changes?
3. Group all capabilities into one flow, or separate commands?
