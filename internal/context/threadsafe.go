package context

import (
	"reflect"
	"sync"
)

// ThreadSafeContext wraps a context with a read-write mutex for thread-safe access.
type ThreadSafeContext[T any] struct {
	mu   sync.RWMutex
	ctx  Context[T]
}

// NewThreadSafeContext creates a new thread-safe context wrapper.
func NewThreadSafeContext[T any](ctx Context[T]) *ThreadSafeContext[T] {
	return &ThreadSafeContext[T]{
		ctx: ctx,
	}
}

// Value returns the context data with read lock protection.
func (ts *ThreadSafeContext[T]) Value() T {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.ctx == nil {
		var zero T
		return zero
	}
	return ts.ctx.Value()
}

// Type returns the reflect.Type of the context data.
func (ts *ThreadSafeContext[T]) Type() reflect.Type {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.ctx == nil {
		return reflect.TypeOf((*T)(nil)).Elem()
	}
	return ts.ctx.Type()
}

// Update atomically updates the context.
func (ts *ThreadSafeContext[T]) Update(ctx Context[T]) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.ctx = ctx
}

// GetContext returns the underlying context with read lock protection.
func (ts *ThreadSafeContext[T]) GetContext() Context[T] {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.ctx
}

// ThreadSafeAnyContext provides thread-safe access to AnyContext.
type ThreadSafeAnyContext struct {
	mu  sync.RWMutex
	ctx AnyContext
}

// NewThreadSafeAnyContext creates a new thread-safe AnyContext wrapper.
func NewThreadSafeAnyContext(ctx AnyContext) *ThreadSafeAnyContext {
	return &ThreadSafeAnyContext{
		ctx: ctx,
	}
}

// TypeName returns the string representation of the context type.
func (ts *ThreadSafeAnyContext) TypeName() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.ctx == nil {
		return "nil"
	}
	return ts.ctx.TypeName()
}

// IsNil returns true if the context is nil or contains nil data.
func (ts *ThreadSafeAnyContext) IsNil() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.ctx == nil || ts.ctx.IsNil()
}

// Get returns the underlying context with read lock protection.
func (ts *ThreadSafeAnyContext) Get() AnyContext {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.ctx
}

// Update atomically updates the context.
func (ts *ThreadSafeAnyContext) Update(ctx AnyContext) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.ctx = ctx
}