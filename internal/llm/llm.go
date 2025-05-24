package llm

import (
	"github.com/logkn/agents-go/internal/tools"
	"github.com/logkn/agents-go/internal/types"
)

type LLMResponse struct {
	message types.Message
	delta   types.MessageDelta
	error   error
}

type LLM interface {
	Run(instructions string, messages []types.Message, tools []tools.Tool, responseFormat types.Struct) chan LLMResponse
}
