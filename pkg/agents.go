package agents

import (
	"fmt"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/tools"
)

// ModelConfig contains configuration details for an LLM model.
// Model is the identifier of the model to use and BaseUrl is an optional
// override for the API base URL.
type ModelConfig struct {
	Model   string
	BaseUrl string
}

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []*Agent
}

// agentToolArgs represents the parameters required when running an Agent as a
// tool. The embedded agent field is ignored when generating a JSON schema and
// when unmarshalling parameters.
type agentToolArgs struct {
	// Prompt is the user input passed to the nested agent.
	Prompt string

	agent Agent `json:"-"`
}

// Run executes the wrapped agent using the provided prompt and returns the
// final assistant response content. Errors are returned as strings.
func (a agentToolArgs) Run() any {
	resp, err := runner.Run(a.agent, a.Prompt)
	if err != nil {
		return fmt.Sprintf("error running agent: %v", err)
	}
	return resp.Response().Content
}

// AsTool exposes the agent as an executable Tool. The returned Tool accepts a
// single parameter `prompt` which is used as the input for the agent. When the
// tool is invoked, the agent is run and the final response text is returned.
func (a Agent) AsTool(toolname, description string) tools.Tool {
	return tools.Tool{
		Name:        toolname,
		Description: description,
		Args:        agentToolArgs{agent: a},
	}
}
