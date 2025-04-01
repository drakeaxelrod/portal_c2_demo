import React, { useState } from 'react';
import AgentList from './AgentList';
import CommandPanel from './CommandPanel';
import './Dashboard.css';

const Dashboard = () => {
  const [selectedAgent, setSelectedAgent] = useState(null);

  const handleSelectAgent = (agent) => {
    setSelectedAgent(agent);
  };

  const handleBackToAgents = () => {
    setSelectedAgent(null);
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Portal C2 Framework</h1>
      </header>

      <div className="dashboard-content">
        {selectedAgent ? (
          <CommandPanel
            selectedAgent={selectedAgent}
            onBack={handleBackToAgents}
          />
        ) : (
          <AgentList onSelectAgent={handleSelectAgent} />
        )}
      </div>

      <footer className="dashboard-footer">
        <p>Portal C2 Framework v1.0</p>
      </footer>
    </div>
  );
};

export default Dashboard;