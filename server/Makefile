.PHONY: run build test lint tidy

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./..

lint:
	golangci-lint run

tidy:
	go mod tidy