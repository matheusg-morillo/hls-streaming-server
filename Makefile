.PHONY: help lint lint-fix build run clean install-lint

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: ## Run golangci-lint
	go tool -modfile=golangci-lint.mod golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	go tool -modfile=golangci-lint.mod golangci-lint run --fix ./...

build: ## Build the binary
	go build -o bin/hls-server src/main.go

run: ## Run the server
	go run src/main.go

clean: ## Clean build artifacts
	rm -rf bin/

install-lint: ## Install/update golangci-lint
	go mod init -modfile=golangci-lint.mod matheusflix/hls-streaming-server/golangci-lint || true
	go get -tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
