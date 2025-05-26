package agent

import (
	"github.com/logkn/agents-go/internal/events"
	"github.com/logkn/agents-go/internal/llm"
	"github.com/logkn/agents-go/internal/tools"
)

type handoff struct {
	agent           *Agent
	toolName        string
	toolDescription string
}

func Handoff(agent *Agent) handoff {
	return handoff{agent: agent}
}

func (h handoff) WithToolName(name string) handoff {
	h.toolName = name
	return h
}

func (h handoff) WithToolDescription(name string) handoff {
	h.toolDescription = name
	return h
}

type Hooks struct {
	OnAgentStart func(agent Agent, state any, events events.EventBus)
	OnAgentEnd   func(agent Agent, response any, state any, events events.EventBus)
	OnHandoff    func(from Agent, to Agent, state any, events events.EventBus)
	OnToolCalled func(agent Agent, tool tools.Tool, state any, events events.EventBus)
	OnToolResult func(agent Agent, tool tools.Tool, result any, state any, events events.EventBus)
}

type Agent struct {
	Name         string
	Instructions string
	Model        llm.Model
	Tools        []tools.Tool
	Handoffs     []handoff
	Hooks        Hooks
}
