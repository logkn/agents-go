package main

import (
	"flag"
)

// main is the program entry point.
func main() {
	query := flag.String("q", "", "Send a query message upfront instead of starting interactive chat")
	flag.Parse()

	if *query != "" {
		RunSingleQuery(*query)
	} else {
		RunChat()
	}
}
