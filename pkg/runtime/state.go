package runtime

import "time"

// State represents the current phase of agent execution.
type State string

const (
	StateInit            State = "init"
	StatePlanning        State = "planning"
	StateToolSelect      State = "tool_select"
	StatePolicyCheck     State = "policy_check"
	StateWaitingApproval State = "waiting_approval"
	StateExecuting       State = "executing"
	StateObserving       State = "observing"
	StateCompleted       State = "completed"
	StateFailed          State = "failed"
	StateDenied          State = "denied"
	StateInterrupted     State = "interrupted"
)

// AutonomyLevel controls how much human approval is required.
// Level 0: every tool call requires approval
// Level 1: risky actions require approval (risk_score >= threshold)
// Level 2: only policy-flagged actions require approval
// Level 3: no approval gates, all actions logged
// Level 4: no policy enforcement (dev/test only)
type AutonomyLevel int

const (
	AutonomyManual       AutonomyLevel = 0
	AutonomyCautious     AutonomyLevel = 1
	AutonomyStandard     AutonomyLevel = 2
	AutonomyAutonomous   AutonomyLevel = 3
	AutonomyUnrestricted AutonomyLevel = 4
)

// AgentState holds the full runtime state for an execution.
type AgentState struct {
	RunID      string
	AgentName  string
	Goal       string
	Current    State
	StepCount  int
	MaxSteps   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	History    []StepRecord
	WorkingMem map[string]any
	LastAction *ToolCall
	LastResult *ToolResult
	LastError  error
}

// StepRecord captures one execution step for replay and audit.
type StepRecord struct {
	StepID    string
	Timestamp time.Time
	State     State
	Action    *ToolCall
	Result    *ToolResult
	Reasoning string
	Metadata  map[string]any
}

// ToolCall represents an agent's chosen action.
type ToolCall struct {
	ID      string
	Tool    string
	Version string
	Input   map[string]any
}

// ToolResult is the outcome of executing a tool.
type ToolResult struct {
	ToolID   string
	Output   map[string]any
	Error    string
	Duration time.Duration
}
