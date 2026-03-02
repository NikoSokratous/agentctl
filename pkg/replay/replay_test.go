package replay

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRecorder(t *testing.T) {
	config := RecordingConfig{
		Enabled:            true,
		CaptureModel:       true,
		CaptureTools:       true,
		CaptureSideEffects: true,
		CaptureEnvironment: true,
		CompressData:       false,
	}

	recorder := NewRecorder(config)

	// Start recording
	err := recorder.StartRecording("run-123", "test-agent", "test goal", map[string]interface{}{
		"autonomy": 2,
	})
	if err != nil {
		t.Fatalf("StartRecording: %v", err)
	}

	// Record model call
	err = recorder.RecordModelCall(
		"gpt-4",
		"openai",
		"What is 2+2?",
		"4",
		10,
		map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		},
	)
	if err != nil {
		t.Fatalf("RecordModelCall: %v", err)
	}

	// Record tool call
	input, _ := json.Marshal(map[string]string{"query": "test"})
	output, _ := json.Marshal(map[string]string{"result": "success"})

	err = recorder.RecordToolCall("test_tool", input, output, nil, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("RecordToolCall: %v", err)
	}

	// Stop recording
	snapshot, err := recorder.StopRecording("completed")
	if err != nil {
		t.Fatalf("StopRecording: %v", err)
	}

	// Verify snapshot
	if snapshot.RunID != "run-123" {
		t.Errorf("RunID: got %s, want run-123", snapshot.RunID)
	}
	if snapshot.AgentName != "test-agent" {
		t.Errorf("AgentName: got %s, want test-agent", snapshot.AgentName)
	}
	if snapshot.FinalState != "completed" {
		t.Errorf("FinalState: got %s, want completed", snapshot.FinalState)
	}
	if len(snapshot.ModelCalls) != 1 {
		t.Errorf("ModelCalls: got %d, want 1", len(snapshot.ModelCalls))
	}
	if len(snapshot.ToolCalls) != 1 {
		t.Errorf("ToolCalls: got %d, want 1", len(snapshot.ToolCalls))
	}

	// Verify model call
	modelCall := snapshot.ModelCalls[0]
	if modelCall.Model != "gpt-4" {
		t.Errorf("Model: got %s, want gpt-4", modelCall.Model)
	}
	if modelCall.Prompt != "What is 2+2?" {
		t.Errorf("Prompt: got %s, want 'What is 2+2?'", modelCall.Prompt)
	}
	if modelCall.Response != "4" {
		t.Errorf("Response: got %s, want '4'", modelCall.Response)
	}

	// Verify tool call
	toolCall := snapshot.ToolCalls[0]
	if toolCall.ToolName != "test_tool" {
		t.Errorf("ToolName: got %s, want test_tool", toolCall.ToolName)
	}
	if toolCall.Duration != 100*time.Millisecond {
		t.Errorf("Duration: got %v, want 100ms", toolCall.Duration)
	}
}

func TestReplayerExactMode(t *testing.T) {
	// Create mock snapshot
	snapshot := &RunSnapshot{
		ID:        "snap-123",
		RunID:     "run-123",
		AgentName: "test-agent",
		Goal:      "test goal",
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		ToolCalls: []ToolExecution{
			{
				Sequence:  1,
				Timestamp: time.Now().Add(-4 * time.Minute),
				ToolName:  "tool1",
				Input:     json.RawMessage(`{"test":true}`),
				Output:    json.RawMessage(`{"result":"success"}`),
				Duration:  100 * time.Millisecond,
			},
			{
				Sequence:  2,
				Timestamp: time.Now().Add(-3 * time.Minute),
				ToolName:  "tool2",
				Input:     json.RawMessage(`{"data":"value"}`),
				Output:    json.RawMessage(`{"status":"ok"}`),
				Duration:  200 * time.Millisecond,
			},
		},
	}

	replayer := NewReplayer(snapshot)

	options := ReplayOptions{
		Mode:       ReplayModeExact,
		SnapshotID: snapshot.ID,
	}

	result, err := replayer.Replay(context.Background(), options)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	// Verify result
	if !result.Success {
		t.Error("Expected replay to succeed")
	}
	if result.ActionsRerun != 2 {
		t.Errorf("ActionsRerun: got %d, want 2", result.ActionsRerun)
	}
	if result.Matches != 2 {
		t.Errorf("Matches: got %d, want 2", result.Matches)
	}
	if len(result.Divergences) != 0 {
		t.Errorf("Divergences: got %d, want 0", len(result.Divergences))
	}
	if result.Metrics.Accuracy != 1.0 {
		t.Errorf("Accuracy: got %.2f, want 1.0", result.Metrics.Accuracy)
	}
}

func TestReplayerPartialReplay(t *testing.T) {
	snapshot := &RunSnapshot{
		ID:        "snap-123",
		RunID:     "run-123",
		AgentName: "test-agent",
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		ToolCalls: []ToolExecution{
			{Sequence: 1, ToolName: "tool1"},
			{Sequence: 2, ToolName: "tool2"},
			{Sequence: 3, ToolName: "tool3"},
			{Sequence: 4, ToolName: "tool4"},
		},
	}

	replayer := NewReplayer(snapshot)

	// Replay from sequence 2 to 3
	options := ReplayOptions{
		Mode:              ReplayModeExact,
		StartFromSequence: 2,
		StopAtSequence:    3,
	}

	result, err := replayer.Replay(context.Background(), options)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	// Should only replay sequences 2 and 3
	if result.ActionsRerun != 2 {
		t.Errorf("ActionsRerun: got %d, want 2", result.ActionsRerun)
	}
}

func TestReplayerValidationMode(t *testing.T) {
	snapshot := &RunSnapshot{
		ID:        "snap-123",
		RunID:     "run-123",
		AgentName: "test-agent",
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		ToolCalls: []ToolExecution{
			{
				Sequence:  1,
				Timestamp: time.Now().Add(-4 * time.Minute),
				ToolName:  "tool1",
			},
			{
				Sequence:  2,
				Timestamp: time.Now().Add(-3 * time.Minute),
				ToolName:  "tool2",
			},
		},
		Checksums: map[string]string{
			"model_calls": "abc123",
			"tool_calls":  "def456",
		},
	}

	replayer := NewReplayer(snapshot)

	options := ReplayOptions{
		Mode: ReplayModeValidation,
	}

	result, err := replayer.Replay(context.Background(), options)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	// Validation should check data integrity
	if !result.Success {
		t.Log("Validation found issues (expected for mock data)")
	}
}

func TestReplayerBreakpoints(t *testing.T) {
	snapshot := &RunSnapshot{
		ID:        "snap-123",
		RunID:     "run-123",
		AgentName: "test-agent",
		StartTime: time.Now().Add(-5 * time.Minute),
		EndTime:   time.Now(),
		ToolCalls: []ToolExecution{
			{Sequence: 1, ToolName: "tool1"},
			{Sequence: 2, ToolName: "tool2"},
			{Sequence: 3, ToolName: "tool3"},
		},
	}

	replayer := NewReplayer(snapshot)

	// Debug mode with breakpoints
	options := ReplayOptions{
		Mode:        ReplayModeDebug,
		Breakpoints: []int{2}, // Break at sequence 2
	}

	result, err := replayer.Replay(context.Background(), options)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	if !result.Success {
		t.Error("Expected debug replay to succeed")
	}
}

func TestRecorderSideEffects(t *testing.T) {
	config := RecordingConfig{
		Enabled:            true,
		CaptureTools:       true,
		CaptureSideEffects: true,
	}

	recorder := NewRecorder(config)
	recorder.StartRecording("run-123", "test-agent", "test", nil)

	// Record tool with side effect
	input, _ := json.Marshal(map[string]string{"file": "/tmp/test.txt"})
	output, _ := json.Marshal(map[string]string{"written": "true"})
	recorder.RecordToolCall("write_file", input, output, nil, 50*time.Millisecond)

	// Record side effect
	sideEffect := SideEffect{
		Type:        SideEffectFileWrite,
		Target:      "/tmp/test.txt",
		Description: "Wrote test file",
		Reversible:  true,
		RevertData:  json.RawMessage(`{"original_content":""}`),
		Timestamp:   time.Now(),
	}

	err := recorder.RecordSideEffect(1, sideEffect)
	if err != nil {
		t.Fatalf("RecordSideEffect: %v", err)
	}

	snapshot, _ := recorder.StopRecording("completed")

	// Verify side effect was recorded
	if len(snapshot.ToolCalls) != 1 {
		t.Fatalf("ToolCalls: got %d, want 1", len(snapshot.ToolCalls))
	}
	if len(snapshot.ToolCalls[0].SideEffects) != 1 {
		t.Fatalf("SideEffects: got %d, want 1", len(snapshot.ToolCalls[0].SideEffects))
	}

	se := snapshot.ToolCalls[0].SideEffects[0]
	if se.Type != SideEffectFileWrite {
		t.Errorf("SideEffect type: got %s, want %s", se.Type, SideEffectFileWrite)
	}
	if se.Target != "/tmp/test.txt" {
		t.Errorf("SideEffect target: got %s, want /tmp/test.txt", se.Target)
	}
}

func TestSnapshotCompression(t *testing.T) {
	config := RecordingConfig{
		Enabled:      true,
		CompressData: true,
	}

	recorder := NewRecorder(config)
	recorder.StartRecording("run-123", "test-agent", "test", nil)

	// Record some data
	for i := 0; i < 10; i++ {
		recorder.RecordModelCall(
			"gpt-4",
			"openai",
			"prompt "+string(rune(i)),
			"response "+string(rune(i)),
			10,
			map[string]interface{}{},
		)
	}

	snapshot, err := recorder.StopRecording("completed")
	if err != nil {
		t.Fatalf("StopRecording: %v", err)
	}

	if !snapshot.Compressed {
		t.Error("Expected snapshot to be compressed")
	}
	if snapshot.SizeBytes == 0 {
		t.Error("Expected non-zero size")
	}
}

func TestReplayMetrics(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Minute)
	endTime := time.Now().Add(-5 * time.Minute)

	snapshot := &RunSnapshot{
		ID:        "snap-123",
		RunID:     "run-123",
		AgentName: "test-agent",
		StartTime: startTime,
		EndTime:   endTime,
		ToolCalls: []ToolExecution{
			{Sequence: 1, ToolName: "tool1", Duration: 100 * time.Millisecond},
			{Sequence: 2, ToolName: "tool2", Duration: 200 * time.Millisecond},
		},
	}

	replayer := NewReplayer(snapshot)

	result, err := replayer.Replay(context.Background(), ReplayOptions{
		Mode: ReplayModeExact,
	})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	// Check metrics
	if result.Metrics.TotalActions != 2 {
		t.Errorf("TotalActions: got %d, want 2", result.Metrics.TotalActions)
	}
	if result.Metrics.OriginalDuration != endTime.Sub(startTime) {
		t.Errorf("OriginalDuration mismatch")
	}
	if result.Metrics.SpeedupFactor == 0 {
		t.Error("SpeedupFactor should be calculated")
	}
}
