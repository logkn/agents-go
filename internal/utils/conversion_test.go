package utils

import (
	"iter"
	"testing"
)

func TestMap(t *testing.T) {
	seq := iter.Seq[int](func(yield func(int) bool) {
		for i := 1; i <= 3; i++ {
			if !yield(i) {
				return
			}
		}
	})

	mapped := Map(seq, func(i int) int { return i * 2 })

	expected := []int{2, 4, 6}
	idx := 0
	for v := range mapped {
		if v != expected[idx] {
			t.Fatalf("unexpected value %d at index %d", v, idx)
		}
		idx++
	}
	if idx != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), idx)
	}
}

func TestMapSlice(t *testing.T) {
	in := []int{1, 2, 3}
	out := MapSlice(in, func(i int) int { return i + 1 })
	expected := []int{2, 3, 4}
	if len(out) != len(expected) {
		t.Fatalf("expected len %d got %d", len(expected), len(out))
	}
	for i, v := range expected {
		if out[i] != v {
			t.Fatalf("expected %d at %d got %d", v, i, out[i])
		}
	}
}
