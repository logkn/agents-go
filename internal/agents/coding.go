package agents

import (
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
	agents "github.com/logkn/agents-go/pkg"
)

const Instructions = `You are a coding assistant. Use the tools provided to answer questions.`

var CodingAgent = agents.Agent{
	Name:         "Coding Agent",
	Instructions: Instructions,
	Tools: []tools.Tool{
		tools.FileReadTool,
		tools.FileWriteTool,
		tools.ListTool,
		tools.SearchTool,
		tools.PwdTool,
		tools.PatchTool,
	},
	Model: types.ModelConfig{
		Model:       "qwen3:30b-a3b",
		BaseUrl:     "http://localhost:11434/v1",
		Temperature: 0.6,
	},
}
