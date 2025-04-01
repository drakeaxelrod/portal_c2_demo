#!/bin/bash

# Exit on error
set -e

echo "Portal C2 Framework Setup"
echo "========================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go before continuing."
    exit 1
fi

echo "Installing required Go tools..."

# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Check if Just is installed
if ! command -v just &> /dev/null; then
    echo "Installing Just command runner..."
    # For Linux
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
    # For macOS with Homebrew
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew install just
        else
            echo "Homebrew not found. Please install Homebrew or manually install Just."
        fi
    else
        echo "Please install Just manually: https://github.com/casey/just#installation"
    fi
fi

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Warning: Protocol Buffer compiler (protoc) not found."
    echo "Please install protoc manually to generate gRPC code:"
    echo "  - Ubuntu/Debian: apt install protobuf-compiler"
    echo "  - macOS: brew install protobuf"
    echo "  - Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
fi

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo "Installing Wails..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

echo "Downloading dependencies..."
go mod tidy

echo "Creating directories..."
mkdir -p bin proto/gen

echo ""
echo "Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Generate gRPC code:      just gen-proto"
echo "2. Build all components:    just build"
echo "3. Run the UI application:  just run-ui"
echo ""
