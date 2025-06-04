package types

import (
	"github.com/openai/openai-go"
	"testing"
)

type dummyToolCall struct{ ID, Name, Args string }

func TestToolCallConversions(t *testing.T) {
	tc := ToolCall{ID: "1", Name: "foo", Args: "{}"}
	oa := tc.ToOpenAI()
	if oa.ID != "1" || oa.Function.Name != "foo" {
		t.Fatalf("unexpected openai conversion")
	}
	back := ToolCallFromOpenAI(openai.ChatCompletionMessageToolCall{ID: oa.ID, Function: openai.ChatCompletionMessageToolCallFunction{Name: oa.Function.Name, Arguments: oa.Function.Arguments}})
	if back != tc {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestMessageHelpers(t *testing.T) {
	m := NewUserMessage("hi")
	if m.Role != User || m.Content != "hi" {
		t.Fatalf("user message wrong")
	}
	a := NewAssistantMessage("yo", "bot", nil)
	if a.Role != Assistant || a.Name != "bot" {
		t.Fatalf("assistant message wrong")
	}
	s := NewSystemMessage("sys")
	if s.Role != System {
		t.Fatalf("system message wrong")
	}
	tmsg := NewToolMessage("tid", "out")
	if tmsg.Role != Tool || tmsg.ID != "tid" {
		t.Fatalf("tool message wrong")
	}
}
