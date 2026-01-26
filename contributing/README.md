# Contributing to MUXI CLI

Documentation for contributors working on the CLI codebase.

## Guides

| Document | Description |
|----------|-------------|
| [architecture.md](architecture.md) | CLI architecture and code organization |
| [api-reference.md](api-reference.md) | Server and Formation API reference |
| [streaming-events.md](streaming-events.md) | SSE event types and formats |
| [ux-patterns.md](ux-patterns.md) | TUI patterns and conventions |
| [tui-design.md](tui-design.md) | Design system and components |
| [banners.md](banners.md) | Banner and logo specifications |

## Getting Started

```bash
cd src
go build ./...
go test ./...
```

## Code Style

- Follow existing patterns in the codebase
- Use `ui.*` functions for terminal output
- Handle Ctrl+C gracefully at prompts

## Links

- [CONTRIBUTING.md](https://github.com/muxi-ai/muxi/blob/main/CONTRIBUTING.md) - Contribution guidelines
- [CODE_OF_CONDUCT.md](https://github.com/muxi-ai/muxi/blob/main/CODE_OF_CONDUCT.md) - Code of conduct
- [SECURITY.md](https://github.com/muxi-ai/muxi/blob/main/SECURITY.md) - Security policy
