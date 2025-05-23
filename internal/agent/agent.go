package agent

import (
	"github.com/logkn/agents-go/internal/provider"
	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
)

// Agent represents an AI agent with tools and state
type Agent struct {
	Name             string
	Instructions     string
	Tools            []*tools.Tool
	Model            provider.LLM
	State            any
	StructuredOutput response.StructuredOutput
	Handoffs         []*Agent
}

func (agent *Agent) HandoffMap() map[string]*Agent {
	handoffs := make(map[string]*Agent)
	for _, handoff := range agent.Handoffs {
		handoffs[handoff.Name] = handoff
	}
	return handoffs
}
