package policy

import (
	"testing"
)

func TestPolicyEngineCheck(t *testing.T) {
	config := &PolicyConfig{
		Version: "1",
		Rules: []Rule{
			{
				Name: "deny-prod-writes",
				Match: MatchSpec{
					Tool:        "db_write",
					Environment: "production",
				},
				Action:  ActionDeny,
				Message: "writes blocked in prod",
			},
			{
				Name: "high-risk-approval",
				Match: MatchSpec{
					RiskScore: ">= 0.8",
				},
				Action:    ActionRequireApproval,
				Approvers: []string{"admin"},
			},
		},
	}

	engine := NewEngine(config)

	tests := []struct {
		name           string
		ctx            EvalContext
		expectDeny     bool
		expectApproval bool
	}{
		{
			name: "prod write denied",
			ctx: EvalContext{
				Tool:        "db_write",
				Environment: "production",
				RiskScore:   0.3,
			},
			expectDeny:     true,
			expectApproval: false,
		},
		{
			name: "dev write allowed",
			ctx: EvalContext{
				Tool:        "db_write",
				Environment: "development",
				RiskScore:   0.3,
			},
			expectDeny:     false,
			expectApproval: false,
		},
		{
			name: "high risk needs approval",
			ctx: EvalContext{
				Tool:        "any_tool",
				Environment: "",
				RiskScore:   0.9,
				Input:       make(map[string]any),
			},
			expectDeny:     false,
			expectApproval: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Check(tt.ctx)
			if result.Deny != tt.expectDeny {
				t.Errorf("Deny = %v, want %v", result.Deny, tt.expectDeny)
			}
			if result.RequireApproval != tt.expectApproval {
				t.Errorf("RequireApproval = %v, want %v", result.RequireApproval, tt.expectApproval)
			}
		})
	}
}

func TestDefaultRiskScorer(t *testing.T) {
	scorer := NewDefaultRiskScorer()

	tests := []struct {
		tool     string
		minScore float64
	}{
		{"shell_exec", 0.8},
		{"db_write", 0.7},
		{"http_request", 0.3},
		{"unknown_tool", 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			score := scorer.Score(tt.tool, nil)
			if score < tt.minScore {
				t.Errorf("Score(%s) = %v, want >= %v", tt.tool, score, tt.minScore)
			}
		})
	}
}
