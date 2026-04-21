.PHONY: build run test test-all lint install-lint

build:
	go build -race -o incident-assistant ./cmd/incident-assistant/

run:
	go run ./cmd/incident-assistant/ --file docs/examples/inc102_db_reporting.txt

test:
	go test -race -short ./internal/

test-all:
	go test -race ./internal/

lint:
	golangci-lint run

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
