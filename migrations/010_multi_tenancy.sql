-- Multi-tenancy schema for v0.6

CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT,
    created_at TIMESTAMP NOT NULL,
    settings JSON
);

CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_created ON tenants(created_at);

CREATE TABLE IF NOT EXISTS tenant_members (
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('owner', 'admin', 'member', 'viewer')),
    joined_at TIMESTAMP NOT NULL,
    PRIMARY KEY (tenant_id, user_id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_role ON tenant_members(role);

-- Add tenant_id to existing tables
ALTER TABLE runs ADD COLUMN IF NOT EXISTS tenant_id TEXT;
ALTER TABLE policy_versions ADD COLUMN IF NOT EXISTS tenant_id TEXT;
ALTER TABLE risk_assessments ADD COLUMN IF NOT EXISTS tenant_id TEXT;
ALTER TABLE workflow_states ADD COLUMN IF NOT EXISTS tenant_id TEXT;
ALTER TABLE run_snapshots ADD COLUMN IF NOT EXISTS tenant_id TEXT;

-- Create indexes for tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_runs_tenant ON runs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_policy_versions_tenant ON policy_versions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_tenant ON risk_assessments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_workflow_states_tenant ON workflow_states(tenant_id);
CREATE INDEX IF NOT EXISTS idx_run_snapshots_tenant ON run_snapshots(tenant_id);

-- Sessions table (for authentication)
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    user_email TEXT NOT NULL,
    user_name TEXT,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL,
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
