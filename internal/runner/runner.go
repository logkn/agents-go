package runner

import (
	"context"
	"time"

	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// ToolResult represents the output of a tool call executed by an agent.
// Name is the tool's name, Content is the returned value and ToolCallID is the
// identifier associated with the call.
type ToolResult struct {
	Name       string
	Content    any
	ToolCallID string
}

// AgentEvent is a generic event emitted during a run. Only one of the fields is
// typically populated depending on what occurred.
type AgentEvent struct {
	Timestamp    time.Time
	OfToken      string
	OfMessage    *types.Message
	OfToolResult ToolResult
	OfError      error
}

// Input represents the starting data for a run. Exactly one field should be
// populated.
type Input struct {
	// OfString initiates a new conversation with this user prompt.
	OfString string
	// OfMessages continues an existing conversation.
	OfMessages []types.Message
}

// Token returns the token contained in the event if present.
func (e *AgentEvent) Token() (string, bool) {
	return e.OfToken, e.OfToken != ""
}

// Message returns the message contained in the event if present.
func (e *AgentEvent) Message() (*types.Message, bool) {
	if e.OfMessage != nil {
		return e.OfMessage, true
	}
	return nil, false
}

// ToolResult returns the tool output carried by the event if present.
func (e *AgentEvent) ToolResult() (ToolResult, bool) {
	if e.OfToolResult.Name != "" {
		return e.OfToolResult, true
	}
	return ToolResult{}, false
}

// Error returns the error stored in the event if any.
func (e *AgentEvent) Error() (error, bool) {
	if e.OfError != nil {
		return e.OfError, true
	}
	return nil, false
}

// tokenEvent creates a new AgentEvent containing a token.
func tokenEvent(token string) AgentEvent {
	return AgentEvent{
		OfToken:   token,
		Timestamp: time.Now(),
	}
}

// messageEvent creates a new AgentEvent carrying a message.
func messageEvent(message types.Message) AgentEvent {
	return AgentEvent{
		OfMessage: &message,
		Timestamp: time.Now(),
	}
}

// toolEvent creates a new AgentEvent for a tool result.
func toolEvent(toolResult ToolResult) AgentEvent {
	return AgentEvent{
		OfToolResult: toolResult,
		Timestamp:    time.Now(),
	}
}

// AgentResponse collects all events produced during a run and exposes helper
// methods to access them.
type AgentResponse struct {
	// events is the internal event bus used during streaming.
	events chan AgentEvent
	// pastEvents stores everything that has already been observed.
	pastEvents   []AgentEvent
	pastMessages []types.Message
}

// newAgentResponse creates an AgentResponse bound to the provided channel.
func newAgentResponse(ch chan AgentEvent, pastMessages []types.Message) *AgentResponse {
	return &AgentResponse{
		events:       ch,
		pastEvents:   []AgentEvent{},
		pastMessages: pastMessages,
	}
}

// Stream returns a channel that yields events in real time while also
// accumulating them for later retrieval.
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

// waitForStreamCompletion drains the event stream until it closes.
func (ar *AgentResponse) waitForStreamCompletion() {
	for range ar.Stream() {
	}
}

// Response returns the last message produced in the conversation.
func (ar *AgentResponse) Response() types.Message {
	allMessages := ar.FinalConversation()
	lastMessage := allMessages[len(allMessages)-1]

	return lastMessage
}

// FinalConversation waits for streaming to finish and returns every message
// that occurred during the run.
func (ar *AgentResponse) FinalConversation() []types.Message {
	ar.waitForStreamCompletion()
	finalMessages := make([]types.Message, 0, len(ar.pastMessages)+len(ar.pastEvents))
	finalMessages = append(finalMessages, ar.pastMessages...)
	return finalMessages
}

// Run executes the agent against the provided input and returns an
// AgentResponse for consuming the results.
// Run executes the agent and streams events back through an AgentResponse.
// If input.OfMessages is provided it is treated as the existing conversation
// history. Otherwise a new conversation is started with input.OfString as the
// user prompt.
func Run(agent agents.Agent, input Input) (AgentResponse, error) {
	var messages []types.Message
	switch {
	case len(input.OfMessages) > 0:
		messages = input.OfMessages
	default:
		messages = []types.Message{
			types.NewSystemMessage(agent.Instructions),
			types.NewUserMessage(input.OfString),
		}
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
	// if _, err := client.Models.Get(context.TODO(), agent.Model.Model); err != nil {
	// 	return AgentResponse{}, err
	// }

	openAITools := make([]openai.ChatCompletionToolParam, len(agent.Tools))
	for i, tool := range agent.Tools {
		openAITools[i] = tool.ToOpenAITool()
	}

	eventChannel := make(chan AgentEvent, 10)
	agentResponse := newAgentResponse(eventChannel, messages)

	go func() {
		for {
			openaiMessages := utils.MapSlice(messages, types.Message.ToOpenAI)
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
				break
			}
			openaimsg := choices[0].Message
			msg := types.AssistantMessageFromOpenAI(openaimsg, agent.Name)
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
						eventChannel <- messageEvent(toolmessage)
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

	return *agentResponse, nil
}
