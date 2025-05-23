package provider

import (
	"context"

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
