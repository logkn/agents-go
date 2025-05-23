package main

import (
	"context"
	"fmt"
	"time"

	"github.com/logkn/agents-go/internal/agent"
	"github.com/logkn/agents-go/internal/provider"
	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/state"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/utils"
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

var WeatherTool = tools.CreateTool(
	GetWeather,
	tools.WithName("GetWeather"),
	tools.WithDescription("Get current weather for a city"),
)

func main() {
	globalState := &state.ExampleGlobalState{
		WeatherAPIKey: "your-weather-api-key",
		DefaultUnits:  "fahrenheit",
		EmailConfig: state.EmailConfig{
			SMTPHost: "smtp.gmail.com",
			SMTPPort: 587,
			Username: "your-email@example.com",
			Password: "your-password",
		},
	}

	agent := agent.Agent{
		Name:         "ExampleAgent",
		Instructions: "You are a helpful assistant. You can get weather information and send emails.",
		Tools: []tools.Tool{
			WeatherTool,
			tools.ThinkTool,
		},
		Model: provider.NewOpenAIProvider("gpt-4o-mini"),
		State: globalState,
	}

	responseChan := make(chan response.AgentResponse)
	go runner.Run(&agent, "Your task is to find the weather in the hometown of Patrick Stewart. First think about where he is from.", context.Background(), responseChan)

	for resp := range responseChan {
		switch resp.Type {
		case response.ResponseTypeFinal:
			fmt.Println("Final Response:", resp.Content)
		case response.ResponseTypeToolCall:
			fmt.Println("Tool Call:", utils.JsonDumps(resp.ToolCall))
		}
	}
}
