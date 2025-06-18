package agents

import (
	agentcontext "github.com/logkn/agents-go/internal/context"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

// RunSimple is a convenience function for running an agent with a simple string input.
func RunSimple(agent Agent, input string) (runner.AgentResponse, error) {
	return runner.Run(types.Agent(agent), runner.Input{OfString: input}, nil)
}

// RunWithContext runs an agent with a global context that tools can access.
func RunWithContext(agent Agent, input string, globalContext agentcontext.AnyContext) (runner.AgentResponse, error) {
	return runner.Run(types.Agent(agent), runner.Input{OfString: input}, globalContext)
}
