package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muxi-ai/cli/pkg/chat"
	"github.com/muxi-ai/cli/pkg/formation"
	"github.com/muxi-ai/cli/pkg/telemetry"
	"github.com/muxi-ai/cli/pkg/ui"
	"github.com/spf13/cobra"
)

const maxAVFileSize = 100 * 1024 * 1024 // 100MB

var validAudioExtensions = map[string]string{
	".mp3":  "audio/mpeg",
	".m4a":  "audio/mp4",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",
	".aac":  "audio/aac",
	".wma":  "audio/x-ms-wma",
	".webm": "audio/webm",
}

var validVideoExtensions = map[string]string{
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".webm": "video/webm",
	".wmv":  "video/x-ms-wmv",
	".flv":  "video/x-flv",
	".m4v":  "video/x-m4v",
}

var chatCmd = &cobra.Command{
	Use:     "chat [message]",
	Short:   "Chat with a formation",
	GroupID: "formation",
	Long: `Start an interactive chat session with a formation.

In interactive mode, you can send messages and receive streaming responses.
Press / to see available commands, ? for keyboard shortcuts.

One-shot modes:
  - Text: Pass a message as argument
  - Voice note: Use --file with audio (transcribed as your message)  
  - File analysis: Use --file with a prompt to analyze audio/video`,
	Example: `  # Interactive chat
  muxi chat
  muxi chat -s sess_abc123           # Resume session

  # One-shot text mode
  muxi chat "What's the weather?"
  echo "Analyze this" | muxi chat

  # Voice note (audio transcribed as message)
  muxi chat --file voice.m4a

  # File analysis (with prompt)
  muxi chat --file video.mp4 "Summarize this"
  muxi chat --file meeting.mp3 "Extract action items"`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	formation.AddCommonFlags(chatCmd)
	chatCmd.Flags().StringP("session", "s", "", "Resume session ID")
	chatCmd.Flags().String("file", "", "Audio/video file to send (max 100MB)")
	chatCmd.Flags().Bool("no-stream", false, "Disable streaming (wait for full response)")
	chatCmd.Flags().BoolP("verbose", "v", false, "Show all streaming events (thinking, planning, progress)")
	chatCmd.Flags().Bool("debug", false, "Enable debug output to stderr")
}

func runChat(cmd *cobra.Command, args []string) error {
	// Track telemetry
	state := telemetry.Load()
	state.IncrementChat()
	state.FlushIfDue()
	defer state.Save()

	flags := formation.GetCommonFlags(cmd)
	fileFlag, _ := cmd.Flags().GetString("file")
	sessionID, _ := cmd.Flags().GetString("session")
	noStream, _ := cmd.Flags().GetBool("no-stream")

	// Resolve formation ID (required)
	formationID, err := formation.ResolveFormationID(flags.FormationID)
	if err != nil {
		return err
	}

	// Resolve server profile
	profile := formation.ResolveProfile(flags.Profile)

	// Resolve user ID (required for chat)
	userID := formation.ResolveUserID(flags.UserID)
	if userID == "" {
		return &formation.UserIDRequiredError{}
	}

	// Handle --file flag
	if fileFlag != "" {
		if len(args) > 0 {
			// File with prompt → /chat with file attachment
			prompt := strings.Join(args, " ")
			return runChatWithFile(cmd, fileFlag, prompt, formationID, profile, userID, sessionID, noStream)
		}
		// Audio file only (voice note) → /audiochat
		return runAudioChat(cmd, fileFlag, formationID, profile, userID, sessionID, noStream)
	}

	// Handle one-shot text mode (message provided as argument)
	if len(args) > 0 {
		message := strings.Join(args, " ")
		return runTextChatOneShot(cmd, message, formationID, profile, userID, sessionID, noStream)
	}

	// Check for piped input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Input is being piped
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if len(lines) > 0 {
			message := strings.Join(lines, "\n")
			return runTextChatOneShot(cmd, message, formationID, profile, userID, sessionID, noStream)
		}
	}

	// Start interactive chat - create client
	draft := formation.ResolveDraftMode(flags.Draft)
	client, err := formation.NewClientFromContext(profile, formationID, draft)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// TODO: Uncomment health check - disabled for direct runtime testing
	// if _, err := client.Health(); err != nil {
	// 	return fmt.Errorf("cannot connect to formation `%s` on `%s` - is it running?", formationID, profile)
	// }

	verbose, _ := cmd.Flags().GetBool("verbose")
	debug, _ := cmd.Flags().GetBool("debug")

	cfg := chat.Config{
		FormationID: formationID,
		ServerID:    profile,
		UserID:      userID,
		SessionID:   sessionID,

		Client:  client,
		Verbose: verbose,
		Debug:   debug,
	}

	return chat.Run(cfg)
}

func runTextChatOneShot(cmd *cobra.Command, message, formationID, profile, userID, sessionID string, noStream bool) error {
	draft, _ := cmd.Flags().GetBool("draft")
	client, err := formation.NewClientFromContext(profile, formationID, formation.ResolveDraftMode(draft))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	req := &formation.ChatRequest{
		Message:   message,
		SessionID: sessionID,
		Stream:    !noStream,
	}

	if noStream {
		resp, err := client.Chat(req)
		if err != nil {
			return fmt.Errorf("chat failed: %w", err)
		}
		fmt.Printf("  %s %s\n\n", ui.GoldText("𝐌"), resp.GetResponseText())
		// Save and display artifacts
		for _, art := range resp.GetResponseArtifacts() {
			path, err := formation.SaveArtifact(art, formationID)
			if err != nil {
				ui.Warning(fmt.Sprintf("  Failed to save %s: %v", art.Filename, err))
			} else {
				size := formation.FormatArtifactSize(art.Metadata.SizeBytes)
				fmt.Printf("  📎  %s (%s)\n     %s\n\n", art.Filename, size, ui.DimmedText(path))
			}
		}
		return nil
	}

	// Streaming mode
	resp, err := client.ChatStream(req, userID)
	if err != nil {
		return fmt.Errorf("chat failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("chat failed: %s", resp.Status)
	}

	return streamSSEResponse(resp)
}

// runAudioChat handles audio-only voice notes → /audiochat
func runAudioChat(cmd *cobra.Command, filePath, formationID, profile, userID, sessionID string, noStream bool) error {
	// Validate file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		ui.ErrorBlock("File not found", filePath, "")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Validate file size
	if info.Size() > maxAVFileSize {
		ui.ErrorBlock(
			"File too large",
			fmt.Sprintf("%s is %.1f MB (max 100 MB)", filepath.Base(filePath), float64(info.Size())/(1024*1024)),
			"",
		)
		return nil
	}

	// Validate file type - audio only for /audiochat (voice note mode)
	ext := strings.ToLower(filepath.Ext(filePath))
	contentType, ok := validAudioExtensions[ext]
	if !ok {
		ui.ErrorBlock(
			"Prompt required",
			fmt.Sprintf("%s is not an audio file", filepath.Base(filePath)),
			"Add a prompt to analyze the file:\n  muxi chat --file "+filepath.Base(filePath)+" \"describe this\"\n\nVoice notes (no prompt needed): mp3, m4a, wav, ogg, flac",
		)
		return nil
	}

	// Read and encode file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)

	draftFlag, _ := cmd.Flags().GetBool("draft")
	client, err := formation.NewClientFromContext(profile, formationID, formation.ResolveDraftMode(draftFlag))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	fmt.Print(dimStyle.Render(fmt.Sprintf("  ⠋ Sending %s (%.1f KB)...", filepath.Base(filePath), float64(info.Size())/1024)))

	req := &formation.AudioChatRequest{
		Files: []formation.ChatFile{
			{
				Filename:    filepath.Base(filePath),
				Content:     encoded,
				ContentType: contentType,
				Size:        info.Size(),
			},
		},
		UserID:    userID,
		SessionID: sessionID,
		Stream:    !noStream,
	}

	if noStream {
		resp, err := client.AudioChat(req)
		if err != nil {
			return fmt.Errorf("audiochat failed: %w", err)
		}
		fmt.Printf("%s\n", resp.GetResponseText())
		return nil
	}

	// Streaming mode
	resp, err := client.AudioChatStream(req, userID)
	if err != nil {
		return fmt.Errorf("audiochat failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("audiochat failed: %s", resp.Status)
	}

	return streamSSEResponse(resp)
}

// runChatWithFile handles file + prompt → /chat with file attachment
func runChatWithFile(cmd *cobra.Command, filePath, prompt, formationID, profile, userID, sessionID string, noStream bool) error {
	// Validate file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		ui.ErrorBlock("File not found", filePath, "")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Validate file size
	if info.Size() > maxAVFileSize {
		ui.ErrorBlock(
			"File too large",
			fmt.Sprintf("%s is %.1f MB (max 100 MB)", filepath.Base(filePath), float64(info.Size())/(1024*1024)),
			"",
		)
		return nil
	}

	// Get content type (audio or video)
	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := getAVContentType(ext)
	if contentType == "" {
		ui.ErrorBlock(
			"Invalid file type",
			fmt.Sprintf("%s is not a supported audio/video format", ext),
			"Supported: mp3, m4a, wav, ogg, flac, mp4, mov, avi, mkv, webm",
		)
		return nil
	}

	// Read and encode file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)

	draftFlag, _ := cmd.Flags().GetBool("draft")
	client, err := formation.NewClientFromContext(profile, formationID, formation.ResolveDraftMode(draftFlag))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	fmt.Print(dimStyle.Render(fmt.Sprintf("  ⠋ Sending %s (%.1f KB)...", filepath.Base(filePath), float64(info.Size())/1024)))

	req := &formation.ChatRequest{
		Message:   prompt,
		SessionID: sessionID,
		Stream:    !noStream,
		Files: []formation.ChatFile{
			{
				Filename:    filepath.Base(filePath),
				Content:     encoded,
				ContentType: contentType,
				Size:        info.Size(),
			},
		},
	}

	if noStream {
		resp, err := client.Chat(req)
		if err != nil {
			return fmt.Errorf("chat failed: %w", err)
		}
		fmt.Printf("%s\n", resp.GetResponseText())
		return nil
	}

	// Streaming mode
	resp, err := client.ChatStream(req, userID)
	if err != nil {
		return fmt.Errorf("chat failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("chat failed: %s", resp.Status)
	}

	return streamSSEResponse(resp)
}

func getAVContentType(ext string) string {
	if ct, ok := validAudioExtensions[ext]; ok {
		return ct
	}
	if ct, ok := validVideoExtensions[ext]; ok {
		return ct
	}
	// Fallback to mime package
	ct := mime.TypeByExtension(ext)
	if strings.HasPrefix(ct, "audio/") || strings.HasPrefix(ct, "video/") {
		return ct
	}
	return ""
}

func isAVFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return getAVContentType(ext) != ""
}

func streamSSEResponse(resp *http.Response) error {
	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	// Spinner frames
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Start/update spinner (replaces previous line)
	startSpinner := func(msg string) {
		// Truncate message to prevent line wrapping (leave room for spinner + padding)
		maxLen := 70
		if len(msg) > maxLen {
			msg = msg[:maxLen] + "..."
		}
		fmt.Print("\r\033[K") // Clear line
		fmt.Print(dimStyle.Render(fmt.Sprintf("  %s %s", spinnerFrames[spinnerIdx], msg)))
		spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip event type lines
		if strings.HasPrefix(line, "event:") {
			eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			if eventType == "done" {
				break
			}
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		// Parse MUXI token format: {"token": {...}}
		var muxiToken struct {
			Token struct {
				Type    string `json:"type"`
				Stage   string `json:"stage"`
				Content string `json:"content"`
			} `json:"token"`
		}

		if err := json.Unmarshal([]byte(data), &muxiToken); err == nil && muxiToken.Token.Type != "" {
			token := muxiToken.Token

			switch token.Type {
			case "thinking", "progress", "planning":
				startSpinner(token.Content)

			case "content", "text", "response":
				fullResponse.WriteString(token.Content)

			case "completed":
				// Use completed content if we didn't accumulate any
				if fullResponse.Len() == 0 && token.Content != "" && token.Content != "done" {
					fullResponse.WriteString(token.Content)
				}

			case "error":
				fmt.Print("\r\033[K")
				return fmt.Errorf("%s", token.Content)
			}
		}
	}

	// Clear spinner line and print response
	fmt.Print("\r\033[K")
	if fullResponse.Len() > 0 {
		fmt.Printf("%s\n", fullResponse.String())
	}

	return scanner.Err()
}
