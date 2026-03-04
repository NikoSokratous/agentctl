import { useCallback } from 'react';

interface NodePaletteProps {
  onAddNode: (type: string, position: { x: number; y: number }) => void;
}

const nodeTypes = [
  {
    type: 'agent',
    label: 'Agent',
    icon: '🤖',
    description: 'Execute an agent with a goal',
    color: '#3b82f6',
  },
  {
    type: 'tool',
    label: 'Tool',
    icon: '🔧',
    description: 'Execute a specific tool',
    color: '#8b5cf6',
  },
  {
    type: 'condition',
    label: 'Condition',
    icon: '❓',
    description: 'Conditional execution (CEL)',
    color: '#f59e0b',
  },
  {
    type: 'parallel',
    label: 'Parallel',
    icon: '⚡',
    description: 'Execute steps in parallel',
    color: '#10b981',
  },
  {
    type: 'join',
    label: 'Join',
    icon: '🔗',
    description: 'Wait for parallel steps',
    color: '#6366f1',
  },
];

export default function NodePalette({ onAddNode: _onAddNode }: NodePaletteProps) {
  const handleDragStart = useCallback(
    (event: React.DragEvent, nodeType: string) => {
      event.dataTransfer.setData('application/reactflow', nodeType);
      event.dataTransfer.effectAllowed = 'move';
    },
    []
  );

  return (
    <div className="node-palette">
      <h3 className="palette-title">Node Types</h3>
      <div className="palette-items">
        {nodeTypes.map((nodeType) => (
          <div
            key={nodeType.type}
            className="palette-item"
            draggable
            onDragStart={(e) => handleDragStart(e, nodeType.type)}
            style={{ borderLeftColor: nodeType.color }}
          >
            <div className="item-icon">{nodeType.icon}</div>
            <div className="item-content">
              <div className="item-label">{nodeType.label}</div>
              <div className="item-description">{nodeType.description}</div>
            </div>
          </div>
        ))}
      </div>

      <style>{`
        .node-palette {
          width: 250px;
          background: white;
          border-right: 1px solid #e5e7eb;
          padding: 1rem;
          overflow-y: auto;
        }

        .palette-title {
          font-size: 1.125rem;
          font-weight: 600;
          margin-bottom: 1rem;
          color: #111827;
        }

        .palette-items {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }

        .palette-item {
          display: flex;
          align-items: center;
          padding: 0.75rem;
          background: #f9fafb;
          border: 1px solid #e5e7eb;
          border-left: 3px solid;
          border-radius: 0.375rem;
          cursor: grab;
          transition: all 0.2s;
        }

        .palette-item:hover {
          background: #f3f4f6;
          transform: translateX(2px);
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
        }

        .palette-item:active {
          cursor: grabbing;
        }

        .item-icon {
          font-size: 1.5rem;
          margin-right: 0.75rem;
        }

        .item-content {
          flex: 1;
        }

        .item-label {
          font-weight: 500;
          color: #111827;
          margin-bottom: 0.125rem;
        }

        .item-description {
          font-size: 0.75rem;
          color: #6b7280;
        }
      `}</style>
    </div>
  );
}
