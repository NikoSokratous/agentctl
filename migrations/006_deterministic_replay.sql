-- Deterministic replay tables for v0.5

-- Run snapshots table
CREATE TABLE IF NOT EXISTS run_snapshots (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    agent_name TEXT NOT NULL,
    goal TEXT,
    agent_config JSON,
    model_calls JSON NOT NULL,
    tool_calls JSON NOT NULL,
    environment JSON,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    final_state TEXT NOT NULL CHECK(final_state IN ('completed', 'failed', 'cancelled')),
    checksums JSON,
    compressed BOOLEAN DEFAULT false,
    encrypted BOOLEAN DEFAULT false,
    size_bytes INTEGER,
    FOREIGN KEY (run_id) REFERENCES runs(run_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snapshots_run ON run_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_created ON run_snapshots(created_at);
CREATE INDEX IF NOT EXISTS idx_snapshots_agent ON run_snapshots(agent_name);
CREATE INDEX IF NOT EXISTS idx_snapshots_state ON run_snapshots(final_state);

-- Freeze points table
CREATE TABLE IF NOT EXISTS freeze_points (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    snapshot_id TEXT,
    sequence INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    reason TEXT NOT NULL,
    state JSON NOT NULL,
    pending_action JSON NOT NULL,
    options JSON NOT NULL,
    decision JSON,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    FOREIGN KEY (run_id) REFERENCES runs(run_id) ON DELETE CASCADE,
    FOREIGN KEY (snapshot_id) REFERENCES run_snapshots(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_freeze_run ON freeze_points(run_id);
CREATE INDEX IF NOT EXISTS idx_freeze_resolved ON freeze_points(resolved);
CREATE INDEX IF NOT EXISTS idx_freeze_timestamp ON freeze_points(timestamp);

-- Replay results table
CREATE TABLE IF NOT EXISTS replay_results (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    mode TEXT NOT NULL CHECK(mode IN ('exact', 'live', 'mixed', 'debug', 'validation')),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    duration_ms INTEGER,
    success BOOLEAN DEFAULT false,
    actions_rerun INTEGER DEFAULT 0,
    matches INTEGER DEFAULT 0,
    divergences JSON,
    final_state JSON,
    error TEXT,
    metrics JSON,
    FOREIGN KEY (snapshot_id) REFERENCES run_snapshots(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_replay_snapshot ON replay_results(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_replay_started ON replay_results(started_at);
CREATE INDEX IF NOT EXISTS idx_replay_success ON replay_results(success);
CREATE INDEX IF NOT EXISTS idx_replay_mode ON replay_results(mode);
