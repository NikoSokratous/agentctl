-- Workflow state persistence for v0.6

CREATE TABLE IF NOT EXISTS workflow_states (
    id TEXT PRIMARY KEY,
    workflow_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('running', 'completed', 'failed', 'cancelled')),
    current_step TEXT,
    outputs JSON,
    started_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_workflow_states_status ON workflow_states(status);
CREATE INDEX IF NOT EXISTS idx_workflow_states_updated ON workflow_states(updated_at);
CREATE INDEX IF NOT EXISTS idx_workflow_states_name ON workflow_states(workflow_name);

CREATE TABLE IF NOT EXISTS workflow_step_states (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    step_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed', 'skipped')),
    output JSON,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (workflow_id) REFERENCES workflow_states(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workflow_step_states_workflow ON workflow_step_states(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_states_status ON workflow_step_states(status);
CREATE INDEX IF NOT EXISTS idx_workflow_step_states_started ON workflow_step_states(started_at);
