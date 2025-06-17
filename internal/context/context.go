package context

import (
	"fmt"
	"reflect"
)

// Context represents the execution context available to agents and tools.
// It provides type-safe access to shared data throughout the agent execution lifecycle.
type Context[T any] interface {
	// Value returns the context data
	Value() T
	// Type returns the reflect.Type of the context data
	Type() reflect.Type
}

// ContextWrapper implements Context[T] and holds the actual context data.
type ContextWrapper[T any] struct {
	data T
}

// NewContext creates a new context wrapper with the provided data.
func NewContext[T any](data T) Context[T] {
	return &ContextWrapper[T]{data: data}
}

// Value returns the wrapped context data.
func (c *ContextWrapper[T]) Value() T {
	return c.data
}

// Type returns the reflect.Type of the context data.
func (c *ContextWrapper[T]) Type() reflect.Type {
	return reflect.TypeOf(c.data)
}

// AnyContext is a type-erased context interface for internal use.
// This allows the framework to handle contexts of different types.
type AnyContext interface {
	// TypeName returns the string representation of the context type
	TypeName() string
	// IsNil returns true if the context is nil or contains nil data
	IsNil() bool
}

// contextWrapper implements AnyContext for any Context[T].
type contextWrapper struct {
	ctx       interface{}
	typeName  string
	isNilFunc func() bool
}

// ToAnyContext converts a typed Context[T] to AnyContext for internal framework use.
func ToAnyContext[T any](ctx Context[T]) AnyContext {
	return &contextWrapper{
		ctx:      ctx,
		typeName: reflect.TypeOf((*T)(nil)).Elem().String(),
		isNilFunc: func() bool {
			if ctx == nil {
				return true
			}
			val := ctx.Value()
			v := reflect.ValueOf(val)
			return !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil())
		},
	}
}

// TypeName returns the string representation of the context type.
func (c *contextWrapper) TypeName() string {
	return c.typeName
}

// IsNil returns true if the context is nil or contains nil data.
func (c *contextWrapper) IsNil() bool {
	return c.isNilFunc()
}

// FromAnyContext attempts to convert an AnyContext back to a typed Context[T].
// Returns an error if the types don't match.
func FromAnyContext[T any](anyCtx AnyContext) (Context[T], error) {
	if anyCtx == nil {
		return nil, fmt.Errorf("anyContext is nil")
	}
	
	expectedType := reflect.TypeOf((*T)(nil)).Elem().String()
	if anyCtx.TypeName() != expectedType {
		return nil, fmt.Errorf("context type mismatch: expected %s, got %s", expectedType, anyCtx.TypeName())
	}
	
	wrapper, ok := anyCtx.(*contextWrapper)
	if !ok {
		return nil, fmt.Errorf("invalid context wrapper type")
	}
	
	ctx, ok := wrapper.ctx.(Context[T])
	if !ok {
		return nil, fmt.Errorf("failed to cast context to expected type")
	}
	
	return ctx, nil
}

// ContextFactory is a function that creates a new context instance.
// Used by agents to initialize context for each run.
type ContextFactory[T any] func() T

// NoContext represents an empty context when no context is needed.
type NoContext struct{}

// EmptyContext creates a context with no data.
func EmptyContext() Context[NoContext] {
	return NewContext(NoContext{})
}