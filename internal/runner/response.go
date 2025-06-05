package runner

import "github.com/logkn/agents-go/internal/types"

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
