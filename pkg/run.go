package agents

import (
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

// Run executes the specified Agent using the provided user input and returns a
// streaming response handle. It is a convenience wrapper around
// runner.Run.
func Run(agent Agent, input string) (runner.AgentResponse, error) {
	return runner.Run(types.Agent(agent), runner.Input{OfString: input})
}
