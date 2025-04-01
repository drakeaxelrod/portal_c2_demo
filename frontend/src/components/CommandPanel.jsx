import React, { useState } from 'react';
import { SendCommand } from "../../wailsjs/go/main/App";
import './CommandPanel.css';

const COMMAND_TYPES = [
  { value: "shell", label: "Shell Command" },
  { value: "upload", label: "Upload File" },
  { value: "download", label: "Download File" },
  { value: "screenshot", label: "Take Screenshot" },
  { value: "system", label: "System Info" },
  { value: "process", label: "Process List" },
];

const CommandPanel = ({ selectedAgent, onBack }) => {
  const [commandType, setCommandType] = useState("shell");
  const [commandPayload, setCommandPayload] = useState("");
  const [result, setResult] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSendCommand = async () => {
    if (!commandPayload.trim() && commandType === "shell") {
      setError("Please enter a command");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const response = await SendCommand(
        selectedAgent.id,
        commandType,
        commandPayload
      );

      setResult(response);
    } catch (err) {
      setError("Failed to send command: " + err.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="command-panel">
      <div className="command-panel-header">
        <button className="back-button" onClick={onBack}>← Back</button>
        <h2>Send Command to {selectedAgent.hostname}</h2>
      </div>

      <div className="agent-details">
        <div><strong>Agent ID:</strong> {selectedAgent.id}</div>
        <div><strong>Operating System:</strong> {selectedAgent.os} ({selectedAgent.arch})</div>
        <div><strong>IP Address:</strong> {selectedAgent.ip}</div>
        <div><strong>Username:</strong> {selectedAgent.username}</div>
      </div>

      <div className="command-form">
        <div className="form-group">
          <label htmlFor="command-type">Command Type:</label>
          <select
            id="command-type"
            value={commandType}
            onChange={(e) => setCommandType(e.target.value)}
          >
            {COMMAND_TYPES.map((type) => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="command-payload">Command:</label>
          <textarea
            id="command-payload"
            value={commandPayload}
            onChange={(e) => setCommandPayload(e.target.value)}
            placeholder={commandType === "shell" ? "Enter shell command..." : `Enter ${commandType} parameters...`}
            rows={3}
          />
        </div>

        {error && <div className="error-message">{error}</div>}

        <button
          className="send-button"
          onClick={handleSendCommand}
          disabled={isLoading}
        >
          {isLoading ? "Sending..." : "Send Command"}
        </button>
      </div>

      {result && (
        <div className="command-result">
          <h3>Command Result</h3>
          <pre>{result}</pre>
        </div>
      )}
    </div>
  );
};

export default CommandPanel;