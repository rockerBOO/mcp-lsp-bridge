#!/bin/bash

# Development Container Management Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Development Container Management Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build       Build the development containers
    dev         Start interactive development environment
    ci          Run full CI pipeline (matches CI environment exactly)
    test        Run tests in CI-like environment
    lint        Run linter in CI-like environment
    security    Run security scan in CI-like environment
    clean       Clean up containers and volumes
    shell       Open shell in development container
    logs        Show logs from running containers
    help        Show this help message

Examples:
    $0 build                    # Build all containers
    $0 ci                       # Run full CI pipeline
    $0 test                     # Run just tests
    $0 dev                      # Start development environment
    $0 shell                    # Open interactive shell
    $0 clean                    # Clean up everything

Environment Variables:
    DOCKER_BUILDKIT=1          # Enable BuildKit for faster builds
    COMPOSE_DOCKER_CLI_BUILD=1 # Use Docker CLI for building
EOF
}

# Check if Docker and Docker Compose are available
check_dependencies() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
}

# Build containers
build_containers() {
    print_status "Building development containers..."
    export DOCKER_BUILDKIT=1
    export COMPOSE_DOCKER_CLI_BUILD=1
    
    docker-compose -f docker-compose.dev.yml build
    print_success "Containers built successfully"
}

# Start development environment
start_dev() {
    print_status "Starting development environment..."
    docker-compose -f docker-compose.dev.yml up -d dev
    print_success "Development environment started"
    print_status "Run '$0 shell' to open an interactive shell"
}

# Run CI pipeline
run_ci() {
    print_status "Running CI pipeline in Docker (matches CI environment)..."
    docker-compose -f docker-compose.dev.yml run --rm ci
    
    if [ $? -eq 0 ]; then
        print_success "CI pipeline completed successfully"
    else
        print_error "CI pipeline failed"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_status "Running tests in CI-like environment..."
    docker-compose -f docker-compose.dev.yml run --rm test
    
    if [ $? -eq 0 ]; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Run linter
run_lint() {
    print_status "Running linter in CI-like environment..."
    docker-compose -f docker-compose.dev.yml run --rm lint
    
    if [ $? -eq 0 ]; then
        print_success "Linting passed"
    else
        print_error "Linting failed"
        exit 1
    fi
}

# Run security scan
run_security() {
    print_status "Running security scan in CI-like environment..."
    docker-compose -f docker-compose.dev.yml run --rm security
    
    if [ $? -eq 0 ]; then
        print_success "Security scan passed"
    else
        print_error "Security scan failed"
        exit 1
    fi
}

# Open shell in development container
open_shell() {
    print_status "Opening shell in development container..."
    
    # Check if dev container is running
    if ! docker-compose -f docker-compose.dev.yml ps dev | grep -q "Up"; then
        print_status "Starting development container..."
        docker-compose -f docker-compose.dev.yml up -d dev
    fi
    
    docker-compose -f docker-compose.dev.yml exec dev bash
}

# Show logs
show_logs() {
    docker-compose -f docker-compose.dev.yml logs -f
}

# Clean up
cleanup() {
    print_warning "Cleaning up containers and volumes..."
    docker-compose -f docker-compose.dev.yml down -v
    docker system prune -f
    print_success "Cleanup completed"
}

# Main script logic
main() {
    check_dependencies
    
    case "$1" in
        build)
            build_containers
            ;;
        dev)
            build_containers
            start_dev
            ;;
        ci)
            build_containers
            run_ci
            ;;
        test)
            build_containers
            run_tests
            ;;
        lint)
            build_containers
            run_lint
            ;;
        security)
            build_containers
            run_security
            ;;
        shell)
            build_containers
            open_shell
            ;;
        logs)
            show_logs
            ;;
        clean)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            print_error "No command specified"
            show_help
            exit 1
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"