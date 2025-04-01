import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Dashboard from './components/Dashboard';
import AgentCommander from './components/AgentCommander';
import './App.css';

function App() {
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [agentCount, setAgentCount] = useState(0);

  const handleSelectAgent = (agent) => {
    setSelectedAgent(agent);
  };

  const handleAgentCountUpdate = (count) => {
    setAgentCount(count);
  };

  return (
    <Router>
      <div className="app">
        <Routes>
          <Route
            path="/"
            element={
              <Dashboard
                onSelectAgent={handleSelectAgent}
                agentCount={agentCount}
                onAgentCountUpdate={handleAgentCountUpdate}
              />
            }
          />
          <Route
            path="/agent/:id"
            element={
              <AgentCommander
                agent={selectedAgent}
              />
            }
          />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
