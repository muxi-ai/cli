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

	"github.com/muxi-ai/cli/pkg/chat"
	"github.com/muxi-ai/cli/pkg/formation"
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
Use slash commands like /help, /agents, /exit for navigation.

One-shot mode: Pass a message as argument or use --file for audio/video.`,
	Example: `  # Interactive chat
  muxi chat
  muxi chat -s sess_abc123           # Resume session

  # One-shot text mode
  muxi chat "What's the weather?"
  echo "Analyze this" | muxi chat

  # One-shot audio/video mode
  muxi chat --file recording.m4a
  muxi chat --file video.mp4 "Summarize this video"`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	formation.AddCommonFlags(chatCmd)
	chatCmd.Flags().StringP("session", "s", "", "Resume session ID")
	chatCmd.Flags().StringP("group", "g", "", "Agent group for routing")
	chatCmd.Flags().String("file", "", "Audio/video file to send (max 100MB)")
	chatCmd.Flags().Bool("no-stream", false, "Disable streaming (wait for full response)")
	chatCmd.Flags().Bool("no-splash", false, "Skip welcome banner")
}

func runChat(cmd *cobra.Command, args []string) error {
	flags := formation.GetCommonFlags(cmd)
	fileFlag, _ := cmd.Flags().GetString("file")
	sessionID, _ := cmd.Flags().GetString("session")
	groupID, _ := cmd.Flags().GetString("group")
	noStream, _ := cmd.Flags().GetBool("no-stream")

	// Resolve formation ID
	formationID, err := formation.ResolveFormationID(flags.FormationID)
	if err != nil {
		formationID = "my-formation"
	}

	// Resolve server profile
	profile := formation.ResolveProfile(flags.Profile)
	if profile == "" {
		profile = "local"
	}

	// Resolve user ID
	userID := formation.ResolveUserID(flags.UserID)
	if userID == "" {
		userID = "default-user"
	}

	// Handle --file flag (one-shot avchat mode)
	if fileFlag != "" {
		prompt := ""
		if len(args) > 0 {
			prompt = strings.Join(args, " ")
		}
		return runAVChatOneShot(cmd, fileFlag, prompt, formationID, profile, userID, sessionID, noStream)
	}

	// Handle one-shot text mode (message provided as argument)
	if len(args) > 0 {
		message := strings.Join(args, " ")
		return runTextChatOneShot(cmd, message, formationID, profile, userID, sessionID, groupID, noStream)
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
			return runTextChatOneShot(cmd, message, formationID, profile, userID, sessionID, groupID, noStream)
		}
	}

	// Start interactive chat - create client
	client, err := formation.NewClientFromContext(profile, formationID)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	cfg := chat.Config{
		FormationID: formationID,
		ServerID:    profile,
		UserID:      userID,
		SessionID:   sessionID,
		GroupID:     groupID,
		Client:      client,
	}

	return chat.Run(cfg)
}

func runTextChatOneShot(cmd *cobra.Command, message, formationID, profile, userID, sessionID, groupID string, noStream bool) error {
	client, err := formation.NewClientFromContext(profile, formationID)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()

	req := &formation.ChatRequest{
		Message:   message,
		SessionID: sessionID,
		GroupID:   groupID,
		Stream:    !noStream,
	}

	if noStream {
		resp, err := client.Chat(req)
		if err != nil {
			return fmt.Errorf("chat failed: %w", err)
		}
		fmt.Printf("  %s %s\n\n", ui.GoldText("𝐌"), resp.Response)
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

	fmt.Printf("  %s ", ui.GoldText("𝐌"))
	return streamSSEResponse(resp)
}

func runAVChatOneShot(cmd *cobra.Command, filePath, prompt, formationID, profile, userID, sessionID string, noStream bool) error {
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

	// Validate file type
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

	client, err := formation.NewClientFromContext(profile, formationID)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	formation.PrintBadgeFromFlags(cmd)
	fmt.Println()
	fmt.Printf("  Sending %s (%.1f MB)...\n\n", filepath.Base(filePath), float64(info.Size())/(1024*1024))

	req := &formation.AVChatRequest{
		Files: []formation.ChatFile{
			{
				Filename:    filepath.Base(filePath),
				Content:     encoded,
				ContentType: contentType,
				Size:        info.Size(),
			},
		},
		UserID:         userID,
		SessionID:      sessionID,
		PromptTemplate: prompt,
	}

	if noStream {
		resp, err := client.AVChat(req)
		if err != nil {
			return fmt.Errorf("avchat failed: %w", err)
		}
		fmt.Printf("  %s %s\n\n", ui.GoldText("𝐌"), resp.Response)
		return nil
	}

	// Streaming mode
	resp, err := client.AVChatStream(req, userID)
	if err != nil {
		return fmt.Errorf("avchat failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("avchat failed: %s", resp.Status)
	}

	fmt.Printf("  %s ", ui.GoldText("𝐌"))
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
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if data != "" && data != "[DONE]" {
				// Parse JSON chunk and extract content delta
				var chunk struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if err := json.Unmarshal([]byte(data), &chunk); err == nil {
					if len(chunk.Choices) > 0 {
						fmt.Print(chunk.Choices[0].Delta.Content)
					}
				} else {
					// Fallback: print raw data if not OpenAI-style chunk
					fmt.Print(data)
				}
			}
		}
	}
	fmt.Println()
	fmt.Println()
	return scanner.Err()
}
