package cli

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
)

const (
	gap       = "\n\n"
	grayColor = "#737373"
	ant       = "#b06227"
)

func RunTUI() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// type responseSpinner struct{}
var responseSpinner = spinner.Spinner{
	Frames: []string{"🞅", "🞆", "🞇", "🞈", "🞉", "●", "🞉", "🞈", "🞇", "🞆"},
	FPS:    time.Second / 8,
}

type (
	errMsg      error
	tokenMsg    string
	streamReady chan string
)

type responseModel struct {
	responseBuffer string
	streamChan     chan string
	spinner        spinner.Model
}

type model struct {
	viewport        viewport.Model
	messages        []types.Message
	textarea        textarea.Model
	textareaHeight  int
	senderStyle     lipgloss.Style
	err             error
	thinkingSpinner spinner.Model
	responseBuffer  string
	isThinking      bool
	streamChan      chan string
	streamSpinner   spinner.Model
	agent           agents.Agent
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Cursor.SetMode(cursor.CursorStatic)

	ta.Prompt = " > "

	ta.SetWidth(30)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	// Add rounded border styling
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(grayColor))

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)

	ta.KeyMap.InsertNewline.SetEnabled(true)

	// thinking spinner
	ts := spinner.New()
	ts.Spinner = spinner.MiniDot
	ts.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ant))

	// stream spinner
	ss := spinner.New()
	ss.Spinner = responseSpinner

	// Configure agent with silent logger
	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	agent := agents.Agent{
		Name:         "Assistant",
		Instructions: "You are a helpful assistant. Answer questions clearly and concisely.",
		Tools:        []tools.Tool{}, // No tools for now
		Model:        agents.ModelConfig{Model: "qwen3:30b-a3b", BaseUrl: "http://localhost:11434/v1"},
		Logger:       silentLogger,
	}

	return model{
		textarea:        ta,
		textareaHeight:  1,
		messages:        []types.Message{},
		viewport:        vp,
		senderStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color(grayColor)),
		err:             nil,
		thinkingSpinner: ts,
		streamSpinner:   ss,
		agent:           agent,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.thinkingSpinner.Tick, m.streamSpinner.Tick)
}

func (m *model) renderStream() string {
	// isStreaming := m.streamChan != nil
	spinner := m.streamSpinner.View()
	return fmt.Sprintf("%s %s", spinner, m.responseBuffer)
}

func renderMessage(msg types.Message) string {
	switch msg.Role {
	case types.User:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(grayColor)).Render("> " + msg.Content)
	case types.Assistant:
		return "● " + msg.Content
	default:
		return msg.Content
	}
}

func (m *model) realAgentResponse() tea.Cmd {
	return func() tea.Msg {
		// Prepare conversation history - always include all messages
		// The runner will handle adding system message if needed
		input := runner.Input{OfMessages: m.messages}

		// Run the agent
		resp, err := runner.Run(m.agent, input)
		if err != nil {
			// Create a simple error channel
			errorChan := make(chan string, 1)
			errorChan <- "Error: " + err.Error()
			close(errorChan)
			return streamReady(errorChan)
		}

		// Convert agent stream to token channel
		tokenChan := make(chan string, 100)
		go func() {
			defer close(tokenChan)
			for event := range resp.Stream() {
				if token, hasToken := event.Token(); hasToken && token != "" {
					tokenChan <- token
				}
			}
		}()

		return streamReady(tokenChan)
	}
}

func (m *model) renderSpinner() string {
	spinner := m.thinkingSpinner.View()
	message := "Thinking..."
	renderedMessage := lipgloss.NewStyle().Foreground(lipgloss.Color(ant)).Render(message)
	return fmt.Sprintf("%s %s", spinner, renderedMessage)
}

func (m *model) updateViewport() {
	curResponse := m.responseBuffer

	lines := utils.MapSlice(m.messages, renderMessage)
	if m.isThinking {
		lines = append(lines, m.renderSpinner())
	}

	// handle the streaming response
	if len(curResponse) > 0 {
		lines = append(lines, m.renderStream())
	}
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(lines, gap)))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd  tea.Cmd
		vpCmd  tea.Cmd
		tspCmd tea.Cmd
		sspCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.thinkingSpinner, tspCmd = m.thinkingSpinner.Update(msg)
	m.streamSpinner, sspCmd = m.streamSpinner.Update(msg)

	m.updateViewport()

	// newline binding
	newlineBinding := key.NewBinding()
	newlineBinding.SetKeys("ctrl+j")
	m.textarea.KeyMap.InsertNewline = newlineBinding

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reset terminal on resize
		fmt.Print("\033[2J\033[H")

		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.updateViewport()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		// fmt.Println(msg)
		switch msg.Type {
		case tea.KeyCtrlC:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			// Regular Enter: send user message
			userMessage := types.NewUserMessage(m.textarea.Value())
			m.messages = append(m.messages, userMessage)
			m.isThinking = true
			m.textarea.Reset()
			m.viewport.GotoBottom()

			// real stream from agent
			return m, m.realAgentResponse()
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp {
			m.viewport.LineUp(3)
		} else if msg.Type == tea.MouseWheelDown {
			m.viewport.LineDown(3)
		}

	case streamReady:
		// Stream is ready, store the channel and start reading
		m.streamChan = chan string(msg)
		return m, func() tea.Msg {
			return tokenMsg("start")
		}

	case tokenMsg:
		switch msg {
		case "start":
			// Start streaming - get first token
			return m, func() tea.Msg {
				if m.streamChan != nil {
					token, ok := <-m.streamChan
					if ok && token != "" {
						m.isThinking = false
						return tokenMsg(token)
					}
				}
				return tokenMsg("done")
			}
		case "done":
			// Streaming finished, add complete response to messages
			if len(m.responseBuffer) > 0 {
				aiMessage := types.NewAssistantMessage(m.responseBuffer, "AI", []types.ToolCall{})
				m.messages = append(m.messages, aiMessage)
				m.responseBuffer = ""
			}
			m.isThinking = false
			m.streamChan = nil
			return m, nil
		default:
			// Regular token
			m.isThinking = false
			m.responseBuffer += string(msg)
			return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
				if m.streamChan != nil {
					token, ok := <-m.streamChan
					if ok && token != "" {
						return tokenMsg(token)
					}
				}
				return tokenMsg("done")
			})
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, tspCmd, sspCmd)
}

func (m model) View() string {
	lines := []any{
		m.viewport.View(),
		gap,
		m.textarea.View(),
	}
	return fmt.Sprintf(
		strings.Repeat("%s", len(lines)),
		lines...,
	)
}
