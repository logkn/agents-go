package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	agents "github.com/logkn/agents-go/pkg"
)

// APIContext holds HTTP client and configuration for external API calls
type APIContext struct {
	Client  *http.Client
	BaseURL string
	APIKey  string
	UserID  string
}

// WeatherTool fetches weather data using the API context
type WeatherTool struct {
	City string `json:"city" description:"City name to get weather for"`
}

func (w WeatherTool) Run() any {
	return "Weather API not available - missing context"
}

func (w WeatherTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return w.Run()
	}

	apiCtx, err := agents.FromAnyContext[APIContext](ctx)
	if err != nil {
		return w.Run()
	}

	api := apiCtx.Value()

	// Simulate API call (in real scenario, you'd call a weather API)
	url := fmt.Sprintf("%s/weather?city=%s&key=%s", api.BaseURL, w.City, api.APIKey)

	// For demo purposes, return mock data
	mockResponse := map[string]any{
		"city":         w.City,
		"temperature":  "22°C",
		"condition":    "Sunny",
		"humidity":     "65%",
		"requested_by": api.UserID,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}

	// Log the API call
	fmt.Printf("🌤️ API Call: GET %s (User: %s)\n", url, api.UserID)

	return fmt.Sprintf("Weather in %s: %s, %s, Humidity: %s",
		w.City, mockResponse["temperature"], mockResponse["condition"], mockResponse["humidity"])
}

// NewsToolArgs fetches news headlines
type NewsTool struct {
	Topic string `json:"topic" description:"News topic to search for"`
	Limit int    `json:"limit" description:"Number of articles to fetch (max 10)"`
}

func (n NewsTool) Run() any {
	return "News API not available - missing context"
}

func (n NewsTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return n.Run()
	}

	apiCtx, err := agents.FromAnyContext[APIContext](ctx)
	if err != nil {
		return n.Run()
	}

	api := apiCtx.Value()

	// Limit validation
	if n.Limit <= 0 || n.Limit > 10 {
		n.Limit = 5
	}

	url := fmt.Sprintf("%s/news?topic=%s&limit=%d&key=%s", api.BaseURL, n.Topic, n.Limit, api.APIKey)

	// Mock news data
	mockArticles := []map[string]string{
		{
			"title":  fmt.Sprintf("Breaking: %s developments continue", n.Topic),
			"source": "Tech News Daily",
			"time":   "2 hours ago",
		},
		{
			"title":  fmt.Sprintf("Analysis: The future of %s industry", n.Topic),
			"source": "Industry Weekly",
			"time":   "4 hours ago",
		},
		{
			"title":  fmt.Sprintf("Opinion: Why %s matters now more than ever", n.Topic),
			"source": "Opinion Tribune",
			"time":   "6 hours ago",
		},
	}

	// Limit to requested number
	if len(mockArticles) > n.Limit {
		mockArticles = mockArticles[:n.Limit]
	}

	fmt.Printf("📰 API Call: GET %s (User: %s)\n", url, api.UserID)

	result := fmt.Sprintf("Found %d articles about '%s':\n", len(mockArticles), n.Topic)
	for i, article := range mockArticles {
		result += fmt.Sprintf("%d. %s (%s - %s)\n", i+1, article["title"], article["source"], article["time"])
	}

	return result
}

// UserPreferencesTool manages user preferences via API
type UserPreferencesTool struct {
	Action string `json:"action" description:"Action to perform: 'get' or 'set'"`
	Key    string `json:"key,omitempty" description:"Preference key (for set action)"`
	Value  string `json:"value,omitempty" description:"Preference value (for set action)"`
}

func (u UserPreferencesTool) Run() any {
	return "User preferences API not available - missing context"
}

func (u UserPreferencesTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return u.Run()
	}

	apiCtx, err := agents.FromAnyContext[APIContext](ctx)
	if err != nil {
		return u.Run()
	}

	api := apiCtx.Value()

	switch u.Action {
	case "get":
		// Mock user preferences
		preferences := map[string]string{
			"theme":       "dark",
			"language":    "en",
			"timezone":    "UTC",
			"news_topics": "technology,science",
		}

		url := fmt.Sprintf("%s/users/%s/preferences?key=%s", api.BaseURL, api.UserID, api.APIKey)
		fmt.Printf("👤 API Call: GET %s\n", url)

		result := "Current user preferences:\n"
		for key, value := range preferences {
			result += fmt.Sprintf("- %s: %s\n", key, value)
		}
		return result

	case "set":
		if u.Key == "" || u.Value == "" {
			return "Error: Both 'key' and 'value' are required for set action"
		}

		url := fmt.Sprintf("%s/users/%s/preferences?key=%s", api.BaseURL, api.UserID, api.APIKey)
		fmt.Printf("👤 API Call: PUT %s (Setting %s=%s)\n", url, u.Key, u.Value)

		return fmt.Sprintf("Successfully updated preference: %s = %s", u.Key, u.Value)

	default:
		return "Error: Action must be 'get' or 'set'"
	}
}

func main() {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create API context
	apiContext := agents.NewContext(APIContext{
		Client:  httpClient,
		BaseURL: "https://api.example.com/v1", // Mock API base URL
		APIKey:  "demo-api-key-12345",
		UserID:  "user789",
	})

	// Create lifecycle hooks for API monitoring
	hooks := &agents.LifecycleHooks{
		BeforeRun: func(ctx agents.AnyContext) error {
			if apiCtx, err := agents.FromAnyContext[APIContext](ctx); err == nil {
				api := apiCtx.Value()
				fmt.Printf("🚀 Starting API agent session for user: %s\n", api.UserID)
			}
			return nil
		},
		BeforeToolCall: func(ctx agents.AnyContext, toolName string, args string) error {
			fmt.Printf("🔧 Preparing API call for tool: %s\n", toolName)
			return nil
		},
		AfterToolCall: func(ctx agents.AnyContext, toolName string, result any) error {
			fmt.Printf("✅ API tool %s completed successfully\n", toolName)
			return nil
		},
		AfterRun: func(ctx agents.AnyContext, result any) error {
			fmt.Println("🏁 API agent session completed")
			return nil
		},
	}

	// Create contextual tools
	weatherTool := agents.NewContextualTool(
		"get_weather",
		"Get current weather information for a city",
		&WeatherTool{},
		apiContext,
	)

	newsTool := agents.NewContextualTool(
		"get_news",
		"Fetch news articles about a specific topic",
		&NewsTool{},
		apiContext,
	)

	prefsTool := agents.NewContextualTool(
		"user_preferences",
		"Get or set user preferences",
		&UserPreferencesTool{},
		apiContext,
	)

	// Create agent configuration
	config := agents.AgentConfig{
		Name: "API Assistant",
		Instructions: `You are an API-powered assistant that can:
1. Fetch weather information for any city
2. Get news articles on specific topics
3. Manage user preferences

You have access to external APIs through your context. Always be helpful and provide clear, formatted responses to users. When fetching data, explain what you're doing and provide useful summaries.`,
		Model: agents.ModelConfig{
			Model: "gpt-4o-mini",
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	// Create agent with context and tools
	agent := agents.NewAgentWithContext(config, apiContext)
	agent = agents.WithTools(agent, weatherTool, newsTool, prefsTool)
	agent = agents.WithHooks(agent, hooks)

	// Demo interactions
	fmt.Println("=== API Context Demo ===")
	fmt.Println("Agent has access to Weather, News, and User Preferences APIs")
	fmt.Println("Current user context: user789")
	fmt.Println()

	// Demo requests
	requests := []string{
		"What's the weather like in New York?",
		"Can you get me the latest news about artificial intelligence? Limit it to 3 articles.",
		"Show me my current user preferences.",
		"Can you get weather for London and news about climate change?",
	}

	for i, request := range requests {
		fmt.Printf("\n--- Request %d: %s ---\n", i+1, request)

		response, err := agents.Run(context.Background(), agent, agents.Input{
			OfString: request,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Stream the response
		for event := range response.Stream() {
			if token, ok := event.Token(); ok && token != "" {
				fmt.Print(token)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("This demo showed how contexts can provide:")
	fmt.Println("• API clients and configuration")
	fmt.Println("• User-specific data and preferences")
	fmt.Println("• Cross-tool data sharing")
	fmt.Println("• Request tracking and monitoring")
}
