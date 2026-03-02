-- Policy versioning tables for v0.5

-- Policy versions table
CREATE TABLE IF NOT EXISTS policy_versions (
    id TEXT PRIMARY KEY,
    policy_name TEXT NOT NULL,
    version TEXT NOT NULL,
    content BLOB NOT NULL,
    format TEXT DEFAULT 'yaml',
    author TEXT,
    changelog TEXT,
    effective_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    supersedes TEXT,
    active BOOLEAN DEFAULT false,
    metadata JSON,
    UNIQUE(policy_name, version),
    FOREIGN KEY (supersedes) REFERENCES policy_versions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_policy_versions_name ON policy_versions(policy_name);
CREATE INDEX IF NOT EXISTS idx_policy_versions_active ON policy_versions(policy_name, active);
CREATE INDEX IF NOT EXISTS idx_policy_versions_created ON policy_versions(created_at);

-- Policy audit log table
CREATE TABLE IF NOT EXISTS policy_audit (
    id TEXT PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    run_id TEXT,
    agent_name TEXT NOT NULL,
    policy_name TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    action TEXT NOT NULL,
    tool TEXT NOT NULL,
    decision TEXT NOT NULL CHECK(decision IN ('allow', 'deny', 'alert')),
    risk_score REAL,
    deny_reason TEXT,
    context JSON,
    reviewed_by TEXT,
    reviewed_at TIMESTAMP,
    FOREIGN KEY (run_id) REFERENCES runs(run_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_policy_audit_timestamp ON policy_audit(timestamp);
CREATE INDEX IF NOT EXISTS idx_policy_audit_run ON policy_audit(run_id);
CREATE INDEX IF NOT EXISTS idx_policy_audit_agent ON policy_audit(agent_name);
CREATE INDEX IF NOT EXISTS idx_policy_audit_decision ON policy_audit(decision);
CREATE INDEX IF NOT EXISTS idx_policy_audit_risk ON policy_audit(risk_score);
CREATE INDEX IF NOT EXISTS idx_policy_audit_policy ON policy_audit(policy_name, policy_version);

-- Policy simulation results table
CREATE TABLE IF NOT EXISTS policy_simulations (
    id TEXT PRIMARY KEY,
    policy_name TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    run_id TEXT,
    mode TEXT NOT NULL CHECK(mode IN ('audit', 'simulation', 'shadow')),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    total_actions INTEGER DEFAULT 0,
    allowed INTEGER DEFAULT 0,
    denied INTEGER DEFAULT 0,
    alerts INTEGER DEFAULT 0,
    results JSON,
    FOREIGN KEY (run_id) REFERENCES runs(run_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_policy_simulations_policy ON policy_simulations(policy_name, policy_version);
CREATE INDEX IF NOT EXISTS idx_policy_simulations_run ON policy_simulations(run_id);
CREATE INDEX IF NOT EXISTS idx_policy_simulations_started ON policy_simulations(started_at);
