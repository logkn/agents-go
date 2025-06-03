package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func TypeName(v any) string {
	if v == nil {
		return "nil"
	}
	t := reflect.TypeOf(v)
	if t == nil {
		return "nil"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func NewInstance(v any) any {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface()
}

func AsString(v any) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case error:
		return val.Error()
	default:
		// Try to JSON marshal if possible
		if jsonBytes, err := json.Marshal(val); err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", val)
	}
}
