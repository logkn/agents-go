package agents

import "github.com/logkn/agents-go/tools"

type ModelConfig struct {
	Model   string
	BaseUrl string
}

type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []*Agent
}

// func (a Agent) AsTool(toolname, description string) tools.Tool {
// }
