package main

import (
	"github.com/logkn/agents-go/cli"
	agents "github.com/logkn/agents-go/pkg"
	"github.com/logkn/agents-go/tools"
)

var agent = agents.BaseAgent(agents.NewModel("qwen3:30b-a3b", agents.WithBaseURL("http://localhost:11434/v1"))).WithBaseTools(tools.SearchTool)

func main() {
	cli.RunTUI(agent, agents.Null, cli.LogToFile("logs.txt"))
}
