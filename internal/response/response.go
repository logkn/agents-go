package response

// ResponseType represents different types of agent responses
type ResponseType string

const (
	ResponseTypeThought      ResponseType = "thought"
	ResponseTypeIntermediate ResponseType = "intermediate"
	ResponseTypeToolCall     ResponseType = "tool_call"
	ResponseTypeFinal        ResponseType = "final"
	ResponseTypeHandoff      ResponseType = "handoff"
)

// AgentResponse represents a response from an agent
type AgentResponse struct {
	Type     ResponseType   `json:"type"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	ToolCall *ToolCall      `json:"tool_call,omitempty"`
	Handoff  *AgentHandoff  `json:"handoff,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters"`
	Result     any            `json:"result,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// AgentHandoff represents transferring control to another agent
type AgentHandoff struct {
	ToAgent string `json:"to_agent"`
	Reason  string `json:"reason"`
	Context string `json:"context"`
}