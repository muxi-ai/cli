# MUXI CLI - Terminal UI Design System

**Version:** 1.0  
**Date:** 2025-11-26

---

## Philosophy

**Goals:**
- Clean, professional appearance
- Consistent visual language
- Clear status communication
- Readable at a glance
- No emojis (except status symbols)
- Help text dimmed (~80% opacity) so commands stand out

---

## Color Palette

```
Success:  Green   (#00D800 / \033[32m / color.FgGreen)
Error:    Red     (#FF0000 / \033[31m / color.FgRed)
Warning:  Yellow  (#FFFF00 / \033[33m / color.FgYellow)
Info:     Blue    (#0080FF / \033[34m / color.FgBlue)
Skipped:  Cyan    (#00FFFF / \033[36m / color.FgCyan)
Dimmed:   Gray    (#808080 / \033[90m / color.Faint)
Bold:     \033[1m / color.Bold
```

---

## Status Symbols

### Core Symbols

| Symbol | Meaning | Color | Weight | Usage |
|--------|---------|-------|--------|-------|
| ✓ | Success | Green | Bold | Completed actions, valid input |
| ✗ | Error | Red | Bold | Failed actions, invalid input |
| ⚠ | Warning | Yellow | Bold | Warnings, destructive actions |
| ℹ | Info | Blue | Normal | Informational messages |
| ⊘ | Skipped | Cyan | Normal | Optional fields skipped |
| ● | In Progress | Blue | Normal | Currently processing |

### Usage Examples

```
✓ Formation 'my-bot' created successfully!
✗ Formation directory 'my-bot' already exists
⚠ You are operating on formation 'my-bot' (production)
ℹ Using default profile: localhost
⊘ Description: skipped
● Uploading formation...
```

---

## Interactive Prompts (Next.js Style)

### Standard Prompt

**Before input:**
```
Formation ID: _
```

**After valid input:**
```
✓ Formation ID: my-bot
```

**After invalid input (loop):**
```
✗ Formation ID: My-Bot

  IDs must be lowercase and start with a letter

Formation ID: _
```

### Optional Prompt

**Before input:**
```
Description (optional, press Enter to skip): _
```

**After skipping:**
```
⊘ Description: skipped
```

**After providing:**
```
✓ Description: My awesome chatbot
```

### Password Prompt

**Before input:**
```
Setup secrets:
  [1/1] OPENAI_API_KEY (optional)
    Enter API key (leave empty to skip): _
```

**After input:**
```
✓ OPENAI_API_KEY: configured
```

**After skip:**
```
⊘ OPENAI_API_KEY: skipped
```

---

## Success Messages

### Two Types

**1. Progress Steps** - Only icon colored, message normal:
```
✓ Directory structure created
✓ Formation keys generated
```
Use `ui.Step()` for these.

**2. Final Success** - Full line colored (green, bold):
```
✓ Formation 'my-bot' created successfully!
```
Use `ui.Success()` for these.

### Format

```
✓ Operation completed successfully! (green, bold - full line)

Summary or next steps (dimmed text ~80%)
```

### Examples

**Progress steps (icon only colored):**
```
Creating formation 'my-bot'...
✓ Directory structure created
✓ Formation keys generated
✓ Secrets configured
```

**Final success (full line colored):**
```
✓ Formation 'my-bot' created successfully!
```

**Success with next steps:**
```
✓ Formation 'my-bot' created successfully!

Next steps:
  cd my-bot
  muxi secrets set OPENAI_API_KEY
  muxi deploy --profile production
```

**Success with details:**
```
✓ Deployed to 3/3 servers successfully

Summary:
  • us-east-1: deployed (v1.2.0)
  • us-west-1: deployed (v1.2.0)
  • eu-central-1: deployed (v1.2.0)
```

---

## Error Messages

### Format

```
✗ ERROR_TITLE (red, bold)

  Error details here (normal, red text)
  More context or explanation
  
  Suggestion or fix:
    command to run (dimmed)
```

### Examples

**Simple error:**
```
✗ Formation not found

  Formation 'my-bot' does not exist on server
```

**Error with suggestion:**
```
✗ Formation directory exists

  Directory 'my-bot' already exists
  
  Choose a different name or remove:
    rm -rf my-bot
```

**Error with multiple issues:**
```
✗ Validation failed

  Found 3 errors in formation.yaml:
    • Line 12: Missing required field 'overlord.persona'
    • Line 25: Invalid model format 'gpt4' (use 'openai/gpt-4')
    • Line 40: Undefined secret reference 'DATABASE_URL'
  
  Fix these errors and try again:
    vim formation.yaml
```

**Error with context:**
```
✗ Cannot delete secret 'OPENAI_API_KEY'

  This secret is referenced in:
    • formation.yaml (line 52)
    • agents/research-agent.yaml (line 18)
  
  Remove references first, then delete
```

---

## Warning Messages

### Format

```
⚠ WARNING_TITLE (yellow, bold)

  Warning details (normal text)
  
  Confirmation or action
```

### Examples

**Simple warning:**
```
⚠ This is a destructive operation

  Formation 'my-bot' will be permanently deleted
  
  Are you sure? [y/N]: _
```

**Warning with details:**
```
⚠ secrets.enc is out of sync

  You added a new secret but secrets template is missing it
  
  Add 'NEW_SECRET_KEY' to secrets? [Y/n]: _
```

---

## Info Messages

### Format

```
ℹ Info message (blue, normal)

  Details if needed (normal text)
```

### Examples

```
ℹ Using default profile: localhost

ℹ No server profiles configured
  Add one with: muxi profile add production
```

---

## Progress Indicators

### Spinner (Single Operation)

```
● Validating formation...
● Creating bundle...
● Uploading to production...
✓ Formation 'my-bot' deployed
```

**Spinner Frames:**
```
⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
```

### Multi-Step Progress

```
[1/3] Validating formation...
  ✓ formation.yaml valid
  ✓ All secrets referenced
  ✓ Agent configs valid

[2/3] Creating bundle...
  ✓ Bundle created (2.3 MB)

[3/3] Uploading to production...
  ● Uploading...
```

### Progress Bar

```
Uploading to production...
████████████████────────────── 60% | 1.4 MB / 2.3 MB | 1.2 MB/s
```

---

## Lists

### Bullet Lists

```
Files created:
  • .gitignore, .key, .muxi
  • formation.yaml, secrets
  • README.md
  • 6 directories
```

**Note:** Use dimmed bullet (•) with normal text

### Status Lists

```
Found 5 secret references:
  ✓ FORMATION_ADMIN_API_KEY (set)
  ✓ FORMATION_CLIENT_API_KEY (set)
  ✗ OPENAI_API_KEY (not set)
  ✗ DATABASE_URL (not set)
  ⊘ SLACK_WEBHOOK (optional)
```

### Numbered Lists

```
Next steps:
  1. Configure secrets
  2. Validate formation
  3. Deploy to production
```

---

## Help Text

### Inline Help (Dimmed)

```
Formation ID: my-bot

  The formation ID becomes the directory name and system identifier
  Must be lowercase, start with a letter, 3-50 characters
  (entire help block is dimmed ~80%)
```

### Error Context (Normal)

```
✗ Invalid formation ID 'My-Bot'

  Formation IDs must:
    • Be lowercase
    • Start with a letter
    • Contain only letters, numbers, and hyphens
    • Be 3-50 characters long
  
  Example: my-bot
```

---

## Confirmation Prompts

### Yes/No

```
Delete formation 'my-bot'? [y/N]: _
```

**Defaults:**
- `[Y/n]` - Default is yes (capital Y)
- `[y/N]` - Default is no (capital N)

### With Warning

```
⚠ You are operating on formation 'my-bot' (production)

  Are you sure you want to delete agent 'weather'? [y/N]: _
```

---

## Tables

### Simple Table

```
ID       STATUS   MODEL          LAST ACTIVE
weather  active   gpt-4o         2m ago
support  active   gpt-4o-mini    5m ago
research inactive gpt-4          1d ago
```

### Table with Status Icons

```
FORMATION  STATUS  SERVER         UPTIME
my-bot     ✓       us-east-1      5d 12h
support    ✓       us-west-1      2d 8h
analytics  ✗       eu-central-1   stopped
```

---

## Code & Commands

### Inline Commands

```
Run: muxi deploy (bright/bold)
```

### Code Blocks

```
  $ cd my-formation
  $ muxi secrets set OPENAI_API_KEY
  $ muxi deploy --profile production
```

### File Paths

```
Created: agents/weather.yaml
Updated: formation.yaml (line 25)
```

---

## Multi-Select Prompts

### Format

```
Select secrets to configure:
  ◉ OPENAI_API_KEY
  ◯ DATABASE_URL
  ◉ SLACK_WEBHOOK
  
Use arrow keys to move, space to select, enter to confirm
```

**Symbols:**
- `◉` Selected
- `◯` Unselected

---

## Tree Views

### File Structure

```
my-formation/
├── agents/
│   ├── weather.yaml
│   └── support.yaml
├── mcps/
│   └── postgres.yaml
├── formation.yaml
└── README.md
```

### Hierarchical Data

```
formation 'my-bot'
├─ overlord (active)
├─ agents
│  ├─ weather (active)
│  └─ support (inactive)
└─ mcps
   └─ postgres (active)
```

---

## Section Headers

### Major Sections

```
Create new formation:
  Field: value
  Field: value
```

### Sub-Sections

```
Setup secrets:
  [1/3] OPENAI_API_KEY
    Enter value: _
```

---

## Indentation Standards

**Consistent 2-space indentation:**

```
✓ Formation created

  Files created:
    • file1
    • file2
  
  Next steps:
    cd my-formation
```

---

## Design Tokens

### Spacing

```
Section spacing: 1 blank line between sections
List item spacing: No blank lines between items
Error detail spacing: 1 blank line before suggestions
```

### Width

```
Max line width: 80 characters (wrap longer text)
Indentation: 2 spaces per level
Code blocks: Indent by 2 spaces
```

---

## Implementation Guidelines

### Helper Functions

All formatting should use helper functions from `pkg/ui`:

```go
// Status messages
ui.Success("Formation created")          // Final success (full line colored)
ui.Step("Directory structure created")   // Progress step (icon only colored)
ui.Error("Formation not found")
ui.Warning("This is destructive")
ui.Info("Using default profile")
ui.Skipped("Optional field skipped")
ui.InProgress("Processing...")

// Formatting
ui.Dimmed("Help text here")
ui.Bold("Important text")

// Prompts
ui.PromptSuccess("Formation name", "my-bot")
ui.PromptError("Formation name", "My-Bot", err)
ui.PromptSkipped("Description")

// Lists
ui.List(items)
ui.StatusList(items)

// Blocks
ui.ErrorBlock(title, details, suggestion)
ui.SuccessBlock(title, nextSteps)
ui.Section("Setup secrets:")
```

### Consistency Rules

1. **Always use helper functions** - Never manually format colors
2. **One blank line between sections** - Consistent spacing
3. **Status symbols at start** - Always begin status lines with symbol
4. **Dimmed help text** - Help/context is always dimmed
5. **Bold for emphasis** - Use sparingly for key information
6. **Indent suggestions** - Commands/paths indented 2 spaces

---

## Testing Display

### Test Command

```bash
muxi ui test
```

Shows all UI patterns for visual verification:
- Success messages
- Error messages
- Warnings
- Lists
- Progress indicators
- Prompts
- Tables

---

## Examples by Command

### `muxi new formation`

```
✓ Formation ID: my-bot
✓ Description: My awesome bot

Creating formation 'my-bot'...
✓ Directory structure created
✓ Formation keys generated

Setup secrets:
  [1/1] OPENAI_API_KEY (optional)
    Enter API key (leave empty to skip): ********
✓   OPENAI_API_KEY: configured

✓ Formation 'my-bot' created successfully!

Next steps:
  cd my-bot
  muxi validate
  muxi deploy --profile production
```

### `muxi deploy`

```
● Validating formation...
✓ Formation validated

● Creating bundle...
✓ Bundle created (2.3 MB)

Deploying to profile 'production' (3 servers)...

[1/3] us-east-1 (https://east.company.com:7890)
  ● Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

[2/3] us-west-1 (https://west.company.com:7890)
  ● Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

[3/3] eu-central-1 (https://eu.company.com:7890)
  ● Uploading formation (2.3 MB)...
  ✓ Formation 'my-bot' deployed

✓ Deployed to 3/3 servers successfully
```

### `muxi agent list`

```
ℹ Using formation: my-bot (production)

ID       STATUS   MODEL          LAST ACTIVE
weather  active   gpt-4o         2m ago
support  active   gpt-4o-mini    5m ago
research inactive gpt-4          1d ago

3 agents (2 active, 1 inactive)
```

---

## Input History & Line Editing

All text input prompts support **history navigation** and **line editing** for a better user experience.

### Features

**History Navigation:**
- **↑ (Up Arrow):** Previous input
- **↓ (Down Arrow):** Next input

**Line Editing:**
- **← / →:** Navigate within line
- **Ctrl+A / Ctrl+E:** Jump to start/end
- **Ctrl+W:** Delete word
- **Ctrl+U / Ctrl+K:** Clear line

**Session History:**
- All valid inputs saved to session history (up to 100 entries)
- History persists across prompts in same session
- Invalid inputs (validation errors) not saved

### Example Workflow

```bash
# First attempt with error
Endpoint URL: https://api.example.
✗ Hostname cannot end with a dot

# Press ↑ to recall, edit, and fix
Endpoint URL: https://api.example.com
✓ Endpoint URL: https://api.example.com

# Later in session, reuse previous input
Endpoint URL: [Press ↑]
Endpoint URL: https://api.example.com [Edit as needed]
```

**See:** [INPUT-HISTORY.md](./INPUT-HISTORY.md) for complete documentation

---

## Version History

**1.0 (2025-11-26):**
- Initial design system
- Core symbols and colors defined
- Interactive prompt patterns (Next.js style)
- Error, success, warning formats
- List and table layouts
- Progress indicators
- Confirmation prompts

---

**Status:** Complete design system - ready for implementation
**Next:** Create `pkg/ui` package with helper functions
