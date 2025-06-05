# Makefile for Go project
.PHONY: all build test run debug

run:
	go run cmd/*.go

debug:
	LOG_LEVEL=DEBUG go run cmd/*.go

test:
	go test ./...
