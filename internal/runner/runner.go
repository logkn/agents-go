package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type ToolResult struct {
	Name       string
	Content    any
	ToolCallID string
}

type AgentEvent struct {
	Timestamp    time.Time
	OfToken      string
	OfMessage    *types.Message
	OfToolResult ToolResult
	OfError      error
}

func (e *AgentEvent) Token() (string, bool) {
	return e.OfToken, e.OfToken != ""
}

func (e *AgentEvent) Message() (*types.Message, bool) {
	if e.OfMessage != nil {
		return e.OfMessage, true
	}
	return nil, false
}

func (e *AgentEvent) ToolResult() (ToolResult, bool) {
	if e.OfToolResult.Name != "" {
		return e.OfToolResult, true
	}
	return ToolResult{}, false
}

func (e *AgentEvent) Error() (error, bool) {
	if e.OfError != nil {
		return e.OfError, true
	}
	return nil, false
}

func tokenEvent(token string) AgentEvent {
	return AgentEvent{
		OfToken:   token,
		Timestamp: time.Now(),
	}
}

func messageEvent(message types.Message) AgentEvent {
	return AgentEvent{
		OfMessage: &message,
		Timestamp: time.Now(),
	}
}

func toolEvent(toolResult ToolResult) AgentEvent {
	return AgentEvent{
		OfToolResult: toolResult,
		Timestamp:    time.Now(),
	}
}

type AgentResponse struct {
	// The event bus
	events chan AgentEvent
	// An accumulator of events
	pastEvents   []AgentEvent
	pastMessages []types.Message
}

func newAgentResponse(ch chan AgentEvent, pastMessages []types.Message) *AgentResponse {
	return &AgentResponse{
		events:       ch,
		pastEvents:   []AgentEvent{},
		pastMessages: pastMessages,
	}
}

func (ar *AgentResponse) Stream() <-chan AgentEvent {
	outchan := make(chan AgentEvent, 10)
	go func() {
		defer close(outchan)
		for event := range ar.events {
			ar.pastEvents = append(ar.pastEvents, event)
			outchan <- event
		}
	}()
	return outchan
}

func (ar *AgentResponse) waitForStreamCompletion() {
	// Wait for the stream to complete
	for range ar.Stream() {
	}
	fmt.Println("Waited for stream completion")
}

func (ar *AgentResponse) Response() types.Message {
	allMessages := ar.FinalConversation()
	lastMessage := allMessages[len(allMessages)-1]

	return lastMessage
}

func (ar *AgentResponse) FinalConversation() []types.Message {
	ar.waitForStreamCompletion()
	finalMessages := make([]types.Message, 0, len(ar.pastMessages)+len(ar.pastEvents))
	finalMessages = append(finalMessages, ar.pastMessages...)
	return finalMessages
}

func Run(agent agents.Agent, input string) AgentResponse {
	message := input
	messages := []types.Message{
		types.NewSystemMessage(agent.Instructions),
		types.NewUserMessage(message),
	}

	var client openai.Client
	if agent.Model.BaseUrl != "" {
		client = openai.NewClient(
			option.WithBaseURL(agent.Model.BaseUrl),
		)
	} else {
		client = openai.NewClient()
	}
	// check that the model exists
	if _, err := client.Models.Get(context.TODO(), agent.Model.Model); err != nil {
		fmt.Printf("Error getting model %s: %v\n", agent.Model.Model, err)
		return *newAgentResponse(nil, messages)
	}

	openAITools := make([]openai.ChatCompletionToolParam, len(agent.Tools))
	for i, tool := range agent.Tools {
		openAITools[i] = tool.ToOpenAITool()
	}

	eventChannel := make(chan AgentEvent, 10)
	agentResponse := newAgentResponse(eventChannel, messages)

	go func() {
		for {
			openaiMessages := utils.MapSlice(messages, types.Message.ToOpenAI)
			fmt.Println(utils.AsString(openaiMessages))
			params := openai.ChatCompletionNewParams{
				Messages: openaiMessages,
				Model:    agent.Model.Model,
				Tools:    openAITools,
			}
			stream := client.Chat.Completions.NewStreaming(context.TODO(), params)
			acc := openai.ChatCompletionAccumulator{}
			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)

				if len(chunk.Choices) > 0 {
					token := chunk.Choices[0].Delta.Content
					eventChannel <- tokenEvent(token)
				}
			}
			choices := acc.Choices
			// if no choices, break the loop
			if len(choices) == 0 {
				fmt.Println("No choices in the stream, breaking the loop")
				break
			}
			openaimsg := choices[0].Message
			log.Println("OpenAI message:", utils.JsonDumpsObj(openaimsg))
			msg := types.AssistantMessageFromOpenAI(openaimsg, agent.Name)
			log.Println("Converted message:", utils.JsonDumpsObj(msg))
			messages = append(messages, msg)

			eventChannel <- messageEvent(msg)

			toolcalls := msg.ToolCalls

			if len(toolcalls) == 0 {
				break
			}

			for _, toolcall := range toolcalls {
				funcname := toolcall.Name
				// get the tool by name
				for _, tool := range agent.Tools {
					if tool.CompleteName() == funcname {
						result := tool.RunOnArgs(toolcall.Args)
						toolmessage := types.NewToolMessage(toolcall.ID, result)
						messages = append(messages, toolmessage)
						eventChannel <- toolEvent(ToolResult{
							Name:       tool.CompleteName(),
							Content:    result,
							ToolCallID: toolcall.ID,
						})
						break
					}
				}
			}
		}
		close(eventChannel)
	}()

	return *agentResponse
}
