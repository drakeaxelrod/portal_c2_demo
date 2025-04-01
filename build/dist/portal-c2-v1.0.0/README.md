# Portal C2 Framework

Portal is a Command and Control (C2) framework built with Go, gRPC, and Wails. It provides a modern, efficient, and secure way to manage remote agents with a beautiful user interface.

## Features

- **gRPC Communication**: Bidirectional streaming for real-time command and control
- **Cross-Platform**: Agents work on Windows, Linux, and macOS
- **Modern UI**: Built with Wails and React for a responsive user experience
- **Extensible**: Easily add new command types and functionality
- **Secure**: Encrypted communications between server and agents

## Project Structure

- `/cmd`: Command-line tools
  - `/server`: Server CLI
  - `/agent`: Agent CLI
- `/pkg`: Core packages
  - `/server`: C2 server implementation
  - `/agent`: C2 agent implementation
  - `/common`: Shared utilities
- `/proto`: Protocol Buffer definitions
- `/frontend`: Wails UI frontend

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Wails v2
- Protocol Buffers compiler
- Just command runner

### Setup

1. Install dependencies:

```bash
just setup
```

2. Generate gRPC code:

```bash
just gen-proto
```

3. Build all components:

```bash
just build
```

### Running

#### Start the server:

```bash
just run-server
```

#### Start the GUI:

```bash
just run-ui
```

#### Start an agent:

```bash
just run-agent SERVER=your-server-ip:50051
```

## Development

For live development of the UI:

```bash
just run-ui
```

## Building for Distribution

```bash
just build
```

## License

[MIT License](LICENSE)
