# Makefile for Go project
.PHONY: all build test run debug format

run:
	LOG_LEVEL=ERROR go run cmd/*.go

debug:
	LOG_LEVEL=DEBUG go run cmd/*.go

test:
	go test ./...

format:
	go fmt ./...
