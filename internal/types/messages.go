package types

// "user", "assistant", "system", or "tool"
type Role string

const (
	User      Role = "user"
	Assistant Role = "assistant"
	System    Role = "system"
	Tool      Role = "tool"
)

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content,omitempty"`
	Name      string     `json:"name,omitempty"`
	Toolcalls []ToolCall `json:"tool_calls,omitempty"`
}

type MessageDelta struct {
	Role  Role   `json:"role"`
	Name  string `json:"name,omitempty"`
	Delta string `json:"delta,omitempty"`
}
