package utils

import (
	"reflect"
)

// GenerateSchema generates a JSON schema for the given struct type.
//
// Top-level fields are:
//   - type: "object"
//   - properties: a map of field names to their schemas
//   - required: a list of required field names
//
// The function handles nested structs, arrays, maps, and basic types.
func GenerateSchema(obj any) map[string]any {
	objType := reflect.TypeOf(obj)

	// Handle pointer types
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	return buildTypeSchema(objType)
}

func buildTypeSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer", "minimum": 0}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice, reflect.Array:
		schema := map[string]any{
			"type": "array",
		}
		// Add items schema for array elements
		elemSchema := buildTypeSchema(t.Elem())
		schema["items"] = elemSchema
		return schema
	case reflect.Map:
		schema := map[string]any{
			"type": "object",
		}
		// Add additionalProperties schema for map values
		if t.Elem().Kind() != reflect.Interface {
			valueSchema := buildTypeSchema(t.Elem())
			schema["additionalProperties"] = valueSchema
		} else {
			schema["additionalProperties"] = true
		}
		return schema
	case reflect.Struct:
		return buildStructSchema(t)
	case reflect.Ptr:
		// Handle pointer to types
		return buildTypeSchema(t.Elem())
	case reflect.Interface:
		// For interface{} types, allow any type
		return map[string]any{}
	default:
		// Fallback for unknown types
		return map[string]any{"type": "string"}
	}
}

func buildStructSchema(t reflect.Type) map[string]any {
	properties := make(map[string]any)
	required := []string{}

	// Iterate through struct fields
	for i := range t.NumField() {
		field := t.Field(i)

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
		isOptional := false
		if commaIdx := findCommaIndex(jsonTag); commaIdx != -1 {
			fieldName = jsonTag[:commaIdx]
			tagOptions := jsonTag[commaIdx+1:]
			isOptional = contains(tagOptions, "omitempty")
		}

		// Build field schema - recursively handle nested types
		fieldSchema := buildTypeSchema(field.Type)

		// Add description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema["description"] = desc
		}

		properties[fieldName] = fieldSchema

		// Add to required if not optional
		if !isOptional {
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

func findCommaIndex(s string) int {
	for i, r := range s {
		if r == ',' {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
