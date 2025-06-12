package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textInput     textinput.Model
	messages      []string
	terminalWidth int
	initialized   bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something and press Enter..."
	ti.Focus()
	ti.Cursor.SetMode(0) // Disable blinking

	return model{
		textInput:     ti,
		messages:      []string{},
		terminalWidth: 0, // Will be set on first WindowSizeMsg
		initialized:   false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.initialized {
			// First time - just set the width without clearing
			m.initialized = true
			m.terminalWidth = msg.Width
			m.textInput.Width = msg.Width - 4
		} else if msg.Width != m.terminalWidth {
			// Actual resize - clear and update
			m.terminalWidth = msg.Width
			m.textInput.Width = msg.Width - 4
			return m, tea.Sequence(tea.ClearScreen, tea.EnterAltScreen, tea.ExitAltScreen)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.textInput.Value() != "" {
				// Add the submitted text to messages
				m.messages = append(m.messages, m.textInput.Value())
				// Clear the input
				m.textInput.SetValue("")
			}
		}
	}

	// Update the text input
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Build the output from all submitted messages
	var output strings.Builder

	for _, message := range m.messages {
		output.WriteString("> " + message + "\n\n")
	}

	// Style the input box with centered margins
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(m.textInput.Width + 2) // Account for padding

	// Combine messages and input
	return fmt.Sprintf(
		"%s\n%s\n\nPress Ctrl+C to quit",
		output.String(),
		inputStyle.Render(m.textInput.View()),
	)
}

func RunTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
