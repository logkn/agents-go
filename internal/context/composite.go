package context

import (
	"fmt"
	"reflect"
	"sync"
)

// CompositeContext holds multiple contexts of different types and provides
// type-safe access to each. This enables context composition and inheritance patterns.
type CompositeContext struct {
	mu       sync.RWMutex
	contexts map[reflect.Type]AnyContext
}

// NewCompositeContext creates a new composite context.
func NewCompositeContext() *CompositeContext {
	return &CompositeContext{
		contexts: make(map[reflect.Type]AnyContext),
	}
}

// Add adds a typed context to the composite. If a context of the same type
// already exists, it will be replaced.
func (cc *CompositeContext) Add(ctx AnyContext) error {
	if ctx == nil {
		return &ContextError{
			Op:  "CompositeContext.Add",
			Err: fmt.Errorf("cannot add nil context"),
		}
	}
	
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	// Extract the actual type from the context
	typeName := ctx.TypeName()
	typ := reflect.TypeOf(typeName) // This is a placeholder - we need the actual type
	
	cc.contexts[typ] = ctx
	return nil
}

// AddTyped is a generic function to add a typed context to the composite.
func AddTyped[T any](cc *CompositeContext, ctx Context[T]) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	typ := reflect.TypeOf((*T)(nil)).Elem()
	anyCtx := ToAnyContext(ctx)
	cc.contexts[typ] = anyCtx
}

// Get is a generic function to retrieve a context of the specified type.
func Get[T any](cc *CompositeContext) (Context[T], error) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	typ := reflect.TypeOf((*T)(nil)).Elem()
	anyCtx, exists := cc.contexts[typ]
	if !exists {
		return nil, &ContextError{
			Op:       "CompositeContext.Get",
			Expected: typ.String(),
			Err:      fmt.Errorf("context not found"),
		}
	}
	
	return FromAnyContext[T](anyCtx)
}

// Has is a generic function to check if a context of the specified type exists.
func Has[T any](cc *CompositeContext) bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	typ := reflect.TypeOf((*T)(nil)).Elem()
	_, exists := cc.contexts[typ]
	return exists
}

// Remove is a generic function to remove a context of the specified type.
func Remove[T any](cc *CompositeContext) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	typ := reflect.TypeOf((*T)(nil)).Elem()
	delete(cc.contexts, typ)
}

// Count returns the number of contexts in the composite.
func (cc *CompositeContext) Count() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	return len(cc.contexts)
}

// Types returns a list of all context type names in the composite.
func (cc *CompositeContext) Types() []string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	types := make([]string, 0, len(cc.contexts))
	for _, ctx := range cc.contexts {
		types = append(types, ctx.TypeName())
	}
	return types
}

// compositeContextWrapper implements AnyContext for CompositeContext
type compositeContextWrapper struct {
	composite *CompositeContext
}

// ToAnyCompositeContext converts a CompositeContext to AnyContext.
func ToAnyCompositeContext(cc *CompositeContext) AnyContext {
	return &compositeContextWrapper{composite: cc}
}

func (c *compositeContextWrapper) TypeName() string {
	return "CompositeContext"
}

func (c *compositeContextWrapper) IsNil() bool {
	return c.composite == nil || c.composite.Count() == 0
}

// GetComposite extracts the CompositeContext from an AnyContext if it's a composite.
func GetComposite(ctx AnyContext) (*CompositeContext, bool) {
	if wrapper, ok := ctx.(*compositeContextWrapper); ok {
		return wrapper.composite, true
	}
	return nil, false
}

// ContextChain represents a chain of contexts with fallback behavior.
// When looking up a context type, it searches from first to last.
type ContextChain struct {
	mu       sync.RWMutex
	contexts []AnyContext
}

// NewContextChain creates a new context chain.
func NewContextChain(contexts ...AnyContext) *ContextChain {
	return &ContextChain{
		contexts: contexts,
	}
}

// Append adds a context to the end of the chain (lowest priority).
func (cc *ContextChain) Append(ctx AnyContext) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.contexts = append(cc.contexts, ctx)
}

// Prepend adds a context to the beginning of the chain (highest priority).
func (cc *ContextChain) Prepend(ctx AnyContext) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.contexts = append([]AnyContext{ctx}, cc.contexts...)
}

// Find is a generic function to search for a context of the specified type in the chain.
func Find[T any](cc *ContextChain) (Context[T], error) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	expectedType := reflect.TypeOf((*T)(nil)).Elem().String()
	
	for _, anyCtx := range cc.contexts {
		if anyCtx == nil {
			continue
		}
		
		// Check if it's a composite context
		if composite, ok := GetComposite(anyCtx); ok {
			if ctx, err := Get[T](composite); err == nil {
				return ctx, nil
			}
			continue
		}
		
		// Try direct conversion
		if anyCtx.TypeName() == expectedType {
			return FromAnyContext[T](anyCtx)
		}
	}
	
	return nil, &ContextError{
		Op:       "ContextChain.Find",
		Expected: expectedType,
		Err:      fmt.Errorf("context not found in chain"),
	}
}