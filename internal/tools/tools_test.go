package tools

import (
	"testing"

	"github.com/logkn/agents-go/internal/events"
)

// Example

type WebSearch struct {
	Query string `json:"query" description:"The query to search for"`
}

func (w WebSearch) Execute(state any, events events.EventBus) (any, error) {
	res := "Here are your search results for " + w.Query
	return res, nil
}

// Mock ToolDef implementations for testing
type MockTool struct {
	Name string `json:"name" description:"Name of the mock tool"`
	Age  int    `json:"age" description:"Age parameter"`
}

func (m MockTool) Execute(state any, events events.EventBus) (any, error) {
	return "mock result", nil
}

type MockToolWithOptional struct {
	Required string `json:"required" description:"Required field"`
	Optional string `json:"optional,omitempty" description:"Optional field"`
}

func (m MockToolWithOptional) Execute(state any, events events.EventBus) (any, error) {
	return "mock result with optional", nil
}

type MockToolWithPointer struct {
	Value string `json:"value" description:"A value"`
}

func (m *MockToolWithPointer) Execute(state any, events events.EventBus) (any, error) {
	return "pointer mock result", nil
}

func TestNewTool(t *testing.T) {
	mockDef := MockTool{Name: "test", Age: 25}
	tool := NewTool(mockDef)

	if tool.def == nil {
		t.Error("Expected def to be set")
	}
	if tool.name != "" {
		t.Error("Expected name to be empty initially")
	}
	if tool.description != "" {
		t.Error("Expected description to be empty initially")
	}
}

func TestWithName(t *testing.T) {
	mockDef := MockTool{}
	tool := NewTool(mockDef)

	updatedTool := (&tool).WithName("CustomName")

	if updatedTool.name != "CustomName" {
		t.Errorf("Expected name to be 'CustomName', got '%s'", updatedTool.name)
	}
}

func TestWithDescription(t *testing.T) {
	mockDef := MockTool{}
	tool := NewTool(mockDef)

	updatedTool := (&tool).WithDescription("Custom description")

	if updatedTool.description != "Custom description" {
		t.Errorf("Expected description to be 'Custom description', got '%s'", updatedTool.description)
	}
}

func TestDefaultName(t *testing.T) {
	mockDef := MockTool{}
	tool := NewTool(mockDef)

	defaultName := tool.defaultName()
	expected := "MockTool"

	if defaultName != expected {
		t.Errorf("Expected default name to be '%s', got '%s'", expected, defaultName)
	}
}

func TestDefaultNameWithPointer(t *testing.T) {
	mockDef := &MockToolWithPointer{}
	tool := NewTool(mockDef)

	defaultName := tool.defaultName()
	expected := "MockToolWithPointer"

	if defaultName != expected {
		t.Errorf("Expected default name to be '%s', got '%s'", expected, defaultName)
	}
}

func TestName(t *testing.T) {
	t.Run("uses custom name when set", func(t *testing.T) {
		mockDef := MockTool{}
		tool := NewTool(mockDef)
		updatedTool := (&tool).WithName("CustomName")

		name := updatedTool.Name()
		expected := "CustomName"

		if name != expected {
			t.Errorf("Expected name to be '%s', got '%s'", expected, name)
		}
	})

	t.Run("uses default name when not set", func(t *testing.T) {
		mockDef := MockTool{}
		tool := NewTool(mockDef)

		name := tool.Name()
		expected := "MockTool"

		if name != expected {
			t.Errorf("Expected name to be '%s', got '%s'", expected, name)
		}
	})
}

func TestDefaultDescription(t *testing.T) {
	mockDef := MockTool{}
	tool := NewTool(mockDef).WithName("TestTool")

	defaultDesc := tool.defaultDescription()
	expected := "Calls the TestTool tool."

	if defaultDesc != expected {
		t.Errorf("Expected default description to be '%s', got '%s'", expected, defaultDesc)
	}
}

func TestDescription(t *testing.T) {
	t.Run("uses custom description when set", func(t *testing.T) {
		mockDef := MockTool{}
		tool := NewTool(mockDef).WithDescription("Custom description")

		desc := tool.Description()
		expected := "Custom description"

		if desc != expected {
			t.Errorf("Expected description to be '%s', got '%s'", expected, desc)
		}
	})

	t.Run("uses default description when not set", func(t *testing.T) {
		mockDef := MockTool{}
		tool := NewTool(mockDef).WithName("TestTool")

		desc := tool.Description()
		expected := "Calls the TestTool tool."

		if desc != expected {
			t.Errorf("Expected description to be '%s', got '%s'", expected, desc)
		}
	})
}

func TestJSONSchema(t *testing.T) {
	t.Run("generates schema for basic types", func(t *testing.T) {
		mockDef := MockTool{}
		tool := NewTool(mockDef)

		schema := tool.JSONSchema()

		// Check top-level structure
		if schema["type"] != "object" {
			t.Error("Expected schema type to be 'object'")
		}

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("Expected properties to be a map")
		}

		// Check name field
		nameField, ok := properties["name"].(map[string]any)
		if !ok {
			t.Fatal("Expected name field to exist")
		}
		if nameField["type"] != "string" {
			t.Error("Expected name field type to be 'string'")
		}
		if nameField["description"] != "Name of the mock tool" {
			t.Error("Expected name field description to match tag")
		}

		// Check age field
		ageField, ok := properties["age"].(map[string]any)
		if !ok {
			t.Fatal("Expected age field to exist")
		}
		if ageField["type"] != "integer" {
			t.Error("Expected age field type to be 'integer'")
		}
		if ageField["description"] != "Age parameter" {
			t.Error("Expected age field description to match tag")
		}

		// Check required fields
		required, ok := schema["required"].([]string)
		if !ok {
			t.Fatal("Expected required to be a string slice")
		}

		expectedRequired := []string{"name", "age"}
		if len(required) != len(expectedRequired) {
			t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
		}

		for _, field := range expectedRequired {
			found := false
			for _, req := range required {
				if req == field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected field '%s' to be required", field)
			}
		}
	})

	t.Run("handles optional fields correctly", func(t *testing.T) {
		mockDef := MockToolWithOptional{}
		tool := NewTool(mockDef)

		schema := tool.JSONSchema()

		required, ok := schema["required"].([]string)
		if !ok {
			t.Fatal("Expected required to be a string slice")
		}

		// Should only have required field, not optional
		if len(required) != 1 || required[0] != "required" {
			t.Errorf("Expected only 'required' field to be required, got %v", required)
		}

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("Expected properties to be a map")
		}

		// Both fields should be in properties
		if len(properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(properties))
		}
	})

	t.Run("handles pointer types", func(t *testing.T) {
		mockDef := &MockToolWithPointer{}
		tool := NewTool(mockDef)

		schema := tool.JSONSchema()

		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("Expected properties to be a map")
		}

		valueField, ok := properties["value"].(map[string]any)
		if !ok {
			t.Fatal("Expected value field to exist")
		}
		if valueField["type"] != "string" {
			t.Error("Expected value field type to be 'string'")
		}
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "llo wo", true},
		{"hello world", "xyz", false},
		{"hello", "hello world", false},
		{"", "test", false},
		{"test", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_contains_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestWebSearchExample(t *testing.T) {
	webSearch := WebSearch{Query: "test query"}
	eventBus := events.NewEventBus()

	result, err := webSearch.Execute(nil, eventBus)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "Here are your search results for test query"
	if result != expected {
		t.Errorf("Expected result to be '%s', got '%s'", expected, result)
	}
}

func TestWebSearchTool(t *testing.T) {
	webSearch := WebSearch{}
	tool := NewTool(webSearch)

	// Test name
	name := tool.Name()
	if name != "WebSearch" {
		t.Errorf("Expected name to be 'WebSearch', got '%s'", name)
	}

	// Test schema
	schema := tool.JSONSchema()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	queryField, ok := properties["query"].(map[string]any)
	if !ok {
		t.Fatal("Expected query field to exist")
	}
	if queryField["type"] != "string" {
		t.Error("Expected query field type to be 'string'")
	}
	if queryField["description"] != "The query to search for" {
		t.Error("Expected query field description to match tag")
	}
}

func TestMethodChaining(t *testing.T) {
	mockDef := MockTool{}
	tool := NewTool(mockDef).
		WithName("ChainedTool").
		WithDescription("A tool created with method chaining")

	if tool.Name() != "ChainedTool" {
		t.Errorf("Expected name to be 'ChainedTool', got '%s'", tool.Name())
	}

	if tool.Description() != "A tool created with method chaining" {
		t.Errorf("Expected description to be 'A tool created with method chaining', got '%s'", tool.Description())
	}
}
