import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Greet } from '../../wailsjs/go/main/App';
import AgentList from './AgentList';
import './Dashboard.css';

const Dashboard = ({ onSelectAgent, agentCount, onAgentCountUpdate }) => {
  const [serverStatus, setServerStatus] = useState('unknown');
  const [statusMessage, setStatusMessage] = useState('Checking server status...');
  const navigate = useNavigate();

  // Check server status on component mount
  useEffect(() => {
    const checkServerStatus = async () => {
      try {
        // Try to call a simple method to check if the server is responding
        const greeting = await Greet("User");
        console.log("Server response:", greeting);
        setServerStatus('online');
        setStatusMessage('Server online');
      } catch (error) {
        console.error("Server status check failed:", error);
        setServerStatus('offline');
        setStatusMessage('Server connection failed');
      }
    };

    checkServerStatus();

    // Set up periodic status check every 30 seconds
    const interval = setInterval(checkServerStatus, 30000);

    return () => clearInterval(interval);
  }, []);

  const handleSelectAgent = (agent) => {
    onSelectAgent(agent);
    navigate(`/agent/${agent.id}`);
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="dashboard-title">
          <h1>Portal C2 Framework</h1>
          <div className="dashboard-status">
            <div className="status-indicator">
              <span className={`status-dot ${serverStatus}`}></span>
              <span className="status-text">{statusMessage}</span>
            </div>
            <div className="agents-counter">
              <span className="agents-count">{agentCount || 0}</span>
              <span className="agents-label">Agent{agentCount !== 1 ? 's' : ''} Connected</span>
            </div>
          </div>
        </div>
        <div className="dashboard-actions">
          <button className="action-button">Settings</button>
          <button className="action-button primary">New Agent</button>
        </div>
      </header>

      <main className="dashboard-content">
        <section className="dashboard-section">
          <AgentList onSelectAgent={handleSelectAgent} onAgentCountUpdate={onAgentCountUpdate} />
        </section>
      </main>

      <footer className="dashboard-footer">
        <div className="footer-links">
          <a href="#">About</a>
          <a href="#">Documentation</a>
        </div>
        <div className="footer-info">
          <span>Portal C2 Framework v1.0.0</span>
        </div>
      </footer>
    </div>
  );
};

export default Dashboard;