package main

import (
	"portal/pkg/server"
)

// API provides server methods for Wails frontend
type API struct {
	c2Server *server.C2Server
}

// NewAPI creates a new API instance
func NewAPI(c2Server *server.C2Server) *API {
	return &API{
		c2Server: c2Server,
	}
}

// GetAgents returns a list of all agents
func (a *API) GetAgents() []map[string]interface{} {
	agents := a.c2Server.GetAgentList()
	result := make([]map[string]interface{}, 0, len(agents))

	for _, agent := range agents {
		// Convert agent info to map for JSON serialization
		agentMap := map[string]interface{}{
			"agent_id":          agent.AgentId,
			"id":                agent.AgentId,            // Include both formats for flexibility
			"hostname":          agent.Hostname,
			"os":                agent.Os,
			"ip_address":        agent.IpAddress,
			"architecture":      agent.Architecture,
			"arch":              agent.Architecture,       // Include both formats for flexibility
			"username":          agent.Username,
			"registration_time": agent.RegistrationTime,
		}

		result = append(result, agentMap)
	}

	return result
}