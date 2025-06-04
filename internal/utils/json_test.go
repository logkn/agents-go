package utils

import "testing"

type sample struct {
	Name string `json:"name"`
}

func TestJsonDumps(t *testing.T) {
	m := map[string]any{"a": 1}
	s := JsonDumps(m, 2)
	expected := "{\n  \"a\": 1\n}"
	if s != expected {
		t.Fatalf("expected %q got %q", expected, s)
	}
}

func TestJsonDumpsObj(t *testing.T) {
	obj := sample{Name: "foo"}
	s := JsonDumpsObj(obj)
	expected := "{\n  \"name\": \"foo\"\n}"
	if s != expected {
		t.Fatalf("expected %q got %q", expected, s)
	}
}

func TestUnescapeUnicode(t *testing.T) {
	in := "Hello \\u0041"
	out := unescapeUnicode(in)
	if out != "Hello A" {
		t.Fatalf("expected 'Hello A' got %q", out)
	}
}
