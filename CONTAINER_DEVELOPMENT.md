# Container Development Guide

## 🚀 Quick Start

### Option 1: VS Code Dev Container (Recommended)
```bash
# 1. Install "Dev Containers" extension in VS Code
# 2. Open project: Ctrl+Shift+P → "Dev Containers: Reopen in Container"
# 3. Container auto-builds with Go 1.24 + hot reload + debugging
```

### Option 2: Docker Compose (Integration Testing)
```bash
# Start development environment with mock server
docker-compose up micv-dev

# Test against mock APIs at localhost:8081
curl http://localhost:8081/secret
```

### Option 3: Local CLI
```bash
# Build and run locally
go build -o micv .
./micv "John Doe" "john@example.com" "Software Engineer"
```

## 🔧 Essential Commands

```bash
# Run tests
go test -v ./...

# Build for development
go build -o micv .

# Run with hot reload (in DevContainer)
air

# Start mock server for testing
docker-compose up micv-dev
```

## 🏗️ Building & Releasing

### Local Development Build
```bash
go build -ldflags "-s -w" -o micv .
```

### Cross-Platform Release Builds
```bash
# Create dist/ directory
mkdir -p dist

# Build for all platforms
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o "dist/micv-windows-amd64.exe" .
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o "dist/micv-darwin-amd64" .
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o "dist/micv-darwin-arm64" .
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "dist/micv-linux-amd64" .
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o "dist/micv-linux-arm64" .
```

### Automated Build Script
Save as `build-all.sh`:
```bash
#!/bin/bash
VERSION=${1:-"dev"}
LDFLAGS="-s -w -X main.Version=${VERSION}"

echo "Building micv ${VERSION} for all platforms..."
mkdir -p dist

GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-windows-amd64.exe" .
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-darwin-amd64" .
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/micv-darwin-arm64" .
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/micv-linux-amd64" .
GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/micv-linux-arm64" .

echo "✅ Build complete! Binaries in dist/ directory"
ls -la dist/
```

Usage:
```bash
chmod +x build-all.sh
./build-all.sh v1.0.0    # With version
./build-all.sh           # Development build
```

## 🧪 Testing

### Unit Tests
```bash
go test -v ./...
```

### Integration Tests (with mock server)
```bash
# Terminal 1: Start mock server
docker-compose up micv-dev

# Terminal 2: Test CLI
./micv "Test User" "test@example.com" "Developer"
```

## 🔧 Configuration

### Environment Variables
- **Development**: Uses mock server at `localhost:8081`
- **Production**: Uses real APIs at `au.mitimes.com`

### Override behavior
- `docker-compose.override.yml` automatically redirects to mock server
- DevContainer and local builds use real endpoints by default

## 🚨 Troubleshooting

```bash
# Verify Docker Compose config
docker-compose config

# Check mock server
curl -v http://localhost:8081/secret

# View logs
docker-compose logs micv-dev

# Clean rebuild
docker-compose down -v && docker-compose up micv-dev
```

## � Related Documentation

- **[README.md](./README.md)** - Main project documentation
- **[REQUIREMENTS.md](./REQUIREMENTS.md)** - Project requirements and specifications
