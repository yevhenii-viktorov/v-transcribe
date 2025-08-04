.PHONY: help build lint test local clean

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the Go API binary
	@echo "Building Go API..."
	cd api && go mod tidy && go build -o ../bin/api main.go
	@echo "✓ Built binary: bin/api"

lint: ## Run golangci-lint on the Go code
	@echo "Running golangci-lint..."
	cd api && golangci-lint run
	@echo "✓ Linting complete"

test: ## Run unit tests
	@echo "Running unit tests..."
	cd api && go test -v ./...
	@echo "✓ Tests complete"

local: ## Start local development environment
	@echo "Starting local development environment..."
	./scripts/test_locally.sh

clean: ## Clean up containers, volumes, and build artifacts
	@echo "Cleaning up..."
	docker-compose -f docker-compose.local.yml down -v 2>/dev/null || true
	rm -rf bin/
	rm -rf transcripts/
	rm -rf whisper-models/
	@echo "✓ Cleanup complete"