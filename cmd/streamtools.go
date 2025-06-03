package main

import (
	"context"
	"fmt"

	"github.com/logkn/agents-go/internal/utils"
	"github.com/logkn/agents-go/tools"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func RunStreamedToolChat() {
	// User story: As a user, I want to chat with an LLM that has tools enabled,
	// and stream back the response as it is generated.
	tools := []tools.Tool{
		{
			Name:        "search_web",
			Description: "Search the web for information",
			Args:        SearchWeb{},
		},
	}

	openAITools := make([]openai.ChatCompletionToolParam, len(tools))
	for i, tool := range tools {
		openAITools[i] = tool.ToOpenAITool()
	}
	message := "What classes are there in Daggerheart?"
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(message),
	}
	client := openai.NewClient(
		option.WithBaseURL("http://localhost:11434/v1"),
	)
	for {

		params := openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    "qwen3:30b-a3b",
			Tools:    openAITools,
		}
		stream := client.Chat.Completions.NewStreaming(context.TODO(), params)

		acc := openai.ChatCompletionAccumulator{}
		for stream.Next() {
			chunk := stream.Current()
			fmt.Println("Received chunk:", chunk)
			acc.AddChunk(chunk)

			if content, ok := acc.JustFinishedContent(); ok {
				println("Stream finished:", content)
			}

			if refusal, ok := acc.JustFinishedRefusal(); ok {
				println("Refusal stream finished:", refusal)
			}

			// use chunks after JustFinished events
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		}
		fmt.Println()
		choices := acc.Choices
		if len(choices) == 0 {
			fmt.Println("No choices in the response")
			break
		}
		assistantMessage := choices[0].Message
		messages = append(messages, assistantMessage.ToParam())
		toolcalls := assistantMessage.ToolCalls

		if len(toolcalls) == 0 {
			fmt.Println("Done with chat")

			finalContent := acc.Choices[0].Message.Content
			fmt.Println("Final content:", finalContent)
			break
		}

		for _, toolcall := range toolcalls {
			funcname := toolcall.Function.Name
			// get the tool by name
			for _, tool := range tools {
				if tool.CompleteName() == funcname {
					result := tool.RunOnArgs(toolcall.Function.Arguments)
					toolmessage := openai.ToolMessage(utils.AsString(result), toolcall.ID)
					messages = append(messages, toolmessage)
					break
				}
			}
		}

		if stream.Err() != nil {
			panic(stream.Err())
		}

	}
}
