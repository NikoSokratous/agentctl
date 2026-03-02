package risk

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// AggregatedRiskStats provides risk analytics.
type AggregatedRiskStats struct {
	Period           string               `json:"period"` // day, week, month
	TotalAssessments int                  `json:"total_assessments"`
	AverageScore     float64              `json:"average_score"`
	ByLevel          map[RiskLevel]int    `json:"by_level"`
	ByCategory       map[string]float64   `json:"by_category"`
	ByTool           map[string]float64   `json:"by_tool"`
	ByEnvironment    map[string]float64   `json:"by_environment"`
	ByDecision       map[RiskDecision]int `json:"by_decision"`
	TopRiskyActions  []RiskAssessment     `json:"top_risky_actions"`
	TrendData        []TrendPoint         `json:"trend_data"`
}

// TrendPoint represents a data point in risk trends.
type TrendPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	AverageScore float64   `json:"average_score"`
	Count        int       `json:"count"`
}

// RiskAnalyzer provides analytics and reporting.
type RiskAnalyzer struct {
	db *sql.DB
}

// NewRiskAnalyzer creates a new risk analyzer.
func NewRiskAnalyzer(db *sql.DB) *RiskAnalyzer {
	return &RiskAnalyzer{db: db}
}

// GetStats retrieves aggregated risk statistics.
func (a *RiskAnalyzer) GetStats(ctx context.Context, startTime, endTime time.Time) (*AggregatedRiskStats, error) {
	stats := &AggregatedRiskStats{
		Period:        fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		ByLevel:       make(map[RiskLevel]int),
		ByCategory:    make(map[string]float64),
		ByTool:        make(map[string]float64),
		ByEnvironment: make(map[string]float64),
		ByDecision:    make(map[RiskDecision]int),
		TrendData:     make([]TrendPoint, 0),
	}

	// Query risk assessments from database
	query := `
		SELECT 
			action_context,
			risk_score,
			decision,
			timestamp
		FROM risk_assessments
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
	`

	rows, err := a.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("query risk assessments: %w", err)
	}
	defer rows.Close()

	totalScore := 0.0
	count := 0

	for rows.Next() {
		var actionCtxJSON, riskScoreJSON string
		var decision string
		var timestamp time.Time

		if err := rows.Scan(&actionCtxJSON, &riskScoreJSON, &decision, &timestamp); err != nil {
			continue
		}

		var riskScore RiskScore
		if err := json.Unmarshal([]byte(riskScoreJSON), &riskScore); err != nil {
			continue
		}

		var actionCtx ActionContext
		if err := json.Unmarshal([]byte(actionCtxJSON), &actionCtx); err != nil {
			continue
		}

		// Aggregate statistics
		count++
		totalScore += riskScore.Score

		stats.ByLevel[riskScore.Level]++
		stats.ByDecision[RiskDecision(decision)]++

		if actionCtx.ToolName != "" {
			stats.ByTool[actionCtx.ToolName] += riskScore.Score
		}
		if actionCtx.Environment != "" {
			stats.ByEnvironment[actionCtx.Environment] += riskScore.Score
		}

		for category, score := range riskScore.Breakdown {
			stats.ByCategory[category] += score
		}
	}

	if count > 0 {
		stats.TotalAssessments = count
		stats.AverageScore = totalScore / float64(count)

		// Normalize category scores
		for category := range stats.ByCategory {
			stats.ByCategory[category] /= float64(count)
		}
		for tool := range stats.ByTool {
			stats.ByTool[tool] /= float64(count)
		}
		for env := range stats.ByEnvironment {
			stats.ByEnvironment[env] /= float64(count)
		}
	}

	return stats, nil
}

// GetTopRiskyActions retrieves highest risk assessments.
func (a *RiskAnalyzer) GetTopRiskyActions(ctx context.Context, limit int, minScore float64) ([]RiskAssessment, error) {
	query := `
		SELECT 
			action_context,
			risk_score,
			decision,
			timestamp,
			assessor_id,
			version
		FROM risk_assessments
		WHERE json_extract(risk_score, '$.score') >= ?
		ORDER BY json_extract(risk_score, '$.score') DESC
		LIMIT ?
	`

	rows, err := a.db.QueryContext(ctx, query, minScore, limit)
	if err != nil {
		return nil, fmt.Errorf("query risky actions: %w", err)
	}
	defer rows.Close()

	assessments := make([]RiskAssessment, 0, limit)

	for rows.Next() {
		var actionCtxJSON, riskScoreJSON, decision, assessorID, version string
		var timestamp time.Time

		if err := rows.Scan(&actionCtxJSON, &riskScoreJSON, &decision, &timestamp, &assessorID, &version); err != nil {
			continue
		}

		var assessment RiskAssessment
		json.Unmarshal([]byte(actionCtxJSON), &assessment.ActionContext)
		json.Unmarshal([]byte(riskScoreJSON), &assessment.RiskScore)
		assessment.Decision = RiskDecision(decision)
		assessment.Timestamp = timestamp
		assessment.AssessorID = assessorID
		assessment.Version = version

		assessments = append(assessments, assessment)
	}

	return assessments, nil
}

// GetTrend retrieves risk score trends over time.
func (a *RiskAnalyzer) GetTrend(ctx context.Context, startTime, endTime time.Time, interval string) ([]TrendPoint, error) {
	// Group by interval (hour, day, week)
	var groupBy string
	switch interval {
	case "hour":
		groupBy = "strftime('%Y-%m-%d %H:00', timestamp)"
	case "day":
		groupBy = "strftime('%Y-%m-%d', timestamp)"
	case "week":
		groupBy = "strftime('%Y-%W', timestamp)"
	default:
		groupBy = "strftime('%Y-%m-%d', timestamp)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as period,
			AVG(json_extract(risk_score, '$.score')) as avg_score,
			COUNT(*) as count
		FROM risk_assessments
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY period
		ORDER BY period ASC
	`, groupBy)

	rows, err := a.db.QueryContext(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("query trend: %w", err)
	}
	defer rows.Close()

	trends := make([]TrendPoint, 0)

	for rows.Next() {
		var period string
		var avgScore float64
		var count int

		if err := rows.Scan(&period, &avgScore, &count); err != nil {
			continue
		}

		// Parse timestamp from period
		timestamp, _ := time.Parse("2006-01-02", period)

		trends = append(trends, TrendPoint{
			Timestamp:    timestamp,
			AverageScore: avgScore,
			Count:        count,
		})
	}

	return trends, nil
}

// RiskStore manages risk assessment persistence.
type RiskStore struct {
	db *sql.DB
}

// NewRiskStore creates a new risk store.
func NewRiskStore(db *sql.DB) *RiskStore {
	return &RiskStore{db: db}
}

// Save persists a risk assessment.
func (s *RiskStore) Save(ctx context.Context, assessment *RiskAssessment) error {
	actionCtxJSON, err := json.Marshal(assessment.ActionContext)
	if err != nil {
		return fmt.Errorf("marshal action context: %w", err)
	}

	riskScoreJSON, err := json.Marshal(assessment.RiskScore)
	if err != nil {
		return fmt.Errorf("marshal risk score: %w", err)
	}

	query := `
		INSERT INTO risk_assessments (
			id, action_context, risk_score, decision, timestamp, assessor_id, version
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		generateID(),
		string(actionCtxJSON),
		string(riskScoreJSON),
		string(assessment.Decision),
		assessment.Timestamp,
		assessment.AssessorID,
		assessment.Version,
	)

	return err
}

// Get retrieves a risk assessment by ID.
func (s *RiskStore) Get(ctx context.Context, id string) (*RiskAssessment, error) {
	query := `
		SELECT action_context, risk_score, decision, timestamp, assessor_id, version
		FROM risk_assessments
		WHERE id = ?
	`

	var actionCtxJSON, riskScoreJSON, decision, assessorID, version string
	var timestamp time.Time

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&actionCtxJSON, &riskScoreJSON, &decision, &timestamp, &assessorID, &version,
	)
	if err != nil {
		return nil, err
	}

	var assessment RiskAssessment
	json.Unmarshal([]byte(actionCtxJSON), &assessment.ActionContext)
	json.Unmarshal([]byte(riskScoreJSON), &assessment.RiskScore)
	assessment.Decision = RiskDecision(decision)
	assessment.Timestamp = timestamp
	assessment.AssessorID = assessorID
	assessment.Version = version

	return &assessment, nil
}

// generateID generates a unique ID (placeholder).
func generateID() string {
	return fmt.Sprintf("risk-%d", time.Now().UnixNano())
}
