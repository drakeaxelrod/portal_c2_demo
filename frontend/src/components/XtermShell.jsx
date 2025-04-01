import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { SendCommand } from '../../wailsjs/go/main/App';
import '@xterm/xterm/css/xterm.css';
import './XtermShell.css';

const XtermShell = ({ agent }) => {
  const terminalRef = useRef(null);
  const terminalContainerRef = useRef(null);
  const [isConnected, setIsConnected] = useState(false);
  const [currentCommand, setCurrentCommand] = useState('');
  const [commandHistory, setCommandHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);

  // Create and configure terminal
  useEffect(() => {
    if (!agent || !terminalContainerRef.current) return;

    // Initialize terminal
    const terminal = new Terminal({
      fontFamily: '"Cascadia Code", "Fira Code", monospace',
      fontSize: 14,
      theme: {
        background: '#1e293b',
        foreground: '#f8fafc',
        cursor: '#3b82f6',
        black: '#1e293b',
        red: '#ef4444',
        green: '#10b981',
        yellow: '#f59e0b',
        blue: '#3b82f6',
        magenta: '#a855f7',
        cyan: '#06b6d4',
        white: '#f8fafc',
        brightBlack: '#475569',
        brightRed: '#f87171',
        brightGreen: '#34d399',
        brightYellow: '#fbbf24',
        brightBlue: '#60a5fa',
        brightMagenta: '#c084fc',
        brightCyan: '#22d3ee',
        brightWhite: '#f1f5f9',
      },
      cursorBlink: true,
      scrollback: 1000,
      allowTransparency: true,
    });

    // Set up addons
    const fitAddon = new FitAddon();
    const webLinksAddon = new WebLinksAddon();

    terminal.loadAddon(fitAddon);
    terminal.loadAddon(webLinksAddon);

    // Open terminal
    terminal.open(terminalContainerRef.current);
    terminalRef.current = terminal;

    // Fit to container
    fitAddon.fit();

    // Display welcome message
    terminal.writeln('\x1b[1;34m=== Portal C2 Framework - Interactive Shell ===\x1b[0m');
    terminal.writeln(`\x1b[1;32mConnected to agent: \x1b[1;33m${agent.hostname} (${agent.id})\x1b[0m`);
    terminal.writeln(`\x1b[32mOS: \x1b[0m${agent.os} ${agent.arch}`);
    terminal.writeln(`\x1b[32mIP: \x1b[0m${agent.ip}`);
    terminal.writeln('\x1b[90m-----------------------------------------\x1b[0m');
    terminal.writeln('Type commands and press Enter to execute. Type "exit" to close the session.');
    terminal.writeln('Type "shell" to start an interactive shell in a new window.');
    terminal.writeln('');

    // Write prompt
    writePrompt(terminal);
    setIsConnected(true);

    // Handle terminal resizing
    const handleResize = () => {
      fitAddon.fit();
    };

    window.addEventListener('resize', handleResize);

    // Handle input
    terminal.onData((data) => {
      // Handle control characters
      const code = data.charCodeAt(0);

      // If Enter is pressed (CR or LF)
      if (data === '\r' || data === '\n') {
        handleCommand(terminal);
      }
      // Handle backspace
      else if (data === '\x7f') {
        if (currentCommand.length > 0) {
          terminal.write('\b \b');
          setCurrentCommand(prev => prev.slice(0, -1));
        }
      }
      // Handle arrow up (navigate command history)
      else if (code === 27 && data.length === 3 && data.charCodeAt(1) === 91 && data.charCodeAt(2) === 65) {
        handleArrowUp(terminal);
      }
      // Handle arrow down (navigate command history)
      else if (code === 27 && data.length === 3 && data.charCodeAt(1) === 91 && data.charCodeAt(2) === 66) {
        handleArrowDown(terminal);
      }
      // Handle printable characters
      else if (code >= 32 && code !== 127) {
        terminal.write(data);
        setCurrentCommand(prev => prev + data);
      }
    });

    // Clean up on unmount
    return () => {
      window.removeEventListener('resize', handleResize);
      if (terminalRef.current) {
        terminalRef.current.dispose();
      }
      setIsConnected(false);
    };
  }, [agent]);

  // Write prompt to terminal
  const writePrompt = (terminal) => {
    terminal.write('\r\n\x1b[1;36m$ \x1b[0m');
  };

  // Open interactive shell in new window
  const openInteractiveShell = async (agentId) => {
    try {
      const shellUrl = `/interactive-shell/${agentId}`;
      window.open(shellUrl, `shell_${agentId}`, 'width=800,height=600,resizable=yes');
      return true;
    } catch (error) {
      console.error("Failed to open interactive shell window:", error);
      return false;
    }
  };

  // Handle command execution
  const handleCommand = async (terminal) => {
    if (!agent || !isConnected || !terminal) return;

    const command = currentCommand.trim();

    // Add to command history and reset current command
    terminal.writeln('');

    if (command === '') {
      writePrompt(terminal);
      return;
    }

    // Update command history
    setCommandHistory(prev => [command, ...prev.slice(0, 49)]);
    setHistoryIndex(-1);

    // Handle exit command
    if (command.toLowerCase() === 'exit' || command.toLowerCase() === 'quit') {
      terminal.writeln('\r\n\x1b[1;33mClosing session...\x1b[0m');
      setIsConnected(false);
      return;
    }

    // Handle clear command
    if (command.toLowerCase() === 'clear' || command.toLowerCase() === 'cls') {
      terminal.clear();
      setCurrentCommand('');
      writePrompt(terminal);
      return;
    }

    // Handle interactive shell command
    if (command.toLowerCase() === 'shell') {
      terminal.writeln('\r\n\x1b[1;32mOpening interactive shell in new window...\x1b[0m');
      const success = await openInteractiveShell(agent.id);
      if (!success) {
        terminal.writeln('\r\n\x1b[1;31mFailed to open interactive shell window. Check browser pop-up settings.\x1b[0m');
      }
      setCurrentCommand('');
      writePrompt(terminal);
      return;
    }

    // Execute the command
    try {
      terminal.writeln(`\x1b[90mExecuting: ${command}\x1b[0m`);

      // Send command to the server/agent
      const result = await SendCommand(agent.id, 'shell', command);

      // Display result
      if (result) {
        // Process ANSI color codes if present, otherwise just write the output
        terminal.writeln('\r\n' + result);
      } else {
        terminal.writeln('\r\n\x1b[90m(Command executed with no output)\x1b[0m');
      }
    } catch (error) {
      terminal.writeln(`\r\n\x1b[1;31mError: ${error.message || 'Command execution failed'}\x1b[0m`);
    }

    setCurrentCommand('');
    writePrompt(terminal);
  };

  // Handle up arrow key (previous command)
  const handleArrowUp = (terminal) => {
    if (commandHistory.length === 0) return;

    const newIndex = historyIndex < commandHistory.length - 1 ? historyIndex + 1 : historyIndex;
    if (newIndex >= 0 && newIndex < commandHistory.length) {
      // Clear current line
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[1;36m$ \x1b[0m');

      // Write the historical command
      const historicalCommand = commandHistory[newIndex];
      terminal.write(historicalCommand);

      setCurrentCommand(historicalCommand);
      setHistoryIndex(newIndex);
    }
  };

  // Handle down arrow key (next command)
  const handleArrowDown = (terminal) => {
    if (commandHistory.length === 0) return;

    if (historyIndex > 0) {
      const newIndex = historyIndex - 1;

      // Clear current line
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[1;36m$ \x1b[0m');

      // Write the historical command
      const historicalCommand = commandHistory[newIndex];
      terminal.write(historicalCommand);

      setCurrentCommand(historicalCommand);
      setHistoryIndex(newIndex);
    } else if (historyIndex === 0) {
      // Clear current line
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[1;36m$ \x1b[0m');

      setCurrentCommand('');
      setHistoryIndex(-1);
    }
  };

  return (
    <div className="xterm-shell-container">
      <div
        ref={terminalContainerRef}
        className="xterm-container"
        style={{ width: '100%', height: '100%' }}
      />
    </div>
  );
};

export default XtermShell;