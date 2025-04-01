# Justfile for Portal C2 Framework

# Variables
SERVER_ADDR := "0.0.0.0:50051"

# Default recipe (shows help)
default:
    @just --list

# Install required tools
setup:
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf code
gen-proto:
    @echo "Generating Go code from protobuf files..."
    bash scripts/generate_proto.sh

# Build the C2 server CLI
build-server:
    @echo "Building C2 server..."
    go build -o bin/server cmd/server/main.go

# Build the C2 agent CLI for the current platform
build-agent:
    @echo "Building C2 agent for current platform..."
    go build -o bin/agent cmd/agent/main.go

# Build agents for all platforms (Windows, macOS, Linux)
build-agents-all: build-agent-windows build-agent-macos build-agent-linux
    @echo "All agents built successfully!"

# Build agent for Windows
build-agent-windows:
    @echo "Building agent for Windows..."
    GOOS=windows GOARCH=amd64 go build -o bin/agent-windows-amd64.exe cmd/agent/main.go
    GOOS=windows GOARCH=386 go build -o bin/agent-windows-386.exe cmd/agent/main.go

# Build agent for macOS
build-agent-macos:
    @echo "Building agent for macOS..."
    GOOS=darwin GOARCH=amd64 go build -o bin/agent-macos-amd64 cmd/agent/main.go
    GOOS=darwin GOARCH=arm64 go build -o bin/agent-macos-arm64 cmd/agent/main.go

# Build agent for Linux
build-agent-linux:
    @echo "Building agent for Linux..."
    GOOS=linux GOARCH=amd64 go build -o bin/agent-linux-amd64 cmd/agent/main.go
    GOOS=linux GOARCH=386 go build -o bin/agent-linux-386 cmd/agent/main.go
    GOOS=linux GOARCH=arm64 go build -o bin/agent-linux-arm64 cmd/agent/main.go
    GOOS=linux GOARCH=arm go build -o bin/agent-linux-arm cmd/agent/main.go

# Build the Wails UI
build-ui:
    @echo "Building Wails UI..."
    wails build

# Build all components
build: gen-proto build-server build-agents-all build-ui
    @echo "All components built successfully!"

# Create a distribution package
package: build
    @echo "Creating distribution package..."
    chmod +x scripts/package.sh
    bash scripts/package.sh

# Run the server
run-server: build-server
    @echo "Starting C2 server on ${SERVER_ADDR}..."
    ./bin/server -addr ${SERVER_ADDR}

# Run the agent
run-agent SERVER="localhost:50051":
    @echo "Starting C2 agent connecting to ${SERVER}..."
    ./bin/agent -server ${SERVER}

# Run the GUI
run-ui:
    @echo "Starting Wails UI..."
    wails dev

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf build/
    rm -rf frontend/dist/
    rm -rf dist/

# Full rebuild
rebuild: clean build
    @echo "Rebuild complete!"

# Update all dependencies
update-deps:
    go get -u all
    go mod tidy
