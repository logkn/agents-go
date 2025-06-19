package types

import (
	"log/slog"

	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/utils"
)

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent[Context any] struct {
	// Name of the agent
	Name string
	// Instructions/system prompt
	Instructions AgentInstructions[Context]
	// Tools available
	Tools []tools.Tool[Context]
	// Model configuration
	Model ModelConfig
	// Handoffs to other agents
	Handoffs []Handoff[Context]
	// Logger
	Logger *slog.Logger
	// Hooks define optional lifecycle callbacks
	Hooks *LifecycleHooks[Context]
}

// AllTools returns all tools (regular + handoff).
func (a *Agent[Context]) AllTools() []tools.Tool[Context] {
	handoffTools := make([]tools.Tool[Context], len(a.Handoffs))
	for i, handoff := range a.Handoffs {
		handoffTools[i] = tools.Tool[Context]{
			Name:        handoff.fullname(),
			Description: handoff.description(),
			Args:        handoffToolArgs[Context]{},
		}
	}
	return append(a.Tools, handoffTools...)
}

func NewAgent[Context any](name string, model ModelConfig) Agent[Context] {
	return Agent[Context]{
		Name:         name,
		Model:        model,
		Tools:        []tools.Tool[Context]{},
		Logger:       utils.NilLogger(),
		Hooks:        nil,
		Instructions: StringInstructions[Context]("You are a helpful assistant."),
		Handoffs:     []Handoff[Context]{},
	}
}

// WithTools returns a new agent with the given tools.
func (a *Agent[Context]) WithTools(tools []tools.Tool[Context]) *Agent[Context] {
	a.Tools = append(a.Tools, tools...)
	return a
}

// WithHandoffs returns a new agent with the given handoffs.
func (a *Agent[Context]) WithHandoffs(handoffs []Handoff[Context]) *Agent[Context] {
	a.Handoffs = append(a.Handoffs, handoffs...)
	return a
}

func (a *Agent[Context]) WithInstructions(instructions AgentInstructions[Context]) *Agent[Context] {
	a.Instructions = instructions
	return a
}
