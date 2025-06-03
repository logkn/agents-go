package types

import (
	"log"

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

type UserMessage struct {
	Content string
	Name    string
}

func (m UserMessage) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	msg := openai.UserMessage(m.Content)
	msg.OfUser.Name = openai.String(m.Name)
	return msg
}

type AssistantMessage struct {
	Content   string
	ToolCalls []ToolCall
	Name      string
}

func (m AssistantMessage) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	msg := openai.AssistantMessage(m.Content)
	msg.OfAssistant.Name = openai.String(m.Name)
	openAIToolCalls := utils.MapSlice(m.ToolCalls, ToolCall.ToOpenAI)
	msg.OfAssistant.ToolCalls = openAIToolCalls
	return msg
}

func AssistantMessageFromOpenAI(msg openai.ChatCompletionMessage, name string) Message {
	toolCalls := utils.MapSlice(msg.ToolCalls, ToolCallFromOpenAI)
	ourmessage := NewAssistantMessage(
		msg.Content,
		name,
		toolCalls,
	)
	return ourmessage
}

type SystemMessage struct {
	Content string
}

func (m SystemMessage) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	msg := openai.SystemMessage(m.Content)
	return msg
}

type ToolMessage struct {
	ID      string
	Content any
}

func (m ToolMessage) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	log.Println("Converting ToolMessage to OpenAI format:", m)
	return openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(utils.AsString(m.Content)),
			},
			ToolCallID: m.ID,
		},
	}
}

type Message struct {
	Role Role
	UserMessage
	AssistantMessage
	SystemMessage
	ToolMessage
}

func (m Message) ToOpenAI() openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case User:
		return m.UserMessage.ToOpenAI()
	case Assistant:
		return m.AssistantMessage.ToOpenAI()
	case System:
		return m.SystemMessage.ToOpenAI()
	case Tool:
		return m.ToolMessage.ToOpenAI()
	}
	return openai.ChatCompletionMessageParamUnion{}
}

func NewUserMessage(content string) Message {
	return Message{
		Role: User,
		UserMessage: UserMessage{
			Content: content,
		},
	}
}

func NewAssistantMessage(content, name string, toolcalls []ToolCall) Message {
	return Message{
		Role: Assistant,
		AssistantMessage: AssistantMessage{
			Content:   content,
			ToolCalls: toolcalls,
			Name:      name,
		},
	}
}

func NewSystemMessage(content string) Message {
	return Message{
		Role: System,
		SystemMessage: SystemMessage{
			Content: content,
		},
	}
}

func NewToolMessage(id string, content any) Message {
	return Message{
		Role: Tool,
		ToolMessage: ToolMessage{
			ID:      id,
			Content: content,
		},
	}
}
