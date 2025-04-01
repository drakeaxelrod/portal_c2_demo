#!/bin/bash

# Exit on error
set -e

# Default build and dist directories if not set by environment
BUILD_DIR=${BUILD_DIR:-"build"}
DIST_DIR=${DIST_DIR:-"$BUILD_DIR/dist"}
VERSION="1.0.0"
PACKAGE_NAME="portal-c2-v$VERSION"

echo "Creating distribution package for Portal C2 Framework v$VERSION"

# Create dist directory
mkdir -p $DIST_DIR/$PACKAGE_NAME/agents
mkdir -p $DIST_DIR/$PACKAGE_NAME/server
mkdir -p $DIST_DIR/$PACKAGE_NAME/docs

# Copy binaries
echo "Copying binaries..."

# Server
cp $BUILD_DIR/server/server $DIST_DIR/$PACKAGE_NAME/server/

# Windows agents
cp $BUILD_DIR/agents/agent-windows-amd64.exe $DIST_DIR/$PACKAGE_NAME/agents/
cp $BUILD_DIR/agents/agent-windows-386.exe $DIST_DIR/$PACKAGE_NAME/agents/

# macOS agents
cp $BUILD_DIR/agents/agent-macos-amd64 $DIST_DIR/$PACKAGE_NAME/agents/
cp $BUILD_DIR/agents/agent-macos-arm64 $DIST_DIR/$PACKAGE_NAME/agents/

# Linux agents
cp $BUILD_DIR/agents/agent-linux-amd64 $DIST_DIR/$PACKAGE_NAME/agents/
cp $BUILD_DIR/agents/agent-linux-386 $DIST_DIR/$PACKAGE_NAME/agents/
cp $BUILD_DIR/agents/agent-linux-arm64 $DIST_DIR/$PACKAGE_NAME/agents/
cp $BUILD_DIR/agents/agent-linux-arm $DIST_DIR/$PACKAGE_NAME/agents/

# Copy README and other docs
cp README.md $DIST_DIR/$PACKAGE_NAME/

# Create a simple usage document
cat > $DIST_DIR/$PACKAGE_NAME/USAGE.md << 'EOF'
# Portal C2 Framework Usage Guide

## Server
To start the C2 server:
```
./server/server -addr 0.0.0.0:50051
```

## GUI
The Wails GUI application can be built using:
```
just build-ui
```

## Agents
To run an agent, use the appropriate binary for your target OS and architecture:

### Windows
```
agents\agent-windows-amd64.exe -server your-server-ip:50051
```

### macOS
```
chmod +x agents/agent-macos-amd64
./agents/agent-macos-amd64 -server your-server-ip:50051
```

### Linux
```
chmod +x agents/agent-linux-amd64
./agents/agent-linux-amd64 -server your-server-ip:50051
```

## Commands
The following command types are supported:
- shell: Execute shell commands
- system: Get system information
- process: List running processes
- screenshot: Take a screenshot (not fully implemented)
- upload/download: File transfer (not fully implemented)
EOF

# Create additional documentation
echo "Creating documentation..."

# Agent Documentation
cat > $DIST_DIR/$PACKAGE_NAME/docs/agent.md << 'EOF'
# Agent Documentation

## Overview
The agent component is a lightweight client that connects to the C2 server and executes commands.

## Configuration
Agents support the following command-line parameters:

- `-server`: The server address in the format of host:port (default: localhost:50051)
- `-id`: Custom agent ID (optional)

## Supported Commands
- Shell commands: Run any command on the target system
- System info: Get detailed system information
- Process management: List and manipulate processes
- Screenshot: Capture the screen (coming soon)
- File transfer: Upload/download files (coming soon)

## Persistence
The agent does not currently implement persistence mechanisms. You will need to use OS-specific methods to achieve persistence.

## Network Protocol
The agent communicates with the server using gRPC over TLS (when configured).
EOF

# Server Documentation
cat > $DIST_DIR/$PACKAGE_NAME/docs/server.md << 'EOF'
# Server Documentation

## Overview
The server component is the command and control center that manages connected agents.

## Configuration
The server supports the following command-line parameters:

- `-addr`: The server address to listen on (default: 0.0.0.0:50051)

## Architecture
The server is built using Go and gRPC for efficient, bi-directional streaming communication.

## Security Considerations
- The server does not currently implement TLS. In production, consider:
  - Adding TLS certificates
  - Implementing proper authentication
  - Running behind a reverse proxy
EOF

# Usage Examples Documentation
cat > $DIST_DIR/$PACKAGE_NAME/docs/examples.md << 'EOF'
# Usage Examples

## Basic Usage

### Starting the Server
```bash
# Start the server on default port
./server/server

# Start the server on a custom port
./server/server -addr 0.0.0.0:8080
```

### Connecting an Agent
```bash
# Linux/macOS
./agents/agent-linux-amd64 -server your-server-ip:50051

# Windows
.\agents\agent-windows-amd64.exe -server your-server-ip:50051
```

## Command Examples

### Execute a shell command
Send the following command type: `shell`
With the payload: `whoami` or `dir` (Windows)

### Get system information
Send the following command type: `system`
This will return system details like OS, architecture, CPU, memory, etc.

### Process list
Send the following command type: `process`
This will enumerate running processes on the target system.

## Web UI Usage

The web UI provides a graphical interface to:
1. View connected agents
2. Send commands to agents
3. View command output
4. Monitor agent status

To access the web UI, start the Wails application:
```bash
just run-ui
```

Or build the UI and run it:
```bash
just build-ui
./build/portal
```
EOF

# Create archive
echo "Creating archive..."
cd $DIST_DIR
tar -czvf $PACKAGE_NAME.tar.gz $PACKAGE_NAME

echo "Package created: $DIST_DIR/$PACKAGE_NAME.tar.gz"