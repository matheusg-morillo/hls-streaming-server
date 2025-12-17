.PHONY: help lint lint-fix build run clean install-lint

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: ## Run golangci-lint
	go tool -modfile=tools.mod golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	go tool -modfile=tools.mod golangci-lint run --fix ./...

fmt: ## Format code with gofmt and goimports
	go tool -modfile=tools.mod gofmt -w -s .
	go tool -modfile=tools.mod goimports -w .

fmt-check: ## Check if code needs formatting
	go tool -modfile=tools.mod gofmt -l -s .

build: ## Build the binary
	go build -o bin/hls-server src/main.go

run: ## Run the server
	go run src/main.go

clean: ## Clean build artifacts
	rm -rf bin/

install-lint: ## Install/update development tools
	go mod init -modfile=tools.mod matflix/hls-streaming-server/tools || true
	go get -tool -modfile=tools.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	go get -tool -modfile=tools.mod golang.org/x/tools/cmd/goimports@latest
	go mod tidy -modfile=tools.mod
