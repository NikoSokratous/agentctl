-- Migration 014: Workflow Versioning
-- Multiple versions per workflow template for version management

CREATE TABLE IF NOT EXISTS workflow_versions (
    id TEXT PRIMARY KEY,
    workflow_name TEXT NOT NULL,
    version TEXT NOT NULL,
    template_yaml TEXT NOT NULL,
    parameters TEXT, -- JSON array
    author TEXT,
    changelog TEXT,
    effective_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    metadata TEXT, -- JSON
    UNIQUE(workflow_name, version)
);

CREATE INDEX IF NOT EXISTS idx_workflow_versions_name ON workflow_versions(workflow_name);
CREATE INDEX IF NOT EXISTS idx_workflow_versions_created ON workflow_versions(created_at DESC);
