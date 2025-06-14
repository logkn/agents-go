package tools

import (
	"os"
	"strings"
)

// ReadFile reads the content of a file and returns it as a string.
type ReadFile struct {
	Path string `json:"path" description:"Path to the file"`
}

func (r ReadFile) Run() any {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return string(data)
}

// WriteFile writes the given content to a file, replacing any existing content.
type WriteFile struct {
	Path    string `json:"path" description:"Path to the file"`
	Content string `json:"content" description:"Content to write"`
}

func (w WriteFile) Run() any {
	err := os.WriteFile(w.Path, []byte(w.Content), 0o644)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return "ok"
}

// Replace finds Old in the file at Path and replaces it with New.
// If All is true, all occurrences are replaced. Otherwise exactly one
// occurrence must exist or an error is returned.
type Replace struct {
	Path string `json:"path" description:"Path to the file"`
	Old  string `json:"old" description:"Substring to replace"`
	New  string `json:"new" description:"Replacement text"`
	All  bool   `json:"all" description:"Replace all occurrences"`
}

func (r Replace) Run() any {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	content := string(data)
	if r.All {
		content = strings.ReplaceAll(content, r.Old, r.New)
	} else {
		count := strings.Count(content, r.Old)
		if count == 0 {
			return map[string]any{"error": "substring not found"}
		}
		if count > 1 {
			return map[string]any{"error": "substring occurs more than once"}
		}
		content = strings.Replace(content, r.Old, r.New, 1)
	}
	if err := os.WriteFile(r.Path, []byte(content), 0o644); err != nil {
		return map[string]any{"error": err.Error()}
	}
	return "ok"
}
