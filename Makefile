.PHONY: help install deps build run test clean lint swagger docker-build docker-run

help:
	@echo "FixFX Backend - Available Commands"
	@echo "===================================="
	@echo "  make install      - Install dependencies"
	@echo "  make deps         - Download dependencies"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run with hot-reload (requires air)"
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make lint         - Run linters"
	@echo "  make swagger      - Generate Swagger documentation"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo ""

install: deps
	@echo "Installing FixFX backend..."
	@go build -o fixfx-backend main.go
	@echo "✓ Installation complete"

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

build:
	@echo "Building FixFX backend..."
	@mkdir -p dist
	@go build -o dist/fixfx-backend main.go
	@echo "✓ Build complete: dist/fixfx-backend"

run:
	@echo "Starting FixFX backend..."
	@go run main.go

dev:
	@echo "Starting FixFX backend (dev mode with hot-reload)..."
	@which air > /dev/null || go install github.com/air-verse/air@latest
	@air

test:
	@echo "Running tests..."
	@go test -v ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run

swagger:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@swag init
	@echo "✓ Swagger docs generated"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f fixfx-backend
	@rm -rf dist/
	@rm -rf docs/swagger.*
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✓ Clean complete"

docker-build:
	@echo "Building Docker image..."
	@docker build -t fixfx-backend:latest .
	@echo "✓ Docker image built"

docker-run:
	@echo "Running Docker container..."
	@docker run -p 3001:3001 \
		-e PORT=3001 \
		-e ENVIRONMENT=development \
		fixfx-backend:latest
	@echo "✓ Container running on http://localhost:3001"

# Development targets
.PHONY: install-tools format

install-tools:
	@echo "Installing development tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest
	@echo "✓ Tools installed"

format:
	@echo "Formatting code..."
	@gofmt -w .
	@echo "✓ Code formatted"
