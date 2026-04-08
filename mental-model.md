# Mental Model: MUXI CLI

**Generated:** 2026-04-08  
**Last Updated:** 2026-04-08

## Overview

The MUXI CLI is a Go command-line client for local development, remote formation management, and interactive chat against the MUXI runtime and MUXI server. It has two distinct network surfaces: the server management API (HMAC-authenticated `/rpc/*` endpoints) and proxied formation APIs (client/admin key headers for `/v1/*` runtime endpoints). The codebase is organized around Cobra commands in `src/cmd/` and thin service clients/helpers in `src/pkg/`.

## Architecture

### High-level shape

```text
CLI command (Cobra)
  ├─ pkg/server      -> talks to muxi-server over HMAC `/rpc/*`
  ├─ pkg/formation   -> talks to formation runtime `/v1/*`
  ├─ pkg/chat        -> interactive Bubble Tea chat TUI
  ├─ pkg/ui          -> shared terminal styling/output
  ├─ pkg/context     -> formation/project detection
  ├─ pkg/defaults    -> local defaults/config under ~/.muxi
  └─ pkg/secrets     -> encrypted local secret handling
```

### Key responsibilities

- `src/cmd/` — user-facing commands, flag parsing, command flow orchestration
- `src/pkg/server/` — MUXI server client, HMAC signing, deploy/start/restart SSE handling
- `src/pkg/formation/` — formation API client, chat/memory/MCP/session/runtime helpers
- `src/pkg/chat/` — Bubble Tea TUI for `muxi chat`, streaming event handling, session UX
- `src/pkg/ui/` — error blocks, badges, formatting helpers, consistent terminal output

## Directory Structure

- `src/cmd/` — Cobra command implementations (`chat.go`, `deploy.go`, `profiles.go`, etc.)
- `src/pkg/chat/` — interactive chat model and stream parsing
- `src/pkg/formation/` — runtime API client and response types
- `src/pkg/server/` — remote server client and deploy/start/restart SSE parser
- `src/pkg/scaffold/`, `src/pkg/validate/` — local project creation and validation
- `contributing/` — CLI-specific architecture, API, and UX conventions
- `CHANGELOG.md` — release history for shipped CLI behavior

## Data Flow

### Remote/server operations

```text
cmd/* -> pkg/server.Client -> HMAC-authenticated /rpc/* request
      -> optional SSE stream parser for deploy/start/restart/rollback
      -> ui formatting / command result
```

### Formation chat (one-shot)

```text
cmd/chat.go -> pkg/formation.Client.ChatStream()
            -> runtime /v1/chat SSE stream
            -> pkg/chat.StreamChatEvents()
            -> spinner / response rendering
```

### Formation chat (interactive TUI)

```text
cmd/chat.go -> pkg/chat.Run()
            -> Bubble Tea model in pkg/chat/chat.go
            -> pkg/formation.Client.ChatStream()
            -> processStreamWithEvents()
            -> StreamEvent channel -> TUI state updates
```

## Key Patterns & Conventions

- Commands live in `src/cmd/`; transport and shared logic live in `src/pkg/`
- Use `ui.*` helpers for terminal output instead of ad hoc printing where possible
- Interactive chat intentionally runs inline (no alt screen) so scrollback is preserved
- The CLI uses two auth models:
  - HMAC for server `/rpc/*`
  - `X-MUXI-CLIENT-KEY` / `X-MUXI-ADMIN-KEY` (+ `X-Muxi-User-ID`) for formation APIs
- Build and test from `src/`:
  - `go build ./...`
  - `go test ./...`

## Dependencies & Infrastructure

- Cobra for command structure
- Bubble Tea / Bubbles / Lip Gloss / Glamour for TUI and formatted chat output
- Standard `net/http` clients for both server and formation APIs
- Local state under `~/.muxi/` for profiles, defaults, server credentials, and related config

## Gotchas & Quirks

- The CLI has separate SSE implementations:
  - `pkg/server/client.go` for deploy/start/restart flows
  - chat-specific parsing in `pkg/chat/` / `cmd/chat.go`
  These paths can drift if not kept aligned.
- Formation chat streaming uses a dedicated `streamClient` in `pkg/formation/client.go` with `Timeout: 0`; if chat times out, the bug is often above the HTTP transport layer.
- The interactive chat TUI waits on parsed `StreamEvent`s, not raw socket bytes. If the parser drops valid SSE traffic, the UI can show a timeout even while the connection is still alive.
- `pkg/chat/chat.go` still contains older stream helpers (`processStream`) that are not the main interactive path; check actual call sites before changing parser behavior.

## Lessons Learned

### 2026-04-08 — Chat SSE keepalive handling must be parser-aware

- Runtime `/v1/chat` now sends SSE comment frames like `: keepalive` during slow setup and between token gaps.
- `muxi chat` still timed out because the TUI's 60-second wait was based on parsed `StreamEvent`s, not on raw SSE activity.
- Fix pattern:
  - parse SSE by event blocks (preserve `event:` until blank line)
  - treat comment frames as heartbeat/liveness
  - surface route-level `event: error` payloads
  - be tolerant of newer runtime token/event shapes instead of assuming a tiny allowlist

### 2026-04-08 — Shared parsers reduce drift

- The one-shot and interactive chat flows had duplicated SSE parsing logic with slightly different assumptions.
- A shared chat stream parser in `pkg/chat/` is the safer place for runtime-specific SSE semantics, while command/TUI layers should focus on presentation.
