package types

import (
	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
)

type ToolCall struct {
	ID   string
	Name string
	Args string
}

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

func ToolCallFromOpenAI(call openai.ChatCompletionMessageToolCall) ToolCall {
	return ToolCall{
		ID:   call.ID,
		Name: call.Function.Name,
		Args: call.Function.Arguments,
	}
}

type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content,omitempty"`
	Name      string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ID        string     `json:"id,omitempty"` // for tool messages
}

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

func NewUserMessage(content string) Message {
	return Message{
		Role:    User,
		Content: content,
	}
}

func NewAssistantMessage(content, name string, toolcalls []ToolCall) Message {
	return Message{
		Role:      Assistant,
		Content:   content,
		Name:      name,
		ToolCalls: toolcalls,
	}
}

func NewSystemMessage(content string) Message {
	return Message{
		Role:    System,
		Content: content,
	}
}

func NewToolMessage(id string, content any) Message {
	return Message{
		Role:    Tool,
		ID:      id,
		Content: utils.AsString(content),
	}
}

func AssistantMessageFromOpenAI(msg openai.ChatCompletionMessage, name string) Message {
	toolCalls := utils.MapSlice(msg.ToolCalls, ToolCallFromOpenAI)
	return NewAssistantMessage(
		msg.Content,
		name,
		toolCalls,
	)
}
