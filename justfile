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

# Build the binary
build:
    go build -o did .

# Build and install to ~/.local/bin/
install: build
    mkdir -p ~/.local/bin
    cp did ~/.local/bin/
