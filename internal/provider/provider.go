package provider

import (
	"context"

	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
)

// Message represents a conversation message
type Message struct {
	Role      string              `json:"role"`
	Content   string              `json:"content"`
	ToolID    string              `json:"tool_id,omitempty"`
	ToolCalls []response.ToolCall `json:"tool_calls,omitempty"`
	Data      any                 `json:"data,omitempty"`
}

// LLM abstracts different LLM providers
type LLM interface {
	GenerateResponse(ctx context.Context, messages []Message, tools []*tools.Tool) (*LLMResponse, error)
}

type Streams interface {
	StreamResponse(ctx context.Context, messages []Message, tools []*tools.Tool) (<-chan LLMResponseItem, error)
}

// LLMResponse represents a response from an LLM
type LLMResponse struct {
	Content   string              `json:"content"`
	ToolCalls []response.ToolCall `json:"tool_calls,omitempty"`
	Finished  bool                `json:"finished"`
	Usage     *TokenUsage         `json:"usage,omitempty"`
}

type LLMResponseItem struct {
	LLMResponse
	Delta string `json:"delta"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
