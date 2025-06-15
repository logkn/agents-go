package utils

import "time"

// MockStream emits the provided string one character at a time with the given
// delay between characters. It is useful for testing streaming behaviour.
func MockStream(text string, delay time.Duration) chan string {
	output := make(chan string)

	go func() {
		defer close(output)

		for _, char := range text {
			time.Sleep(delay)
			output <- string(char)
		}
	}()

	return output
}
