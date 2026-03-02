package observe

import "time"

// EventType identifies the kind of event.
type EventType string

const (
	EventInit        EventType = "init"
	EventPlan        EventType = "plan"
	EventToolCall    EventType = "tool_call"
	EventToolResult  EventType = "tool_result"
	EventApproval    EventType = "approval"
	EventPolicyCheck EventType = "policy_check"
	EventError       EventType = "error"
	EventCompleted   EventType = "completed"
	EventInterrupted EventType = "interrupted"
	EventReasoning   EventType = "reasoning"
)

// ModelMeta captures model metadata for audit.
type ModelMeta struct {
	Provider string `json:"provider,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
}

// Event is a structured execution event.
type Event struct {
	RunID     string         `json:"run_id"`
	StepID    string         `json:"step_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Type      EventType      `json:"type"`
	Agent     string         `json:"agent"`
	Data      map[string]any `json:"data,omitempty"`
	Model     ModelMeta      `json:"model,omitempty"`
}
