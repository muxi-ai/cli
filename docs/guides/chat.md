# Chat Guide

Chat with formations using interactive or one-shot modes.

## Interactive Mode

Start an interactive chat session:

```bash
muxi chat                           # Chat with current formation
muxi chat -f my-formation           # Specify formation
muxi chat -s sess_abc123            # Resume existing session
```

### Keyboard Shortcuts

- `/` - Show available commands
- `?` - Show keyboard shortcuts
- `Ctrl+C` - Cancel current request
- `Esc` - Exit chat

## One-Shot Text Mode

Send a single message and get a response:

```bash
muxi chat "What's the weather in London?"
echo "Analyze this data" | muxi chat
```

## Voice Notes (Audio Chat)

Send audio files as voice messages (transcribed automatically):

```bash
muxi chat --file voice.m4a
muxi chat --file recording.mp3
```

Supported audio formats: mp3, m4a, wav, ogg, flac, aac, wma, aiff

The audio is transcribed and used as your message - no prompt needed.

## File Analysis

Analyze audio/video files with a prompt:

```bash
muxi chat --file video.mp4 "Summarize this video"
muxi chat --file meeting.mp3 "Extract action items"
muxi chat --file podcast.m4a "What topics are discussed?"
```

Supported formats:
- Audio: mp3, m4a, wav, ogg, flac, aac, wma, aiff
- Video: mp4, mov, avi, mkv, webm, wmv, flv

Maximum file size: 100 MB

## Options

```
-f, --formation    Formation ID
-p, --profile      Server profile
-u, --user         User ID (required)
-s, --session      Resume session ID
    --file         Audio/video file to send
    --no-stream    Disable streaming (wait for full response)
    --verbose      Show detailed output
    --debug        Enable debug mode
```

## Examples

```bash
# Interactive chat with specific user
muxi chat -u alice

# One-shot with formation context
muxi chat -f customer-support "How do I reset my password?"

# Voice note from WhatsApp
muxi chat --file voice-message.m4a

# Analyze a video
muxi chat --file demo.mp4 "What features are shown?"

# Resume a previous session
muxi chat -s sess_abc123 "Continue where we left off"
```
