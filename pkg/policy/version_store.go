package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// VersionStore manages policy version storage.
type VersionStore struct {
	db       *sql.DB
	fileBase string // Base directory for file storage
}

// NewVersionStore creates a new policy version store.
func NewVersionStore(db *sql.DB, fileBase string) (*VersionStore, error) {
	store := &VersionStore{
		db:       db,
		fileBase: fileBase,
	}

	// Ensure file storage directory exists
	if fileBase != "" {
		if err := os.MkdirAll(fileBase, 0755); err != nil {
			return nil, fmt.Errorf("create policy storage dir: %w", err)
		}
	}

	return store, nil
}

// SaveVersion saves a new policy version.
func (s *VersionStore) SaveVersion(ctx context.Context, version *PolicyVersion) error {
	if version.ID == "" {
		version.ID = uuid.New().String()
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now()
	}
	if version.EffectiveAt.IsZero() {
		version.EffectiveAt = version.CreatedAt
	}

	// Deactivate previous active version
	if version.Active {
		if err := s.deactivateVersions(ctx, version.PolicyName); err != nil {
			return fmt.Errorf("deactivate previous versions: %w", err)
		}
	}

	// Store content in file if configured
	var contentPath string
	if s.fileBase != "" {
		contentPath = filepath.Join(s.fileBase, version.PolicyName, version.Version+".yaml")
		if err := os.MkdirAll(filepath.Dir(contentPath), 0755); err != nil {
			return fmt.Errorf("create policy dir: %w", err)
		}
		if err := os.WriteFile(contentPath, version.Content, 0644); err != nil {
			return fmt.Errorf("write policy file: %w", err)
		}
	}

	// Store metadata in database
	metadataJSON, err := json.Marshal(version.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO policy_versions (
			id, policy_name, version, content, format, author, 
			changelog, effective_at, created_at, supersedes, active, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		version.ID,
		version.PolicyName,
		version.Version,
		version.Content,
		version.Format,
		version.Author,
		version.Changelog,
		version.EffectiveAt,
		version.CreatedAt,
		version.Supersedes,
		version.Active,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("insert policy version: %w", err)
	}

	return nil
}

// GetVersion retrieves a specific policy version.
func (s *VersionStore) GetVersion(ctx context.Context, policyName, version string) (*PolicyVersion, error) {
	query := `
		SELECT id, policy_name, version, content, format, author,
		       changelog, effective_at, created_at, supersedes, active, metadata
		FROM policy_versions
		WHERE policy_name = ? AND version = ?
	`

	var pv PolicyVersion
	var metadataJSON []byte

	err := s.db.QueryRowContext(ctx, query, policyName, version).Scan(
		&pv.ID,
		&pv.PolicyName,
		&pv.Version,
		&pv.Content,
		&pv.Format,
		&pv.Author,
		&pv.Changelog,
		&pv.EffectiveAt,
		&pv.CreatedAt,
		&pv.Supersedes,
		&pv.Active,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy version not found: %s@%s", policyName, version)
	}
	if err != nil {
		return nil, fmt.Errorf("query policy version: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &pv.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &pv, nil
}

// GetActiveVersion retrieves the currently active policy version.
func (s *VersionStore) GetActiveVersion(ctx context.Context, policyName string) (*PolicyVersion, error) {
	query := `
		SELECT id, policy_name, version, content, format, author,
		       changelog, effective_at, created_at, supersedes, active, metadata
		FROM policy_versions
		WHERE policy_name = ? AND active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var pv PolicyVersion
	var metadataJSON []byte

	err := s.db.QueryRowContext(ctx, query, policyName).Scan(
		&pv.ID,
		&pv.PolicyName,
		&pv.Version,
		&pv.Content,
		&pv.Format,
		&pv.Author,
		&pv.Changelog,
		&pv.EffectiveAt,
		&pv.CreatedAt,
		&pv.Supersedes,
		&pv.Active,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active policy found: %s", policyName)
	}
	if err != nil {
		return nil, fmt.Errorf("query active policy: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &pv.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &pv, nil
}

// ListVersions lists all versions of a policy.
func (s *VersionStore) ListVersions(ctx context.Context, policyName string) ([]PolicyVersionMetadata, error) {
	query := `
		SELECT policy_name, version, author, changelog, effective_at, created_at, active
		FROM policy_versions
		WHERE policy_name = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, policyName)
	if err != nil {
		return nil, fmt.Errorf("query policy versions: %w", err)
	}
	defer rows.Close()

	var versions []PolicyVersionMetadata
	for rows.Next() {
		var v PolicyVersionMetadata
		err := rows.Scan(
			&v.PolicyName,
			&v.Version,
			&v.Author,
			&v.Changelog,
			&v.EffectiveAt,
			&v.CreatedAt,
			&v.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("scan policy version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, rows.Err()
}

// ListPolicies lists all policy names.
func (s *VersionStore) ListPolicies(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT policy_name
		FROM policy_versions
		ORDER BY policy_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query policies: %w", err)
	}
	defer rows.Close()

	var policies []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan policy name: %w", err)
		}
		policies = append(policies, name)
	}

	return policies, rows.Err()
}

// SetActiveVersion sets a specific version as active.
func (s *VersionStore) SetActiveVersion(ctx context.Context, policyName, version string) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate all versions
	if _, err := tx.ExecContext(ctx,
		"UPDATE policy_versions SET active = false WHERE policy_name = ?",
		policyName); err != nil {
		return fmt.Errorf("deactivate versions: %w", err)
	}

	// Activate specified version
	result, err := tx.ExecContext(ctx,
		"UPDATE policy_versions SET active = true WHERE policy_name = ? AND version = ?",
		policyName, version)
	if err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("policy version not found: %s@%s", policyName, version)
	}

	return tx.Commit()
}

// DeleteVersion deletes a policy version (soft delete by marking inactive).
func (s *VersionStore) DeleteVersion(ctx context.Context, policyName, version string) error {
	// Check if it's the active version
	active, err := s.GetActiveVersion(ctx, policyName)
	if err == nil && active.Version == version {
		return fmt.Errorf("cannot delete active policy version")
	}

	result, err := s.db.ExecContext(ctx,
		"DELETE FROM policy_versions WHERE policy_name = ? AND version = ?",
		policyName, version)
	if err != nil {
		return fmt.Errorf("delete policy version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("policy version not found: %s@%s", policyName, version)
	}

	// Delete file if exists
	if s.fileBase != "" {
		contentPath := filepath.Join(s.fileBase, policyName, version+".yaml")
		os.Remove(contentPath) // Ignore error
	}

	return nil
}

// deactivateVersions deactivates all versions of a policy.
func (s *VersionStore) deactivateVersions(ctx context.Context, policyName string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE policy_versions SET active = false WHERE policy_name = ?",
		policyName)
	return err
}
