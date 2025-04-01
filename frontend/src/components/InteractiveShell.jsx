import React, { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import './InteractiveShell.css';

const API_URL = 'http://localhost:8080';

const InteractiveShell = ({ agent: propAgent, onClose, isModal = false }) => {
  const { agentId } = useParams();
  const [agent, setAgent] = useState(propAgent);
  const [loading, setLoading] = useState(!propAgent);
  const [error, setError] = useState(null);
  const [connected, setConnected] = useState(false);
  const [connectionChecks, setConnectionChecks] = useState(0);

  const terminalRef = useRef(null);
  const terminalInstanceRef = useRef(null);
  const fitAddonRef = useRef(null);
  const wsRef = useRef(null);

  // Load agent data if not provided via props
  useEffect(() => {
    const loadAgent = async () => {
      if (propAgent) {
        setAgent(propAgent);
        setLoading(false);
        return;
      }

      if (!agentId) {
        setError('No agent ID provided');
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        const response = await fetch(`${API_URL}/api/agents`);

        if (!response.ok) {
          throw new Error(`Failed to fetch agents: ${response.status}`);
        }

        const agents = await response.json();

        // Find the agent with matching ID
        const foundAgent = agents.find(a =>
          (a.agent_id === agentId) || (a.id === agentId)
        );

        if (!foundAgent) {
          throw new Error(`Agent not found: ${agentId}`);
        }

        // Normalize agent data
        const normalizedAgent = {
          id: foundAgent.agent_id || foundAgent.id,
          hostname: foundAgent.hostname || 'Unknown Host',
          os: foundAgent.os || 'Unknown OS',
          arch: foundAgent.architecture || foundAgent.arch || 'Unknown Arch',
          ip: foundAgent.ip_address || foundAgent.ip || 'Unknown IP',
          raw: foundAgent
        };

        setAgent(normalizedAgent);
        setLoading(false);
      } catch (err) {
        console.error('Error loading agent:', err);
        setError(`Failed to load agent: ${err.message}`);
        setLoading(false);
      }
    };

    loadAgent();
  }, [propAgent, agentId]);

  // Initialize terminal and WebSocket connection
  useEffect(() => {
    if (!agent || !terminalRef.current || terminalInstanceRef.current) return;

    // Create terminal
    const terminal = new Terminal({
      cursorBlink: true,
      cursorStyle: 'bar',
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 14,
      lineHeight: 1.2,
      theme: {
        background: '#1e1e1e',
        foreground: '#f0f0f0',
        cursor: '#f0f0f0',
        cursorAccent: '#1e1e1e',
        selection: 'rgba(255, 255, 255, 0.3)',
      }
    });

    // Add fit addon to make terminal responsive
    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    fitAddonRef.current = fitAddon;
    terminalInstanceRef.current = terminal;

    // Open terminal in the DOM
    terminal.open(terminalRef.current);
    fitAddon.fit();

    // Display welcome message
    terminal.writeln('\x1b[1;32mConnecting to interactive shell...\x1b[0m');
    terminal.writeln(`Agent: ${agent.hostname} (${agent.id})`);
    terminal.writeln(`IP: ${agent.ip}`);
    terminal.writeln('');
    terminal.writeln('\x1b[90mEstablishing secure connection...\x1b[0m');

    // Connect to WebSocket for interactive shell
    const connectWebSocket = () => {
      // Close existing connection if any
      if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
        wsRef.current.close();
      }

      // Create new WebSocket connection
      const ws = new WebSocket(`ws://localhost:8080/api/agents/${agent.id}/shell`);
      wsRef.current = ws;

      ws.onopen = () => {
        terminal.writeln('\x1b[1;32mConnected to interactive shell!\x1b[0m');
        terminal.writeln('Type commands below. Use Ctrl+C to abort commands or exit.');
        terminal.writeln('');
        setConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.output) {
            terminal.write(data.output);
          }
        } catch (e) {
          // If not JSON, just write the raw data
          terminal.write(event.data);
        }
      };

      ws.onclose = () => {
        terminal.writeln('\r\n\x1b[1;31mConnection closed\x1b[0m');
        setConnected(false);

        // Attempt to reconnect if not too many attempts
        if (connectionChecks < 3) {
          terminal.writeln('\x1b[33mAttempting to reconnect...\x1b[0m');
          setTimeout(() => {
            setConnectionChecks(prev => prev + 1);
            connectWebSocket();
          }, 2000);
        } else {
          terminal.writeln('\x1b[31mFailed to establish connection after multiple attempts.\x1b[0m');
          terminal.writeln('Please try again later or check agent status.');
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        terminal.writeln('\r\n\x1b[1;31mConnection error\x1b[0m');
        terminal.writeln('WebSocket connection failed. The server might not support interactive shells via WebSocket.');
        terminal.writeln('You can still use the command terminal for basic interaction.');
      };

      // Handle terminal input and send to WebSocket
      terminal.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ input: data }));
        }
      });
    };

    // Start connection
    connectWebSocket();

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit();
      }
    };

    window.addEventListener('resize', handleResize);

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);

      if (wsRef.current) {
        wsRef.current.close();
      }

      if (terminalInstanceRef.current) {
        terminalInstanceRef.current.dispose();
        terminalInstanceRef.current = null;
      }
    };
  }, [agent]);

  // Handle closing the shell
  const handleClose = () => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    if (onClose) {
      onClose();
    }
  };

  if (loading) {
    return <div className="shell-loading">Loading interactive shell...</div>;
  }

  if (error) {
    return (
      <div className="shell-error">
        <h3>Error</h3>
        <p>{error}</p>
        <button onClick={handleClose}>Close</button>
      </div>
    );
  }

  return (
    <div className={`interactive-shell ${isModal ? 'is-modal' : ''}`}>
      <div className="shell-header">
        <div className="agent-info">
          <span className="hostname">{agent?.hostname || 'Unknown'}</span>
          <span className="separator">•</span>
          <span className="status">{connected ? 'Connected' : 'Disconnected'}</span>
        </div>
        {isModal && (
          <button className="close-button" onClick={handleClose}>
            &times;
          </button>
        )}
      </div>

      <div className="terminal-container" ref={terminalRef}></div>

      <div className="shell-footer">
        <div className="shell-instructions">
          <span>Press <code>Ctrl+C</code> to abort a command.</span>
          <span>Type <code>exit</code> to close the session.</span>
        </div>
      </div>
    </div>
  );
};

export default InteractiveShell;