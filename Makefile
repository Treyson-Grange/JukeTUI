.PHONY: build run clean help dev lint test

# Build the JukeTUI binary
build:
	go build -o JukeTUI ./src

# Run JukeTUI directly without building
run:
	go run ./src

# Clean up build artifacts and logs
clean:
	rm -f JukeTUI JukeTUI-new
	rm -f data/logs/*.log

# Install dependencies
deps:
	go mod tidy
	go mod download

# Development mode - run with logging enabled
dev:
	DEVELOPMENT=true go run ./src

# Format code
fmt:
	gofmt -w ./src

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./src

# Display help
help:
	@echo "Available targets:"
	@echo "  build   - Build the JukeTUI binary"
	@echo "  run     - Run JukeTUI directly"
	@echo "  dev     - Run with DEVELOPMENT mode enabled (creates log files)"
	@echo "  clean   - Remove build artifacts and log files"
	@echo "  deps    - Update dependencies"
	@echo "  fmt     - Format code with gofmt"
	@echo "  lint    - Run golangci-lint"
	@echo "  help    - Show this help message"

.DEFAULT_GOAL := help
