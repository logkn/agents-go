// Package tools provides a set of tools for use with agents.
package tools

import (
	"fmt"
	"log/slog"

	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/stoewer/go-strcase"
)

// ToolArgs is implemented by a type that can execute a tool using its own
// parameters.
type ToolArgs[Context any] interface {
	Run(ctx *Context) any
}

// Tool describes an executable function that can be invoked by an agent.
type Tool[Context any] struct {
	Name        string
	Description string
	Args        ToolArgs[Context]
}

// CompleteName returns the explicit name if set or derives one from the
// argument type.
func (t Tool[Context]) CompleteName() string {
	if t.Name != "" {
		return t.Name
	}
	// get the name of the type of Args

	typeName := utils.TypeName(t.Args)
	// snake case
	return strcase.SnakeCase(typeName)
}

// ToOpenAITool converts this tool into the format expected by the OpenAI SDK.
func (t Tool[Context]) ToOpenAITool() openai.ChatCompletionToolParam {
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

// RunOnArgs unmarshals the provided JSON arguments and executes the tool with context.
// This method should be used when the tool requires access to the execution context.
func (t Tool[Context]) RunOnArgs(args string, ctx *Context) any {
	// parse the args into the tool's args type

	slog.Debug("unmarshaling tool arguments", "tool_name", t.CompleteName(), "args", args)
	argsInstance := utils.NewInstance(t.Args).(ToolArgs[Context])

	// execute the tool
	result := argsInstance.Run(ctx)
	slog.Debug("tool execution completed", "tool_name", t.CompleteName())

	return result
}

// NewTool creates a new tool with the given name, description, and args.
func NewTool[T any](name, description string, args ToolArgs[T]) Tool[T] {
	return Tool[T]{
		Name:        name,
		Description: description,
		Args:        args,
	}
}

// BaseTool is a tool that does not depend on context.
// This means it is reusable across different agents.
type BaseTool struct {
	Name        string
	Description string
	Args        baseToolArgs
}

// baseToolArgsAdapter adapts baseToolArgs to work with ToolArgs[Context]
type baseToolArgsAdapter[Context any] struct {
	baseToolArgs
}

func (p baseToolArgsAdapter[Context]) Run(ctx *Context) any {
	return p.baseToolArgs.Run()
}

func AsTool[Context any](base BaseTool) Tool[Context] {
	return Tool[Context]{
		Name:        base.Name,
		Description: base.Description,
		Args:        baseToolArgsAdapter[Context]{base.Args},
	}
}

type baseToolArgs interface {
	Run() any
}
