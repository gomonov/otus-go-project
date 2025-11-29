BIN := "./bin/application"
CLI_BIN := "./bin/cli"
DOCKER_COMPOSE_FILE := "./deployments/docker-compose.yaml"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.PHONY: build build-cli run run-local run-cli cli stop logs version install-lint-deps lint test clean help

run:
	docker compose -f $(DOCKER_COMPOSE_FILE) up --build -d

stop:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

cli:
	docker compose -f $(DOCKER_COMPOSE_FILE) run --rm cli -url http://application:8080 $(filter-out $@,$(MAKECMDGOALS))

logs:
	docker logs bf-application -f

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

build-cli:
	go build -v -o $(CLI_BIN) ./cmd/cli

run-local: build
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
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans

help:
	@echo "Makefile для Anti-Brute Force Service"
	@echo ""
	@echo "Доступные команды:"
	@echo ""
	@echo "  run              - Запуск полного сервиса через Docker Compose (основная команда)"
	@echo "  stop             - Остановка Docker сервисов"
	@echo "  cli              - Запуск CLI команды в Docker (без передачи параметров вывод help)"
	@echo "  logs             - Просмотр логов"
	@echo "  run-local        - Сборка и запуск основного приложения локально"
	@echo "  run-cli          - Сборка и запуск CLI инструмента локально с аргументами"
	@echo "  build            - Сборка основного приложения"
	@echo "  build-cli        - Сборка CLI инструмента"
	@echo "  version          - Показать версию приложения"
	@echo "  lint             - Запуск линтера"
	@echo "  test             - Запуск тестов"
	@echo "  clean            - Удаление артефактов сборки и Docker ресурсов"
	@echo "  help             - Показать эту справку"
	@echo ""
	@echo "Примеры использования:"
	@echo "  make run                              # Запуск полного сервиса (Docker)"
	@echo "  make logs                             # Просмотр логов (можно запустить в другом окне терминала, чтобы видеть логи запросов к сервису)"
	@echo "  make cli                              # Справка по командам"
	@echo "  make cli blacklist add 192.168.1.0/24 # Добавить подсеть в чёрный список"
	@echo "  make cli blacklist list               # Вывести чёрный список"
	@echo "  make cli -- reset --login user1       # Очистить бакеты для логина"
	@echo "  make stop                             # Остановка сервисов"

%:
	@:

.DEFAULT_GOAL := help