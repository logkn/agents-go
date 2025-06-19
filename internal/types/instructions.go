package types

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/logkn/agents-go/internal/context"
	agentcontext "github.com/logkn/agents-go/internal/context"
)

type AgentInstructions struct {
	OfString string
	OfFile   string
}

func StringInstructions(s string) AgentInstructions {
	return AgentInstructions{OfString: s}
}

func FileInstructions(file string) AgentInstructions {
	return AgentInstructions{OfFile: file}
}

func (ins AgentInstructions) getContent() (string, error) {
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

func (ins AgentInstructions) ToString(ctx agentcontext.AnyContext) (string, error) {
	content, err := ins.getContent()
	if err != nil {
		return "", err
	}

	fmt.Println(ctx)

	templ, err := template.New("instructions").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}

	// get the value out of the context
	ctxVal, err := context.FromAnyContext[any](ctx)
	if err != nil {
		return "", err
	}

	// make a buffer to hold the output
	var buffer bytes.Buffer
	err = templ.Execute(&buffer, ctxVal.Value())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
