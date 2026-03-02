-- Migration 013: Security & Reliability
-- Supports audit log encryption, backups, and ML model registry

-- Encrypted audit logs table
CREATE TABLE IF NOT EXISTS audit_logs_encrypted (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    user_id TEXT,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    resource_id TEXT,
    status TEXT NOT NULL,
    details_encrypted TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL,
    encryption_key_version INTEGER NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant ON audit_logs_encrypted(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs_encrypted(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs_encrypted(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs_encrypted(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs_encrypted(timestamp DESC);

-- Encryption keys table
CREATE TABLE IF NOT EXISTS encryption_keys (
    id TEXT PRIMARY KEY,
    version INTEGER UNIQUE NOT NULL,
    key_hash TEXT NOT NULL, -- SHA-256 hash for verification
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    active INTEGER DEFAULT 1,
    revoked_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_encryption_keys_version ON encryption_keys(version DESC);
CREATE INDEX IF NOT EXISTS idx_encryption_keys_active ON encryption_keys(active);

-- Backup metadata table
CREATE TABLE IF NOT EXISTS backup_metadata (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- full, incremental, wal
    size INTEGER NOT NULL,
    location TEXT NOT NULL, -- Storage path (S3, GCS, etc.)
    created_at TIMESTAMP NOT NULL,
    status TEXT NOT NULL, -- pending, completed, failed
    checksum TEXT,
    compressed INTEGER DEFAULT 1,
    encrypted INTEGER DEFAULT 0,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_backup_metadata_type ON backup_metadata(type);
CREATE INDEX IF NOT EXISTS idx_backup_metadata_created ON backup_metadata(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_metadata_status ON backup_metadata(status);

-- Replication status table
CREATE TABLE IF NOT EXISTS replication_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replica_name TEXT NOT NULL,
    region TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    priority INTEGER NOT NULL,
    healthy INTEGER DEFAULT 1,
    last_check TIMESTAMP NOT NULL,
    lag_seconds INTEGER,
    last_sync TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_replication_replica ON replication_status(replica_name);
CREATE INDEX IF NOT EXISTS idx_replication_healthy ON replication_status(healthy);

-- ML models table
CREATE TABLE IF NOT EXISTS ml_models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    framework TEXT NOT NULL, -- tensorflow, pytorch, sklearn
    description TEXT,
    tags TEXT, -- JSON
    metrics TEXT, -- JSON
    parameters TEXT, -- JSON
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    status TEXT NOT NULL, -- draft, ready, deprecated
    UNIQUE (name, version)
);

CREATE INDEX IF NOT EXISTS idx_ml_models_name ON ml_models(name);
CREATE INDEX IF NOT EXISTS idx_ml_models_status ON ml_models(status);
CREATE INDEX IF NOT EXISTS idx_ml_models_framework ON ml_models(framework);

-- Model deployments table
CREATE TABLE IF NOT EXISTS model_deployments (
    id TEXT PRIMARY KEY,
    model_id TEXT NOT NULL,
    model_version TEXT NOT NULL,
    environment TEXT NOT NULL, -- dev, staging, prod
    endpoint TEXT NOT NULL,
    replicas INTEGER NOT NULL,
    status TEXT NOT NULL, -- deploying, running, failed, stopped
    health_check TEXT,
    created_at TIMESTAMP NOT NULL,
    deployed_at TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES ml_models(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_model_deployments_model ON model_deployments(model_id);
CREATE INDEX IF NOT EXISTS idx_model_deployments_env ON model_deployments(environment);
CREATE INDEX IF NOT EXISTS idx_model_deployments_status ON model_deployments(status);

-- Model metrics table
CREATE TABLE IF NOT EXISTS model_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id TEXT NOT NULL,
    version TEXT NOT NULL,
    accuracy REAL,
    precision_score REAL,
    recall_score REAL,
    f1_score REAL,
    latency_ms INTEGER,
    throughput REAL,
    error_rate REAL,
    recorded_at TIMESTAMP NOT NULL,
    FOREIGN KEY (model_id) REFERENCES ml_models(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_model_metrics_model ON model_metrics(model_id);
CREATE INDEX IF NOT EXISTS idx_model_metrics_recorded ON model_metrics(recorded_at DESC);

-- A/B test configurations table
CREATE TABLE IF NOT EXISTS ab_tests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    model_a TEXT NOT NULL,
    model_b TEXT NOT NULL,
    traffic_split REAL NOT NULL, -- 0.0 to 1.0
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    metrics TEXT NOT NULL, -- JSON array
    active INTEGER DEFAULT 1,
    winner TEXT,
    FOREIGN KEY (model_a) REFERENCES ml_models(id),
    FOREIGN KEY (model_b) REFERENCES ml_models(id)
);

CREATE INDEX IF NOT EXISTS idx_ab_tests_active ON ab_tests(active);
CREATE INDEX IF NOT EXISTS idx_ab_tests_start ON ab_tests(start_time DESC);

-- GraphQL query cache table
CREATE TABLE IF NOT EXISTS graphql_query_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query_hash TEXT NOT NULL,
    query TEXT NOT NULL,
    variables TEXT,
    result TEXT NOT NULL,
    cached_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    hits INTEGER DEFAULT 0,
    UNIQUE (query_hash)
);

CREATE INDEX IF NOT EXISTS idx_graphql_cache_hash ON graphql_query_cache(query_hash);
CREATE INDEX IF NOT EXISTS idx_graphql_cache_expires ON graphql_query_cache(expires_at);

-- WAL archive metadata table
CREATE TABLE IF NOT EXISTS wal_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL UNIQUE,
    location TEXT NOT NULL,
    size INTEGER NOT NULL,
    archived_at TIMESTAMP NOT NULL,
    checksum TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wal_archive_archived ON wal_archive(archived_at DESC);

-- Disaster recovery configurations table
CREATE TABLE IF NOT EXISTS dr_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- backup, replication, failover
    schedule TEXT, -- Cron expression
    enabled INTEGER DEFAULT 1,
    config TEXT NOT NULL, -- JSON configuration
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dr_configs_type ON dr_configs(type);
CREATE INDEX IF NOT EXISTS idx_dr_configs_enabled ON dr_configs(enabled);
CREATE INDEX IF NOT EXISTS idx_dr_configs_next_run ON dr_configs(next_run);
