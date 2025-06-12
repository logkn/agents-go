package main

// import (
// 	"fmt"
// 	"strings"
//
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// 	"github.com/logkn/agents-go/internal/runner"
// 	"github.com/logkn/agents-go/internal/tools"
// 	"github.com/logkn/agents-go/internal/types"
// 	"github.com/logkn/agents-go/internal/utils"
// 	agents "github.com/logkn/agents-go/pkg"
// )
//
// var agent = agents.Agent{
// 	Name:         "Main Agent",
// 	Instructions: "You are a helpful assistant. Use the tools provided to answer questions.",
// 	Tools:        []tools.Tool{tools.SearchTool},
// 	Model:        agents.ModelConfig{Model: "qwen3:30b-a3b", BaseUrl: "http://localhost:11434/v1", Temperature: 0.6},
// }
//
// var (
// 	// Styles
// 	userStyle = lipgloss.NewStyle().
// 			Foreground(lipgloss.Color("#545454"))
//
// 	assistantStyle = lipgloss.NewStyle().
// 			Foreground(lipgloss.Color("#ffffff")).
// 			Bold(true)
//
// 	systemStyle = lipgloss.NewStyle().
// 			Foreground(lipgloss.Color("#666666")).
// 			Italic(true)
//
// 	inputStyle = lipgloss.NewStyle().
// 			Border(lipgloss.RoundedBorder()).
// 			BorderForeground(lipgloss.Color("#545454")).
// 			Padding(0, 1)
//
// 	chatStyle = lipgloss.NewStyle().
// 			Padding(1, 2)
// )
//
// type renderable struct {
// 	message   string
// 	toolcalls []types.ToolCall
// }
//
// // Model represents the application state
// type Model struct {
// 	input           string
// 	conversation    []types.Message
// 	currentResponse string
// 	renderables     []renderable
// 	streaming       bool
// 	tokenChan       <-chan streamMsg
// 	doneChan        <-chan streamDoneMsg
// 	err             error
// }
//
// func (m *Model) AppendToken(token string) {
// 	if len(m.renderables[len(m.renderables)-1].toolcalls) > 0 {
// 		m.AppendMessage("")
// 	}
// 	m.renderables[len(m.renderables)-1].message += token
// }
//
// func (m *Model) AppendMessage(messages string) {
// 	m.renderables = append(m.renderables, renderable{
// 		message: messages,
// 	})
// }
//
// func (m *Model) AppendToolCalls(toolcalls []types.ToolCall) {
// 	m.renderables = append(m.renderables, renderable{
// 		toolcalls: toolcalls,
// 	})
// }
//
// // streamMsg is sent when we receive streaming tokens
// // type streamMsg string
// type streamMsg struct {
// 	token     string
// 	toolcalls []types.ToolCall
// }
//
// // streamDoneMsg is sent when streaming is complete
// type streamDoneMsg struct {
// 	conversation []types.Message
// 	response     string
// 	err          error
// }
//
// // Init initializes the model
// func (m Model) Init() tea.Cmd {
// 	return nil
// }
//
// // Update handles messages and updates the model
// func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.Type {
// 		case tea.KeyCtrlC, tea.KeyEsc:
// 			return m, tea.Quit
// 		case tea.KeyEnter:
// 			if m.streaming || strings.TrimSpace(m.input) == "" {
// 				return m, nil
// 			}
//
// 			// Add user message
// 			userMsg := fmt.Sprintf("%s %s", userStyle.Render("> "), m.input)
// 			m.AppendMessage(userMsg)
// 			m.conversation = append(m.conversation, types.NewUserMessage(m.input))
//
// 			// Start streaming
// 			m.streaming = true
// 			m.input = ""
// 			m.AppendMessage("")
// 			m.AppendToken(assistantStyle.Render("● "))
//
// 			return m, m.sendQuery()
// 		case tea.KeyBackspace:
// 			if len(m.input) > 0 {
// 				m.input = m.input[:len(m.input)-1]
// 			}
// 		case tea.KeyCtrlW:
// 			// Delete previous word (Ctrl+W)
// 			m.input = m.deleteWord(m.input, false)
// 		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
// 			// Ignore arrow keys
// 			return m, nil
// 		default:
// 			if !m.streaming && len(msg.Runes) > 0 {
// 				m.input += string(msg.Runes)
// 			}
// 		}
//
// 	case subscriptionStartedMsg:
// 		// Store channels and start listening for tokens
// 		m.tokenChan = msg.tokenSub
// 		m.doneChan = msg.doneSub
// 		return m, listenForTokensCmd(msg.tokenSub, msg.doneSub)
//
// 	case streamMsg:
// 		if len(m.renderables) > 0 {
// 			// append the token to the last message
// 			// m.renderables[len(m.renderables)-1].message += string(msg.token)
// 			m.AppendToken(msg.token)
// 		}
//
// 		if len(msg.toolcalls) > 0 {
// 			// If there are tool calls, append them to the last message
// 			m.AppendToolCalls(msg.toolcalls)
// 		}
//
// 		// Continue listening for more tokens
// 		return m, listenForTokensCmd(m.tokenChan, m.doneChan)
//
// 	case streamDoneMsg:
// 		m.streaming = false
// 		m.conversation = msg.conversation
// 		m.err = msg.err
// 		m.tokenChan = nil
// 		m.doneChan = nil
// 		if msg.err != nil {
// 			m.AppendMessage(systemStyle.Render(fmt.Sprintf("Error: %v", msg.err)))
// 		}
// 		// Note: we don't replace the content since tokens were added individually
// 	}
//
// 	return m, nil
// }
//
// // View renders the UI
// func (m Model) View() string {
// 	s := chatStyle.Render("PascalAI") + "\n\n"
//
// 	// Show conversation history
// 	for _, msg := range m.renderables {
// 		if len(msg.message) > 0 {
// 			s += msg.message + "\n"
// 		}
// 		if len(msg.toolcalls) > 0 {
// 			s += systemStyle.Render("Tool calls:") + "\n"
// 			for _, toolCall := range msg.toolcalls {
// 				// args := toolCall.Args
// 				s += systemStyle.Render(fmt.Sprintf("  - %s: %s", toolCall.Name, toolCall.Args)) + "\n"
// 			}
// 		}
// 	}
//
// 	// Show input field
// 	if !m.streaming {
// 		s += "\n" + inputStyle.Render("› "+m.input+"█")
// 		s += "\n\n" + systemStyle.Render("Press Enter to send, Ctrl+C to quit")
// 	} else {
// 		s += "\n\n" + systemStyle.Render("🔄 Thinking...")
// 	}
//
// 	return s
// }
//
// // deleteWord deletes a word from the input string
// func (m Model) deleteWord(input string, forward bool) string {
// 	if len(input) == 0 {
// 		return input
// 	}
//
// 	if forward {
// 		// Delete forward word (not commonly used in chat, but here for completeness)
// 		return input
// 	} else {
// 		// Delete backward word
// 		runes := []rune(input)
// 		pos := len(runes) - 1
//
// 		// Skip trailing spaces
// 		for pos >= 0 && runes[pos] == ' ' {
// 			pos--
// 		}
//
// 		// Delete the word
// 		for pos >= 0 && runes[pos] != ' ' {
// 			pos--
// 		}
//
// 		return string(runes[:pos+1])
// 	}
// }
//
// // sendQuery sends the query to the agent and starts streaming
// func (m Model) sendQuery() tea.Cmd {
// 	return func() tea.Msg {
// 		resp, err := runner.Run(agent, runner.Input{OfMessages: m.conversation})
// 		if err != nil {
// 			return streamDoneMsg{conversation: m.conversation, err: err}
// 		}
//
// 		// Start streaming tokens by creating a subscription
// 		sub := make(chan streamMsg, 100)
// 		done := make(chan streamDoneMsg, 1)
//
// 		go func() {
// 			defer close(sub)
// 			defer close(done)
//
// 			streamChan := resp.Stream()
// 			for event := range streamChan {
// 				if token, hasToken := event.Token(); hasToken {
// 					sub <- streamMsg{token: token}
// 				}
// 				if msg, hasMsg := event.Message(); hasMsg {
// 					if msg.Role == types.Assistant && len(msg.ToolCalls) > 0 {
// 						sub <- streamMsg{toolcalls: msg.ToolCalls}
// 					}
// 				}
// 			}
//
// 			done <- streamDoneMsg{
// 				conversation: resp.FinalConversation(),
// 				response:     "",
// 				err:          nil,
// 			}
// 		}()
//
// 		return subscriptionStartedMsg{
// 			tokenSub: sub,
// 			doneSub:  done,
// 		}
// 	}
// }
//
// // subscriptionStartedMsg indicates streaming has started
// type subscriptionStartedMsg struct {
// 	tokenSub chan streamMsg
// 	doneSub  chan streamDoneMsg
// }
//
// // listenForTokensCmd creates a command that listens for streaming tokens
// func listenForTokensCmd(tokenSub <-chan streamMsg, doneSub <-chan streamDoneMsg) tea.Cmd {
// 	return func() tea.Msg {
// 		select {
// 		case token, ok := <-tokenSub:
// 			if !ok {
// 				// Token channel closed, wait for done
// 				return <-doneSub
// 			}
// 			return token
// 		case done := <-doneSub:
// 			return done
// 		}
// 	}
// }
//
// // RunChat starts an interactive TUI session with the agent
// func RunChat() {
// 	// Set up structured logging
// 	agent.Logger = utils.SetupLogger()
//
// 	// Initialize the model
// 	m := Model{
// 		conversation: []types.Message{types.NewSystemMessage(agent.Instructions)},
// 		renderables:  []renderable{},
// 	}
//
// 	// Start the Bubble Tea program
// 	p := tea.NewProgram(m, tea.WithAltScreen())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Error running chat: %v\n", err)
// 	}
// }
//
// // RunSingleQuery processes a single query and exits.
// func RunSingleQuery(query string) {
// 	// Set up structured logging
// 	agent.Logger = utils.SetupLogger()
//
// 	conversation := []types.Message{
// 		types.NewSystemMessage(agent.Instructions),
// 		types.NewUserMessage(query),
// 	}
//
// 	resp, err := runner.Run(agent, runner.Input{OfMessages: conversation})
// 	if err != nil {
// 		fmt.Println("Error running agent:", err)
// 		return
// 	}
//
// 	for event := range resp.Stream() {
// 		if token, ok := event.Token(); ok {
// 			fmt.Print(token)
// 		}
// 	}
// 	fmt.Println()
// }
