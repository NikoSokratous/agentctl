import { memo } from 'react';
import { Handle, Position } from 'reactflow';

interface CustomNodeProps {
  data: {
    label: string;
    agent?: string;
    goal?: string;
    condition?: string;
  };
  selected: boolean;
}

const getNodeStyle = (type: string) => {
  const styles: Record<string, any> = {
    agent: { background: '#3b82f6', color: 'white', icon: '🤖' },
    tool: { background: '#8b5cf6', color: 'white', icon: '🔧' },
    condition: { background: '#f59e0b', color: 'white', icon: '❓' },
    parallel: { background: '#10b981', color: 'white', icon: '⚡' },
    join: { background: '#6366f1', color: 'white', icon: '🔗' },
  };
  return styles[type] || styles.agent;
};

function CustomNode({ data, selected }: CustomNodeProps) {
  const nodeType = data.agent ? 'agent' : 'tool';
  const style = getNodeStyle(nodeType);

  return (
    <>
      <Handle type="target" position={Position.Top} />
      <div
        className="custom-node"
        style={{
          background: style.background,
          borderColor: selected ? '#1e40af' : 'transparent',
          boxShadow: selected ? '0 0 0 2px #3b82f6' : undefined,
        }}
      >
        <div className="node-icon">{style.icon}</div>
        <div className="node-content">
          <div className="node-label">{data.label}</div>
          {data.agent && <div className="node-detail">Agent: {data.agent}</div>}
          {data.condition && <div className="node-condition">⚠️ Conditional</div>}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} />

      <style>{`
        .custom-node {
          padding: 0.75rem 1rem;
          border-radius: 0.5rem;
          border: 2px solid transparent;
          min-width: 180px;
          transition: all 0.2s;
        }

        .custom-node:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }

        .node-icon {
          font-size: 1.25rem;
          margin-bottom: 0.25rem;
        }

        .node-content {
          color: inherit;
        }

        .node-label {
          font-weight: 600;
          font-size: 0.875rem;
          margin-bottom: 0.25rem;
        }

        .node-detail {
          font-size: 0.75rem;
          opacity: 0.9;
        }

        .node-condition {
          font-size: 0.7rem;
          margin-top: 0.25rem;
          opacity: 0.95;
        }
      `}</style>
    </>
  );
}

export default memo(CustomNode);
