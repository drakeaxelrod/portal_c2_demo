package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"portal/pkg/server"
)

func main() {
	// Parse command line flags
	addr := flag.String("addr", "0.0.0.0:50051", "The server address in the format of host:port")
	flag.Parse()

	// Create a new C2 server
	c2Server := server.NewC2Server()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting C2 server on %s", *addr)
		if err := c2Server.Start(*addr); err != nil {
			errCh <- err
		}
	}()

	// Wait for server to exit or receive a signal
	select {
	case err := <-errCh:
		log.Fatalf("Server error: %v", err)
	case sig := <-sigCh:
		log.Printf("Received signal: %v, shutting down...", sig)
	}
}