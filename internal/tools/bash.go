package tools

import (
	"bytes"
	"os/exec"
)

// Bash runs a shell command using bash -c.
// The output of the command (stdout and stderr) is returned as a string.
type Bash struct {
	Command string `json:"command" description:"Command to execute"`
}

func (b Bash) Run() any {
	if b.Command == "" {
		return map[string]any{"error": "command cannot be empty"}
	}
	cmd := exec.Command("bash", "-c", b.Command)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]any{"error": err.Error(), "stderr": stderr.String(), "stdout": out.String()}
	}
	if stderr.Len() > 0 {
		return map[string]any{"stderr": stderr.String(), "stdout": out.String()}
	}
	return out.String()
}
