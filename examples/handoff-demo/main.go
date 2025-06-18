package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

func main() {
	// Create specialized agents
	mathAgent := types.Agent{
		Name:         "Math Specialist",
		Instructions: types.AgentInstructions{OfString: "You are a math specialist. Help users with mathematical calculations and problems. If asked about non-math topics, transfer to the general assistant."},
		Model: types.ModelConfig{
			Model: "qwen3:30b-a3b",
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	generalAgent := types.Agent{
		Name:         "General Assistant",
		Instructions: types.AgentInstructions{OfString: "You are a general purpose assistant. Help users with various tasks. If asked complex math questions, transfer to the math specialist."},
		Model: types.ModelConfig{
			Model: "qwen3:30b-a3b",
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	// Set up handoffs - each agent can hand off to the other
	mathAgent.Handoffs = []types.Handoff{
		{
			Agent:           &generalAgent,
			ToolDescription: "Transfer to general assistant for non-math questions",
		},
	}

	generalAgent.Handoffs = []types.Handoff{
		{
			Agent:           &mathAgent,
			ToolDescription: "Transfer to math specialist for complex calculations",
		},
	}

	// Start with the general agent
	response, err := runner.Run(generalAgent, runner.Input{
		OfString: "Can you help me calculate the derivative of x^2 + 3x + 5?",
	}, context.Background(), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Stream and print events
	fmt.Println("=== Agent Handoff Demo ===")
	for event := range response.Stream() {
		if token, ok := event.Token(); ok && token != "" {
			fmt.Print(token)
		}
		if msg, ok := event.Message(); ok {
			if msg.Role == types.Assistant {
				fmt.Printf("\n[%s]: %s\n", "Agent", msg.Content)
			}
		}
		if handoff, ok := event.Handoff(); ok {
			fmt.Printf("\n🔄 HANDOFF: %s → %s\n", handoff.FromAgent, handoff.ToAgent)
			fmt.Printf("Prompt: %s\n", handoff.Prompt)
		}
	}

	fmt.Printf("\n\nFinal response: %s\n", response.Response().Content)
}
