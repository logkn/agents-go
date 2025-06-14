package tools

import (
	"os"
	"strings"
)

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

type fileread struct {
	FilePath string `json:"file_path" description:"The path to the file to read"`
	Limit    int    `json:"limit,omitempty" description:"The maximum number of lines to read"`
	Offset   int    `json:"offset,omitempty" description:"The offset from the beginning of the file"`
}

func (f fileread) Run() any {
	content, err := os.ReadFile(f.FilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	start := max(f.Offset, 0)
	if start >= len(lines) {
		return ""
	}

	end := len(lines)
	if f.Limit > 0 {
		end = min(start+f.Limit, len(lines))
	}

	return strings.Join(lines[start:end], "\n")
}

var FileReadTool = Tool{
	Args:        fileread{},
	Description: "Reads a file and returns its contents as a string. The file is read starting from the offset line and limited to the specified number of lines.",
	Name:        "file_read",
}

type List struct {
	Path string `json:"path" description:"The path to the directory to list"`
}

func (l List) Run() any {
	files, err := os.ReadDir(l.Path)
	if err != nil {
		return err
	}
	var names []string
	for _, file := range files {
		names = append(names, file.Name())
	}
	return names
}

var ListTool = Tool{
	Args:        List{},
	Description: "Lists the files in a directory.",
	Name:        "list",
}

type filewrite struct {
	FilePath string `json:"file_path" description:"The path to the file to write"`
	Content  string `json:"content" description:"The content to write to the file"`
}

func (f filewrite) Run() any {
	err := os.WriteFile(f.FilePath, []byte(f.Content), 0644)
	if err != nil {
		return err
	}
	return nil
}

var FileWriteTool = Tool{
	Args:        filewrite{},
	Description: "Writes text to a file.",
	Name:        "file_write",
}

type patch struct {
	FilePath  string `json:"file_path" description:"The path to the file to patch"`
	OldString string `json:"old_string" description:"The string to replace"`
	NewString string `json:"new_string" description:"The new string to replace the old string with"`
	All       bool   `json:"all" description:"Whether to replace all occurrences of the old string"`
}

func (p patch) Run() any {
	content, err := os.ReadFile(p.FilePath)
	if err != nil {
		return err
	}

	if p.All {
		content = []byte(strings.ReplaceAll(string(content), p.OldString, p.NewString))
	} else {
		content = []byte(strings.Replace(string(content), p.OldString, p.NewString, 1))
	}

	err = os.WriteFile(p.FilePath, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

var PatchTool = Tool{
	Args:        patch{},
	Description: "Patches a file by replacing a string with another string.",
	Name:        "patch",
}

// type bash struct {
// 	Command string `json:"command" description:"The command to run"`
// }
//
// func (b bash) Run() any {
// 	cmd := exec.Command("bash", "-c", b.Command)
//
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return err
// 	}
// 	return string(output)
// }
//
// var BashTool = Tool{
// 	Args:        bash{},
// 	Description: "Runs a bash command and returns the output.",
// 	Name:        "bash",
// }
