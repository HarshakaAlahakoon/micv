# Enhanced Makefile for Production-Ready Build

.PHONY: build build-prod test test-coverage benchmark lint fmt security docs clean install-tools pre-commit docker-build release

# Build variables
APP_NAME := micv
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# Go variables
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GO_VERSION := $(shell go version | awk '{print $$3}')

# Directories
BUILD_DIR := build
DIST_DIR := dist
DOCS_DIR := docs

# Default target
all: build

# Development build
build:
	@echo "Building $(APP_NAME) for development..."
	go build $(LDFLAGS) -o $(APP_NAME) .
	@echo "Build complete: $(APP_NAME)"

# Production build with optimizations
build-prod:
	@echo "Building $(APP_NAME) for production..."
	CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS) -s -w" \
		-trimpath \
		-o $(APP_NAME) .
	@echo "Production build complete: $(APP_NAME)"

# Cross-platform builds
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS) -s -w" \
		-trimpath \
		-o $(DIST_DIR)/$(APP_NAME)-linux-amd64 .

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS) -s -w" \
		-trimpath \
		-o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 .

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS) -s -w" \
		-trimpath \
		-o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe .

# Testing
test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Code quality
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

lint:
	@echo "Running linter..."
	golangci-lint run

security:
	@echo "Running security scan..."
	gosec ./...

# Documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	godoc -html . > $(DOCS_DIR)/api.html
	@echo "API documentation generated: $(DOCS_DIR)/api.html"

# Development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Pre-commit hooks
pre-commit: fmt lint security test

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Release
release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(DIST_DIR) $(APP_NAME)-linux-amd64
	tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(DIST_DIR) $(APP_NAME)-darwin-amd64
	zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(DOCS_DIR)
	rm -f $(APP_NAME) coverage.out coverage.html

# Install to system
install: build-prod
	@echo "Installing $(APP_NAME) to system..."
	sudo cp $(APP_NAME) /usr/local/bin/
	@echo "$(APP_NAME) installed to /usr/local/bin/"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(APP_NAME) from system..."
	sudo rm -f /usr/local/bin/$(APP_NAME)

# Show build info
info:
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit Hash: $(COMMIT_HASH)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build for development"
	@echo "  build-prod    - Build for production with optimizations"
	@echo "  build-all     - Build for all platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  benchmark     - Run benchmarks"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  security      - Run security scan"
	@echo "  docs          - Generate documentation"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install to system"
	@echo "  uninstall     - Uninstall from system"
	@echo "  docker-build  - Build Docker image"
	@echo "  release       - Create release archives"
	@echo "  info          - Show build information"
	@echo "  help          - Show this help"
