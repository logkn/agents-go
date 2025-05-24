package utils

import "time"

func ReadWithTimeout[T any](ch <-chan T, timeout int) (T, bool) {
	var result T
	select {
	case res := <-ch:
		return res, true
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return result, false
	}
}
