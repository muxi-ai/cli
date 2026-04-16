# Changelog

## 0.20260416.0 - 2026-04-16

### Fixed
- Map `muxi scheduler list` and `muxi scheduler show` to the runtime scheduler job fields (`status`, `is_recurring`, `cron_expression`, `scheduled_for`, `original_prompt`, `last_run_at`, `total_failures`) so active jobs render correctly instead of appearing disabled with empty columns.
- Compute scheduler next-run values client-side for recurring cron jobs and reuse `scheduled_for` for one-time jobs when the runtime does not provide a precomputed `next_run`.

## 0.20260413.0 - 2026-04-13

### Fixed
- Align `muxi validate` with the updated MCP declaration spec by recognizing `mcps.servers` while remaining compatible with legacy `mcp.servers` manifests.
- Match MCP config files from both `mcps/` and legacy `mcp/` directories so declaration checks and required-field validation report the correct results during migration.

## 0.20260408.0 - 2026-04-08

### Fixed
- Prevent `muxi chat` from timing out on long-running streamed requests that only emit SSE keepalive comment frames during slow setup or between token gaps.
- Parse chat SSE streams using proper event-block handling so route-level `event: error` frames surface correctly and newer runtime progress/tool activity no longer gets dropped by the CLI.

## 0.20260324.0 - 2026-03-24

### Fixed
- Prevent CLI chat from hanging indefinitely while waiting for streaming events by adding a 60s timeout with a user-facing error.

## 0.20260323.0 - 2026-03-23

### Changed
- Use CDN-based release download URLs (`https://releases.muxi.org/...`) instead of direct GitHub release URLs in CLI upgrade flow and release workflow examples.

## 0.20260314.0 - 2026-03-14

### Breaking Changes
- **`muxi server` renamed to `muxi remote`** - All deployed formation management commands moved from `muxi server <action>` to `muxi remote <action>` to avoid confusion with `muxi-server` (the server daemon).

### Added
- `muxi server` now proxies to `muxi-server` binary -- forwards all args, signals, and exit codes

### Changed
- Deploy and registry bundles now skip undeclared component files (backward compat: no declarations = include all)
- `muxi validate` shows file names in errors/warnings, splits multi-line parse errors into individual entries
- Registry publish error handling: detect `{"error":true}` responses returned with 200 status

## 0.20260306.0 - 2026-03-06

### Breaking Changes
- **Explicit component declaration** - Components must now be listed in `formation.yaml` / `formation.afs` to be loaded. Files in `agents/`, `mcps/`, `a2a/` are definitions only -- the `active` field is removed.

### Added
- `muxi new agent` now auto-registers the agent in the formation manifest (`agents:` list)
- `muxi new mcp` (formation-level) now auto-registers in the formation manifest (`mcp.servers:` list)
- `muxi new a2a-service` now auto-registers in the formation manifest (`a2a.outbound.services:` list)
- `muxi validate` warns when component files exist but are not declared (won't be loaded), and errors when declared IDs have no matching file
- Formation scaffold includes commented `agents:` and `mcp:` sections showing declaration syntax
- `muxi artifacts list` - List saved artifacts grouped by formation
- `muxi artifacts open` - Open artifacts directory in file manager
- `muxi artifacts cleanup` - Remove old artifacts with `--days` and `--formation` filters
- Artifacts commands auto-scope to current formation when inside a formation directory
- File artifacts in chat: streaming and non-streaming responses save binary/text artifacts to `~/.muxi/cli/artifacts/`

### Changed
- Chat banner redesigned with new MUXI logo (gradient) and formation info panel
- Scanner buffer increased to 100MB (fixes "disconnected by peer" with large artifact responses)
- Artifacts directory changed from `~/.muxi/cli/outputs/` to `~/.muxi/cli/artifacts/`
- `ChatResponse.Response` now handles string, object, and array response formats

### Removed
- `active: true` field from all component templates (agent, MCP, A2A)

## Unreleased

### Added
- `muxi up` - Start formation in local development mode (runs from source directory)
- `muxi down` - Stop formation running in local development mode

## 0.20260126.0 - 2026-01-26

### Core Features

- **Formation Management** - Create, deploy, and manage AI agent formations
- **Interactive Chat** - TUI-based chat with streaming, markdown rendering, and slash commands
- **One-Shot Chat** - Text, voice notes, and file analysis modes
- **Server Profiles** - Multi-server support with HMAC authentication
- **Registry Integration** - Push/pull formations from the MUXI registry

### Commands

#### Scaffolding
- `muxi new formation` - Create new formation with interactive wizard
- `muxi new agent` - Add agent to formation
- `muxi new mcp` - Add MCP server (formation or agent level)
- `muxi new trigger` - Add webhook trigger
- `muxi new sop` - Add Standard Operating Procedure

#### Configuration
- `muxi config llm` - Configure LLM providers and models
- `muxi config overlord` - Configure Overlord soul and behavior
- `muxi config memory` - Configure memory settings
- `muxi config logging` - Configure logging streams
- `muxi config async` - Configure async response settings
- `muxi config a2a` - Configure Agent-to-Agent communication
- `muxi config security` - Configure user credential handling

#### Secrets
- `muxi secrets setup` - Interactive secrets wizard
- `muxi secrets set/get/list/delete` - Manage encrypted secrets

#### Development
- `muxi dev` - Run formation locally
- `muxi validate` - Validate formation configuration
- `muxi chat` - Interactive chat session
- `muxi bump` - Bump formation version

#### Deployment
- `muxi deploy` - Deploy formation to server
- `muxi remote list/get/stop/start/restart/delete/rollback` - Manage deployed formations
- `muxi logs` - Stream formation logs with filtering

#### Formation API
- `muxi info` - View formation runtime status
- `muxi agents list/show` - View agents in running formation
- `muxi mcp list/show/tools` - View MCP servers and tools
- `muxi sessions list/show/messages` - Manage user sessions
- `muxi memory status/list/add/delete/buffer` - Manage user memories
- `muxi triggers list/show/run` - View and execute triggers
- `muxi sops list/show` - View SOPs
- `muxi scheduler status/list/show/add/remove` - Manage scheduled jobs
- `muxi requests list/show/cancel` - Track chat requests

#### Registry
- `muxi login/logout` - Registry authentication
- `muxi push/pull` - Publish and download formations
- `muxi search/show` - Discover formations

#### Profiles
- `muxi profiles add/list/remove` - Manage server profiles
- `muxi formations add/list/show/remove` - Save formation configs for quick access
- `muxi set default profile/user` - Set defaults

### Technical Features

- **SSE Streaming** - Real-time progress for deploy, start, restart, rollback
- **HMAC Authentication** - Secure server communication
- **Fernet Encryption** - Local secrets encryption
- **Shell Completions** - Bash, Zsh, Fish, PowerShell with caching
- **Telemetry** - Anonymous usage analytics (opt-out available)
- **X-Muxi-SDK Header** - SDK identification for all API calls

### Platforms

- Linux (amd64, arm64)
- macOS (Intel, Apple Silicon)
- Windows (amd64, arm64)

### Installation

```bash
# Homebrew (macOS/Linux)
brew install muxi-ai/tap/muxi

# Direct download
curl -L https://github.com/muxi-ai/cli/releases/latest/download/muxi-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o muxi
chmod +x muxi && sudo mv muxi /usr/local/bin/
```
