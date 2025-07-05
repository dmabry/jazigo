




#!/bin/bash

set -e  # Exit on error
set -x  # Print commands for debugging

# Install dependencies first to avoid issues with staticcheck
go mod download

# Install and run static analysis tools
go install golang.org/x/vuln/cmd/govulncheck@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Format code
gofmt -s -w .

# Run linters
staticcheck ./...

# Tidy up dependencies
go mod tidy

# Check for vulnerabilities
govulncheck ./...

# Enable CGO for testing
go env -w CGO_ENABLED=1

# Run tests with race detector and verbose output
go test -race -v ./...

# Disable CGO after testing
go env -w CGO_ENABLED=0

# Build and install the project
go install ./...

# Clean up environment
go env -u CGO_ENABLED

echo "Build completed successfully!"




