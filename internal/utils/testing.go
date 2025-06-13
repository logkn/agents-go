package utils

import "time"

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
