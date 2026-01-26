# MUXI CLI

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/muxi-ai/cli/badge)](https://scorecard.dev/viewer/?uri=github.com/muxi-ai/cli)

The official command-line interface for creating, deploying, and managing MUXI AI agent formations.

> [!IMPORTANT]
> This repository is part of the MUXI ecosystem. See [ARCHITECTURE.md](https://github.com/muxi-ai/muxi/blob/main/ARCHITECTURE.md) for how all repositories fit together.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install muxi-ai/tap/muxi
```

### Direct Download

```bash
# macOS Apple Silicon
curl -L https://github.com/muxi-ai/cli/releases/latest/download/muxi-darwin-arm64 -o muxi

# macOS Intel
curl -L https://github.com/muxi-ai/cli/releases/latest/download/muxi-darwin-amd64 -o muxi

# Linux x86_64
curl -L https://github.com/muxi-ai/cli/releases/latest/download/muxi-linux-amd64 -o muxi

# Linux ARM64
curl -L https://github.com/muxi-ai/cli/releases/latest/download/muxi-linux-arm64 -o muxi

chmod +x muxi && sudo mv muxi /usr/local/bin/
```

### Windows

Download from [GitHub Releases](https://github.com/muxi-ai/cli/releases/latest):
- `muxi-windows-amd64.exe` (x86_64)
- `muxi-windows-arm64.exe` (ARM64)

## Quick Start

```bash
# Create a new formation
muxi new formation my-bot
cd my-bot

# Configure secrets
muxi secrets setup

# Run locally
muxi dev

# Chat with your formation
muxi chat "Hello!"

# Deploy to server
muxi deploy --profile production
```

## Documentation

- **[CLI Cheatsheet](https://muxi.dev/docs/cli/cheatsheet)** - Quick reference for all commands
- **[Getting Started](https://muxi.dev/docs/cli/setup)** - Installation and setup guide
- **[Full Documentation](https://muxi.dev/docs/cli)** - Complete CLI documentation

## Features

- **Formation Management** - Create, deploy, and manage AI agent formations
- **Interactive Chat** - TUI-based chat with streaming and markdown rendering
- **Server Profiles** - Multi-server support with HMAC authentication
- **Registry Integration** - Push/pull formations from the MUXI registry
- **Shell Completions** - Bash, Zsh, Fish, PowerShell

## Development

```bash
cd src
go build ./...
go test ./...
```

See [contributing/](contributing/) for contributor documentation.

## Building from Source

```bash
# Clone the repository
git clone https://github.com/muxi-ai/cli.git
cd cli/src

# Build for current platform
go build -o muxi .

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o muxi-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o muxi-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o muxi-windows-amd64.exe .
```

## Links

- [MUXI Documentation](https://muxi.dev/docs)
- [GitHub Issues](https://github.com/muxi-ai/cli/issues)
- [Contributing Guidelines](https://github.com/muxi-ai/muxi/blob/main/CONTRIBUTING.md)
- [Code of Conduct](https://github.com/muxi-ai/muxi/blob/main/CODE_OF_CONDUCT.md)
- [Security Policy](https://github.com/muxi-ai/muxi/blob/main/SECURITY.md)

## License

[Apache 2.0](LICENSE-Apache-2.0)
