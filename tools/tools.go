package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/stoewer/go-strcase"
)

// ToolArgs is implemented by a type that can execute a tool using its own
// parameters.
type ToolArgs interface {
	Run() any
}

// Tool describes an executable function that can be invoked by an agent.
type Tool struct {
	Name        string
	Description string
	Args        ToolArgs
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
