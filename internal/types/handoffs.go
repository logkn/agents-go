package types

import (
	"strings"

	"github.com/stoewer/go-strcase"
)

type Handoff[Context any] struct {
	Agent           *Agent[Context]
	ToolName        string
	ToolDescription string
}

func (h Handoff[Context]) defaultName() string {
	// "transfer_to_{agent_name}"
	snakecaseName := strings.ReplaceAll(h.Agent.Name, " ", "_")
	snakecaseName = strcase.SnakeCase(snakecaseName)
	return "transfer_to_" + snakecaseName
}

func (h Handoff[Context]) fullname() string {
	if h.ToolName != "" {
		return h.ToolName
	}
	return h.defaultName()
}

func (h Handoff[Context]) defaultDescription() string {
	return "Handoff to the " + h.Agent.Name + " agent to handle the request."
}

func (h Handoff[Context]) description() string {
	if h.ToolDescription != "" {
		return h.ToolDescription
	}
	return h.defaultDescription()
}

type handoffToolArgs[Context any] struct{}

func (h handoffToolArgs[Context]) Run(ctx *Context) any {
	return "handoff_executed"
}
