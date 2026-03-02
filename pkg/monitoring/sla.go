package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// SLAMonitor tracks SLA metrics (uptime, latency, error rate)
type SLAMonitor struct {
	db            *sql.DB
	metrics       map[string]*SLAMetrics
	targets       map[string]*SLATarget
	mu            sync.RWMutex
	flushInterval time.Duration
	stopChan      chan struct{}
}

// SLAMetrics holds current SLA metrics
type SLAMetrics struct {
	ServiceID       string
	TenantID        string
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalLatency    time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	LastUpdate      time.Time
	Uptime          time.Duration
	Downtime        time.Duration
}

// SLATarget defines SLA targets for a service
type SLATarget struct {
	ServiceID          string
	TenantID           string
	UptimeTarget       float64 // e.g., 99.9 for 99.9%
	LatencyTargetP50   time.Duration
	LatencyTargetP95   time.Duration
	LatencyTargetP99   time.Duration
	ErrorRateTarget    float64 // e.g., 0.01 for 1%
	AlertOnViolation   bool
	ViolationThreshold int // Number of violations before alert
}

// SLAReport represents an SLA report
type SLAReport struct {
	ServiceID       string
	TenantID        string
	Period          string
	Uptime          float64
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	ErrorRate       float64
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	SLAViolations   []string
	GeneratedAt     time.Time
}

// NewSLAMonitor creates a new SLA monitor
func NewSLAMonitor(db *sql.DB) *SLAMonitor {
	return &SLAMonitor{
		db:            db,
		metrics:       make(map[string]*SLAMetrics),
		targets:       make(map[string]*SLATarget),
		flushInterval: 60 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the SLA monitor
func (sm *SLAMonitor) Start(ctx context.Context) {
	go sm.flushLoop(ctx)
	go sm.checkSLALoop(ctx)
}

// Stop stops the SLA monitor
func (sm *SLAMonitor) Stop() {
	close(sm.stopChan)
}

// SetTarget sets an SLA target for a service
func (sm *SLAMonitor) SetTarget(target *SLATarget) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", target.TenantID, target.ServiceID)
	sm.targets[key] = target
}

// RecordRequest records a request and its outcome
func (sm *SLAMonitor) RecordRequest(ctx context.Context, serviceID, tenantID string, latency time.Duration, success bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", tenantID, serviceID)
	metrics, exists := sm.metrics[key]
	if !exists {
		metrics = &SLAMetrics{
			ServiceID:  serviceID,
			TenantID:   tenantID,
			MinLatency: latency,
			MaxLatency: latency,
			LastUpdate: time.Now(),
		}
		sm.metrics[key] = metrics
	}

	metrics.TotalRequests++
	if success {
		metrics.SuccessRequests++
	} else {
		metrics.FailedRequests++
	}

	metrics.TotalLatency += latency
	if latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}
	metrics.LastUpdate = time.Now()
}

// RecordUptime records uptime/downtime
func (sm *SLAMonitor) RecordUptime(serviceID, tenantID string, duration time.Duration, available bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", tenantID, serviceID)
	metrics, exists := sm.metrics[key]
	if !exists {
		metrics = &SLAMetrics{
			ServiceID:  serviceID,
			TenantID:   tenantID,
			LastUpdate: time.Now(),
		}
		sm.metrics[key] = metrics
	}

	if available {
		metrics.Uptime += duration
	} else {
		metrics.Downtime += duration
	}
	metrics.LastUpdate = time.Now()
}

// GetMetrics returns current metrics for a service
func (sm *SLAMonitor) GetMetrics(serviceID, tenantID string) (*SLAMetrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", tenantID, serviceID)
	metrics, exists := sm.metrics[key]
	if !exists {
		return nil, fmt.Errorf("metrics not found for service: %s", serviceID)
	}

	// Return a copy
	return &SLAMetrics{
		ServiceID:       metrics.ServiceID,
		TenantID:        metrics.TenantID,
		TotalRequests:   metrics.TotalRequests,
		SuccessRequests: metrics.SuccessRequests,
		FailedRequests:  metrics.FailedRequests,
		TotalLatency:    metrics.TotalLatency,
		MinLatency:      metrics.MinLatency,
		MaxLatency:      metrics.MaxLatency,
		LastUpdate:      metrics.LastUpdate,
		Uptime:          metrics.Uptime,
		Downtime:        metrics.Downtime,
	}, nil
}

// GenerateReport generates an SLA report
func (sm *SLAMonitor) GenerateReport(ctx context.Context, serviceID, tenantID string, start, end time.Time) (*SLAReport, error) {
	// Query database for historical data
	query := `
		SELECT 
			COUNT(*) as total_requests,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as success_requests,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_requests,
			AVG(latency_ms) as avg_latency,
			MIN(latency_ms) as min_latency,
			MAX(latency_ms) as max_latency
		FROM sla_metrics
		WHERE service_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?
	`

	var totalRequests, successRequests, failedRequests int64
	var avgLatency, minLatency, maxLatency float64

	err := sm.db.QueryRowContext(ctx, query, serviceID, tenantID, start, end).Scan(
		&totalRequests, &successRequests, &failedRequests,
		&avgLatency, &minLatency, &maxLatency,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query metrics: %w", err)
	}

	// Calculate uptime
	_ = end.Sub(start) // totalTime calculated but not used - uptime % is calculated from uptime_records
	uptimeQuery := `
		SELECT COALESCE(SUM(uptime_seconds), 0), COALESCE(SUM(downtime_seconds), 0)
		FROM uptime_records
		WHERE service_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?
	`

	var uptimeSeconds, downtimeSeconds float64
	err = sm.db.QueryRowContext(ctx, uptimeQuery, serviceID, tenantID, start, end).Scan(
		&uptimeSeconds, &downtimeSeconds,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query uptime: %w", err)
	}

	// Calculate percentages
	var uptimePercent, errorRate float64
	if uptimeSeconds+downtimeSeconds > 0 {
		uptimePercent = (uptimeSeconds / (uptimeSeconds + downtimeSeconds)) * 100
	} else {
		uptimePercent = 100.0
	}

	if totalRequests > 0 {
		errorRate = float64(failedRequests) / float64(totalRequests)
	}

	// Check for SLA violations
	violations := sm.checkViolations(serviceID, tenantID, uptimePercent, errorRate, time.Duration(avgLatency)*time.Millisecond)

	report := &SLAReport{
		ServiceID:       serviceID,
		TenantID:        tenantID,
		Period:          fmt.Sprintf("%s to %s", start.Format(time.RFC3339), end.Format(time.RFC3339)),
		Uptime:          uptimePercent,
		AvgLatency:      time.Duration(avgLatency) * time.Millisecond,
		P50Latency:      time.Duration(avgLatency*0.5) * time.Millisecond, // Simplified
		P95Latency:      time.Duration(avgLatency*1.5) * time.Millisecond, // Simplified
		P99Latency:      time.Duration(avgLatency*2.0) * time.Millisecond, // Simplified
		ErrorRate:       errorRate,
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		SLAViolations:   violations,
		GeneratedAt:     time.Now(),
	}

	return report, nil
}

// checkViolations checks for SLA violations
func (sm *SLAMonitor) checkViolations(serviceID, tenantID string, uptime, errorRate float64, avgLatency time.Duration) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var violations []string
	key := fmt.Sprintf("%s:%s", tenantID, serviceID)
	target, exists := sm.targets[key]
	if !exists {
		return violations
	}

	if uptime < target.UptimeTarget {
		violations = append(violations, fmt.Sprintf("Uptime %.2f%% below target %.2f%%", uptime, target.UptimeTarget))
	}

	if errorRate > target.ErrorRateTarget {
		violations = append(violations, fmt.Sprintf("Error rate %.2f%% above target %.2f%%", errorRate*100, target.ErrorRateTarget*100))
	}

	if avgLatency > target.LatencyTargetP95 {
		violations = append(violations, fmt.Sprintf("Average latency %v above target %v", avgLatency, target.LatencyTargetP95))
	}

	return violations
}

// flushLoop periodically flushes metrics to database
func (sm *SLAMonitor) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(sm.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sm.flush()
			return
		case <-sm.stopChan:
			sm.flush()
			return
		case <-ticker.C:
			sm.flush()
		}
	}
}

// flush writes metrics to database
func (sm *SLAMonitor) flush() {
	sm.mu.Lock()
	metrics := make([]*SLAMetrics, 0, len(sm.metrics))
	for _, m := range sm.metrics {
		metrics = append(metrics, m)
	}
	// Don't clear metrics, keep accumulating
	sm.mu.Unlock()

	if len(metrics) == 0 {
		return
	}

	// Insert metrics (simplified)
	// In production, batch insert properly
	for _, m := range metrics {
		if m.TotalRequests > 0 {
			avgLatency := m.TotalLatency.Milliseconds() / m.TotalRequests

			// Insert request metrics
			_, _ = sm.db.Exec(`
				INSERT INTO sla_metrics (
					service_id, tenant_id, success, latency_ms, timestamp
				) VALUES (?, ?, ?, ?, ?)
			`, m.ServiceID, m.TenantID, 1, avgLatency, m.LastUpdate)
		}

		if m.Uptime > 0 || m.Downtime > 0 {
			// Insert uptime metrics
			_, _ = sm.db.Exec(`
				INSERT INTO uptime_records (
					service_id, tenant_id, uptime_seconds, downtime_seconds, timestamp
				) VALUES (?, ?, ?, ?, ?)
			`, m.ServiceID, m.TenantID, m.Uptime.Seconds(), m.Downtime.Seconds(), m.LastUpdate)
		}
	}
}

// checkSLALoop periodically checks for SLA violations
func (sm *SLAMonitor) checkSLALoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.checkAllSLAs(ctx)
		}
	}
}

// checkAllSLAs checks all SLA targets
func (sm *SLAMonitor) checkAllSLAs(ctx context.Context) {
	sm.mu.RLock()
	targets := make([]*SLATarget, 0, len(sm.targets))
	for _, target := range sm.targets {
		targets = append(targets, target)
	}
	sm.mu.RUnlock()

	for _, target := range targets {
		// Generate report for last hour
		end := time.Now()
		start := end.Add(-1 * time.Hour)

		report, err := sm.GenerateReport(ctx, target.ServiceID, target.TenantID, start, end)
		if err != nil {
			continue
		}

		if len(report.SLAViolations) > 0 && target.AlertOnViolation {
			// Send alert (simplified - in production, integrate with alerting system)
			fmt.Printf("SLA ALERT for %s: %v\n", target.ServiceID, report.SLAViolations)
		}
	}
}
