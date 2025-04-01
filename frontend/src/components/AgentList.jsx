import React, { useState, useEffect } from 'react';
import { GetAgents } from "../../wailsjs/go/main/App";
import './AgentList.css';

const AgentList = ({ onSelectAgent }) => {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Function to load agents
  const loadAgents = async () => {
    try {
      setLoading(true);
      const agentList = await GetAgents();
      setAgents(agentList);
      setError(null);
    } catch (err) {
      setError("Failed to load agents: " + err.message);
    } finally {
      setLoading(false);
    }
  };

  // Load agents when component mounts
  useEffect(() => {
    loadAgents();

    // Refresh agent list every 10 seconds
    const interval = setInterval(loadAgents, 10000);

    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="agent-list">
      <div className="agent-list-header">
        <h2>Connected Agents</h2>
        <button onClick={loadAgents} disabled={loading}>
          {loading ? "Loading..." : "Refresh"}
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      {agents.length === 0 && !loading ? (
        <div className="no-agents">No agents connected</div>
      ) : (
        <div className="agent-table">
          <div className="agent-row header">
            <div>ID</div>
            <div>Hostname</div>
            <div>OS</div>
            <div>IP</div>
            <div>Actions</div>
          </div>

          {agents.map((agent) => (
            <div
              key={agent.id}
              className="agent-row"
            >
              <div>{agent.id}</div>
              <div>{agent.hostname}</div>
              <div>{agent.os} ({agent.arch})</div>
              <div>{agent.ip}</div>
              <div>
                <button onClick={() => onSelectAgent(agent)}>
                  Command
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AgentList;