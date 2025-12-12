# Config Commands Development Guide

This document covers the implementation patterns for `muxi config *` commands.

## Available Commands

| Command | Description | File |
|---------|-------------|------|
| `muxi config a2a` | Agent-to-Agent communication | `components.go` |
| `muxi config async` | Async response settings | `async.go` |
| `muxi config llm` | LLM providers and models | `llm.go` |
| `muxi config logging` | Logging streams | `logging.go` |
| `muxi config memory` | Memory configuration | `memory.go` |
| `muxi config overlord` | Overlord behavior | `overlord.go` |
| `muxi config security` | User credential handling | `security.go` |

---

## Remote Config (Formation API)

Config commands support a `--remote` flag to fetch configuration from a running Formation. This is **read-only** and useful for inspecting deployed formations.

To change configuration, update your `formation.afs` locally and deploy a new version using `muxi deploy`.

### Commands with --remote Support

```bash
# Full formation config
muxi config --remote
muxi config --remote -o json    # JSON output

# Specific sections
muxi config llm --remote        # GET /llm/settings
muxi config memory --remote     # GET /memory
muxi config overlord --remote   # GET /overlord

# With formation/profile flags
muxi config llm --remote -F my-formation -p production
```

### Output Formats

```bash
muxi config --remote              # Default: YAML
muxi config --remote -o yaml      # Explicit YAML
muxi config --remote -o json      # JSON output
```

### Implementation Pattern

For commands that support both local (interactive wizard) and remote modes:

```go
var configLLMCmd = &cobra.Command{
    Use:   "llm",
    Short: "Configure LLM providers and models",
    RunE:  runConfigLLM,
}

func runConfigLLM(cmd *cobra.Command, args []string) error {
    remote, _ := cmd.Flags().GetBool("remote")

    if remote {
        client, err := formation.ClientFromFlags(cmd)
        if err != nil {
            return err
        }

        llmSettings, err := client.GetLLMSettings()
        if err != nil {
            return fmt.Errorf("failed to get LLM settings: %w", err)
        }

        output, _ := cmd.Flags().GetString("output")
        return printConfigOutput(llmSettings, output)
    }

    // Local: run interactive wizard
    return scaffold.ConfigureLLM()
}

func init() {
    configLLMCmd.Flags().Bool("remote", false, "Fetch from Formation API")
    configLLMCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
    formation.AddCommonFlags(configLLMCmd)
}
```

### Flags Added for Remote Mode

| Flag | Short | Description |
|------|-------|-------------|
| `--remote` | | Fetch config from Formation API (read-only) |
| `--output` | `-o` | Output format: `yaml` (default) or `json` |
| `--formation` | `-F` | Formation ID |
| `--profile` | `-p` | Server profile |

---

## YAML Formatting Rules

All config commands follow consistent YAML formatting via helpers in `yaml.go`:

### 1. Two-Space Indentation

Use `marshalYAML()` instead of `yaml.Marshal()`:

```go
output, err := marshalYAML(&root)  // Uses 2-space indent
```

### 2. Blank Lines Before Top-Level Keys

```yaml
llm:
  settings:
    temperature: 0.7

memory:        # ← blank line before top-level key
  working:
    mode: "local"
```

### 3. Blank Lines Before Second-Level Keys with Children

```yaml
server:
  host: "0.0.0.0"     # NO blank line (simple value)
  port: 8271          # NO blank line (simple value)

  api_keys:           # YES blank line (has children)
    admin_key: "..."
    client_key: "..."
```

### 4. Implementation

Call these helpers after writing YAML:

```go
content, err := os.ReadFile(formationPath)
if err == nil {
    cleaned := removeCommentedSection(string(content), "logging")
    cleaned = cleanupAdditionalConfigSection(cleaned)
    cleaned = ensureBlankLineBeforeTopLevel(cleaned)
    os.WriteFile(formationPath, []byte(cleaned), 0644)
}
```

---

## Commented Sections Cleanup

When adding real config, remove the corresponding commented template:

| Config Added | Commented Section Removed |
|--------------|--------------------------|
| `logging:` | `# logging:` block |
| `memory:` | `# memory:` block |
| `user_credentials:` | `# user_credentials:` block |

The "Additional configuration" section always stays at the bottom of the file.

---

## Password Masking

All sensitive inputs use `PromptPassword()` for hidden input:

```go
// Hidden input (shows nothing while typing)
token, err := wizard.PromptPassword("Bearer token", false)

// Show masked value in success message
ui.PromptSuccess("Token", "********")
```

### Sensitive Fields

| Field Type | Secret Name Pattern |
|------------|---------------------|
| Bearer token | `*_BEARER_TOKEN` |
| API key | `*_API_KEY` |
| Password | `*_PASSWORD` |
| Encryption key | `*_ENCRYPTION_KEY` |

---

## Secrets Pattern

Secrets are stored in `.muxi/secrets.enc` and referenced in YAML:

```yaml
auth:
  token: "${{ secrets.LOGGING_BEARER_TOKEN }}"
```

### Creating Secrets

```go
sm := secrets.NewManager(rootDir)
if err := sm.Set("LOGGING_BEARER_TOKEN", token, true); err != nil {
    return fmt.Errorf("failed to save token: %w", err)
}
```

---

## Wizard Patterns

### Banner

```go
ui.Banner(`╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure Logging                                   MUXI │
│──────────────────────────────────────────────────────────────│
│ Configure where logs and events are sent.                    │
╰──────────────────────────────────────────────────────────────╯`)
```

### Select with Current Value Indicator

```go
options := []wizard.SelectOption{
    {Value: "option1", Label: "Option 1"},
    {Value: "option2", Label: "Option 2 [current]"},  // Mark current
}
```

### Flow Selection

```go
options := []wizard.SelectOption{
    {Value: "add", Label: "Add a new item"},
    {Value: "view", Label: "View/edit current items"},
    {Value: "remove", Label: "Remove an item"},
}
action, _ := wizard.PromptSelect("What would you like to do?", options, 0)
```

---

## URL Validation

Use consistent URL validation with auto-prefix:

```go
destination := normalizeURL(input)      // Adds https:// if missing
if err := validateURL(destination); err != nil {
    ui.PromptError("URL", input, err)
    continue
}
```

---

## File Structure

```
src/pkg/scaffold/
├── yaml.go          # YAML formatting helpers
├── async.go         # muxi config async
├── logging.go       # muxi config logging
├── security.go      # muxi config security  
├── memory.go        # muxi config memory
├── overlord.go      # muxi config overlord
├── llm.go           # muxi config llm
├── components.go    # muxi config a2a (and new agent/mcp/etc)
└── templates.go     # Formation template with "Additional config" section
```

---

## Adding a New Config Command

1. Create `src/pkg/scaffold/{name}.go` with `Configure{Name}()` function
2. Add command in `src/cmd/config.go`:
   ```go
   var config{Name}Cmd = &cobra.Command{
       Use:   "{name}",
       Short: "Configure {name}",
       RunE: func(cmd *cobra.Command, args []string) error {
           return scaffold.Configure{Name}()
       },
   }
   ```
3. Register in `init()`: `configCmd.AddCommand(config{Name}Cmd)`
4. Use YAML cleanup helpers after writing
5. Create user guide in `docs/guides/{name}.md`

---

## Testing

```bash
cd src && go build -o muxi .
cd ../playground/test-formation
../../src/muxi config logging
```
