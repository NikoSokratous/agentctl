-- Risk scoring and analytics tables for v0.5

-- Risk assessments table
CREATE TABLE IF NOT EXISTS risk_assessments (
    id TEXT PRIMARY KEY,
    run_id TEXT,
    action_context JSON NOT NULL,
    risk_score JSON NOT NULL,
    decision TEXT NOT NULL CHECK(decision IN ('allow', 'allow_with_log', 'require_review', 'deny')),
    timestamp TIMESTAMP NOT NULL,
    assessor_id TEXT NOT NULL,
    version TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES runs(run_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_risk_timestamp ON risk_assessments(timestamp);
CREATE INDEX IF NOT EXISTS idx_risk_decision ON risk_assessments(decision);
CREATE INDEX IF NOT EXISTS idx_risk_run ON risk_assessments(run_id);
CREATE INDEX IF NOT EXISTS idx_risk_score ON risk_assessments(json_extract(risk_score, '$.score'));
CREATE INDEX IF NOT EXISTS idx_risk_level ON risk_assessments(json_extract(risk_score, '$.level'));

-- Compliance reports table
CREATE TABLE IF NOT EXISTS compliance_reports (
    id TEXT PRIMARY KEY,
    report_type TEXT NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    generated_at TIMESTAMP NOT NULL,
    total_actions INTEGER NOT NULL,
    high_risk_count INTEGER NOT NULL,
    denied_count INTEGER NOT NULL,
    approval_count INTEGER NOT NULL,
    summary JSON NOT NULL,
    findings JSON,
    recommendations JSON,
    metadata JSON
);

CREATE INDEX IF NOT EXISTS idx_compliance_period ON compliance_reports(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_compliance_type ON compliance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_compliance_generated ON compliance_reports(generated_at);

-- Risk anomalies table
CREATE TABLE IF NOT EXISTS risk_anomalies (
    id TEXT PRIMARY KEY,
    assessment_id TEXT NOT NULL,
    anomaly_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK(severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    FOREIGN KEY (assessment_id) REFERENCES risk_assessments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_anomaly_detected ON risk_anomalies(detected_at);
CREATE INDEX IF NOT EXISTS idx_anomaly_resolved ON risk_anomalies(resolved);
CREATE INDEX IF NOT EXISTS idx_anomaly_severity ON risk_anomalies(severity);
