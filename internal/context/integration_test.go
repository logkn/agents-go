package context

import (
	"fmt"
	"testing"
)

// Mock tool for testing context integration
type TestTool struct {
	Message string `json:"message" description:"Test message"`
}

// Run implements basic ToolArgs interface (fallback)
func (t TestTool) Run() any {
	return "Basic execution: " + t.Message
}

// RunWithAnyContext implements contextual tool interface
func (t TestTool) RunWithAnyContext(ctx AnyContext) any {
	if ctx == nil {
		return t.Run()
	}

	// Try to get user context
	if userCtx, err := FromAnyContext[SimpleContext](ctx); err == nil {
		user := userCtx.Value()
		return "Contextual execution for " + user.Value + ": " + t.Message
	}

	// Try to get complex context
	if complexCtx, err := FromAnyContext[ComplexContext](ctx); err == nil {
		complex := complexCtx.Value()
		return "Contextual execution for " + complex.Name + ": " + t.Message
	}

	// Fallback if context type doesn't match
	return t.Run()
}

func TestContextualToolIntegration(t *testing.T) {
	t.Run("tool with matching context", func(t *testing.T) {
		ctx := NewContext(SimpleContext{Value: "testuser"})
		anyCtx := ToAnyContext(ctx)

		tool := TestTool{Message: "hello"}
		result := tool.RunWithAnyContext(anyCtx)

		expected := "Contextual execution for testuser: hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("tool with non-matching context", func(t *testing.T) {
		ctx := NewContext("wrong type")
		anyCtx := ToAnyContext(ctx)

		tool := TestTool{Message: "hello"}
		result := tool.RunWithAnyContext(anyCtx)

		expected := "Basic execution: hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("tool with nil context", func(t *testing.T) {
		tool := TestTool{Message: "hello"}
		result := tool.RunWithAnyContext(nil)

		expected := "Basic execution: hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("tool with composite context", func(t *testing.T) {
		composite := NewCompositeContext()
		AddTyped(composite, NewContext(SimpleContext{Value: "compositeuser"}))
		AddTyped(composite, NewContext(123))

		compositeAnyCtx := ToAnyCompositeContext(composite)

		tool := TestTool{Message: "composite hello"}

		// The tool should fallback since it doesn't know about composite contexts
		result := tool.RunWithAnyContext(compositeAnyCtx)
		expected := "Basic execution: composite hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// Advanced contextual tool that knows about composite contexts
type AdvancedTestTool struct {
	Message string `json:"message" description:"Test message"`
}

func (a AdvancedTestTool) Run() any {
	return "Advanced basic execution: " + a.Message
}

func (a AdvancedTestTool) RunWithAnyContext(ctx AnyContext) any {
	if ctx == nil {
		return a.Run()
	}

	// Check if it's a composite context
	if composite, ok := GetComposite(ctx); ok {
		// Try to get user context from composite
		if userCtx, err := Get[SimpleContext](composite); err == nil {
			user := userCtx.Value()
			return "Advanced composite execution for " + user.Value + ": " + a.Message
		}
	}

	// Try direct context conversion
	if userCtx, err := FromAnyContext[SimpleContext](ctx); err == nil {
		user := userCtx.Value()
		return "Advanced direct execution for " + user.Value + ": " + a.Message
	}

	return a.Run()
}

func TestAdvancedContextualTool(t *testing.T) {
	t.Run("advanced tool with composite context", func(t *testing.T) {
		composite := NewCompositeContext()
		AddTyped(composite, NewContext(SimpleContext{Value: "advanceduser"}))
		AddTyped(composite, NewContext(ComplexContext{ID: 42, Name: "config"}))

		compositeAnyCtx := ToAnyCompositeContext(composite)

		tool := AdvancedTestTool{Message: "advanced hello"}
		result := tool.RunWithAnyContext(compositeAnyCtx)

		expected := "Advanced composite execution for advanceduser: advanced hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("advanced tool with direct context", func(t *testing.T) {
		ctx := NewContext(SimpleContext{Value: "directuser"})
		anyCtx := ToAnyContext(ctx)

		tool := AdvancedTestTool{Message: "direct hello"}
		result := tool.RunWithAnyContext(anyCtx)

		expected := "Advanced direct execution for directuser: direct hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

func TestContextChainWithTools(t *testing.T) {
	// Create a context chain with different priorities
	userCtx := NewContext(SimpleContext{Value: "chainuser"})
	configCtx := NewContext(ComplexContext{ID: 1, Name: "chainconfig"})

	chain := NewContextChain(
		ToAnyContext(userCtx),   // Higher priority
		ToAnyContext(configCtx), // Lower priority
	)

	// Create a tool that tries to find user context from the chain
	type ChainTool struct {
		Message string `json:"message"`
	}

	_ = ChainTool{Message: "chain test"} // Tool definition for reference

	// Simulate tool execution that searches the chain
	foundUserCtx, err := Find[SimpleContext](chain)
	if err != nil {
		t.Fatalf("Failed to find user context in chain: %v", err)
	}

	user := foundUserCtx.Value()
	expected := "chainuser"
	if user.Value != expected {
		t.Errorf("Expected user value %s, got %s", expected, user.Value)
	}

	// Test finding config context
	foundConfigCtx, err := Find[ComplexContext](chain)
	if err != nil {
		t.Fatalf("Failed to find config context in chain: %v", err)
	}

	config := foundConfigCtx.Value()
	if config.Name != "chainconfig" {
		t.Errorf("Expected config name chainconfig, got %s", config.Name)
	}
}

func TestErrorHandlingInContextTools(t *testing.T) {
	// Test tool that properly handles context errors
	type ErrorHandlingTool struct {
		Message string `json:"message"`
	}

	_ = &ErrorHandlingTool{Message: "error test"} // Tool definition for reference

	// Create a function that simulates tool execution with error handling
	executeWithContext := func(ctx AnyContext) (any, error) {
		if ctx == nil {
			return "No context provided", nil
		}

		// Try to convert context with proper error handling
		userCtx, err := FromAnyContext[SimpleContext](ctx)
		if err != nil {
			// Check if it's a context error
			if contextErr, ok := err.(*ContextError); ok {
				return fmt.Sprintf("Context error: %s (Expected: %s, Got: %s)",
					contextErr.Op, contextErr.Expected, contextErr.Got), nil
			}
			return fmt.Sprintf("Unknown error: %v", err), nil
		}

		user := userCtx.Value()
		return fmt.Sprintf("Success for user: %s", user.Value), nil
	}

	t.Run("with correct context", func(t *testing.T) {
		ctx := NewContext(SimpleContext{Value: "erroruser"})
		anyCtx := ToAnyContext(ctx)

		result, err := executeWithContext(anyCtx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := "Success for user: erroruser"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("with wrong context type", func(t *testing.T) {
		ctx := NewContext("wrong type")
		anyCtx := ToAnyContext(ctx)

		result, err := executeWithContext(anyCtx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should contain context error information
		if !containsString(result.(string), "Context error") {
			t.Errorf("Expected context error message, got: %s", result)
		}

		if !containsString(result.(string), "context.SimpleContext") {
			t.Errorf("Expected expected type in error, got: %s", result)
		}

		if !containsString(result.(string), "string") {
			t.Errorf("Expected actual type in error, got: %s", result)
		}
	})

	t.Run("with nil context", func(t *testing.T) {
		result, err := executeWithContext(nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := "No context provided"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
