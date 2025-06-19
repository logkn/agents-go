package types

// LifecycleHooks defines optional hooks that can be called during agent execution.
type LifecycleHooks[Context any] struct {
	BeforeRun      func(ctx *Context) error
	AfterRun       func(ctx *Context, result any) error
	BeforeToolCall func(ctx *Context, toolName string, args string) error
	AfterToolCall  func(ctx *Context, toolName string, result any) error
}
