BIN := "./bin/application"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

run: build
	$(BIN) -config ./configs/config.toml

version: build
	$(BIN) version