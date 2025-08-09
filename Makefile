# Makefile for pingy-http library

.PHONY: help test test-unit test-integration test-race test-goroutines test-coverage test-bench test-memory test-all clean deps

# Default target
help:
	@echo "Available commands:"
	@echo "  deps            - Download dependencies"
	@echo "  test-unit       - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-race       - Run race condition tests"
	@echo "  test-goroutines - Run goroutine leak tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  test-bench      - Run benchmark tests"
	@echo "  test-memory     - Run memory usage tests"
	@echo "  test-all        - Run all tests"
	@echo "  clean          - Clean up test artifacts"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v -run "^Test(New|Set|HTTP|Error|Is)" .

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -run "^Test(TLS|Context|Retry)" .

# Run race condition tests
test-race:
	@echo "Running race condition tests..."
	go test -v -race -run "^Test(Concurrent|Race)" .

# Run goroutine leak tests
test-goroutines:
	@echo "Running goroutine leak tests..."
	go test -v -run "^Test.*Leak|^TestNoGoroutine" .

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic .
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem -run=^Benchmark .

# Run memory usage tests
test-memory:
	@echo "Running memory usage tests..."
	go test -v -run "^Test(Memory|HighVolume)" .

# Run all tests
test-all: deps test-unit test-integration test-race test-goroutines test-coverage test-bench test-memory
	@echo "All tests completed successfully!"

# Short test run (excludes long-running tests)
test-short:
	@echo "Running short tests..."
	go test -v -short .

# Clean up test artifacts
clean:
	@echo "Cleaning up test artifacts..."
	rm -f coverage.out coverage.html
	go clean -testcache

# Run tests in verbose mode
test-verbose:
	@echo "Running all tests in verbose mode..."
	go test -v .

# Run tests with specific timeout
test-timeout:
	@echo "Running tests with 5 minute timeout..."
	go test -v -timeout=5m .

# Production readiness test suite
test-production: deps test-coverage test-race test-goroutines test-bench
	@echo "Production readiness tests completed!"
	@echo "Coverage report: coverage.html"

# CI/CD test suite (for automated testing)
test-ci: deps
	@echo "Running CI/CD test suite..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic .
	go tool cover -func=coverage.out 