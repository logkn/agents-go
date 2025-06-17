# Agent Context System

This document describes the context system implementation inspired by the OpenAI agents Python SDK.

## Overview

The context system provides type-safe context sharing across agents, tools, and lifecycle hooks. Context data is available locally during execution but is NOT sent to the LLM, maintaining a clear separation between local execution state and conversation history.

## Key Components

### Context Types

- **`Context[T]`**: Generic interface for type-safe context
- **`AnyContext`**: Type-erased context for internal framework use
- **`ContextFactory[T]`**: Function type for creating context instances

### Context Creation

```go
// Create a typed context
userCtx := agents.NewContext(UserData{
    UserID: "123",
    Name: "Alice",
})

// Create an empty context
emptyCtx := agents.EmptyContext()
```

### Agent Creation with Context

```go
// Agent without context
agent := agents.NewAgent(agents.AgentConfig{
    Name: "MyAgent",
    Instructions: "You are a helpful assistant",
    Model: agents.ModelConfig{Model: "gpt-4o-mini"},
})

// Agent with context
agentWithCtx := agents.NewAgentWithContext(config, userCtx)
```

### Contextual Tools

Tools can access context during execution:

```go
type MyTool struct {
    Message string `json:"message" description:"The message to process"`
}

// Basic tool (fallback)
func (t MyTool) Run() any {
    return fmt.Sprintf("Hello! %s", t.Message)
}

// Contextual tool
func (t MyTool) RunWithAnyContext(ctx agents.AnyContext) any {
    // Try to get typed context
    userCtx, err := agents.FromAnyContext[UserData](ctx)
    if err != nil {
        return t.Run() // Fallback
    }
    
    user := userCtx.Value()
    return fmt.Sprintf("Hello %s! %s", user.Name, t.Message)
}

// Create contextual tool
tool := agents.NewContextualTool(
    "my_tool",
    "Description of my tool",
    &MyTool{},
    userCtx,
)
```

### Lifecycle Hooks

Context-aware lifecycle hooks for monitoring execution:

```go
hooks := &agents.LifecycleHooks{
    BeforeRun: func(ctx agents.AnyContext) error {
        // Called before agent starts
        return nil
    },
    AfterRun: func(ctx agents.AnyContext, result any) error {
        // Called after agent completes
        return nil
    },
    BeforeToolCall: func(ctx agents.AnyContext, toolName string, args string) error {
        // Called before each tool execution
        return nil
    },
    AfterToolCall: func(ctx agents.AnyContext, toolName string, result any) error {
        // Called after each tool execution
        return nil
    },
}

agent = agents.WithHooks(agent, hooks)
```

## Usage Patterns

### 1. User Session Context

```go
type SessionContext struct {
    UserID    string
    SessionID string
    Preferences map[string]string
}

sessionCtx := agents.NewContext(SessionContext{
    UserID: "user123",
    SessionID: "sess456",
    Preferences: map[string]string{"theme": "dark"},
})

agent := agents.NewAgentWithContext(config, sessionCtx)
```

### 2. Database Connection Context

```go
type DBContext struct {
    Connection *sql.DB
    UserID     string
}

dbCtx := agents.NewContext(DBContext{
    Connection: db,
    UserID: "user123",
})

// Tools can access the database through context
type QueryTool struct {
    Query string `json:"query"`
}

func (q QueryTool) RunWithAnyContext(ctx agents.AnyContext) any {
    dbCtx, err := agents.FromAnyContext[DBContext](ctx)
    if err != nil {
        return "Database not available"
    }
    
    db := dbCtx.Value()
    // Use db.Connection to execute queries
    return "Query executed"
}
```

### 3. Configuration Context

```go
type ConfigContext struct {
    Environment string
    APIKeys     map[string]string
    Features    []string
}

configCtx := agents.NewContext(ConfigContext{
    Environment: "production",
    APIKeys: map[string]string{
        "api_key": "secret",
    },
    Features: []string{"feature1", "feature2"},
})
```

## API Reference

### Core Functions

- `NewContext[T](data T) Context[T]` - Create typed context
- `EmptyContext() Context[NoContext]` - Create empty context
- `FromAnyContext[T](ctx AnyContext) (Context[T], error)` - Convert from AnyContext
- `ToAnyContext[T](ctx Context[T]) AnyContext` - Convert to AnyContext

### Agent Functions

- `NewAgent(config AgentConfig) Agent` - Create agent without context
- `NewAgentWithContext[T](config AgentConfig, ctx Context[T]) Agent` - Create agent with context
- `WithTools(agent Agent, tools ...Tool) Agent` - Add tools to agent
- `WithHooks(agent Agent, hooks *LifecycleHooks) Agent` - Add lifecycle hooks

### Tool Functions

- `NewTool(name, description string, args ToolArgs) Tool` - Create basic tool
- `NewContextualTool[T](name, description string, args AnyContextualToolArgs, ctx Context[T]) Tool` - Create contextual tool

## Design Principles

1. **Type Safety**: Generic context ensures compile-time type checking
2. **Local Only**: Context never sent to LLM, stays in local execution
3. **Fallback Support**: Tools work with or without context
4. **Consistency**: Same context type across entire agent execution chain
5. **Flexibility**: Support for any context type (structs, primitives, interfaces)

## Examples

See `examples/context_demo.go` and `examples/simple_context_demo.go` for complete working examples.