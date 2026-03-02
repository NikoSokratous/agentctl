package policy

// RiskScorer computes a risk score for an action.
type RiskScorer interface {
	Score(tool string, input map[string]any) float64
}

// DefaultRiskScorer uses static rules for risk scoring.
type DefaultRiskScorer struct {
	toolScores map[string]float64
}

// NewDefaultRiskScorer creates a default risk scorer.
func NewDefaultRiskScorer() *DefaultRiskScorer {
	return &DefaultRiskScorer{
		toolScores: map[string]float64{
			"http_request": 0.4,
			"file_write":   0.7,
			"shell_exec":   0.9,
			"db_write":     0.8,
			"send_email":   0.5,
		},
	}
}

// Score returns a risk score between 0 and 1.
func (r *DefaultRiskScorer) Score(tool string, input map[string]any) float64 {
	if s, ok := r.toolScores[tool]; ok {
		return s
	}
	return 0.3
}
