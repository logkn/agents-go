package llm

import (
	"context"

	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type OpenAI struct {
	baseUrl string
	model   string
}

func (o OpenAI) llm() LLM {
	return openaiLLM{
		config: o,
	}
}

type openaiLLM struct {
	config OpenAI
}

func (llm openaiLLM) Run(instructions string, messages []types.Message, tools []tools.Tool, responseFormat types.ResponseFormat) chan LLMResponse {
	// setup the client
	client := openai.NewClient()
	conf := llm.config
	if conf.baseUrl != "" {
		client.Options = append(client.Options, option.WithBaseURL(conf.baseUrl))
	}
	// convert our messages to OpenAI

	openaiMessages := utils.MapSlicePointerFn(messages, messageToOpenAI)
	openaiTools := utils.MapSlicePointerFn(tools, toolToOpenAI)

	// create the request

	params := openai.ChatCompletionNewParams{
		Messages:       openaiMessages,
		Model:          conf.model,
		ResponseFormat: responseFormatToOpenAI(responseFormat),
		Tools:          openaiTools,
	}

	stream := client.Chat.Completions.NewStreaming(context.TODO(), params)
}

// ================== Type Conversion ==================

func responseFormatToOpenAI(responseFormat types.ResponseFormat) openai.ChatCompletionNewParamsResponseFormatUnion {
	if responseFormat.String {
		oftext := shared.NewResponseFormatTextParam()
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &oftext,
		}
	}

	structured := responseFormat.Structured
	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        structured.Name,
				Description: param.NewOpt(structured.Description),
				Schema:      structured.Schema(),
				Strict:      param.NewOpt(true),
			},
		},
	}
}

func messageFromOpenAI(msg openai.ChatCompletionMessage) types.Message {
	role := types.Role(string(msg.Role))
	content := msg.Content

	toolCalls := utils.MapSlice(msg.ToolCalls, toolCallFromOpenAI)

	return types.Message{
		Role:    role,
		Content: content,
		// Name:      msg.Name, // ChatCompletionMessage doesn't have Name field
		Toolcalls: toolCalls,
	}
}

func messageToOpenAI(msg *types.Message) openai.ChatCompletionMessageParamUnion {
	switch msg.Role {
	case types.User:
		return openai.UserMessage(msg.Content)
	case types.Assistant:
		// handle tool calls
		toolCalls := utils.MapSlicePointerFn(msg.Toolcalls, toolCallToOpenAI)
		msg := openai.AssistantMessage(msg.Content)
		msg.OfAssistant.ToolCalls = toolCalls
		return msg
	case types.System:
		return openai.SystemMessage(msg.Content)
	case types.Tool:
		return openai.ToolMessage(msg.Name, msg.Content)
	default:
		return openai.UserMessage(msg.Content)
	}
}

func toolCallToOpenAI(msg *types.ToolCall) openai.ChatCompletionMessageToolCallParam {
	return openai.ChatCompletionMessageToolCallParam{
		ID: msg.ID,
		Function: openai.ChatCompletionMessageToolCallFunctionParam{
			Name:      msg.Name,
			Arguments: msg.Arguments,
		},
	}
}

func toolCallFromOpenAI(toolcall openai.ChatCompletionMessageToolCall) types.ToolCall {
	return types.ToolCall{
		ID:        toolcall.ID,
		Name:      toolcall.Function.Name,
		Arguments: toolcall.Function.Arguments,
	}
}

func toolToOpenAI(tool *tools.Tool) openai.ChatCompletionToolParam {
	schema := tool.Schema()
	name := tool.Name()
	description := tool.Description()
	return openai.ChatCompletionToolParam{
		Type: "function",
		Function: openai.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(description),
			Parameters:  openai.FunctionParameters(schema),
		},
	}
}
