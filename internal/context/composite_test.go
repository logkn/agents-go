package context

import (
	"fmt"
	"testing"
)

func TestCompositeContext(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cc := NewCompositeContext()
		
		// Add different types of contexts
		userCtx := NewContext(SimpleContext{Value: "user"})
		AddTyped(cc, userCtx)
		
		configCtx := NewContext(ComplexContext{
			ID:   1,
			Name: "config",
			Settings: map[string]any{
				"debug": true,
			},
		})
		AddTyped(cc, configCtx)
		
		// Verify count
		if count := cc.Count(); count != 2 {
			t.Errorf("Expected 2 contexts, got %d", count)
		}
		
		// Retrieve contexts
		retrievedUser, err := Get[SimpleContext](cc)
		if err != nil {
			t.Fatalf("Failed to get SimpleContext: %v", err)
		}
		if retrievedUser.Value().Value != "user" {
			t.Errorf("Expected 'user', got %s", retrievedUser.Value().Value)
		}
		
		retrievedConfig, err := Get[ComplexContext](cc)
		if err != nil {
			t.Fatalf("Failed to get ComplexContext: %v", err)
		}
		if retrievedConfig.Value().Name != "config" {
			t.Errorf("Expected 'config', got %s", retrievedConfig.Value().Name)
		}
		
		// Check existence
		if !Has[SimpleContext](cc) {
			t.Error("Expected SimpleContext to exist")
		}
		if !Has[ComplexContext](cc) {
			t.Error("Expected ComplexContext to exist")
		}
		if Has[PointerContext](cc) {
			t.Error("Expected PointerContext to not exist")
		}
		
		// Remove a context
		Remove[SimpleContext](cc)
		if Has[SimpleContext](cc) {
			t.Error("SimpleContext should have been removed")
		}
		if count := cc.Count(); count != 1 {
			t.Errorf("Expected 1 context after removal, got %d", count)
		}
	})
	
	t.Run("get non-existent context", func(t *testing.T) {
		cc := NewCompositeContext()
		
		_, err := Get[SimpleContext](cc)
		if err == nil {
			t.Fatal("Expected error when getting non-existent context")
		}
		
		contextErr, ok := err.(*ContextError)
		if !ok {
			t.Fatalf("Expected ContextError, got %T", err)
		}
		
		if contextErr.Op != "CompositeContext.Get" {
			t.Errorf("Expected Op = CompositeContext.Get, got %s", contextErr.Op)
		}
	})
	
	t.Run("types listing", func(t *testing.T) {
		cc := NewCompositeContext()
		
		AddTyped(cc, NewContext("string"))
		AddTyped(cc, NewContext(123))
		AddTyped(cc, NewContext(SimpleContext{Value: "test"}))
		
		types := cc.Types()
		if len(types) != 3 {
			t.Errorf("Expected 3 types, got %d", len(types))
		}
		
		// Verify type names are present
		typeMap := make(map[string]bool)
		for _, typ := range types {
			typeMap[typ] = true
		}
		
		expectedTypes := []string{"string", "int", "context.SimpleContext"}
		for _, expected := range expectedTypes {
			if !typeMap[expected] {
				t.Errorf("Expected type %s not found in types list", expected)
			}
		}
	})
}

func TestContextChain(t *testing.T) {
	t.Run("basic chain operations", func(t *testing.T) {
		// Create contexts
		userCtx := NewContext(SimpleContext{Value: "user"})
		configCtx := NewContext(ComplexContext{ID: 1, Name: "config"})
		
		// Create chain
		chain := NewContextChain(
			ToAnyContext(userCtx),
			ToAnyContext(configCtx),
		)
		
		// Find contexts
		foundUser, err := Find[SimpleContext](chain)
		if err != nil {
			t.Fatalf("Failed to find SimpleContext: %v", err)
		}
		if foundUser.Value().Value != "user" {
			t.Errorf("Expected 'user', got %s", foundUser.Value().Value)
		}
		
		foundConfig, err := Find[ComplexContext](chain)
		if err != nil {
			t.Fatalf("Failed to find ComplexContext: %v", err)
		}
		if foundConfig.Value().ID != 1 {
			t.Errorf("Expected ID 1, got %d", foundConfig.Value().ID)
		}
		
		// Try to find non-existent type
		_, err = Find[PointerContext](chain)
		if err == nil {
			t.Error("Expected error when finding non-existent type")
		}
	})
	
	t.Run("chain with composite context", func(t *testing.T) {
		// Create a composite context
		composite := NewCompositeContext()
		AddTyped(composite, NewContext(SimpleContext{Value: "composite"}))
		AddTyped(composite, NewContext(123))
		
		// Create another standalone context
		standaloneCtx := NewContext(ComplexContext{ID: 2, Name: "standalone"})
		
		// Create chain with composite first
		chain := NewContextChain(
			ToAnyCompositeContext(composite),
			ToAnyContext(standaloneCtx),
		)
		
		// Find from composite
		foundSimple, err := Find[SimpleContext](chain)
		if err != nil {
			t.Fatalf("Failed to find SimpleContext: %v", err)
		}
		if foundSimple.Value().Value != "composite" {
			t.Errorf("Expected 'composite', got %s", foundSimple.Value().Value)
		}
		
		foundInt, err := Find[int](chain)
		if err != nil {
			t.Fatalf("Failed to find int: %v", err)
		}
		if foundInt.Value() != 123 {
			t.Errorf("Expected 123, got %d", foundInt.Value())
		}
		
		// Find from standalone
		foundComplex, err := Find[ComplexContext](chain)
		if err != nil {
			t.Fatalf("Failed to find ComplexContext: %v", err)
		}
		if foundComplex.Value().Name != "standalone" {
			t.Errorf("Expected 'standalone', got %s", foundComplex.Value().Name)
		}
	})
	
	t.Run("prepend and append", func(t *testing.T) {
		chain := NewContextChain()
		
		// Append contexts
		chain.Append(ToAnyContext(NewContext(SimpleContext{Value: "first"})))
		chain.Append(ToAnyContext(NewContext(SimpleContext{Value: "second"})))
		
		// Should find the first one (higher priority)
		found, err := Find[SimpleContext](chain)
		if err != nil {
			t.Fatalf("Failed to find SimpleContext: %v", err)
		}
		if found.Value().Value != "first" {
			t.Errorf("Expected 'first', got %s", found.Value().Value)
		}
		
		// Prepend a new one
		chain.Prepend(ToAnyContext(NewContext(SimpleContext{Value: "prepended"})))
		
		// Should now find the prepended one
		found, err = Find[SimpleContext](chain)
		if err != nil {
			t.Fatalf("Failed to find SimpleContext: %v", err)
		}
		if found.Value().Value != "prepended" {
			t.Errorf("Expected 'prepended', got %s", found.Value().Value)
		}
	})
}

func TestCompositeContext_ThreadSafety(t *testing.T) {
	cc := NewCompositeContext()
	
	// Run concurrent operations
	done := make(chan bool)
	
	// Writers
	go func() {
		for i := 0; i < 100; i++ {
			AddTyped(cc, NewContext(i))
			AddTyped(cc, NewContext(fmt.Sprintf("string%d", i)))
		}
		done <- true
	}()
	
	// Readers
	go func() {
		for i := 0; i < 100; i++ {
			Has[int](cc)
			Has[string](cc)
			cc.Count()
			cc.Types()
		}
		done <- true
	}()
	
	// Mixed operations
	go func() {
		for i := 0; i < 50; i++ {
			AddTyped(cc, NewContext(SimpleContext{Value: fmt.Sprintf("ctx%d", i)}))
			if Has[SimpleContext](cc) {
				Get[SimpleContext](cc)
			}
			Remove[SimpleContext](cc)
		}
		done <- true
	}()
	
	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
