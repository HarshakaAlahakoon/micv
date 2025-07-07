# MiCV - Enterprise Job Application Tool

A production-ready, enterprise-grade command-line tool for submitting job applications to MiTimes careers portal. Built with software engineering practices including dependency injection, functional programming patterns, circuit breakers, and comprehensive observability.

## Architecture Overview

This application follows clean architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                      │
│                        (main.go)                            │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                        │
│                      (services.go)                          │
├─────────────────────────────────────────────────────────────┤
│                     Domain Layer                            │
│                 (types.go, functional.go)                   │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                       │
│            (client.go, config.go, errors.go)                │
└─────────────────────────────────────────────────────────────┘
```

## Key Features

### 🏗️ **Enterprise Architecture**
- **Dependency Injection**: Fully testable with mockable dependencies
- **Service Layer Pattern**: Clean separation of business logic
- **Circuit Breaker Pattern**: Resilient network calls with automatic recovery
- **Retry Mechanism**: Exponential backoff for transient failures
- **Functional Programming**: Immutable data structures and pure functions

### 🔍 **Observability**
- **Structured Logging**: JSON-formatted logs with contextual information
- **Error Categorization**: Detailed error codes and context
- **Performance Metrics**: Built-in benchmarking and profiling support
- **Distributed Tracing**: Request correlation across components

### 🧪 **Testing Excellence**
- **Comprehensive Test Suite**: Unit, integration, and benchmark tests
- **Mocking Framework**: Fully mockable dependencies
- **Property-Based Testing**: Functional validation with edge cases
- **Performance Testing**: Built-in benchmarks for critical paths

### 🚀 **Production Ready**
- **Configuration Management**: Environment variables, config files, CLI flags
- **Graceful Degradation**: Circuit breakers and fallback mechanisms
- **Resource Management**: Proper cleanup and timeout handling
- **Version Management**: Build-time version injection

## Installation

### Prerequisites

- Go 1.21+ (with generics support)
- Git (for version information)
- Make (for build automation)

### Quick Start

```bash
# Clone and build
git clone <repository-url>
cd micv
make build

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

## Command-Line Options

The application supports various command-line flags for flexible configuration and operation:

### Basic Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--verbose` | boolean | Enable verbose logging (debug level) | `--verbose` |
| `--help` | boolean | Show help message and usage information | `--help` |
| `--version` | boolean | Display version, build time, and commit hash | `--version` |

### Configuration Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--config` | string | Path to configuration file | `--config config.json` |
| `--secret-url` | string | URL for the secret endpoint | `--secret-url https://custom.com/secret` |
| `--app-url` | string | URL for the application endpoint | `--app-url https://custom.com/apply` |
| `--timeout` | int | Request timeout in seconds | `--timeout 60` |

### Data Management Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--data` | string | Path to JSON file containing application data | `--data application.json` |
| `--generate-data-json` | boolean | Generate sample data.json file and exit | `--generate-data-json` |
| `--generate-config-json` | boolean | Generate sample config.json file and exit | `--generate-config-json` |

### Usage Examples

#### Verbose Mode
```bash
# Enable detailed debug logging
./micv --verbose "John Doe" "john@example.com" "Software Engineer"
```

#### Generate Sample Data File
```bash
# Generate a sample data.json file with realistic examples
./micv --generate-data-json
# This creates a 'data.json' file you can edit and use with:
./micv --data data.json
```

#### Generate Sample Configuration File
```bash
# Generate a sample config.json file with default settings
./micv --generate-config-json
# This creates a 'config.json' file you can edit and use with:
./micv --config config.json
```

#### Generate Both Configuration and Data Files
```bash
# Generate both config.json and data.json files at once
./micv --generate-config-json --generate-data-json
# This creates both files which you can then use together:
./micv --config config.json --data data.json
```

#### Combined Flags
```bash
# Use multiple flags together
./micv --verbose --timeout 60 --config custom-config.json "John Doe" "john@example.com" "Software Engineer"

# Use custom endpoints with verbose logging
./micv --verbose --secret-url https://staging.com/secret --app-url https://staging.com/apply "John Doe" "john@example.com" "Software Engineer"
```

#### Configuration File with Data File
```bash
# Use both configuration and data files
./micv --config production.json --data application.json --verbose
```

### Understanding Verbose Mode

When `--verbose` is enabled, the application provides detailed debug information including:

- HTTP request/response details
- Configuration loading steps
- Authentication token retrieval process
- Network timeout and retry information
- Detailed error context and stack traces
- Performance metrics and timing information

Example verbose output:
```
🔧 Debug: Loading configuration from file: config.json
🔧 Debug: Overriding secret URL from command line
🔧 Debug: Timeout set to 45 seconds
🌐 Debug: Making request to secret endpoint
🔧 Debug: Auth token received (length: 64 characters)
🌐 Debug: Submitting application with token
✅ Application submitted successfully
```

### File Generation Features

The application provides convenient commands to generate sample configuration and data files:

#### Configuration File Generation
```bash
./micv --generate-config-json
```
This generates a `config.json` file with default settings:
```json
{
  "secret_url": "https://au.mitimes.com/careers/apply/secret",
  "application_url": "https://au.mitimes.com/careers/apply", 
  "timeout_seconds": 30
}
```

#### Data File Generation
```bash
./micv --generate-data-json
```
This generates a `data.json` file with sample application data including:
- Personal information (name, email, job title)
- Professional attributes and skills
- Work experience and previous roles
- Key projects and achievements
- Technical skills and programming languages
- Availability and preferences

Both generated files can be edited with your actual information and used with the `--config` and `--data` flags respectively.

## Configuration

### Configuration Hierarchy (highest to lowest priority)

1. **Command Line Flags** (highest priority)
2. **Environment Variables**
3. **Configuration File**
4. **Default Values** (lowest priority)

### Environment Variables

```bash
export MICV_SECRET_URL="https://au.mitimes.com/careers/apply/secret"
export MICV_APPLICATION_URL="https://au.mitimes.com/careers/apply"
export MICV_TIMEOUT="30"
```

### Configuration File Example

```json
{
  "secret_url": "https://au.mitimes.com/careers/apply/secret",
  "application_url": "https://au.mitimes.com/careers/apply",
  "timeout_seconds": 30
}
```

## Advanced Usage

### Functional Programming Patterns

The application demonstrates functional programming concepts:

```go
// Result type for better error handling
result := validateApplicationDataFunctional(appData)
if result.IsError() {
    return result.Error
}

// Pipeline pattern for operations
pipeline := NewPipeline[ApplicationData]().
    Add(validateData).
    Add(enrichData).
    Add(submitData)

result := pipeline.Execute(appData)
```

### Circuit Breaker Configuration

```go
// Customize circuit breaker behavior
circuitBreaker := NewCircuitBreaker(
    maxFailures: 5,
    resetTimeout: 60*time.Second,
    logger: logger,
)
```

### Retry Configuration

```go
// Custom retry strategy
retryConfig := RetryConfig{
    MaxAttempts:  5,
    InitialDelay: 2 * time.Second,
    MaxDelay:     60 * time.Second,
    Multiplier:   2.0,
}
```

## Development

### Development Environments

This project supports **three different development approaches**:

#### 1. 🏠 **Local Development** (Traditional)
```bash
# Direct Go development
go run main.go "John Doe" "john@example.com" "Software Engineer"
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
make test
make test-coverage
make benchmark

# Container testing
docker-compose --profile test up micv-test

# Integration testing with mocks
docker-compose up  # Automatically includes mock server
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security scan
make security

# Generate documentation
make docs
```

### Build Options

```bash
# Development build
make build

# Production build with optimizations
make build-prod

# Cross-platform builds
make build-all

# Docker builds
make docker-build        # Production
make docker-build-dev    # Development
```

## API Reference

### Core Services

#### ApplicationService
```go
type ApplicationService struct {
    deps Dependencies
}

func (s *ApplicationService) SubmitApplication(ctx context.Context, appData ApplicationData) error
```

#### AuthTokenService
```go
type AuthTokenService struct {
    deps Dependencies
}

func (s *AuthTokenService) GetToken(ctx context.Context) (string, error)
```

#### ConfigService
```go
type ConfigService struct {
    deps Dependencies
}

func (s *ConfigService) ValidateConfig() error
```

### Functional Utilities

#### Result Type
```go
type Result[T any] struct {
    Value T
    Error error
}

func (r Result[T]) Map(fn func(T) T) Result[T]
func (r Result[T]) FlatMap(fn func(T) Result[T]) Result[T]
func (r Result[T]) Filter(predicate func(T) bool, errorMsg string) Result[T]
```

#### Pipeline
```go
type Pipeline[T any] struct {
    operations []func(T) Result[T]
}

func (p *Pipeline[T]) Add(op func(T) Result[T]) *Pipeline[T]
func (p *Pipeline[T]) Execute(input T) Result[T]
```

## Performance Considerations

### Benchmarks

```bash
BenchmarkApplicationService-8     1000000    1000 ns/op    240 B/op    5 allocs/op
BenchmarkCircuitBreaker-8        10000000     100 ns/op     64 B/op    1 allocs/op
BenchmarkRetryMechanism-8         5000000     200 ns/op    128 B/op    2 allocs/op
```

### Memory Usage

The application is designed for minimal memory footprint:
- Zero-copy JSON parsing where possible
- Efficient string handling with builders
- Pooled HTTP clients for connection reuse
- Structured logging with minimal allocations

## Monitoring and Observability

### Structured Logging

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Application submission started",
  "name": "John Doe",
  "email": "john@example.com",
  "job_title": "Software Engineer",
  "operation": "submit_application",
  "trace_id": "abc123"
}
```

### Error Codes

| Code | Description | Action Required |
|------|-------------|----------------|
| `NETWORK_ERROR` | Network connectivity issues | Check network, retry |
| `VALIDATION_ERROR` | Input validation failed | Fix input data |
| `CONFIG_ERROR` | Configuration problems | Check configuration |
| `AUTH_ERROR` | Authentication failed | Check credentials |
| `TIMEOUT_ERROR` | Request timeout | Increase timeout or retry |

## Contributing

### Development Setup

```bash
# Install dependencies
go mod download

# Install development tools
make install-tools

# Run pre-commit hooks
make pre-commit
```

### Code Standards

- **Go Style Guide**: Follow effective Go practices
- **Functional Programming**: Prefer immutable data structures
- **Error Handling**: Use structured error types
- **Testing**: Maintain >90% test coverage
- **Documentation**: Document all public APIs

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- GitHub Issues: Create detailed bug reports
- Documentation: Check the wiki for additional examples
- Code Review: Submit PRs with comprehensive tests
