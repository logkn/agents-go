package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"

	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
)

const (
	gap   = "\n\n"
	gray  = "#737373"
	ant   = "#b06227"
	green = "#2a7d2f"
)

type StreamHandler struct {
	response *runner.AgentResponse
}

func (sh *StreamHandler) Stop() {
	if sh.response != nil {
		sh.response.Stop()
	} else {
		fmt.Println("no response to stop")
	}
}

var grayColor = lipgloss.Color(gray)

// State

type UIMessage struct {
	types.Message
}

func (m UIMessage) RenderMessage(hideThoughts bool) string {
	msg := m.Message
	content := RenderMarkdown(msg.Content, hideThoughts)
	switch msg.Role {
	case types.User:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Render("> " + content)
	case types.Assistant:
		return "● " + content
	default:
		return ""
	}
}

type MessageArea struct {
	vp viewport.Model
}

func (ma MessageArea) Update(msg tea.Msg) (MessageArea, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case runner.AgentEvent:
		return ma, nil

	default:
		vp, vpCmd := ma.vp.Update(msg)
		ma.vp = vp
		cmd = vpCmd
		return ma, cmd
	}
}

type AppStateComponents struct {
	viewport viewport.Model
	inputBox textarea.Model
}

type CallAndResponse struct {
	call     types.ToolCall
	response string
}

func (cr CallAndResponse) View() string {
	// first render the tool call
	funcNamePart := lipgloss.NewStyle().Bold(true).Render(cr.call.Name)
	argsPart := cr.call.Args
	if argsPart == "{}" {
		argsPart = ""
	}

	bullet := "●"
	if cr.response != "" {
		bullet = lipgloss.NewStyle().Foreground(lipgloss.Color(green)).Render(bullet)
	}

	rendered := fmt.Sprintf("%s %s(%s)", bullet, funcNamePart, argsPart)

	if cr.response != "" {
		rendered += "\n  ⎿  " + truncateWithEllipsis(cr.response, 40)
	}
	return rendered
}

type MessageAreaItem struct {
	OfMessage *UIMessage
	OfTool    *CallAndResponse
}

func (item MessageAreaItem) View(hideThoughts bool) string {
	switch {
	case item.OfMessage != nil:
		return item.OfMessage.RenderMessage(hideThoughts)
	case item.OfTool != nil:
		return item.OfTool.View()
	}
	return ""
}

type AppState struct {
	components AppStateComponents
	// messages is purely for state & logic, not rendering
	messages       []types.Message
	items          []MessageAreaItem
	responseBuffer string
	agent          *agents.Agent
	streamHandler  StreamHandler
	hideThoughts   bool
}

func (s *AppState) pushMessage(msg types.Message) {
	s.items = append(s.items, MessageAreaItem{OfMessage: &UIMessage{msg}})
	s.messages = append(s.messages, msg)
}

func textArea(vpWidth int) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Cursor.SetMode(cursor.CursorStatic)

	ta.Prompt = " > "

	ta.SetWidth(vpWidth)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(grayColor)

	// Add rounded border styling
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(grayColor)

	ta.ShowLineNumbers = false

	newlineBinding := key.NewBinding()
	newlineBinding.SetKeys("ctrl+j")
	ta.KeyMap.InsertNewline = newlineBinding
	ta.KeyMap.InsertNewline.SetEnabled(true)
	return ta
}

func viewPort() viewport.Model {
	vp := viewport.New(80, 5)
	return vp
}

func initialComponents() AppStateComponents {
	vp := viewPort()
	return AppStateComponents{
		viewport: viewPort(),
		inputBox: textArea(vp.Width),
	}
}

func initialModel(agent *agents.Agent, hideThoughts bool) AppState {
	agent.Logger = utils.NilLogger()
	return AppState{
		components:   initialComponents(),
		messages:     []types.Message{},
		agent:        agent,
		hideThoughts: hideThoughts,
	}
}

type (
	StreamStart struct{}
	StreamEnd   struct{}
)

// Tea.Model implementation

func (s AppState) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (s *AppState) ProcessCommand(userMessage string) bool {
	userMessage = strings.TrimSpace(userMessage)
	switch userMessage {
	case "/clear":
		s.streamHandler.Stop()
		s.responseBuffer = ""
		s.items = []MessageAreaItem{}
		s.messages = []types.Message{}
		s.refreshViewport()
	default:
		return false
	}
	return true
}

func (s AppState) OnEvent(event runner.AgentEvent) (tea.Model, tea.Cmd) {
	if token, hasToken := event.Token(); hasToken {
		s.responseBuffer += token
	}

	if message, hasMessage := event.Message(); hasMessage {
		s.pushMessage(*message)
		s.responseBuffer = ""

		// handle tool calls
		for _, toolcall := range message.ToolCalls {
			s.items = append(s.items, MessageAreaItem{OfTool: &CallAndResponse{toolcall, ""}})
		}

		// if tool message, update the associated tool call item
		if message.Role == types.Tool {
			s.registerToolResponse(message.ID, message.Content)
		}
	}

	s.GoToBottom()

	return s, nil
}

func (s *AppState) registerToolResponse(id string, response string) {
	// find the item where item.OfTool.call.ID == id
	// and update the response

	for i, item := range s.items {
		if item.OfTool != nil && item.OfTool.call.ID == id {
			s.items[i].OfTool.response = response
			return

		}
	}
}

func (s *AppState) refreshViewport() {
	vp := &s.components.viewport

	lines := utils.MapSlice(s.items, func(item MessageAreaItem) string {
		return item.View(s.hideThoughts)
	})

	// Add current response buffer as temporary content without modifying s.items
	if len(s.responseBuffer) > 0 {
		respMessage := types.NewAssistantMessage(s.responseBuffer, s.agent.Name, []types.ToolCall{})
		uiMessage := UIMessage{respMessage}
		lines = append(lines, uiMessage.RenderMessage(s.hideThoughts))
	}

	content := strings.Join(lines, gap)
	content = lipgloss.NewStyle().Width(vp.Width).Render(content)

	vp.SetContent(lipgloss.NewStyle().Width(vp.Width).Render(content))
}

func (s *AppState) GoToBottom() {
	s.refreshViewport()
	s.components.viewport.GotoBottom()
}

func (s AppState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	s.components.inputBox, tiCmd = s.components.inputBox.Update(msg)
	s.components.viewport, vpCmd = s.components.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return s, tea.Quit
		case tea.KeyEsc:
			s.streamHandler.Stop()
			// add the current response buffer to conversation
			if len(s.responseBuffer) > 0 {
				respMessage := types.NewAssistantMessage(s.responseBuffer, s.agent.Name, []types.ToolCall{})
				s.pushMessage(respMessage)
				s.responseBuffer = ""
			}
		case tea.KeyEnter:
			msg := s.components.inputBox.Value()
			userMessage := types.NewUserMessage(msg)
			if s.ProcessCommand(msg) {
				s.components.inputBox.Reset()
				return s, nil
			}
			s.pushMessage(userMessage)
			s.components.inputBox.Reset()

			// Initialize stream control
			agentResponse := StreamAgent(s.agent, s.messages)
			s.streamHandler.response = agentResponse

			go func() {
				defer s.streamHandler.Stop()
				defer p.Send(StreamEnd{})
				p.Send(StreamStart{})

				for event := range s.streamHandler.response.Stream() {
					p.Send(event)
				}
			}()
		}

	case tea.WindowSizeMsg:
		s.components.viewport.Width = msg.Width
		s.components.viewport.Height = msg.Height - s.components.inputBox.Height() - lipgloss.Height(gap)
		s.components.inputBox.SetWidth(s.components.viewport.Width)

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			s.components.viewport.ScrollUp(3)
		case tea.MouseButtonWheelDown:
			s.components.viewport.ScrollDown(3)
		}

	case StreamStart:
		// fmt.Println("start")

	case StreamEnd:
		// fmt.Println("end")

	case runner.AgentEvent:
		return s.OnEvent(msg)

	}

	return s, tea.Batch(tiCmd, vpCmd)
}

func (s *AppState) renderViewport() string {
	s.refreshViewport()
	return s.components.viewport.View()
}

func (s AppState) renderInput() string {
	return s.components.inputBox.View()
}

func (s AppState) View() string {
	vp := s.renderViewport()
	input := s.renderInput()
	lines := []any{
		vp,
		gap,
		input,
	}

	return fmt.Sprintf(
		strings.Repeat("%s", len(lines)),
		lines...,
	)
}

// executable

var p *tea.Program

func RunTUI(agent agents.Agent, hideThoughts bool) {
	p = tea.NewProgram(initialModel(&agent, hideThoughts), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func StreamAgent(agent *agents.Agent, messages []types.Message) *runner.AgentResponse {
	agentResponse, err := runner.Run(*agent, runner.Input{OfMessages: messages})
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return &agentResponse
}
