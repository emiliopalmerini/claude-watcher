.PHONY: all fmt vet build run test clean generate sqlc templ

all: generate fmt vet test build

generate: sqlc templ

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

sqlc:
	sqlc generate

templ:
	templ generate

build: vet sqlc templ
	go build -o claude-watcher ./cmd

run: build
	./claude-watcher

test: vet
	go test -v ./...

clean:
	rm -f claude-watcher
	go clean ./...
