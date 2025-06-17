package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

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

// findHandoffByToolName searches for a handoff that matches the given tool name
func findHandoffByToolName(agent types.Agent, toolName string) *types.Handoff {
	for _, handoff := range agent.Handoffs {
		if handoff.ToolName == toolName ||
			(handoff.ToolName == "" && strings.HasPrefix(toolName, "transfer_to_")) {
			return &handoff
		}
	}
	return nil
}

// isHandoffTool checks if the given tool name corresponds to a handoff tool
func isHandoffTool(agent types.Agent, toolName string) bool {
	return findHandoffByToolName(agent, toolName) != nil
}

// Run executes the agent against the provided input and returns an
// AgentResponse for consuming the results.
// Run executes the agent and streams events back through an AgentResponse.
// If input.OfMessages is provided it is treated as the existing conversation
// history. Otherwise a new conversation is started with input.OfString as the
// user prompt.
func Run(ctx context.Context, agent types.Agent, input Input) (AgentResponse, error) {
	logger := agent.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("starting agent run",
		"agent_name", agent.Name,
		"model", agent.Model.Model,
		"num_tools", len(agent.Tools),
		"has_existing_messages", len(input.OfMessages) > 0)

	// Execute BeforeRun hook
	if agent.Hooks != nil && agent.Hooks.BeforeRun != nil {
		if err := agent.Hooks.BeforeRun(agent.Context); err != nil {
			logger.Error("BeforeRun hook failed", "error", err)
			return AgentResponse{}, fmt.Errorf("BeforeRun hook failed: %w", err)
		}
	}

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

	allTools := agent.AllToolsWithContext()
	openAITools := make([]openai.ChatCompletionToolParam, len(allTools))
	for i, tool := range allTools {
		openAITools[i] = tool.ToOpenAITool()
	}

	eventChannel := make(chan AgentEvent, 10)
	agentResponse := newAgentResponse(eventChannel, messages)

	go func() {
		for {
			logger.Debug("sending request to LLM", "message_count", len(messages))
			openaiMessages := utils.MapSlice(messages, types.Message.ToOpenAI)
			params := openai.ChatCompletionNewParams{
				Messages:    openaiMessages,
				Model:       agent.Model.Model,
				Tools:       openAITools,
				Temperature: openai.Float(0.6),
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

				// Check if this is a handoff tool
				if handoff := findHandoffByToolName(agent, funcname); handoff != nil {
					logger.Info("executing handoff",
						"from_agent", agent.Name,
						"to_agent", handoff.Agent.Name,
						"tool_call_id", toolcall.ID)

					// Parse handoff arguments to get the prompt
					var args struct {
						Prompt string `json:"prompt"`
					}
					if err := json.Unmarshal([]byte(toolcall.Args), &args); err != nil {
						logger.Error("failed to parse handoff arguments", "error", err)
						continue
					}

					// Emit handoff event
					eventChannel <- handoffEvent(HandoffEvent{
						FromAgent: agent.Name,
						ToAgent:   handoff.Agent.Name,
						Prompt:    args.Prompt,
					})

					// Create tool result message for the handoff
					toolmessage := types.NewToolMessage(toolcall.ID, "Transferring to "+handoff.Agent.Name+" agent")
					messages = append(messages, toolmessage)
					eventChannel <- messageEvent(toolmessage)

					// Switch to the handoff agent and continue with the new prompt
					agent = *handoff.Agent
					if agent.Logger == nil {
						agent.Logger = logger
					}

					// Add the handoff prompt as a user message
					messages = append(messages, types.NewUserMessage(args.Prompt))

					// Update tool list for the new agent
					allTools = agent.AllToolsWithContext()
					openAITools = make([]openai.ChatCompletionToolParam, len(allTools))
					for i, tool := range allTools {
						openAITools[i] = tool.ToOpenAITool()
					}

					logger.Info("handoff completed", "new_agent", agent.Name)
					continue
				}

				// Regular tool execution
				toolFound := false
				for _, tool := range allTools {
					if tool.CompleteName() == funcname {
						toolFound = true
						
						// Execute BeforeToolCall hook
						if agent.Hooks != nil && agent.Hooks.BeforeToolCall != nil {
							if err := agent.Hooks.BeforeToolCall(agent.Context, funcname, toolcall.Args); err != nil {
								logger.Error("BeforeToolCall hook failed", "error", err, "tool_name", funcname)
								continue
							}
						}
						
						var result any
						// Use contextual execution if tool has context, otherwise use regular execution
						if tool.Context != nil {
							result = tool.RunOnArgsWithContext(toolcall.Args)
						} else {
							result = tool.RunOnArgs(toolcall.Args)
						}
						
						// Execute AfterToolCall hook
						if agent.Hooks != nil && agent.Hooks.AfterToolCall != nil {
							if err := agent.Hooks.AfterToolCall(agent.Context, funcname, result); err != nil {
								logger.Error("AfterToolCall hook failed", "error", err, "tool_name", funcname)
							}
						}
						
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
		
		// Execute AfterRun hook before stopping
		if agent.Hooks != nil && agent.Hooks.AfterRun != nil {
			// Get the final response content for the hook
			finalResponse := ""
			if len(messages) > 0 {
				for i := len(messages) - 1; i >= 0; i-- {
					if messages[i].Role == types.Assistant && messages[i].Content != "" {
						finalResponse = messages[i].Content
						break
					}
				}
			}
			if err := agent.Hooks.AfterRun(agent.Context, finalResponse); err != nil {
				logger.Error("AfterRun hook failed", "error", err)
			}
		}
		
		agentResponse.Stop()
	}()

	logger.Debug("agent run initiated successfully")
	return *agentResponse, nil
}
