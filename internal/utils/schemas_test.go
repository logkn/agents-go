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
	if schema["type"] != "object" {
		t.Fatalf("expected object type")
	}
}
