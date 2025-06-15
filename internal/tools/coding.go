package tools

import "os"

// pwd is the argument structure for the PwdTool. It has no fields.
type pwd struct{}

func (p pwd) Run() any {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	return path
}

// PwdTool exposes the current working directory as a tool callable by an
// agent.
var PwdTool = Tool{
	Name:        "pwd",
	Description: "Get the current working directory.",
	Args:        pwd{},
}
