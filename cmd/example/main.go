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
	agentExecutor := executor.NewAgentExecutor()
	agentExecutor.RegisterAgent("assistant", agent)

	// Execute agent
	ctx := context.Background()
	responseChan := make(chan response.AgentResponse, 10)

	// Test parallel tool execution
	fmt.Println("=== Testing Parallel Tool Execution ===")
	go func() {
		err := agentExecutor.Execute(ctx, "assistant", "Check the weather and send me an email about it", responseChan)
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
		err := agentExecutor.Execute(ctx, "assistant", "What's the weather in New York?", responseChan)
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

	// Test structured output
	fmt.Println("\n=== Testing Structured Output ===")

	// Define a structured output type
	type WeatherReport struct {
		Location    string  `json:"location" description:"The location for the weather report"`
		Temperature float64 `json:"temperature" description:"Temperature in the specified units"`
		Condition   string  `json:"condition" description:"Weather condition (e.g., sunny, cloudy, rainy)"`
		Humidity    int     `json:"humidity,omitempty" description:"Humidity percentage"`
		WindSpeed   float64 `json:"wind_speed,omitempty" description:"Wind speed"`
	}

	// Create structured output schema
	structuredSchema := response.NewStructuredOutputSchema[WeatherReport]()

	// Print the generated schema
	schemaJSON, _ := json.MarshalIndent(structuredSchema.JSONSchema(), "", "  ")
	fmt.Printf("Generated JSON Schema for WeatherReport:\n%s\n", schemaJSON)

	// Create agent with structured output
	structuredAgent := &executor.Agent{
		Name:             "WeatherAgent",
		Instructions:     "You are a weather agent that returns weather data in structured format. Always respond with valid JSON matching the WeatherReport schema.",
		Tools:            []tools.Tool{weatherTool},
		Provider:         &provider.MockLLMProvider{},
		State:            globalState,
		StructuredOutput: structuredSchema,
	}

	// Register and execute structured agent
	agentExecutor.RegisterAgent("weather-agent", structuredAgent)
	responseChan = make(chan response.AgentResponse, 10)

	go func() {
		err := agentExecutor.Execute(ctx, "weather-agent", "Get weather for San Francisco", responseChan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Handle structured responses
	for response := range responseChan {
		fmt.Printf("[%s] %s\n", response.Type, response.Content)
		if response.StructuredData != nil {
			structuredJSON, _ := json.MarshalIndent(response.StructuredData, "", "  ")
			fmt.Printf("  Structured Data: %s\n", structuredJSON)
		}
		if response.Metadata != nil {
			if err, exists := response.Metadata["structured_output_error"]; exists {
				fmt.Printf("  Structured Output Error: %v\n", err)
			}
		}
	}
}
