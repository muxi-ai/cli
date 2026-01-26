# Wizard Banners

**Date:** 2025-11-27
**Status:** ✅ Implemented
**Purpose:** Contextual banners at the start of each wizard

---

## Overview

All interactive wizards now display contextual banners that explain what the wizard does and provide helpful tips. Banners use Unicode box-drawing characters for a professional appearance.

**Design Principle:** Users should know exactly what they're doing before answering any questions.

---

## Banner Format

### Standard Width
All banners are **64 characters wide** (including borders):
```
╭──────────────────────────────────────────────────────────────╮
│ Content here (62 chars max)                                  │
╰──────────────────────────────────────────────────────────────╯
```

### With Separator
For multi-section banners (action + info/warning):
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Header action                                            │
│──────────────────────────────────────────────────────────────│
│ ⚠ Additional context or warning                              │
│                                                              │
│ Can span multiple lines for detailed explanations.           │
╰──────────────────────────────────────────────────────────────╯
```

---

## Icon Legend

| Icon | Meaning | Usage |
|------|---------|-------|
| `[+]` | Add/Create | Creating new files or resources |
| `[✎]` | Edit | Modifying existing (future) |
| `[⚙]` | Configure | Changing settings |
| `⚠` | Warning | Important notices (no brackets in content) |
| `ℹ` | Info | Helpful context (no brackets in content) |

---

## Implementation

### 1. Formation
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Create new formation                                     │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new formation`
**When:** Interactive mode only (not shown with `--no-wizard`)

---

### 2. Agent

#### Simple (no A2A)
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new agent                                            │
╰──────────────────────────────────────────────────────────────╯
```

#### With A2A Info
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new agent                                            │
│──────────────────────────────────────────────────────────────│
│ This formation has A2A enabled. You can make this agent      │
│ visible externally for Agent-to-Agent communication.         │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new agent`
**Condition:** Checks if `a2a.enabled: true` in `formation.yaml`
**When:** Interactive mode only

---

### 3. MCP

#### Formation-Level
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new MCP to formation                                 │
│──────────────────────────────────────────────────────────────│
│ ⚠ Formation-level MCPs can be used by all agents.            │ (RED)
│                                                              │
│ For tools that are going to be used primarily by a specific  │
│ agent, we recommend adding the MCP on the agent-level:       │
│   $ muxi new mcp --agent <agent-id>                          │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new mcp`
**When:** Interactive mode, no `--agent` flag
**Note:** Warning line is displayed in **red bold** for emphasis

#### Agent-Level
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new MCP to: Weather Assistant                        │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new mcp --agent weather-assistant`
**Note:** Agent name is dynamically inserted (title-cased from agent ID)
**When:** Interactive mode with `--agent` flag

---

### 4. A2A Configuration

#### Generic (no direction flag)
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure A2A communication settings                     │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi config a2a`
**When:** Interactive mode, no `--inbound` or `--outbound` flag

#### Inbound
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure A2A inbound settings                           │
│──────────────────────────────────────────────────────────────│
│ Inbound A2A allows external formations to discover and       │
│ connect to your agents.                                      │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi config a2a --inbound`
**When:** Interactive mode

#### Outbound
```
╭──────────────────────────────────────────────────────────────╮
│ [⚙] Configure A2A outbound settings                          │
│──────────────────────────────────────────────────────────────│
│ Outbound A2A allows your agents to discover and connect      │
│ to external formations.                                      │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi config a2a --outbound`
**When:** Interactive mode
**Status:** Not yet implemented (wizard pending)

---

### 5. SOP
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new SOP (Standard Operating Procedure) document      │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new sop`
**When:** Interactive mode only

---

### 6. Trigger
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new trigger                                          │
│──────────────────────────────────────────────────────────────│
│ Triggers allow external systems to invoke your agents        │
│ via webhooks or scheduled events.                            │
╰──────────────────────────────────────────────────────────────╯
```

**Command:** `muxi new trigger`
**When:** Interactive mode only

---

## Code Implementation

### UI Functions
```go
// src/pkg/ui/ui.go

// Standard banner (plain text)
func Banner(banner string) {
    fmt.Println(banner)
    fmt.Println()
}

// Special banner for formation-level MCP with red warning
func FormationMCPBanner() {
    // Prints banner with warning line in red bold
    fmt.Print("│ ")
    red.Print("⚠ Formation-level MCPs can be used by all agents.")
    fmt.Println("            │")
}
```

**Usage:**
```go
// Simple banner
ui.Banner("╭──────────────────────────────────────────────────────────────╮\n│ [+] Add new agent                                            │\n╰──────────────────────────────────────────────────────────────╯")

// Formation MCP banner with red warning
ui.FormationMCPBanner()
```

### Pattern in Wizards
```go
func CreateAgent(name string, noWizard bool) error {
    // Formation detection first
    ctx, err := context.MustDetectFormation()
    if err != nil {
        // ... error handling
    }

    // Show banner in interactive mode
    if !noWizard {
        if isA2AEnabled(ctx.RootDir) {
            // Show A2A-aware banner
        } else {
            // Show simple banner
        }
    }

    // ... rest of wizard
}
```

**Key Points:**
- Banner shown AFTER formation detection (so we know context)
- Banner shown BEFORE first prompt (user sees what they're doing)
- Banner ONLY in interactive mode (respects `--no-wizard`)
- Banner followed by blank line (visual separation)

---

## Design Decisions

### 1. Why Fixed Width (64 chars)?
- Professional appearance across terminals
- Easy to read and scan
- Matches terminal conventions (80-char standard with margins)
- Consistent with other CLI tools (kubectl, gh, etc.)

### 2. Why Show Only in Interactive Mode?
- `--no-wizard` mode is for automation/scripts
- Banners add no value when output is piped
- Keeps non-interactive output clean

### 3. Why Context-Aware Banners?
- **Formation-level MCP:** Users need to know it's available to ALL agents
- **Agent-level MCP:** Users need to see which agent they're configuring
- **A2A in Agent:** Users should know about external visibility option
- **Inbound vs Outbound:** Different purposes, different explanations

### 4. Why Icons in Brackets vs Standalone?
- **Headers:** `[+]` in brackets - action labels
- **Content:** `⚠` standalone - visual emphasis without clutter
- **Consistency:** Matches common CLI patterns

---

## Testing

### Manual Verification
```bash
# Formation banner
muxi new formation
# Should show: [+] Create new formation

# Agent banner (without A2A)
cd test-formation
muxi new agent
# Should show: [+] Add new agent

# Enable A2A manually
echo "a2a:\n  enabled: true" >> formation.yaml

# Agent banner (with A2A)
muxi new agent agent2
# Should show: [+] Add new agent + A2A info

# Formation-level MCP
muxi new mcp weather
# Should show: [+] Add new MCP to formation + warning

# Agent-level MCP
muxi new mcp --agent agent2 tools
# Should show: [+] Add new MCP to: Agent2

# A2A inbound
muxi config a2a --inbound
# Should show: [⚙] Configure A2A inbound settings + info

# No banner in non-interactive mode
muxi new formation test --no-wizard
# Should NOT show banner
```

---

## Future Enhancements

### 1. Edit Wizards (Future)
When we add edit functionality:
```
╭──────────────────────────────────────────────────────────────╮
│ [✎] Edit agent: Weather Assistant                            │
╰──────────────────────────────────────────────────────────────╯
```

### 2. Dynamic Tips
Based on formation state:
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new agent                                            │
│──────────────────────────────────────────────────────────────│
│ You have 3 formation-level MCPs available:                   │
│   • weather-api, database, web-search                        │
╰──────────────────────────────────────────────────────────────╯
```

### 3. Progress Indicators
For multi-step wizards:
```
╭──────────────────────────────────────────────────────────────╮
│ [+] Add new MCP to formation                     [Step 1/3]  │
│──────────────────────────────────────────────────────────────│
│ ⚠ Formation-level MCPs can be used by all agents.            │
╰──────────────────────────────────────────────────────────────╯
```

---

## Related Documentation

- [tui-design.md](tui-design.md) - Full design system
- [ux-patterns.md](ux-patterns.md) - UX patterns and conventions
