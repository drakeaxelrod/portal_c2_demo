#!/bin/bash

# Exit on error
set -e

VERSION="1.0.0"
PACKAGE_NAME="portal-c2-v$VERSION"

echo "Creating distribution package for Portal C2 Framework v$VERSION"

# Create directory structure
mkdir -p dist/$PACKAGE_NAME/agents
mkdir -p dist/$PACKAGE_NAME/server

# Copy binaries
echo "Copying binaries..."

# Server
cp bin/server dist/$PACKAGE_NAME/server/

# Windows agents
cp bin/agent-windows-amd64.exe dist/$PACKAGE_NAME/agents/
cp bin/agent-windows-386.exe dist/$PACKAGE_NAME/agents/

# macOS agents
cp bin/agent-macos-amd64 dist/$PACKAGE_NAME/agents/
cp bin/agent-macos-arm64 dist/$PACKAGE_NAME/agents/

# Linux agents
cp bin/agent-linux-amd64 dist/$PACKAGE_NAME/agents/
cp bin/agent-linux-386 dist/$PACKAGE_NAME/agents/
cp bin/agent-linux-arm64 dist/$PACKAGE_NAME/agents/
cp bin/agent-linux-arm dist/$PACKAGE_NAME/agents/

# Copy README and other docs
cp README.md dist/$PACKAGE_NAME/

# Create a simple usage document
cat > dist/$PACKAGE_NAME/USAGE.md << 'EOF'
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

# Create archive
echo "Creating archive..."
cd dist
tar -czvf $PACKAGE_NAME.tar.gz $PACKAGE_NAME

echo "Package created: dist/$PACKAGE_NAME.tar.gz"