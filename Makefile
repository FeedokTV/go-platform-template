# Load .env
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run test build

# Export variables for build infoin internal/platform/buildoinfo
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X 'go-platform-template/internal/platform/buildinfo.Version=$(VERSION)' \
			-X 'go-platform-template/internal/platform/buildinfo.Commit=$(COMMIT)' \
			-X 'go-platform-template/internal/platform/buildinfo.Date=$(DATE)'

help: ## List available targets
	@echo "Go-platform template by Feedok"
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

run: ## Run application with .env loaded
	set -a && . ./.env 2>/dev/null || true && set +a && go run ./cmd

build: ## Build application in binary
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd

docker-build: ## Build the container image with version info
	docker build -f deploy/Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t go-platform-template:$(VERSION) .

test: ## Run tests
	go test ./... -short

test-integration: ## Run integration tests
	go test ./... -v

lint: ## Run golangci-lint over the code
	golangci-lint run

up: ## Up the docker-compose with containers
	docker compose up -d

down: ## Down the docker-compose with containers
	docker compose down

migrate-up: ## Up migrations in database
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) up

migrate-down: ## Down migrations in database (for dev only)
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE) down
