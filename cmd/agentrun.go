package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/utils"
	agents "github.com/logkn/agents-go/pkg"
	"github.com/logkn/agents-go/tools"
)

// SearchWeb represents arguments for the search tool example.
type SearchWeb struct {
	// The query to search for
	Query string
}

// Run performs the mock web search and returns a hard coded result.
func (s SearchWeb) Run() any {
	// Simulate a search operation
	// In a real implementation, this would perform an actual search
	return "There are two classes in Daggerheart: the Warrior and the Mage."
}

var SearchTool = tools.Tool{
	Args: SearchWeb{},
}

var agent = agents.Agent{
	Name:         "Main Agent",
	Instructions: "You are a helpful assistant. Use the tools provided to answer questions.",
	Tools:        []tools.Tool{SearchTool},
	Model:        agents.ModelConfig{Model: "hf.co/unsloth/DeepSeek-R1-0528-Qwen3-8B-GGUF:Q6_K_XL", BaseUrl: "http://127.0.0.1:11434/v1"},
}

// RunAgent demonstrates running a simple agent with one tool.
func RunAgent() {
	// Configure logging level from environment variable or default to INFO
	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}
	
	// Create a structured logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	
	// Set the agent's logger
	agent.Logger = logger
	
	input := "What are the classes in Daggerheart?"
	agentResponse, err := runner.Run(agent, runner.Input{OfString: input})
	if err != nil {
		fmt.Println("Error running agent:", err)
		return
	}

	for event := range agentResponse.Stream() {
		if msg, ok := event.Message(); ok {
			fmt.Println("Message:", utils.JsonDumpsObj(msg))
		}
	}
}
