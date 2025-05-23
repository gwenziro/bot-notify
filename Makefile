.PHONY: build run clean dev setup test

# Application name
APP_NAME=whatsapp-bot

# Build the application
build:
	go build -o $(APP_NAME) ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Clean up build artifacts
clean:
	rm -f $(APP_NAME)
	rm -rf tmp

# Development mode with hot-reloading
dev:
	air

# Setup development environment
setup:
	go run ./scripts/setup_dev.go

# Install project dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test -v ./...
