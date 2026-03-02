package observe

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger writes structured events to one or more outputs.
type Logger struct {
	mu      sync.Mutex
	writers []io.Writer
	agent   string
}

// NewLogger creates a logger that writes to the given writers.
func NewLogger(agent string, writers ...io.Writer) *Logger {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Logger{
		writers: writers,
		agent:   agent,
	}
}

// Log writes an event to all configured writers.
func (l *Logger) Log(runID, stepID string, typ EventType, data map[string]any, model ModelMeta) {
	evt := Event{
		RunID:     runID,
		StepID:    stepID,
		Timestamp: time.Now().UTC(),
		Type:      typ,
		Agent:     l.agent,
		Data:      data,
		Model:     model,
	}
	bytes, err := json.Marshal(evt)
	if err != nil {
		return
	}
	line := append(bytes, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, w := range l.writers {
		_, _ = w.Write(line)
	}
}

// AddWriter appends a writer for output.
func (l *Logger) AddWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, w)
}

// LogInit logs run initialization.
func (l *Logger) LogInit(runID string, goal string) {
	l.Log(runID, "", EventInit, map[string]any{"goal": goal}, ModelMeta{})
}

// LogPlan logs a planning step.
func (l *Logger) LogPlan(runID, stepID string, reasoning string, model ModelMeta) {
	l.Log(runID, stepID, EventPlan, map[string]any{"reasoning": reasoning}, model)
}

// LogToolCall logs a tool invocation.
func (l *Logger) LogToolCall(runID, stepID string, tool, version string, input map[string]any) {
	l.Log(runID, stepID, EventToolCall, map[string]any{
		"tool":    tool,
		"version": version,
		"input":   input,
	}, ModelMeta{})
}

// LogToolResult logs a tool result.
func (l *Logger) LogToolResult(runID, stepID string, tool string, output map[string]any, errMsg string, duration string) {
	d := map[string]any{"tool": tool, "output": output}
	if errMsg != "" {
		d["error"] = errMsg
	}
	if duration != "" {
		d["duration"] = duration
	}
	l.Log(runID, stepID, EventToolResult, d, ModelMeta{})
}

// LogError logs an error event.
func (l *Logger) LogError(runID, stepID string, err error) {
	l.Log(runID, stepID, EventError, map[string]any{"message": err.Error()}, ModelMeta{})
}

// LogCompleted logs run completion.
func (l *Logger) LogCompleted(runID string) {
	l.Log(runID, "", EventCompleted, nil, ModelMeta{})
}

// WithFile adds a file output to the logger.
func (l *Logger) WithFile(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	l.AddWriter(f)
	return l, nil
}
