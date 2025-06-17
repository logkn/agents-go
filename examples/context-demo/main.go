package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	agents "github.com/logkn/agents-go/pkg"
)

// UserContext demonstrates a custom context type for user-specific data
type UserContext struct {
	UserID    string
	UserName  string
	SessionID string
	Preferences map[string]string
}

// greetingTool demonstrates a tool that uses context to personalize responses
type greetingTool struct {
	Message string `json:"message" description:"The greeting message to personalize"`
}

// Run implements the basic ToolArgs interface for fallback
func (g greetingTool) Run() any {
	return fmt.Sprintf("Hello! %s", g.Message)
}

// RunWithAnyContext implements the contextual tool interface
func (g greetingTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return g.Run()
	}
	
	// Try to convert to our expected context type
	userCtx, err := agents.FromAnyContext[UserContext](ctx)
	if err != nil {
		return g.Run() // Fallback to non-contextual
	}
	
	user := userCtx.Value()
	return fmt.Sprintf("Hello %s (ID: %s)! %s", user.UserName, user.UserID, g.Message)
}

// userInfoTool demonstrates a tool that accesses user context data
type userInfoTool struct{}

func (u userInfoTool) Run() any {
	return "User information not available without context"
}

func (u userInfoTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return u.Run()
	}
	
	userCtx, err := agents.FromAnyContext[UserContext](ctx)
	if err != nil {
		return u.Run()
	}
	
	user := userCtx.Value()
	return fmt.Sprintf("User: %s (ID: %s), Session: %s, Preferences: %v", 
		user.UserName, user.UserID, user.SessionID, user.Preferences)
}

func main() {
	// Create a user context
	userContext := agents.NewContext(UserContext{
		UserID:    "user123",
		UserName:  "Alice",
		SessionID: "session456",
		Preferences: map[string]string{
			"theme": "dark",
			"lang":  "en",
		},
	})

	// Create lifecycle hooks to demonstrate context usage
	hooks := &agents.LifecycleHooks{
		BeforeRun: func(ctx agents.AnyContext) error {
			if ctx != nil {
				userCtx, err := agents.FromAnyContext[UserContext](ctx)
				if err == nil {
					user := userCtx.Value()
					fmt.Printf("🚀 Starting session for user: %s\n", user.UserName)
				}
			}
			return nil
		},
		AfterRun: func(ctx agents.AnyContext, result any) error {
			if ctx != nil {
				userCtx, err := agents.FromAnyContext[UserContext](ctx)
				if err == nil {
					user := userCtx.Value()
					fmt.Printf("✅ Completed session for user: %s\n", user.UserName)
				}
			}
			return nil
		},
		BeforeToolCall: func(ctx agents.AnyContext, toolName string, args string) error {
			fmt.Printf("🔧 About to call tool: %s\n", toolName)
			return nil
		},
		AfterToolCall: func(ctx agents.AnyContext, toolName string, result any) error {
			fmt.Printf("✅ Tool %s completed\n", toolName)
			return nil
		},
	}

	// Create tools with context support
	greetingToolInstance := &greetingTool{}
	userInfoToolInstance := &userInfoTool{}

	contextualGreeting := agents.NewContextualTool(
		"personalized_greeting",
		"Provides a personalized greeting using user context",
		greetingToolInstance,
		userContext,
	)

	contextualUserInfo := agents.NewContextualTool(
		"user_info",
		"Retrieves user information from context",
		userInfoToolInstance,
		userContext,
	)

	// Create an agent with context and tools
	config := agents.AgentConfig{
		Name:         "Context Demo Agent",
		Instructions: "You are a helpful assistant that can access user context. Use the available tools to provide personalized responses.",
		Model: agents.ModelConfig{
			Model:   "gpt-4o-mini",
			BaseUrl: "", // Use OpenAI API
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	agent := agents.NewAgentWithContext(config, userContext)
	agent = agents.WithTools(agent, contextualGreeting, contextualUserInfo)
	agent = agents.WithHooks(agent, hooks)

	// Demonstrate the context system
	fmt.Println("=== Context-Aware Agent Demo ===")
	
	// Run the agent with a request that will use context
	response, err := agents.Run(context.Background(), agent, agents.Input{
		OfString: "Hi! Can you greet me personally and tell me about my user information?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Stream and print events
	for event := range response.Stream() {
		if token, ok := event.Token(); ok && token != "" {
			fmt.Print(token)
		}
		if msg, ok := event.Message(); ok {
			if msg.Role == agents.Assistant {
				fmt.Printf("\n[Agent]: %s\n", msg.Content)
			}
		}
	}

	fmt.Printf("\n\nFinal response: %s\n", response.Response().Content)
	
	// Demonstrate without context (create a new agent without context)
	fmt.Println("\n=== Without Context Demo ===")
	
	agentNoContext := agents.NewAgent(config)
	// Add tools without context
	regularGreeting := agents.NewTool("greeting", "Provides a basic greeting", &greetingTool{})
	regularUserInfo := agents.NewTool("user_info", "Attempts to get user info", &userInfoTool{})
	agentNoContext = agents.WithTools(agentNoContext, regularGreeting, regularUserInfo)

	responseNoContext, err := agents.Run(context.Background(), agentNoContext, agents.Input{
		OfString: "Hi! Can you greet me and tell me about my user information?",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Stream and print events
	for event := range responseNoContext.Stream() {
		if token, ok := event.Token(); ok && token != "" {
			fmt.Print(token)
		}
		if msg, ok := event.Message(); ok {
			if msg.Role == agents.Assistant {
				fmt.Printf("\n[Agent]: %s\n", msg.Content)
			}
		}
	}

	fmt.Printf("\n\nFinal response (no context): %s\n", responseNoContext.Response().Content)
}