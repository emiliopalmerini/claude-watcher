.PHONY: all fmt vet build run test clean generate sqlc test-prompter dashboard

all: generate fmt vet test build

generate: sqlc

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

sqlc:
	sqlc generate

build: vet sqlc
	go build -o claude-watcher ./cmd

run: build
	./claude-watcher

test: vet
	go test -v ./...

clean:
	rm -f claude-watcher test-prompter dashboard
	go clean ./...

test-prompter: vet
	go build -o test-prompter ./cmd/test-prompter && ./test-prompter

dashboard: vet
	go build -o dashboard ./cmd/dashboard && ./dashboard
