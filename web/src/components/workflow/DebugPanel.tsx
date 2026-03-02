import { useState, useEffect } from 'react';

interface DebugPanelProps {
  workflowId: string;
  isActive: boolean;
  onClose: () => void;
}

interface DebugState {
  workflow_id: string;
  current_step: string;
  paused: boolean;
  variables: Record<string, any>;
  breakpoints: Record<string, any>;
  history: string[];
}

export default function DebugPanel({ workflowId, isActive, onClose }: DebugPanelProps) {
  const [debugState, setDebugState] = useState<DebugState | null>(null);
  const [breakpoints, setBreakpoints] = useState<string[]>([]);
  const [newBreakpoint, setNewBreakpoint] = useState({ nodeId: '', condition: '' });

  useEffect(() => {
    if (!isActive) return;

    const interval = setInterval(() => {
      fetchDebugState();
    }, 1000);

    return () => clearInterval(interval);
  }, [isActive, workflowId]);

  const fetchDebugState = async () => {
    try {
      const response = await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/state`);
      if (response.ok) {
        const state = await response.json();
        setDebugState(state);
      }
    } catch (error) {
      console.error('Failed to fetch debug state:', error);
    }
  };

  const handleStart = async () => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/start`, {
        method: 'POST',
      });
      fetchDebugState();
    } catch (error) {
      console.error('Failed to start debug:', error);
    }
  };

  const handleStop = async () => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/stop`, {
        method: 'POST',
      });
      onClose();
    } catch (error) {
      console.error('Failed to stop debug:', error);
    }
  };

  const handlePause = async () => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/pause`, {
        method: 'POST',
      });
      fetchDebugState();
    } catch (error) {
      console.error('Failed to pause:', error);
    }
  };

  const handleResume = async () => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/resume`, {
        method: 'POST',
      });
      fetchDebugState();
    } catch (error) {
      console.error('Failed to resume:', error);
    }
  };

  const handleStep = async () => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/step`, {
        method: 'POST',
      });
      fetchDebugState();
    } catch (error) {
      console.error('Failed to step:', error);
    }
  };

  const handleAddBreakpoint = async () => {
    if (!newBreakpoint.nodeId) return;

    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/breakpoint`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newBreakpoint),
      });
      setBreakpoints([...breakpoints, newBreakpoint.nodeId]);
      setNewBreakpoint({ nodeId: '', condition: '' });
      fetchDebugState();
    } catch (error) {
      console.error('Failed to add breakpoint:', error);
    }
  };

  const handleRemoveBreakpoint = async (nodeId: string) => {
    try {
      await fetch(`http://localhost:8080/v1/workflows/${workflowId}/debug/breakpoint/${nodeId}`, {
        method: 'DELETE',
      });
      setBreakpoints(breakpoints.filter((bp) => bp !== nodeId));
      fetchDebugState();
    } catch (error) {
      console.error('Failed to remove breakpoint:', error);
    }
  };

  if (!isActive) return null;

  return (
    <div className="debug-panel">
      <div className="panel-header">
        <h3>🐛 Workflow Debugger</h3>
        <button className="close-btn" onClick={onClose}>
          ✕
        </button>
      </div>

      <div className="panel-content">
        <div className="debug-controls">
          <button className="control-btn" onClick={handleStart}>
            ▶️ Start
          </button>
          <button className="control-btn" onClick={handleStop}>
            ⏹️ Stop
          </button>
          {debugState?.paused ? (
            <>
              <button className="control-btn primary" onClick={handleResume}>
                ▶️ Resume
              </button>
              <button className="control-btn" onClick={handleStep}>
                ⏭️ Step
              </button>
            </>
          ) : (
            <button className="control-btn" onClick={handlePause}>
              ⏸️ Pause
            </button>
          )}
        </div>

        {debugState && (
          <div className="debug-info">
            <div className="info-section">
              <h4>Current State</h4>
              <div className="info-item">
                <span className="label">Step:</span>
                <span className="value">{debugState.current_step || 'N/A'}</span>
              </div>
              <div className="info-item">
                <span className="label">Status:</span>
                <span className={`status ${debugState.paused ? 'paused' : 'running'}`}>
                  {debugState.paused ? '⏸️ Paused' : '▶️ Running'}
                </span>
              </div>
            </div>

            <div className="info-section">
              <h4>Breakpoints</h4>
              <div className="breakpoint-list">
                {Object.keys(debugState.breakpoints || {}).map((nodeId) => (
                  <div key={nodeId} className="breakpoint-item">
                    <span>{nodeId}</span>
                    <button
                      className="remove-btn"
                      onClick={() => handleRemoveBreakpoint(nodeId)}
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
              <div className="add-breakpoint">
                <input
                  type="text"
                  placeholder="Node ID"
                  value={newBreakpoint.nodeId}
                  onChange={(e) =>
                    setNewBreakpoint({ ...newBreakpoint, nodeId: e.target.value })
                  }
                />
                <input
                  type="text"
                  placeholder="Condition (optional)"
                  value={newBreakpoint.condition}
                  onChange={(e) =>
                    setNewBreakpoint({ ...newBreakpoint, condition: e.target.value })
                  }
                />
                <button className="add-btn" onClick={handleAddBreakpoint}>
                  + Add
                </button>
              </div>
            </div>

            <div className="info-section">
              <h4>Variables</h4>
              <pre className="variables-view">
                {JSON.stringify(debugState.variables, null, 2)}
              </pre>
            </div>

            <div className="info-section">
              <h4>Execution History</h4>
              <div className="history-list">
                {debugState.history?.map((step, idx) => (
                  <div key={idx} className="history-item">
                    {idx + 1}. {step}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      <style>{`
        .debug-panel {
          position: fixed;
          right: 0;
          top: 0;
          bottom: 0;
          width: 400px;
          background: white;
          border-left: 1px solid #e5e7eb;
          box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
          display: flex;
          flex-direction: column;
          z-index: 100;
        }

        .panel-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1rem;
          border-bottom: 1px solid #e5e7eb;
          background: #f9fafb;
        }

        .panel-header h3 {
          margin: 0;
          font-size: 1.125rem;
        }

        .close-btn {
          background: none;
          border: none;
          font-size: 1.25rem;
          cursor: pointer;
          padding: 0.25rem;
        }

        .panel-content {
          flex: 1;
          overflow-y: auto;
          padding: 1rem;
        }

        .debug-controls {
          display: flex;
          gap: 0.5rem;
          margin-bottom: 1.5rem;
          flex-wrap: wrap;
        }

        .control-btn {
          padding: 0.5rem 0.75rem;
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 0.375rem;
          cursor: pointer;
          font-size: 0.875rem;
          transition: all 0.2s;
        }

        .control-btn:hover {
          background: #e5e7eb;
        }

        .control-btn.primary {
          background: #3b82f6;
          color: white;
          border-color: #3b82f6;
        }

        .control-btn.primary:hover {
          background: #2563eb;
        }

        .debug-info {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }

        .info-section {
          background: #f9fafb;
          border: 1px solid #e5e7eb;
          border-radius: 0.375rem;
          padding: 1rem;
        }

        .info-section h4 {
          margin: 0 0 0.75rem 0;
          font-size: 0.875rem;
          font-weight: 600;
          color: #374151;
        }

        .info-item {
          display: flex;
          justify-content: space-between;
          margin-bottom: 0.5rem;
          font-size: 0.875rem;
        }

        .info-item .label {
          color: #6b7280;
        }

        .info-item .value {
          font-weight: 500;
        }

        .status {
          padding: 0.25rem 0.5rem;
          border-radius: 0.25rem;
          font-size: 0.75rem;
          font-weight: 500;
        }

        .status.running {
          background: #d1fae5;
          color: #065f46;
        }

        .status.paused {
          background: #fef3c7;
          color: #92400e;
        }

        .breakpoint-list,
        .history-list {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }

        .breakpoint-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0.5rem;
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.25rem;
          font-size: 0.875rem;
        }

        .remove-btn {
          background: #fee2e2;
          color: #991b1b;
          border: none;
          padding: 0.125rem 0.375rem;
          border-radius: 0.25rem;
          cursor: pointer;
          font-size: 0.75rem;
        }

        .add-breakpoint {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
          margin-top: 0.75rem;
        }

        .add-breakpoint input {
          padding: 0.5rem;
          border: 1px solid #d1d5db;
          border-radius: 0.25rem;
          font-size: 0.875rem;
        }

        .add-btn {
          padding: 0.5rem;
          background: #3b82f6;
          color: white;
          border: none;
          border-radius: 0.25rem;
          cursor: pointer;
          font-weight: 500;
        }

        .add-btn:hover {
          background: #2563eb;
        }

        .variables-view {
          background: white;
          padding: 0.75rem;
          border-radius: 0.25rem;
          font-size: 0.75rem;
          overflow-x: auto;
          margin: 0;
        }

        .history-item {
          padding: 0.5rem;
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.25rem;
          font-size: 0.875rem;
        }
      `}</style>
    </div>
  );
}
