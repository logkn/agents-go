package agents

import (
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

func Run(agent Agent, input string) (runner.AgentResponse, error) {
	return runner.Run(types.Agent(agent), runner.Input{OfString: input})
}
