package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SimulationMode defines how policies are simulated.
type SimulationMode string

const (
	SimulationModeAudit      SimulationMode = "audit"      // Record what would happen
	SimulationModeSimulation SimulationMode = "simulation" // Test without execution
	SimulationModeShadow     SimulationMode = "shadow"     // Run alongside production
)

// Simulator tests policies without executing real actions.
type Simulator struct {
	store    *VersionStore
	executor *Executor
}

// NewSimulator creates a new policy simulator.
func NewSimulator(store *VersionStore, executor *Executor) *Simulator {
	return &Simulator{
		store:    store,
		executor: executor,
	}
}

// SimulationRequest represents a policy simulation request.
type SimulationRequest struct {
	PolicyName    string                 `json:"policy_name"`
	PolicyVersion string                 `json:"policy_version"`
	RunID         string                 `json:"run_id,omitempty"`
	Mode          SimulationMode         `json:"mode"`
	Actions       []ActionToSimulate     `json:"actions"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// ActionToSimulate represents an action to test.
type ActionToSimulate struct {
	Sequence int                    `json:"sequence"`
	Tool     string                 `json:"tool"`
	Input    map[string]interface{} `json:"input"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// SimulationResult represents the result of a policy simulation.
type SimulationResult struct {
	ID            string                   `json:"id"`
	PolicyName    string                   `json:"policy_name"`
	PolicyVersion string                   `json:"policy_version"`
	Mode          SimulationMode           `json:"mode"`
	StartedAt     time.Time                `json:"started_at"`
	CompletedAt   time.Time                `json:"completed_at"`
	TotalActions  int                      `json:"total_actions"`
	Allowed       int                      `json:"allowed"`
	Denied        int                      `json:"denied"`
	Alerts        int                      `json:"alerts"`
	Details       []ActionSimulationResult `json:"details"`
	Summary       SimulationSummary        `json:"summary"`
}

// ActionSimulationResult represents the result of simulating one action.
type ActionSimulationResult struct {
	Sequence   int                    `json:"sequence"`
	Action     string                 `json:"action"`
	Tool       string                 `json:"tool"`
	Input      map[string]interface{} `json:"input"`
	Allowed    bool                   `json:"allowed"`
	DenyReason string                 `json:"deny_reason,omitempty"`
	RiskScore  float64                `json:"risk_score"`
	WouldAlert bool                   `json:"would_alert"`
	Duration   time.Duration          `json:"duration"`
}

// SimulationSummary summarizes simulation results.
type SimulationSummary struct {
	AllowRate      float64           `json:"allow_rate"`
	DenyRate       float64           `json:"deny_rate"`
	AlertRate      float64           `json:"alert_rate"`
	AvgRiskScore   float64           `json:"avg_risk_score"`
	TopDenyReasons []DenyReasonCount `json:"top_deny_reasons"`
}

// DenyReasonCount counts occurrences of deny reasons.
type DenyReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

// Simulate runs a policy simulation.
func (s *Simulator) Simulate(ctx context.Context, req SimulationRequest) (*SimulationResult, error) {
	// Load policy version
	policy, err := s.store.GetVersion(ctx, req.PolicyName, req.PolicyVersion)
	if err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}

	result := &SimulationResult{
		ID:            uuid.New().String(),
		PolicyName:    req.PolicyName,
		PolicyVersion: req.PolicyVersion,
		Mode:          req.Mode,
		StartedAt:     time.Now(),
		TotalActions:  len(req.Actions),
		Details:       make([]ActionSimulationResult, 0, len(req.Actions)),
	}

	// Parse policy
	policyDoc, err := ParsePolicy(policy.Content)
	if err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}

	// Simulate each action
	denyReasons := make(map[string]int)
	totalRiskScore := 0.0

	for _, action := range req.Actions {
		actionStart := time.Now()

		// Evaluate action against policy
		evalResult := s.evaluateAction(policyDoc, action, req.Context)

		actionResult := ActionSimulationResult{
			Sequence:   action.Sequence,
			Action:     fmt.Sprintf("%s(%v)", action.Tool, action.Input),
			Tool:       action.Tool,
			Input:      action.Input,
			Allowed:    evalResult.Allowed,
			DenyReason: evalResult.DenyReason,
			RiskScore:  evalResult.RiskScore,
			WouldAlert: evalResult.WouldAlert,
			Duration:   time.Since(actionStart),
		}

		result.Details = append(result.Details, actionResult)

		// Update counters
		if evalResult.Allowed {
			result.Allowed++
		} else {
			result.Denied++
			if evalResult.DenyReason != "" {
				denyReasons[evalResult.DenyReason]++
			}
		}
		if evalResult.WouldAlert {
			result.Alerts++
		}

		totalRiskScore += evalResult.RiskScore
	}

	result.CompletedAt = time.Now()

	// Calculate summary
	result.Summary = s.calculateSummary(result, denyReasons)

	return result, nil
}

// PolicyEvalResult represents the result of evaluating an action.
type PolicyEvalResult struct {
	Allowed    bool
	DenyReason string
	RiskScore  float64
	WouldAlert bool
}

// evaluateAction evaluates a single action against the policy.
func (s *Simulator) evaluateAction(policy *PolicyDocument, action ActionToSimulate, context map[string]interface{}) PolicyEvalResult {
	// Simple evaluation logic - will be enhanced
	result := PolicyEvalResult{
		Allowed:   true,
		RiskScore: 0.0,
	}

	// Check if tool is allowed
	for _, rule := range policy.Rules {
		if matchesRule(rule, action.Tool, action.Input, context) {
			if rule.Effect == "deny" {
				result.Allowed = false
				result.DenyReason = rule.Reason
				result.RiskScore = rule.RiskScore
			}
			if rule.Alert {
				result.WouldAlert = true
			}
			break
		}
	}

	return result
}

// matchesRule checks if an action matches a policy rule.
func matchesRule(rule PolicyRule, tool string, input, context map[string]interface{}) bool {
	// Simple matching - will be enhanced with expression evaluation
	if rule.Tool != "" && rule.Tool != tool && rule.Tool != "*" {
		return false
	}
	return true
}

// calculateSummary calculates simulation summary statistics.
func (s *Simulator) calculateSummary(result *SimulationResult, denyReasons map[string]int) SimulationSummary {
	summary := SimulationSummary{}

	if result.TotalActions > 0 {
		summary.AllowRate = float64(result.Allowed) / float64(result.TotalActions)
		summary.DenyRate = float64(result.Denied) / float64(result.TotalActions)
		summary.AlertRate = float64(result.Alerts) / float64(result.TotalActions)
	}

	// Calculate average risk score
	totalRisk := 0.0
	for _, detail := range result.Details {
		totalRisk += detail.RiskScore
	}
	if len(result.Details) > 0 {
		summary.AvgRiskScore = totalRisk / float64(len(result.Details))
	}

	// Get top deny reasons
	for reason, count := range denyReasons {
		summary.TopDenyReasons = append(summary.TopDenyReasons, DenyReasonCount{
			Reason: reason,
			Count:  count,
		})
	}

	return summary
}

// PolicyDocument represents a parsed policy.
type PolicyDocument struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   PolicyMeta   `yaml:"metadata"`
	Spec       PolicySpec   `yaml:"spec"`
	Rules      []PolicyRule `yaml:"rules"`
}

// PolicyMeta holds policy metadata.
type PolicyMeta struct {
	Name      string            `yaml:"name"`
	Version   string            `yaml:"version"`
	Changelog []ChangelogEntry  `yaml:"changelog,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// ChangelogEntry represents a single changelog entry.
type ChangelogEntry struct {
	Version string    `yaml:"version"`
	Date    time.Time `yaml:"date"`
	Author  string    `yaml:"author"`
	Changes string    `yaml:"changes"`
}

// PolicySpec defines policy specifications.
type PolicySpec struct {
	DefaultEffect string `yaml:"defaultEffect"` // "allow" or "deny"
	Mode          string `yaml:"mode"`          // "enforcing", "permissive", "audit"
}

// PolicyRule represents a single policy rule.
type PolicyRule struct {
	ID              string  `yaml:"id"`
	Tool            string  `yaml:"tool"`
	Effect          string  `yaml:"effect"` // "allow" or "deny"
	Condition       string  `yaml:"condition,omitempty"`
	Reason          string  `yaml:"reason,omitempty"`
	Alert           bool    `yaml:"alert,omitempty"`
	RiskScore       float64 `yaml:"riskScore,omitempty"`
	RequireApproval bool    `yaml:"requireApproval,omitempty"`
}

// ParsePolicy parses a policy document.
func ParsePolicy(content []byte) (*PolicyDocument, error) {
	var policy PolicyDocument
	// Simple JSON unmarshal for now - can add YAML support
	if err := json.Unmarshal(content, &policy); err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}
	return &policy, nil
}
