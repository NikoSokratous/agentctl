package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Breakpoint represents a debugging breakpoint
type Breakpoint struct {
	NodeID    string
	Condition string // CEL expression
	Enabled   bool
}

// DebugSession manages a workflow debugging session
type DebugSession struct {
	WorkflowID    string
	Breakpoints   map[string]*Breakpoint
	CurrentStep   string
	Variables     map[string]interface{}
	StepHistory   []string
	Paused        bool
	mu            sync.RWMutex
	pauseChan     chan struct{}
	resumeChan    chan struct{}
	stepChan      chan struct{}
	condEvaluator ConditionEvaluator
}

// NewDebugSession creates a new debug session
func NewDebugSession(workflowID string) *DebugSession {
	condEval, _ := NewCELEvaluator()
	return &DebugSession{
		WorkflowID:    workflowID,
		Breakpoints:   make(map[string]*Breakpoint),
		Variables:     make(map[string]interface{}),
		StepHistory:   []string{},
		pauseChan:     make(chan struct{}, 1),
		resumeChan:    make(chan struct{}, 1),
		stepChan:      make(chan struct{}, 1),
		condEvaluator: condEval,
	}
}

// AddBreakpoint adds a breakpoint to the session
func (ds *DebugSession) AddBreakpoint(nodeID string, condition string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.Breakpoints[nodeID] = &Breakpoint{
		NodeID:    nodeID,
		Condition: condition,
		Enabled:   true,
	}
}

// RemoveBreakpoint removes a breakpoint
func (ds *DebugSession) RemoveBreakpoint(nodeID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.Breakpoints, nodeID)
}

// EnableBreakpoint enables/disables a breakpoint
func (ds *DebugSession) EnableBreakpoint(nodeID string, enabled bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if bp, ok := ds.Breakpoints[nodeID]; ok {
		bp.Enabled = enabled
	}
}

// ShouldPause checks if execution should pause at a node
func (ds *DebugSession) ShouldPause(nodeID string, context map[string]interface{}) (bool, error) {
	ds.mu.RLock()
	bp, exists := ds.Breakpoints[nodeID]
	ds.mu.RUnlock()

	if !exists || !bp.Enabled {
		return false, nil
	}

	// If no condition, always pause
	if bp.Condition == "" {
		return true, nil
	}

	// Evaluate condition
	result, err := ds.condEvaluator.Evaluate(bp.Condition, context)
	if err != nil {
		return false, fmt.Errorf("evaluate breakpoint condition: %w", err)
	}

	return result, nil
}

// Pause pauses execution
func (ds *DebugSession) Pause() {
	ds.mu.Lock()
	ds.Paused = true
	ds.mu.Unlock()

	select {
	case ds.pauseChan <- struct{}{}:
	default:
	}
}

// Resume resumes execution
func (ds *DebugSession) Resume() {
	ds.mu.Lock()
	ds.Paused = false
	ds.mu.Unlock()

	select {
	case ds.resumeChan <- struct{}{}:
	default:
	}
}

// StepOver executes one step and pauses
func (ds *DebugSession) StepOver() {
	select {
	case ds.stepChan <- struct{}{}:
	default:
	}
}

// WaitIfPaused blocks execution if paused
func (ds *DebugSession) WaitIfPaused(ctx context.Context, nodeID string) error {
	ds.mu.Lock()
	ds.CurrentStep = nodeID
	ds.StepHistory = append(ds.StepHistory, nodeID)
	paused := ds.Paused
	ds.mu.Unlock()

	if !paused {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ds.resumeChan:
		return nil
	case <-ds.stepChan:
		// Step over - pause again after this step
		ds.Pause()
		return nil
	}
}

// UpdateVariables updates the variable state
func (ds *DebugSession) UpdateVariables(vars map[string]interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for k, v := range vars {
		ds.Variables[k] = v
	}
}

// GetState returns current debug state
func (ds *DebugSession) GetState() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return map[string]interface{}{
		"workflow_id":  ds.WorkflowID,
		"current_step": ds.CurrentStep,
		"paused":       ds.Paused,
		"variables":    ds.Variables,
		"breakpoints":  ds.Breakpoints,
		"history":      ds.StepHistory,
	}
}

// DebugExecutor wraps Executor with debugging capabilities
type DebugExecutor struct {
	*Executor
	session *DebugSession
}

// NewDebugExecutor creates a new debug-enabled executor
func NewDebugExecutor(stateStore *StateStore, session *DebugSession) *DebugExecutor {
	return &DebugExecutor{
		Executor: NewExecutor(stateStore, session.condEvaluator),
		session:  session,
	}
}

// ExecuteDAGWithDebug executes a DAG with debugging support
func (de *DebugExecutor) ExecuteDAGWithDebug(ctx context.Context, dag *DAG, inputs map[string]interface{}) (map[string]interface{}, error) {
	levels, err := dag.GetExecutionLevels()
	if err != nil {
		return nil, fmt.Errorf("get execution levels: %w", err)
	}

	outputs := make(map[string]interface{})

	// Initialize context with inputs
	executionContext := map[string]interface{}{
		"inputs":  inputs,
		"outputs": outputs,
	}

	for levelIdx, level := range levels {
		// Check for pause at level start
		if err := de.session.WaitIfPaused(ctx, fmt.Sprintf("level-%d", levelIdx)); err != nil {
			return nil, err
		}

		// Execute nodes in this level
		for _, nodeID := range level {
			// Update debug context
			executionContext["step"] = map[string]interface{}{
				"name":   nodeID,
				"status": "running",
			}
			de.session.UpdateVariables(executionContext)

			// Check breakpoint
			shouldPause, err := de.session.ShouldPause(nodeID, executionContext)
			if err != nil {
				return nil, fmt.Errorf("check breakpoint: %w", err)
			}

			if shouldPause {
				de.session.Pause()
			}

			// Wait if paused
			if err := de.session.WaitIfPaused(ctx, nodeID); err != nil {
				return nil, err
			}

			// Execute node (simplified - in real implementation, call actual agent)
			result := map[string]interface{}{
				"node_id":   nodeID,
				"executed":  true,
				"timestamp": time.Now(),
			}

			outputs[nodeID] = result
			executionContext["outputs"] = outputs

			// Update debug state
			executionContext["step"] = map[string]interface{}{
				"name":   nodeID,
				"status": "completed",
				"result": result,
			}
			de.session.UpdateVariables(executionContext)
		}
	}

	return outputs, nil
}
