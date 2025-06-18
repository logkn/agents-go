package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
	agents "github.com/logkn/agents-go/pkg"
)

// SessionContext holds session-specific data
type SessionContext struct {
	SessionID string
	UserName  string
}

// sessionInfoTool demonstrates accessing context in a tool
type sessionInfoTool struct{}

func (s sessionInfoTool) Run() any {
	return "Session information not available"
}

func (s sessionInfoTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return s.Run()
	}

	sessionCtx, err := agents.FromAnyContext[SessionContext](ctx)
	if err != nil {
		return fmt.Sprintf("Context error: %v", err)
	}

	session := sessionCtx.Value()
	return fmt.Sprintf("Session ID: %s, User: %s", session.SessionID, session.UserName)
}

func main() {
	// Create session context
	sessionCtx := agents.NewContext(SessionContext{
		SessionID: "sess_123",
		UserName:  "Bob",
	})

	// Create contextual tool
	sessionTool := agents.NewContextualTool(
		"session_info",
		"Get current session information",
		&sessionInfoTool{},
		sessionCtx,
	)

	// Create agent
	agent := types.Agent{
		Name:         "Session Agent",
		Instructions: "You are an assistant that can access session information. Use the session_info tool when asked about the current session.",
		Model: types.ModelConfig{
			Model: "gpt-4o-mini",
		},
		Tools:  []agents.Tool{sessionTool},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	// Run agent with global context
	response, err := runner.Run(agent, runner.Input{
		OfString: "What's my current session information?",
	}, context.Background(), agents.ToAnyContext(sessionCtx))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Context Demo ===")
	for event := range response.Stream() {
		if token, ok := event.Token(); ok && token != "" {
			fmt.Print(token)
		}
	}

	fmt.Printf("\n\nFinal response: %s\n", response.Response().Content)
}
