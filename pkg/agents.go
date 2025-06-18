package agents

import (
	"context"
	"fmt"
	"log/slog"

	agentcontext "github.com/logkn/agents-go/internal/context"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
)

type (
	Agent                 = types.Agent
	ModelConfig           = types.ModelConfig
	LifecycleHooks        = types.LifecycleHooks
	AnyContext            = agentcontext.AnyContext
	Tool                  = tools.Tool
	ToolArgs              = tools.ToolArgs
	AnyContextualToolArgs = tools.AnyContextualToolArgs
	Input                 = runner.Input
	AgentResponse         = runner.AgentResponse
	Role                  = types.Role
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
type agentToolArgs struct {
	// Prompt is the user input passed to the nested agent.
	Prompt string

	agent types.Agent `json:"-"`
}

// Run executes the wrapped agent using the provided prompt and returns the
// final assistant response content. Errors are returned as strings.
func (a agentToolArgs) Run() any {
	resp, err := runner.Run(a.agent, runner.Input{OfString: a.Prompt}, context.Background())
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

// Context Creation Functions

// ContextFactory is a function that creates a new context instance.
type ContextFactory[T any] = agentcontext.ContextFactory[T]

// NewContext creates a new typed context with the provided data.
func NewContext[T any](data T) agentcontext.Context[T] {
	return agentcontext.NewContext(data)
}

// EmptyContext creates a context with no data for agents that don't need context.
func EmptyContext() agentcontext.Context[agentcontext.NoContext] {
	return agentcontext.EmptyContext()
}

// FromAnyContext attempts to convert an AnyContext back to a typed Context[T].
func FromAnyContext[T any](anyCtx AnyContext) (agentcontext.Context[T], error) {
	return agentcontext.FromAnyContext[T](anyCtx)
}

// ToAnyContext converts a typed Context[T] to AnyContext for internal use.
func ToAnyContext[T any](ctx agentcontext.Context[T]) AnyContext {
	return agentcontext.ToAnyContext(ctx)
}

// Agent Builder Functions

// AgentConfig holds the basic configuration for creating an agent.
type AgentConfig struct {
	Name         string
	Instructions string
	Model        ModelConfig
	Logger       *slog.Logger
}

// NewAgent creates a new agent without context.
func NewAgent(config AgentConfig) Agent {
	return Agent{
		Name:         config.Name,
		Instructions: config.Instructions,
		Model:        config.Model,
		Logger:       config.Logger,
		Tools:        []tools.Tool{},
		Handoffs:     []types.Handoff{},
	}
}

// NewAgentWithContext creates a new agent with typed context.
func NewAgentWithContext[T any](config AgentConfig, ctx agentcontext.Context[T]) Agent {
	agent := NewAgent(config)
	agent.Context = agentcontext.ToAnyContext(ctx)
	return agent
}

// WithTools adds tools to an agent.
func WithTools(agent Agent, tools ...Tool) Agent {
	agent.Tools = append(agent.Tools, tools...)
	return agent
}

// WithHooks adds lifecycle hooks to an agent.
func WithHooks(agent Agent, hooks *LifecycleHooks) Agent {
	agent.Hooks = hooks
	return agent
}

// WithHandoffs adds handoff configurations to an agent.
func WithHandoffs(agent Agent, handoffs ...types.Handoff) Agent {
	agent.Handoffs = append(agent.Handoffs, handoffs...)
	return agent
}

// Tool Creation Functions

// NewTool creates a new tool with the given configuration.
func NewTool(name, description string, args ToolArgs) Tool {
	return tools.NewTool(name, description, args)
}

// NewContextualTool creates a new tool with context support.
func NewContextualTool[T any](name, description string, args AnyContextualToolArgs, ctx agentcontext.Context[T]) Tool {
	return tools.NewContextualTool(name, description, args, ctx)
}

// Run executes an agent with the given input.
func Run(ctx context.Context, agent Agent, input Input) (AgentResponse, error) {
	return runner.Run(agent, input, ctx)
}
