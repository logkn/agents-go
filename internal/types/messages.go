package types

import (
	"log/slog"

	"github.com/logkn/agents-go/internal/utils"
	"github.com/logkn/agents-go/tools"
	"github.com/openai/openai-go"
)

// ToolCall represents an invocation of a tool by the language model.
type ToolCall struct {
	ID   string
	Name string
	Args string
}

// ToOpenAI converts the tool call into the OpenAI SDK representation.
func (t ToolCall) ToOpenAI() openai.ChatCompletionMessageToolCallParam {
	msg := openai.ChatCompletionMessageToolCallParam{
		ID: t.ID,
		Function: openai.ChatCompletionMessageToolCallFunctionParam{
			Arguments: t.Args,
			Name:      t.Name,
		},
	}
	return msg
}

// ToolCallFromOpenAI converts an OpenAI tool call into our internal type.
func ToolCallFromOpenAI(call openai.ChatCompletionMessageToolCall) ToolCall {
	return ToolCall{
		ID:   call.ID,
		Name: call.Function.Name,
		Args: call.Function.Arguments,
	}
}

// Message represents a single message in the conversation transcript.
type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content,omitempty"`
	Name      string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ID        string     `json:"id,omitempty"` // for tool messages
}

// ToOpenAI converts the message into the OpenAI SDK representation.
func (m Message) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case User:
		msg := openai.UserMessage(m.Content)
		if m.Name != "" {
			msg.OfUser.Name = openai.String(m.Name)
		}
		return msg
	case Assistant:
		msg := openai.AssistantMessage(m.Content)
		if m.Name != "" {
			msg.OfAssistant.Name = openai.String(m.Name)
		}
		if len(m.ToolCalls) > 0 {
			openAIToolCalls := utils.MapSlice(m.ToolCalls, ToolCall.ToOpenAI)
			msg.OfAssistant.ToolCalls = openAIToolCalls
		}
		return msg
	case System:
		return openai.SystemMessage(m.Content)
	case Tool:
		return openai.ChatCompletionMessageParamUnion{
			OfTool: &openai.ChatCompletionToolMessageParam{
				Content: openai.ChatCompletionToolMessageParamContentUnion{
					OfString: openai.String(utils.AsString(m.Content)),
				},
				ToolCallID: m.ID,
			},
		}
	}
	return openai.ChatCompletionMessageParamUnion{}
}

// NewUserMessage creates a user message with the provided content.
func NewUserMessage(content string) Message {
	return Message{
		Role:    User,
		Content: content,
	}
}

// NewAssistantMessage constructs an assistant message with optional tool calls.
func NewAssistantMessage(content, name string, toolcalls []ToolCall) Message {
	return Message{
		Role:      Assistant,
		Content:   content,
		Name:      name,
		ToolCalls: toolcalls,
	}
}

// NewSystemMessage creates a system message with the given content.
func NewSystemMessage(content string) Message {
	return Message{
		Role:    System,
		Content: content,
	}
}

// NewToolMessage creates a message that captures the output of a tool.
func NewToolMessage(id string, content any) Message {
	return Message{
		Role:    Tool,
		ID:      id,
		Content: utils.AsString(content),
	}
}

// AssistantMessageFromOpenAI converts an OpenAI assistant message into our internal structure.
func AssistantMessageFromOpenAI(msg openai.ChatCompletionMessage, name string) Message {
	toolCalls := utils.MapSlice(msg.ToolCalls, ToolCallFromOpenAI)
	return NewAssistantMessage(
		msg.Content,
		name,
		toolCalls,
	)
}

// ModelConfig contains configuration details for an LLM model.
// Model is the identifier of the model to use and BaseUrl is an optional
// override for the API base URL.
type ModelConfig struct {
	Model   string
	BaseUrl string
}

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []*Agent
	Logger       *slog.Logger
}
