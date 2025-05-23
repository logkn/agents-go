package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
)

// MockLLM is a simple implementation for testing
type MockLLM struct{}

func (m *MockLLM) GenerateResponse(ctx context.Context, messages []Message, tools []tools.Tool) (*LLMResponse, error) {
	// Simple mock - just echo the last message
	if len(messages) == 0 {
		return &LLMResponse{
			Content:  "Hello! How can I help you?",
			Finished: true,
		}, nil
	}

	lastMsg := messages[len(messages)-1]

	// Debug: Print message conversation for debugging (uncomment for debugging)
	// fmt.Printf("Mock Provider - Messages (%d):\n", len(messages))
	// for i, msg := range messages {
	// 	fmt.Printf("  [%d] Role: %s, Content: %s\n", i, msg.Role, msg.Content)
	// }

	// Check if we have tool results in the conversation
	hasToolResults := false
	for _, msg := range messages {
		if msg.Role == "tool" {
			hasToolResults = true
			break
		}
	}

	// If user mentions weather and email together and we haven't used tools yet
	content := strings.ToLower(lastMsg.Content)
	if strings.Contains(content, "weather") && strings.Contains(content, "email") && !hasToolResults {
		// Missing required parameters - ask for them
		return &LLMResponse{
			Content:  "Please provide me with the city you'd like to check the weather for and the email address where you want the weather information sent.",
			Finished: true,
		}, nil
	}

	// If user mentions weather and we haven't used tools yet, use the weather tool
	if strings.Contains(content, "weather") && !hasToolResults {
		// Extract city from the message or use a default
		city := "New York" // Default for "What's the weather in New York?"
		if strings.Contains(content, "new york") {
			city = "New York"
		} else if strings.Contains(content, "san francisco") {
			city = "San Francisco"
		}

		for _, tool := range tools {
			if tool.Name() == "GetWeather" {
				return &LLMResponse{
					Content: "I'll check the weather for you.",
					ToolCalls: []response.ToolCall{
						{
							ID:   "call_1",
							Name: "GetWeather",
							Parameters: map[string]any{
								"city": city,
							},
						},
					},
					Finished: false,
				}, nil
			}
		}
	}

	// If this is a structured output request, return mock JSON
	if strings.Contains(content, "san francisco") {
		return &LLMResponse{
			Content:  `{"location": "San Francisco, CA", "temperature": 68.5, "condition": "partly cloudy", "humidity": 65, "wind_speed": 12.3}`,
			Finished: true,
		}, nil
	}

	// If we have tool results, provide a final summary
	if hasToolResults {
		// Check if any system message mentions structured output
		isStructuredOutput := false
		for _, msg := range messages {
			if msg.Role == "system" && strings.Contains(strings.ToLower(msg.Content), "json") {
				isStructuredOutput = true
				break
			}
		}

		if isStructuredOutput {
			// Return structured JSON based on tool results
			return &LLMResponse{
				Content:  `{"location": "San Francisco, CA", "temperature": 68.5, "condition": "partly cloudy", "humidity": 65, "wind_speed": 12.3}`,
				Finished: true,
			}, nil
		}

		return &LLMResponse{
			Content:  "Based on the tool results, I've completed your request.",
			Finished: true,
		}, nil
	}

	return &LLMResponse{
		Content:  fmt.Sprintf("I received: %s", lastMsg.Content),
		Finished: true,
	}, nil
}

func (m *MockLLM) SupportsStreaming() bool {
	return false
}
