// Package agents
package agents

import (
	"fmt"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
)

type (
	Agent[Context any]          = types.Agent[Context]
	ModelConfig                 = types.ModelConfig
	LifecycleHooks[Context any] = types.LifecycleHooks[Context]
	Handoff[Context any]        = types.Handoff[Context]
	Tool[Context any]           = tools.Tool[Context]
	ToolArgs[Context any]       = tools.ToolArgs[Context]
	Input                       = runner.Input
	AgentResponse               = runner.AgentResponse
	Role                        = types.Role
)

// Role constants
const (
	User      = types.User
	Assistant = types.Assistant
	System    = types.System
	ToolRole  = types.Tool
)

// agentToolArgs represents the parameters required when running an Agent as a
// tool. The embedded agent field is ignored when generating a JSON schema and
// when unmarshalling parameters.
type agentToolArgs[Context any] struct {
	// Prompt is the user input passed to the nested agent.
	Prompt string

	agent types.Agent[Context] `json:"-"`
}

// Run executes the wrapped agent using the provided prompt and returns the
// final assistant response content. Errors are returned as strings.
func (a agentToolArgs[Context]) Run(ctx *Context) any {
	resp, err := runner.Run(a.agent, runner.Input{OfString: a.Prompt}, nil)
	if err != nil {
		return fmt.Sprintf("error running agent: %v", err)
	}
	return resp.Response().Content
}

// AsTool exposes the agent as an executable Tool. The returned Tool accepts a
// single parameter `prompt` which is used as the input for the agent. When the
// tool is invoked, the agent is run and the final response text is returned.
func AsTool[Context any](a Agent[Context], toolname, description string) tools.Tool[Context] {
	return tools.Tool[Context]{
		Name:        toolname,
		Description: description,
		Args:        agentToolArgs[Context]{agent: a},
	}
}

func NewAgent[Context any](model ModelConfig) Agent[Context] {
	return Agent[Context]{
		Model:        model,
		Tools:        []tools.Tool[Context]{},
		Logger:       utils.NilLogger(),
		Hooks:        nil,
		Name:         "Agent",
		Instructions: types.StringInstructions[Context]("You are a helpful assistant."),
	}
}
