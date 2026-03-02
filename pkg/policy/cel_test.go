package policy

import (
	"testing"
)

func TestCELRiskScore(t *testing.T) {
	ctx := EvalContext{
		Tool:        "any_tool",
		Environment: "",
		RiskScore:   0.9,
		Input:       make(map[string]any),
	}

	// Test the actual CEL expression
	expr := "risk_score >= 0.8"
	result, err := EvalBool(expr, ctx)
	if err != nil {
		t.Fatalf("CEL eval error: %v", err)
	}
	t.Logf("CEL result for '%s' with risk=0.9: %v", expr, result)

	if !result {
		t.Errorf("Expected true for risk_score 0.9 >= 0.8, got false")
	}
}

func TestPolicyEngineCheckDebug(t *testing.T) {
	config := &PolicyConfig{
		Version: "1",
		Rules: []Rule{
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
	ctx := EvalContext{
		Tool:        "any_tool",
		Environment: "",
		RiskScore:   0.9,
		Input:       make(map[string]any),
	}

	t.Logf("Testing with context: %+v", ctx)
	t.Logf("Rule match spec: %+v", config.Rules[0].Match)

	// Manually test if rule matches
	matches := engine.matches(config.Rules[0], ctx)
	t.Logf("Rule matches: %v", matches)

	result := engine.Check(ctx)
	t.Logf("Check result: %+v", result)

	if !result.RequireApproval {
		t.Errorf("Expected RequireApproval=true, got false")
	}
}
