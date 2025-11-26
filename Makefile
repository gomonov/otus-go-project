BIN := "./bin/application"
CLI_BIN := "./bin/cli"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.PHONY: build build-cli run run-cli version install-lint-deps lint test clean help

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

build-cli:
	go build -v -o $(CLI_BIN) ./cmd/cli

run: build
	$(BIN) -config ./configs/config.toml

run-cli: build-cli
	$(CLI_BIN) -url http://localhost:8080 $(filter-out $@,$(MAKECMDGOALS))

version: build
	$(BIN) version

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	golangci-lint run ./...

lint-fix: install-lint-deps
	golangci-lint run --fix ./...

test:
	go test -race -count 100 ./internal/ratelimit

clean:
	rm -rf ./bin

help:
	@echo "Anti-Brute Force Service Makefile"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "  build           - Build main application"
	@echo "  build-cli       - Build CLI administration tool"
	@echo "  run             - Build and run main application"
	@echo "  run-cli         - Build and run CLI tool"
	@echo "  version         - Show application version"
	@echo "  lint            - Run linter"
	@echo "  test            - Run tests"
	@echo "  clean           - Remove build artifacts"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run        # Start the main service"
	@echo "  make run-cli    # Start CLI tool (in another terminal)"
	@echo "  make lint       # Check code quality"
	@echo "  make test       # Run tests"
	@echo ""
	@echo "CLI Usage after run-cli:"
	@echo "  ./bin/cli blacklist list"
	@echo "  ./bin/cli whitelist add 10.0.0.0/8"
	@echo "  ./bin/cli reset --login user1"
	@echo "  ./bin/cli auth user1 pass123 192.168.1.100"

# Default target
.DEFAULT_GOAL := help