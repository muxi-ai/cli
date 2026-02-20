# Changelog

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
- `muxi server list/get/stop/start/restart/delete/rollback` - Manage deployed formations
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
