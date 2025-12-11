# Plan: .afs Extension Support

**Date:** 2025-12-11
**Status:** In Progress

## Overview

Support `.afs` (Agent Formation Schema) as the default file extension for formation configs, while maintaining backward compatibility with `.yaml`.

## Behavior

### New Files (Creation)
- Use `file_extension` from `~/.muxi/cli/defaults.yaml` (default: `"afs"`)
- Generates: `formation.afs`, `agents/foo.afs`, `mcps/bar.afs`, `a2a/baz.afs`

### Existing Files (Read/Edit)
- **Always detect on disk** - don't rely on config
- Check `.afs` first, then `.yaml`
- Use whichever exists

### Bundling/Validation
- Accept **both** `.afs` and `.yaml` files
- Pack both when deploying

## Files Affected

### Phase 1: Foundation

| File | Changes |
|------|---------|
| `pkg/defaults/defaults.go` | Add `FileExtension` field, `GetFileExtension()` helper |
| `pkg/context/formation.go` | New `FindFormationFile()` that checks both extensions |

### Phase 2: Detection Helpers

Create or update helpers:
```go
// Check for file with either extension, return path that exists
func FindConfigFile(dir, baseName string) (string, error)

// Check if filename has .afs or .yaml extension
func HasConfigExtension(name string) bool

// Get preferred extension for new files
func GetPreferredExtension() string  // ".afs" or ".yaml"
```

### Phase 3: Validation & Bundling

| File | Changes |
|------|---------|
| `pkg/validate/validate.go` | Accept both extensions in `HasSuffix` checks |
| `pkg/registry/bundle.go` | Include `*.afs` patterns alongside `*.yaml` |
| `pkg/secrets/secrets.go` | Scan both extensions for secret refs |

### Phase 4: Scaffold (New File Creation)

| File | Changes |
|------|---------|
| `pkg/scaffold/formation.go` | Use preferred extension for new formations |
| `pkg/scaffold/components.go` | Use preferred extension for agents, mcps, a2a |
| `pkg/scaffold/templates.go` | Update README templates |

### Phase 5: Scaffold (Config Wizards - Detect Existing)

| File | Changes |
|------|---------|
| `pkg/scaffold/llm.go` | Detect formation.afs or formation.yaml |
| `pkg/scaffold/memory.go` | Detect formation.afs or formation.yaml |
| `pkg/scaffold/overlord.go` | Detect formation.afs or formation.yaml |
| `pkg/scaffold/security.go` | Detect formation.afs or formation.yaml |
| `pkg/scaffold/async.go` | Detect formation.afs or formation.yaml |
| `pkg/scaffold/logging.go` | Detect formation.afs or formation.yaml |

### Phase 6: Commands

| File | Changes |
|------|---------|
| `cmd/deploy.go` | Detect formation file |
| `cmd/edit.go` | Detect correct file to open |
| `cmd/registry.go` | Detect formation file, count both extensions |
| `cmd/validate.go` | Update help text |
| `cmd/shortcuts.go` | Detect formation file |

## Extension Mapping

| Directory | Extensions |
|-----------|------------|
| Root | `formation.afs`, `formation.yaml` |
| agents/ | `*.afs`, `*.yaml` |
| mcps/ | `*.afs`, `*.yaml` |
| a2a/ | `*.afs`, `*.yaml` |
| triggers/ | `*.yaml` only (markdown templates) |
| sops/ | `*.md` only |
| knowledge/ | `*.md`, `*.txt` |

## Helper Functions

```go
// pkg/context/formation.go

// FindFormationFile returns the path to formation.afs or formation.yaml
func FindFormationFile(dir string) (string, bool) {
    afs := filepath.Join(dir, "formation.afs")
    if _, err := os.Stat(afs); err == nil {
        return afs, true
    }
    yaml := filepath.Join(dir, "formation.yaml")
    if _, err := os.Stat(yaml); err == nil {
        return yaml, true
    }
    return "", false
}

// FindConfigFile finds a config file with either extension
func FindConfigFile(dir, baseName string) (string, bool) {
    afs := filepath.Join(dir, baseName+".afs")
    if _, err := os.Stat(afs); err == nil {
        return afs, true
    }
    yaml := filepath.Join(dir, baseName+".yaml")
    if _, err := os.Stat(yaml); err == nil {
        return yaml, true
    }
    return "", false
}

// HasConfigExtension checks if filename ends with .afs or .yaml
func HasConfigExtension(name string) bool {
    return strings.HasSuffix(name, ".afs") || strings.HasSuffix(name, ".yaml")
}

// pkg/defaults/defaults.go

// GetFileExtension returns preferred extension for new files (without dot)
func GetFileExtension() string {
    config, _ := Load()
    if config.FileExtension == "yaml" {
        return "yaml"
    }
    return "afs" // default
}
```

## Implementation Order

1. `pkg/defaults/defaults.go` - Add FileExtension field
2. `pkg/context/formation.go` - Add FindFormationFile, FindConfigFile, HasConfigExtension
3. `pkg/validate/validate.go` - Use HasConfigExtension
4. `pkg/registry/bundle.go` - Add *.afs patterns
5. `pkg/secrets/secrets.go` - Add *.afs patterns  
6. `pkg/scaffold/formation.go` - Use GetFileExtension for new files
7. `pkg/scaffold/components.go` - Use GetFileExtension, FindConfigFile
8. `pkg/scaffold/*.go` (wizards) - Use FindFormationFile
9. `cmd/*.go` - Use FindFormationFile, FindConfigFile
10. Update help text and templates

## Testing

- Create new formation → should use .afs
- Set file_extension: yaml, create new → should use .yaml
- Edit existing .yaml formation → should find and edit .yaml
- Validate formation with mixed extensions → should work
- Deploy formation with .afs files → should bundle correctly
