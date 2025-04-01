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
