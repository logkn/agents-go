package agents

import "github.com/logkn/agents-go/internal/types"

type (
	Instructions[Context any] = types.AgentInstructions[Context]
)

func StringInstructions[Context any](s string) Instructions[Context] {
	return types.AgentInstructions[Context]{OfString: s}
}

func FileInstructions[Context any](file string) Instructions[Context] {
	return types.AgentInstructions[Context]{OfFile: file}
}
