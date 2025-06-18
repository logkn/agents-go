package main

import (
	"github.com/logkn/agents-go/internal/agents"
	"github.com/logkn/agents-go/internal/cli"
)

func main() {
	cli.RunTUI(agents.CodingAgent, true, agents.NewCodingContext())
}
