package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/logkn/agents-go/internal/provider"
	"github.com/logkn/agents-go/internal/response"
	"github.com/logkn/agents-go/internal/tools"
)

// Agent represents an AI agent with tools and state
type Agent struct {
	Name         string
	Instructions string
	Tools        []tools.Tool
	Provider     provider.LLMProvider
	State        any
}

// AgentExecutor handles agent execution and coordination
type AgentExecutor struct {
	agents map[string]*Agent
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor() *AgentExecutor {
	return &AgentExecutor{
		agents: make(map[string]*Agent),
	}
}

// RegisterAgent adds an agent to the executor
func (e *AgentExecutor) RegisterAgent(name string, agent *Agent) {
	e.agents[name] = agent
}

// Execute runs an agent with the given input
func (e *AgentExecutor) Execute(ctx context.Context, agentName string, input string, responseChan chan<- response.AgentResponse) error {
	agent, exists := e.agents[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	defer close(responseChan)

	// Send initial thought
	responseChan <- response.AgentResponse{
		Type:    response.ResponseTypeThought,
		Content: fmt.Sprintf("Processing request: %s", input),
	}

	// Build conversation context
	messages := []provider.Message{
		{Role: "system", Content: agent.Instructions},
		{Role: "user", Content: input},
	}

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		// Get LLM response
		llmResp, err := agent.Provider.GenerateResponse(ctx, messages, agent.Tools)
		if err != nil {
			responseChan <- response.AgentResponse{
				Type:    response.ResponseTypeFinal,
				Content: fmt.Sprintf("Error: %v", err),
			}
			return err
		}

		// Add assistant message
		messages = append(messages, provider.Message{
			Role:    "assistant",
			Content: llmResp.Content,
		})

		// Handle tool calls (potentially in parallel)
		if len(llmResp.ToolCalls) > 0 {
			// Execute tools in parallel and collect results
			toolResults := e.executeToolsParallel(ctx, agent, llmResp.ToolCalls, responseChan)

			// Add all tool results to conversation
			for _, result := range toolResults {
				messages = append(messages, provider.Message{
					Role:    "tool",
					Content: fmt.Sprintf("Tool result: %v", result.Result),
					ToolID:  result.ID,
				})
			}

			continue // Get next LLM response
		}

		// If no tool calls and response is finished, we're done
		if llmResp.Finished {
			responseChan <- response.AgentResponse{
				Type:    response.ResponseTypeFinal,
				Content: llmResp.Content,
			}
			break
		}
	}

	return nil
}

// ExecuteAgentAsTool allows one agent to invoke another as a tool
func (e *AgentExecutor) ExecuteAgentAsTool(ctx context.Context, agentName string, input string) (string, error) {
	responseChan := make(chan response.AgentResponse, 10)

	go func() {
		e.Execute(ctx, agentName, input, responseChan)
	}()

	var finalResult string
	for response := range responseChan {
		if response.Type == response.ResponseTypeFinal {
			finalResult = response.Content
		}
	}

	return finalResult, nil
}

// HandoffToAgent transfers control from one agent to another
func (e *AgentExecutor) HandoffToAgent(ctx context.Context, fromAgent, toAgent, input string, responseChan chan<- response.AgentResponse) error {
	responseChan <- response.AgentResponse{
		Type:    response.ResponseTypeHandoff,
		Content: fmt.Sprintf("Handing off from %s to %s", fromAgent, toAgent),
		Handoff: &response.AgentHandoff{
			ToAgent: toAgent,
			Reason:  "User request requires specialized handling",
			Context: input,
		},
	}

	return e.Execute(ctx, toAgent, input, responseChan)
}

// executeToolsParallel executes multiple tools concurrently and returns results
func (e *AgentExecutor) executeToolsParallel(ctx context.Context, agent *Agent, toolCalls []response.ToolCall, responseChan chan<- response.AgentResponse) []response.ToolCall {
	if len(toolCalls) == 1 {
		// Single tool call - execute directly for better error reporting
		toolCall := toolCalls[0]
		responseChan <- response.AgentResponse{
			Type:     response.ResponseTypeToolCall,
			Content:  fmt.Sprintf("Calling tool: %s", toolCall.Name),
			ToolCall: &toolCall,
		}

		result, err := e.executeTool(ctx, agent, toolCall)
		toolCall.Result = result
		if err != nil {
			toolCall.Error = err.Error()
		}

		responseChan <- response.AgentResponse{
			Type:    response.ResponseTypeIntermediate,
			Content: fmt.Sprintf("Tool %s completed: %v", toolCall.Name, result),
		}

		return []response.ToolCall{toolCall}
	}

	// Multiple tool calls - execute in parallel
	responseChan <- response.AgentResponse{
		Type:    response.ResponseTypeToolCall,
		Content: fmt.Sprintf("Executing %d tools in parallel", len(toolCalls)),
	}

	// Channel to collect results
	type toolResult struct {
		index  int
		result any
		err    error
	}

	resultChan := make(chan toolResult, len(toolCalls))

	// Start all tool executions
	for i, toolCall := range toolCalls {
		go func(index int, tc response.ToolCall) {
			result, err := e.executeTool(ctx, agent, tc)
			resultChan <- toolResult{
				index:  index,
				result: result,
				err:    err,
			}
		}(i, toolCall)
	}

	// Collect results as they complete
	completedCount := 0
	results := make([]response.ToolCall, len(toolCalls))
	copy(results, toolCalls) // Copy original tool calls

	for completedCount < len(toolCalls) {
		select {
		case res := <-resultChan:
			results[res.index].Result = res.result
			if res.err != nil {
				results[res.index].Error = res.err.Error()
			}

			completedCount++

			// Send progress update
			toolName := results[res.index].Name
			responseChan <- response.AgentResponse{
				Type:    response.ResponseTypeIntermediate,
				Content: fmt.Sprintf("Tool %s completed (%d/%d): %v", toolName, completedCount, len(toolCalls), res.result),
			}

		case <-ctx.Done():
			// Context cancelled - return partial results
			responseChan <- response.AgentResponse{
				Type:    response.ResponseTypeIntermediate,
				Content: fmt.Sprintf("Tool execution cancelled, completed %d/%d", completedCount, len(toolCalls)),
			}
			return results
		}
	}

	responseChan <- response.AgentResponse{
		Type:    response.ResponseTypeIntermediate,
		Content: fmt.Sprintf("All %d tools completed successfully", len(toolCalls)),
	}

	return results
}

func (e *AgentExecutor) executeTool(ctx context.Context, agent *Agent, toolCall response.ToolCall) (any, error) {
	// Find the tool
	var tool tools.Tool
	for _, t := range agent.Tools {
		if t.Name() == toolCall.Name {
			tool = t
			break
		}
	}

	if tool == nil {
		return nil, fmt.Errorf("tool %s not found", toolCall.Name)
	}

	// Convert parameters to JSON
	paramsJSON, err := json.Marshal(toolCall.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool parameters: %w", err)
	}

	// Execute the tool with context timeout protection
	return tool.Execute(ctx, agent.State, paramsJSON)
}