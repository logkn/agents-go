package llm

import (
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
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

func (llm openaiLLM) Run(instructions string, messages []types.Message, tools []tools.Tool, responseFormat types.Struct) chan LLMResponse {
	// setup the client
	client := openai.NewClient()
	conf := llm.config
	if conf.baseUrl != "" {
		client.Options = append(client.Options, option.WithBaseURL(conf.baseUrl))
	}
}

// ================== Type Conversion ==================

func messageFromOpenAI(msg openai.ChatCompletionMessage) types.Message

func messageToOpenAI(msg *types.Message) openai.ChatCompletionMessageParamUnion

func toolCallToOpenAI(msg *types.ToolCall) openai.ChatCompletionMessageToolCall

func toolCallFromOpenAI(toolcall openai.ChatCompletionMessageToolCall)

func toolToOpenAI(tool *tools.Tool) openai.ChatCompletionToolParam
