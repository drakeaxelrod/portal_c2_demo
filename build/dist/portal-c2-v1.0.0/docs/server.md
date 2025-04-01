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
