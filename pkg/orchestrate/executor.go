package orchestrate

import (
	"context"
	"time"
)

// StepExecutor executes a single workflow step. When nil, SimulatedExecutor is used.
type StepExecutor interface {
	ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error)
}

// SimulatedExecutor simulates step execution with a short delay (no real agent run).
type SimulatedExecutor struct{}

// ExecuteStep simulates execution and returns a placeholder result.
func (SimulatedExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	time.Sleep(100 * time.Millisecond)
	return &StepResult{
		Name:        agentName,
		Agent:       agentName,
		Status:      "completed",
		RunID:       "run-" + agentName + "-simulated",
		Output:      map[string]interface{}{"agent": agentName, "goal": goal},
		Duration:    100 * time.Millisecond,
		StartedAt:   time.Now().Add(-100 * time.Millisecond),
		CompletedAt: time.Now(),
	}, nil
}
