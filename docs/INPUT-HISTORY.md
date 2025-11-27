# Input History & Line Editing

The MUXI CLI provides a powerful input experience with history navigation and line editing capabilities, powered by [readline](https://github.com/chzyer/readline).

## Features

### 🔄 History Navigation

Navigate through your previous inputs using arrow keys:

- **Up Arrow (↑)**: Go to previous input
- **Down Arrow (↓)**: Go to next input (if you went back)

**Example workflow:**
```
Endpoint URL: https://api.example.
✗ Hostname cannot end with a dot

Endpoint URL: [Press ↑]
Endpoint URL: https://api.example.  [Edit to add 'com']
Endpoint URL: https://api.example.com
✓ Endpoint URL: https://api.example.com
```

### ✏️ Line Editing

Edit your input efficiently with keyboard shortcuts:

| Shortcut | Action |
|----------|--------|
| **Left Arrow (←)** | Move cursor left |
| **Right Arrow (→)** | Move cursor right |
| **Ctrl+A** | Jump to start of line |
| **Ctrl+E** | Jump to end of line |
| **Ctrl+W** | Delete word before cursor |
| **Ctrl+U** | Clear line before cursor |
| **Ctrl+K** | Clear line after cursor |
| **Ctrl+L** | Clear screen |
| **Backspace** | Delete character before cursor |
| **Delete** | Delete character under cursor |

### 📚 Session History

History persists throughout your CLI session:

- All valid inputs are saved to session history
- History limit: 100 entries (oldest entries removed automatically)
- History is shared across all prompts in the same session
- Invalid inputs (validation errors) are **not** saved to history

**Example:**
```bash
# First MCP
muxi new mcp weather-api
Endpoint URL: https://api.weather.com
✓ Endpoint URL: https://api.weather.com

# Second MCP - reuse previous input
muxi new mcp forecast-api
Endpoint URL: [Press ↑]
Endpoint URL: https://api.weather.com  [Edit domain]
```

## Implementation Details

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ wizard.PromptString()                                       │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Try: readline.NewEx() (with history)               │    │
│  │   ↓ Success                                         │    │
│  │   • Load session history                            │    │
│  │   • Enable arrow keys & line editing                │    │
│  │   • Handle validation errors                        │    │
│  │   • Save valid inputs to history                    │    │
│  └────────────────────────────────────────────────────┘    │
│                 ↓ Fail                                      │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Fallback: promptStringFallback() (basic)           │    │
│  │   • Basic line input with bufio.Reader              │    │
│  │   • No history navigation                           │    │
│  │   • Works everywhere (even without terminal)        │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Fallback Behavior

The CLI automatically falls back to basic input if readline initialization fails:

- **When fallback is used:**
  - Non-interactive terminals (CI/CD, piped input)
  - Unsupported terminal types
  - Terminal control issues

- **Fallback limitations:**
  - No arrow key history navigation
  - Basic line editing only
  - All other features work normally

### History Management

```go
// Session-wide history (in-memory only)
var inputHistory []string

// History lifecycle:
1. User enters valid input → Added to inputHistory
2. User enters invalid input → NOT added to history
3. Next prompt → Previous history loaded into readline
4. Session ends → History cleared (not persisted)
```

**Why not persist history to disk?**
- Privacy: Avoid storing potentially sensitive inputs (URLs, endpoints)
- Simplicity: No file management, permissions, or cleanup needed
- Security: No risk of leaking secrets through history files
- Session-scoped is sufficient for wizard workflows

## Examples

### URL Validation with History

```bash
muxi new mcp weather-api

# First attempt (invalid)
Endpoint URL: https://api.weather.
✗ Hostname cannot end with a dot

# Press ↑ to recall, then edit
Endpoint URL: https://api.weather.com
✓ Endpoint URL: https://api.weather.com

# Later in same session
muxi new mcp forecast-api
Endpoint URL: [Press ↑]
Endpoint URL: https://api.weather.com  [Available!]
```

### Line Editing

```bash
MCP ID: weather-forcast-api
       [Ctrl+←]  # Jump to "forcast"
       [Backspace x3]  # Delete "cas"
       [Type "eca"]  # Insert "eca"
✓ MCP ID: weather-forecast-api
```

### Multi-Input Wizards

```bash
muxi new agent sales-assistant

# All these inputs support history:
Agent ID: sales-assistant
✓ Agent ID: sales-assistant

Name [Sales Assistant]: [Enter]
✓ Name: Sales Assistant

System message: You are a helpful sales assistant...
✓ System message: configured

# Later inputs can use history from earlier in the same session
```

## Technical Notes

### Dependencies

- **Library:** `github.com/chzyer/readline` v1.5.1
- **Purpose:** Terminal input handling with history and editing
- **Size:** ~100KB (compiled)
- **License:** MIT

### Platform Support

| Platform | History Navigation | Line Editing | Fallback |
|----------|-------------------|--------------|----------|
| **Linux** | ✅ Full support | ✅ Full support | ✅ Available |
| **macOS** | ✅ Full support | ✅ Full support | ✅ Available |
| **Windows** | ✅ Full support | ✅ Full support | ✅ Available |
| **CI/CD** | ❌ Fallback | ❌ Fallback | ✅ Works |

### Performance

- **Initialization:** ~1ms per prompt
- **History lookup:** O(n) where n = history size (max 100)
- **Memory overhead:** ~10KB for 100 history entries
- **No file I/O:** History is in-memory only

## Comparison: Before vs After

### Before (Basic Input)

```bash
Endpoint URL: https://asdasd.
✗ Hostname cannot end with a dot

Endpoint URL: [Type entire URL again from scratch] 😞
```

**Problems:**
- ❌ Can't recall previous input
- ❌ Arrow keys print escape sequences (^[[A/B)
- ❌ Must retype entire input
- ❌ Slow workflow for corrections

### After (With Readline)

```bash
Endpoint URL: https://asdasd.
✗ Hostname cannot end with a dot

Endpoint URL: [Press ↑ → cursor at end → type "com"] 🎉
✓ Endpoint URL: https://asdasd.com
```

**Benefits:**
- ✅ Recall previous input with ↑
- ✅ Navigate with ← → to edit
- ✅ Quick corrections
- ✅ Fast workflow
- ✅ Professional CLI experience

## See Also

- [TUI Design System](./TUI-DESIGN.md) - Complete UI/UX patterns
- [Wizard System](../src/pkg/wizard/wizard.go) - Implementation details
- [readline Documentation](https://github.com/chzyer/readline) - Library reference
