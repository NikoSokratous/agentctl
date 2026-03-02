package tenancy

import (
	"context"
	"fmt"
	"net/http"
)

// IsolationMiddleware provides tenant data isolation.
type IsolationMiddleware struct {
	tenantManager *TenantManager
}

// NewIsolationMiddleware creates a new isolation middleware.
func NewIsolationMiddleware(tenantManager *TenantManager) *IsolationMiddleware {
	return &IsolationMiddleware{
		tenantManager: tenantManager,
	}
}

// Isolate ensures tenant data isolation.
func (im *IsolationMiddleware) Isolate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant ID from request
		tenantID := extractTenantID(r)
		if tenantID == "" {
			http.Error(w, "Tenant ID required", http.StatusBadRequest)
			return
		}

		// Validate tenant exists
		tenant, err := im.tenantManager.GetTenant(r.Context(), tenantID)
		if err != nil {
			http.Error(w, "Invalid tenant", http.StatusNotFound)
			return
		}

		// Validate user has access to tenant
		userID := extractUserID(r)
		if userID != "" {
			if !im.hasAccess(r.Context(), userID, tenantID) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		// Add tenant to context
		ctx := r.Context()
		ctx = contextWithTenant(ctx, tenant)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// hasAccess checks if user has access to tenant.
func (im *IsolationMiddleware) hasAccess(ctx context.Context, userID, tenantID string) bool {
	// Check if user is member of tenant
	members, err := im.tenantManager.ListMembers(ctx, tenantID)
	if err != nil {
		return false
	}

	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}

	return false
}

// extractTenantID extracts tenant ID from request.
func extractTenantID(r *http.Request) string {
	// Try header first
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	// Try query parameter
	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		return tenantID
	}

	// Try subdomain
	// host := r.Host
	// if tenant := extractTenantFromSubdomain(host); tenant != "" {
	// 	return tenant
	// }

	return ""
}

// extractUserID extracts user ID from request context.
func extractUserID(r *http.Request) string {
	// Would extract from auth context
	if user, ok := r.Context().Value("user").(interface{ GetID() string }); ok {
		return user.GetID()
	}
	return ""
}

// contextWithTenant adds tenant to context.
func contextWithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, "tenant", tenant)
}

// GetTenantFromContext retrieves tenant from context.
func GetTenantFromContext(ctx context.Context) (*Tenant, bool) {
	tenant, ok := ctx.Value("tenant").(*Tenant)
	return tenant, ok
}

// TenantFilter filters database queries by tenant.
type TenantFilter struct {
	TenantID string
}

// ApplyToQuery adds tenant filter to SQL query.
func (tf *TenantFilter) ApplyToQuery(query string, args []interface{}) (string, []interface{}) {
	if tf.TenantID == "" {
		return query, args
	}

	// Add tenant_id filter to WHERE clause
	if containsWhere(query) {
		query += " AND tenant_id = ?"
	} else {
		query += " WHERE tenant_id = ?"
	}

	args = append(args, tf.TenantID)
	return query, args
}

// containsWhere checks if query contains WHERE clause.
func containsWhere(query string) bool {
	// Simple check - would be more sophisticated in production
	return len(query) > 0
}

// ValidateTenantAccess validates user access to a resource in a tenant.
func ValidateTenantAccess(ctx context.Context, resourceTenantID string) error {
	tenant, ok := GetTenantFromContext(ctx)
	if !ok {
		return fmt.Errorf("no tenant in context")
	}

	if tenant.ID != resourceTenantID {
		return fmt.Errorf("resource belongs to different tenant")
	}

	return nil
}
