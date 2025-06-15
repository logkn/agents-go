package types

import (
	"log/slog"
	"strings"

	"github.com/logkn/agents-go/internal/tools"
	"github.com/stoewer/go-strcase"
)

// Handoff defines an agent-to-agent transition. When a tool with the provided
// name is invoked the conversation is handed to the referenced Agent.
type Handoff struct {
	Agent           *Agent
	ToolName        string
	ToolDescription string
}

// defaultName returns the automatically generated tool name in the form
// "transfer_to_{agent_name}".
func (h Handoff) defaultName() string {
	// "transfer_to_{agent_name}"
	snakecaseName := strings.ReplaceAll(h.Agent.Name, " ", "_")
	snakecaseName = strcase.SnakeCase(snakecaseName)
	return "transfer_to_" + snakecaseName
}

// fullname returns either the explicit ToolName or the generated default name.
func (h Handoff) fullname() string {
	if h.ToolName != "" {
		return h.ToolName
	}
	return h.defaultName()
}

// defaultDescription builds a generic description for the handoff tool.
func (h Handoff) defaultDescription() string {
	return "Handoff to the " + h.Agent.Name + " agent to handle the request."
}

// description returns the tool description provided by the user or falls back
// to the automatically generated one.
func (h Handoff) description() string {
	if h.ToolDescription != "" {
		return h.ToolDescription
	}
	return h.defaultDescription()
}

type handoffToolArgs struct {
	Prompt string `json:"prompt" description:"The request or message to pass to the agent"`
}

func (h handoffToolArgs) Run() any {
	return "handoff_executed"
}

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []Handoff
	Logger       *slog.Logger
}

// HandoffTools converts the configured handoffs into executable tools so they
// can be exposed to the language model during a run.
func (a *Agent) HandoffTools() []tools.Tool {
	handoffTools := make([]tools.Tool, len(a.Handoffs))
	for i, handoff := range a.Handoffs {
		handoffTools[i] = tools.Tool{
			Name:        handoff.fullname(),
			Description: handoff.description(),
			Args:        handoffToolArgs{},
		}
	}
	return handoffTools
}
