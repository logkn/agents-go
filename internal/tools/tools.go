// Package tools provides a set of tools for use with agents.
package tools

import (
	"encoding/json"
	"fmt"
	"reflect"

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
	schema, err := utils.CreateSchema(t.Args)
	if err != nil {
		fmt.Println("Error creating schema for tool arguments:", err)
		return openai.ChatCompletionToolParam{}
	}
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

	// Special handling for baseToolArgsAdapter to access the underlying baseToolArgs
	if adapter, ok := t.Args.(baseToolArgsAdapter[Context]); ok {
		// Create a new instance of the underlying baseToolArgs type
		argsInstancePtr := utils.NewInstance(adapter.baseToolArgs)

		// unmarshal JSON args into the instance
		if err := json.Unmarshal([]byte(args), argsInstancePtr); err != nil {
			return fmt.Sprintf("Error unmarshaling arguments: %v", err)
		}

		// Dereference the pointer and cast to baseToolArgs
		argsValue := reflect.ValueOf(argsInstancePtr).Elem().Interface()
		if baseArgs, ok := argsValue.(baseToolArgs); ok {
			result := baseArgs.Run()
			return result
		}

		return fmt.Sprintf("Error: cannot cast %T to baseToolArgs", argsValue)
	}

	// Regular tool handling
	argsInstance := utils.NewInstance(t.Args)

	// unmarshal JSON args into the instance
	if err := json.Unmarshal([]byte(args), argsInstance); err != nil {
		return fmt.Sprintf("Error unmarshaling arguments: %v", err)
	}

	toolArgs := argsInstance.(ToolArgs[Context])

	// execute the tool
	result := toolArgs.Run(ctx)

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

type baseToolArgs interface {
	Run() any
}

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

func CoerceBaseTool[Context any](base BaseTool) Tool[Context] {
	return Tool[Context]{
		Name:        base.Name,
		Description: base.Description,
		Args:        baseToolArgsAdapter[Context]{baseToolArgs: base.Args},
	}
}
