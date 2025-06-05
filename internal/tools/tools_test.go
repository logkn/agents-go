package tools

import (
	"encoding/json"
	"testing"
)

type adder struct{ A, B int }

func (a adder) Run() any { return a.A + a.B }

func TestCompleteName(t *testing.T) {
	tool := Tool{Args: adder{}}
	if tool.CompleteName() != "adder" {
		t.Fatalf("unexpected name %s", tool.CompleteName())
	}
}

func TestRunOnArgs(t *testing.T) {
	tool := Tool{Name: "add", Args: adder{}}
	args := adder{A: 1, B: 2}
	data, _ := json.Marshal(args)
	res := tool.RunOnArgs(string(data))
	if res != 3 {
		t.Fatalf("expected 3 got %v", res)
	}
}
