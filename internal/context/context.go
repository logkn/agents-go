// Package context
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
	ctx       any
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

// ContextError represents an error that occurred during context operations.
type ContextError struct {
	Op       string // Operation that failed
	Expected string // Expected type
	Got      string // Actual type
	Err      error  // Underlying error
}

func (e *ContextError) Error() string {
	if e.Expected != "" && e.Got != "" {
		return fmt.Sprintf("context error in %s: type mismatch (expected %s, got %s)", e.Op, e.Expected, e.Got)
	}
	if e.Err != nil {
		return fmt.Sprintf("context error in %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("context error in %s", e.Op)
}

func (e *ContextError) Unwrap() error {
	return e.Err
}

// FromAnyContext attempts to convert an AnyContext back to a typed Context[T].
// Returns an error if the types don't match.
func FromAnyContext[T any](anyCtx AnyContext) (Context[T], error) {
	if anyCtx == nil {
		return nil, &ContextError{
			Op:  "FromAnyContext",
			Err: fmt.Errorf("anyContext is nil"),
		}
	}

	expectedType := reflect.TypeOf((*T)(nil)).Elem().String()
	if anyCtx.TypeName() != expectedType {
		return nil, &ContextError{
			Op:       "FromAnyContext",
			Expected: expectedType,
			Got:      anyCtx.TypeName(),
		}
	}

	wrapper, ok := anyCtx.(*contextWrapper)
	if !ok {
		return nil, &ContextError{
			Op:  "FromAnyContext",
			Err: fmt.Errorf("invalid context wrapper type: got %T", anyCtx),
		}
	}

	ctx, ok := wrapper.ctx.(Context[T])
	if !ok {
		return nil, &ContextError{
			Op:  "FromAnyContext",
			Err: fmt.Errorf("failed to cast context to expected type %s", expectedType),
		}
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
