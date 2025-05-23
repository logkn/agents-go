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

// OpenAIProvider implements LLM using OpenAI's API
type OpenAIProvider struct {
	client openai.Client
	Model  string
}

// NewOpenAIProvider creates a new OpenAI provider instance
// If apiKey is empty, it will attempt to use the OPENAI_API_KEY environment variable
func NewOpenAIProvider(model string) LLM {
	if model == "" {
		model = string(openai.ChatModelGPT4o)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return OpenAIProvider{
		client: client,
		Model:  model,
	}
}

// GenerateResponse generates a response using OpenAI's chat completion API
func (p OpenAIProvider) GenerateResponse(ctx context.Context, messages []Message, tools []*tools.Tool) (*LLMResponse, error) {
	// Convert our messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls - use manual construction
				toolCalls := make([]openai.ChatCompletionMessageToolCall, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					// Convert parameters back to JSON string
					argsJSON, err := json.Marshal(tc.Parameters)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal tool call parameters: %w", err)
					}

					toolCalls[i] = openai.ChatCompletionMessageToolCall{
						ID:   tc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunction{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					}
				}

				assistantMsg := openai.ChatCompletionMessage{
					Role:      "assistant",
					Content:   msg.Content,
					ToolCalls: toolCalls,
				}
				openaiMessages = append(openaiMessages, assistantMsg.ToParam())
			} else {
				openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
			}
		case "tool":
			openaiMessages = append(openaiMessages, openai.ToolMessage(msg.Content, msg.ToolID))
		case "system":
			openaiMessages = append(openaiMessages, openai.SystemMessage(msg.Content))
		}
	}

	// Convert our tools to OpenAI function schema format
	openaiTools := make([]openai.ChatCompletionToolParam, 0, len(tools))

	for _, ptool := range tools {
		tool := *ptool
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
		Model:    openai.ChatModel(p.Model),
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
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool call arguments: %w", err)
			}

			toolID := toolCall.ID

			toolCalls = append(toolCalls, response.ToolCall{
				ID:         toolID,
				Name:       toolCall.Function.Name,
				Parameters: args,
			})
		}

		llmResponse.ToolCalls = toolCalls
		llmResponse.Finished = false // Tool calls mean we're not finished
	}

	return llmResponse, nil
}

func (p OpenAIProvider) StreamResponse(ctx context.Context, messages []Message, tools []*tools.Tool) (<-chan LLMResponseItem, error) {
	openaiMessages, err := convertMessages(&messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages to OpenAI format: %w", err)
	}

	openaiTools := convertTools(tools)

	// Prepare the request parameters for streaming
	params := openai.ChatCompletionNewParams{
		Messages: openaiMessages,
		Model:    openai.ChatModel(p.Model),
	}

	// Add tools if available
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
	}

	// Create the streaming request
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	// Create response channel
	responseCh := make(chan LLMResponseItem)

	// Handle streaming in a goroutine
	go func() {
		defer close(responseCh)
		defer stream.Close()

		fullContent := ""
		accumulatedToolCalls := []response.ToolCall{}
		usage := TokenUsage{}

		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]
			delta := choice.Delta

			// Handle content delta
			if delta.Content != "" {
				fullContent += delta.Content

				item := LLMResponseItem{
					LLMResponse: LLMResponse{
						Content:  fullContent,
						Finished: false,
					},
					Delta: delta.Content,
				}

				select {
				case responseCh <- item:
				case <-ctx.Done():
					return
				}
			}

			// Handle tool calls
			if len(delta.ToolCalls) > 0 {
				for _, toolCall := range delta.ToolCalls {
					// Find or create the tool call in our accumulated list
					toolCallIndex := int(toolCall.Index)

					// Ensure we have enough space in the slice
					for len(accumulatedToolCalls) <= toolCallIndex {
						accumulatedToolCalls = append(accumulatedToolCalls, response.ToolCall{})
					}

					// Update the accumulated tool call
					if toolCall.ID != "" {
						accumulatedToolCalls[toolCallIndex].ID = toolCall.ID
					}
					if toolCall.Function.Name != "" {
						accumulatedToolCalls[toolCallIndex].Name = toolCall.Function.Name
					}
					if toolCall.Function.Arguments != "" {
						// Accumulate arguments
						if accumulatedToolCalls[toolCallIndex].Parameters == nil {
							accumulatedToolCalls[toolCallIndex].Parameters = make(map[string]any)
						}

						// Try to parse the accumulated arguments
						var args map[string]any
						if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
							accumulatedToolCalls[toolCallIndex].Parameters = args
						}
					}
				}

				item := LLMResponseItem{
					LLMResponse: LLMResponse{
						Content:   fullContent,
						ToolCalls: accumulatedToolCalls,
						Finished:  false,
					},
					Delta: "",
				}

				select {
				case responseCh <- item:
				case <-ctx.Done():
					return
				}
			}

			// Handle usage information
			if chunk.Usage.PromptTokens > 0 {
				usage = TokenUsage{
					InputTokens:  int(chunk.Usage.PromptTokens),
					OutputTokens: int(chunk.Usage.CompletionTokens),
					TotalTokens:  int(chunk.Usage.TotalTokens),
				}
			}

			// Check if we're finished
			if string(choice.FinishReason) == "stop" || string(choice.FinishReason) == "tool_calls" {
				finalItem := LLMResponseItem{
					LLMResponse: LLMResponse{
						Content:   fullContent,
						ToolCalls: accumulatedToolCalls,
						Finished:  true,
						Usage:     &usage,
					},
					Delta: "",
				}

				select {
				case responseCh <- finalItem:
				case <-ctx.Done():
					return
				}
			}
		}

		// Check for streaming errors
		if err := stream.Err(); err != nil {
			// Send error by closing the channel - the caller should handle this
			return
		}
	}()

	return responseCh, nil
}
