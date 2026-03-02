package runtime

import (
	"testing"
	"time"
)

func TestAutonomyLevelRequiresApproval(t *testing.T) {
	tests := []struct {
		name             string
		level            AutonomyLevel
		policyRequires   bool
		riskScore        float64
		expectedApproval bool
	}{
		{"Manual always requires approval", AutonomyManual, false, 0.1, true},
		{"Cautious requires on high risk", AutonomyCautious, false, 0.6, true},
		{"Cautious no approval on low risk", AutonomyCautious, false, 0.3, false},
		{"Cautious requires on policy", AutonomyCautious, true, 0.3, true},
		{"Standard only policy", AutonomyStandard, true, 0.9, true},
		{"Standard no policy", AutonomyStandard, false, 0.9, false},
		{"Autonomous never requires", AutonomyAutonomous, true, 0.9, false},
		{"Unrestricted never requires", AutonomyUnrestricted, true, 0.9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresApproval(tt.level, tt.policyRequires, tt.riskScore)
			if result != tt.expectedApproval {
				t.Errorf("RequiresApproval(%v, %v, %v) = %v, want %v",
					tt.level, tt.policyRequires, tt.riskScore, result, tt.expectedApproval)
			}
		})
	}
}

func TestEnforcesPolicy(t *testing.T) {
	tests := []struct {
		level    AutonomyLevel
		expected bool
	}{
		{AutonomyManual, true},
		{AutonomyCautious, true},
		{AutonomyStandard, true},
		{AutonomyAutonomous, true},
		{AutonomyUnrestricted, false},
	}

	for _, tt := range tests {
		result := EnforcesPolicy(tt.level)
		if result != tt.expected {
			t.Errorf("EnforcesPolicy(%v) = %v, want %v", tt.level, result, tt.expected)
		}
	}
}

func TestAgentStateInitialization(t *testing.T) {
	state := &AgentState{
		RunID:      "test-run",
		AgentName:  "test-agent",
		Goal:       "test goal",
		Current:    StateInit,
		MaxSteps:   10,
		WorkingMem: make(map[string]any),
	}

	if state.RunID != "test-run" {
		t.Errorf("RunID = %v, want test-run", state.RunID)
	}
	if state.Current != StateInit {
		t.Errorf("Current = %v, want %v", state.Current, StateInit)
	}
	if state.StepCount != 0 {
		t.Errorf("StepCount = %v, want 0", state.StepCount)
	}
}

func TestStepRecordCreation(t *testing.T) {
	now := time.Now()
	record := StepRecord{
		StepID:    "step-1",
		Timestamp: now,
		State:     StatePlanning,
		Action: &ToolCall{
			ID:      "call-1",
			Tool:    "test-tool",
			Version: "1",
			Input:   map[string]any{"key": "value"},
		},
		Reasoning: "test reasoning",
	}

	if record.StepID != "step-1" {
		t.Errorf("StepID = %v, want step-1", record.StepID)
	}
	if record.Action.Tool != "test-tool" {
		t.Errorf("Action.Tool = %v, want test-tool", record.Action.Tool)
	}
}
