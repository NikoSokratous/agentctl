package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Engine evaluates policy rules.
type Engine struct {
	config *PolicyConfig
}

// NewEngine creates a policy engine from config.
func NewEngine(cfg *PolicyConfig) *Engine {
	return &Engine{config: cfg}
}

// Config returns the policy configuration.
func (e *Engine) Config() *PolicyConfig {
	return e.config
}

// LoadEngine loads a policy engine from a YAML file.
func LoadEngine(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy: %w", err)
	}
	return LoadEngineFromBytes(data)
}

// LoadEngineFromBytes loads a policy engine from YAML bytes.
func LoadEngineFromBytes(data []byte) (*Engine, error) {
	var cfg PolicyConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}
	return NewEngine(&cfg), nil
}

// CheckResult is the outcome of a policy check.
type CheckResult struct {
	Allow           bool
	Deny            bool
	Message         string
	RequireApproval bool
	Approvers       []string
}

// Check evaluates rules for a tool execution and returns the decision.
func (e *Engine) Check(ctx EvalContext) CheckResult {
	for _, r := range e.config.Rules {
		if !e.matches(r, ctx) {
			continue
		}
		switch r.Action {
		case ActionDeny:
			return CheckResult{Deny: true, Message: r.Message}
		case ActionRequireApproval:
			return CheckResult{Allow: true, RequireApproval: true, Approvers: r.Approvers}
		default:
			return CheckResult{Allow: true}
		}
	}
	return CheckResult{Allow: true}
}

func (e *Engine) matches(r Rule, ctx EvalContext) bool {
	// Match tool if specified
	if r.Match.Tool != "" && r.Match.Tool != ctx.Tool {
		return false
	}
	// Match environment if specified
	if r.Match.Environment != "" && r.Match.Environment != ctx.Environment {
		return false
	}
	// Match condition if specified
	if r.Match.Condition != "" {
		ok, err := EvalBool(r.Match.Condition, ctx)
		if err != nil || !ok {
			return false
		}
	}
	// Match risk score if specified
	if r.Match.RiskScore != "" {
		// Parse and evaluate expression like ">= 0.8"
		// Convert to "risk_score >= 0.8"
		expr := "risk_score " + r.Match.RiskScore
		ok, err := EvalBool(expr, ctx)
		if err != nil {
			return false
		}
		if !ok {
			return false
		}
	}
	return true
}
