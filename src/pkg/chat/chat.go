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
		config:    cfg,
		textarea:  ta,
		renderer:  renderer,
		messages:  []Message{},
		thinking:  []ThinkingStep{},
		asyncMode: "auto",
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
			// Clear input first, quit on second press
			if strings.TrimSpace(m.textarea.Value()) != "" {
				m.textarea.Reset()
				return m, nil
			}
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
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, nil
			}
			if m.isThinking {
				m.isThinking = false
				m.thinking = append(m.thinking, ThinkingStep{
					Text:      "Cancelled by user",
					Completed: true,
				})
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
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, m.simulateNextThinking(len(m.thinking))

	case thinkingCompleteMsg:
		if int(msg) < len(m.thinking) {
			m.thinking[msg].Completed = true
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
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

	// Check for ? as first/only character - show help immediately (not during streaming)
	if m.textarea.Value() == "?" && !m.isThinking {
		m.showHelp = true
		m.showCommands = false
		m.textarea.Reset()
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	} else if m.showHelp && m.textarea.Value() != "" {
		// Clear help when user starts typing something else
		m.showHelp = false
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
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

	// Command menu or submenu (above input) - only one at a time
	if m.showSubmenu {
		b.WriteString(m.renderSubmenu())
	} else if m.showCommands {
		b.WriteString(m.renderCommands())
	}

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

	return b.String()
}

func (m Model) renderHeader() string {
	gold := goldStyle.Render
	dim := dimmedStyle.Render

	// Fixed width box (65 chars)
	boxWidth := 65

	// M logo lines (11 chars visual width each)
	mLine1 := gold("███") + dim("╗") + "   " + gold("███") + dim("╗")
	mLine2 := gold("████") + dim("╗") + " " + gold("████") + dim("║")
	mLine3 := gold("██") + dim("║╚") + gold("██") + dim("╔╝") + gold("██") + dim("║")
	mLine4 := gold("██") + dim("║") + " " + dim("╚═╝") + " " + gold("██") + dim("║")
	mLine5 := dim("╚═╝") + "     " + dim("╚═╝")

	// Right content - pad to fill remaining space (65 - 1 - 2 - 11 - 2 - 1 - 1 = 47 chars)
	rightWidth := 47
	pad := func(s string, w int) string {
		if len(s) >= w {
			return s[:w]
		}
		return s + strings.Repeat(" ", w-len(s))
	}

	var b strings.Builder

	// Top border: ╭── MUXI Chat ─────...─╮
	titleText := "── " + gold("MUXI Chat") + " "
	dashesNeeded := boxWidth - 15 - 2 // 15 = "── MUXI Chat " visual, 2 = corners
	b.WriteString(dim("╭") + dim("── ") + gold("MUXI Chat") + dim(" " + strings.Repeat("─", dashesNeeded) + "╮"))
	b.WriteString("\n")
	_ = titleText

	// Line 1: empty
	b.WriteString(dim("│") + "  " + "           " + "  " + dim("│") + pad("", rightWidth) + dim("│") + "\n")

	// Lines 2-6: logo + content
	mLines := []string{mLine1, mLine2, mLine3, mLine4, mLine5}
	rightContents := []string{
		" Chatting with:",
		"  ⌬ Formation: " + m.config.FormationID,
		"  ⏍ Server: " + m.config.ServerID,
		"  ♛ User: " + m.config.UserID,
		"",
	}

	for i, mLine := range mLines {
		b.WriteString(dim("│") + "  ")
		b.WriteString(mLine)
		b.WriteString("  " + dim("│"))
		content := rightContents[i]
		if i == 0 {
			b.WriteString(dim(pad(content, rightWidth)))
		} else {
			b.WriteString(pad(content, rightWidth))
		}
		b.WriteString(dim("│") + "\n")
	}

	// Bottom border
	b.WriteString(dim("╰" + strings.Repeat("─", boxWidth-2) + "╯"))
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
			b.WriteString(goldStyle.Render("𝐌"))
			b.WriteString(" ")
			
			lines := strings.Split(strings.TrimSpace(rendered), "\n")
			for j, line := range lines {
				if j > 0 {
					b.WriteString(" ")
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
	)

	_, err := p.Run()
	return err
}
