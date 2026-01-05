.PHONY: test test-short test-coverage fmt lint build examples clean

# Run all tests
test:
	go test -v ./...

# Run tests without integration tests
test-short:
	go test -v -short ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

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

