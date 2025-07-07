# MiCV - Enterprise Job Application Tool

A production-ready, enterprise-grade command-line tool for submitting job applications to MiTimes careers portal. Built with software engineering practices including dependency injection, functional programming patterns, circuit breakers, and comprehensive observability.

## Table of Contents

- [Documentation](#documentation)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
- [Development](#development)
  - [Development Environments](#development-environments)
    - [1. 🏠 Local Development (Traditional)](#1--local-development-traditional)
    - [2. 🐳 Docker Compose (Integration Testing)](#2--docker-compose-integration-testing)
    - [3. 📦 VS Code DevContainer (Recommended)](#3--vs-code-devcontainer-recommended)
  - [Docker Override File Explained](#docker-override-file-explained)
  - [Running Tests](#running-tests)
  - [Code Quality](#code-quality)
  - [Build Options](#build-options)
- [Contributing](#contributing)
  - [Development Setup](#development-setup)
  - [Code Standards](#code-standards)

## Documentation

For detailed technical documentation, API references, and advanced usage examples, see [DOCUMENTATION.md](./DOCUMENTATION.md).

## Installation

### Prerequisites

- Go 1.21+ (with generics support)
- Git (for version information)

### Quick Start

```bash
# Clone and build
git clone <repository-url>
cd micv

# Build directly with Go
go build -o micv .

# Run with default configuration
./micv "John Doe" "john@example.com" "Software Engineer"

# Run with configuration file
./micv --data application.json

# Generate sample files for easy setup
./micv --generate-config-json  # Creates config.json
./micv --generate-data-json   # Creates data.json

# Run with custom endpoints
./micv --secret-url https://custom.com/secret \
       --app-url https://custom.com/apply \
       "John Doe" "john@example.com" "Software Engineer"
```

## Development

### Development Environments

This project supports **three different development approaches**:

#### 1. 🏠 **Local Development** (Traditional)
```bash
# Direct Go development
go run . "John Doe" "john@example.com" "Software Engineer"
go test -v ./...
```

#### 2. 🐳 **Docker Compose** (Integration Testing)
```bash
# Development with mock server
docker-compose up micv-dev     # Includes mock API server

# Testing with coverage
docker-compose --profile test up micv-test
```

#### 3. 📦 **VS Code DevContainer** (Recommended)
```bash
# Open in VS Code and use "Reopen in Container"
# Provides full Go environment with debugging support
```

See [CONTAINER_DEVELOPMENT.md](./CONTAINER_DEVELOPMENT.md) for detailed container workflows.

### Docker Override File Explained

The `docker-compose.override.yml` automatically provides:

| Feature | Purpose |
|---------|---------|
| **Mock Server** | WireMock server on `localhost:8081` for safe testing |
| **Environment Override** | Points URLs to mock server instead of production |
| **SSH Key Mounting** | Git operations inside containers |
| **Interactive TTY** | Better debugging experience |

**When to use:**
- **Local CLI**: Daily usage and real job applications  
- **Docker Compose**: Integration testing with mock APIs
- **DevContainer**: Development and debugging in VS Code

### Running Tests

```bash
# Local testing
go test -v ./...
go test -v -race -coverprofile=coverage.out ./...
go test -bench=. -benchmem ./...

# Container testing
docker-compose --profile test up micv-test

# Integration testing with mocks
docker-compose up  # Automatically includes mock server
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Run security scan (requires gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
gosec ./...

# Generate documentation
godoc -http=:6060  # Then visit http://localhost:6060
```

### Build Options

```bash
# Development build
go build -o micv .

# Production build with optimizations
CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o micv .

# Cross-platform builds
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o micv-linux-amd64 .
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o micv-darwin-amd64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o micv-windows-amd64.exe .

# Docker builds
docker build -t micv:prod .        # Production
docker build -f Dockerfile.dev -t micv:dev .    # Development
```

## Contributing

### Development Setup

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run pre-commit checks
go fmt ./... && golangci-lint run && gosec ./... && go test -v ./...
```

### Code Standards

- **Go Style Guide**: Follow effective Go practices
- **Functional Programming**: Prefer immutable data structures
- **Error Handling**: Use structured error types
- **Testing**: Maintain >90% test coverage