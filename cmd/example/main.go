package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/logkn/agents-go/internal/executor"
	"github.com/logkn/agents-go/internal/provider"
	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/state"
	"github.com/logkn/agents-go/internal/tools"
)

// Example tool functions
func GetWeather(
	ctx context.Context,
	state state.WeatherStateReader,
	params struct {
		City  string `json:"city" description:"The city to get weather for"`
		Units string `json:"units,omitempty" description:"Temperature units (celsius/fahrenheit)"`
	},
) (string, error) {
	apiKey := state.GetWeatherAPIKey()
	units := params.Units
	if units == "" {
		units = state.GetDefaultUnits()
	}

	// Simulate API call
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("Weather in %s: 72°F, sunny (API key: %s, units: %s)",
		params.City, apiKey, units), nil
}

func SendEmail(
	ctx context.Context,
	state state.EmailStateReader,
	params struct {
		To      string `json:"to" description:"Recipient email address"`
		Subject string `json:"subject" description:"Email subject"`
		Body    string `json:"body" description:"Email body"`
	},
) (string, error) {
	config := state.GetEmailConfig()

	// Simulate sending email
	time.Sleep(200 * time.Millisecond)
	return fmt.Sprintf("Email sent to %s via %s:%d",
		params.To, config.SMTPHost, config.SMTPPort), nil
}

func main() {
	// Create global state
	globalState := &state.ExampleGlobalState{
		WeatherAPIKey: "weather-api-key-123",
		DefaultUnits:  "fahrenheit",
		EmailConfig: state.EmailConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password",
		},
	}

	// Register tools
	weatherTool := tools.RegisterTool(GetWeather, tools.WithDescription("Get current weather for any city"))
	emailTool := tools.RegisterTool(SendEmail, tools.WithDescription("Send an email"))

	// Create agent
	agent := &executor.Agent{
		Name:         "Assistant",
		Instructions: "You are a helpful assistant that can check weather and send emails.",
		Tools:        []tools.Tool{weatherTool, emailTool},
		Provider:     &provider.MockLLMProvider{},
		State:        globalState,
	}

	// Create executor and register agent
	executor := executor.NewAgentExecutor()
	executor.RegisterAgent("assistant", agent)

	// Execute agent
	ctx := context.Background()
	responseChan := make(chan response.AgentResponse, 10)

	// Test parallel tool execution
	fmt.Println("=== Testing Parallel Tool Execution ===")
	go func() {
		err := executor.Execute(ctx, "assistant", "Check the weather and send me an email about it", responseChan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Handle responses
	for response := range responseChan {
		fmt.Printf("[%s] %s\n", response.Type, response.Content)
		if response.ToolCall != nil {
			fmt.Printf("  Tool: %s(%v) -> %v\n",
				response.ToolCall.Name,
				response.ToolCall.Parameters,
				response.ToolCall.Result)
		}
	}

	// Test single tool execution
	fmt.Println("\n=== Testing Single Tool Execution ===")
	responseChan = make(chan response.AgentResponse, 10)
	go func() {
		err := executor.Execute(ctx, "assistant", "What's the weather in New York?", responseChan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Handle responses
	for response := range responseChan {
		fmt.Printf("[%s] %s\n", response.Type, response.Content)
		if response.ToolCall != nil {
			fmt.Printf("  Tool: %s(%v) -> %v\n",
				response.ToolCall.Name,
				response.ToolCall.Parameters,
				response.ToolCall.Result)
		}
	}

	// Demonstrate tool schema generation
	fmt.Println("\n=== Tool Schemas ===")
	for _, tool := range agent.Tools {
		schema, _ := json.MarshalIndent(tool.JSONSchema(), "", "  ")
		fmt.Printf("\n%s:\n%s\n", tool.Name(), schema)
	}
}