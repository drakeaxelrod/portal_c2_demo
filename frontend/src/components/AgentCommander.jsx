import React, { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import InteractiveShell from './InteractiveShell';
import './AgentCommander.css';

const API_URL = 'http://localhost:8080';

const AgentCommander = ({ agent: initialAgent }) => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [agent, setAgent] = useState(initialAgent);
  const [commandInput, setCommandInput] = useState('');
  const [history, setHistory] = useState([]);
  const [isExecuting, setIsExecuting] = useState(false);
  const [showShellModal, setShowShellModal] = useState(false);
  const [error, setError] = useState(null);
  const commandEndRef = useRef(null);

  // Load agent data if not provided in props
  useEffect(() => {
    const loadAgent = async () => {
      if (initialAgent) {
        setAgent(initialAgent);
        return;
      }

      if (!id) {
        navigate('/');
        return;
      }

      try {
        const response = await fetch(`${API_URL}/api/agents`);
        if (!response.ok) {
          throw new Error(`Failed to load agents: ${response.status}`);
        }

        const agents = await response.json();
        console.log('Loaded agents:', agents);

        // Find the agent with matching ID
        const foundAgent = agents.find(a =>
          (a.agent_id === id) || (a.id === id)
        );

        if (!foundAgent) {
          throw new Error(`Agent not found: ${id}`);
        }

        // Normalize agent data
        const normalizedAgent = {
          id: foundAgent.agent_id || foundAgent.id,
          hostname: foundAgent.hostname || 'Unknown Host',
          os: foundAgent.os || 'Unknown OS',
          arch: foundAgent.architecture || foundAgent.arch || 'Unknown Arch',
          ip: foundAgent.ip_address || foundAgent.ip || 'Unknown IP',
          username: foundAgent.username || 'Unknown User',
          lastSeen: foundAgent.registration_time ? new Date(foundAgent.registration_time * 1000) : null,
          raw: foundAgent
        };

        setAgent(normalizedAgent);

      } catch (err) {
        console.error('Error loading agent:', err);
        setError(`Failed to load agent: ${err.message}`);
      }
    };

    loadAgent();
  }, [id, initialAgent, navigate]);

  // Auto-scroll to bottom of command history
  const scrollToBottom = () => {
    commandEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [history]);

  // Add message to history
  const appendToHistory = (content, type) => {
    const entry = {
      type, // 'command', 'response', or 'error'
      content,
      timestamp: new Date().toISOString()
    };
    setHistory(prev => [...prev, entry]);
  };

  // Execute a command
  const executeCommand = async (command) => {
    if (!command || isExecuting || !agent) return;

    try {
      setIsExecuting(true);

      // Add command to history
      appendToHistory(command, 'command');
      setCommandInput('');

      // Send command to the server
      const response = await fetch(`${API_URL}/api/agents/${agent.id}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          command,
          type: 'shell',
        }),
      });

      // Handle error responses
      if (!response.ok) {
        if (response.status === 404) {
          appendToHistory(`Error: Agent ${agent.id} not found`, 'error');
        } else {
          const errorText = await response.text();
          appendToHistory(`Error: ${errorText}`, 'error');
        }
        return;
      }

      // Process successful response
      const result = await response.json();

      if (result.success) {
        appendToHistory(result.result, 'response');
      } else {
        appendToHistory(`Error: ${result.error || 'Command failed'}`, 'error');
      }

    } catch (err) {
      console.error('Error executing command:', err);
      appendToHistory(`Error: ${err.message}`, 'error');
    } finally {
      setIsExecuting(false);
    }
  };

  // Handle form submission
  const handleCommandSubmit = async (e) => {
    e.preventDefault();
    if (!commandInput.trim() || !agent) return;
    await executeCommand(commandInput);
  };

  // Execute quick command
  const handleQuickCommand = async (command) => {
    if (isExecuting) return;
    await executeCommand(command);
  };

  // Interactive shell handling
  const openInteractiveShell = () => {
    if (!agent) return;
    setShowShellModal(true);
  };

  const closeInteractiveShell = () => {
    setShowShellModal(false);
  };

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // ESC key to close modal
      if (e.key === 'Escape' && showShellModal) {
        closeInteractiveShell();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [showShellModal]);

  // Get OS icon based on agent OS
  const getOSIcon = (os) => {
    const osLower = (os || '').toLowerCase();
    if (osLower.includes('windows')) return '🪟';
    if (osLower.includes('mac') || osLower.includes('darwin')) return '🍎';
    if (osLower.includes('linux')) return '🐧';
    return '💻';
  };

  // Common quick commands
  const quickCommands = [
    { label: 'System Info', command: 'sysinfo' },
    { label: 'List Files', command: 'ls -la' },
    { label: 'Network Info', command: 'ifconfig || ipconfig' },
    { label: 'Current User', command: 'whoami' }
  ];

  // Handle loading and error states
  if (error) {
    return (
      <div className="agent-error">
        <h3>Error</h3>
        <p>{error}</p>
        <button onClick={() => navigate('/')}>Back to Agent List</button>
      </div>
    );
  }

  if (!agent) {
    return <div className="agent-loading">Loading agent details...</div>;
  }

  return (
    <div className="agent-commander">
      <header className="commander-header">
        <div className="agent-info">
          <button className="back-button" onClick={() => navigate('/')}>
            ← Back
          </button>
          <div className="agent-badge">
            <span className="os-icon">{getOSIcon(agent.os)}</span>
            <span className="agent-name">{agent.hostname}</span>
            <span className="agent-detail">{agent.os} ({agent.arch})</span>
            <span className="agent-ip">{agent.ip}</span>
          </div>
        </div>
        <div className="commander-actions">
          <button
            className="action-button primary"
            onClick={openInteractiveShell}
          >
            Interactive Shell
          </button>
          <button className="action-button danger">Terminate</button>
        </div>
      </header>

      <div className="command-terminal">
        <div className="terminal-header">
          <span>Terminal Session</span>
        </div>
        <div className="terminal-output">
          <div className="welcome-message">
            <h3>Connected to Agent {agent.id}</h3>
            <p>Type commands below or select a quick command to interact with the agent.</p>
            <p>For a fully interactive shell experience, click the "Interactive Shell" button above.</p>
          </div>

          {history.map((entry, index) => (
            <div key={index} className={`terminal-entry ${entry.type}`}>
              {entry.type === 'command' ? (
                <div className="command-entry">
                  <span className="prompt">$ </span>
                  <span className="command-text">{entry.content}</span>
                </div>
              ) : entry.type === 'response' ? (
                <div className="response-entry">
                  <pre>{entry.content}</pre>
                </div>
              ) : (
                <div className="error-entry">
                  <pre>{entry.content}</pre>
                </div>
              )}
            </div>
          ))}

          <div ref={commandEndRef} />
        </div>

        <form className="command-input-form" onSubmit={handleCommandSubmit}>
          <div className="prompt">$</div>
          <input
            type="text"
            value={commandInput}
            onChange={(e) => setCommandInput(e.target.value)}
            placeholder="Enter command..."
            disabled={isExecuting}
            autoFocus
          />
          <button
            type="submit"
            disabled={isExecuting || !commandInput.trim()}
          >
            {isExecuting ? 'Executing...' : 'Send'}
          </button>
        </form>
      </div>

      <div className="quick-commands">
        <div className="quick-commands-header">Quick Commands</div>
        <div className="quick-commands-list">
          {quickCommands.map((cmd, index) => (
            <button
              key={index}
              className="quick-command-button"
              onClick={() => handleQuickCommand(cmd.command)}
              disabled={isExecuting}
            >
              {cmd.label}
            </button>
          ))}
        </div>
      </div>

      {/* Interactive Shell Modal */}
      {showShellModal && (
        <div className="modal-overlay" onClick={closeInteractiveShell}>
          <div className="interactive-shell-modal" onClick={(e) => e.stopPropagation()}>
            <InteractiveShell
              agent={agent}
              onClose={closeInteractiveShell}
              isModal={true}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default AgentCommander;