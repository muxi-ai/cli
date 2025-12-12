# Version Management Guide

This guide covers managing formation versions with the `muxi bump` command.

## Overview

Before deploying updates to a formation, you must increment the version. The `muxi bump` command simplifies this by automatically updating the `version` field in your `formation.afs`.

## Commands

### Bump Patch Version (Default)

```bash
muxi bump
```

Increments the patch version: `1.0.0` → `1.0.1`

This is the most common bump for bug fixes and small changes.

### Bump Minor Version

```bash
muxi bump minor
```

Increments the minor version and resets patch: `1.0.3` → `1.1.0`

Use for new features that are backward compatible.

### Bump Major Version

```bash
muxi bump major
```

Increments the major version and resets minor/patch: `1.2.3` → `2.0.0`

Use for breaking changes.

### Set Specific Version

```bash
muxi bump --set 2.0.0
```

Sets an exact version. Useful for:
- Aligning with external release schedules
- Skipping versions
- Initial version setup

## Output

```bash
$ muxi bump
✓ Formation version updated!
  1.0.0 => 1.0.1
```

When no version exists:
```bash
$ muxi bump
✓ Formation version updated!
  (none) => 1.0.0
```

## Version Field Location

The command adds or updates the `version` field in your `formation.afs`:

```yaml
schema: "1.0.0"
id: my-formation
version: "1.0.1"    # Added/updated by muxi bump
name: My Formation
```

If no version exists, the command:
1. Creates it with `1.0.0` (for patch bump)
2. Places it after the `id` field

## Typical Workflow

```bash
# Make changes to your formation
muxi edit formation

# Validate changes
muxi validate

# Bump version before deploy
muxi bump

# Deploy to server
muxi deploy
```

## Semver Format

Versions follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

| Component | When to Increment |
|-----------|-------------------|
| MAJOR | Breaking changes |
| MINOR | New features (backward compatible) |
| PATCH | Bug fixes, small changes |

## Error Handling

### Not in Formation Directory

```bash
$ muxi bump
✗ Not in formation directory

  This command must be run inside a formation directory.

  Navigate to your formation:
    cd my-formation
```

### Invalid Version Format

```bash
$ muxi bump --set invalid
✗ Invalid version

  'invalid' is not a valid semver version.

  Use format: major.minor.patch (e.g., 2.0.0)
```

### Invalid Bump Type

```bash
$ muxi bump foo
✗ Invalid bump type

  'foo' is not a valid bump type.

  Use: patch, minor, or major
```

## Tips

1. **Always bump before deploy** - The server rejects deployments with the same or lower version
2. **Use patch for most updates** - It's the default for a reason
3. **Commit after bumping** - Include the version bump in your git commit
4. **Check current version** - Look at `formation.afs` or use `muxi info` after deploying

## Related Commands

- `muxi deploy` - Deploy formation (requires version bump for updates)
- `muxi validate` - Validate formation before deploy
- `muxi info` - View deployed formation info including version
