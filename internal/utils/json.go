package utils

import (
	"reflect"

	"github.com/logkn/agents-go/internal/types"
)

func JSONSchema(obj types.Struct) map[string]any {
	objType := reflect.TypeOf(obj)

	// Handle pointer types
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	properties := make(map[string]any)
	required := []string{}

	// Iterate through struct fields
	for i := range objType.NumField() {
		field := objType.Field(i)

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
