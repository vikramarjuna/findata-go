.PHONY: test test-short test-parallel test-pkg coverage fmt lint lint-config build examples clean version tag-version help

# Version management
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
VERSION_FILE = VERSION

# Run all tests
test:
	go test -v -race -parallel 4 ./...

# Run tests without integration tests
test-short:
	go test -v -short -race -parallel 4 ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests in parallel with verbose output
test-parallel:
	go test -v -race -parallel 8 -timeout 10m ./...

# Run a specific package's tests
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Error: PKG is required. Usage: make test-pkg PKG=./provider/nse"; \
		exit 1; \
	fi
	go test -v -race -parallel 4 $(PKG)

# Format code
fmt:
	go fmt ./...

# Verify linter config
lint-config:
	golangci-lint config verify

# Lint code
lint: lint-config
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
	@echo "  test          - Run all tests with race detector and parallelism"
	@echo "  test-short    - Run tests without integration tests"
	@echo "  test-parallel - Run tests with higher parallelism (8 workers)"
	@echo "  test-pkg      - Run tests for a specific package (PKG=./path/to/pkg)"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  fmt           - Format code"
	@echo "  lint-config   - Verify golangci-lint configuration"
	@echo "  lint          - Verify config and run linter"
	@echo "  examples      - Build example binaries"
	@echo "  run-nse       - Run NSE quote example"
	@echo "  run-mf        - Run MF NAV example"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  version       - Show current version"
	@echo "  tag-version   - Create a new version tag (VERSION=v1.0.0)"
	@echo "  help          - Show this help message"
