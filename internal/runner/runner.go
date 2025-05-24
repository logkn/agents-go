package runner

import (
	"github.com/logkn/agents-go/internal/agent"
	"github.com/logkn/agents-go/internal/events"
)

type AgentInput any

func Run(agent *agent.Agent, input AgentInput, eventbus events.EventBus) RunResult {
	// TODO: Implement the run function
	return RunResult{}
}
