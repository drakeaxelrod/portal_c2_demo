#!/bin/bash

# Exit on error
set -e

echo "Generating Go code from protobuf definitions..."

# Ensure protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "protoc could not be found. Please install Protocol Buffers compiler."
    exit 1
fi

# Get the GOPATH
GOPATH=$(go env GOPATH)
PROTOC_GEN_GO="${GOPATH}/bin/protoc-gen-go"
PROTOC_GEN_GO_GRPC="${GOPATH}/bin/protoc-gen-go-grpc"

# Check if the protoc plugins exist
if [ ! -f "${PROTOC_GEN_GO}" ]; then
    echo "protoc-gen-go not found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
fi

if [ ! -f "${PROTOC_GEN_GO_GRPC}" ]; then
    echo "protoc-gen-go-grpc not found. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
fi

# Create the output directory if it doesn't exist
mkdir -p proto/gen

# Generate Go code
protoc --plugin=protoc-gen-go="${PROTOC_GEN_GO}" \
       --plugin=protoc-gen-go-grpc="${PROTOC_GEN_GO_GRPC}" \
       --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/c2.proto

# Clean up go.mod and go.sum
go mod tidy

echo "Code generation completed successfully."