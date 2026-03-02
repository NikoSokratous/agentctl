package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Registry is the plugin registry client and manager.
type Registry struct {
	db        *sql.DB
	serverURL string
	localPath string
}

// NewRegistry creates a new registry.
func NewRegistry(db *sql.DB, serverURL, localPath string) *Registry {
	return &Registry{
		db:        db,
		serverURL: serverURL,
		localPath: localPath,
	}
}

// Register publishes a plugin to the registry.
func (r *Registry) Register(ctx context.Context, manifest *PluginManifest, artifactData []byte) error {
	// Validate metadata
	if err := ValidateMetadata(&manifest.Metadata); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	// Calculate checksum
	manifest.Metadata.Checksum = CalculateChecksum(artifactData)
	manifest.Metadata.CreatedAt = time.Now()
	manifest.Metadata.UpdatedAt = time.Now()

	// Store in database
	return r.storePlugin(ctx, &manifest.Metadata, artifactData)
}

// Search searches for plugins.
func (r *Registry) Search(ctx context.Context, filter *SearchFilter) (*SearchResult, error) {
	query := `
		SELECT id, name, version, author, description, type, runtime, 
		       license, repository, permissions, checksum, created_at
		FROM plugins
		WHERE 1=1
	`

	args := make([]interface{}, 0)

	// Build query
	if filter.Query != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		searchTerm := "%" + filter.Query + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if filter.Type != "" {
		query += " AND type = ?"
		args = append(args, filter.Type)
	}

	if filter.Runtime != "" {
		query += " AND runtime = ?"
		args = append(args, filter.Runtime)
	}

	if filter.Author != "" {
		query += " AND author = ?"
		args = append(args, filter.Author)
	}

	if filter.MinRating > 0 {
		query += " AND rating >= ?"
		args = append(args, filter.MinRating)
	}

	// Sort
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Limit and offset
	limit := 50
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plugins := make([]PluginMetadata, 0, limit)

	for rows.Next() {
		var p PluginMetadata
		var permsJSON string

		err := rows.Scan(
			&p.ID, &p.Name, &p.Version, &p.Author, &p.Description,
			&p.Type, &p.Runtime, &p.License, &p.Repository,
			&permsJSON, &p.Checksum, &p.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Parse permissions
		if permsJSON != "" {
			json.Unmarshal([]byte(permsJSON), &p.Permissions)
		}

		plugins = append(plugins, p)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM plugins WHERE 1=1"
	if filter.Query != "" {
		countQuery += " AND (name LIKE ? OR description LIKE ?)"
	}

	var total int
	countArgs := args[:len(args)-2] // Remove LIMIT and OFFSET
	r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)

	return &SearchResult{
		Plugins: plugins,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// Get retrieves a specific plugin.
func (r *Registry) Get(ctx context.Context, pluginID string) (*PluginMetadata, error) {
	query := `
		SELECT id, name, version, author, description, type, runtime,
		       license, repository, permissions, checksum, signature,
		       created_at, updated_at
		FROM plugins
		WHERE id = ?
	`

	var p PluginMetadata
	var permsJSON string

	err := r.db.QueryRowContext(ctx, query, pluginID).Scan(
		&p.ID, &p.Name, &p.Version, &p.Author, &p.Description,
		&p.Type, &p.Runtime, &p.License, &p.Repository,
		&permsJSON, &p.Checksum, &p.Signature,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse permissions
	if permsJSON != "" {
		json.Unmarshal([]byte(permsJSON), &p.Permissions)
	}

	return &p, nil
}

// GetStats retrieves plugin statistics.
func (r *Registry) GetStats(ctx context.Context, pluginID string) (*PluginStats, error) {
	query := `
		SELECT plugin_id, downloads, rating, 
		       (SELECT COUNT(*) FROM plugin_reviews WHERE plugin_id = ?) as review_count
		FROM plugins
		WHERE id = ?
	`

	var stats PluginStats
	err := r.db.QueryRowContext(ctx, query, pluginID, pluginID).Scan(
		&stats.PluginID,
		&stats.Downloads,
		&stats.Rating,
		&stats.ReviewCount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// IncrementDownloads increments download counter.
func (r *Registry) IncrementDownloads(ctx context.Context, pluginID string) error {
	query := "UPDATE plugins SET downloads = downloads + 1 WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, pluginID)
	return err
}

// AddReview adds a plugin review.
func (r *Registry) AddReview(ctx context.Context, review *PluginReview) error {
	if review.Rating < 1 || review.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	review.ID = fmt.Sprintf("rev-%d", time.Now().UnixNano())
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	query := `
		INSERT INTO plugin_reviews (
			id, plugin_id, user_id, rating, comment, helpful, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		review.ID,
		review.PluginID,
		review.UserID,
		review.Rating,
		review.Comment,
		0, // initial helpful count
		review.CreatedAt,
		review.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Update plugin rating
	return r.updatePluginRating(ctx, review.PluginID)
}

// GetReviews retrieves reviews for a plugin.
func (r *Registry) GetReviews(ctx context.Context, pluginID string, limit int) ([]PluginReview, error) {
	query := `
		SELECT id, plugin_id, user_id, rating, comment, helpful, created_at, updated_at
		FROM plugin_reviews
		WHERE plugin_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, pluginID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]PluginReview, 0, limit)

	for rows.Next() {
		var r PluginReview
		err := rows.Scan(
			&r.ID, &r.PluginID, &r.UserID, &r.Rating,
			&r.Comment, &r.Helpful, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			continue
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

// storePlugin stores plugin in database.
func (r *Registry) storePlugin(ctx context.Context, meta *PluginMetadata, artifactData []byte) error {
	metaJSON, _ := json.Marshal(meta.Metadata)

	query := `
		INSERT INTO plugins (
			id, name, version, author, type, metadata,
			artifact_url, downloads, rating, published_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			updated_at = CURRENT_TIMESTAMP,
			metadata = excluded.metadata
	`

	_, err := r.db.ExecContext(ctx, query,
		meta.ID,
		meta.Name,
		meta.Version,
		meta.Author,
		meta.Type,
		string(metaJSON),
		"",  // artifact_url would be set after upload
		0,   // initial downloads
		0.0, // initial rating
		meta.CreatedAt,
	)

	return err
}

// updatePluginRating recalculates plugin rating.
func (r *Registry) updatePluginRating(ctx context.Context, pluginID string) error {
	query := `
		UPDATE plugins 
		SET rating = (
			SELECT AVG(rating) 
			FROM plugin_reviews 
			WHERE plugin_id = ?
		)
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, pluginID, pluginID)
	return err
}

// ListInstalledPlugins lists locally installed plugins.
func (r *Registry) ListInstalledPlugins() ([]PluginMetadata, error) {
	// Placeholder: would scan local plugin directory
	return []PluginMetadata{}, nil
}

// IsInstalled checks if a plugin is installed locally.
func (r *Registry) IsInstalled(pluginID string) bool {
	// Placeholder: would check local filesystem
	return false
}

// GetInstallPath returns the local installation path for a plugin.
func (r *Registry) GetInstallPath(pluginID, version string) string {
	return fmt.Sprintf("%s/%s/%s", r.localPath, pluginID, version)
}
