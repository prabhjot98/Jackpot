# Set the default shell to bash
set shell := ["bash", "-c"]

# Define variables
_BINARY_NAME := "jackpot"
_MAIN_PATH := "./src/main.go"

# Default recipe (runs when you just type 'just')
default:
    @just --list

# Build the project
build:
    go build -o {{_BINARY_NAME}} {{_MAIN_PATH}}

# Run the project
run:
    go run {{_MAIN_PATH}}

# Clean build artifacts
clean:
    go clean
    rm -f {{_BINARY_NAME}}

# Run tests
test:
    go test ./...

# Get dependencies (less necessary with Go modules, but included for completeness)
deps:
    go get ./...

# Build and run
all: build run

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Generate code coverage report
cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Show help
help:
    @just --list
