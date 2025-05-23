package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIProvider implements LLMProvider using OpenAI's API
type OpenAIProvider struct {
	client openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider instance
// If apiKey is empty, it will attempt to use the OPENAI_API_KEY environment variable
func NewOpenAIProvider(apiKey string, model string) *OpenAIProvider {
	if model == "" {
		model = string(openai.ChatModelGPT4o)
	}

	// If no API key provided, try to get it from environment variable
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIProvider{
		client: client,
		model:  model,
	}
}

// GenerateResponse generates a response using OpenAI's chat completion API
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, messages []Message, tools []tools.Tool) (*LLMResponse, error) {
	// Convert our messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			if len(msg.ToolID) > 0 {
				// This is a tool call response
				openaiMessages = append(openaiMessages, openai.ToolMessage(msg.ToolID, msg.Content))
			} else {
				openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
			}
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		}
	}

	// Convert our tools to OpenAI function schema format
	openaiTools := make([]openai.ChatCompletionToolParam, 0, len(tools))

	for _, tool := range tools {
		schema := tool.JSONSchema()

		// Convert our schema to OpenAI function format
		functionDef := openai.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
		}

		// Convert parameters schema
		if schema != nil {
			functionDef.Parameters = schema
		}

		openaiTools = append(openaiTools, openai.ChatCompletionToolParam{
			Type:     "function",
			Function: functionDef,
		})
	}

	// Prepare the request parameters
	params := openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    openai.ChatModel(p.model),
	}

	// Add tools if available
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
	}

	// Make the API call
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := completion.Choices[0]

	// Parse the response
	llmResponse := &LLMResponse{
		Content:  choice.Message.Content,
		Finished: string(choice.FinishReason) == "stop",
	}

	// Handle token usage if available
	if completion.Usage.PromptTokens > 0 {
		llmResponse.Usage = &TokenUsage{
			InputTokens:  int(completion.Usage.PromptTokens),
			OutputTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:  int(completion.Usage.TotalTokens),
		}
	}

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]response.ToolCall, 0, len(choice.Message.ToolCalls))

		for _, toolCall := range choice.Message.ToolCalls {
			// Parse function arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool call arguments: %w", err)
			}

			toolCalls = append(toolCalls, response.ToolCall{
				ID:         toolCall.ID,
				Name:       toolCall.Function.Name,
				Parameters: args,
			})
		}

		llmResponse.ToolCalls = toolCalls
		llmResponse.Finished = false // Tool calls mean we're not finished
	}

	return llmResponse, nil
}

// SupportsStreaming returns whether this provider supports streaming responses
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true // OpenAI supports streaming, though not implemented yet
}
