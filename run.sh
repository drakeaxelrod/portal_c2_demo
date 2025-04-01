#!/bin/bash

# Exit on any error
set -e

# Kill any existing server and agent processes
echo "Stopping any existing server and agent processes..."
pkill -f "go run cmd/server/main.go" || true
pkill -f "go run cmd/agent/main.go" || true

# Start the server
echo "Starting C2 server..."
go run cmd/server/main.go cmd/server/api.go &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 2

# Start the agent
echo "Starting test agent..."
go run cmd/agent/main.go &
AGENT_PID=$!

echo "Server running with PID: $SERVER_PID"
echo "Test agent running with PID: $AGENT_PID"
echo ""
echo "API is available at: http://localhost:8080/api/agents"
echo "WebSocket shell is available at: ws://localhost:8080/api/agents/{agentId}/shell"
echo ""
echo "Press Ctrl+C to stop all processes"

# Trap Ctrl+C to kill both processes
trap "echo 'Stopping all processes...'; kill $SERVER_PID $AGENT_PID 2>/dev/null || true" INT

# Wait for Ctrl+C
wait