package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/logkn/agents-go/internal/types"
	"github.com/logkn/agents-go/internal/utils"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// ToolResult represents the output of a tool call executed by an agent.
// Name is the tool's name, Content is the returned value and ToolCallID is the
// identifier associated with the call.
type ToolResult struct {
	Name       string
	Content    any
	ToolCallID string
}

// Input represents the starting data for a run. Exactly one field should be
// populated.
type Input struct {
	// OfString initiates a new conversation with this user prompt.
	OfString string
	// OfMessages continues an existing conversation.
	OfMessages []types.Message
}

// Run executes the agent against the provided input and returns an
// AgentResponse for consuming the results.
// Run executes the agent and streams events back through an AgentResponse.
// If input.OfMessages is provided it is treated as the existing conversation
// history. Otherwise a new conversation is started with input.OfString as the
// user prompt.
func Run(agent types.Agent, input Input) (AgentResponse, error) {
	logger := agent.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("starting agent run",
		"agent_name", agent.Name,
		"model", agent.Model.Model,
		"num_tools", len(agent.Tools),
		"has_existing_messages", len(input.OfMessages) > 0)

	var messages []types.Message
	switch {
	case len(input.OfMessages) > 0:
		messages = input.OfMessages
		logger.Debug("using existing conversation", "message_count", len(input.OfMessages))
	default:
		messages = []types.Message{
			types.NewSystemMessage(agent.Instructions),
			types.NewUserMessage(input.OfString),
		}
		logger.Debug("starting new conversation", "user_prompt", input.OfString)
	}

	var client openai.Client
	if agent.Model.BaseUrl != "" {
		logger.Debug("using custom base URL", "base_url", agent.Model.BaseUrl)
		client = openai.NewClient(
			option.WithBaseURL(agent.Model.BaseUrl),
		)
	} else {
		logger.Debug("using OpenAI API")
		client = openai.NewClient()
	}
	// check that the model exists
	// if _, err := client.Models.Get(context.TODO(), agent.Model.Model); err != nil {
	// 	return AgentResponse{}, err
	// }

	openAITools := make([]openai.ChatCompletionToolParam, len(agent.Tools))
	for i, tool := range agent.Tools {
		openAITools[i] = tool.ToOpenAITool()
	}

	eventChannel := make(chan AgentEvent, 10)
	agentResponse := newAgentResponse(eventChannel, messages)

	go func() {
		for {
			logger.Debug("sending request to LLM", "message_count", len(messages))
			openaiMessages := utils.MapSlice(messages, types.Message.ToOpenAI)
			params := openai.ChatCompletionNewParams{
				Messages: openaiMessages,
				Model:    agent.Model.Model,
				Tools:    openAITools,
			}
			stream := client.Chat.Completions.NewStreaming(context.TODO(), params)
			acc := openai.ChatCompletionAccumulator{}
			tokenCount := 0
			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)

				if len(chunk.Choices) > 0 {
					token := chunk.Choices[0].Delta.Content
					if token != "" {
						tokenCount++
					}
					eventChannel <- tokenEvent(token)
				}
			}
			logger.Debug("received response from LLM", "tokens_received", tokenCount)
			choices := acc.Choices
			// if no choices, break the loop
			if len(choices) == 0 {
				logger.Debug("no choices returned from LLM, ending conversation")
				break
			}
			openaimsg := choices[0].Message

			// check for refusals
			if openaimsg.Refusal != "" {
				err := fmt.Errorf("LLM refusal: %s", openaimsg.Refusal)
				logger.Error("LLM refused to respond", "refusal", openaimsg.Refusal)
				eventChannel <- errorEvent(err)
				return
			}

			msg := types.AssistantMessageFromOpenAI(openaimsg, agent.Name)
			messages = append(messages, msg)

			eventChannel <- messageEvent(msg)

			toolcalls := msg.ToolCalls

			if len(toolcalls) == 0 {
				logger.Info("assistant response completed", "content_length", len(msg.Content))
				break
			}

			logger.Info("processing tool calls", "tool_call_count", len(toolcalls))

			for _, toolcall := range toolcalls {
				funcname := toolcall.Name
				logger.Debug("executing tool",
					"tool_name", funcname,
					"tool_call_id", toolcall.ID,
					"args_length", len(toolcall.Args))

				// get the tool by name
				toolFound := false
				for _, tool := range agent.Tools {
					if tool.CompleteName() == funcname {
						toolFound = true
						result := tool.RunOnArgs(toolcall.Args)
						logger.Info("tool execution completed",
							"tool_name", funcname,
							"tool_call_id", toolcall.ID)

						toolmessage := types.NewToolMessage(toolcall.ID, result)
						messages = append(messages, toolmessage)
						eventChannel <- messageEvent(toolmessage)
						eventChannel <- toolEvent(ToolResult{
							Name:       tool.CompleteName(),
							Content:    result,
							ToolCallID: toolcall.ID,
						})
						break
					}
				}
				if !toolFound {
					logger.Error("tool not found", "tool_name", funcname)
				}
			}
		}
		close(eventChannel)
	}()

	logger.Debug("agent run initiated successfully")
	return *agentResponse, nil
}
