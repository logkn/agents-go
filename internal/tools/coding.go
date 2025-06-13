package tools

import "os"

type pwd struct{}

func (p pwd) Run() any {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	return path
}

var PwdTool = Tool{
	Name:        "pwd",
	Description: "Get the current working directory.",
	Args:        pwd{},
}
