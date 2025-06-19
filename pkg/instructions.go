package agents

import "github.com/logkn/agents-go/internal/types"

type (
	Instructions = types.AgentInstructions
)

func StringInstructions(s string) Instructions {
	return types.AgentInstructions{OfString: s}
}

func FileInstructions(file string) Instructions {
	return types.AgentInstructions{OfFile: file}
}
