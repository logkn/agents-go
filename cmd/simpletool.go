package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

func RunSimpleToolChat() {
	// user can get a chat back from an LLM
	message := "What classes are there in Daggerheart?"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(message),
	}
	client := openai.NewClient()
	for {
		chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    "gpt-4.1-nano",
			Tools: []openai.ChatCompletionToolParam{
				{
					Function: openai.FunctionDefinitionParam{
						Name:        "search_web",
						Description: openai.String("Search the web for information"),
						Parameters: openai.FunctionParameters{
							"type": "object",
							"properties": map[string]any{
								"query": map[string]string{
									"type":        "string",
									"description": "The search query to use",
								},
							},
							"required": []string{"query"},
						},
					},
				},
			},
		})
		if err != nil {
			panic(err.Error())
		}
		messages = append(messages, chatCompletion.Choices[0].Message.ToParam())
		if content := chatCompletion.Choices[0].Message.Content; content != "" {
			fmt.Println(content)
			break
		}
		// fmt.Println(chatCompletion.Choices[0].Message.ToolCalls)
		toolcalls := chatCompletion.Choices[0].Message.ToolCalls
		for _, toolcall := range toolcalls {
			funcname := toolcall.Function.Name
			fmt.Println(funcname)
			var args map[string]any
			err := json.Unmarshal([]byte(toolcall.Function.Arguments), &args)
			if err != nil {
				fmt.Println("Error unmarshalling function arguments:", err)
				continue
			}
			// query := args["query"].(string)

			response := "There are two classes in Daggerheart: the Warrior and the Mage."

			messages = append(messages, openai.ToolMessage(
				response,
				toolcall.ID,
			))
		}
	}
	fmt.Println("Done with chat")
}
