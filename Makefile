# Makefile for Go project
.PHONY: all build test run debug format check

build:
	go build ./...

run:
	GOEXPERIMENT=aliastypeparams LOG_LEVEL=ERROR go run cmd/*.go

debug:
	GOEXPERIMENT=aliastypeparams LOG_LEVEL=DEBUG go run cmd/*.go

test:
	go test ./...

format:
	go fmt ./...

check: format test
