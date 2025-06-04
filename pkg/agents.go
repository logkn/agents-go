package agents

import (
	"fmt"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/tools"
)

type (
	Agent       = types.Agent
	ModelConfig = types.ModelConfig
)

// agentToolArgs represents the parameters required when running an Agent as a
// tool. The embedded agent field is ignored when generating a JSON schema and
// when unmarshalling parameters.
type agentToolArgs struct {
	// Prompt is the user input passed to the nested agent.
	Prompt string

	agent types.Agent `json:"-"`
}

// Run executes the wrapped agent using the provided prompt and returns the
// final assistant response content. Errors are returned as strings.
func (a agentToolArgs) Run() any {
	resp, err := runner.Run(a.agent, runner.Input{OfString: a.Prompt})
	if err != nil {
		return fmt.Sprintf("error running agent: %v", err)
	}
	return resp.Response().Content
}

// AsTool exposes the agent as an executable Tool. The returned Tool accepts a
// single parameter `prompt` which is used as the input for the agent. When the
// tool is invoked, the agent is run and the final response text is returned.
func AsTool(a Agent, toolname, description string) tools.Tool {
	return tools.Tool{
		Name:        toolname,
		Description: description,
		Args:        agentToolArgs{agent: types.Agent(a)},
	}
}
