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
			Foreground(lipgloss.Color("252"))

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	goldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
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

		case tea.KeyEnter:
			// Check for Shift+Enter or Alt+Enter to add newline
			// Note: Shift+Enter detection varies by terminal
			if msg.Alt || msg.String() == "shift+enter" {
				m.textarea.InsertString("\n")
				// Grow textarea height as needed
				lines := strings.Count(m.textarea.Value(), "\n") + 1
				if lines > m.textarea.Height() && lines <= 10 {
					m.textarea.SetHeight(lines)
				}
				return m, nil
			}
			if !m.isThinking && strings.TrimSpace(m.textarea.Value()) != "" {
				input := m.textarea.Value()
				
				// Handle slash commands
				if strings.HasPrefix(input, "/") {
					return m.handleCommand(input)
				}

				// Add user message
				m.messages = append(m.messages, Message{
					Role:      "user",
					Content:   input,
					Timestamp: time.Now(),
				})
				m.textarea.Reset()
				m.textarea.SetHeight(1)

				// Start dummy thinking (TODO: replace with actual API call)
				m.isThinking = true
				m.thinkingStart = time.Now()
				m.thinking = []ThinkingStep{}
				return m, m.simulateThinking()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 9 // ASCII header
		statusHeight := 3 // Status bar + dividers
		inputHeight := 5  // Input area

		m.textarea.SetWidth(msg.Width - 4)

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-statusHeight-inputHeight)
			m.viewport.SetContent(m.renderMessages())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - statusHeight - inputHeight
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

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Messages viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Thinking status
	if m.isThinking {
		b.WriteString(m.renderThinking())
	}

	// Input divider
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Input area
	b.WriteString(">  ")
	b.WriteString(m.textarea.View())
	b.WriteString("\n")

	// Bottom divider
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderHeader() string {
	gold := goldStyle.Render

	header := fmt.Sprintf(`╭── %s ────────────────────────────────────────────────╮
│               │                                             │
│  ███╗   ███╗  │ Chatting with:                              │
│  ████╗ ████║  │  ⌬ Formation: %-30s│
│  ██║╚██╔╝██║  │  ⏍ Server: %-33s│
│  ██║ ╚═╝ ██║  │  ♕ User: %-35s│
│  ╚═╝     ╚═╝  │                                             │
╰─────────────────────────────────────────────────────────────╯`,
		gold("MUXI Chat"),
		m.config.FormationID,
		m.config.ServerID,
		m.config.UserID,
	)

	return header
}

func (m Model) renderMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		if msg.Role == "user" {
			b.WriteString(userStyle.Render(">  " + msg.Content))
			b.WriteString("\n\n")
		} else {
			// Render markdown for assistant messages
			rendered, err := m.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}
			b.WriteString(goldStyle.Render("𝐌  "))
			b.WriteString(assistantStyle.Render(strings.TrimSpace(rendered)))
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (m Model) renderThinking() string {
	var b strings.Builder
	
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	elapsed := time.Since(m.thinkingStart)
	frame := frames[int(elapsed.Milliseconds()/100)%len(frames)]

	for _, step := range m.thinking {
		if step.Completed {
			b.WriteString(completedStyle.Render("⏺  " + step.Text))
		} else {
			b.WriteString(thinkingStyle.Render(frame + "  " + step.Text))
		}
		b.WriteString("\n")
	}

	// Current thinking with elapsed time
	if len(m.thinking) == 0 || m.thinking[len(m.thinking)-1].Completed {
		b.WriteString(thinkingStyle.Render(fmt.Sprintf("%s  Thinking... %.1fs  (ESC to stop)", frame, elapsed.Seconds())))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	left := fmt.Sprintf("%s@%s://%s",
		m.config.UserID,
		m.config.ServerID,
		m.config.FormationID,
	)

	right := "? help  / commands  ESC stop"

	gap := m.width - len(left) - len(right) - 2
	if gap < 0 {
		gap = 1
	}

	return statusBarStyle.Render(left + strings.Repeat(" ", gap) + right)
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
				return responseMsg("This is a **dummy response** from the formation.\n\nIt supports:\n- Markdown formatting\n- Code blocks\n- Lists\n\n```go\nfmt.Println(\"Hello, MUXI!\")\n```\n\n*Replace this with actual API integration.*")
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
