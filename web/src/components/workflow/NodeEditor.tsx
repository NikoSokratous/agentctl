import { useState } from 'react';
import { WorkflowNode } from '../hooks/useWorkflowDesigner';

interface NodeEditorProps {
  node: WorkflowNode | null;
  onUpdate: (nodeId: string, data: Partial<WorkflowNode['data']>) => void;
  onClose: () => void;
}

export default function NodeEditor({ node, onUpdate, onClose }: NodeEditorProps) {
  const [formData, setFormData] = useState(node?.data || {});

  if (!node) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onUpdate(node.id, formData);
    onClose();
  };

  const handleChange = (field: string, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div className="node-editor-overlay" onClick={onClose}>
      <div className="node-editor" onClick={(e) => e.stopPropagation()}>
        <div className="editor-header">
          <h3>Edit Node: {node.data.label}</h3>
          <button className="close-btn" onClick={onClose}>
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="editor-form">
          <div className="form-group">
            <label>Label</label>
            <input
              type="text"
              value={formData.label || ''}
              onChange={(e) => handleChange('label', e.target.value)}
              required
            />
          </div>

          {node.type === 'agent' && (
            <>
              <div className="form-group">
                <label>Agent Name</label>
                <input
                  type="text"
                  value={formData.agent || ''}
                  onChange={(e) => handleChange('agent', e.target.value)}
                  placeholder="e.g., coder, researcher"
                  required
                />
              </div>

              <div className="form-group">
                <label>Goal</label>
                <textarea
                  value={formData.goal || ''}
                  onChange={(e) => handleChange('goal', e.target.value)}
                  placeholder="What should this agent accomplish?"
                  rows={3}
                  required
                />
              </div>
            </>
          )}

          <div className="form-group">
            <label>Output Key (optional)</label>
            <input
              type="text"
              value={formData.outputKey || ''}
              onChange={(e) => handleChange('outputKey', e.target.value)}
              placeholder="e.g., result, data"
            />
          </div>

          <div className="form-group">
            <label>Condition (CEL expression, optional)</label>
            <input
              type="text"
              value={formData.condition || ''}
              onChange={(e) => handleChange('condition', e.target.value)}
              placeholder="e.g., outputs.prev_step.status == 'success'"
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Timeout (optional)</label>
              <input
                type="text"
                value={formData.timeout || ''}
                onChange={(e) => handleChange('timeout', e.target.value)}
                placeholder="e.g., 5m, 30s"
              />
            </div>

            <div className="form-group">
              <label>Retry Count (optional)</label>
              <input
                type="number"
                value={formData.retry || ''}
                onChange={(e) => handleChange('retry', parseInt(e.target.value) || 0)}
                min="0"
                max="10"
              />
            </div>
          </div>

          <div className="editor-actions">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary">
              Save Changes
            </button>
          </div>
        </form>

        <style>{`
          .node-editor-overlay {
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

          .node-editor {
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
            width: 90%;
            max-width: 500px;
            max-height: 80vh;
            overflow-y: auto;
          }

          .editor-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 1.5rem;
            border-bottom: 1px solid #e5e7eb;
          }

          .editor-header h3 {
            margin: 0;
            font-size: 1.25rem;
            font-weight: 600;
          }

          .close-btn {
            background: none;
            border: none;
            font-size: 1.5rem;
            cursor: pointer;
            color: #6b7280;
            padding: 0;
            width: 2rem;
            height: 2rem;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 0.25rem;
          }

          .close-btn:hover {
            background: #f3f4f6;
          }

          .editor-form {
            padding: 1.5rem;
          }

          .form-group {
            margin-bottom: 1.25rem;
          }

          .form-group label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 500;
            color: #374151;
            font-size: 0.875rem;
          }

          .form-group input,
          .form-group textarea {
            width: 100%;
            padding: 0.625rem;
            border: 1px solid #d1d5db;
            border-radius: 0.375rem;
            font-size: 0.875rem;
            font-family: inherit;
          }

          .form-group input:focus,
          .form-group textarea:focus {
            outline: none;
            border-color: #3b82f6;
            box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
          }

          .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 1rem;
          }

          .editor-actions {
            display: flex;
            gap: 0.75rem;
            justify-content: flex-end;
            margin-top: 2rem;
            padding-top: 1.5rem;
            border-top: 1px solid #e5e7eb;
          }

          .btn-primary,
          .btn-secondary {
            padding: 0.625rem 1.25rem;
            border-radius: 0.375rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
            border: none;
            font-size: 0.875rem;
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
    </div>
  );
}
