# Container Development Guide

This project supports **three complementary development approaches**, each optimized for different workflows:

## 🏗️ Development Approaches Overview

| Approach | Use Case | When to Use |
|----------|----------|-------------|
| **DevContainer** | Daily coding, debugging | Primary development workflow |
| **Docker Compose** | Integration testing | Testing with mock APIs |
| **Local Go** | CLI execution | Running the actual CLI tool |

## 🚀 Quick Start

### Option 1: VS Code Dev Container (Recommended for Development)
```bash
# 1. Install "Dev Containers" extension in VS Code
# 2. Open project: Ctrl+Shift+P → "Dev Containers: Reopen in Container"
# 3. Container builds automatically with full Go environment
```

**Features:**
- Full Go 1.24 environment with hot reload (Air)
- Integrated debugging with Delve (port 2345)
- All Go tools pre-installed (golangci-lint, goimports)
- Direct file editing with real-time compilation

### Option 2: Docker Compose (Integration Testing)
```bash
# Development with mock server (automatic with override file)
docker-compose up micv-dev

# Test integration with mock APIs
curl http://localhost:8081/secret  # Mock server endpoint

# Run comprehensive tests
docker-compose --profile test up micv-test
```

**Features:**
- **Mock Server**: WireMock on port 8081 with test data
- **Environment Override**: Automatically redirects to local mock APIs
- **Service Integration**: Full application stack testing

### Option 3: Local CLI Execution (Actual Usage)
```bash
# Build and run the CLI tool locally
go build -o micv .
./micv "John Doe" "john@example.com" "Software Engineer"

# Or run directly
go run main.go "John Doe" "john@example.com" "Software Engineer"

# With configuration file
./micv --data application.json

# With custom endpoints  
./micv --secret-url https://custom.com/secret "John Doe" "john@example.com" "Software Engineer"
```

## 🔧 Docker Override File Explanation

The `docker-compose.override.yml` **automatically merges** with `docker-compose.yml` when you run Docker Compose commands:

### What it does:
```yaml
# Overrides environment variables
environment:
  - MICV_SECRET_URL=http://localhost:8081/secret      # Points to mock
  - MICV_APPLICATION_URL=http://localhost:8081/apply  # Points to mock
  
# Adds mock server service
mock-server:
  image: wiremock/wiremock:latest
  ports: ["8081:8080"]
  volumes: ["./testdata:/home/wiremock/mappings:ro"]

# Development conveniences  
volumes:
  - ${HOME}/.ssh:/home/vscode/.ssh:ro  # SSH keys for git
stdin_open: true    # Interactive terminal
tty: true          # Better debugging
```

### When Override File is Active:
- ✅ **Active**: `docker-compose up` (any Docker Compose command)
- ❌ **Not Active**: DevContainer workflow
- ❌ **Not Active**: Direct `docker build`/`docker run`

## 📁 Container Files Overview

| File | Purpose | Used By |
|------|---------|---------|
| `.devcontainer/devcontainer.json` | VS Code dev environment config | DevContainer |
| `Dockerfile` | Production multi-stage build | CI/CD, Production |
| `Dockerfile.dev` | Development with hot reload | Docker Compose |
| `docker-compose.yml` | Base service definitions | Docker Compose |
| `docker-compose.override.yml` | Dev overrides + mock server | Docker Compose |
| `.air.toml` | Hot reload configuration | DevContainer, Docker Compose |

### Development Features

### Hot Reload
- Uses [Air](https://github.com/air-verse/air) for automatic rebuilds
- Watches Go files and restarts on changes
- Excludes test files and vendor directories

### Debugging
- Delve debugger pre-installed
- Debug port exposed on 2345
- VS Code debugging configuration included

### Tools Included
- Go 1.24.4
- Air (hot reload)
- Delve (debugger)
- golangci-lint
- goimports
- Git and GitHub CLI
- Docker-in-Docker support

## 🔧 Environment Variables & Configuration

### Development Environment (with Override)
When using `docker-compose up`, the override file automatically sets:
```bash
MICV_SECRET_URL=http://localhost:8081/secret          # Mock server
MICV_APPLICATION_URL=http://localhost:8081/apply      # Mock server  
MICV_TIMEOUT=30
```

### Production Environment
```bash
MICV_SECRET_URL=https://au.mitimes.com/careers/apply/secret    # Real API
MICV_APPLICATION_URL=https://au.mitimes.com/careers/apply     # Real API
MICV_TIMEOUT=30
```

### DevContainer Environment
```bash
GO111MODULE=on
GOPROXY=https://proxy.golang.org
GOSUMDB=sum.golang.org
# No URL overrides - uses configuration from config files or CLI
```

## 🧪 Testing Workflows

### 1. Unit Testing (Any Environment)
```bash
# DevContainer or local
go test -v ./...

# Docker Compose  
docker-compose --profile test up micv-test
```

### 2. Integration Testing (Docker Compose Only)
```bash
# Start with mock server
docker-compose up micv-dev

# Test CLI against mock endpoints (from another terminal)
./micv "John Doe" "john@example.com" "Software Engineer"

# Mock server provides predictable responses from testdata/
```

### 3. CLI Distribution Testing
```bash
# Build production binary
go build -ldflags "-s -w" -o micv .

# Test binary directly
./micv "Real Name" "real@email.com" "Real Job"

# Cross-platform builds for distribution
GOOS=windows GOARCH=amd64 go build -o micv-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -o micv-windows-arm64.exe .
GOOS=darwin GOARCH=amd64 go build -o micv-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o micv-darwin-arm64 .  # Apple Silicon
GOOS=linux GOARCH=amd64 go build -o micv-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o micv-linux-arm64 .   # ARM servers
```

## 🔄 Mock Server Integration

The `docker-compose.override.yml` includes a **WireMock server** that:

### Features:
- **Port**: `8081` (mapped to container port `8080`)
- **Mock Data**: Uses files from `./testdata/` directory
- **Response Templates**: Global response templating enabled
- **Verbose Logging**: Full request/response logging

### Mock Endpoints:
```bash
# Secret endpoint (returns test token)
curl http://localhost:8081/secret

# Application endpoint (accepts POST with JSON)
curl -X POST http://localhost:8081/apply \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{"name":"Test","email":"test@example.com","job_title":"Developer"}'
```

### Test Data Files:
- `testdata/test-secret-response.json` - Mock secret response
- `testdata/test-valid-data.json` - Valid application data
- `testdata/test-invalid-data.json` - Invalid data scenarios
- `testdata/test-empty-fields.json` - Empty field validation

## 🐛 Debugging

### VS Code Debugging
1. Open the project in the dev container
2. Set breakpoints in your code
3. Press F5 or use the Debug panel
4. The debugger will attach automatically

### Manual Debugging
```bash
# Start with Delve
docker run -p 2345:2345 -v $(pwd):/app micv:dev dlv debug --headless --listen=:2345 --api-version=2

# Connect from your IDE or dlv client
dlv connect localhost:2345
```

## 🔄 Common Commands

```bash
# Build and start development environment
make dev-up

# Stop development environment
make dev-down

# Run tests
make test

# Build production image
make build

# Clean up containers and volumes
make clean
```

## 📊 Monitoring

### Health Checks
- Production container includes health checks
- Endpoint: `./micv --version`
- Interval: 30s

### Logs
```bash
# View development logs
docker-compose logs -f micv-dev

# View production logs
docker-compose --profile production logs -f micv-prod
```

## 🚨 Troubleshooting

### Common Issues & Solutions

#### 1. **Docker Override Not Working**
```bash
# Verify override file is being used
docker-compose config

# Should show merged configuration with localhost:8081 URLs
```

#### 2. **Port Conflicts**
```bash
# Change ports in docker-compose.yml if needed
ports:
  - "8081:8080"  # Change first port (host)
  - "2346:2345"  # Change debug port
```

#### 3. **Mock Server Not Responding**
```bash
# Check mock server status
docker-compose logs mock-server

# Verify test data files exist
ls -la testdata/

# Test mock server directly
curl -v http://localhost:8081/secret
```

#### 4. **DevContainer vs Docker Compose Confusion**
| Issue | Solution |
|-------|----------|
| "Override not working in DevContainer" | Override only works with `docker-compose`, not DevContainer |
| "Mock server not available in DevContainer" | Use `docker-compose up` for mock server testing |
| "Can't access localhost:8081" | Ensure you're using Docker Compose, not DevContainer |

#### 5. **Hot Reload Issues**
```bash
# Check Air configuration
cat .air.toml

# Verify file watching
docker-compose logs micv-dev | grep "watching"

# Restart if needed
docker-compose restart micv-dev
```

#### 6. **Permission Issues**
```bash
# Fix ownership (Linux/macOS)
sudo chown -R $USER:$USER .

# Windows WSL2
wsl --shutdown
# Restart WSL and Docker Desktop
```

#### 7. **Go Module Issues**
```bash
# Clear module cache
docker-compose down -v
docker system prune -f
docker-compose up micv-dev
```

## � Development Workflow Comparison

### When to Use Each Approach:

#### DevContainer ✅
- **Daily coding and debugging**
- **File editing with immediate feedback**
- **Running unit tests**
- **Integrated VS Code experience**
- **Learning Go development**

#### Docker Compose ✅  
- **Integration testing with mock APIs**
- **Testing CLI behavior with controlled responses**
- **Validating environment configurations**
- **CI/CD pipeline testing**
- **Demo/presentation setups with mock data**

#### Local CLI ✅
- **Actual CLI usage and execution**
- **Performance benchmarking**
- **Real job application submissions**
- **Cross-platform binary testing**
- **End-user experience validation**

### Workflow Examples:

```bash
# Typical development day:
# 1. Start with DevContainer for coding
code .  # Opens in DevContainer

# 2. Switch to Docker Compose for integration testing  
docker-compose up micv-dev  # Test with mocks

# 3. Build and test CLI locally
go build -o micv .
./micv "Test User" "test@example.com" "Software Engineer"

# 4. Cross-platform builds for distribution
make build-all  # Creates binaries for different platforms
```

## 🚀 CLI Distribution

Since this is a **CLI tool**, the main "production" concern is **binary distribution**:

### Building Optimized Binaries
```bash
# Optimized build (smaller binary)
go build -ldflags "-s -w" -o micv .

# With version info
go build -ldflags "-s -w -X main.Version=v1.0.0" -o micv .

# Cross-platform builds (AMD64)
GOOS=windows GOARCH=amd64 go build -o micv-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 go build -o micv-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o micv-linux-amd64 .

# Cross-platform builds (ARM64)
GOOS=windows GOARCH=arm64 go build -o micv-windows-arm64.exe .
GOOS=darwin GOARCH=arm64 go build -o micv-darwin-arm64 .  # Apple Silicon
GOOS=linux GOARCH=arm64 go build -o micv-linux-arm64 .   # ARM servers/Raspberry Pi

# Build all platforms at once
make build-all  # If Makefile includes multi-arch targets
```

### Distribution Methods
```bash
# GitHub Releases (recommended) - include all architectures
gh release create v1.0.0 \
  micv-linux-amd64 \
  micv-linux-arm64 \
  micv-darwin-amd64 \
  micv-darwin-arm64 \
  micv-windows-amd64.exe \
  micv-windows-arm64.exe

# Package managers
# go install github.com/yourorg/micv@latest

# Docker multi-arch images (if needed)
docker buildx build --platform linux/amd64,linux/arm64 -t micv:latest .
docker run --rm micv:latest "John Doe" "john@example.com" "Software Engineer"
```

### Platform-Specific Notes
| Platform | Architecture | Notes |
|----------|--------------|--------|
| **Windows** | AMD64 | Standard 64-bit Windows |
| **Windows** | ARM64 | Windows on ARM (Surface Pro X, etc.) |
| **macOS** | AMD64 | Intel Macs |
| **macOS** | ARM64 | Apple Silicon (M1/M2/M3 Macs) |
| **Linux** | AMD64 | Standard 64-bit Linux |
| **Linux** | ARM64 | ARM servers, Raspberry Pi, AWS Graviton |

**Note**: The Dockerfile creates a containerized CLI that can be useful for:
- Consistent runtime environments
- CI/CD pipelines  
- Environments where Go isn't installed

### Automated Build Script
```bash
#!/bin/bash
# build-all.sh - Build for all platforms

VERSION=${1:-"dev"}
LDFLAGS="-s -w -X main.Version=${VERSION}"

echo "Building micv ${VERSION} for all platforms..."

# AMD64 builds
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-windows-amd64.exe" .
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-darwin-amd64" .
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-linux-amd64" .

# ARM64 builds  
GOOS=windows GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/micv-windows-arm64.exe" .
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/micv-darwin-arm64" .
GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/micv-linux-arm64" .

echo "✅ Build complete! Binaries in dist/ directory"
ls -la dist/
```

Usage:
```bash
# Build with version
./build-all.sh v1.0.0

# Build dev version
./build-all.sh
```

## 🔗 Related Documentation

- **[README.md](./README.md)** - Main project documentation
- **[REQUIREMENTS.md](./REQUIREMENTS.md)** - Project requirements and specifications
- **`.devcontainer/devcontainer.json`** - DevContainer configuration
- **`docker-compose.yml`** - Service definitions
- **`docker-compose.override.yml`** - Development overrides
- **`.air.toml`** - Hot reload configuration

## 📞 Support

For container-specific issues:

1. **Check logs**: `docker-compose logs [service-name]`
2. **Verify configuration**: `docker-compose config`
3. **Test connectivity**: `curl -v http://localhost:8081/secret`
4. **Review documentation**: Check this file and README.md
5. **Create issue**: Include container logs and environment details
