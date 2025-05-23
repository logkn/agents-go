package response

type ResponseItemType string

const (
	ResponseItemTypeToken    ResponseItemType = "token"
	ResponseItemTypeThought  ResponseItemType = "thought"
	ResponseItemTypeToolCall ResponseItemType = "tool_call"
	ResponseItemTypeFinal    ResponseItemType = "final"
	ResponseItemTypeHandoff  ResponseItemType = "handoff"
)
