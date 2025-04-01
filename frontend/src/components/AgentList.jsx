import React, { useState, useEffect, useCallback } from 'react';
// Import fetch instead of Wails functions
// import { GetAgents } from "../../wailsjs/go/main/App";
import './AgentList.css';

const API_URL = 'http://localhost:8080';

const AgentList = ({ onSelectAgent, onAgentCountUpdate }) => {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterOS, setFilterOS] = useState('all');
  const [refreshInterval, setRefreshInterval] = useState(2000); // 2 seconds default

  // Function to load agents from API
  const loadAgents = useCallback(async () => {
    try {
      setLoading(true);

      const response = await fetch(`${API_URL}/api/agents`);

      if (!response.ok) {
        throw new Error(`API returned status ${response.status}`);
      }

      const data = await response.json();
      console.log('API Response:', data);

      // Normalize agent data to handle different field naming conventions
      const normalizedAgents = data.map(agent => ({
        id: agent.agent_id || agent.id || '',
        hostname: agent.hostname || 'Unknown Host',
        os: agent.os || 'Unknown OS',
        arch: agent.architecture || agent.arch || 'Unknown Arch',
        ip: agent.ip_address || agent.ip || 'Unknown IP',
        username: agent.username || 'Unknown User',
        // Use registration_time for last seen time
        lastSeen: agent.registration_time ? new Date(agent.registration_time * 1000) : null,
        // Calculate status based on last seen time
        status: calculateAgentStatus(agent.registration_time),
        // Store the raw data for debugging
        raw: agent
      }));

      console.log('Normalized Agents:', normalizedAgents);
      setAgents(normalizedAgents);

      // Update connected agent count
      if (onAgentCountUpdate) {
        const connectedCount = normalizedAgents.filter(
          agent => agent.status === 'active' || agent.status === 'idle'
        ).length;
        onAgentCountUpdate(connectedCount);
      }

      setError(null);
    } catch (err) {
      console.error('Error loading agents:', err);
      setError(`Failed to load agents: ${err.message}`);
    } finally {
      setLoading(false);
    }
  }, [onAgentCountUpdate]);

  // Calculate agent status based on last seen time
  const calculateAgentStatus = (timestamp) => {
    if (!timestamp) return 'offline';

    const now = new Date();
    const lastSeen = new Date(timestamp * 1000);
    const diffSeconds = (now - lastSeen) / 1000;

    if (diffSeconds < 30) return 'active';
    if (diffSeconds < 120) return 'idle';
    return 'offline';
  };

  // Format time since last seen
  const formatTimeSince = (date) => {
    if (!date) return 'Never';

    const now = new Date();
    const diffSeconds = Math.floor((now - date) / 1000);

    if (diffSeconds < 60) return `${diffSeconds} sec ago`;
    if (diffSeconds < 3600) return `${Math.floor(diffSeconds / 60)} min ago`;
    if (diffSeconds < 86400) return `${Math.floor(diffSeconds / 3600)} hrs ago`;
    return `${Math.floor(diffSeconds / 86400)} days ago`;
  };

  // Load agents on component mount and set up refresh interval
  useEffect(() => {
    // Initial load
    loadAgents();

    // Set up interval for refreshing
    const interval = setInterval(loadAgents, refreshInterval);

    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, [loadAgents, refreshInterval]);

  // Handle OS icon display
  const getOSIcon = (os) => {
    const osLower = (os || '').toLowerCase();
    if (osLower.includes('windows')) return '🪟';
    if (osLower.includes('mac') || osLower.includes('darwin')) return '🍎';
    if (osLower.includes('linux')) return '🐧';
    return '💻';
  };

  // Get readable status text
  const getStatusText = (status) => {
    switch (status) {
      case 'active': return 'Online';
      case 'idle': return 'Idle';
      case 'offline': return 'Offline';
      default: return 'Unknown';
    }
  };

  // Filter agents based on search term and OS filter
  const filteredAgents = agents.filter(agent => {
    // Filter by search term
    const matchesSearch =
      agent.hostname.toLowerCase().includes(searchTerm.toLowerCase()) ||
      agent.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      agent.ip.toLowerCase().includes(searchTerm.toLowerCase());

    // Filter by OS
    const matchesOS =
      filterOS === 'all' ||
      (filterOS === 'windows' && agent.os.toLowerCase().includes('windows')) ||
      (filterOS === 'mac' && (agent.os.toLowerCase().includes('mac') || agent.os.toLowerCase().includes('darwin'))) ||
      (filterOS === 'linux' && agent.os.toLowerCase().includes('linux'));

    return matchesSearch && matchesOS;
  });

  return (
    <div className="agent-list-container">
      <div className="list-controls">
        <div className="search-filter">
          <input
            type="text"
            placeholder="Search agents..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
          <select
            value={filterOS}
            onChange={(e) => setFilterOS(e.target.value)}
            className="os-filter"
          >
            <option value="all">All OS</option>
            <option value="windows">Windows</option>
            <option value="mac">macOS</option>
            <option value="linux">Linux</option>
          </select>
          <button
            className="refresh-button"
            onClick={loadAgents}
            disabled={loading}
          >
            {loading ? 'Loading...' : 'Refresh'}
          </button>
        </div>
        <div className="agent-count">
          {filteredAgents.length} agent{filteredAgents.length !== 1 ? 's' : ''}
        </div>
      </div>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      <div className="agent-list">
        {filteredAgents.length === 0 ? (
          <div className="no-agents">
            {loading ? 'Loading agents...' : 'No agents found'}
          </div>
        ) : (
          filteredAgents.map((agent) => (
            <div
              key={agent.id}
              className={`agent-row ${agent.status === 'offline' ? 'offline' : ''}`}
            >
              <div className="agent-status">
                <span className={`status-indicator-dot ${agent.status || 'offline'}`}></span>
                <span className="status-text">{getStatusText(agent.status)}</span>
              </div>
              <div className="agent-os">
                <span className="os-icon">{getOSIcon(agent.os)}</span>
                <span>{agent.os}</span>
                <span className="arch">({agent.arch})</span>
              </div>
              <div className="agent-host">
                <div className="hostname">{agent.hostname}</div>
                <div className="ip-address">{agent.ip}</div>
                <div className="agent-id">{agent.id}</div>
              </div>
              <div className="agent-registered">
                {formatTimeSince(agent.lastSeen)}
              </div>
              <div className="agent-actions">
                <button
                  className="action-button command"
                  onClick={() => onSelectAgent(agent)}
                  disabled={agent.status === 'offline'}
                >
                  <span>Terminal</span>
                </button>
                <button className="action-button info">
                  <span>Info</span>
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default AgentList;