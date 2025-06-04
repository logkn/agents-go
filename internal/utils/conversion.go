package utils

import (
	"iter"
	"slices"
)

// Map applies the provided function to each element in the sequence and
// returns a new sequence of the results.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for a := range seq {
			if !yield(f(a)) {
				return
			}
		}
	}
}

// MapSlice converts a slice of T to a slice of U using the supplied mapper
// function.
func MapSlice[T, U any](seq []T, f func(T) U) []U {
	seqT := slices.Values(seq)
	seqU := Map(iter.Seq[T](seqT), f)

	sliceU := []U{}
	for u := range seqU {
		sliceU = append(sliceU, u)
	}
	return sliceU
}
