# MUXI CLI - UX Design Patterns

**Last Updated:** 2025-11-27  
**Status:** Living Document

This document captures the UX design patterns, conventions, and best practices established during MUXI CLI development. These patterns ensure a consistent, polished, and user-friendly experience throughout the CLI.

---

## Table of Contents

1. [Validation & Error Handling](#validation--error-handling)
2. [URL & Endpoint Handling](#url--endpoint-handling)
3. [Menu Selections](#menu-selections)
4. [Message Formatting](#message-formatting)
5. [User Input Patterns](#user-input-patterns)
6. [Natural Language](#natural-language)
7. [Ctrl+C Handling](#ctrlc-handling)
8. [Secret Management](#secret-management)
9. [State Management](#state-management)

---

## Validation & Error Handling

### ✅ **Pattern: Validation Loops (Not Exits)**

**Bad (Old Pattern):**
```go
registriesStr, err := wizard.PromptString("Registry URLs", "", nil)
if registriesStr == "" {
    return fmt.Errorf("at least one registry URL is required")
    // User has to restart entire wizard! ❌
}
```

**Good (New Pattern):**
```go
// Loop until valid input
for {
    registriesStr, err := wizard.PromptString("Registry URLs", existingRegistries, nil)
    if err != nil {
        // Ctrl+C - exit gracefully
        fmt.Println()
        ui.Dimmed("Configuration cancelled")
        return nil
    }
    
    if registriesStr == "" {
        ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("at least one registry URL is required"))
        continue  // ← Re-prompt, don't exit!
    }
    
    // Validate...
    if !valid {
        ui.PromptError("Registry URLs", registriesStr, fmt.Errorf("invalid format"))
        continue  // ← Re-prompt, don't exit!
    }
    
    ui.PromptSuccess("Registries", fmt.Sprintf("%d added", len(registries)))
    break  // ← Success!
}
```

**Benefits:**
- ✅ Users can fix typos immediately
- ✅ No need to restart wizard
- ✅ Less frustrating UX
- ✅ Ctrl+C still works to cancel

---

### 📏 **Pattern: Error Message Line Length**

**Rule:** Max 70 characters per detail line

**Bad:**
```go
return fmt.Errorf("the registry URL you provided is not valid because it contains an invalid hostname format")
// ❌ Too long!
```

**Good:**
```go
// Error wrapper splits into multiple lines automatically
ui.PromptError("Registry URLs", input, fmt.Errorf("invalid URL: %s (malformed hostname)", url))

// Output:
// ✗ Registry URLs: example..com
//
//   invalid URL: https://example..com (malformed hostname)
```

**Implementation:**
```go
// In ui.PromptError():
words := strings.Fields(message)
var lines []string
currentLine := "  "

for _, word := range words {
    if len(currentLine)+len(word)+1 > 70 {
        lines = append(lines, currentLine)
        currentLine = "  " + word
    } else {
        if currentLine != "  " {
            currentLine += " "
        }
        currentLine += word
    }
}
```

---

## URL & Endpoint Handling

### 🔗 **Pattern: Auto-Normalize URLs**

**Rule:** Auto-add `https://` if missing, reject `http://`

```go
func normalizeURL(url string) (string, error) {
    url = strings.TrimSpace(url)
    
    // Reject http:// (insecure)
    if strings.HasPrefix(url, "http://") {
        return "", fmt.Errorf("must use https:// (http:// is not secure)")
    }
    
    // Auto-add https:// if missing
    if !strings.HasPrefix(url, "https://") {
        url = "https://" + url
    }
    
    // Validate format
    parsed, err := neturl.Parse(url)
    if err != nil || parsed.Host == "" {
        return "", fmt.Errorf("invalid format")
    }
    
    // Check for malformed hostname
    if strings.Contains(parsed.Host, "..") || strings.HasPrefix(parsed.Host, ".") {
        return "", fmt.Errorf("malformed hostname")
    }
    
    return url, nil
}
```

**User Experience:**
```bash
# Input: registry.com
# Normalized: https://registry.com ✅

# Input: https://registry.com
# Normalized: https://registry.com ✅

# Input: http://registry.com
# Error: must use https:// ❌

# Input: example..com
# Error: malformed hostname ❌
```

---

### 🎯 **Pattern: Flexible Input Parsing**

**Rule:** Accept comma, space, OR newline separated values (but don't advertise it!)

```go
// parseURLList parses comma, space, or line-separated URLs
func parseURLList(input string) []string {
    if input == "" {
        return []string{}
    }

    // Split by comma, space, or newline (flexible input)
    parts := strings.FieldsFunc(input, func(r rune) bool {
        return r == ',' || r == '\n' || r == ' '
    })

    var result []string
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }

    return result
}
```

**Prompt Text:**
```
Registry URLs (comma or line-separated):  ← Keep it simple!
```

**Why hide space-support?**
- Don't overwhelm users with "comma, space, or line-separated"
- They'll naturally try spaces and it'll "just work"
- Feels smart and forgiving

---

## Menu Selections

### 🎨 **Pattern: Green Bold Highlighting**

**Rule:** Selected option in arrow-key menus should be green + bold

```go
func renderOptions(options []Option, selectedIndex int) {
    green := color.New(color.FgGreen, color.Bold)
    
    for i, opt := range options {
        if i == selectedIndex {
            // Selected: green + bold
            fmt.Print("  ")
            green.Print("◉ ")
            green.Println(opt.Label)
        } else {
            // Not selected: dimmed
            fmt.Printf("  \033[2m◯ %s\033[0m\n", opt.Label)
        }
    }
}
```

**Visual Output:**
```
  Authentication (↑↓ to select):
    ◉ API Key               ← Green bold
    ○ Bearer Token          ← Dimmed
    ○ Basic Auth            ← Dimmed
    ○ None                  ← Dimmed
```

---

### 📋 **Pattern: Multi-Line Prompts**

**Rule:** If prompt text exceeds 60 chars, wrap to next line

```go
func PromptString(prompt string, defaultValue string, options []string) (string, error) {
    // Multi-line if > 60 chars
    if len(prompt) > 60 {
        bold := color.New(color.Bold)
        bold.Println(prompt)
        fmt.Print("  ")  // Indent input on next line
    } else {
        // Single line
        bold := color.New(color.Bold)
        bold.Printf("%s: ", prompt)
    }
    
    // ... rest of prompt logic
}
```

**Visual Output:**
```
# Short prompt (single line):
Registry URLs (comma or line-separated): _

# Long prompt (multi-line):
This is a very long prompt that exceeds sixty characters
  _
```

---

## Message Formatting

### 💬 **Pattern: Natural Language**

**Rule:** Use natural, friendly language instead of technical terms

**User-Facing Messages:**
- ✅ `in the formation` (NOT: `in formation.yaml`)
- ✅ `added to the formation` (NOT: `added to formation.yaml`)
- ✅ `to agent 'weather'` (NOT: `to agents/weather.yaml`)

**Error Messages:**
- ✅ `failed to read the formation` (NOT: `failed to read formation.yaml`)
- ✅ `failed to update the formation` (NOT: `failed to update formation.yaml`)

**Success Messages:**
```go
// Bad
ui.Success("A2A inbound configuration added to formation.yaml")

// Good
ui.Success("A2A inbound configuration added to the formation")
```

**Exception:** "Created" messages still show file paths (helpful to know where):
```
✓ Created agents/weather.yaml
✓ Created mcps/weather-api.yaml
```

---

### 🎯 **Pattern: Banner Verbs in "-ing" Form**

**Rule:** Use present continuous for active processes

```
# Creating formation...     ← "-ing" form
# Configuring A2A...        ← "-ing" form
# Adding MCP...             ← "-ing" form
```

**Implementation:**
```go
ui.Banner("Configuring A2A Inbound")  // ✅
ui.Banner("Configure A2A Inbound")    // ❌
```

---

### 🏷️ **Pattern: MUXI Branding**

**Rule:** All banners show "MUXI" in right corner

```
  ╔══════════════════════════════════════════════════════════════ MUXI ══╗
  ║                                                                       ║
  ║                     Configuring A2A Inbound                           ║
  ║                                                                       ║
  ╚═══════════════════════════════════════════════════════════════════════╝
```

**Implementation:**
```go
func Banner(title string) {
    width := 75
    titleLen := len(title)
    padding := (width - titleLen - 2) / 2
    
    brandingWidth := 10  // " MUXI ══╗"
    topBarContent := strings.Repeat("═", width-brandingWidth-3)
    
    fmt.Printf("  ╔%s MUXI ══╗\n", topBarContent)
    fmt.Println("  ║" + strings.Repeat(" ", width) + "║")
    fmt.Printf("  ║%s%s%s║\n",
        strings.Repeat(" ", padding),
        title,
        strings.Repeat(" ", width-padding-titleLen))
    fmt.Println("  ║" + strings.Repeat(" ", width) + "║")
    fmt.Println("  ╚" + strings.Repeat("═", width) + "╝")
    fmt.Println()
}
```

---

## User Input Patterns

### ⌨️ **Pattern: Pre-Fill Existing Values**

**Rule:** When editing, show existing values as defaults

```go
var existingRegistries string
if alreadyConfigured {
    existingRegistries = extractA2ARegistries(content)
}

// Show as default
registriesStr, err := wizard.PromptString(
    "Registry URLs",
    existingRegistries,  // ← Pre-filled!
    nil,
)
```

**User Experience:**
```
Registry URLs [https://old-registry.com]: 
  ← Press Enter to keep existing
  ← Or type new value to replace
```

---

### ⚠️ **Pattern: Confirmation Before Replacing**

**Rule:** Warn + confirm before replacing existing configuration

```go
if alreadyConfigured && !noWizard {
    fmt.Println()
    red := color.New(color.FgRed, color.Bold)
    red.Println("  ⚠ A2A inbound is already configured in the formation")
    fmt.Println()
    
    ui.Dimmed("This will replace the entire A2A inbound configuration.")
    ui.Dimmed("Existing values will be shown as defaults - press Enter to keep them.")
    fmt.Println()
    
    confirm, err := wizard.PromptString("Continue and replace? (y/N)", "", nil)
    if confirm != "y" {
        ui.Dimmed("Configuration cancelled")
        return nil
    }
}
```

---

## Ctrl+C Handling

### 🛑 **Pattern: Graceful Exit at All Prompts**

**Rule:** Check errors from `wizard.PromptString()` and exit gracefully

```go
registriesStr, err := wizard.PromptString("Registry URLs", existingRegistries, nil)
if err != nil {
    // User pressed Ctrl+C
    fmt.Println()
    ui.Dimmed("Configuration cancelled")
    return nil  // ← Clean exit
}
```

**Don't ignore errors:**
```go
// Bad
registriesStr, _ := wizard.PromptString(...)  // ❌ Ctrl+C treated as empty!

// Good
registriesStr, err := wizard.PromptString(...)  // ✅ Ctrl+C exits gracefully
if err != nil {
    fmt.Println()
    ui.Dimmed("Configuration cancelled")
    return nil
}
```

---

## Secret Management

### 🔐 **Pattern: Masked Display**

**Rule:** Show last 4 chars only for secrets

```go
authKey, err := wizard.PromptString("API Key value", "", nil)
if len(authKey) > 4 {
    ui.PromptSuccess("API Key", "***"+authKey[len(authKey)-4:])
} else {
    ui.PromptSuccess("API Key", "***")
}
```

**Output:**
```
API Key value: sk_live_12345678
✓ API Key: ***5678
```

---

### 📝 **Pattern: Secret Placeholders**

**Rule:** Store placeholders in formation, actual values in secrets file

```go
// In formation.yaml
auth:
  type: "api_key"
  key: "${{ secrets.A2A_INBOUND_API_KEY }}"

// In secrets file
A2A_INBOUND_API_KEY=

// User fills this in later (or via secrets command)
```

---

## State Management

### 🔄 **Pattern: Asymmetric Enable/Disable Flow**

**Rule:** Disable exits, Enable continues to wizard

**Rationale:**
- **Disabling** = "turn it off and I'm done" → Exit makes sense
- **Enabling** = "turn it on and let me configure it" → Continue to wizard makes sense

```go
if isEnabled {
    togglePrompt = "Disable inbound A2A? (y/N)"
} else {
    togglePrompt = "Enable inbound A2A? (y/N)"
}

toggle, err := wizard.PromptString(togglePrompt, "", nil)

if toggle == "y" {
    if isEnabled {
        // Disable and EXIT
        disableA2AInbound(rootDir)
        ui.Success("A2A inbound disabled in the formation")
        return nil  // ← Done!
    } else {
        // Enable and CONTINUE to wizard
        enableA2AInbound(rootDir)
        ui.Success("A2A inbound enabled in the formation")
        fmt.Println()
        // Fall through to wizard below ← Continue!
    }
}
```

**User Experience:**

| Current State | User Action | Result |
|---------------|-------------|---------|
| Enabled | Disable? y | Disable and EXIT |
| Enabled | Disable? n | Continue to wizard (keep enabled) |
| Disabled | Enable? y | Enable and CONTINUE to wizard ⭐ |
| Disabled | Enable? n | Continue to wizard (keep disabled) |

**Benefit:** One command to enable + configure!

---

### 📊 **Pattern: Smart State Detection**

**Rule:** Show appropriate warning based on current state

```go
var isEnabled bool
if alreadyConfigured {
    isEnabled = extractA2AInboundEnabled(content)
}

if isEnabled {
    red.Println("  ⚠ A2A inbound is already enabled in the formation")
} else {
    red.Println("  ⚠ A2A inbound is already configured in the formation")
}
```

**User sees:**
- `already enabled` → Implies it's currently active
- `already configured` → Implies it exists but might be disabled

---

## Summary: Quick Reference

### **Validation**
- ✅ Use validation loops, not exits
- ✅ Max 70 chars per error line
- ✅ Re-prompt on error, don't exit

### **URLs**
- ✅ Auto-add `https://`
- ✅ Reject `http://`
- ✅ Accept comma/space/newline (don't advertise space)

### **Selections**
- ✅ Green bold for selected
- ✅ Multi-line prompts >60 chars

### **Language**
- ✅ "the formation" not "formation.yaml"
- ✅ "-ing" verbs in banners
- ✅ MUXI branding in all banners

### **Input**
- ✅ Pre-fill existing values
- ✅ Confirm before replacing
- ✅ Check errors for Ctrl+C

### **Secrets**
- ✅ Masked display (***5678)
- ✅ Placeholders in formation
- ✅ Actual values in secrets file

### **State**
- ✅ Disable exits
- ✅ Enable continues to wizard
- ✅ Smart state-aware prompts

---

## Evolution

This document will evolve as we establish new patterns. When adding new wizards or commands, refer to these patterns for consistency.

**Questions?** Check existing implementations in:
- `src/pkg/scaffold/components.go` - A2A wizards (reference implementation)
- `src/pkg/wizard/wizard.go` - Prompt patterns
- `src/pkg/ui/ui.go` - Banner and message formatting

---

**Last major update:** A2A configuration wizards (2025-11-27)  
**Next patterns to establish:** Validation command, Secrets management
