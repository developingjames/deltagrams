# Deltagram Makefile
# Standard Go project build configuration

# Variables
BINARY_NAME=deltagram
MAIN_PACKAGE=./cmd/deltagram
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)"

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(shell go env GOEXE) $(MAIN_PACKAGE)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)$(shell go env GOEXE)"

# Cross-platform builds
.PHONY: build-all
build-all: build-linux build-darwin build-windows build-linux-arm build-darwin-arm

.PHONY: build-linux
build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

.PHONY: build-linux-arm
build-linux-arm:
	@echo "Building for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

.PHONY: build-darwin-arm
build-darwin-arm:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Development targets
.PHONY: dev
dev: build
	@echo "Development build complete"

.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PACKAGE)

# Testing
.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -v ./pkg/... ./internal/...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/...

.PHONY: test-all
test-all: test test-integration
	@echo "All tests completed"

.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	@go test -race -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -v ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
.PHONY: lint
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

.PHONY: mod-tidy
mod-tidy:
	@echo "Tidying go.mod..."
	@go mod tidy

# Cleanup
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all artifacts..."
	@go clean -cache
	@go clean -testcache

# Release preparation
.PHONY: release-prep
release-prep: clean fmt vet mod-tidy test-all build-all
	@echo "Release preparation complete"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          Build for current platform"
	@echo "  build-all      Build for all supported platforms"
	@echo "  build-linux    Build for Linux (amd64)"
	@echo "  build-linux-arm Build for Linux (arm64)"
	@echo "  build-darwin   Build for macOS (amd64)"
	@echo "  build-darwin-arm Build for macOS (arm64)"
	@echo "  build-windows  Build for Windows (amd64)"
	@echo "  dev            Development build"
	@echo "  install        Install to GOPATH/bin"
	@echo "  test           Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-all       Run all tests"
	@echo "  test-race      Run tests with race detection"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  lint           Run linter"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  mod-tidy       Tidy go.mod"
	@echo "  clean          Clean build artifacts"
	@echo "  clean-all      Clean all artifacts including caches"
	@echo "  release-prep   Prepare for release (full pipeline)"
	@echo "  help           Show this help"

# Default help when no target specified
.DEFAULT_GOAL := help