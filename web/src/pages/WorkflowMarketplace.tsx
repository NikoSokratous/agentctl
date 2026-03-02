import { useState, useEffect } from 'react';
import { workflowsAPI, WorkflowTemplate } from '../api/workflows';

export default function WorkflowMarketplace() {
  const [templates, setTemplates] = useState<WorkflowTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showPublish, setShowPublish] = useState(false);

  const categories = ['all', 'data-processing', 'code-review', 'research', 'deployment', 'testing'];

  useEffect(() => {
    loadTemplates();
  }, [selectedCategory, searchQuery]);

  const loadTemplates = async () => {
    setLoading(true);
    try {
      const filters: any = { limit: 50 };
      if (selectedCategory !== 'all') filters.category = selectedCategory;
      if (searchQuery) filters.search = searchQuery;

      const result = await workflowsAPI.listTemplates(filters);
      setTemplates(result);
    } catch (error) {
      console.error('Failed to load templates:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleInstall = async (template: WorkflowTemplate) => {
    try {
      const result = await workflowsAPI.installTemplate(template.id);
      
      // Download YAML file
      const blob = new Blob([result.yaml], { type: 'text/yaml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${template.name}.yaml`;
      a.click();
      URL.revokeObjectURL(url);

      alert(`Template "${template.name}" installed successfully!`);
    } catch (error) {
      alert('Failed to install template');
    }
  };

  const handleRate = async (template: WorkflowTemplate, rating: number) => {
    try {
      await workflowsAPI.rateTemplate(template.id, rating);
      loadTemplates();
    } catch (error) {
      alert('Failed to rate template');
    }
  };

  return (
    <div className="marketplace">
      <div className="marketplace-header">
        <h1>Workflow Marketplace</h1>
        <button className="btn-publish" onClick={() => setShowPublish(true)}>
          + Publish Template
        </button>
      </div>

      <div className="marketplace-filters">
        <div className="search-box">
          <input
            type="text"
            placeholder="Search workflows..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="category-tabs">
          {categories.map((category) => (
            <button
              key={category}
              className={`tab ${selectedCategory === category ? 'active' : ''}`}
              onClick={() => setSelectedCategory(category)}
            >
              {category.replace('-', ' ')}
            </button>
          ))}
        </div>
      </div>

      <div className="marketplace-content">
        {loading ? (
          <div className="loading">Loading templates...</div>
        ) : templates.length === 0 ? (
          <div className="empty">No templates found</div>
        ) : (
          <div className="template-grid">
            {templates.map((template) => (
              <div key={template.id} className="template-card">
                <div className="card-header">
                  <h3>{template.name}</h3>
                  <span className="category-badge">{template.category}</span>
                </div>

                <p className="description">{template.description}</p>

                <div className="card-meta">
                  <span>by {template.author}</span>
                  <span>v{template.version}</span>
                </div>

                <div className="card-tags">
                  {template.tags.map((tag) => (
                    <span key={tag} className="tag">
                      {tag}
                    </span>
                  ))}
                </div>

                <div className="card-stats">
                  <div className="stat">
                    <span className="stat-icon">⭐</span>
                    <span>{template.rating.toFixed(1)}</span>
                    <span className="stat-label">({template.rating_count})</span>
                  </div>
                  <div className="stat">
                    <span className="stat-icon">📥</span>
                    <span>{template.downloads}</span>
                    <span className="stat-label">downloads</span>
                  </div>
                </div>

                <div className="card-actions">
                  <button className="btn-install" onClick={() => handleInstall(template)}>
                    📥 Install
                  </button>
                  <div className="rating-stars">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <button
                        key={star}
                        className="star-btn"
                        onClick={() => handleRate(template, star)}
                      >
                        {star <= template.rating ? '⭐' : '☆'}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {showPublish && (
        <PublishModal onClose={() => setShowPublish(false)} onPublish={loadTemplates} />
      )}

      <style>{`
        .marketplace {
          padding: 2rem;
          max-width: 1400px;
          margin: 0 auto;
        }

        .marketplace-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
        }

        .marketplace-header h1 {
          margin: 0;
          font-size: 2rem;
        }

        .btn-publish {
          padding: 0.75rem 1.5rem;
          background: #3b82f6;
          color: white;
          border: none;
          border-radius: 0.5rem;
          font-weight: 600;
          cursor: pointer;
          transition: background 0.2s;
        }

        .btn-publish:hover {
          background: #2563eb;
        }

        .marketplace-filters {
          margin-bottom: 2rem;
        }

        .search-box {
          margin-bottom: 1rem;
        }

        .search-box input {
          width: 100%;
          max-width: 500px;
          padding: 0.75rem 1rem;
          border: 1px solid #d1d5db;
          border-radius: 0.5rem;
          font-size: 1rem;
        }

        .category-tabs {
          display: flex;
          gap: 0.5rem;
          flex-wrap: wrap;
        }

        .tab {
          padding: 0.5rem 1rem;
          background: #f3f4f6;
          border: 1px solid #e5e7eb;
          border-radius: 0.5rem;
          cursor: pointer;
          font-weight: 500;
          text-transform: capitalize;
          transition: all 0.2s;
        }

        .tab:hover {
          background: #e5e7eb;
        }

        .tab.active {
          background: #3b82f6;
          color: white;
          border-color: #3b82f6;
        }

        .marketplace-content {
          min-height: 400px;
        }

        .loading,
        .empty {
          text-align: center;
          padding: 3rem;
          color: #6b7280;
        }

        .template-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
          gap: 1.5rem;
        }

        .template-card {
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.75rem;
          padding: 1.5rem;
          transition: all 0.2s;
        }

        .template-card:hover {
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
          transform: translateY(-2px);
        }

        .card-header {
          display: flex;
          justify-content: space-between;
          align-items: start;
          margin-bottom: 0.75rem;
        }

        .card-header h3 {
          margin: 0;
          font-size: 1.25rem;
          flex: 1;
        }

        .category-badge {
          padding: 0.25rem 0.75rem;
          background: #f3f4f6;
          border-radius: 1rem;
          font-size: 0.75rem;
          font-weight: 500;
          text-transform: capitalize;
        }

        .description {
          color: #6b7280;
          margin-bottom: 1rem;
          line-height: 1.5;
        }

        .card-meta {
          display: flex;
          justify-content: space-between;
          font-size: 0.875rem;
          color: #6b7280;
          margin-bottom: 1rem;
        }

        .card-tags {
          display: flex;
          gap: 0.5rem;
          flex-wrap: wrap;
          margin-bottom: 1rem;
        }

        .tag {
          padding: 0.25rem 0.5rem;
          background: #ede9fe;
          color: #6b21a8;
          border-radius: 0.25rem;
          font-size: 0.75rem;
        }

        .card-stats {
          display: flex;
          gap: 1.5rem;
          padding: 0.75rem 0;
          border-top: 1px solid #e5e7eb;
          border-bottom: 1px solid #e5e7eb;
          margin-bottom: 1rem;
        }

        .stat {
          display: flex;
          align-items: center;
          gap: 0.25rem;
          font-size: 0.875rem;
        }

        .stat-icon {
          font-size: 1rem;
        }

        .stat-label {
          color: #6b7280;
        }

        .card-actions {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .btn-install {
          padding: 0.5rem 1rem;
          background: #10b981;
          color: white;
          border: none;
          border-radius: 0.375rem;
          font-weight: 500;
          cursor: pointer;
          transition: background 0.2s;
        }

        .btn-install:hover {
          background: #059669;
        }

        .rating-stars {
          display: flex;
          gap: 0.125rem;
        }

        .star-btn {
          background: none;
          border: none;
          font-size: 1.25rem;
          cursor: pointer;
          padding: 0;
        }
      `}</style>
    </div>
  );
}

function PublishModal({ onClose, onPublish }: { onClose: () => void; onPublish: () => void }) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'data-processing',
    tags: '',
    template_yaml: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await workflowsAPI.publishTemplate({
        ...formData,
        tags: formData.tags.split(',').map((t) => t.trim()),
      });
      alert('Template published successfully!');
      onPublish();
      onClose();
    } catch (error) {
      alert('Failed to publish template');
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>Publish Template</h2>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Template Name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />
          <textarea
            placeholder="Description"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            rows={3}
            required
          />
          <select
            value={formData.category}
            onChange={(e) => setFormData({ ...formData, category: e.target.value })}
          >
            <option value="data-processing">Data Processing</option>
            <option value="code-review">Code Review</option>
            <option value="research">Research</option>
            <option value="deployment">Deployment</option>
            <option value="testing">Testing</option>
          </select>
          <input
            type="text"
            placeholder="Tags (comma-separated)"
            value={formData.tags}
            onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
          />
          <textarea
            placeholder="Workflow YAML"
            value={formData.template_yaml}
            onChange={(e) => setFormData({ ...formData, template_yaml: e.target.value })}
            rows={10}
            required
          />
          <div className="modal-actions">
            <button type="button" onClick={onClose}>
              Cancel
            </button>
            <button type="submit">Publish</button>
          </div>
        </form>

        <style>{`
          .modal-overlay {
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

          .modal {
            background: white;
            border-radius: 0.5rem;
            padding: 2rem;
            width: 90%;
            max-width: 600px;
            max-height: 80vh;
            overflow-y: auto;
          }

          .modal h2 {
            margin: 0 0 1.5rem 0;
          }

          .modal form {
            display: flex;
            flex-direction: column;
            gap: 1rem;
          }

          .modal input,
          .modal textarea,
          .modal select {
            padding: 0.75rem;
            border: 1px solid #d1d5db;
            border-radius: 0.375rem;
            font-family: inherit;
          }

          .modal textarea {
            font-family: monospace;
            font-size: 0.875rem;
          }

          .modal-actions {
            display: flex;
            gap: 1rem;
            justify-content: flex-end;
            margin-top: 1rem;
          }

          .modal-actions button {
            padding: 0.75rem 1.5rem;
            border: none;
            border-radius: 0.375rem;
            cursor: pointer;
            font-weight: 500;
          }

          .modal-actions button[type='button'] {
            background: #f3f4f6;
            color: #374151;
          }

          .modal-actions button[type='submit'] {
            background: #3b82f6;
            color: white;
          }
        `}</style>
      </div>
    </div>
  );
}
