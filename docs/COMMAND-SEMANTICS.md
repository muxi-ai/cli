# Command Semantics - `muxi new` vs `muxi config`

**Date:** 2025-11-27  
**Status:** ✅ Implemented  
**Decision:** Semantic separation between file creation and configuration modification

---

## The Problem

Initially, we had an inconsistency in command semantics:

```bash
muxi new formation    # ✅ Creates NEW directory + files
muxi new agent       # ✅ Creates NEW file (agents/*.yaml)
muxi new mcp         # ✅ Creates NEW file (mcps/*.yaml)
muxi new a2a         # ❌ Modifies EXISTING formation.yaml  <- INCONSISTENT!
```

**Issue:** `muxi new a2a` didn't create a NEW file - it modified an EXISTING file (formation.yaml). This violated user expectations and command semantics.

---

## The Solution

Split commands by their action:

### **`muxi new` = Create new files/directories**
```bash
muxi new formation          # Creates formation directory + all files
muxi new agent              # Creates agents/{name}.yaml
muxi new mcp                # Creates mcps/{name}.yaml (formation-level)
muxi new mcp --agent=X      # Appends to agents/{X}.yaml (agent-level)
muxi new sop                # Creates sops/{name}.md
muxi new trigger            # Creates triggers/{name}.yaml
```

**Future:**
```bash
muxi new a2a-service        # Creates a2a/{name}.yaml (outbound service definition)
```

### **`muxi config` = Modify existing formation.yaml**
```bash
muxi config a2a --inbound   # Modifies formation.yaml (a2a.inbound section)
muxi config a2a --outbound  # Modifies formation.yaml (a2a.outbound section)
```

**Future:**
```bash
muxi config llm             # Modify formation.yaml (llm section)
muxi config logging         # Modify formation.yaml (logging section)
muxi config scheduler       # Modify formation.yaml (scheduler section)
muxi config filtering       # Modify formation.yaml (a2a.filtering section)
muxi config webhooks        # Modify formation.yaml (webhooks section)
muxi config secrets --provider aws  # Modify formation.yaml (secrets.provider)
```

---

## Benefits

### 1. **Semantic Clarity**
- `new` = "I want to create something"
- `config` = "I want to change settings"

### 2. **Consistency**
All `muxi new` commands create files. No exceptions.

### 3. **Discoverability**
```bash
muxi new --help      # Shows what you can create
muxi config --help   # Shows what you can configure
```

Users know exactly where to look.

### 4. **Future-Proof**
Natural pattern for all formation.yaml sections:

```yaml
# formation.yaml sections → muxi config commands
llm:                  # muxi config llm
  providers: [...]
logging:              # muxi config logging
  streams: [...]
scheduler:            # muxi config scheduler
  enabled: true
a2a:                  # muxi config a2a
  inbound: {...}
  outbound: {...}
webhooks:             # muxi config webhooks
  url: "..."
```

### 5. **Mental Model**
**Files & Directories:**
```
formation/
├── formation.yaml    # muxi config <section>
├── agents/           # muxi new agent
├── mcps/             # muxi new mcp
├── a2a/              # muxi new a2a-service (future)
├── sops/             # muxi new sop
└── triggers/         # muxi new trigger
```

---

## A2A Specifics

### **Formation-Level A2A (in `formation.yaml`)**
```bash
muxi config a2a --inbound
muxi config a2a --outbound
```

Configures:
```yaml
# formation.yaml
a2a:
  enabled: true
  inbound:
    enabled: true
    registries: [...]
    auth: {...}
  outbound:
    enabled: true
    registries: [...]
```

### **A2A Service Files (in `a2a/*.yaml`)** - Future
```bash
muxi new a2a-service billing
```

Creates:
```yaml
# a2a/billing.yaml
schema: "1.0.0"
id: "billing"
name: "External Billing Service"
url: "https://billing.external.com/a2a"
auth:
  type: "api_key"
  key: "${{ secrets.BILLING_API_KEY }}"
```

---

## Implementation

### Files Changed
1. **Created:** `src/cmd/config.go` - New config command with a2a subcommand
2. **Modified:** `src/cmd/new.go` - Removed A2A command

### Code Organization
```go
// src/cmd/config.go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Configure formation settings",
}

var configA2ACmd = &cobra.Command{
    Use:   "a2a",
    Short: "Configure A2A (Agent-to-Agent) communication",
    RunE: func(cmd *cobra.Command, args []string) error {
        inbound, _ := cmd.Flags().GetBool("inbound")
        outbound, _ := cmd.Flags().GetBool("outbound")
        noWizard, _ := cmd.Flags().GetBool("no-wizard")
        
        return scaffold.ConfigureA2A(inbound, outbound, noWizard)
    },
}
```

### Shared Logic
The `scaffold.ConfigureA2A()` function remains unchanged - only the CLI command changed.

---

## Migration Guide

### For Users

**Old command:**
```bash
muxi new a2a --inbound
```

**New command:**
```bash
muxi config a2a --inbound
```

**Impact:** Command renamed for semantic clarity. Functionality identical.

### For Developers

No changes to scaffold logic. Only command routing changed:
- **Old:** `src/cmd/new.go` → `scaffold.ConfigureA2A()`
- **New:** `src/cmd/config.go` → `scaffold.ConfigureA2A()`

---

## Future Commands

As we add more formation.yaml configuration, use `muxi config`:

| Section | Command |
|---------|---------|
| `llm:` | `muxi config llm` |
| `logging:` | `muxi config logging` |
| `scheduler:` | `muxi config scheduler` |
| `a2a.filtering:` | `muxi config filtering` |
| `webhooks:` | `muxi config webhooks` |
| `secrets.provider:` | `muxi config secrets --provider aws` |

**Rule:** If it modifies formation.yaml, it's `muxi config <section>`.

---

## Rationale Summary

**Why not `muxi new a2a`?**
- Violates semantic expectations ("new" should create files)
- Inconsistent with other `muxi new` commands
- Confusing for users ("where's my new a2a file?")

**Why `muxi config a2a`?**
- Matches user intent ("I want to configure A2A settings")
- Consistent pattern for all formation.yaml sections
- Clear separation of concerns (create vs modify)
- Future-proof for additional config commands

---

## References

- **Schema Documentation:** `/Users/ran/Projects/muxi/code/schemas/formation/README.md`
- **Implementation:** `src/cmd/config.go`, `src/pkg/scaffold/components.go`
- **Related Commands:** `muxi new formation`, `muxi new agent`, `muxi new mcp`
