#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

case "${1:-help}" in
    build)
        echo "Building bar..."
        go build -o bin/bar ./cmd/bar
        echo "✅ Built: bin/bar"
        ;;
    
    run)
        shift
        go run ./cmd/bar "$@"
        ;;
    
    test)
        echo "Running tests..."
        go test -v ./...
        ;;
    
    test-cover)
        echo "Running tests with coverage..."
        go test -v -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
        echo "✅ Coverage report: coverage.html"
        ;;
    
    lint)
        echo "Running linter..."
        if command -v golangci-lint &> /dev/null; then
            golangci-lint run
        else
            echo "golangci-lint not found, using go vet..."
            go vet ./...
        fi
        ;;
    
    fmt)
        echo "Formatting code..."
        go fmt ./...
        echo "✅ Code formatted"
        ;;
    
    deps)
        echo "Installing dependencies..."
        go mod download
        go mod tidy
        echo "✅ Dependencies installed"
        ;;
    
    clean)
        echo "Cleaning..."
        rm -rf bin/ coverage.out coverage.html
        go clean
        echo "✅ Cleaned"
        ;;
    
    install)
        echo "Installing bar to /usr/local/bin..."
        go build -o bin/bar ./cmd/bar
        sudo cp bin/bar /usr/local/bin/bar
        echo "✅ Installed: /usr/local/bin/bar"
        ;;
    
    help|*)
        echo "Development helper script for BAR"
        echo ""
        echo "Usage: ./scripts/dev.sh <command>"
        echo ""
        echo "Commands:"
        echo "  build       Build the bar binary"
        echo "  run         Run bar with arguments (e.g., ./scripts/dev.sh run init)"
        echo "  test        Run all tests"
        echo "  test-cover  Run tests with coverage report"
        echo "  lint        Run linter"
        echo "  fmt         Format code"
        echo "  deps        Install/update dependencies"
        echo "  clean       Clean build artifacts"
        echo "  install     Install bar to /usr/local/bin"
        echo "  help        Show this help message"
        ;;
esac
