package main

import (
	"fmt"

	"github.com/logkn/agents-go/internal/runner"
	"github.com/logkn/agents-go/internal/types"
)

// RunSimpleChat provides a fallback chat interface
func RunSimpleChat() {
	fmt.Println("🤖 AI Agent Chat (Simple Mode)")
	fmt.Println("Enter 'exit' or 'quit' to stop")
	fmt.Println()

	var input string
	conversation := []types.Message{types.NewSystemMessage(agent.Instructions)}

	for {
		fmt.Print(userStyle.Render("You: "))
		fmt.Scanln(&input)

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
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Print(assistantStyle.Render("Assistant: "))
		for event := range resp.Stream() {
			if token, ok := event.Token(); ok {
				fmt.Print(token)
			}
		}
		fmt.Println()
		conversation = resp.FinalConversation()
	}
}
