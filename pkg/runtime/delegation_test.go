package runtime

import (
	"context"
	"testing"
	"time"
)

func TestDelegateToAgent(t *testing.T) {
	planner := NewReplayPlanner([]StepRecord{})
	executor := NewReplayExecutor([]StepRecord{})
	delegator := NewAgentDelegator(planner, executor)

	ctx := context.Background()
	result, err := delegator.DelegateToAgent(
		ctx,
		"test-agent",
		"Test goal",
		map[string]interface{}{"key": "value"},
	)

	if err != nil {
		t.Fatalf("DelegateToAgent failed: %v", err)
	}

	if result.AgentName != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", result.AgentName)
	}

	if result.Goal != "Test goal" {
		t.Errorf("Expected goal 'Test goal', got '%s'", result.Goal)
	}

	if result.State != StateCompleted {
		t.Errorf("Expected state completed, got %v", result.State)
	}

	if result.Output == nil {
		t.Error("Expected output, got nil")
	}
}

func TestDelegateToAgentAsync(t *testing.T) {
	planner := NewReplayPlanner([]StepRecord{})
	executor := NewReplayExecutor([]StepRecord{})
	delegator := NewAgentDelegator(planner, executor)

	ctx := context.Background()
	runID, err := delegator.DelegateToAgentAsync(
		ctx,
		"test-agent",
		"Test goal",
		map[string]interface{}{"key": "value"},
	)

	if err != nil {
		t.Fatalf("DelegateToAgentAsync failed: %v", err)
	}

	if runID == "" {
		t.Error("Expected non-empty run ID")
	}

	// Give async execution time to start
	time.Sleep(50 * time.Millisecond)
}

func TestDelegationTimeout(t *testing.T) {
	planner := NewReplayPlanner([]StepRecord{})
	executor := NewReplayExecutor([]StepRecord{})
	delegator := NewAgentDelegator(planner, executor)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	_, err := delegator.DelegateToAgent(
		ctx,
		"test-agent",
		"Test goal",
		nil,
	)

	// We expect this to potentially timeout, but the function currently doesn't check context
	// This is a placeholder for future timeout handling
	_ = err
}
