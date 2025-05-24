package tools

import "github.com/logkn/agents-go/internal/events"

type ThinkTool struct {
	Thought string `json:"thought" description:"The thought to append to the log"`
}

func (t ThinkTool) Execute(state any, events events.EventBus) (any, error) {
	return true, nil
}

var thinkTool Tool = NewTool(ThinkTool{}).WithName("Think").WithDescription("Use the tool to think about something. It will not obtain new information or change the database, but just append the thought to the log. Use it when complex reasoning or some cache memory is needed.")
