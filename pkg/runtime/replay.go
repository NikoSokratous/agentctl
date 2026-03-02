package runtime

import (
	"context"
	"encoding/json"
)

// ReplayPlanner replays a run by feeding recorded decisions back instead of calling the LLM.
type ReplayPlanner struct {
	history []StepRecord
	index   int
}

// NewReplayPlanner creates a planner that replays recorded steps.
func NewReplayPlanner(history []StepRecord) *ReplayPlanner {
	return &ReplayPlanner{
		history: history,
		index:   0,
	}
}

// Plan implements LLMPlanner by returning the next recorded action.
func (r *ReplayPlanner) Plan(ctx context.Context, input StepInput) (*PlannedAction, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if r.index >= len(r.history) {
		return &PlannedAction{Type: "complete", Reasoning: "replay end"}, nil
	}

	step := r.history[r.index]
	r.index++

	if step.Action == nil {
		return &PlannedAction{Type: "complete", Reasoning: step.Reasoning}, nil
	}

	return &PlannedAction{
		Type:      "tool_call",
		Tool:      step.Action.Tool,
		Version:   step.Action.Version,
		Input:     step.Action.Input,
		Reasoning: step.Reasoning,
	}, nil
}

// ReplayExecutor replays recorded tool outputs instead of executing tools.
type ReplayExecutor struct {
	history []StepRecord
	index   int
}

// NewReplayExecutor creates an executor that returns recorded results.
func NewReplayExecutor(history []StepRecord) *ReplayExecutor {
	return &ReplayExecutor{
		history: history,
		index:   0,
	}
}

// Execute implements ToolExecutor by returning the next recorded result.
func (r *ReplayExecutor) Execute(ctx context.Context, tool, version string, input json.RawMessage) (*ToolResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if r.index >= len(r.history) {
		return nil, nil
	}

	for i := r.index; i < len(r.history); i++ {
		step := r.history[i]
		if step.Result != nil {
			r.index = i + 1
			return step.Result, nil
		}
	}
	return nil, nil
}
