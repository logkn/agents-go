package tools

import (
	"encoding/json"
	"fmt"

	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/stoewer/go-strcase"
)

type ToolArgs interface {
	Run() any
}

type Tool struct {
	Name        string
	Description string
	Args        ToolArgs
}

func (t Tool) CompleteName() string {
	if t.Name != "" {
		return t.Name
	}
	// get the name of the type of Args

	typeName := utils.TypeName(t.Args)
	// snake case
	return strcase.SnakeCase(typeName)
}

func (t Tool) ToOpenAITool() openai.ChatCompletionToolParam {
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

func (t Tool) RunOnArgs(args string) any {
	argsInstance := utils.NewInstance(t.Args).(ToolArgs)
	err := json.Unmarshal([]byte(args), argsInstance)
	if err != nil {
		fmt.Println("Error unmarshalling function arguments:", err)
	}
	return argsInstance.Run()
}
