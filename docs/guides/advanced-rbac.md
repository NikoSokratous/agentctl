# Advanced RBAC

Fine-grained roles, org hierarchy, and delegation.

## Role Templates

Create custom roles derived from base roles:

```go
engine := tenancy.NewAdvancedRBAC()
engine.RegisterRoleTemplate(&tenancy.RoleTemplate{
    Name:        "developer",
    BaseRole:    tenancy.RoleMember,
    ExtraPerms:  []tenancy.Permission{tenancy.PermissionManageTenant},
    RevokedPerms: []tenancy.Permission{},
    Description: "Developer with tenant config access",
})
has := engine.HasPermissionWithTemplate("developer", tenancy.PermissionManageTenant) // true
```

## Org Hierarchy

Org units support parent-child relationships:

```go
type OrgUnit struct {
    ID       string
    TenantID string
    ParentID string  // Empty for root
    Name     string
}
```

Use for scoped access (e.g. permit access only within an org subtree).

## Delegation

Users can delegate permissions to others:

```go
type Delegation struct {
    TenantID    string
    DelegatorID string
    DelegateeID string
    Permission  Permission
    Resource    string  // Optional scope
    ExpiresAt   *string
}
```

Implement `DelegationStore` to persist delegations and use `ValidateAccessWithDelegation`.

## Schema

| Table | Purpose |
|-------|---------|
| tenant_members | user_id, tenant_id, role |
| role_templates | name, base_role, extra_permissions, revoked_permissions |
| org_units | id, tenant_id, parent_id, name |
| delegations | delegator_id, delegatee_id, permission, expires_at |

## UI for Role Assignment

Use `TenantManager.ListMembers` and `AddMember`/`RemoveMember` for role assignment. Expose via API:

- `GET /tenants/:id/members` – list members
- `POST /tenants/:id/members` – add member with role
- `DELETE /tenants/:id/members/:userId` – remove member
