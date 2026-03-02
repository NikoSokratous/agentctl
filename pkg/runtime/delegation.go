package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DelegationResult contains the result of delegating to another agent.
type DelegationResult struct {
	RunID       string                 `json:"run_id"`
	AgentName   string                 `json:"agent_name"`
	Goal        string                 `json:"goal"`
	State       State                  `json:"state"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
}

// AgentDelegator handles delegation to other agents.
type AgentDelegator struct {
	planner  LLMPlanner
	executor ToolExecutor
	results  map[string]*DelegationResult
	mu       sync.RWMutex
}

// NewAgentDelegator creates a new agent delegator.
func NewAgentDelegator(planner LLMPlanner, executor ToolExecutor) *AgentDelegator {
	return &AgentDelegator{
		planner:  planner,
		executor: executor,
		results:  make(map[string]*DelegationResult),
	}
}

// DelegateToAgent delegates a task to another agent and waits for completion.
func (d *AgentDelegator) DelegateToAgent(
	ctx context.Context,
	agentName string,
	goal string,
	contextData map[string]interface{},
) (*DelegationResult, error) {

	result := &DelegationResult{
		RunID:     fmt.Sprintf("delegated-%s-%d", agentName, time.Now().Unix()),
		AgentName: agentName,
		Goal:      goal,
		StartedAt: time.Now(),
	}

	// Create a child agent engine with default configuration
	engineConfig := EngineConfig{
		AgentName: agentName,
		Goal:      goal,
		MaxSteps:  50,
		Timeout:   5 * time.Minute,
		Autonomy:  AutonomyStandard,
		RunID:     result.RunID,
	}

	childEngine := NewEngine(engineConfig, d.planner, d.executor)

	// Add context data to working memory
	if contextData != nil {
		for k, v := range contextData {
			childEngine.state.WorkingMem[k] = v
		}
	}

	// Execute and wait for completion
	state, err := childEngine.Run(ctx)

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.State = state.Current

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Extract output from final state
	result.Output = map[string]interface{}{
		"agent":      agentName,
		"goal":       goal,
		"status":     string(state.Current),
		"steps":      state.StepCount,
		"final_step": state.LastAction,
	}

	// Store result for async retrieval
	d.mu.Lock()
	d.results[result.RunID] = result
	d.mu.Unlock()

	return result, nil
}

// DelegateToAgentAsync delegates a task without waiting for completion.
func (d *AgentDelegator) DelegateToAgentAsync(
	ctx context.Context,
	agentName string,
	goal string,
	contextData map[string]interface{},
) (runID string, error error) {

	runID = fmt.Sprintf("delegated-async-%s-%d", agentName, time.Now().Unix())

	// Launch agent execution in background goroutine
	go func() {
		// Create background context that won't be canceled with parent
		backgroundCtx := context.Background()

		// Copy timeout from original context if present
		if deadline, ok := ctx.Deadline(); ok {
			var cancel context.CancelFunc
			backgroundCtx, cancel = context.WithDeadline(backgroundCtx, deadline)
			defer cancel()
		}

		// Execute agent asynchronously
		result, err := d.DelegateToAgent(backgroundCtx, agentName, goal, contextData)
		if err != nil {
			// Store error in result
			d.mu.Lock()
			if result != nil {
				d.results[runID] = result
			}
			d.mu.Unlock()
		}
	}()

	return runID, nil
}

// GetDelegationResult retrieves the result of an async delegation.
func (d *AgentDelegator) GetDelegationResult(runID string) (*DelegationResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result, ok := d.results[runID]
	if !ok {
		return nil, fmt.Errorf("delegation result not found for run ID: %s", runID)
	}

	return result, nil
}

// CancelDelegation cancels a delegated agent execution.
func (d *AgentDelegator) CancelDelegation(runID string) error {
	// Note: This is a simplified implementation.
	// In production, you would need to track context.CancelFunc for each delegation
	// and call it here to actually cancel the running agent.

	d.mu.Lock()
	defer d.mu.Unlock()

	result, ok := d.results[runID]
	if !ok {
		return fmt.Errorf("delegation not found for run ID: %s", runID)
	}

	// Mark as interrupted if still running
	if result.State == StatePlanning || result.State == StateExecuting {
		result.State = StateInterrupted
		result.Error = "Delegation canceled by user"
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
	}

	return nil
}
