package tools2

import "github.com/logkn/agents-go/internal/events"

// type Tool interface {
// 	Name() string
// 	Description() string
// 	JSONSchema() map[string]any
// 	Execute(state any, events events.EventBus) (any, error)
// }

type ToolDef interface {
	Execute(state any, events events.EventBus) (any, error)
}

type Tool struct {
	name        string
	description string
	def         ToolDef
}

func NewTool(def ToolDef) Tool {
	return Tool{
		def: def,
	}
}

func (t *Tool) WithName(name string) Tool {
	t.name = name
	return *t
}

func (t *Tool) WithDescription(description string) Tool {
	t.description = description
	return *t
}

func (t *Tool) defaultDescription() string {
	return "Calls the " + t.Name() + " tool."
}

func (t *Tool) Description() string {
	// if explicitly set, use that
	// otherwise, use the default description
	description := t.description
	if description == "" {
		description = t.defaultDescription()
	}
	return description
}

func (t *Tool) defaultName() string {
	// TODO: Return the name of the struct implementing ToolDef
}

func (t *Tool) Name() string {
	// if explicitly set, use that
	// otherwise, use the name from the struct implementing ToolDef
	name := t.name
	if name == "" {
		name = t.defaultName()
	}
	return name
}

func (t *Tool) JSONSchema() map[string]any {
	// TODO: Implement JSON schema generation
}

// Example

type WebSearch struct {
	Query string `json:"query" description:"The query to search for"`
}

func (w WebSearch) Execute(state any, events events.EventBus) (any, error) {
	res := "Here are your search results for " + w.Query
	return res, nil
}
