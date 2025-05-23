package tools

import "context"

var ThinkTool Tool = CreateTool(
	func(ctx context.Context, state any, paramStruct struct {
		Thought string `json:"thought" description:"The thought to append to the log"`
	},
	) (any, error) {
		return true, nil
	},
	WithName("Think"),
	WithDescription("Use the tool to think about something. It will not obtain new information or change the database, but just append the thought to the log. Use it when complex reasoning or some cache memory is needed."),
)
