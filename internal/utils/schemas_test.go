package utils

import "testing"

type sampleStruct struct {
	Field string `json:"field"`
}

func TestCreateSchema(t *testing.T) {
	schema, err := CreateSchema(sampleStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("expected $defs in schema")
	}
	s, ok := defs["sampleStruct"].(map[string]any)
	if !ok {
		t.Fatalf("expected sampleStruct definition")
	}
	if s["type"] != "object" {
		t.Fatalf("expected object type")
	}
}
