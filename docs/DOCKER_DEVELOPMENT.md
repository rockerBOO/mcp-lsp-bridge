# Docker Development Environment

This document describes how to use Docker for development and testing to ensure consistency with CI environments.

## Overview

The Docker development setup provides:
- **Exact CI environment replication** for testing
- **Consistent development environment** across all machines
- **Isolated testing** without affecting local Go installation
- **Security scanning** with the same tools as CI
- **Hot reload development** with live code changes

## Quick Start

```bash
# Run the complete CI pipeline locally
./scripts/dev-container.sh ci

# Start interactive development environment
./scripts/dev-container.sh dev

# Run just tests in CI-like environment
./scripts/dev-container.sh test

# Run linting in CI-like environment
./scripts/dev-container.sh lint

# Run security scan in CI-like environment
./scripts/dev-container.sh security
```

## Available Commands

### Development Commands

```bash
# Build all Docker containers
./scripts/dev-container.sh build

# Start development environment with hot reload
./scripts/dev-container.sh dev

# Open interactive shell in development container
./scripts/dev-container.sh shell
```

### Testing Commands

```bash
# Run complete CI pipeline (recommended before pushing)
./scripts/dev-container.sh ci

# Run only tests
./scripts/dev-container.sh test

# Run only linting
./scripts/dev-container.sh lint

# Run only security scanning
./scripts/dev-container.sh security
```

### Utility Commands

```bash
# View logs from running containers
./scripts/dev-container.sh logs

# Clean up containers and volumes
./scripts/dev-container.sh clean

# Show help
./scripts/dev-container.sh help
```

## Container Architecture

### Base Container (`base`)
- Go 1.21 Alpine Linux
- Essential build tools (git, make, curl)
- golangci-lint for linting
- nancy for dependency scanning
- Source code and dependencies

### Development Container (`dev`)
- Extends base container
- Adds development tools (air, delve)
- Hot reload capabilities
- Debugging support (port 2345)
- Interactive shell access

### CI Container (`ci`)
- Identical to CI environment
- Runs as root user (matches CI)
- Executes complete CI pipeline
- Includes security scanning

## File Structure

```
├── Dockerfile.dev              # Multi-stage Docker configuration
├── docker-compose.dev.yml      # Development services configuration
├── scripts/dev-container.sh    # Container management script
├── .dockerignore               # Files to exclude from Docker context
└── docs/DOCKER_DEVELOPMENT.md  # This documentation
```

## Development Workflow

### 1. Setting Up Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd mcp-lsp-bridge

# Start development environment
./scripts/dev-container.sh dev

# Open shell in development container
./scripts/dev-container.sh shell
```

### 2. Making Changes

```bash
# In the development container, code changes are automatically reflected
# Run tests as you develop
go test ./...

# Run specific tests
go test ./mcpserver/tools -v

# Build the application
make build
```

### 3. Pre-Commit Validation

```bash
# Run the complete CI pipeline before committing
./scripts/dev-container.sh ci

# This runs:
# - go test ./...
# - make lint
# - make security-scan
# - make build
```

### 4. Debugging Failed CI

```bash
# If CI fails, reproduce locally:
./scripts/dev-container.sh ci

# Or run specific steps:
./scripts/dev-container.sh test     # Just tests
./scripts/dev-container.sh lint     # Just linting
./scripts/dev-container.sh security # Just security scan
```

## Environment Variables

The containers use these environment variables to match CI:

```bash
CGO_ENABLED=0     # Disable CGO for static binaries
GOOS=linux        # Target Linux (matches CI)
```

## Volume Mounts

- **Source code**: `.:/workspace` (live editing)
- **Go modules**: `go-mod-cache:/go/pkg/mod` (persistent cache)
- **Docker socket**: `/var/run/docker.sock:/var/run/docker.sock` (for security scanning)

## Troubleshooting

### Container Build Issues

```bash
# Clean up and rebuild
./scripts/dev-container.sh clean
./scripts/dev-container.sh build

# Check Docker version
docker --version
docker-compose --version
```

### Permission Issues

```bash
# If you encounter permission issues, the containers run as root
# This matches the CI environment behavior
```

### Network Issues

```bash
# If containers can't access external resources:
docker network ls
docker network inspect bridge
```

### Performance Issues

```bash
# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

## CI Environment Matching

The Docker setup exactly matches CI environments:

| Aspect | Local Docker | CI Environment |
|--------|-------------|----------------|
| OS | Alpine Linux | Alpine Linux |
| Go Version | 1.21 | 1.21 |
| User | root | root |
| Build Tools | Same versions | Same versions |
| Environment Variables | Identical | Identical |
| File Permissions | Same | Same |

## Security Considerations

- Containers run as root (matches CI)
- Security scanning uses same tools as CI
- File permissions match CI requirements
- Temp directories behave identically to CI

## Integration with IDE

### VS Code

Add to `.vscode/settings.json`:

```json
{
    "go.toolsEnvVars": {
        "CGO_ENABLED": "0",
        "GOOS": "linux"
    },
    "go.buildEnvVars": {
        "CGO_ENABLED": "0",
        "GOOS": "linux"
    }
}
```

### GoLand/IntelliJ

Configure Go environment:
- GOOS: linux
- CGO_ENABLED: 0

## Best Practices

1. **Always run CI pipeline before pushing**:
   ```bash
   ./scripts/dev-container.sh ci
   ```

2. **Use development container for consistent environment**:
   ```bash
   ./scripts/dev-container.sh shell
   ```

3. **Test specific scenarios that failed in CI**:
   ```bash
   ./scripts/dev-container.sh test
   ```

4. **Clean up regularly to save disk space**:
   ```bash
   ./scripts/dev-container.sh clean
   ```

5. **Keep containers updated**:
   ```bash
   ./scripts/dev-container.sh build
   ```

## Examples

### Reproducing CI Test Failure

```bash
# CI failed on TestTryLoadConfig
./scripts/dev-container.sh shell

# Inside container:
go test -v -run "TestTryLoadConfig"
```

### Running Security Scan Locally

```bash
# Run the same security scan as CI
./scripts/dev-container.sh security

# Should output:
# Gosec: 0 issues
# Nancy: 0 vulnerable dependencies
```

### Development with Hot Reload

```bash
# Start development environment
./scripts/dev-container.sh dev

# In another terminal, open shell
./scripts/dev-container.sh shell

# Code changes are automatically available in container
```

This Docker setup ensures that local development and testing exactly matches the CI environment, preventing CI failures due to environment differences.