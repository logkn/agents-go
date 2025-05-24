package tools2

import (
	"reflect"

	"github.com/logkn/agents-go/internal/events"
)

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
	// Return the name of the struct implementing ToolDef
	return reflect.TypeOf(t.def).Name()
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
	defType := reflect.TypeOf(t.def)
	
	// Handle pointer types
	if defType.Kind() == reflect.Ptr {
		defType = defType.Elem()
	}
	
	properties := make(map[string]any)
	required := []string{}
	
	// Iterate through struct fields
	for i := 0; i < defType.NumField(); i++ {
		field := defType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		
		// Parse json tag (handle "fieldname,omitempty" format)
		fieldName := jsonTag
		if commaIdx := len(jsonTag); commaIdx > 0 {
			for j, r := range jsonTag {
				if r == ',' {
					fieldName = jsonTag[:j]
					break
				}
			}
		}
		
		// Build field schema
		fieldSchema := make(map[string]any)
		
		// Add type based on Go type
		switch field.Type.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			fieldSchema["type"] = "number"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		case reflect.Slice:
			fieldSchema["type"] = "array"
		case reflect.Map, reflect.Struct:
			fieldSchema["type"] = "object"
		default:
			fieldSchema["type"] = "string"
		}
		
		// Add description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema["description"] = desc
		}
		
		properties[fieldName] = fieldSchema
		
		// Check if field is required (no omitempty tag)
		if !contains(jsonTag, "omitempty") {
			required = append(required, fieldName)
		}
	}
	
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Example

type WebSearch struct {
	Query string `json:"query" description:"The query to search for"`
}

func (w WebSearch) Execute(state any, events events.EventBus) (any, error) {
	res := "Here are your search results for " + w.Query
	return res, nil
}
