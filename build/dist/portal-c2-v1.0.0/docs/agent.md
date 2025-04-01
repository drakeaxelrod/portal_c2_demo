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
