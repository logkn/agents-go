package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
)

var agent = agents.Agent{
	Name:         "Main Agent",
	Instructions: "You are a helpful assistant. Use the tools provided to answer questions.",
	Tools:        []tools.Tool{tools.SearchTool},
	Model:        agents.ModelConfig{Model: "qwen3:30b-a3b", BaseUrl: "http://localhost:11434/v1"},
}

var (
	// Styles
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4AA")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B9D")).
			Bold(true)

	systemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)

	chatStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

// Model represents the application state
type Model struct {
	input        string
	conversation []types.Message
	messages     []string
	streaming    bool
	currentResp  *runner.AgentResponse
	err          error
}

// streamMsg is sent when we receive streaming tokens
type streamMsg string

// streamDoneMsg is sent when streaming is complete
type streamDoneMsg struct {
	conversation []types.Message
	response     string
	err          error
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.streaming || strings.TrimSpace(m.input) == "" {
				return m, nil
			}

			// Add user message
			userMsg := fmt.Sprintf("%s %s", userStyle.Render("You:"), m.input)
			m.messages = append(m.messages, userMsg)
			m.conversation = append(m.conversation, types.NewUserMessage(m.input))

			// Start streaming
			m.streaming = true
			m.input = ""
			m.messages = append(m.messages, assistantStyle.Render("Assistant: "))

			return m, m.sendQuery()
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case tea.KeyCtrlW:
			// Delete previous word (Ctrl+W)
			m.input = m.deleteWord(m.input, false)
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			// Ignore arrow keys
			return m, nil
		default:
			if !m.streaming && len(msg.Runes) > 0 {
				m.input += string(msg.Runes)
			}
		}

	case streamMsg:
		if len(m.messages) > 0 {
			m.messages[len(m.messages)-1] += string(msg)
		}

	case streamDoneMsg:
		m.streaming = false
		m.conversation = msg.conversation
		m.err = msg.err
		if msg.err != nil {
			m.messages = append(m.messages, systemStyle.Render(fmt.Sprintf("Error: %v", msg.err)))
		} else if msg.response != "" {
			// Replace the "Assistant: " with the full response
			if len(m.messages) > 0 {
				m.messages[len(m.messages)-1] = assistantStyle.Render("Assistant: ") + msg.response
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	s := chatStyle.Render("🤖 AI Agent Chat") + "\n\n"

	// Show conversation history
	for _, msg := range m.messages {
		s += msg + "\n"
	}

	// Show input field
	if !m.streaming {
		s += "\n" + inputStyle.Render("› "+m.input+"█")
		s += "\n\n" + systemStyle.Render("Press Enter to send, Ctrl+C to quit")
	} else {
		s += "\n\n" + systemStyle.Render("🔄 Thinking...")
	}

	return s
}

// deleteWord deletes a word from the input string
func (m Model) deleteWord(input string, forward bool) string {
	if len(input) == 0 {
		return input
	}

	if forward {
		// Delete forward word (not commonly used in chat, but here for completeness)
		return input
	} else {
		// Delete backward word
		runes := []rune(input)
		pos := len(runes) - 1

		// Skip trailing spaces
		for pos >= 0 && runes[pos] == ' ' {
			pos--
		}

		// Delete the word
		for pos >= 0 && runes[pos] != ' ' {
			pos--
		}

		return string(runes[:pos+1])
	}
}

// sendQuery sends the query to the agent and returns a command to handle streaming
func (m Model) sendQuery() tea.Cmd {
	return func() tea.Msg {
		resp, err := runner.Run(agent, runner.Input{OfMessages: m.conversation})
		if err != nil {
			return streamDoneMsg{conversation: m.conversation, err: err}
		}

		// Process streaming response synchronously
		var responseText strings.Builder
		for event := range resp.Stream() {
			if token, ok := event.Token(); ok {
				responseText.WriteString(token)
			}
		}

		return streamDoneMsg{
			conversation: resp.FinalConversation(),
			response:     responseText.String(),
			err:          nil,
		}
	}
}

// RunChat starts an interactive TUI session with the agent
func RunChat() {
	// Set up structured logging
	agent.Logger = utils.SetupLogger()

	// Initialize the model
	m := Model{
		conversation: []types.Message{types.NewSystemMessage(agent.Instructions)},
		messages:     []string{},
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running chat: %v\n", err)
	}
}

// RunSingleQuery processes a single query and exits.
func RunSingleQuery(query string) {
	// Set up structured logging
	agent.Logger = utils.SetupLogger()

	conversation := []types.Message{
		types.NewSystemMessage(agent.Instructions),
		types.NewUserMessage(query),
	}

	resp, err := runner.Run(agent, runner.Input{OfMessages: conversation})
	if err != nil {
		fmt.Println("Error running agent:", err)
		return
	}

	for event := range resp.Stream() {
		if token, ok := event.Token(); ok {
			fmt.Print(token)
		}
	}
	fmt.Println()
}
