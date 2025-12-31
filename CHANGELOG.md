# Changelog

All notable changes to the MUXI CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [ScalVer](https://scalver.org/) versioning: `MAJOR.YYYYMMDD.PATCH`.

## [Unreleased]

### Added
- Interactive chat with TUI (streaming, markdown rendering, slash commands)
- One-shot chat modes (text, voice notes, file analysis)
- `/chat` endpoint with file attachments
- `/audiochat` endpoint for voice notes
- Request cancellation with ESC key
- Full Formation API support (agents, sessions, triggers, secrets, etc.)
- User guides for all commands

### Fixed
- SSE streaming with `DisableCompression` for proper event buffering
- Line wrapping in markdown rendering

### Changed
- Renamed `AVChatRequest` to `AudioChatRequest`
- Improved error messages for file uploads
