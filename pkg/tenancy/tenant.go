package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Tenant represents a tenant in the multi-tenant system.
type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
	Active      bool                   `json:"active"`
}

// TenantManager manages tenants.
type TenantManager struct {
	db *sql.DB
}

// NewTenantManager creates a new tenant manager.
func NewTenantManager(db *sql.DB) *TenantManager {
	return &TenantManager{db: db}
}

// CreateTenant creates a new tenant.
func (tm *TenantManager) CreateTenant(ctx context.Context, name, displayName string) (*Tenant, error) {
	tenant := &Tenant{
		ID:          generateTenantID(),
		Name:        name,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Settings:    make(map[string]interface{}),
		Active:      true,
	}

	query := `
		INSERT INTO tenants (id, name, display_name, created_at, settings)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := tm.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.DisplayName,
		tenant.CreatedAt,
		"{}",
	)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	query := `
		SELECT id, name, display_name, created_at, settings
		FROM tenants
		WHERE id = ?
	`

	var tenant Tenant
	var settingsJSON string

	err := tm.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.DisplayName,
		&tenant.CreatedAt,
		&settingsJSON,
	)
	if err != nil {
		return nil, err
	}

	tenant.Active = true
	tenant.Settings = make(map[string]interface{})

	return &tenant, nil
}

// GetTenantByName retrieves a tenant by name.
func (tm *TenantManager) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	query := `
		SELECT id, name, display_name, created_at
		FROM tenants
		WHERE name = ?
	`

	var tenant Tenant

	err := tm.db.QueryRowContext(ctx, query, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.DisplayName,
		&tenant.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// ListTenants lists all tenants.
func (tm *TenantManager) ListTenants(ctx context.Context, limit int) ([]Tenant, error) {
	query := `
		SELECT id, name, display_name, created_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := tm.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := make([]Tenant, 0, limit)

	for rows.Next() {
		var t Tenant
		err := rows.Scan(&t.ID, &t.Name, &t.DisplayName, &t.CreatedAt)
		if err != nil {
			continue
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

// UpdateTenant updates tenant information.
func (tm *TenantManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	query := `
		UPDATE tenants
		SET display_name = ?, settings = ?
		WHERE id = ?
	`

	_, err := tm.db.ExecContext(ctx, query,
		tenant.DisplayName,
		"{}",
		tenant.ID,
	)

	return err
}

// DeleteTenant deletes a tenant.
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	query := "DELETE FROM tenants WHERE id = ?"
	_, err := tm.db.ExecContext(ctx, query, tenantID)
	return err
}

// AddMember adds a user to a tenant.
func (tm *TenantManager) AddMember(ctx context.Context, tenantID, userID, role string) error {
	query := `
		INSERT INTO tenant_members (tenant_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := tm.db.ExecContext(ctx, query, tenantID, userID, role, time.Now())
	return err
}

// RemoveMember removes a user from a tenant.
func (tm *TenantManager) RemoveMember(ctx context.Context, tenantID, userID string) error {
	query := "DELETE FROM tenant_members WHERE tenant_id = ? AND user_id = ?"
	_, err := tm.db.ExecContext(ctx, query, tenantID, userID)
	return err
}

// ListMembers lists all members of a tenant.
func (tm *TenantManager) ListMembers(ctx context.Context, tenantID string) ([]TenantMember, error) {
	query := `
		SELECT tenant_id, user_id, role, joined_at
		FROM tenant_members
		WHERE tenant_id = ?
		ORDER BY joined_at ASC
	`

	rows, err := tm.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]TenantMember, 0)

	for rows.Next() {
		var m TenantMember
		err := rows.Scan(&m.TenantID, &m.UserID, &m.Role, &m.JoinedAt)
		if err != nil {
			continue
		}
		members = append(members, m)
	}

	return members, nil
}

// GetUserTenants returns all tenants a user belongs to.
func (tm *TenantManager) GetUserTenants(ctx context.Context, userID string) ([]Tenant, error) {
	query := `
		SELECT t.id, t.name, t.display_name, t.created_at
		FROM tenants t
		JOIN tenant_members tm ON t.id = tm.tenant_id
		WHERE tm.user_id = ?
		ORDER BY t.name ASC
	`

	rows, err := tm.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := make([]Tenant, 0)

	for rows.Next() {
		var t Tenant
		err := rows.Scan(&t.ID, &t.Name, &t.DisplayName, &t.CreatedAt)
		if err != nil {
			continue
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

// TenantMember represents a tenant member.
type TenantMember struct {
	TenantID string    `json:"tenant_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// generateTenantID generates a unique tenant ID.
func generateTenantID() string {
	return fmt.Sprintf("tenant-%d", time.Now().UnixNano())
}
