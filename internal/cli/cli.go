package cli

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

var mdRenderer, _ = glamour.NewTermRenderer(
	glamour.WithStylePath("customstyle.json"),
	glamour.WithWordWrap(0),
)

var agent = agents.Agent{
	Name:         "Main Agent",
	Instructions: "You are a helpful assistant. Use the tools provided to answer questions.",
	Tools: []tools.Tool{
		tools.SearchTool,
		tools.PwdTool,
	},
	Model: types.ModelConfig{
		Model:       "qwen3:30b-a3b",
		BaseUrl:     "http://localhost:11434/v1",
		Temperature: 0.6,
	},
}

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
	errMsg        error
	tokenMsg      string
	streamReady   chan string
	agentResponse *runner.AgentResponse
)

type responseModel struct {
	responseBuffer string
	streamChan     chan string
	spinner        spinner.Model
}

type model struct {
	viewport          viewport.Model
	messages          []types.Message
	textarea          textarea.Model
	textareaHeight    int
	senderStyle       lipgloss.Style
	err               error
	thinkingSpinner   spinner.Model
	responseBuffer    string
	isThinking        bool
	streamChan        chan string
	streamSpinner     spinner.Model
	streamInterrupted bool
	agent             agents.Agent
	mdRenderer        *glamour.TermRenderer
	currentResponse   *runner.AgentResponse
	pendingMessage    *types.Message
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
	agent.Logger = silentLogger

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

type ContentSegment struct {
	Text       string
	IsThinking bool
}

func normalizeThinkTags(text string) string {
	text = strings.TrimSpace(text)
	thinkStartRe := regexp.MustCompile(`<think>\s*`)
	thinkEndRe := regexp.MustCompile(`\s*</think>`)

	if thinkStartRe.MatchString(text) && !thinkEndRe.MatchString(text) {
		text += "</think>"
	}
	return text
}

func parseContentSegments(text string) []ContentSegment {
	text = normalizeThinkTags(text)

	thinkTagRe := regexp.MustCompile(`(?s)<think>\s*(.*?)\s*</think>`)
	matches := thinkTagRe.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		return []ContentSegment{{Text: text, IsThinking: false}}
	}

	var segments []ContentSegment
	lastEnd := 0

	for _, match := range matches {
		if match[0] > lastEnd {
			content := text[lastEnd:match[0]]
			if strings.TrimSpace(content) != "" {
				segments = append(segments, ContentSegment{Text: content, IsThinking: false})
			}
		}

		thinkingContent := text[match[2]:match[3]]
		if strings.TrimSpace(thinkingContent) != "" {
			segments = append(segments, ContentSegment{Text: thinkingContent, IsThinking: true})
		}

		lastEnd = match[1]
	}

	if lastEnd < len(text) {
		content := text[lastEnd:]
		if strings.TrimSpace(content) != "" {
			segments = append(segments, ContentSegment{Text: content, IsThinking: false})
		}
	}

	return segments
}

func renderContent(content string, isThinking bool) string {
	if mdRenderer == nil {
		if isThinking {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(grayColor)).Italic(true).Render(content)
		}
		return content
	}

	rendered, err := mdRenderer.Render(content)
	if err != nil {
		rendered = content
	} else {
		rendered = strings.TrimSpace(rendered)
	}

	if isThinking {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(grayColor)).Italic(true).Render(rendered)
	}
	return rendered
}

func renderMarkdown(text string) string {
	segments := parseContentSegments(text)
	if len(segments) == 0 {
		return ""
	}

	var result strings.Builder
	for i, segment := range segments {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(renderContent(segment.Text, segment.IsThinking))
	}

	return strings.TrimSpace(result.String())
}

func (m *model) renderStream() string {
	// isStreaming := m.streamChan != nil
	spinner := m.streamSpinner.View()
	fmtBuffer := renderMarkdown(m.responseBuffer)
	return fmt.Sprintf("%s %s", spinner, fmtBuffer)
}

func (m *model) renderMessage(msg types.Message) string {
	content := renderMarkdown(msg.Content)
	switch msg.Role {
	case types.User:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(grayColor)).Render("> " + content)
	case types.Assistant:
		return "● " + content
	default:
		return content
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

		// Store the response object for later use
		m.currentResponse = &resp

		// Convert agent stream to enhanced event channel while also collecting final conversation
		tokenChan := make(chan string, 100)
		go func() {
			defer close(tokenChan)
			// Process the stream but let the response object collect events too
			stream := resp.Stream()
			for event := range stream {
				// Handle different event types for display
				if token, hasToken := event.Token(); hasToken && token != "" {
					tokenChan <- token
				} else if message, hasMessage := event.Message(); hasMessage {
					// Store the complete message for later addition to conversation
					m.pendingMessage = message
					// Signal that we have a complete message - clear buffer
					tokenChan <- "::MESSAGE::" + message.Content
				} else if toolResult, hasToolResult := event.ToolResult(); hasToolResult {
					// Show tool execution
					tokenChan <- "::TOOL::" + toolResult.Name + "::" + fmt.Sprintf("%v", toolResult.Content)
				} else if handoff, hasHandoff := event.Handoff(); hasHandoff {
					// Show agent handoff
					tokenChan <- "::HANDOFF::" + handoff.FromAgent + " -> " + handoff.ToAgent + ": " + handoff.Prompt
				} else if eventErr, hasError := event.Error(); hasError {
					// Show errors
					tokenChan <- "::ERROR::" + eventErr.Error()
				}
			}
			// Signal that we can now get the final conversation
			tokenChan <- "::FINAL::"
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

	lines := utils.MapSlice(m.messages, m.renderMessage)
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
		case tea.KeyEsc:
			// Escape key: interrupt stream if currently streaming
			if m.streamChan != nil && !m.streamInterrupted {
				// Set interrupt flag - let the goroutine close the channel
				m.streamInterrupted = true
				return m, nil
			}
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
		m.streamInterrupted = false // Reset interrupt flag for new stream
		return m, func() tea.Msg {
			return tokenMsg("start")
		}

	case tokenMsg:
		switch msg {
		case "start":
			// Start streaming - get first token
			return m, func() tea.Msg {
				if m.streamChan != nil && !m.streamInterrupted {
					token, ok := <-m.streamChan
					if ok && token != "" {
						m.isThinking = false
						return tokenMsg(token)
					}
				}
				return tokenMsg("done")
			}
		case "done":
			// Streaming finished, add pending message if available
			if m.pendingMessage != nil {
				m.messages = append(m.messages, *m.pendingMessage)
				m.pendingMessage = nil
			}
			m.responseBuffer = ""
			m.isThinking = false
			m.streamChan = nil
			m.streamInterrupted = false
			m.currentResponse = nil
			return m, nil
		default:
			// Check for special event messages
			token := string(msg)
			if token == "::FINAL::" {
				// Stream is complete, add the pending message to conversation
				if m.pendingMessage != nil {
					m.messages = append(m.messages, *m.pendingMessage)
					m.pendingMessage = nil
				}
				m.responseBuffer = ""
				m.isThinking = false
				m.streamChan = nil
				m.streamInterrupted = false
				m.currentResponse = nil
				return m, nil
			} else if strings.HasPrefix(token, "::MESSAGE::") {
				// Complete message received - clear buffer and use the message content
				messageContent := strings.TrimPrefix(token, "::MESSAGE::")
				m.responseBuffer = messageContent
				// Continue to get next token
				return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
					if m.streamChan != nil && !m.streamInterrupted {
						nextToken, ok := <-m.streamChan
						if ok && nextToken != "" {
							return tokenMsg(nextToken)
						}
					}
					return tokenMsg("done")
				})
			} else if strings.HasPrefix(token, "::TOOL::") {
				// Tool execution - add visual indicator
				parts := strings.Split(strings.TrimPrefix(token, "::TOOL::"), "::")
				if len(parts) >= 2 {
					toolName := parts[0]
					toolResult := parts[1]
					toolIndicator := fmt.Sprintf("\n**Tool:** %s\n**Result:** %s\n", toolName, toolResult)
					m.responseBuffer += toolIndicator
				}
				// Continue to get next token
				return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
					if m.streamChan != nil && !m.streamInterrupted {
						nextToken, ok := <-m.streamChan
						if ok && nextToken != "" {
							return tokenMsg(nextToken)
						}
					}
					return tokenMsg("done")
				})
			} else if strings.HasPrefix(token, "::HANDOFF::") {
				// Agent handoff - add visual indicator
				handoffMsg := strings.TrimPrefix(token, "::HANDOFF::")
				handoffIndicator := fmt.Sprintf("\n🔄 **Agent Handoff:** %s\n", handoffMsg)
				m.responseBuffer += handoffIndicator
				// Continue to get next token
				return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
					if m.streamChan != nil && !m.streamInterrupted {
						nextToken, ok := <-m.streamChan
						if ok && nextToken != "" {
							return tokenMsg(nextToken)
						}
					}
					return tokenMsg("done")
				})
			} else if strings.HasPrefix(token, "::ERROR::") {
				// Error event - add visual indicator
				errorMsg := strings.TrimPrefix(token, "::ERROR::")
				errorIndicator := fmt.Sprintf("\n❌ **Error:** %s\n", errorMsg)
				m.responseBuffer += errorIndicator
				// Continue to get next token
				return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
					if m.streamChan != nil && !m.streamInterrupted {
						nextToken, ok := <-m.streamChan
						if ok && nextToken != "" {
							return tokenMsg(nextToken)
						}
					}
					return tokenMsg("done")
				})
			} else {
				// Regular token - check if stream was interrupted
				if m.streamInterrupted {
					// Stream was interrupted, add pending message if available
					if m.pendingMessage != nil {
						m.messages = append(m.messages, *m.pendingMessage)
						m.pendingMessage = nil
					}
					m.responseBuffer = ""
					m.isThinking = false
					m.streamChan = nil
					m.streamInterrupted = false
					m.currentResponse = nil
					return m, nil
				}

				m.isThinking = false
				m.responseBuffer += token
				return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg {
					if m.streamChan != nil && !m.streamInterrupted {
						nextToken, ok := <-m.streamChan
						if ok && nextToken != "" {
							return tokenMsg(nextToken)
						}
					}
					return tokenMsg("done")
				})
			}
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
