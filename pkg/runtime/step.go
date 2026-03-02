package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StepInput provides context for a single execution step.
type StepInput struct {
	RunID      string
	AgentName  string
	Goal       string
	StepNum    int
	History    []StepRecord
	WorkingMem map[string]any
}

// StepResult is the outcome of one execution step.
type StepResult struct {
	ID        string
	Timestamp time.Time
	State     State
	Action    *ToolCall
	Result    *ToolResult
	Reasoning string
	Metadata  map[string]any
}

// ToolExecutor is called to run a tool. Implemented by pkg/tool.
type ToolExecutor interface {
	Execute(ctx context.Context, tool string, version string, input json.RawMessage) (*ToolResult, error)
}

// LLMPlanner returns the next action (tool call or completion) based on context.
// Implemented by pkg/llm integration.
type LLMPlanner interface {
	Plan(ctx context.Context, input StepInput) (*PlannedAction, error)
}

// PlannedAction is the LLM's proposed next step.
type PlannedAction struct {
	Type      string         // "tool_call" | "complete" | "retry"
	Tool      string         // tool name if Type == "tool_call"
	Version   string         // tool version
	Input     map[string]any // tool input if Type == "tool_call"
	Reasoning string
}

// ExecuteStep runs one iteration of the agent loop.
func ExecuteStep(ctx context.Context, input StepInput, planner LLMPlanner, executor ToolExecutor) (*StepResult, error) {
	planned, err := planner.Plan(ctx, input)
	if err != nil {
		return &StepResult{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			State:     StateFailed,
			Metadata:  map[string]any{"error": err.Error()},
		}, err
	}

	result := &StepResult{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Reasoning: planned.Reasoning,
		Metadata:  make(map[string]any),
	}

	switch planned.Type {
	case "complete":
		result.State = StateCompleted
		result.Metadata["goal_reached"] = true
		return result, nil

	case "tool_call":
		inputBytes, _ := json.Marshal(planned.Input)
		output, execErr := executor.Execute(ctx, planned.Tool, planned.Version, inputBytes)
		result.Action = &ToolCall{
			ID:      uuid.New().String(),
			Tool:    planned.Tool,
			Version: planned.Version,
			Input:   planned.Input,
		}
		result.Result = output
		result.State = StateObserving
		if execErr != nil {
			result.State = StateFailed
			result.Metadata["error"] = execErr.Error()
			return result, execErr
		}
		return result, nil

	default:
		result.State = StateFailed
		result.Metadata["error"] = "unknown action type: " + planned.Type
		return result, nil
	}
}
