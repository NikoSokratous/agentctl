import { useCallback, useRef, useState } from 'react';
import ReactFlow, { Background, Controls, MiniMap, ReactFlowProvider } from 'reactflow';
import 'reactflow/dist/style.css';
import { useWorkflowDesigner, WorkflowNode } from '../hooks/useWorkflowDesigner';
import NodePalette from '../components/workflow/NodePalette';
import NodeEditor from '../components/workflow/NodeEditor';
import CustomNode from '../components/workflow/CustomNode';

const nodeTypes = {
  agent: CustomNode,
  tool: CustomNode,
  condition: CustomNode,
  parallel: CustomNode,
  join: CustomNode,
};

function WorkflowDesignerContent() {
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [selectedNode, setSelectedNode] = useState<WorkflowNode | null>(null);
  const [yamlImport, setYamlImport] = useState('');
  const [showImport, setShowImport] = useState(false);

  const {
    nodes,
    edges,
    isValid,
    errors,
    onNodesChange,
    onEdgesChange,
    onConnect,
    addNode,
    updateNode,
    validateWorkflow,
    exportToYAML,
    importFromYAML,
  } = useWorkflowDesigner();

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow');
      if (!type || !reactFlowWrapper.current) return;

      const reactFlowBounds = reactFlowWrapper.current.getBoundingClientRect();
      const position = {
        x: event.clientX - reactFlowBounds.left,
        y: event.clientY - reactFlowBounds.top,
      };

      addNode(type, position);
    },
    [addNode]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onNodeClick = useCallback((_: any, node: WorkflowNode) => {
    setSelectedNode(node);
  }, []);

  const handleExport = () => {
    const yaml = exportToYAML();
    const blob = new Blob([yaml], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'workflow.yaml';
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImport = () => {
    try {
      importFromYAML(yamlImport);
      setShowImport(false);
      setYamlImport('');
    } catch (error) {
      alert('Failed to import YAML: ' + (error as Error).message);
    }
  };

  return (
    <div className="workflow-designer">
      <div className="designer-header">
        <h1>Workflow Designer</h1>
        <div className="header-actions">
          <button className="btn-action" onClick={() => setShowImport(true)}>
            📥 Import YAML
          </button>
          <button className="btn-action" onClick={handleExport}>
            📤 Export YAML
          </button>
          <button className="btn-action" onClick={validateWorkflow}>
            ✓ Validate
          </button>
          <div className={`status-indicator ${isValid ? 'valid' : 'invalid'}`}>
            {isValid ? '✓ Valid' : '✗ Invalid'}
          </div>
        </div>
      </div>

      {errors.length > 0 && (
        <div className="error-banner">
          <strong>Validation Errors:</strong>
          <ul>
            {errors.map((error, idx) => (
              <li key={idx}>{error}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="designer-content">
        <NodePalette onAddNode={addNode} />
        
        <div className="canvas-wrapper" ref={reactFlowWrapper}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onDrop={onDrop}
            onDragOver={onDragOver}
            onNodeClick={onNodeClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
      </div>

      {selectedNode && (
        <NodeEditor
          node={selectedNode}
          onUpdate={updateNode}
          onClose={() => setSelectedNode(null)}
        />
      )}

      {showImport && (
        <div className="import-modal-overlay" onClick={() => setShowImport(false)}>
          <div className="import-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Import Workflow YAML</h3>
            <textarea
              value={yamlImport}
              onChange={(e) => setYamlImport(e.target.value)}
              placeholder="Paste your workflow YAML here..."
              rows={15}
            />
            <div className="modal-actions">
              <button className="btn-secondary" onClick={() => setShowImport(false)}>
                Cancel
              </button>
              <button className="btn-primary" onClick={handleImport}>
                Import
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{`
        .workflow-designer {
          height: 100vh;
          display: flex;
          flex-direction: column;
          background: #f9fafb;
        }

        .designer-header {
          background: white;
          border-bottom: 1px solid #e5e7eb;
          padding: 1rem 1.5rem;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .designer-header h1 {
          margin: 0;
          font-size: 1.5rem;
          font-weight: 600;
        }

        .header-actions {
          display: flex;
          gap: 0.75rem;
          align-items: center;
        }

        .btn-action {
          padding: 0.5rem 1rem;
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 0.375rem;
          cursor: pointer;
          font-weight: 500;
          transition: all 0.2s;
        }

        .btn-action:hover {
          background: #e5e7eb;
        }

        .status-indicator {
          padding: 0.5rem 1rem;
          border-radius: 0.375rem;
          font-weight: 500;
          font-size: 0.875rem;
        }

        .status-indicator.valid {
          background: #d1fae5;
          color: #065f46;
        }

        .status-indicator.invalid {
          background: #fee2e2;
          color: #991b1b;
        }

        .error-banner {
          background: #fef2f2;
          border-bottom: 1px solid #fecaca;
          padding: 1rem 1.5rem;
          color: #991b1b;
        }

        .error-banner ul {
          margin: 0.5rem 0 0 0;
          padding-left: 1.5rem;
        }

        .designer-content {
          flex: 1;
          display: flex;
          overflow: hidden;
        }

        .canvas-wrapper {
          flex: 1;
          position: relative;
        }

        .import-modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
        }

        .import-modal {
          background: white;
          border-radius: 0.5rem;
          padding: 1.5rem;
          width: 90%;
          max-width: 600px;
        }

        .import-modal h3 {
          margin: 0 0 1rem 0;
        }

        .import-modal textarea {
          width: 100%;
          padding: 0.75rem;
          border: 1px solid #d1d5db;
          border-radius: 0.375rem;
          font-family: monospace;
          font-size: 0.875rem;
        }

        .modal-actions {
          display: flex;
          gap: 0.75rem;
          justify-content: flex-end;
          margin-top: 1rem;
        }

        .btn-primary,
        .btn-secondary {
          padding: 0.625rem 1.25rem;
          border-radius: 0.375rem;
          font-weight: 500;
          cursor: pointer;
          border: none;
        }

        .btn-primary {
          background: #3b82f6;
          color: white;
        }

        .btn-primary:hover {
          background: #2563eb;
        }

        .btn-secondary {
          background: #f3f4f6;
          color: #374151;
        }

        .btn-secondary:hover {
          background: #e5e7eb;
        }
      `}</style>
    </div>
  );
}

export default function WorkflowDesigner() {
  return (
    <ReactFlowProvider>
      <WorkflowDesignerContent />
    </ReactFlowProvider>
  );
}
