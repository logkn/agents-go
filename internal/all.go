package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// ResponseType represents different types of agent responses
type ResponseType string

const (
	ResponseTypeThought      ResponseType = "thought"
	ResponseTypeIntermediate ResponseType = "intermediate"
	ResponseTypeToolCall     ResponseType = "tool_call"
	ResponseTypeFinal        ResponseType = "final"
	ResponseTypeHandoff      ResponseType = "handoff"
)

// AgentResponse represents a response from an agent
type AgentResponse struct {
	Type     ResponseType   `json:"type"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	ToolCall *ToolCall      `json:"tool_call,omitempty"`
	Handoff  *AgentHandoff  `json:"handoff,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters"`
	Result     any            `json:"result,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// AgentHandoff represents transferring control to another agent
type AgentHandoff struct {
	ToAgent string `json:"to_agent"`
	Reason  string `json:"reason"`
	Context string `json:"context"`
}

// ============================================================================
// Tool Framework
// ============================================================================

// Tool interface that all tools must implement
type Tool interface {
	Name() string
	Description() string
	JSONSchema() map[string]any
	Execute(ctx context.Context, state any, paramsJSON []byte) (any, error)
}

// ToolOption allows customizing tool registration
type ToolOption func(*toolConfig)

type toolConfig struct {
	name        string
	description string
}

// WithName sets a custom name for the tool
func WithName(name string) ToolOption {
	return func(c *toolConfig) {
		c.name = name
	}
}

// WithDescription sets a custom description for the tool
func WithDescription(desc string) ToolOption {
	return func(c *toolConfig) {
		c.description = desc
	}
}

// reflectedTool wraps a function to implement the Tool interface
type reflectedTool struct {
	fn          reflect.Value
	fnType      reflect.Type
	paramsType  reflect.Type
	stateType   reflect.Type
	name        string
	description string
	schema      map[string]any
}

// RegisterTool converts any properly-structured function into a Tool
func RegisterTool(fn any, opts ...ToolOption) Tool {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		panic("RegisterTool: argument must be a function")
	}

	// Validate function signature: func(context.Context, StateInterface, ParamsStruct) (ResultType, error)
	if fnType.NumIn() != 3 {
		panic("RegisterTool: function must have exactly 3 parameters (ctx, state, params)")
	}

	if fnType.NumOut() != 2 {
		panic("RegisterTool: function must return (result, error)")
	}

	// Check parameter types
	ctxType := fnType.In(0)
	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic("RegisterTool: first parameter must be context.Context")
	}

	stateType := fnType.In(1)
	paramsType := fnType.In(2)

	// Check return types
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !fnType.Out(1).Implements(errorType) {
		panic("RegisterTool: second return value must be error")
	}

	// Extract configuration
	config := &toolConfig{
		name:        extractFunctionName(fnValue),
		description: fmt.Sprintf("Executes %s", extractFunctionName(fnValue)),
	}

	for _, opt := range opts {
		opt(config)
	}

	// Generate JSON schema from params struct
	schema := generateJSONSchema(paramsType)

	return &reflectedTool{
		fn:          fnValue,
		fnType:      fnType,
		paramsType:  paramsType,
		stateType:   stateType,
		name:        config.name,
		description: config.description,
		schema:      schema,
	}
}

func (t *reflectedTool) Name() string {
	return t.name
}

func (t *reflectedTool) Description() string {
	return t.description
}

func (t *reflectedTool) JSONSchema() map[string]any {
	return t.schema
}

func (t *reflectedTool) Execute(ctx context.Context, state any, paramsJSON []byte) (any, error) {
	// Create new instance of params struct
	paramsValue := reflect.New(t.paramsType).Elem()

	// Unmarshal JSON into params struct
	paramsPtr := paramsValue.Addr().Interface()
	if err := json.Unmarshal(paramsJSON, paramsPtr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	// Type assert state to required interface
	stateValue := reflect.ValueOf(state)
	if !stateValue.Type().AssignableTo(t.stateType) {
		return nil, fmt.Errorf("state does not implement required interface %s", t.stateType)
	}

	// Call the function
	results := t.fn.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		stateValue,
		paramsValue,
	})

	// Extract results
	result := results[0].Interface()
	errValue := results[1].Interface()

	if errValue != nil {
		return result, errValue.(error)
	}

	return result, nil
}

// Helper functions for reflection and schema generation

func extractFunctionName(fn reflect.Value) string {
	fullName := fn.Type().String()
	// Extract just the function name from the full type string
	if idx := strings.LastIndex(fullName, "."); idx != -1 {
		return fullName[idx+1:]
	}
	return fullName
}

func generateJSONSchema(t reflect.Type) map[string]any {
	if t.Kind() != reflect.Struct {
		panic("generateJSONSchema: type must be a struct")
	}

	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		isRequired := true

		// Parse json tag
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			// Check for omitempty
			for _, part := range parts[1:] {
				if part == "omitempty" {
					isRequired = false
				}
			}
		}

		// Generate property schema
		prop := generateFieldSchema(field.Type)

		// Add description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			prop["description"] = desc
		}

		properties[fieldName] = prop

		if isRequired {
			required = append(required, fieldName)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func generateFieldSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": generateFieldSchema(t.Elem()),
		}
	case reflect.Struct:
		return generateJSONSchema(t)
	case reflect.Ptr:
		return generateFieldSchema(t.Elem())
	default:
		return map[string]any{"type": "string"} // fallback
	}
}

// ============================================================================
// LLM Provider Abstraction
// ============================================================================

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	ToolID  string `json:"tool_id,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// LLMProvider abstracts different LLM providers
type LLMProvider interface {
	GenerateResponse(ctx context.Context, messages []Message, tools []Tool) (*LLMResponse, error)
	SupportsStreaming() bool
}

// LLMResponse represents a response from an LLM
type LLMResponse struct {
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
	Finished  bool        `json:"finished"`
	Usage     *TokenUsage `json:"usage,omitempty"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// MockLLMProvider is a simple implementation for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, messages []Message, tools []Tool) (*LLMResponse, error) {
	// Simple mock - just echo the last message
	if len(messages) == 0 {
		return &LLMResponse{
			Content:  "Hello! How can I help you?",
			Finished: true,
		}, nil
	}

	lastMsg := messages[len(messages)-1]

	// If user mentions weather and email together, use both tools in parallel
	content := strings.ToLower(lastMsg.Content)
	if strings.Contains(content, "weather") && strings.Contains(content, "email") {
		var toolCalls []ToolCall

		for _, tool := range tools {
			if tool.Name() == "get_weather" {
				toolCalls = append(toolCalls, ToolCall{
					ID:   "call_weather",
					Name: "get_weather",
					Parameters: map[string]any{
						"city": "San Francisco",
					},
				})
			}
			if tool.Name() == "send_email" {
				toolCalls = append(toolCalls, ToolCall{
					ID:   "call_email",
					Name: "send_email",
					Parameters: map[string]any{
						"to":      "user@example.com",
						"subject": "Weather Update",
						"body":    "Here's your weather update!",
					},
				})
			}
		}

		if len(toolCalls) > 0 {
			return &LLMResponse{
				Content:   "I'll check the weather and send you an email with the information.",
				ToolCalls: toolCalls,
				Finished:  false,
			}, nil
		}
	}

	// If user mentions weather, use the weather tool
	if strings.Contains(content, "weather") {
		for _, tool := range tools {
			if tool.Name() == "get_weather" {
				return &LLMResponse{
					Content: "I'll check the weather for you.",
					ToolCalls: []ToolCall{
						{
							ID:   "call_1",
							Name: "get_weather",
							Parameters: map[string]any{
								"city": "San Francisco",
							},
						},
					},
					Finished: false,
				}, nil
			}
		}
	}

	return &LLMResponse{
		Content:  fmt.Sprintf("I received: %s", lastMsg.Content),
		Finished: true,
	}, nil
}

func (m *MockLLMProvider) SupportsStreaming() bool {
	return false
}

// ============================================================================
// Agent System
// ============================================================================

// Agent represents an AI agent with tools and state
type Agent struct {
	Name         string
	Instructions string
	Tools        []Tool
	Provider     LLMProvider
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
func (e *AgentExecutor) Execute(ctx context.Context, agentName string, input string, responseChan chan<- AgentResponse) error {
	agent, exists := e.agents[agentName]
	if !exists {
		return fmt.Errorf("agent %s not found", agentName)
	}

	defer close(responseChan)

	// Send initial thought
	responseChan <- AgentResponse{
		Type:    ResponseTypeThought,
		Content: fmt.Sprintf("Processing request: %s", input),
	}

	// Build conversation context
	messages := []Message{
		{Role: "system", Content: agent.Instructions},
		{Role: "user", Content: input},
	}

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		// Get LLM response
		llmResp, err := agent.Provider.GenerateResponse(ctx, messages, agent.Tools)
		if err != nil {
			responseChan <- AgentResponse{
				Type:    ResponseTypeFinal,
				Content: fmt.Sprintf("Error: %v", err),
			}
			return err
		}

		// Add assistant message
		messages = append(messages, Message{
			Role:    "assistant",
			Content: llmResp.Content,
		})

		// Handle tool calls (potentially in parallel)
		if len(llmResp.ToolCalls) > 0 {
			// Execute tools in parallel and collect results
			toolResults := e.executeToolsParallel(ctx, agent, llmResp.ToolCalls, responseChan)

			// Add all tool results to conversation
			for _, result := range toolResults {
				messages = append(messages, Message{
					Role:    "tool",
					Content: fmt.Sprintf("Tool result: %v", result.Result),
					ToolID:  result.ID,
				})
			}

			continue // Get next LLM response
		}

		// If no tool calls and response is finished, we're done
		if llmResp.Finished {
			responseChan <- AgentResponse{
				Type:    ResponseTypeFinal,
				Content: llmResp.Content,
			}
			break
		}
	}

	return nil
}

// ExecuteAgentAsTool allows one agent to invoke another as a tool
func (e *AgentExecutor) ExecuteAgentAsTool(ctx context.Context, agentName string, input string) (string, error) {
	responseChan := make(chan AgentResponse, 10)

	go func() {
		e.Execute(ctx, agentName, input, responseChan)
	}()

	var finalResult string
	for response := range responseChan {
		if response.Type == ResponseTypeFinal {
			finalResult = response.Content
		}
	}

	return finalResult, nil
}

// HandoffToAgent transfers control from one agent to another
func (e *AgentExecutor) HandoffToAgent(ctx context.Context, fromAgent, toAgent, input string, responseChan chan<- AgentResponse) error {
	responseChan <- AgentResponse{
		Type:    ResponseTypeHandoff,
		Content: fmt.Sprintf("Handing off from %s to %s", fromAgent, toAgent),
		Handoff: &AgentHandoff{
			ToAgent: toAgent,
			Reason:  "User request requires specialized handling",
			Context: input,
		},
	}

	return e.Execute(ctx, toAgent, input, responseChan)
}

// executeToolsParallel executes multiple tools concurrently and returns results
func (e *AgentExecutor) executeToolsParallel(ctx context.Context, agent *Agent, toolCalls []ToolCall, responseChan chan<- AgentResponse) []ToolCall {
	if len(toolCalls) == 1 {
		// Single tool call - execute directly for better error reporting
		toolCall := toolCalls[0]
		responseChan <- AgentResponse{
			Type:     ResponseTypeToolCall,
			Content:  fmt.Sprintf("Calling tool: %s", toolCall.Name),
			ToolCall: &toolCall,
		}

		result, err := e.executeTool(ctx, agent, toolCall)
		toolCall.Result = result
		if err != nil {
			toolCall.Error = err.Error()
		}

		responseChan <- AgentResponse{
			Type:    ResponseTypeIntermediate,
			Content: fmt.Sprintf("Tool %s completed: %v", toolCall.Name, result),
		}

		return []ToolCall{toolCall}
	}

	// Multiple tool calls - execute in parallel
	responseChan <- AgentResponse{
		Type:    ResponseTypeToolCall,
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
		go func(index int, tc ToolCall) {
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
	results := make([]ToolCall, len(toolCalls))
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
			responseChan <- AgentResponse{
				Type:    ResponseTypeIntermediate,
				Content: fmt.Sprintf("Tool %s completed (%d/%d): %v", toolName, completedCount, len(toolCalls), res.result),
			}

		case <-ctx.Done():
			// Context cancelled - return partial results
			responseChan <- AgentResponse{
				Type:    ResponseTypeIntermediate,
				Content: fmt.Sprintf("Tool execution cancelled, completed %d/%d", completedCount, len(toolCalls)),
			}
			return results
		}
	}

	responseChan <- AgentResponse{
		Type:    ResponseTypeIntermediate,
		Content: fmt.Sprintf("All %d tools completed successfully", len(toolCalls)),
	}

	return results
}

func (e *AgentExecutor) executeTool(ctx context.Context, agent *Agent, toolCall ToolCall) (any, error) {
	// Find the tool
	var tool Tool
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

// ============================================================================
// Example Usage and State Interfaces
// ============================================================================

// Example state interfaces
type WeatherStateReader interface {
	GetWeatherAPIKey() string
	GetDefaultUnits() string
}

type EmailStateReader interface {
	GetEmailConfig() EmailConfig
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

// Example global state that implements multiple interfaces
type ExampleGlobalState struct {
	WeatherAPIKey string
	DefaultUnits  string
	EmailConfig   EmailConfig
}

func (s *ExampleGlobalState) GetWeatherAPIKey() string {
	return s.WeatherAPIKey
}

func (s *ExampleGlobalState) GetDefaultUnits() string {
	return s.DefaultUnits
}

func (s *ExampleGlobalState) GetEmailConfig() EmailConfig {
	return s.EmailConfig
}

// Example tool functions
func GetWeather(
	ctx context.Context,
	state WeatherStateReader,
	params struct {
		City  string `json:"city" description:"The city to get weather for"`
		Units string `json:"units,omitempty" description:"Temperature units (celsius/fahrenheit)"`
	},
) (string, error) {
	apiKey := state.GetWeatherAPIKey()
	units := params.Units
	if units == "" {
		units = state.GetDefaultUnits()
	}

	// Simulate API call
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("Weather in %s: 72°F, sunny (API key: %s, units: %s)",
		params.City, apiKey, units), nil
}

func SendEmail(
	ctx context.Context,
	state EmailStateReader,
	params struct {
		To      string `json:"to" description:"Recipient email address"`
		Subject string `json:"subject" description:"Email subject"`
		Body    string `json:"body" description:"Email body"`
	},
) (string, error) {
	config := state.GetEmailConfig()

	// Simulate sending email
	time.Sleep(200 * time.Millisecond)
	return fmt.Sprintf("Email sent to %s via %s:%d",
		params.To, config.SMTPHost, config.SMTPPort), nil
}

// Example usage
func main() {
	// Create global state
	globalState := &ExampleGlobalState{
		WeatherAPIKey: "weather-api-key-123",
		DefaultUnits:  "fahrenheit",
		EmailConfig: EmailConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password",
		},
	}

	// Register tools
	weatherTool := RegisterTool(GetWeather, WithDescription("Get current weather for any city"))
	emailTool := RegisterTool(SendEmail, WithDescription("Send an email"))

	// Create agent
	agent := &Agent{
		Name:         "Assistant",
		Instructions: "You are a helpful assistant that can check weather and send emails.",
		Tools:        []Tool{weatherTool, emailTool},
		Provider:     &MockLLMProvider{},
		State:        globalState,
	}

	// Create executor and register agent
	executor := NewAgentExecutor()
	executor.RegisterAgent("assistant", agent)

	// Execute agent
	ctx := context.Background()
	responseChan := make(chan AgentResponse, 10)

	// Test parallel tool execution
	fmt.Println("=== Testing Parallel Tool Execution ===")
	go func() {
		err := executor.Execute(ctx, "assistant", "Check the weather and send me an email about it", responseChan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Handle responses
	for response := range responseChan {
		fmt.Printf("[%s] %s\n", response.Type, response.Content)
		if response.ToolCall != nil {
			fmt.Printf("  Tool: %s(%v) -> %v\n",
				response.ToolCall.Name,
				response.ToolCall.Parameters,
				response.ToolCall.Result)
		}
	}

	// Test single tool execution
	fmt.Println("\n=== Testing Single Tool Execution ===")
	responseChan = make(chan AgentResponse, 10)
	go func() {
		err := executor.Execute(ctx, "assistant", "What's the weather in New York?", responseChan)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Handle responses
	for response := range responseChan {
		fmt.Printf("[%s] %s\n", response.Type, response.Content)
		if response.ToolCall != nil {
			fmt.Printf("  Tool: %s(%v) -> %v\n",
				response.ToolCall.Name,
				response.ToolCall.Parameters,
				response.ToolCall.Result)
		}
	}

	// Demonstrate tool schema generation
	fmt.Println("\n=== Tool Schemas ===")
	for _, tool := range agent.Tools {
		schema, _ := json.MarshalIndent(tool.JSONSchema(), "", "  ")
		fmt.Printf("\n%s:\n%s\n", tool.Name(), schema)
	}
}
