.PHONY: test test-short coverage fmt lint build examples clean version tag-version help

# Version management
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
VERSION_FILE = VERSION

# Run all tests
test:
	go test -v ./...

# Run tests without integration tests
test-short:
	go test -v -short ./...

# Run tests with coverage
coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build examples
examples:
	go build -o bin/nse_quote ./examples/nse_quote
	go build -o bin/mf_nav ./examples/mf_nav

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run example: NSE quote
run-nse:
	go run ./examples/nse_quote/main.go

# Run example: MF NAV
run-mf:
	go run ./examples/mf_nav/main.go

# Show current version
version:
	@echo $(VERSION)

# Create a new version tag (usage: make tag-version VERSION=v1.0.0)
tag-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make tag-version VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$$'; then \
		echo "Error: VERSION must follow semantic versioning (e.g., v1.0.0, v1.0.0-beta.1)"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag created. Push with: git push origin $(VERSION)"

# Help target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-short    - Run tests without integration tests"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  examples      - Build example binaries"
	@echo "  run-nse       - Run NSE quote example"
	@echo "  run-mf        - Run MF NAV example"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  version       - Show current version"
	@echo "  tag-version   - Create a new version tag (VERSION=v1.0.0)"
	@echo "  help          - Show this help message"
