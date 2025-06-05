package runner

import (
	"time"

	"github.com/logkn/agents-go/internal/types"
)

// AgentEvent is a generic event emitted during a run. Only one of the fields is
// typically populated depending on what occurred.
type AgentEvent struct {
	Timestamp    time.Time
	OfToken      string
	OfMessage    *types.Message
	OfToolResult ToolResult
	OfError      error
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

func errorEvent(err error) AgentEvent {
	return AgentEvent{
		OfError:   err,
		Timestamp: time.Now(),
	}
}
