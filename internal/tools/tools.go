package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/logkn/agents-go/internal/context"
	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/stoewer/go-strcase"
)

// ToolArgs is implemented by a type that can execute a tool using its own
// parameters.
type ToolArgs interface {
	Run() any
}

// ContextualToolArgs is implemented by a type that can execute a tool using its own
// parameters and has access to the agent's execution context.
type ContextualToolArgs[T any] interface {
	RunWithContext(ctx context.Context[T]) any
}

// Tool describes an executable function that can be invoked by an agent.
type Tool struct {
	Name        string
	Description string
	Args        ToolArgs
	// Context holds the execution context if this tool requires it
	Context context.AnyContext
}

// CompleteName returns the explicit name if set or derives one from the
// argument type.
func (t Tool) CompleteName() string {
	if t.Name != "" {
		return t.Name
	}
	// get the name of the type of Args

	typeName := utils.TypeName(t.Args)
	// snake case
	return strcase.SnakeCase(typeName)
}

// ToOpenAITool converts this tool into the format expected by the OpenAI SDK.
func (t Tool) ToOpenAITool() openai.ChatCompletionToolParam {
	slog.Debug("converting tool to OpenAI format", "tool_name", t.CompleteName())
	schema, err := utils.CreateSchema(t.Args)
	if err != nil {
		slog.Error("failed to create schema for tool", "tool_name", t.CompleteName(), "error", err)
		fmt.Println("Error creating schema for tool arguments:", err)
		return openai.ChatCompletionToolParam{}
	}
	slog.Debug("tool schema created successfully", "tool_name", t.CompleteName())
	return openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        t.CompleteName(),
			Description: openai.String(t.Description),
			Parameters:  schema,
		},
	}
}

// RunOnArgs unmarshals the provided JSON arguments and executes the tool.
func (t Tool) RunOnArgs(args string) any {
	slog.Debug("unmarshaling tool arguments", "tool_name", t.CompleteName(), "args", args)
	argsInstance := utils.NewInstance(t.Args).(ToolArgs)
	err := json.Unmarshal([]byte(args), argsInstance)
	if err != nil {
		slog.Error("failed to unmarshal tool arguments",
			"tool_name", t.CompleteName(),
			"args", args,
			"error", err)
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to unmarshal tool arguments: %v", err),
			"tool":  t.CompleteName(),
			"args":  args,
		}
	}

	slog.Debug("executing tool", "tool_name", t.CompleteName())
	result := argsInstance.Run()
	slog.Debug("tool execution completed", "tool_name", t.CompleteName())
	return result
}

// RunOnArgsWithContext unmarshals the provided JSON arguments and executes the tool with context.
// This method should be used when the tool requires access to the execution context.
func (t Tool) RunOnArgsWithContext(args string) any {
	if t.Context == nil {
		slog.Warn("tool has no context but RunOnArgsWithContext was called", "tool_name", t.CompleteName())
		return t.RunOnArgs(args)
	}

	slog.Debug("unmarshaling contextual tool arguments", "tool_name", t.CompleteName(), "args", args)
	argsInstance := utils.NewInstance(t.Args)
	err := json.Unmarshal([]byte(args), argsInstance)
	if err != nil {
		slog.Error("failed to unmarshal contextual tool arguments",
			"tool_name", t.CompleteName(),
			"args", args,
			"error", err)
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to unmarshal tool arguments: %v", err),
			"tool":  t.CompleteName(),
			"args":  args,
		}
	}

	slog.Debug("executing contextual tool", "tool_name", t.CompleteName(), "context_type", t.Context.TypeName())
	
	// Try to execute as contextual tool first, fallback to regular tool
	if result, ok := t.tryRunWithContext(argsInstance); ok {
		slog.Debug("contextual tool execution completed", "tool_name", t.CompleteName())
		return result
	}
	
	// Fallback to regular tool execution
	slog.Debug("contextual execution failed, falling back to regular execution", "tool_name", t.CompleteName())
	if toolArgs, ok := argsInstance.(ToolArgs); ok {
		result := toolArgs.Run()
		slog.Debug("tool execution completed (fallback)", "tool_name", t.CompleteName())
		return result
	}
	
	return map[string]interface{}{
		"error": "Tool does not implement ToolArgs interface",
		"tool":  t.CompleteName(),
	}
}

// tryRunWithContext attempts to run the tool with context using reflection to handle the generic type.
func (t Tool) tryRunWithContext(argsInstance interface{}) (any, bool) {
	// Check if the args instance implements AnyContextualToolArgs
	if contextualTool, ok := argsInstance.(AnyContextualToolArgs); ok {
		result := contextualTool.RunWithAnyContext(t.Context)
		return result, true
	}
	
	return nil, false
}

// AnyContextualToolArgs is a marker interface for tools that can work with any context type.
// This provides a bridge between the generic ContextualToolArgs[T] and runtime execution.
type AnyContextualToolArgs interface {
	ToolArgs // Still implements basic ToolArgs for fallback
	RunWithAnyContext(ctx context.AnyContext) any
}

// NewTool creates a new tool with the given name, description, and args.
func NewTool(name, description string, args ToolArgs) Tool {
	return Tool{
		Name:        name,
		Description: description,
		Args:        args,
	}
}

// NewContextualTool creates a new tool with context support.
func NewContextualTool[T any](name, description string, args AnyContextualToolArgs, ctx context.Context[T]) Tool {
	return Tool{
		Name:        name,
		Description: description,
		Args:        args,
		Context:     context.ToAnyContext(ctx),
	}
}
