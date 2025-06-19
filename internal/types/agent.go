package types

import (
	"log/slog"
	"strings"

	"github.com/logkn/agents-go/internal/tools"
	"github.com/stoewer/go-strcase"
)

type Handoff[Context any] struct {
	Agent           *Agent[Context]
	ToolName        string
	ToolDescription string
}

func (h Handoff[Context]) defaultName() string {
	// "transfer_to_{agent_name}"
	snakecaseName := strings.ReplaceAll(h.Agent.Name, " ", "_")
	snakecaseName = strcase.SnakeCase(snakecaseName)
	return "transfer_to_" + snakecaseName
}

func (h Handoff[Context]) fullname() string {
	if h.ToolName != "" {
		return h.ToolName
	}
	return h.defaultName()
}

func (h Handoff[Context]) defaultDescription() string {
	return "Handoff to the " + h.Agent.Name + " agent to handle the request."
}

func (h Handoff[Context]) description() string {
	if h.ToolDescription != "" {
		return h.ToolDescription
	}
	return h.defaultDescription()
}

type handoffToolArgs[Context any] struct{}

func (h handoffToolArgs[Context]) Run(ctx *Context) any {
	return "handoff_executed"
}

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

// LifecycleHooks defines optional hooks that can be called during agent execution.
type LifecycleHooks[Context any] struct {
	BeforeRun      func(ctx *Context) error
	AfterRun       func(ctx *Context, result any) error
	BeforeToolCall func(ctx *Context, toolName string, args string) error
	AfterToolCall  func(ctx *Context, toolName string, result any) error
}

func (a *Agent[Context]) HandoffTools() []tools.Tool[Context] {
	handoffTools := make([]tools.Tool[Context], len(a.Handoffs))
	for i, handoff := range a.Handoffs {
		handoffTools[i] = tools.Tool[Context]{
			Name:        handoff.fullname(),
			Description: handoff.description(),
			Args:        handoffToolArgs[Context]{},
		}
	}
	return handoffTools
}

// AllTools returns all tools (regular + handoff).
func (a *Agent[Context]) AllTools() []tools.Tool[Context] {
	handoffTools := a.HandoffTools()
	return append(a.Tools, handoffTools...)
}
