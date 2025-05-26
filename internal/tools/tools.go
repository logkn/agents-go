package tools

import (
	"reflect"

	"github.com/logkn/agents-go/internal/events"
	"github.com/logkn/agents-go/internal/utils"
)

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

func (t Tool) WithName(name string) Tool {
	t.name = name
	return t
}

func (t Tool) WithDescription(description string) Tool {
	t.description = description
	return t
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
	// Return the name of the struct implementing ToolDef
	defType := reflect.TypeOf(t.def)

	// Handle pointer types
	if defType.Kind() == reflect.Ptr {
		defType = defType.Elem()
	}
	return defType.Name()
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

func (t *Tool) Schema() map[string]any {
	return utils.GenerateSchema(t.def)
}
