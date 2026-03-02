-- Migration 012: Cost, Performance & Analytics
-- Supports cost tracking, SLA monitoring, and analytics

-- Cost entries table
CREATE TABLE IF NOT EXISTS cost_entries (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    user_id TEXT,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    cost REAL NOT NULL,
    call_count INTEGER DEFAULT 1,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cost_entries_agent ON cost_entries(agent_id);
CREATE INDEX IF NOT EXISTS idx_cost_entries_tenant ON cost_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cost_entries_user ON cost_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_cost_entries_timestamp ON cost_entries(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_cost_entries_provider ON cost_entries(provider, model);

-- Budgets table
CREATE TABLE IF NOT EXISTS budgets (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    agent_id TEXT,
    user_id TEXT,
    limit_amount REAL NOT NULL,
    period TEXT NOT NULL, -- daily, weekly, monthly
    alert_threshold REAL NOT NULL, -- 0.0 to 1.0
    enabled INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_budgets_tenant ON budgets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_budgets_agent ON budgets(agent_id);

-- SLA metrics table
CREATE TABLE IF NOT EXISTS sla_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    success INTEGER NOT NULL, -- 0 or 1
    latency_ms INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sla_metrics_service ON sla_metrics(service_id);
CREATE INDEX IF NOT EXISTS idx_sla_metrics_tenant ON sla_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sla_metrics_timestamp ON sla_metrics(timestamp DESC);

-- Uptime records table
CREATE TABLE IF NOT EXISTS uptime_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    uptime_seconds REAL NOT NULL,
    downtime_seconds REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_uptime_records_service ON uptime_records(service_id);
CREATE INDEX IF NOT EXISTS idx_uptime_records_timestamp ON uptime_records(timestamp DESC);

-- SLA targets table
CREATE TABLE IF NOT EXISTS sla_targets (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    uptime_target REAL NOT NULL, -- Percentage (e.g., 99.9)
    latency_target_p50_ms INTEGER NOT NULL,
    latency_target_p95_ms INTEGER NOT NULL,
    latency_target_p99_ms INTEGER NOT NULL,
    error_rate_target REAL NOT NULL, -- 0.0 to 1.0
    alert_on_violation INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sla_targets_service ON sla_targets(service_id);

-- Scaling policies table
CREATE TABLE IF NOT EXISTS scaling_policies (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    min_replicas INTEGER NOT NULL,
    max_replicas INTEGER NOT NULL,
    scale_up_threshold REAL NOT NULL,
    scale_down_threshold REAL NOT NULL,
    cooldown_period_seconds INTEGER NOT NULL,
    enabled INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scaling_policies_service ON scaling_policies(service_id);

-- Scaling events table
CREATE TABLE IF NOT EXISTS scaling_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    old_replicas INTEGER NOT NULL,
    new_replicas INTEGER NOT NULL,
    reason TEXT NOT NULL,
    metric_value REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (policy_id) REFERENCES scaling_policies(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scaling_events_policy ON scaling_events(policy_id);
CREATE INDEX IF NOT EXISTS idx_scaling_events_timestamp ON scaling_events(timestamp DESC);

-- Performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    metric_type TEXT NOT NULL, -- cpu, memory, throughput, etc.
    metric_value REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_agent ON performance_metrics(agent_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type ON performance_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_timestamp ON performance_metrics(timestamp DESC);

-- Analytics aggregates table (for faster dashboard queries)
CREATE TABLE IF NOT EXISTS analytics_aggregates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id TEXT NOT NULL,
    period TEXT NOT NULL, -- hourly, daily, weekly, monthly
    period_start TIMESTAMP NOT NULL,
    total_cost REAL NOT NULL,
    total_requests INTEGER NOT NULL,
    avg_latency_ms REAL NOT NULL,
    error_rate REAL NOT NULL,
    active_agents INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE (tenant_id, period, period_start)
);

CREATE INDEX IF NOT EXISTS idx_analytics_aggregates_tenant ON analytics_aggregates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_analytics_aggregates_period ON analytics_aggregates(period, period_start DESC);
