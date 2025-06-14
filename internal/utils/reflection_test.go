package utils

import "testing"

func TestTypeName(t *testing.T) {
	type myStruct struct{}
	var v *myStruct
	if TypeName(v) != "myStruct" {
		t.Fatalf("unexpected name %s", TypeName(v))
	}
}

func TestNewInstance(t *testing.T) {
	type myStruct struct{ A int }
	inst := NewInstance(myStruct{}).(*myStruct)
	if inst.A != 0 {
		t.Fatalf("expected zero value")
	}
}

func TestAsString(t *testing.T) {
	if AsString(nil) != "[No tool output]" {
		t.Fatalf("nil case failed")
	}
	if AsString("foo") != "foo" {
		t.Fatalf("string case failed")
	}
	if AsString([]byte("bar")) != "bar" {
		t.Fatalf("bytes case failed")
	}
	if AsString(5) != "5" {
		t.Fatalf("int case failed")
	}
}
