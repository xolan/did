# Justfile for did - a simple task tracking tool

# Default recipe shows available commands
default:
    @just --list

# Install tools and fetch Go dependencies
setup:
    mise install
    go mod download

# Run the test suite
test:
    go test ./...

# Format code
format:
    go fmt ./...

# Run linter
lint:
    go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Build the binary
build:
    go build -o did .

# Build and install to ~/.local/bin/
install: build
    mkdir -p ~/.local/bin
    cp did ~/.local/bin/
