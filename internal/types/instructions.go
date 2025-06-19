package types

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type AgentInstructions[Context any] struct {
	OfString string
	OfFile   string
}

func StringInstructions[Context any](s string) AgentInstructions[Context] {
	return AgentInstructions[Context]{OfString: s}
}

func FileInstructions[Context any](file string) AgentInstructions[Context] {
	return AgentInstructions[Context]{OfFile: file}
}

func (ins AgentInstructions[Context]) getContent() (string, error) {
	// Case: OfString
	if ins.OfString != "" {
		return ins.OfString, nil
	}

	// Case: Neither
	if ins.OfFile == "" {
		return "", errors.New("no instruction provided for agent")
	}

	// Case: OfFile
	// Expand user home directory if path starts with ~
	filePath := ins.OfFile
	if strings.HasPrefix(filePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Read file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (ins AgentInstructions[Context]) ToString(ctx *Context) (string, error) {
	content, err := ins.getContent()
	if err != nil {
		return "", err
	}

	templ, err := template.New("instructions").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}

	// get the value out of the context
	if err != nil {
		return "", err
	}

	// make a buffer to hold the output
	var buffer bytes.Buffer
	err = templ.Execute(&buffer, ctx)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
