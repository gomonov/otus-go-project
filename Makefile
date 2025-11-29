BIN := "./bin/application"
CLI_BIN := "./bin/cli"
DOCKER_COMPOSE_FILE := "./deployments/docker-compose.yaml"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.PHONY: build build-cli run run-local run-cli cli stop logs run-detached version install-lint-deps lint test clean help

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

build-cli:
	go build -v -o $(CLI_BIN) ./cmd/cli

run:
	docker compose -f $(DOCKER_COMPOSE_FILE) up --build

run-local: build
	$(BIN) -config ./configs/config.toml

run-cli: build-cli
	$(CLI_BIN) -url http://localhost:8080 $(filter-out $@,$(MAKECMDGOALS))

cli:
	docker compose -f $(DOCKER_COMPOSE_FILE) run --rm cli -url http://application:8080 $(filter-out $@,$(MAKECMDGOALS))

stop:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

logs:
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

run-detached:
	docker compose -f $(DOCKER_COMPOSE_FILE) up --build -d

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
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans

help:
	@echo "Makefile для Anti-Brute Force Service"
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  run              - Запуск полного сервиса через Docker Compose (основная команда)"
	@echo "  run-local        - Сборка и запуск основного приложения локально"
	@echo "  run-cli          - Сборка и запуск CLI инструмента локально с аргументами"
	@echo "  cli              - Запуск CLI команды в Docker"
	@echo "  stop             - Остановка Docker сервисов"
	@echo "  logs             - Просмотр логов"
	@echo "  run-detached     - Запуск в фоновом режиме"
	@echo "  build            - Сборка основного приложения"
	@echo "  build-cli        - Сборка CLI инструмента"
	@echo "  version          - Показать версию приложения"
	@echo "  lint             - Запуск линтера"
	@echo "  test             - Запуск тестов"
	@echo "  clean            - Удаление артефактов сборки и Docker ресурсов"
	@echo "  help             - Показать эту справку"
	@echo ""
	@echo "Примеры использования:"
	@echo "  make run                           # Запуск полного сервиса (Docker)"
	@echo "  make run-local                     # Локальный запуск для разработки"
	@echo "  make run-cli blacklist list        # Использование CLI локально"
	@echo "  make cli blacklist list            # Использование CLI в Docker"
	@echo "  make cli -- reset --login user1    # Использование CLI с флагами"
	@echo "  make stop                          # Остановка сервисов"
	@echo "  make logs                          # Просмотр логов"

%:
	@:

.DEFAULT_GOAL := help