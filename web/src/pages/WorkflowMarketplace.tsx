import { useState, useEffect } from 'react';
import { workflowsAPI, WorkflowTemplate } from '../api/workflows';
import './WorkflowMarketplace.css';

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
                  <span>v{template.version ?? '1.0'}</span>
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
                    <span className="stat-label">({template.rating_count ?? 0})</span>
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
      </div>
    </div>
  );
}
