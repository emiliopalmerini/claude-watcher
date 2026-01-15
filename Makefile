.PHONY: all fmt vet build run test clean generate sqlc session-tracker dashboard

all: generate fmt vet test build

generate: sqlc

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

sqlc:
	sqlc generate

build: vet sqlc dashboard session-tracker

dashboard:
	go build -o claude-watcher ./cmd/dashboard

session-tracker:
	go build -o session-tracker ./cmd/session-tracker

run: dashboard
	./claude-watcher

test: vet
	go test -v ./...

clean:
	rm -f claude-watcher session-tracker
	go clean ./...
