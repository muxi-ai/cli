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
	GroupID     string
	Client      *formation.Client
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
	requestAborted  bool            // Flag to ignore response after abort
	sessionID       string          // Current session ID (from API response)
	streamingText   strings.Builder // Accumulated streaming response
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
	{"/clear", "Start a new session (clears context)", nil},
	{"/config", "Show current configuration", nil},
	{"/cost", "Show token usage for the session", nil},
	{"/exit", "Exit the chat", nil},
	{"/help", "Show help information", nil},
	{"/history", "Show conversation history", nil},
	{"/model", "Show or change the current model", nil},
	{"/session", "Show session information", nil},
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
		glamour.WithAutoStyle(),
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
		sessionID:    cfg.SessionID,
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
				m.requestAborted = false // Reset abort flag for new request

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
			return m, printAssistantMessageAbove(response)
		}
		m.isThinking = false
		// No response text received - request completed but server didn't stream text
		return m, printServerError("Request completed but no response text was streamed. The server may not be configured for text streaming.")

	case responseMsg:
		// Ignore response if request was aborted
		if m.requestAborted {
			m.requestAborted = false
			return m, nil
		}
		m.isThinking = false
		// Print all thinking steps in order, then the response
		var cmds []tea.Cmd
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

// renderThinkingLive renders the current thinking state (completed steps + spinner)
func (m Model) renderThinkingLive() string {
	var b strings.Builder
	margin := " "

	// Show completed thinking steps first (in order)
	for _, step := range m.thinking {
		if step.Completed {
			b.WriteString(margin)
			b.WriteString(completedStyle.Render("⏺  " + step.Text))
			b.WriteString("\n")
		}
	}

	// Add line space before spinner if there are completed steps above
	hasCompleted := false
	for _, step := range m.thinking {
		if step.Completed {
			hasCompleted = true
			break
		}
	}
	if hasCompleted {
		b.WriteString("\n")
	}

	// Show spinner with elapsed time
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	elapsed := time.Since(m.thinkingStart)
	frame := frames[int(elapsed.Milliseconds()/100)%len(frames)]

	// Show current thinking with spinner and ESC hint (very dim)
	veryDimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	escHint := veryDimStyle.Render("(ESC to cancel)")
	b.WriteString(margin)
	b.WriteString(thinkingStyle.Render(fmt.Sprintf("%s  Thinking... %.1fs", frame, elapsed.Seconds())))
	b.WriteString("  ")
	b.WriteString(escHint)
	b.WriteString("\n")

	return b.String()
}

// printMessageAbove returns a command to print a message above the TUI (persists in scrollback)
func printUserMessageAbove(content string) tea.Cmd {
	// Empty line before, message, empty line after
	return tea.Println("\n " + goldStyle.Render(">") + "  " + userStyle.Render(content) + "\n")
}

func printAssistantMessageAbove(content string) tea.Cmd {
	// Render markdown with dark style (avoid terminal color query that causes escape sequence leak)
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(76), // Reduced to account for indentation
	)
	rendered := content
	if err == nil {
		if r, err := renderer.Render(content); err == nil {
			rendered = strings.TrimSpace(r)
			// Indent each line to align with other content
			lines := strings.Split(rendered, "\n")
			for i, line := range lines {
				if i == 0 {
					// First line: no extra indent (follows 𝐌 directly)
					lines[i] = line
				} else {
					lines[i] = "  " + line // 2 spaces for subsequent lines
				}
			}
			rendered = strings.Join(lines, "\n")
		}
	}
	// Empty line before, message, TWO empty lines after (space before input)
	return tea.Println("\n " + goldStyle.Render("𝐌") + rendered + "\n\n")
}

// printThinkingStepAbove prints a completed thinking step to history (no trailing space)
func printThinkingStepAbove(text string) tea.Cmd {
	return tea.Println(" " + completedStyle.Render("⏺  "+text))
}

// printAbortMessage prints the abort message in red
func printAbortMessage() tea.Cmd {
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
	return tea.Println("\n " + redStyle.Render("✕  Request aborted by user") + "\n")
}

// printServerError prints a server error message in yellow/warning style
func printServerError(err string) tea.Cmd {
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00"))
	return tea.Println("\n " + warnStyle.Render("⚠  Server returned an error:") + "\n    " + err + "\n")
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
		m.viewport.SetContent("")
		m.textarea.Reset()
		m.textarea.SetHeight(1)

	case "help", "?":
		helpText := `**Available Commands:**
- /exit, /quit, /q - Exit chat
- /clear - Clear screen
- /help, /? - Show this help
- /agents - List available agents
- /session - Show session ID
- /new - Start new session

**Keyboard Shortcuts:**
- Enter - Send message
- ESC - Cancel current response
- Ctrl+C - Exit
- Ctrl+L - Clear screen`
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   helpText,
			Timestamp: time.Now(),
		})
		m.textarea.Reset()
		m.textarea.SetHeight(1)

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
		errMsg := fmt.Sprintf("Unknown command: `/%s`. Type `/help` for available commands.", parts[0])
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

// sendChatMessage sends a message to the chat API and streams the response
func (m Model) sendChatMessage(message string) tea.Cmd {
	return func() tea.Msg {
		if m.config.Client == nil {
			return streamErrorMsg{err: fmt.Errorf("no API client configured")}
		}

		req := &formation.ChatRequest{
			Message:   message,
			SessionID: m.sessionID,
			GroupID:   m.config.GroupID,
			Stream:    true,
		}

		resp, err := m.config.Client.ChatStream(req, m.config.UserID)
		if err != nil {
			return streamErrorMsg{err: err}
		}

		// Process stream in a goroutine and return tokens via channel
		return processStream(resp.Body)
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
		Status    string `json:"status"`
	} `json:"token"`
}

// processStream reads SSE events and returns the first token or completion
func processStream(body io.ReadCloser) tea.Msg {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var fullResponse strings.Builder
	var sessionID string
	var lastContent string

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
			case "text", "response":
				// Actual response text
				if token.Content != "" {
					fullResponse.WriteString(token.Content)
				}
			case "completed":
				// Check if content has the actual response
				if token.Content != "" && token.Content != "done" && token.Content != "" {
					lastContent = token.Content
				}
			case "progress", "thinking", "planning":
				// Status updates - we could emit these as thinking steps
				// For now, just track them
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
		return responseMsg(fullResponse.String())
	}

	// Fall back to lastContent if no text tokens were received
	if lastContent != "" {
		return responseMsg(lastContent)
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
