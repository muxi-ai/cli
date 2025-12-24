// Package chat implements the MUXI Chat TUI using Bubble Tea.
//
// # Architecture Decision: No Alt Screen (Inline Mode)
//
// This implementation uses Bubble Tea WITHOUT tea.WithAltScreen() to enable
// terminal scrollback. Messages are printed above the TUI using tea.Println()
// which persists them in the terminal's scrollback buffer.
//
// Key patterns:
//   - tea.Println() prints persistent content above the TUI (stays in scrollback)
//   - View() only renders the input area (small, fixed height)
//   - Header is printed once at startup via tea.Println()
//   - Messages/thinking steps are printed via tea.Println() as they arrive
//
// Trade-off: Input area moves when menus (?, /) open/close. This is acceptable
// and matches behavior of other CLI tools like Droid/Claude Code.
//
// # Message Formats
//
// User messages:
//
//	>  message content
//
// Assistant messages (with markdown rendering via glamour):
//
//	𝐌 response content
//
// Thinking steps (printed as they complete):
//
//	⏺  Analyzing request...
//	⏺  Routing to agent...
//
// Request aborted (red, when user presses ESC during thinking):
//
//	✕  Request aborted by user
//
// Server error (yellow/warning):
//
//	⚠  Server returned an error:
//	   (error details)
//
// # UX Features
//
//   - Input history: ↑/↓ arrows navigate previous inputs
//   - Ctrl+C: First press clears input (shows "Press Ctrl+C again to exit" for 3s)
//   - ESC during thinking: Aborts request, shows red abort message
//   - ? key: Shows help screen
//   - / key: Shows command menu with submenus
package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muxi-ai/cli/pkg/formation"
)

// Config holds chat session configuration
type Config struct {
	FormationID string
	ServerID    string
	UserID      string
	SessionID   string

	Client      *formation.Client
	Verbose     bool // Show all streaming events (don't replace)
	Debug       bool // Enable debug output to stderr
}

// StreamEvent represents a streaming event from the server
type StreamEvent struct {
	Type      string // progress, thinking, planning, content, completed, error
	Stage     string // init, tool_call, response_preparation, etc.
	Content   string
	ToolName  string
	SessionID string // Session ID from server
	RequestID string // Request ID from server (for cancellation)
}

// Message represents a chat message
type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// ThinkingStep represents a thinking update
type ThinkingStep struct {
	Text      string
	Completed bool
}

// Model is the Bubble Tea model for the chat UI
type Model struct {
	config          Config
	messages        []Message
	thinking        []ThinkingStep
	isThinking      bool
	thinkingStart   time.Time
	textarea        textarea.Model
	viewport        viewport.Model
	renderer        *glamour.TermRenderer
	width           int
	height          int
	ready           bool
	quitting        bool
	showHelp        bool
	showCommands    bool
	commandSelected int
	showSubmenu     bool
	submenuSelected int
	submenuParent   string
	asyncMode       string // "auto", "on", "off" - shown as ⚡/A/S indicator
	err             error
	inputHistory    []string  // History of user inputs
	historyIndex    int       // Current position in history (-1 = new input)
	currentInput    string    // Saved current input when navigating history
	showExitHint    bool      // Show "Ctrl+C again to exit" hint
	exitHintStart   time.Time // When the exit hint was shown
	requestAborted   bool   // Flag to ignore response after abort
	sessionID        string // Current session ID (from API response)
	currentRequestID string // Current request ID (for cancellation)
	lastError        string        // Last error message from server
	streamBody       io.ReadCloser // Current stream body (for cancellation)
	streamingText   strings.Builder // Accumulated streaming response
	eventChan       chan StreamEvent // Channel for streaming events
	currentEvent    *StreamEvent     // Current event being displayed
	eventHistory    []StreamEvent    // History of events (for verbose mode)
}

// Command represents a slash command
type Command struct {
	Name        string
	Description string
	Submenu     []SubItem
}

// SubItem represents a submenu option
type SubItem struct {
	Name        string
	Description string
}

// Available commands
var commands = []Command{
	{"/async", "Toggle async mode", []SubItem{
		{"auto", "Let formation decide"},
		{"on", "Enable async mode"},
		{"off", "Disable async mode"},
	}},
	{"/clear", "Clear history and start new session", nil},
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c98b45")) // Dimmed orange for user messages

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Same as footer

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Same as footer

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Explicit gray for all terminals

	goldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c98b45")) // Gold (matches muxi brand)

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Explicit gray for all terminals
)

// chatStyle is a minimal glamour style without horizontal rules or header decorations
var chatStyle = []byte(`{
	"document": { "margin": 0 },
	"heading": { "bold": true },
	"h1": { "bold": true },
	"h2": { "bold": true },
	"h3": { "bold": true },
	"strong": { "bold": true },
	"emph": { "italic": true },
	"list": { "level_indent": 2 },
	"item": { "block_prefix": "• " },
	"code": {},
	"code_block": { "margin": 2 },
	"horizontal_rule": { "format": "" },
	"paragraph": {}
}`)

// New creates a new chat model
func New(cfg Config) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false

	// Remove all borders and backgrounds for clean look
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLineNumber = lipgloss.NewStyle()
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle()
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()

	// Clear the prompt character (removes left border)
	ta.Prompt = ""

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("notty"),
		glamour.WithWordWrap(80),
	)

	return Model{
		config:       cfg,
		textarea:     ta,
		renderer:     renderer,
		messages:     []Message{},
		thinking:     []ThinkingStep{},
		asyncMode:    "auto",
		inputHistory: []string{},
		historyIndex: -1,
		sessionID:    cfg.SessionID, // Empty on first request, server returns ID
		eventChan:    make(chan StreamEvent, 10),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.printHeaderAbove(),
	)
}

// printHeaderAbove prints the header above the TUI (persists in scrollback)
func (m Model) printHeaderAbove() tea.Cmd {
	gold := goldStyle.Render
	dim := dimmedStyle.Render

	header := "\n" +
		dim("╭── ") + gold("MUXI Chat") + dim(" ─────────────────────────────────────────────────╮") + "\n" +
		dim("│") + "               " + dim("│") + "                                              " + dim("│") + "\n" +
		dim("│") + "  " + gold("███") + dim("╗") + "   " + gold("███") + dim("╗") + "  " + dim("│") + "  " + dim("Chatting with:") + "                              " + dim("│") + "\n" +
		dim("│") + "  " + gold("████") + dim("╗") + " " + gold("████") + dim("║") + "  " + dim("│") + "   " + gold("⌬") + "  " + dim("Formation:") + " " + m.config.FormationID + strings.Repeat(" ", max(0, 29-len(m.config.FormationID))) + dim("│") + "\n" +
		dim("│") + "  " + gold("██") + dim("║╚") + gold("██") + dim("╔╝") + gold("██") + dim("║") + "  " + dim("│") + "   " + gold("⚙︎") + "  " + dim("Server:") + " " + m.config.ServerID + strings.Repeat(" ", max(0, 32-len(m.config.ServerID))) + dim("│") + "\n" +
		dim("│") + "  " + gold("██") + dim("║") + " " + dim("╚═╝") + " " + gold("██") + dim("║") + "  " + dim("│") + "   " + gold("♛") + "  " + dim("User:") + " " + m.config.UserID + strings.Repeat(" ", max(0, 34-len(m.config.UserID))) + dim("│") + "\n" +
		dim("│") + "  " + dim("╚═╝") + "     " + dim("╚═╝") + "  " + dim("│") + "                                              " + dim("│") + "\n" +
		dim("╰──────────────────────────────────────────────────────────────╯") + "\n\n" +
		dim("   ENTER to send • \\ + ENTER for a new line • Ctrl+C to exit") + "\n\n\n\n\n"

	return tea.Println(header)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			// Clear input first, quit on second press
			if strings.TrimSpace(m.textarea.Value()) != "" {
				m.textarea.Reset()
				m.showExitHint = true
				m.exitHintStart = time.Now()
				m.textarea.Placeholder = "Press Ctrl+C again to exit"
				// Start tick to clear the hint after 5 seconds
				return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
					return tickMsg{}
				})
			}
			// If hint is showing or input is empty, quit
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEsc:
			// Clear input first if not empty
			if strings.TrimSpace(m.textarea.Value()) != "" && !m.showCommands && !m.showSubmenu && !m.showHelp {
				m.textarea.Reset()
				return m, nil
			}
			if m.showSubmenu {
				m.showSubmenu = false
				m.submenuSelected = 0
				m.showCommands = true // Go back to main menu
				m.textarea.SetValue("/")
				return m, nil
			}
			if m.showCommands {
				m.showCommands = false
				m.commandSelected = 0
				m.textarea.Reset()
				return m, nil
			}
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.isThinking {
				m.isThinking = false
				m.thinking = []ThinkingStep{} // Clear thinking steps
				m.requestAborted = true       // Flag to ignore incoming response
				// Cancel request on server if we have a request ID
				if m.currentRequestID != "" && m.config.Client != nil {
					go m.config.Client.CancelRequest(m.currentRequestID, m.config.UserID)
					m.currentRequestID = ""
				}
				// Print abort message in red
				return m, printAbortMessage()
			}
			return m, nil

		case tea.KeyUp:
			if m.showSubmenu {
				cmd := m.getCommandByName(m.submenuParent)
				if cmd != nil {
					if m.submenuSelected > 0 {
						m.submenuSelected--
					} else {
						m.submenuSelected = len(cmd.Submenu) - 1
					}
				}
				return m, nil
			}
			if m.showCommands {
				filtered := m.filterCommands(m.textarea.Value())
				if m.commandSelected > 0 {
					m.commandSelected--
				} else {
					m.commandSelected = len(filtered) - 1 // Wrap to bottom
				}
				return m, nil
			}
			// Input history navigation (up = older)
			if len(m.inputHistory) > 0 {
				if m.historyIndex == -1 {
					// Save current input before navigating
					m.currentInput = m.textarea.Value()
					m.historyIndex = len(m.inputHistory) - 1
				} else if m.historyIndex > 0 {
					m.historyIndex--
				}
				m.textarea.SetValue(m.inputHistory[m.historyIndex])
				m.textarea.CursorEnd()
				return m, nil
			}

		case tea.KeyDown:
			if m.showSubmenu {
				cmd := m.getCommandByName(m.submenuParent)
				if cmd != nil {
					if m.submenuSelected < len(cmd.Submenu)-1 {
						m.submenuSelected++
					} else {
						m.submenuSelected = 0
					}
				}
				return m, nil
			}
			if m.showCommands {
				filtered := m.filterCommands(m.textarea.Value())
				if m.commandSelected < len(filtered)-1 {
					m.commandSelected++
				} else {
					m.commandSelected = 0 // Wrap to top
				}
				return m, nil
			}
			// Input history navigation (down = newer)
			if m.historyIndex >= 0 {
				if m.historyIndex < len(m.inputHistory)-1 {
					m.historyIndex++
					m.textarea.SetValue(m.inputHistory[m.historyIndex])
				} else {
					// Back to current input
					m.historyIndex = -1
					m.textarea.SetValue(m.currentInput)
				}
				m.textarea.CursorEnd()
				return m, nil
			}

		case tea.KeyTab:
			if m.showSubmenu {
				// Select submenu item
				cmd := m.getCommandByName(m.submenuParent)
				if cmd != nil && m.submenuSelected < len(cmd.Submenu) {
					m.asyncMode = m.handleSubmenuSelection(cmd, m.submenuSelected)
				}
				m.showSubmenu = false
				m.showCommands = false
				m.submenuSelected = 0
				m.textarea.Reset()
				return m, nil
			}
			if m.showCommands {
				filtered := m.filterCommands(m.textarea.Value())
				if len(filtered) > 0 && m.commandSelected < len(filtered) {
					selected := filtered[m.commandSelected]
					// Check if command has submenu
					if len(selected.Submenu) > 0 {
						m.showSubmenu = true
						m.submenuParent = selected.Name
						m.submenuSelected = 0
						m.showCommands = false
					} else if selected.Name == "/clear" {
						// Execute /clear immediately
						m.showCommands = false
						m.commandSelected = 0
						m.textarea.Reset()
						m.messages = []Message{}
						m.eventHistory = []StreamEvent{}
						m.viewport.SetContent("")
						m.sessionID = ""
						return m, printSystemMessage("Session cleared. New session will start with next message.")
					} else {
						// Select the command
						m.textarea.SetValue(selected.Name + " ")
						m.showCommands = false
						m.commandSelected = 0
					}
				}
				return m, nil
			}

		case tea.KeyCtrlJ:
			// Shift+Enter sends ctrl+j in iTerm2
			m.insertNewline()
			return m, nil

		case tea.KeyEnter:
			// Alt+Enter (Option+Enter) adds newline - works in all terminals
			if msg.Alt {
				m.insertNewline()
				return m, nil
			}
			// Handle submenu selection with Enter
			if m.showSubmenu {
				cmd := m.getCommandByName(m.submenuParent)
				if cmd != nil && m.submenuSelected < len(cmd.Submenu) {
					m.asyncMode = m.handleSubmenuSelection(cmd, m.submenuSelected)
				}
				m.showSubmenu = false
				m.showCommands = false
				m.submenuSelected = 0
				m.textarea.Reset()
				return m, nil
			}
			// Handle command menu selection with Enter
			if m.showCommands {
				filtered := m.filterCommands(m.textarea.Value())
				if len(filtered) > 0 && m.commandSelected < len(filtered) {
					selected := filtered[m.commandSelected]
					if len(selected.Submenu) > 0 {
						m.showSubmenu = true
						m.submenuParent = selected.Name
						m.submenuSelected = 0
						m.showCommands = false
					} else if selected.Name == "/clear" {
						// Execute /clear immediately
						m.showCommands = false
						m.commandSelected = 0
						m.textarea.Reset()
						m.messages = []Message{}
						m.eventHistory = []StreamEvent{}
						m.viewport.SetContent("")
						m.sessionID = ""
						return m, printSystemMessage("Session cleared. New session will start with next message.")
					} else {
						m.textarea.SetValue(selected.Name + " ")
						m.showCommands = false
						m.commandSelected = 0
					}
				}
				return m, nil
			}
			// Backslash + Enter = line continuation (like bash)
			val := m.textarea.Value()
			if strings.HasSuffix(val, "\\") {
				m.textarea.SetValue(strings.TrimSuffix(val, "\\"))
				m.insertNewline()
				return m, nil
			}
			if !m.isThinking && strings.TrimSpace(val) != "" {
				input := m.textarea.Value()
				
				// Handle slash commands
				if strings.HasPrefix(input, "/") {
					return m.handleCommand(input)
				}

				// Clear help and previous thinking steps
				m.showHelp = false
				m.thinking = []ThinkingStep{}

				// Add to input history (avoid duplicates of last entry)
				if len(m.inputHistory) == 0 || m.inputHistory[len(m.inputHistory)-1] != input {
					m.inputHistory = append(m.inputHistory, input)
				}
				m.historyIndex = -1
				m.currentInput = ""
				m.requestAborted = false  // Reset abort flag for new request
				m.currentRequestID = "" // Clear previous request ID
				m.lastError = ""        // Clear previous error

				// Add user message and print above TUI (persists in scrollback)
				m.messages = append(m.messages, Message{
					Role:      "user",
					Content:   input,
					Timestamp: time.Now(),
				})
				m.textarea.Reset()
				m.textarea.SetHeight(1)

				// Start API call
				m.isThinking = true
				m.thinkingStart = time.Now()
				m.streamingText.Reset()
				m.currentEvent = nil
				m.eventChan = make(chan StreamEvent, 10) // New channel for this request
				return m, tea.Batch(printUserMessageAbove(input), m.sendChatMessage(input))
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		if !m.ready {
			m.ready = true
		}

	case thinkingMsg:
		// Ignore if request was aborted
		if m.requestAborted {
			return m, nil
		}
		m.thinking = append(m.thinking, ThinkingStep{
			Text:      string(msg),
			Completed: false,
		})
		return m, nil

	case thinkingCompleteMsg:
		// Ignore if request was aborted
		if m.requestAborted {
			return m, nil
		}
		if int(msg) < len(m.thinking) {
			m.thinking[msg].Completed = true
		}
		return m, nil

	case streamErrorMsg:
		m.isThinking = false
		m.thinking = []ThinkingStep{}
		return m, printServerError(msg.err.Error())

	case streamDoneMsg:
		// Update session ID if provided
		if msg.sessionID != "" {
			m.sessionID = msg.sessionID
		}
		// If we have accumulated text, treat it as response
		if m.streamingText.Len() > 0 {
			response := m.streamingText.String()
			m.streamingText.Reset()
			m.isThinking = false
			m.messages = append(m.messages, Message{
				Role:      "assistant",
				Content:   response,
				Timestamp: time.Now(),
			})
			return m, tea.Sequence(clearThinkingLine(), printAssistantMessageAbove(response))
		}
		m.isThinking = false
		m.currentEvent = nil
		// Check if we have an error message
		if m.lastError != "" {
			errMsg := m.lastError
			m.lastError = ""
			return m, printServerError(errMsg)
		}
		// No response text received
		return m, printServerError("No response received from server")

	case streamEventMsg:
		event := msg.event

		// Always capture request ID (even if aborted, for cancellation)
		if event.RequestID != "" && m.currentRequestID == "" {
			m.currentRequestID = event.RequestID
			// If request was aborted but we just got the ID, send cancel now
			if m.requestAborted && m.config.Client != nil {
				go m.config.Client.CancelRequest(m.currentRequestID, m.config.UserID)
			}
		}

		// Ignore display if request was aborted
		if m.requestAborted {
			return m, m.listenForEvents() // Keep listening to drain channel
		}

		m.currentEvent = &event
		m.eventHistory = append(m.eventHistory, event)

		// Update session ID if provided
		if event.SessionID != "" && m.sessionID == "" {
			m.sessionID = event.SessionID
		}
		// Track error events
		if event.Type == "error" && event.Content != "" {
			m.lastError = event.Content
		}

		// In verbose mode, print each event; otherwise just update currentEvent
		var cmd tea.Cmd
		if m.config.Verbose {
			cmd = printStreamEventAbove(event)
		}

		// Continue listening for more events
		return m, tea.Batch(cmd, m.listenForEvents())

	case streamCompleteMsg:
		// Ignore if request was aborted
		if m.requestAborted {
			m.requestAborted = false
			return m, nil
		}

		// Update session ID if provided
		if msg.sessionID != "" {
			m.sessionID = msg.sessionID
		}

		m.isThinking = false
		m.currentEvent = nil

		// Build commands for printing
		var cmds []tea.Cmd

		// Clear the thinking line before printing response
		cmds = append(cmds, clearThinkingLine())

		// Add assistant message and print above TUI
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   msg.response,
			Timestamp: time.Now(),
		})
		cmds = append(cmds, printAssistantMessageAbove(msg.response))

		// Clear event history for next request
		m.eventHistory = nil

		return m, tea.Sequence(cmds...)

	case streamResponseMsg:
		// Ignore response if request was aborted
		if m.requestAborted {
			m.requestAborted = false
			return m, nil
		}
		// Update session ID if provided
		if msg.sessionID != "" {
			m.sessionID = msg.sessionID
		}
		m.isThinking = false
		// Clear thinking line and print response
		var cmds []tea.Cmd
		cmds = append(cmds, clearThinkingLine())
		for _, step := range msg.thinking {
			cmds = append(cmds, printThinkingStepAbove(step))
		}
		// Add assistant message and print above TUI
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   msg.response,
			Timestamp: time.Now(),
		})
		cmds = append(cmds, printAssistantMessageAbove(msg.response))
		return m, tea.Sequence(cmds...)

	case responseMsg:
		// Ignore response if request was aborted
		if m.requestAborted {
			m.requestAborted = false
			return m, nil
		}
		m.isThinking = false
		// Clear thinking line and print response
		var cmds []tea.Cmd
		cmds = append(cmds, clearThinkingLine())
		for _, step := range m.thinking {
			cmds = append(cmds, printThinkingStepAbove(step.Text))
		}
		m.thinking = []ThinkingStep{} // Clear for next request
		// Add assistant message and print above TUI (persists in scrollback)
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   string(msg),
			Timestamp: time.Now(),
		})
		cmds = append(cmds, printAssistantMessageAbove(string(msg)))
		return m, tea.Sequence(cmds...)

	case tickMsg:
		// Check if exit hint should be cleared (after 3 seconds)
		if m.showExitHint && time.Since(m.exitHintStart) > 3*time.Second {
			m.showExitHint = false
			m.textarea.Placeholder = "Type your message..."
		}

		// Continue ticking if thinking or showing exit hint
		if m.isThinking || m.showExitHint {
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg{}
			})
		}
	}

	// Update textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	// Clear exit hint when user starts typing
	if m.showExitHint && m.textarea.Value() != "" {
		m.showExitHint = false
		m.textarea.Placeholder = "Type your message..."
	}

	// Check for ? as first/only character - show help immediately (not during streaming)
	if m.textarea.Value() == "?" && !m.isThinking {
		m.showHelp = true
		m.showCommands = false
		m.textarea.Reset()
	} else if m.showHelp && m.textarea.Value() != "" {
		// Clear help when user starts typing something else
		m.showHelp = false
	}

	// Check for / command menu
	val := m.textarea.Value()
	if strings.HasPrefix(val, "/") && !m.isThinking && !m.showSubmenu {
		filtered := m.filterCommands(val)
		if len(filtered) > 0 {
			m.showCommands = true
			m.showHelp = false
			// Keep selection in bounds
			if m.commandSelected >= len(filtered) {
				m.commandSelected = len(filtered) - 1
			}
		} else {
			// No matching commands - hide menu
			m.showCommands = false
		}
	} else if !strings.HasPrefix(val, "/") {
		// Clear both menus when / is deleted
		m.showCommands = false
		m.showSubmenu = false
		m.commandSelected = 0
		m.submenuSelected = 0
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model - only renders input area (messages printed above via tea.Println)
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder
	margin := " "

	// Help screen (shown above input when ? is pressed)
	if m.showHelp {
		b.WriteString(m.renderHelp())
	}

	// Thinking indicators (if active) - shown in real-time, printed to history when complete
	if m.isThinking {
		b.WriteString(m.renderThinkingLive())
	}

	// Command menu or submenu (above input)
	if m.showSubmenu {
		b.WriteString(m.renderSubmenu())
	} else if m.showCommands {
		b.WriteString(m.renderCommands())
	}

	// Add spacing before input if there's content above (thinking, menus, help)
	if m.isThinking || m.showHelp || m.showCommands || m.showSubmenu {
		b.WriteString("\n")
	}

	// Input divider
	dividerWidth := m.width - 4
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	b.WriteString(margin)
	b.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n")

	// Input area
	inputLines := strings.Split(m.textarea.View(), "\n")
	for i, line := range inputLines {
		b.WriteString(margin)
		if i == 0 {
			b.WriteString(goldStyle.Render(">"))
			b.WriteString("  ")
		} else {
			b.WriteString("   ")
		}
		b.WriteString(line)
		if i < len(inputLines)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Status bar
	b.WriteString(margin)
	b.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n")
	b.WriteString(margin)
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n") // Space after input footer

	return b.String()
}

// renderThinkingLive renders the current thinking state (current event + spinner)
func (m Model) renderThinkingLive() string {
	var b strings.Builder

	// Show spinner with elapsed time
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	elapsed := time.Since(m.thinkingStart)
	frame := frames[int(elapsed.Milliseconds()/100)%len(frames)]

	// Very dim ESC hint
	veryDimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	escHint := veryDimStyle.Render("(ESC to cancel)")

	// If we have a current event from streaming, show it
	if m.currentEvent != nil {
		event := *m.currentEvent
		italicStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#808080"))
		contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

		// Calculate wrap width (terminal width - indent)
		wrapWidth := m.width - 8
		if wrapWidth < 40 {
			wrapWidth = 40
		}

		switch event.Type {
		case "thinking":
			b.WriteString(" " + frame + " ")
			b.WriteString(italicStyle.Render("Thinking:"))
			b.WriteString("  ")
			b.WriteString(escHint)
			b.WriteString("\n")
			// Wrap the content
			wrapped := wrapText(event.Content, wrapWidth)
			for i, line := range wrapped {
				if i == 0 {
					b.WriteString(contentStyle.Render("    ↳ " + line))
				} else {
					b.WriteString(contentStyle.Render("      " + line))
				}
				b.WriteString("\n")
			}

		case "planning":
			b.WriteString(" " + frame + " ")
			b.WriteString(italicStyle.Render("Planning:"))
			b.WriteString("  ")
			b.WriteString(escHint)
			b.WriteString("\n")
			// Wrap the content
			wrapped := wrapText(event.Content, wrapWidth)
			for i, line := range wrapped {
				if i == 0 {
					b.WriteString(contentStyle.Render("    ↳ " + line))
				} else {
					b.WriteString(contentStyle.Render("      " + line))
				}
				b.WriteString("\n")
			}

		case "progress":
			b.WriteString(" " + frame + " ")
			if event.Stage == "tool_call" && event.ToolName != "" {
				toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c98b45"))
				b.WriteString(contentStyle.Render("Using "))
				b.WriteString(toolStyle.Render(event.ToolName))
				b.WriteString(contentStyle.Render(" tool..."))
			} else {
				b.WriteString(contentStyle.Render(event.Content))
			}
			b.WriteString("  ")
			b.WriteString(escHint)
			b.WriteString("\n")

		default:
			b.WriteString(" " + frame + " ")
			b.WriteString(contentStyle.Render(event.Content))
			b.WriteString("  ")
			b.WriteString(escHint)
			b.WriteString("\n")
		}
	} else {
		// No event yet, show generic thinking
		b.WriteString(" ")
		b.WriteString(thinkingStyle.Render(fmt.Sprintf("%s  Thinking... %.1fs", frame, elapsed.Seconds())))
		b.WriteString("  ")
		b.WriteString(escHint)
		b.WriteString("\n")
	}

	return b.String()
}

// printMessageAbove returns a command to print a message above the TUI (persists in scrollback)
func printUserMessageAbove(content string) tea.Cmd {
	// Empty line before, message, empty line after
	return tea.Println("\n " + goldStyle.Render(">") + "  " + userStyle.Render(content) + "\n")
}

func printAssistantMessageAbove(content string) tea.Cmd {
	// Remove horizontal rules from content (server sometimes sends these)
	content = removeHorizontalRules(content)

	// Check if content has markdown formatting
	hasMarkdown := strings.Contains(content, "```") ||
		strings.Contains(content, "**") ||
		strings.Contains(content, "##") ||
		strings.Contains(content, "- ") ||
		strings.Contains(content, "1. ")

	var rendered string
	if hasMarkdown {
		// Use custom style without horizontal rules
		// Wrap at 72 to account for 4-char indentation (keeps total under 80)
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStylesFromJSONBytes(chatStyle),
			glamour.WithWordWrap(72),
		)
		if err == nil {
			if r, err := renderer.Render(content); err == nil {
				rendered = strings.TrimSpace(r)
				// Strip trailing spaces from each line (glamour pads to terminal width)
				lines := strings.Split(rendered, "\n")
				for i, line := range lines {
					line = strings.TrimRight(line, " ")
					if i == 0 {
						lines[i] = "  " + line // 2 spaces after 𝐌
					} else {
						lines[i] = "    " + line // 4 spaces to align with content after "𝐌  "
					}
				}
				rendered = strings.Join(lines, "\n")
			} else {
				rendered = indentContent(content)
			}
		} else {
			rendered = indentContent(content)
		}
	} else {
		rendered = indentContent(content)
	}
	// Print with clear line sequences to prevent TUI bleed-through
	// Each line ends with clear-to-end-of-line to prevent divider bleeding
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		lines[i] = line + "\033[K" // Clear to end of line after each line
	}
	rendered = strings.Join(lines, "\n")
	return tea.Println(" " + goldStyle.Render("𝐌") + rendered + "\n")
}

// printThinkingStepAbove prints a completed thinking step to history (no trailing space)
func printThinkingStepAbove(text string) tea.Cmd {
	return tea.Println(" " + completedStyle.Render("⏺  "+text))
}

// printStreamEventAbove prints a stream event to history (for verbose mode)
func printStreamEventAbove(event StreamEvent) tea.Cmd {
	return tea.Println(formatStreamEvent(event, false))
}

// formatStreamEvent formats a stream event for display
func formatStreamEvent(event StreamEvent, withSpinner bool) string {
	italicStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#808080"))
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

	prefix := " ⏺ "
	if withSpinner {
		// Use spinner frame based on time
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := frames[int(time.Now().UnixMilli()/80)%len(frames)]
		prefix = " " + frame + " "
	}

	switch event.Type {
	case "thinking":
		// Thinking: (italic)
		//   ↳ content
		header := italicStyle.Render("Thinking:")
		content := contentStyle.Render("  ↳ " + event.Content)
		return prefix + header + "\n" + content

	case "planning":
		// Planning: (italic)
		//   ↳ content
		header := italicStyle.Render("Planning:")
		content := contentStyle.Render("  ↳ " + event.Content)
		return prefix + header + "\n" + content

	case "progress":
		// For tool_call, highlight the tool name
		if event.Stage == "tool_call" && event.ToolName != "" {
			toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c98b45"))
			return prefix + contentStyle.Render("Using ") + toolStyle.Render(event.ToolName) + contentStyle.Render(" tool...")
		}
		// Regular progress
		return prefix + contentStyle.Render(event.Content)

	default:
		return prefix + contentStyle.Render(event.Content)
	}
}

// indentContent adds proper indentation and wrapping to multi-line content
func indentContent(content string) string {
	// First wrap long lines at 72 chars (accounts for 4-char indent + 4-char prefix)
	var wrappedLines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, " ")
		if len(line) <= 72 {
			wrappedLines = append(wrappedLines, line)
		} else {
			// Wrap long lines
			wrappedLines = append(wrappedLines, wrapText(line, 72)...)
		}
	}
	
	// Add indentation
	for i, line := range wrappedLines {
		if i == 0 {
			wrappedLines[i] = "  " + line // 2 spaces after 𝐌
		} else {
			wrappedLines[i] = "    " + line // 4 spaces to align with content after "𝐌  "
		}
	}
	return strings.Join(wrappedLines, "\n")
}

// removeHorizontalRules strips markdown horizontal rules from content
func removeHorizontalRules(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip lines that are just dashes, asterisks, or underscores (3+ chars)
		if len(trimmed) >= 3 {
			allDashes := strings.Trim(trimmed, "-") == ""
			allAsterisks := strings.Trim(trimmed, "*") == ""
			allUnderscores := strings.Trim(trimmed, "_") == ""
			if allDashes || allAsterisks || allUnderscores {
				continue
			}
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// generateSessionID creates a unique session ID if none provided
func generateSessionID(provided string) string {
	if provided != "" {
		return provided
	}
	// Generate: session_<unix_timestamp_ms>_<random>
	return fmt.Sprintf("session_%d_%d", time.Now().UnixMilli(), time.Now().UnixNano()%10000)
}

// truncateForDebug truncates a string for debug output
func truncateForDebug(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// printAbortMessage prints the abort message in red
func printAbortMessage() tea.Cmd {
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
	return tea.Println(" " + redStyle.Render("✕  Request aborted by user") + "\n")
}

// printServerError prints a server error message in yellow/warning style
func printServerError(err string) tea.Cmd {
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00"))
	return tea.Println(" " + warnStyle.Render("⚠  Server returned an error:") + "\n    " + err + "\n")
}

// printSystemMessage prints a system message in dim style
func printSystemMessage(msg string) tea.Cmd {
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	return tea.Println("\n\n " + dimStyle.Render("ℹ  "+msg) + "\n\n\n")
}

// clearThinkingLine clears the thinking area before printing response
func clearThinkingLine() tea.Cmd {
	// Clear current line fully
	return tea.Printf("\033[2K\r")
}

func (m Model) renderHeader() string {
	gold := goldStyle.Render
	dim := dimmedStyle.Render

	// Fixed 64-char box
	// Left: │ + 2 + logo(11) + 2 + │ = 17 chars
	// Right: 64 - 17 - 1 = 46 chars
	rightWidth := 46

	// M logo lines (11 chars each)
	mLine1 := gold("███") + dim("╗") + "   " + gold("███") + dim("╗")
	mLine2 := gold("████") + dim("╗") + " " + gold("████") + dim("║")
	mLine3 := gold("██") + dim("║╚") + gold("██") + dim("╔╝") + gold("██") + dim("║")
	mLine4 := gold("██") + dim("║") + " " + dim("╚═╝") + " " + gold("██") + dim("║")
	mLine5 := dim("╚═╝") + "     " + dim("╚═╝")

	var b strings.Builder

	// Top border: ╭─── MUXI Chat ─────...─╮ (64 chars)
	b.WriteString(dim("╭─── ") + gold("MUXI Chat") + dim(" " + strings.Repeat("─", 48) + "╮"))
	b.WriteString("\n")

	// Line 1: empty (64 chars)
	b.WriteString(dim("│") + "               " + dim("│") + strings.Repeat(" ", 46) + dim("│") + "\n")

	// Lines 2-6: logo + content
	// Build right content with explicit padding (46 chars each)
	// Use rune count for visual width (Unicode symbols are 1 visual char)
	runeLen := func(s string) int {
		return len([]rune(s))
	}
	fmtLine := func(prefix, label, value string, totalWidth int) string {
		content := prefix + label + value
		visualLen := runeLen(content)
		if visualLen < totalWidth {
			return content + strings.Repeat(" ", totalWidth-visualLen)
		}
		return string([]rune(content)[:totalWidth])
	}

	mLines := []string{mLine1, mLine2, mLine3, mLine4, mLine5}
	rightContents := []string{
		fmtLine("  ", "Chatting with:", "", rightWidth),
		fmtLine("    ", "⌬ Formation: ", m.config.FormationID, rightWidth),
		fmtLine("    ", "⏍ Server: ", m.config.ServerID, rightWidth),
		fmtLine("    ", "♛ User: ", m.config.UserID, rightWidth),
		strings.Repeat(" ", rightWidth),
	}

	for i, mLine := range mLines {
		b.WriteString(dim("│") + "  ")
		b.WriteString(mLine)
		b.WriteString("  " + dim("│"))
		if i == 0 {
			b.WriteString(dim(rightContents[i]))
		} else {
			b.WriteString(rightContents[i])
		}
		b.WriteString(dim("│") + "\n")
	}

	// Bottom border (64 chars)
	b.WriteString(dim("╰" + strings.Repeat("─", 62) + "╯"))
	b.WriteString("\n")

	// Hint text (left-aligned)
	hint := "ENTER to send • \\ + ENTER for a new line • Ctrl+C to exit"
	b.WriteString("\n")
	b.WriteString(dim("   " + hint))

	return b.String()
}

func (m Model) renderMessages() string {
	var b strings.Builder
	margin := " "

	// Initial padding to push header down (12 lines from input area)
	b.WriteString(strings.Repeat("\n", 12))

	// Header (scrolls with content)
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Spacing below header
	b.WriteString(strings.Repeat("\n", 5))

	for i, msg := range m.messages {
		// Add 2 line breaks between messages (not before first)
		if i > 0 {
			b.WriteString("\n\n")
		}

		if msg.Role == "user" {
			// Add margin to each line of user message
			lines := strings.Split(msg.Content, "\n")
			for j, line := range lines {
				b.WriteString(margin)
				if j == 0 {
					b.WriteString(userStyle.Render(">  " + line))
				} else {
					b.WriteString(userStyle.Render("   " + line))
				}
				b.WriteString("\n")
			}
			
			// Show thinking after user message (before response)
			// Check if next message is assistant or if we're still thinking
			isLastMsg := i == len(m.messages)-1
			nextIsAssistant := !isLastMsg && m.messages[i+1].Role == "assistant"
			if (isLastMsg || nextIsAssistant) && (m.isThinking || len(m.thinking) > 0) {
				b.WriteString("\n")
				b.WriteString(m.renderThinking())
			}
		} else {
			// Render markdown for assistant messages
			rendered, err := m.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			// Add M prefix on first line, then align rest
			b.WriteString(" ")
			b.WriteString(goldStyle.Render("𝐌"))
			
			lines := strings.Split(strings.TrimSpace(rendered), "\n")
			for j, line := range lines {
				if j > 0 {
					b.WriteString("  ") // Align with " 𝐌"
				}
				b.WriteString(strings.TrimLeft(line, " "))
				b.WriteString("\n")
			}
		}
	}

	// Show help screen (temporary, clears on next message)
	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(m.renderHelp())
	}

	return b.String()
}

func (m Model) renderHelp() string {
	dim := dimmedStyle.Render
	key := userStyle.Render // #c98b45 for keys
	
	var b strings.Builder
	b.WriteString(dim("╭───────────────────────────────────────────────────────────╮") + "\n")
	b.WriteString(dim("│  Basics                                                   │") + "\n")
	b.WriteString(dim("│  Send                                              ") + key("Enter") + dim("  │") + "\n")
	b.WriteString(dim("│  New line                                      ") + key("\\ + Enter") + dim("  │") + "\n")
	b.WriteString(dim("│  Cancel request                                      ") + key("Esc") + dim("  │") + "\n")
	b.WriteString(dim("│  History or line navigation                          ") + key("↑/↓") + dim("  │") + "\n")
	b.WriteString(dim("╰───────────────────────────────────────────────────────────╯") + "\n")
	b.WriteString(dim("╭───────────────────────────────────────────────────────────╮") + "\n")
	b.WriteString(dim("│  Commands                                              ") + key("/") + dim("  │") + "\n")
	b.WriteString(dim("│  Exit chat                                        ") + key("Ctrl+C") + dim("  │") + "\n")
	b.WriteString(dim("╰───────────────────────────────────────────────────────────╯") + "\n")
	b.WriteString(dim("╭───────────────────────────────────────────────────────────╮") + "\n")
	b.WriteString(dim("│  Navigation                                               │") + "\n")
	b.WriteString(dim("│  Jump to line start/end                        ") + key("Cmd + ←/→") + dim("  │") + "\n")
	b.WriteString(dim("│  Delete word                             ") + key("Option + Delete") + dim("  │") + "\n")
	b.WriteString(dim("│  Delete line                                ") + key("Cmd + Delete") + dim("  │") + "\n")
	b.WriteString(dim("╰───────────────────────────────────────────────────────────╯") + "\n")
	
	return b.String()
}

func (m Model) filterCommands(input string) []Command {
	var filtered []Command
	input = strings.ToLower(input)
	for _, cmd := range commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), input) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

func (m Model) renderCommands() string {
	dim := dimmedStyle.Render
	sel := lipgloss.NewStyle().Foreground(lipgloss.Color("#c98b45")).Bold(true).Render
	
	filtered := m.filterCommands(m.textarea.Value())
	if len(filtered) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(dim("╭──────────────────────────────────────────────────────────────╮") + "\n")
	
	for i, cmd := range filtered {
		prefix := "  "
		if i == m.commandSelected {
			prefix = "> "
		}
		
		// Format: prefix + command name (padded) + description
		name := cmd.Name
		desc := cmd.Description
		// Truncate description if too long
		maxDesc := 40
		if len(desc) > maxDesc {
			desc = desc[:maxDesc-3] + "..."
		}
		
		line := fmt.Sprintf("%-12s  %-42s", name, desc)
		if i == m.commandSelected {
			b.WriteString(dim("│ ") + sel(prefix+line) + dim("   │") + "\n")
		} else {
			b.WriteString(dim("│ "+prefix) + fmt.Sprintf("%-12s", name) + dim("  "+fmt.Sprintf("%-42s", desc)) + dim("   │") + "\n")
		}
	}
	
	b.WriteString(dim("╰──────────────────────────────────────────────────────────────╯") + "\n")
	b.WriteString(dim("  ↑/↓ navigate • Tab/Enter select • Esc cancel") + "\n")
	
	return b.String()
}

func (m Model) getCommandByName(name string) *Command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func (m Model) handleSubmenuSelection(cmd *Command, idx int) string {
	if cmd.Name == "/async" && idx < len(cmd.Submenu) {
		return cmd.Submenu[idx].Name
	}
	return m.asyncMode
}

func (m Model) renderSubmenu() string {
	dim := dimmedStyle.Render
	sel := lipgloss.NewStyle().Foreground(lipgloss.Color("#c98b45")).Bold(true).Render
	
	cmd := m.getCommandByName(m.submenuParent)
	if cmd == nil || len(cmd.Submenu) == 0 {
		return ""
	}

	var b strings.Builder
	// Box width: 44 chars total (42 dashes + 2 corners)
	b.WriteString(dim("╭──────────────────────────────────────────╮") + "\n")
	b.WriteString(dim("│") + fmt.Sprintf("  %-40s", cmd.Name) + dim("│") + "\n")
	b.WriteString(dim("├──────────────────────────────────────────┤") + "\n")
	
	for i, item := range cmd.Submenu {
		prefix := "  "
		if i == m.submenuSelected {
			prefix = "> "
		}
		
		content := fmt.Sprintf("%-10s %-27s", item.Name, item.Description)
		if i == m.submenuSelected {
			b.WriteString(dim("│") + sel(" "+prefix+content) + dim(" │") + "\n")
		} else {
			b.WriteString(dim("│ "+prefix+content+" │") + "\n")
		}
	}
	
	b.WriteString(dim("╰──────────────────────────────────────────╯") + "\n")
	b.WriteString(dim("  ↑/↓ navigate • Tab/Enter select • Esc cancel") + "\n")
	
	return b.String()
}

func (m Model) renderThinking() string {
	var b strings.Builder
	margin := " "
	
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	elapsed := time.Since(m.thinkingStart)
	frame := frames[int(elapsed.Milliseconds()/100)%len(frames)]

	for _, step := range m.thinking {
		b.WriteString(margin)
		if step.Completed {
			b.WriteString(completedStyle.Render("⏺  " + step.Text))
		} else {
			b.WriteString(thinkingStyle.Render(frame + "  " + step.Text))
		}
		b.WriteString("\n")
	}

	// Current thinking with elapsed time (only when actively thinking)
	if m.isThinking && (len(m.thinking) == 0 || m.thinking[len(m.thinking)-1].Completed) {
		b.WriteString(margin)
		b.WriteString(thinkingStyle.Render(fmt.Sprintf("%s  Thinking... %.1fs  (ESC to stop)", frame, elapsed.Seconds())))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	dim := dimmedStyle.Render
	gold := goldStyle.Render
	bold := lipgloss.NewStyle().Bold(true).Render

	// Async indicator
	var indicator, indicatorVisual string
	switch m.asyncMode {
	case "on":
		indicator = gold("A")
		indicatorVisual = "A"
	case "auto":
		indicator = gold("⚡")
		indicatorVisual = "⚡"
	default:
		indicator = dim("S")
		indicatorVisual = "S"
	}

	left := fmt.Sprintf("%s@%s://%s",
		m.config.UserID,
		m.config.ServerID,
		m.config.FormationID,
	)

	// Visual length of right side (without ANSI codes)
	rightVisual := indicatorVisual + " • ? for help • / for commands"
	right := indicator + dim(" • ") + bold("?") + dim(" for help • ") + bold("/") + dim(" for commands")

	gap := m.width - len(left) - len(rightVisual) + 1
	if gap < 0 {
		gap = 1
	}

	return dim(left) + strings.Repeat(" ", gap) + right
}

func (m *Model) insertNewline() {
	m.textarea.InsertString("\n")
	// Grow textarea height as needed (max 10 lines)
	lines := strings.Count(m.textarea.Value(), "\n") + 1
	if lines > m.textarea.Height() && lines <= 10 {
		m.textarea.SetHeight(lines)
	}
}

func (m Model) handleCommand(input string) (tea.Model, tea.Cmd) {
	cmd := strings.TrimPrefix(input, "/")
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "exit", "quit", "q":
		m.quitting = true
		return m, tea.Quit

	case "clear":
		m.messages = []Message{}
		m.eventHistory = []StreamEvent{}
		m.viewport.SetContent("")
		m.textarea.Reset()
		m.textarea.SetHeight(1)
		m.sessionID = "" // Clear session ID - server will generate new one on next message
		return m, printSystemMessage("Session cleared. New session will start with next message.")

	case "agents":
		// TODO: Fetch from API
		agentList := `**Available Agents:**
- weather-agent (specialist) - Weather information
- code-helper (specialist) - Code assistance  
- general-assistant (generalist) - General queries`
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   agentList,
			Timestamp: time.Now(),
		})
		m.textarea.Reset()
		m.textarea.SetHeight(1)

	default:
		errMsg := fmt.Sprintf("Unknown command: `/%s`. Press `/` to see available commands.", parts[0])
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   errMsg,
			Timestamp: time.Now(),
		})
		m.textarea.Reset()
		m.textarea.SetHeight(1)
	}

	return m, nil
}

// Message types for tea.Cmd
type thinkingMsg string
type thinkingCompleteMsg int
type responseMsg string
type tickMsg struct{}
type streamTokenMsg string
type streamDoneMsg struct{ sessionID string }
type streamErrorMsg struct{ err error }
type streamResponseMsg struct {
	thinking  []string
	response  string
	sessionID string
}
type streamEventMsg struct {
	event     StreamEvent
	sessionID string
}
type streamCompleteMsg struct {
	response  string
	sessionID string
}
type streamBodyMsg struct {
	body io.ReadCloser
}

// sendChatMessage sends a message to the chat API and streams the response
func (m Model) sendChatMessage(message string) tea.Cmd {
	eventChan := m.eventChan // Capture the channel
	debug := m.config.Debug

	return func() tea.Msg {
		if m.config.Client == nil {
			return streamErrorMsg{err: fmt.Errorf("no API client configured")}
		}

		req := &formation.ChatRequest{
			Message:   message,
			SessionID: m.sessionID,
			Stream:    true,
		}
		// Set UseAsync based on mode (nil for auto lets server decide)
		if m.asyncMode == "on" {
			t := true
			req.UseAsync = &t
		} else if m.asyncMode == "off" {
			f := false
			req.UseAsync = &f
		}

		resp, err := m.config.Client.ChatStream(req, m.config.UserID)
		if err != nil {
			return streamErrorMsg{err: err}
		}

		// Start processing stream and sending events through channel
		go processStreamWithEvents(resp.Body, eventChan, debug)

		// Return command to listen for first event
		return waitForEventFromChan(eventChan)
	}
}

// waitForEventFromChan waits for the next stream event from a channel
func waitForEventFromChan(eventChan chan StreamEvent) tea.Msg {
	event, ok := <-eventChan
	if !ok {
		// Channel closed, stream complete
		return streamDoneMsg{}
	}

	if event.Type == "completed" {
		return streamCompleteMsg{
			response:  event.Content,
			sessionID: event.SessionID,
		}
	}

	// Return event message for model to process
	return streamEventMsg{event: event}
}

// listenForEvents returns a command to continue listening for events
func (m Model) listenForEvents() tea.Cmd {
	eventChan := m.eventChan // Capture the channel
	return func() tea.Msg {
		return waitForEventFromChan(eventChan)
	}
}

// MuxiToken represents the MUXI streaming token format
type MuxiToken struct {
	Token struct {
		RequestID string `json:"request_id"`
		UserID    string `json:"user_id"`
		SessionID string `json:"session_id"`
		Type      string `json:"type"`    // progress, thinking, planning, completed, text, response
		Content   string `json:"content"` // The actual content/message
		Stage     string `json:"stage"`
		AgentName string `json:"agent_name"`
		AgentUsed string `json:"agent_used"`
		ToolName  string `json:"tool_name"`
		Status    string `json:"status"`
	} `json:"token"`
}

// processStreamWithEvents reads SSE events and sends them through a channel
func processStreamWithEvents(body io.ReadCloser, eventChan chan StreamEvent, debug bool) {
	defer body.Close()
	defer close(eventChan)

	scanner := bufio.NewScanner(body)
	var fullResponse strings.Builder
	var sessionID string
	var requestID string

	// Debug file for reliable logging (bypasses TUI buffering)
	var debugFile *os.File
	if debug {
		debugFile, _ = os.OpenFile("/tmp/muxi-chat-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if debugFile != nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "=== Stream started at %s ===\n", time.Now().Format(time.RFC3339))
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if debug && debugFile != nil {
			fmt.Fprintf(debugFile, "RAW: %s\n", line)
		}

		// Handle event type lines
		if strings.HasPrefix(line, "event:") {
			eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			if eventType == "done" {
				break
			}
			continue
		}

		// Handle data lines
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		// Try parsing as MUXI token format: {"token": {...}}
		var muxiToken MuxiToken
		if err := json.Unmarshal([]byte(data), &muxiToken); err == nil && muxiToken.Token.Type != "" {
			token := muxiToken.Token

			// Capture session ID and request ID if provided
			if token.SessionID != "" {
				sessionID = token.SessionID
			}
			if token.RequestID != "" {
				requestID = token.RequestID
			}

			if debug && debugFile != nil {
				fmt.Fprintf(debugFile, "PARSED: type=%s stage=%s content=%q requestID=%s\n", token.Type, token.Stage, token.Content, token.RequestID)
			}

			switch token.Type {
			case "content", "text", "response":
				// Actual response text - accumulate
				if token.Content != "" {
					fullResponse.WriteString(token.Content)
				}
			case "completed":
				// Final response
				response := token.Content
				if response == "" || response == "done" {
					response = fullResponse.String()
				}
				if debug && debugFile != nil {
					fmt.Fprintf(debugFile, "COMPLETED: response=%q sessionID=%q\n", truncateForDebug(response, 100), sessionID)
				}
				eventChan <- StreamEvent{
					Type:      "completed",
					Content:   response,
					SessionID: sessionID,
					RequestID: requestID,
				}
				return
			case "progress", "thinking", "planning":
				// Send as event
				if token.Content != "" {
					eventChan <- StreamEvent{
						Type:      token.Type,
						Stage:     token.Stage,
						Content:   token.Content,
						ToolName:  token.ToolName,
						SessionID: sessionID,
						RequestID: requestID,
					}
				}
			case "error":
				eventChan <- StreamEvent{
					Type:      "error",
					Stage:     token.Stage,
					Content:   token.Content,
					RequestID: requestID,
				}
				return
			}
			continue
		}

		// Try parsing as finished marker
		var finished struct {
			Finished bool `json:"finished"`
		}
		if err := json.Unmarshal([]byte(data), &finished); err == nil && finished.Finished {
			// Send accumulated response if any
			if fullResponse.Len() > 0 {
				eventChan <- StreamEvent{
					Type:    "completed",
					Content: fullResponse.String(),
				}
			}
			return
		}
	}

	// Stream ended, send any accumulated response
	if fullResponse.Len() > 0 {
		eventChan <- StreamEvent{
			Type:    "completed",
			Content: fullResponse.String(),
		}
	}
}

// processStream reads SSE events and returns the first token or completion
func processStream(body io.ReadCloser) tea.Msg {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var fullResponse strings.Builder
	var sessionID string
	var lastContent string
	var thinkingSteps []string

	for scanner.Scan() {
		line := scanner.Text()

		// Handle event type lines
		if strings.HasPrefix(line, "event:") {
			eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			if eventType == "done" {
				break
			}
			continue
		}

		// Handle data lines
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		// Try parsing as MUXI token format: {"token": {...}}
		var muxiToken MuxiToken
		if err := json.Unmarshal([]byte(data), &muxiToken); err == nil && muxiToken.Token.Type != "" {
			token := muxiToken.Token
			if token.SessionID != "" {
				sessionID = token.SessionID
			}

			switch token.Type {
			case "content", "text", "response":
				// Actual response text
				if token.Content != "" {
					fullResponse.WriteString(token.Content)
				}
			case "completed":
				// Check if content has the actual response
				if token.Content != "" && token.Content != "done" {
					lastContent = token.Content
				}
			case "progress", "thinking", "planning":
				// Collect thinking/progress steps to display
				if token.Content != "" {
					thinkingSteps = append(thinkingSteps, token.Content)
				}
			}
			continue
		}

		// Try parsing as simple token: {"token": "..."}
		var simpleToken struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(data), &simpleToken); err == nil && simpleToken.Token != "" {
			fullResponse.WriteString(simpleToken.Token)
			continue
		}

		// Try parsing as finished marker
		var finished struct {
			Finished bool `json:"finished"`
		}
		if err := json.Unmarshal([]byte(data), &finished); err == nil && finished.Finished {
			break
		}

		// Try parsing as initial envelope
		var envelope struct {
			Data struct {
				StreamStarted bool   `json:"stream_started"`
				SessionID     string `json:"session_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(data), &envelope); err == nil {
			if envelope.Data.SessionID != "" {
				sessionID = envelope.Data.SessionID
			}
			continue
		}

		// Try parsing as OpenAI-style delta
		var oaiChunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &oaiChunk); err == nil && len(oaiChunk.Choices) > 0 {
			if content := oaiChunk.Choices[0].Delta.Content; content != "" {
				fullResponse.WriteString(content)
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return streamErrorMsg{err: err}
	}

	// Return full response if we got text tokens
	if fullResponse.Len() > 0 {
		return streamResponseMsg{
			thinking:  thinkingSteps,
			response:  fullResponse.String(),
			sessionID: sessionID,
		}
	}

	// Fall back to lastContent if no text tokens were received
	if lastContent != "" {
		return streamResponseMsg{
			thinking:  thinkingSteps,
			response:  lastContent,
			sessionID: sessionID,
		}
	}

	// No response text received - this might mean the API doesn't stream text
	return streamDoneMsg{sessionID: sessionID}
}

// Run starts the chat UI
func Run(cfg Config) error {
	p := tea.NewProgram(
		New(cfg),
		// No WithAltScreen() - renders everything in main buffer for scrollback
	)

	_, err := p.Run()
	return err
}
