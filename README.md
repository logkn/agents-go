# Agents Go

Agents Go is a small framework for building and running LLM powered agents in Go. It provides
utilities for streaming responses, running tools and handing off between multiple agents. A
simple terminal user interface is included for interactive use.

## Getting Started

1. **Install dependencies**
   ```bash
   go mod download
   ```
2. **Run the example TUI**
   ```bash
   go run ./cmd
   ```
3. **Explore the API**
   The `pkg` package exposes helper functions to run agents and to wrap an agent as a tool. See
   `examples/handoff_demo.go` for a more advanced scenario demonstrating agent handoffs.

## Developing

Run the test suite with:

```bash
make test
```

This will execute all unit tests under the `internal` packages.


