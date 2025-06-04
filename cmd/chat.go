package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

// RunChat starts an interactive session with the agent allowing multiple turns.
func RunChat() {
	conversation := []types.Message{types.NewSystemMessage(agent.Instructions)}
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\033[1;34mYou: \033[0m")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		conversation = append(conversation, types.NewUserMessage(input))
		resp, err := runner.Run(agent, runner.Input{OfMessages: conversation})
		if err != nil {
			fmt.Println("Error running agent:", err)
			return
		}

		fmt.Print("\033[1;32mAssistant: \033[0m")
		for event := range resp.Stream() {
			if token, ok := event.Token(); ok {
				fmt.Print(token)
			}
		}
		fmt.Println()
		conversation = resp.FinalConversation()
	}
}
