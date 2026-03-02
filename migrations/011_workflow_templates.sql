-- Migration 011: Workflow Templates and Debug Sessions
-- Supports workflow marketplace and debugging features

-- Workflow templates table
CREATE TABLE IF NOT EXISTS workflow_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    tags TEXT NOT NULL, -- JSON array
    template_yaml TEXT NOT NULL,
    parameters TEXT NOT NULL, -- JSON array
    version TEXT NOT NULL,
    license TEXT NOT NULL,
    rating REAL DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    downloads INTEGER DEFAULT 0,
    published_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    metadata TEXT NOT NULL, -- JSON object
    tenant_id TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workflow_templates_category ON workflow_templates(category);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_author ON workflow_templates(author);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_downloads ON workflow_templates(downloads DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_rating ON workflow_templates(rating DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_tenant ON workflow_templates(tenant_id);

-- Debug sessions table
CREATE TABLE IF NOT EXISTS debug_sessions (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    current_step TEXT,
    paused INTEGER DEFAULT 0,
    variables TEXT NOT NULL, -- JSON object
    breakpoints TEXT NOT NULL, -- JSON object
    step_history TEXT NOT NULL, -- JSON array
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    tenant_id TEXT,
    FOREIGN KEY (workflow_id) REFERENCES workflow_states(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_debug_sessions_workflow ON debug_sessions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_debug_sessions_user ON debug_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_debug_sessions_tenant ON debug_sessions(tenant_id);

-- Template ratings table (for individual user ratings)
CREATE TABLE IF NOT EXISTS template_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (template_id) REFERENCES workflow_templates(id) ON DELETE CASCADE,
    UNIQUE (template_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_template_ratings_template ON template_ratings(template_id);
CREATE INDEX IF NOT EXISTS idx_template_ratings_user ON template_ratings(user_id);

-- Template comments table
CREATE TABLE IF NOT EXISTS template_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (template_id) REFERENCES workflow_templates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_template_comments_template ON template_comments(template_id);
CREATE INDEX IF NOT EXISTS idx_template_comments_created ON template_comments(created_at DESC);
