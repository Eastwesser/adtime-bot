.PHONY: help build run test migrate-up migrate-down migrate-status migrate-create docker-up docker-down docker-build

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

TARGET_MAX_CHAR_NUM=20

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Build the application
build:
	@go build -o bin/adtime-bot ./cmd/adtime/main.go

## Run the application locally
run:
	@go run ./cmd/adtime/main.go

## Run tests
test:
	@go test -v ./...

## Apply all migrations
migrate-up:
	@goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" up

## Revert the last migration
migrate-down:
	@goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" down

## Show migration status
migrate-status:
	@goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" status

## Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" create $${name// /_} sql

## Start all containers
docker-up:
	@docker-compose up -d

## Stop all containers
docker-down:
	@docker-compose down

## Build Docker images
docker-build:
	@docker-compose build