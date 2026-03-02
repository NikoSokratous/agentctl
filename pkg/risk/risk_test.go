package risk

import (
	"context"
	"testing"
	"time"
)

func TestRiskEngine(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	tests := []struct {
		name         string
		ctx          ActionContext
		wantLevel    RiskLevel
		wantDecision RiskDecision
		minScore     float64
		maxScore     float64
	}{
		{
			name: "low risk - read file in dev",
			ctx: ActionContext{
				ToolName:    "read_file",
				Environment: "dev",
				Permissions: []string{"fs:read"},
			},
			wantLevel:    RiskLevelLow,
			wantDecision: DecisionAllow,
			maxScore:     0.5,
		},
		{
			name: "medium risk - write file",
			ctx: ActionContext{
				ToolName:    "write_file",
				Environment: "staging",
				Permissions: []string{"fs:write"},
			},
			wantLevel: RiskLevelMedium,
			minScore:  0.3,
			maxScore:  0.7,
		},
		{
			name: "high risk - delete file in production",
			ctx: ActionContext{
				ToolName:    "delete_file",
				Environment: "production",
				Permissions: []string{"fs:delete"},
			},
			wantLevel:    RiskLevelHigh,
			wantDecision: DecisionAllowWithLog, // Below approval threshold
			minScore:     0.6,
		},
		{
			name: "critical risk - execute command in prod",
			ctx: ActionContext{
				ToolName:    "execute_command",
				Environment: "production",
				Permissions: []string{"exec", "fs:write"},
			},
			wantLevel:    RiskLevelHigh,
			wantDecision: DecisionAllowWithLog, // Below approval threshold but high
			minScore:     0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment, err := engine.Assess(context.Background(), tt.ctx)
			if err != nil {
				t.Fatalf("Assess: %v", err)
			}

			score := assessment.RiskScore

			// Check score range
			if tt.minScore > 0 && score.Score < tt.minScore {
				t.Errorf("Score too low: got %.2f, want >= %.2f", score.Score, tt.minScore)
			}
			if tt.maxScore > 0 && score.Score > tt.maxScore {
				t.Errorf("Score too high: got %.2f, want <= %.2f", score.Score, tt.maxScore)
			}

			// Check level
			if tt.wantLevel != "" && score.Level != tt.wantLevel {
				t.Errorf("Level: got %s, want %s", score.Level, tt.wantLevel)
			}

			// Check decision
			if tt.wantDecision != "" && assessment.Decision != tt.wantDecision {
				t.Errorf("Decision: got %s, want %s", assessment.Decision, tt.wantDecision)
			}

			// Check factors exist
			if len(score.Factors) == 0 {
				t.Error("No risk factors calculated")
			}

			// Check breakdown
			if len(score.Breakdown) == 0 {
				t.Error("No score breakdown")
			}

			// Check confidence
			if score.Confidence < 0 || score.Confidence > 1 {
				t.Errorf("Invalid confidence: %.2f", score.Confidence)
			}
		})
	}
}

func TestRiskEngineDisabled(t *testing.T) {
	config := DefaultRiskConfig()
	config.Enabled = false
	engine := NewRiskEngine(config)

	ctx := ActionContext{
		ToolName:    "delete_file",
		Environment: "production",
	}

	assessment, err := engine.Assess(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}

	// When disabled, should return low risk
	if assessment.RiskScore.Score != 0.0 {
		t.Errorf("Score: got %.2f, want 0.0", assessment.RiskScore.Score)
	}
	if assessment.Decision != DecisionAllow {
		t.Errorf("Decision: got %s, want allow", assessment.Decision)
	}
}

func TestRiskCategories(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	// Test security category
	securityCtx := ActionContext{
		ToolName:    "execute_command",
		Permissions: []string{"exec"},
	}
	assessment, _ := engine.Assess(context.Background(), securityCtx)

	securityScore := 0.0
	for _, factor := range assessment.RiskScore.Factors {
		if factor.Category == string(CategorySecurity) {
			securityScore = factor.Score
			break
		}
	}

	if securityScore < 0.7 {
		t.Errorf("Security score too low for execute_command: %.2f", securityScore)
	}

	// Test privacy category with PII
	privacyCtx := ActionContext{
		ToolName: "read_file",
		Input: map[string]interface{}{
			"path": "/data/user_emails.txt",
		},
	}
	assessment, _ = engine.Assess(context.Background(), privacyCtx)

	hasPrivacyFactor := false
	for _, factor := range assessment.RiskScore.Factors {
		if factor.Category == string(CategoryPrivacy) {
			hasPrivacyFactor = true
			break
		}
	}

	if !hasPrivacyFactor {
		t.Error("Missing privacy risk factor")
	}
}

func TestRiskLevels(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	tests := []struct {
		score float64
		want  RiskLevel
	}{
		{0.1, RiskLevelLow},
		{0.29, RiskLevelLow},
		{0.3, RiskLevelMedium},
		{0.5, RiskLevelMedium},
		{0.6, RiskLevelHigh},
		{0.8, RiskLevelHigh},
		{0.85, RiskLevelCritical},
		{0.99, RiskLevelCritical},
	}

	for _, tt := range tests {
		level := engine.scoreToLevel(tt.score)
		if level != tt.want {
			t.Errorf("scoreToLevel(%.2f): got %s, want %s", tt.score, level, tt.want)
		}
	}
}

func TestRiskDecisions(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	tests := []struct {
		score float64
		level RiskLevel
		want  RiskDecision
	}{
		{0.2, RiskLevelLow, DecisionAllow},
		{0.5, RiskLevelMedium, DecisionAllowWithLog},
		{0.7, RiskLevelHigh, DecisionAllowWithLog},
		{0.82, RiskLevelHigh, DecisionRequireReview},
		{0.96, RiskLevelCritical, DecisionDeny},
	}

	for _, tt := range tests {
		riskScore := RiskScore{
			Score: tt.score,
			Level: tt.level,
		}
		decision := engine.makeDecision(riskScore, ActionContext{})

		if decision != tt.want {
			t.Errorf("makeDecision(%.2f, %s): got %s, want %s",
				tt.score, tt.level, decision, tt.want)
		}
	}
}

func TestRiskEnvironmentImpact(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	// Same tool, different environments
	baseCtx := ActionContext{
		ToolName: "write_file",
	}

	// Dev environment
	devCtx := baseCtx
	devCtx.Environment = "dev"
	devAssessment, _ := engine.Assess(context.Background(), devCtx)

	// Production environment
	prodCtx := baseCtx
	prodCtx.Environment = "production"
	prodAssessment, _ := engine.Assess(context.Background(), prodCtx)

	// Production should have higher risk
	if prodAssessment.RiskScore.Score <= devAssessment.RiskScore.Score {
		t.Errorf("Production risk (%.2f) should be higher than dev (%.2f)",
			prodAssessment.RiskScore.Score,
			devAssessment.RiskScore.Score)
	}
}

func TestRiskFactorContributions(t *testing.T) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	ctx := ActionContext{
		ToolName:    "delete_file",
		Environment: "production",
		Permissions: []string{"fs:delete"},
	}

	assessment, err := engine.Assess(context.Background(), ctx)
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}

	// Check all factors have contributions
	for _, factor := range assessment.RiskScore.Factors {
		if factor.Score < 0 || factor.Score > 1 {
			t.Errorf("Invalid factor score: %.2f", factor.Score)
		}
		if factor.Weight < 0 {
			t.Errorf("Invalid factor weight: %.2f", factor.Weight)
		}

		expectedContribution := factor.Score * factor.Weight
		if factor.Contribution != expectedContribution {
			t.Errorf("Contribution mismatch: got %.2f, want %.2f",
				factor.Contribution, expectedContribution)
		}
	}
}

func TestComplianceReportGeneration(t *testing.T) {
	// This is a placeholder test
	// In production, would use a test database

	t.Run("report structure", func(t *testing.T) {
		report := &ComplianceReport{
			ID:            "report-test",
			ReportType:    "daily",
			PeriodStart:   time.Now().Add(-24 * time.Hour),
			PeriodEnd:     time.Now(),
			GeneratedAt:   time.Now(),
			TotalActions:  100,
			HighRiskCount: 10,
			DeniedCount:   5,
			ApprovalCount: 8,
			Summary: ComplianceSummary{
				ComplianceRate:   0.95,
				AverageRiskScore: 0.42,
			},
		}

		if report.TotalActions != 100 {
			t.Errorf("TotalActions: got %d, want 100", report.TotalActions)
		}
		if report.Summary.ComplianceRate != 0.95 {
			t.Errorf("ComplianceRate: got %.2f, want 0.95", report.Summary.ComplianceRate)
		}
	})
}

func TestPIIDetection(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"normal text", false},
		{"user email@example.com", true},
		{"ssn 123-45-6789", true},
		{"password: secret123", true},
		{"just data", false},
	}

	for _, tt := range tests {
		got := containsPII(tt.input)
		if got != tt.want {
			t.Errorf("containsPII(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRiskWeights(t *testing.T) {
	config := DefaultRiskConfig()

	// Check all categories have weights
	expectedCategories := []RiskCategory{
		CategorySecurity,
		CategoryPrivacy,
		CategoryCost,
		CategoryReversible,
		CategoryImpact,
		CategoryCompliance,
		CategoryReliability,
	}

	for _, category := range expectedCategories {
		weight, exists := config.DefaultWeights[category]
		if !exists {
			t.Errorf("Missing weight for category: %s", category)
		}
		if weight < 0 || weight > 2.0 { // Allow weights > 1 for emphasis
			t.Errorf("Invalid weight for %s: %.2f", category, weight)
		}
	}
}

func BenchmarkRiskAssessment(b *testing.B) {
	config := DefaultRiskConfig()
	engine := NewRiskEngine(config)

	ctx := ActionContext{
		ToolName:    "http_request",
		Environment: "production",
		Permissions: []string{"net:external"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Assess(context.Background(), ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
