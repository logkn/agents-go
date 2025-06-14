package main

import (
	"github.com/logkn/agents-go/internal/cli"
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	agents "github.com/logkn/agents-go/pkg"
)

// main is the program entry point.

var agent = agents.Agent{
	Name:         "Main Agent",
	Instructions: "You are a helpful assistant. Use the tools provided to answer questions.",
	Tools: []tools.Tool{
		tools.SearchTool,
		tools.PwdTool,
	},
	Model: types.ModelConfig{
		Model:       "qwen3:30b-a3b",
		BaseUrl:     "http://localhost:11434/v1",
		Temperature: 0.6,
	},
}

func main() {
	cli.RunTUI(agent)
}
