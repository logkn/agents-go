package agents

import "github.com/logkn/agents-go/tools"

// ModelConfig contains configuration details for an LLM model.
// Model is the identifier of the model to use and BaseUrl is an optional
// override for the API base URL.
type ModelConfig struct {
	Model   string
	BaseUrl string
}

// Agent represents an autonomous entity that can process instructions and use
// tools. Tools are optional helpers, while Handoffs specifies other agents that
// can be delegated work.
type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Model        ModelConfig
	Handoffs     []*Agent
}

// func (a Agent) AsTool(toolname, description string) tools.Tool {
// }
