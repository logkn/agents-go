package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Tool interface that all tools must implement
type Tool interface {
	Name() string
	Description() string
	JSONSchema() map[string]any
	Execute(ctx context.Context, state any, paramsJSON []byte) (any, error)
}

// ToolOption allows customizing tool registration
type ToolOption func(*toolConfig)

type toolConfig struct {
	name        string
	description string
}

// WithName sets a custom name for the tool
func WithName(name string) ToolOption {
	return func(c *toolConfig) {
		c.name = name
	}
}

// WithDescription sets a custom description for the tool
func WithDescription(desc string) ToolOption {
	return func(c *toolConfig) {
		c.description = desc
	}
}

// reflectedTool wraps a function to implement the Tool interface
type reflectedTool struct {
	fn          reflect.Value
	fnType      reflect.Type
	paramsType  reflect.Type
	stateType   reflect.Type
	name        string
	description string
	schema      map[string]any
}

// RegisterTool converts any properly-structured function into a Tool
func RegisterTool(fn any, opts ...ToolOption) Tool {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		panic("RegisterTool: argument must be a function")
	}

	// Validate function signature: func(context.Context, StateInterface, ParamsStruct) (ResultType, error)
	if fnType.NumIn() != 3 {
		panic("RegisterTool: function must have exactly 3 parameters (ctx, state, params)")
	}

	if fnType.NumOut() != 2 {
		panic("RegisterTool: function must return (result, error)")
	}

	// Check parameter types
	ctxType := fnType.In(0)
	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic("RegisterTool: first parameter must be context.Context")
	}

	stateType := fnType.In(1)
	paramsType := fnType.In(2)

	// Check return types
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !fnType.Out(1).Implements(errorType) {
		panic("RegisterTool: second return value must be error")
	}

	// Extract configuration
	config := &toolConfig{
		name:        extractFunctionName(fnValue),
		description: fmt.Sprintf("Executes %s", extractFunctionName(fnValue)),
	}

	for _, opt := range opts {
		opt(config)
	}

	// Generate JSON schema from params struct
	schema := generateJSONSchema(paramsType)

	return &reflectedTool{
		fn:          fnValue,
		fnType:      fnType,
		paramsType:  paramsType,
		stateType:   stateType,
		name:        config.name,
		description: config.description,
		schema:      schema,
	}
}

func (t *reflectedTool) Name() string {
	return t.name
}

func (t *reflectedTool) Description() string {
	return t.description
}

func (t *reflectedTool) JSONSchema() map[string]any {
	return t.schema
}

func (t *reflectedTool) Execute(ctx context.Context, state any, paramsJSON []byte) (any, error) {
	// Create new instance of params struct
	paramsValue := reflect.New(t.paramsType).Elem()

	// Unmarshal JSON into params struct
	paramsPtr := paramsValue.Addr().Interface()
	if err := json.Unmarshal(paramsJSON, paramsPtr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	// Type assert state to required interface
	stateValue := reflect.ValueOf(state)
	if !stateValue.IsValid() || stateValue.IsNil() {
		return nil, fmt.Errorf("state is nil but tool requires %s", t.stateType)
	}
	if !stateValue.Type().AssignableTo(t.stateType) {
		return nil, fmt.Errorf("state does not implement required interface %s", t.stateType)
	}

	// Call the function
	results := t.fn.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		stateValue,
		paramsValue,
	})

	// Extract results
	result := results[0].Interface()
	errValue := results[1].Interface()

	if errValue != nil {
		return result, errValue.(error)
	}

	return result, nil
}

// Helper functions for reflection and schema generation

func extractFunctionName(fn reflect.Value) string {
	// Get the actual function name using runtime.FuncForPC
	pc := fn.Pointer()
	fullName := runtime.FuncForPC(pc).Name()

	// Extract just the function name from the full package path
	if idx := strings.LastIndex(fullName, "."); idx != -1 {
		return fullName[idx+1:]
	}
	return fullName
}

func generateJSONSchema(t reflect.Type) map[string]any {
	if t.Kind() != reflect.Struct {
		panic("generateJSONSchema: type must be a struct")
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
		isRequired := true

		// Parse json tag
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			// Check for omitempty
			for _, part := range parts[1:] {
				if part == "omitempty" {
					isRequired = false
				}
			}
		}

		// Generate property schema
		prop := generateFieldSchema(field.Type)

		// Add description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			prop["description"] = desc
		}

		properties[fieldName] = prop

		if isRequired {
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

func generateFieldSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": generateFieldSchema(t.Elem()),
		}
	case reflect.Struct:
		return generateJSONSchema(t)
	case reflect.Ptr:
		return generateFieldSchema(t.Elem())
	default:
		return map[string]any{"type": "string"} // fallback
	}
}
