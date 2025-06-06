package tools

import (
	"os"
	"testing"
)

func TestReadWriteReplace(t *testing.T) {
	tmp, err := os.CreateTemp("", "fileops")
	if err != nil {
		t.Fatal(err)
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	// WriteFile
	w := WriteFile{Path: path, Content: "foo"}
	if w.Run() != "ok" {
		t.Fatalf("write failed")
	}

	// ReadFile
	r := ReadFile{Path: path}
	out := r.Run()
	if out != "foo" {
		t.Fatalf("read got %v", out)
	}

	// Replace
	rep := Replace{Path: path, Old: "foo", New: "bar", All: false}
	if rep.Run() != "ok" {
		t.Fatalf("replace failed")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "bar" {
		t.Fatalf("replace result %s", data)
	}
}
