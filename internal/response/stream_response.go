package response

type ResponseItemType string

const (
	ResponseItemTypeToken    ResponseItemType = "token"
	ResponseItemTypeThought  ResponseItemType = "thought"
	ResponseItemTypeToolCall ResponseItemType = "tool_call"
	ResponseItemTypeFinal    ResponseItemType = "final"
	ResponseItemTypeHandoff  ResponseItemType = "handoff"
)

type AgentResponseItem struct {
	Type           ResponseItemType `json:"type"`
	Content        string           `json:"content"`
	StructuredData any              `json:"structured_data,omitempty"`
	Metadata       map[string]any   `json:"metadata,omitempty"`
	ToolCall       *ToolCall        `json:"tool_call,omitempty"`
	Handoff        *AgentHandoff    `json:"handoff,omitempty"`
}
