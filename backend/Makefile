.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: build
build: ## Build the application
	go build -o bin/api cmd/api/main.go

.PHONY: run
run: ## Run the application
	go run cmd/api/main.go

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t tg-analysis-tool:latest -f ../infra/docker/Dockerfile.api .

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	go run cmd/migrate/main.go up

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	go run cmd/migrate/main.go down

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create name=add_users_table)
	go run cmd/migrate/main.go create $(name)

.PHONY: dev
dev: ## Run in development mode with hot reload
	air

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest