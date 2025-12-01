# LLM Configuration Guide

Configure language models for your MUXI formation using `muxi config llm`.

## Quick Start

```bash
cd my-formation
muxi config llm
```

## What You Can Configure

### 1. Configure Model for a Capability

Set which model handles each capability:

| Capability | Purpose | Example Models |
|------------|---------|----------------|
| **Text** | Main reasoning model | `openai/gpt-5`, `anthropic/claude-sonnet-4-5` |
| **Vision** | Image understanding | `google/gemini-2.5-flash`, `openai/gpt-5` |
| **Audio** | Speech-to-text | `openai/whisper-1` |
| **Documents** | PDF/doc processing | `openai/gpt-5`, `anthropic/claude-sonnet-4-5` |
| **Embedding** | Vector search | `openai/text-embedding-3-large` |
| **Streaming** | Progress updates | `openai/gpt-5-mini` (fast, cheap) |

**Defaults:** Vision, audio, documents, and streaming fall back to the text model if not configured. Embedding uses a local model (`all-MiniLM-L6-v2`) if not configured.

### 2. Add/Update Provider API Key

Add API keys for LLM providers. The wizard:
- Shows configured providers at the top with masked key preview
- Detects environment variables (e.g., `OPENAI_API_KEY`, `GEMINI_API_KEY`) and offers to copy them
- Saves keys to encrypted secrets store
- Updates `formation.yaml` with secret references

### 3. Global LLM Settings

Configure defaults applied to all models:
- **Temperature** (0.0-1.0): Controls randomness
- **Max tokens**: Maximum response length
- **Timeout**: Request timeout in seconds
- **Max retries**: Retry attempts on failure
- **Fallback model**: Model to use if primary fails
- **Response caching**: Cache similar requests for cost savings

## Model Settings

When configuring a capability, you can set:

### Common Settings (All Capabilities)
```yaml
settings:
  temperature: 0.7
  max_tokens: 4096
  timeout_seconds: 30
  max_retries: 3
  fallback_model: "anthropic/claude-sonnet-4-5"
```

### Vision Settings
```yaml
settings:
  image:
    max_size_mb: 5
    preprocessing:
      resize: true
      max_width: 1024
      max_height: 1024
```

### Audio Settings
```yaml
settings:
  max_size_mb: 10
  language: "auto"  # or "en", "es", "fr", etc.
```

### Video Settings
```yaml
settings:
  max_size_mb: 100
  max_duration_seconds: 300
  include_audio_analysis: true
```

### Documents Settings
```yaml
settings:
  max_size_mb: 20
  extraction:
    chunk_size: 1000
    overlap: 100
    strategy: "adaptive"  # adaptive, semantic, fixed, paragraph
  cache_ttl_seconds: 3600
```

## Example Output

After running `muxi config llm`, your `formation.yaml` will include:

```yaml
llm:
  api_keys:
    openai: "${{ secrets.OPENAI_API_KEY }}"
    anthropic: "${{ secrets.ANTHROPIC_API_KEY }}"
    google: "${{ secrets.GOOGLE_API_KEY }}"
  
  settings:
    temperature: 0.7
    max_tokens: 4096
    timeout_seconds: 30
    max_retries: 3
    fallback_model: "openai/gpt-5-mini"
    caching:
      enabled: true
      max_entries: 10000
      similarity: 0.95
      ttl: 86400
  
  models:
    - text: "openai/gpt-5"
      settings:
        temperature: 0.7
        max_tokens: 4096
        fallback_model: "anthropic/claude-sonnet-4-5"
    
    - vision: "google/gemini-2.5-flash"
      settings:
        fallback_model: "anthropic/claude-sonnet-4-5"
        image:
          max_size_mb: 5
          preprocessing:
            resize: true
            max_width: 1024
            max_height: 1024
    
    - audio: "openai/whisper-1"
      settings:
        max_size_mb: 10
        language: "auto"
    
    - documents: "openai/gpt-5"
      settings:
        max_size_mb: 20
        extraction:
          chunk_size: 1000
          overlap: 100
          strategy: "adaptive"
        cache_ttl_seconds: 3600
    
    - embedding: "openai/text-embedding-3-large"
    
    - streaming: "openai/gpt-5-mini"
```

## Model Order

Models are automatically ordered in formation.yaml:
1. Text
2. Vision
3. Audio
4. Documents
5. Embedding
6. Streaming

## Supported Providers

| Provider | Env Var | Key Prefix |
|----------|---------|------------|
| OpenAI | `OPENAI_API_KEY` | `sk-` |
| Anthropic | `ANTHROPIC_API_KEY` | `sk-ant-` |
| Google | `GOOGLE_API_KEY` or `GEMINI_API_KEY` | - |
| Mistral | `MISTRAL_API_KEY` | - |
| Groq | `GROQ_API_KEY` | `gsk_` |
| xAI | `XAI_API_KEY` | `xai-` |
| DeepSeek | `DEEPSEEK_API_KEY` | `sk-` |
| Cohere | `COHERE_API_KEY` | - |
| Together | `TOGETHER_API_KEY` | - |
| Fireworks | `FIREWORKS_API_KEY` | - |
| Perplexity | `PERPLEXITY_API_KEY` | `pplx-` |
| OpenRouter | `OPENROUTER_API_KEY` | `sk-or-` |

## Claude Model Names

For Anthropic Claude models, use the 4.x naming:
- `anthropic/claude-sonnet-4-5`
- `anthropic/claude-haiku-4-5`
- `anthropic/claude-opus-4-5`

(Claude 3.x models are deprecated)

## Tips

1. **Start with text model** - Other capabilities default to it
2. **Use environment variables** - The wizard detects and offers to import them
3. **Configure fallbacks** - Set fallback models for reliability
4. **Enable caching** - Reduces API costs by 70%+ for repeated queries
5. **Use fast models for streaming** - `gpt-5-mini` or `claude-haiku-4-5` for progress updates

## Related

- [Secrets Guide](secrets.md) - Managing API keys
- [Formations Guide](formations.md) - Formation structure
