# Server & Formation Commands

This guide covers deploying and managing formations on MUXI servers.

## Server Management

### Adding a Server

```bash
muxi server add <name> --url <url> --key-id <id> --secret-key <key>
```

The server credentials are stored in `~/.muxi/cli/servers.yaml`.

**Example:**
```bash
muxi server add production \
  --url https://muxi.company.com:7890 \
  --key-id MUXI_PROD_abc123 \
  --secret-key your-secret-key
```

### Listing Servers

```bash
muxi server list
```

Shows all configured servers with their online/offline status.

### Setting Default Server

```bash
muxi server default <name>
```

Sets the default server for all commands when `--profile` is not specified.

### Server Status

```bash
muxi server status [--profile <name>]
```

Shows server version, uptime, and formation counts.

### Server Ping

```bash
muxi server ping [--profile <name>]
```

Continuous ping with latency statistics, like network ping.

---

## Deploying Formations

### Deploy Command

```bash
muxi deploy [--profile <name>] [--dry-run] [--no-stream]
```

Run from inside a formation directory. The CLI will:
1. Validate the formation
2. Create a tar.gz bundle
3. Auto-detect if this is a new deploy or an update
4. Stream progress from the server

**Flags:**
- `--profile` - Server profile to use (defaults to default server)
- `--dry-run` - Validate and create bundle without deploying
- `--no-stream` - Disable SSE streaming (simpler output)

### New Deploy vs Update

The CLI automatically detects whether to create a new formation or update an existing one:
- If the formation ID doesn't exist on the server → **new deploy**
- If the formation ID exists → **update** (requires higher version)

### Version Validation

When updating, the version in your formation config must be higher than the server version:

```yaml
# formation.afs
id: my-formation
name: My Formation
version: "1.1.0"  # Must be higher than server version
```

If you try to deploy the same or lower version, you'll see:
```
✗ Version conflict

Cannot update 'my-formation' to version 1.0.0
Server already has version 1.0.0

Bump the version in formation.afs and try again.
```

---

## SSE Streaming Progress

All long-running commands show real-time progress from the server:

```
  Deploying: my-formation to localhost

  ✓ Bundle ready (3 files, 1.2 KB)
  ✓ [SERVER] Extracted bundle
  ✓ [SERVER] Validated formation
  ✓ [SERVER] Resolved runtime 0.2025.0
  ◐ [SERVER] Downloading runtime image... 45%
```

### Progress Stages

**Deploy/Update:**
1. `extracting` - Extracting bundle
2. `validating` - Validating formation config
3. `resolving_runtime` - Resolving runtime version
4. `downloading_sif` - Downloading SIF file (with %)
5. `pulling_runner` - Pulling Docker image
6. `spawning` / `spawning_staging` - Starting formation
7. `health_check` - Waiting for health check
8. `swapping` - Switching to new version (update only)
9. `stopping_old` - Stopping old version (update only)

**Start/Restart/Rollback:**
- Similar stages depending on the operation

### Cancellation (Ctrl+C)

Press Ctrl+C during any streaming operation to cancel:
- The CLI sends a cancel request to the server
- Server cleans up staging resources
- Formation continues running (for updates) or is cleaned up (for new deploys)

---

## Formation Lifecycle Commands

### List Formations

```bash
muxi formation list [--profile <name>]
```

Shows all deployed formations with status, version, port, and uptime.

### Get Formation Details

```bash
muxi formation get <id> [-v|--verbose] [--profile <name>]
```

Shows formation details. Use `-v` for internal details like port and PID.

### Stop Formation

```bash
muxi formation stop <id> [-f|--force] [--profile <name>]
```

Stops a running formation. The formation remains registered and can be started again.

### Start Formation

```bash
muxi formation start <id> [--profile <name>]
```

Starts a stopped formation. Shows SSE streaming progress.

### Restart Formation

```bash
muxi formation restart <id> [-f|--force] [--profile <name>]
```

Restarts a formation. Shows SSE streaming progress.

### Rollback Formation

```bash
muxi formation rollback <id> [-f|--force] [--profile <name>]
```

Rolls back to the previous version. Only available if a previous version exists.

### Delete Formation

```bash
muxi formation delete <id> [-f|--force] [--profile <name>]
```

Permanently deletes a formation from the server.

### View Logs

```bash
muxi formation logs <id> [flags]
```

**Flags:**
- `-n, --lines <num>` - Number of lines to show (default 100)
- `-f, --follow` - Stream new logs (like `tail -f`)
- `--stream stdout|stderr` - Filter by stream
- `--profile <name>` - Server profile to use

---

## Shortcut Commands

When inside a formation directory, you can use shortcut commands:

```bash
muxi get [-v]              # Get current formation details
muxi stop [-f]             # Stop current formation
muxi start                 # Start current formation
muxi restart [-f]          # Restart current formation
muxi rollback [-f]         # Rollback current formation
muxi delete [-f]           # Delete current formation
muxi logs [-n 100] [-f]    # View logs
```

These commands automatically detect the formation ID from the formation config.

---

## Formation-Level Settings

Set default server/registry for a formation:

```bash
muxi set server            # Interactive wizard
muxi set registry          # Interactive wizard
```

Settings are saved to `.muxi` file in the formation directory.

---

## Notification Sounds

On macOS, the CLI plays notification sounds:
- **Success:** Glass.aiff
- **Failure:** Sosumi.aiff

On other platforms, a terminal bell is used.

---

## Examples

### Complete Workflow

```bash
# Create a new formation
muxi new formation
cd my-formation

# Configure LLM
muxi config llm

# Set secrets
muxi secrets set OPENAI_API_KEY

# Validate
muxi validate

# Deploy
muxi deploy

# Check status
muxi get -v

# View logs
muxi logs -f

# Update (bump version first)
# Edit formation.yaml to increase version
muxi deploy

# Rollback if needed
muxi rollback

# Stop when done
muxi stop

# Delete
muxi delete -f
```

### Multi-Server Deployment

```bash
# Add servers
muxi server add staging --url https://staging.company.com:7890 ...
muxi server add production --url https://prod.company.com:7890 ...

# Deploy to staging
muxi deploy --profile staging

# Test on staging...

# Deploy to production
muxi deploy --profile production
```
