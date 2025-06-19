package runner

import (
	"testing"

	"github.com/logkn/agents-go/internal/types"
)

func TestAgentEventAccessors(t *testing.T) {
	msg := types.NewUserMessage("hello")
	event := messageEvent(msg)

	if m, ok := event.Message(); !ok || m.Content != "hello" {
		t.Fatalf("message accessor failed")
	}

	tokenEvt := tokenEvent("t")
	if tok, ok := tokenEvt.Token(); !ok || tok != "t" {
		t.Fatalf("token accessor failed")
	}

	tr := ToolResult{Name: "n", Content: "c", ToolCallID: "id"}
	toolEvt := toolEvent(tr)
	if r, ok := toolEvt.ToolResult(); !ok || r.Name != "n" {
		t.Fatalf("tool accessor failed")
	}

	errEvt := AgentEvent{OfError: ErrTest}
	if e, ok := errEvt.Error(); !ok || e != ErrTest {
		t.Fatalf("error accessor failed")
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }

var ErrTest = testErr("boom")
