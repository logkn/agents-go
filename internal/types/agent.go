package types

import (
	"log/slog"
	"strings"

	"github.com/logkn/agents-go/internal/tools"
	"github.com/stoewer/go-strcase"
)

type Handoff struct {
	Agent           *Agent
	ToolName        string
	ToolDescription string
}

func (h Handoff) defaultName() string {
	// "transfer_to_{agent_name}"
	snakecaseName := strings.ReplaceAll(h.Agent.Name, " ", "_")
	snakecaseName = strcase.SnakeCase(snakecaseName)
	return "transfer_to_" + snakecaseName
}

func (h Handoff) fullname() string {
	if h.ToolName != "" {
		return h.ToolName
	}
	return h.defaultName()
}

func (h Handoff) defaultDescription() string {
	return "Handoff to the " + h.Agent.Name + " agent to handle the request."
}

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
