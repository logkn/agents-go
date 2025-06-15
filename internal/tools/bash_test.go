package tools

import "testing"

func TestBashRun(t *testing.T) {
	b := Bash{Command: "echo hello"}
	out := b.Run()
	if out != "hello\n" {
		t.Fatalf("unexpected output: %v", out)
	}
}
