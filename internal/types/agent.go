package types

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/logkn/agents-go/internal/context"
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

type handoffToolArgs struct{}

func (h handoffToolArgs) Run() any {
	return "handoff_executed"
}

type AgentInstructions struct {
	OfString string
	OfFunc   func(ctx context.AnyContext) (string, error)
}

func (ins AgentInstructions) ToString(ctx context.AnyContext) (string, error) {
	if ins.OfString != "" {
		return ins.OfString, nil
	}
	if ins.OfFunc != nil {
		return ins.OfFunc(ctx)
	}
	return "", errors.New("no instruction provided")
}

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent struct {
	Name         string
	Instructions AgentInstructions
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []Handoff
	Logger       *slog.Logger
	// Hooks define optional lifecycle callbacks
	Hooks *LifecycleHooks
}

// LifecycleHooks defines optional hooks that can be called during agent execution.
type LifecycleHooks struct {
	BeforeRun      func(ctx context.AnyContext) error
	AfterRun       func(ctx context.AnyContext, result any) error
	BeforeToolCall func(ctx context.AnyContext, toolName string, args string) error
	AfterToolCall  func(ctx context.AnyContext, toolName string, result any) error
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

// AllTools returns all tools (regular + handoff).
func (a *Agent) AllTools() []tools.Tool {
	handoffTools := a.HandoffTools()
	return append(a.Tools, handoffTools...)
}
