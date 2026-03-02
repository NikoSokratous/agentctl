package runtime

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EngineConfig configures the runtime engine.
type EngineConfig struct {
	AgentName string
	Goal      string
	MaxSteps  int
	Timeout   time.Duration
	Autonomy  AutonomyLevel
	RunID     string // optional; generated if empty
}

// Engine runs the agent execution loop.
type Engine struct {
	config  EngineConfig
	planner LLMPlanner
	exec    ToolExecutor
	state   *AgentState
}

// NewEngine creates a new runtime engine.
func NewEngine(config EngineConfig, planner LLMPlanner, exec ToolExecutor) *Engine {
	runID := config.RunID
	if runID == "" {
		runID = uuid.New().String()
	}
	return &Engine{
		config:  config,
		planner: planner,
		exec:    exec,
		state: &AgentState{
			RunID:      runID,
			AgentName:  config.AgentName,
			Goal:       config.Goal,
			Current:    StateInit,
			StepCount:  0,
			MaxSteps:   config.MaxSteps,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			History:    nil,
			WorkingMem: make(map[string]any),
		},
	}
}

// State returns the current agent state (read-only).
func (e *Engine) State() *AgentState {
	return e.state
}

// Run executes the agent loop until completion, failure, or timeout.
func (e *Engine) Run(ctx context.Context) (*AgentState, error) {
	if e.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.config.Timeout)
		defer cancel()
	}

	e.state.Current = StatePlanning

	for e.state.StepCount < e.state.MaxSteps {
		select {
		case <-ctx.Done():
			e.state.Current = StateInterrupted
			return e.state, ctx.Err()
		default:
		}

		input := StepInput{
			RunID:      e.state.RunID,
			AgentName:  e.state.AgentName,
			Goal:       e.state.Goal,
			StepNum:    e.state.StepCount,
			History:    e.state.History,
			WorkingMem: e.state.WorkingMem,
		}

		result, err := ExecuteStep(ctx, input, e.planner, e.exec)
		if err != nil {
			e.state.Current = StateFailed
			e.state.LastError = err
			return e.state, err
		}

		e.state.StepCount++
		e.state.UpdatedAt = time.Now()
		e.state.LastAction = result.Action
		e.state.LastResult = result.Result
		e.state.Current = result.State

		e.state.History = append(e.state.History, StepRecord{
			StepID:    result.ID,
			Timestamp: result.Timestamp,
			State:     result.State,
			Action:    result.Action,
			Result:    result.Result,
			Reasoning: result.Reasoning,
			Metadata:  result.Metadata,
		})

		if result.State == StateCompleted {
			return e.state, nil
		}
		if result.State == StateFailed {
			return e.state, err
		}
	}

	e.state.Current = StateCompleted
	return e.state, nil
}
