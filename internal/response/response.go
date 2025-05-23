package response

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ResponseType represents different types of agent responses
type ResponseType string

const (
	ResponseTypeThought  ResponseType = "thought"
	ResponseTypeToolCall ResponseType = "tool_call"
	ResponseTypeFinal    ResponseType = "final"
	ResponseTypeHandoff  ResponseType = "handoff"
)

// StructuredOutput defines the interface for structured output schemas
type StructuredOutput interface {
	JSONSchema() map[string]any
	ValidateAndUnmarshal(data []byte) (any, error)
}

// StructuredOutputSchema implements StructuredOutput for a given type
type StructuredOutputSchema[T any] struct {
	schema map[string]any
}

// NewStructuredOutputSchema creates a new structured output schema for type T
func NewStructuredOutputSchema[T any]() *StructuredOutputSchema[T] {
	var zero T
	schema := generateJSONSchema(reflect.TypeOf(zero))
	return &StructuredOutputSchema[T]{schema: schema}
}

// JSONSchema returns the JSON schema for the structured output
func (s *StructuredOutputSchema[T]) JSONSchema() map[string]any {
	return s.schema
}

// ValidateAndUnmarshal validates the JSON data against the schema and unmarshals it
func (s *StructuredOutputSchema[T]) ValidateAndUnmarshal(data []byte) (any, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal structured output: %w", err)
	}
	return result, nil
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	Type           ResponseType   `json:"type"`
	Content        string         `json:"content"`
	StructuredData any            `json:"structured_data,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	ToolCall       *ToolCall      `json:"tool_call,omitempty"`
	Handoff        *AgentHandoff  `json:"handoff,omitempty"`
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

// generateJSONSchema generates a JSON schema from a Go type using reflection
func generateJSONSchema(t reflect.Type) map[string]any {
	schema := map[string]any{
		"type": "object",
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]any{"type": getJSONType(t)}
	}

	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			// Parse json tag (e.g., "name,omitempty")
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldSchema := generateJSONSchema(field.Type)

		// Add description from struct tag if available
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema["description"] = desc
		}

		properties[fieldName] = fieldSchema

		// Check if field is required (no omitempty and not a pointer)
		if !hasOmitEmpty(jsonTag) && field.Type.Kind() != reflect.Ptr {
			required = append(required, fieldName)
		}
	}

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// getJSONType returns the JSON type for a Go type
func getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// hasOmitEmpty checks if the json tag contains omitempty
func hasOmitEmpty(jsonTag string) bool {
	if jsonTag == "" {
		return false
	}
	parts := strings.Split(jsonTag, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == "omitempty" {
			return true
		}
	}
	return false
}
