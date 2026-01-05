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

# Run linter (matches CI settings - pinned to same version)
lint:
    go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2 run

# Build the binary
build:
    go build -o dist/did .
    @echo "Built: {{justfile_directory()}}/dist/did"

# Build and install to ~/.local/bin/
install: build
    mkdir -p ~/.local/bin
    install --mode +x dist/did ~/.local/bin/

# Build release artifacts with GoReleaser (local snapshot, no publish)
release:
    go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean

# Validate GoReleaser configuration
release-check:
    go run github.com/goreleaser/goreleaser/v2@latest check
