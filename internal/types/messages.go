package types

import (
	"github.com/logkn/agents-go/internal/utils"
)

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

type ResponseFormat struct {
	String     bool
	Structured StructuredOutputFormat
}

type StructuredOutputFormat struct {
	Name        string
	Description string
	Zero        *Struct
}

// Schema() returns the JSON schmea for the response format
// The top level of this schema is:
//   - type: "object"
//   - properties: properties of the struct
//   - required: array of required properties
//   - strict: true
func (r *StructuredOutputFormat) Schema() map[string]any {
	return utils.GenerateSchema(r.Zero)
}
