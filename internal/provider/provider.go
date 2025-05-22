package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
)

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	ToolID  string `json:"tool_id,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// LLMProvider abstracts different LLM providers
type LLMProvider interface {
	GenerateResponse(ctx context.Context, messages []Message, tools []tools.Tool) (*LLMResponse, error)
	SupportsStreaming() bool
}

// LLMResponse represents a response from an LLM
type LLMResponse struct {
	Content   string              `json:"content"`
	ToolCalls []response.ToolCall `json:"tool_calls,omitempty"`
	Finished  bool                `json:"finished"`
	Usage     *TokenUsage         `json:"usage,omitempty"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// MockLLMProvider is a simple implementation for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, messages []Message, tools []tools.Tool) (*LLMResponse, error) {
	// Simple mock - just echo the last message
	if len(messages) == 0 {
		return &LLMResponse{
			Content:  "Hello! How can I help you?",
			Finished: true,
		}, nil
	}

	lastMsg := messages[len(messages)-1]

	// If user mentions weather and email together, use both tools in parallel
	content := strings.ToLower(lastMsg.Content)
	if strings.Contains(content, "weather") && strings.Contains(content, "email") {
		var toolCalls []response.ToolCall

		for _, tool := range tools {
			if tool.Name() == "get_weather" {
				toolCalls = append(toolCalls, response.ToolCall{
					ID:   "call_weather",
					Name: "get_weather",
					Parameters: map[string]any{
						"city": "San Francisco",
					},
				})
			}
			if tool.Name() == "send_email" {
				toolCalls = append(toolCalls, response.ToolCall{
					ID:   "call_email",
					Name: "send_email",
					Parameters: map[string]any{
						"to":      "user@example.com",
						"subject": "Weather Update",
						"body":    "Here's your weather update!",
					},
				})
			}
		}

		if len(toolCalls) > 0 {
			return &LLMResponse{
				Content:   "I'll check the weather and send you an email with the information.",
				ToolCalls: toolCalls,
				Finished:  false,
			}, nil
		}
	}

	// If user mentions weather, use the weather tool
	if strings.Contains(content, "weather") {
		for _, tool := range tools {
			if tool.Name() == "get_weather" {
				return &LLMResponse{
					Content: "I'll check the weather for you.",
					ToolCalls: []response.ToolCall{
						{
							ID:   "call_1",
							Name: "get_weather",
							Parameters: map[string]any{
								"city": "San Francisco",
							},
						},
					},
					Finished: false,
				}, nil
			}
		}
	}

	return &LLMResponse{
		Content:  fmt.Sprintf("I received: %s", lastMsg.Content),
		Finished: true,
	}, nil
}

func (m *MockLLMProvider) SupportsStreaming() bool {
	return false
}
