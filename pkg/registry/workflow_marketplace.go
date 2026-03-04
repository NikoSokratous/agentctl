package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// WorkflowTemplate represents a shareable workflow template
type WorkflowTemplate struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Author       string                 `json:"author"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	TemplateYAML string                 `json:"template_yaml"`
	Parameters   []TemplateParameter    `json:"parameters"`
	Downloads    int64                  `json:"downloads"`
	Rating       float64                `json:"rating"`
	RatingCount  int64                  `json:"rating_count"`
	Version      string                 `json:"version"`
	License      string                 `json:"license"`
	PublishedAt  time.Time              `json:"published_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateParameter represents a template parameter
type TemplateParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// WorkflowMarketplace manages workflow templates
type WorkflowMarketplace struct {
	db *sql.DB
}

// NewWorkflowMarketplace creates a new workflow marketplace
func NewWorkflowMarketplace(db *sql.DB) *WorkflowMarketplace {
	return &WorkflowMarketplace{db: db}
}

// Publish publishes a new workflow template
func (wm *WorkflowMarketplace) Publish(ctx context.Context, template *WorkflowTemplate) error {
	tagsJSON, err := json.Marshal(template.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	paramsJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("marshal parameters: %w", err)
	}

	metadataJSON, err := json.Marshal(template.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO workflow_templates (
			id, name, author, description, category, tags, template_yaml,
			parameters, version, license, rating, rating_count,
			downloads, published_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, ?, ?, ?)
	`

	now := time.Now()
	_, err = wm.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Author,
		template.Description,
		template.Category,
		string(tagsJSON),
		template.TemplateYAML,
		string(paramsJSON),
		template.Version,
		template.License,
		now,
		now,
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}

	return nil
}

// Search searches for workflow templates
func (wm *WorkflowMarketplace) Search(ctx context.Context, filters map[string]interface{}) ([]*WorkflowTemplate, error) {
	query := `
		SELECT id, name, author, description, category, tags, template_yaml,
		       parameters, version, license, rating, rating_count, downloads,
		       published_at, updated_at, metadata
		FROM workflow_templates
		WHERE 1=1
	`

	args := []interface{}{}

	if category, ok := filters["category"].(string); ok && category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY downloads DESC, rating DESC"

	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := wm.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var templates []*WorkflowTemplate
	for rows.Next() {
		var t WorkflowTemplate
		var tagsJSON, paramsJSON, metadataJSON string

		err := rows.Scan(
			&t.ID, &t.Name, &t.Author, &t.Description, &t.Category,
			&tagsJSON, &t.TemplateYAML, &paramsJSON, &t.Version, &t.License,
			&t.Rating, &t.RatingCount, &t.Downloads,
			&t.PublishedAt, &t.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(paramsJSON), &t.Parameters); err != nil {
			return nil, fmt.Errorf("unmarshal parameters: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &t.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// Get retrieves a workflow template by ID
func (wm *WorkflowMarketplace) Get(ctx context.Context, id string) (*WorkflowTemplate, error) {
	query := `
		SELECT id, name, author, description, category, tags, template_yaml,
		       parameters, version, license, rating, rating_count, downloads,
		       published_at, updated_at, metadata
		FROM workflow_templates
		WHERE id = ?
	`

	var t WorkflowTemplate
	var tagsJSON, paramsJSON, metadataJSON string

	err := wm.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Author, &t.Description, &t.Category,
		&tagsJSON, &t.TemplateYAML, &paramsJSON, &t.Version, &t.License,
		&t.Rating, &t.RatingCount, &t.Downloads,
		&t.PublishedAt, &t.UpdatedAt, &metadataJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query template: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(paramsJSON), &t.Parameters); err != nil {
		return nil, fmt.Errorf("unmarshal parameters: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &t.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return &t, nil
}

// IncrementDownloads increments the download counter
func (wm *WorkflowMarketplace) IncrementDownloads(ctx context.Context, id string) error {
	query := `UPDATE workflow_templates SET downloads = downloads + 1 WHERE id = ?`
	_, err := wm.db.ExecContext(ctx, query, id)
	return err
}

// Rate adds a rating to a template
func (wm *WorkflowMarketplace) Rate(ctx context.Context, id string, rating float64) error {
	query := `
		UPDATE workflow_templates
		SET rating = ((rating * rating_count) + ?) / (rating_count + 1),
		    rating_count = rating_count + 1
		WHERE id = ?
	`
	_, err := wm.db.ExecContext(ctx, query, rating, id)
	return err
}

// GetPopular returns the most popular templates
func (wm *WorkflowMarketplace) GetPopular(ctx context.Context, limit int) ([]*WorkflowTemplate, error) {
	return wm.Search(ctx, map[string]interface{}{
		"limit": limit,
	})
}

// GetByCategory returns templates by category
func (wm *WorkflowMarketplace) GetByCategory(ctx context.Context, category string, limit int) ([]*WorkflowTemplate, error) {
	return wm.Search(ctx, map[string]interface{}{
		"category": category,
		"limit":    limit,
	})
}

// WorkflowVersion represents a versioned workflow template.
type WorkflowVersion struct {
	ID           string    `json:"id"`
	WorkflowName string    `json:"workflow_name"`
	Version      string    `json:"version"`
	TemplateYAML string    `json:"template_yaml"`
	Parameters   string    `json:"parameters"`
	Author       string    `json:"author"`
	Changelog    string    `json:"changelog"`
	EffectiveAt  time.Time `json:"effective_at"`
	CreatedAt    time.Time `json:"created_at"`
	Metadata     string    `json:"metadata"`
}

// ListWorkflowVersions lists all versions of a workflow.
func (wm *WorkflowMarketplace) ListWorkflowVersions(ctx context.Context, workflowName string) ([]WorkflowVersion, error) {
	query := `
		SELECT id, workflow_name, version, template_yaml, parameters, author,
		       changelog, effective_at, created_at, metadata
		FROM workflow_versions
		WHERE workflow_name = ?
		ORDER BY created_at DESC
	`
	rows, err := wm.db.QueryContext(ctx, query, workflowName)
	if err != nil {
		return nil, fmt.Errorf("query workflow versions: %w", err)
	}
	defer rows.Close()

	var versions []WorkflowVersion
	for rows.Next() {
		var v WorkflowVersion
		err := rows.Scan(
			&v.ID, &v.WorkflowName, &v.Version, &v.TemplateYAML, &v.Parameters,
			&v.Author, &v.Changelog, &v.EffectiveAt, &v.CreatedAt, &v.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// GetWorkflowVersion retrieves a specific workflow version.
func (wm *WorkflowMarketplace) GetWorkflowVersion(ctx context.Context, workflowName, version string) (*WorkflowVersion, error) {
	query := `
		SELECT id, workflow_name, version, template_yaml, parameters, author,
		       changelog, effective_at, created_at, metadata
		FROM workflow_versions
		WHERE workflow_name = ? AND version = ?
	`
	var v WorkflowVersion
	err := wm.db.QueryRowContext(ctx, query, workflowName, version).Scan(
		&v.ID, &v.WorkflowName, &v.Version, &v.TemplateYAML, &v.Parameters,
		&v.Author, &v.Changelog, &v.EffectiveAt, &v.CreatedAt, &v.Metadata,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workflow version not found: %s@%s", workflowName, version)
	}
	if err != nil {
		return nil, fmt.Errorf("query workflow version: %w", err)
	}
	return &v, nil
}

// PublishWorkflowVersion saves a new workflow version.
func (wm *WorkflowMarketplace) PublishWorkflowVersion(ctx context.Context, v *WorkflowVersion) error {
	if v.ID == "" {
		v.ID = fmt.Sprintf("wv-%d", time.Now().UnixNano())
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	if v.EffectiveAt.IsZero() {
		v.EffectiveAt = v.CreatedAt
	}

	query := `
		INSERT INTO workflow_versions (
			id, workflow_name, version, template_yaml, parameters,
			author, changelog, effective_at, created_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := wm.db.ExecContext(ctx, query,
		v.ID, v.WorkflowName, v.Version, v.TemplateYAML, v.Parameters,
		v.Author, v.Changelog, v.EffectiveAt, v.CreatedAt, v.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert workflow version: %w", err)
	}
	return nil
}

// ListWorkflowNames returns distinct workflow names that have versions.
func (wm *WorkflowMarketplace) ListWorkflowNames(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT workflow_name FROM workflow_versions ORDER BY workflow_name`
	rows, err := wm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query workflow names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}
