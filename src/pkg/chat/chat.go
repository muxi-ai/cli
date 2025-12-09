package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Config holds chat session configuration
type Config struct {
	FormationID string
	ServerID    string
	UserID      string
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
	config        Config
	messages      []Message
	thinking      []ThinkingStep
	isThinking    bool
	thinkingStart time.Time
	textarea      textarea.Model
	viewport      viewport.Model
	renderer      *glamour.TermRenderer
	width         int
	height        int
	ready         bool
	quitting      bool
	err           error
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
			Foreground(lipgloss.Color("#555555")) // Dimmer for thinking updates

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")) // Dimmer for completed steps

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080")) // Explicit gray for all terminals

	goldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e48d20")) // Orange (matches muxi brand)

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
		config:   cfg,
		textarea: ta,
		renderer: renderer,
		messages: []Message{},
		thinking: []ThinkingStep{},
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEsc:
			if m.isThinking {
				m.isThinking = false
				m.thinking = append(m.thinking, ThinkingStep{
					Text:      "Cancelled by user",
					Completed: true,
				})
			}
			return m, nil

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

				// Clear previous thinking steps
				m.thinking = []ThinkingStep{}

				// Add user message
				m.messages = append(m.messages, Message{
					Role:      "user",
					Content:   input,
					Timestamp: time.Now(),
				})
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				m.textarea.Reset()
				m.textarea.SetHeight(1)

				// Start dummy thinking (TODO: replace with actual API call)
				m.isThinking = true
				m.thinkingStart = time.Now()
				return m, m.simulateThinking()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		inputAreaHeight := 6 // Input + dividers + status + footer padding

		m.textarea.SetWidth(msg.Width - 4)

		// Viewport gets all space except input area
		chatHeight := msg.Height - inputAreaHeight
		if chatHeight < 5 {
			chatHeight = 5
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width, chatHeight)
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = chatHeight
		}

	case thinkingMsg:
		m.thinking = append(m.thinking, ThinkingStep{
			Text:      string(msg),
			Completed: false,
		})
		return m, m.simulateNextThinking(len(m.thinking))

	case thinkingCompleteMsg:
		if int(msg) < len(m.thinking) {
			m.thinking[msg].Completed = true
		}
		return m, nil

	case responseMsg:
		m.isThinking = false
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   string(msg),
			Timestamp: time.Now(),
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case tickMsg:
		if m.isThinking {
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return tickMsg{}
			})
		}
	}

	// Update textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if !m.ready {
		return "Initializing..."
	}

	margin := " " // Left margin for entire chat
	var b strings.Builder

	// Everything in viewport - scrolls naturally like terminal
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Input divider
	dividerWidth := m.width - 4 // Account for margins
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	b.WriteString(margin)
	b.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n")

	// Input area - add margin to continuation lines
	inputLines := strings.Split(m.textarea.View(), "\n")
	for i, line := range inputLines {
		b.WriteString(margin)
		if i == 0 {
			b.WriteString(goldStyle.Render(">"))
			b.WriteString("  ")
		} else {
			b.WriteString("   ") // Align with first line (same width as ">  ")
		}
		b.WriteString(line)
		if i < len(inputLines)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Bottom divider
	b.WriteString(margin)
	b.WriteString(dividerStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n")

	// Status bar
	b.WriteString(margin)
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n\n") // Padding from terminal edge

	return b.String()
}

func (m Model) renderHeader() string {
	gold := goldStyle.Render
	dim := dimmedStyle.Render
	margin := " "

	// Calculate width - account for margin
	width := m.width - 2
	if width < 65 {
		width = 65
	}

	// Layout: │  <M logo 12 chars>  │ <right content> │
	// Left section: 1 (│) + 2 (spaces) + 12 (logo) + 2 (spaces) + 1 (│) = 18 visual chars
	// Right section: remaining width - 18 - 1 (final │)
	logoWidth := 12
	leftSection := 18 // │ + 2 spaces + logo + 2 spaces + │
	rightWidth := width - leftSection - 1

	// Helper to pad content to exact visual width
	pad := func(s string, w int) string {
		// Count visual width (runes, not bytes)
		visualLen := 0
		for range s {
			visualLen++
		}
		if visualLen >= w {
			return s
		}
		return s + strings.Repeat(" ", w-visualLen)
	}

	// M logo lines (each is exactly 12 visual chars)
	mLine1 := gold("███") + dim("╗") + "   " + gold("███") + dim("╗")  // 3+1+3+3+1 = 11... need 12
	mLine2 := gold("████") + dim("╗") + " " + gold("████") + dim("║") // 4+1+1+4+1 = 11
	mLine3 := gold("██") + dim("║╚") + gold("██") + dim("╔╝") + gold("██") + dim("║") // 2+2+2+2+2+1 = 11
	mLine4 := gold("██") + dim("║") + " " + dim("╚═╝") + " " + gold("██") + dim("║") // 2+1+1+3+1+2+1 = 11
	mLine5 := dim("╚═╝") + "     " + dim("╚═╝") // 3+5+3 = 11
	
	// Adjust: the original logo is 12 chars wide, let me recalculate
	// ███╗   ███╗ = 3+1+3+3+1 = 11 chars
	// So logo is 11 chars, left section = 1+2+11+2+1 = 17
	leftSection = 17
	rightWidth = width - leftSection - 1
	_ = logoWidth // unused now

	// Build lines
	var b strings.Builder

	// Top border: ╭── MUXI Chat ─────...─╮
	titlePart := "── " + gold("MUXI Chat") + " "
	titleVisualLen := 13 // "── MUXI Chat "
	dashesNeeded := width - titleVisualLen - 2 // -2 for ╭ and ╮
	b.WriteString(margin)
	b.WriteString(dim("╭" + "── ") + gold("MUXI Chat") + dim(" " + strings.Repeat("─", dashesNeeded) + "╮"))
	b.WriteString("\n")
	_ = titlePart

	// Content lines
	rightContents := []string{
		"",
		" Chatting with:", // dimmed later
		fmt.Sprintf("  ⌬ Formation: %s", m.config.FormationID),
		fmt.Sprintf("  ⏍ Server: %s", m.config.ServerID),
		fmt.Sprintf("  ♛ User: %s", m.config.UserID),
		"",
	}
	mLines := []string{
		"           ", // empty line (11 spaces)
		mLine1,
		mLine2,
		mLine3,
		mLine4,
		mLine5,
	}

	for i, mLine := range mLines {
		b.WriteString(margin)
		b.WriteString(dim("│") + "  ")
		if i == 0 {
			b.WriteString("           ") // 11 spaces for empty line
		} else {
			b.WriteString(mLine)
		}
		b.WriteString("  " + dim("│"))
		// Pad first, then apply dimming to "Chatting with:" line
		padded := pad(rightContents[i], rightWidth)
		if i == 1 {
			padded = dim(padded)
		}
		b.WriteString(padded)
		b.WriteString(dim("│"))
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(margin)
	b.WriteString(dim("╰" + strings.Repeat("─", width-2) + "╯"))
	b.WriteString("\n")

	// Hint text (centered)
	hint := "ENTER to send • \\ + ENTER for a new line • Ctrl+C to exit"
	hintPadding := (width - len(hint)) / 2
	if hintPadding < 0 {
		hintPadding = 0
	}
	b.WriteString("\n")
	b.WriteString(margin)
	b.WriteString(dim(strings.Repeat(" ", hintPadding) + hint))

	return b.String()
}

func (m Model) renderMessages() string {
	var b strings.Builder
	margin := " "

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
		} else {
			// Render markdown for assistant messages
			rendered, err := m.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			// Add M prefix on first line, then indent rest to align
			b.WriteString(margin)
			b.WriteString(goldStyle.Render("𝐌"))
			
			lines := strings.Split(strings.TrimSpace(rendered), "\n")
			for j, line := range lines {
				if j > 0 {
					b.WriteString(margin + "  ")
				} else {
					b.WriteString(" ") // Single space after M
				}
				b.WriteString(strings.TrimLeft(line, " ")) // Remove leading whitespace
				b.WriteString("\n")
			}
		}
	}

	// Thinking status (part of scrollable content)
	if m.isThinking || len(m.thinking) > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderThinking())
	}

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
	bold := lipgloss.NewStyle().Bold(true).Render

	left := fmt.Sprintf("%s@%s://%s",
		m.config.UserID,
		m.config.ServerID,
		m.config.FormationID,
	)

	// Visual length of right side (without ANSI codes)
	rightVisual := "? for help • / for commands"
	right := bold("?") + dim(" for help • ") + bold("/") + dim(" for commands")

	gap := m.width - len(left) - len(rightVisual) - 2
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
		m.viewport.SetContent(m.renderMessages())
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
		m.viewport.SetContent(m.renderMessages())
		m.textarea.Reset()
		m.textarea.SetHeight(1)

	default:
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   fmt.Sprintf("Unknown command: `/%s`. Type `/help` for available commands.", parts[0]),
			Timestamp: time.Now(),
		})
		m.viewport.SetContent(m.renderMessages())
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

// Simulate thinking steps (TODO: replace with actual API streaming)
func (m Model) simulateThinking() tea.Cmd {
	return tea.Batch(
		tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg{}
		}),
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return thinkingMsg("Analyzing request...")
		}),
	)
}

func (m Model) simulateNextThinking(step int) tea.Cmd {
	switch step {
	case 1:
		return tea.Sequence(
			tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
				return thinkingCompleteMsg(0)
			}),
			tea.Tick(900*time.Millisecond, func(t time.Time) tea.Msg {
				return thinkingMsg("Routing to agent...")
			}),
		)
	case 2:
		return tea.Sequence(
			tea.Tick(600*time.Millisecond, func(t time.Time) tea.Msg {
				return thinkingCompleteMsg(1)
			}),
			tea.Tick(700*time.Millisecond, func(t time.Time) tea.Msg {
				return thinkingMsg("Generating response...")
			}),
		)
	case 3:
		return tea.Sequence(
			tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
				return thinkingCompleteMsg(2)
			}),
			tea.Tick(1100*time.Millisecond, func(t time.Time) tea.Msg {
				return responseMsg(`This is a **dummy response** from the formation.

## Features Supported

| Feature | Status | Notes |
|---------|--------|-------|
| Markdown | ✓ | Full support |
| Code blocks | ✓ | With syntax highlighting |
| Tables | ✓ | GitHub-flavored |
| Lists | ✓ | Ordered and unordered |

### Code Example

` + "```" + `go
package main

import "fmt"

func main() {
    message := "Hello, MUXI!"
    fmt.Println(message)
    
    for i := 0; i < 3; i++ {
        fmt.Printf("Count: %d\n", i)
    }
}
` + "```" + `

### Lists

**Unordered:**
- First item
- Second item
  - Nested item
- Third item

**Ordered:**
1. Step one
2. Step two
3. Step three

> **Note:** This is a blockquote for important information.

*Replace this with actual API integration.*`)
			}),
		)
	}
	return nil
}

// Run starts the chat UI
func Run(cfg Config) error {
	p := tea.NewProgram(
		New(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
