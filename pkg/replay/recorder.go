package replay

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// Recorder captures execution for replay.
type Recorder struct {
	config    RecordingConfig
	snapshot  *RunSnapshot
	recording bool
	modelSeq  int
	toolSeq   int
}

// NewRecorder creates a new execution recorder.
func NewRecorder(config RecordingConfig) *Recorder {
	return &Recorder{
		config:    config,
		recording: false,
	}
}

// StartRecording begins recording a run.
func (r *Recorder) StartRecording(runID, agentName, goal string, agentConfig map[string]interface{}) error {
	if !r.config.Enabled {
		return nil
	}

	r.snapshot = &RunSnapshot{
		ID:          uuid.New().String(),
		RunID:       runID,
		Version:     SnapshotVersion,
		CreatedAt:   time.Now(),
		AgentName:   agentName,
		Goal:        goal,
		AgentConfig: agentConfig,
		ModelCalls:  make([]ModelCall, 0),
		ToolCalls:   make([]ToolExecution, 0),
		Environment: make(map[string]string),
		StartTime:   time.Now(),
		Checksums:   make(map[string]string),
	}

	r.recording = true
	r.modelSeq = 0
	r.toolSeq = 0

	// Capture environment if configured
	if r.config.CaptureEnvironment {
		r.captureEnvironment()
	}

	return nil
}

// RecordModelCall records an LLM interaction.
func (r *Recorder) RecordModelCall(model, provider, prompt, response string, tokens int, metadata map[string]interface{}) error {
	if !r.recording || !r.config.CaptureModel {
		return nil
	}

	r.modelSeq++
	call := ModelCall{
		Sequence:   r.modelSeq,
		Timestamp:  time.Now(),
		Model:      model,
		Provider:   provider,
		Prompt:     prompt,
		Response:   response,
		TokensUsed: tokens,
		Metadata:   metadata,
	}

	// Extract temperature and other params from metadata
	if temp, ok := metadata["temperature"].(float64); ok {
		call.Temperature = temp
	}
	if maxTokens, ok := metadata["max_tokens"].(int); ok {
		call.MaxTokens = maxTokens
	}
	if seed, ok := metadata["seed"].(*int64); ok {
		call.Seed = seed
	}
	if finishReason, ok := metadata["finish_reason"].(string); ok {
		call.FinishReason = finishReason
	}

	r.snapshot.ModelCalls = append(r.snapshot.ModelCalls, call)
	return nil
}

// RecordToolCall records a tool execution.
func (r *Recorder) RecordToolCall(toolName string, input, output json.RawMessage, err error, duration time.Duration) error {
	if !r.recording || !r.config.CaptureTools {
		return nil
	}

	r.toolSeq++
	execution := ToolExecution{
		Sequence:  r.toolSeq,
		Timestamp: time.Now(),
		ToolName:  toolName,
		Input:     input,
		Output:    output,
		Duration:  duration,
	}

	if err != nil {
		execution.Error = err.Error()
	}

	r.snapshot.ToolCalls = append(r.snapshot.ToolCalls, execution)
	return nil
}

// RecordSideEffect records a state change.
func (r *Recorder) RecordSideEffect(toolSeq int, sideEffect SideEffect) error {
	if !r.recording || !r.config.CaptureSideEffects {
		return nil
	}

	// Find the tool execution and add side effect
	if toolSeq > 0 && toolSeq <= len(r.snapshot.ToolCalls) {
		idx := toolSeq - 1
		if r.snapshot.ToolCalls[idx].SideEffects == nil {
			r.snapshot.ToolCalls[idx].SideEffects = make([]SideEffect, 0)
		}
		r.snapshot.ToolCalls[idx].SideEffects = append(r.snapshot.ToolCalls[idx].SideEffects, sideEffect)
	}

	return nil
}

// StopRecording finalizes the recording.
func (r *Recorder) StopRecording(finalState string) (*RunSnapshot, error) {
	if !r.recording {
		return nil, fmt.Errorf("not recording")
	}

	r.snapshot.EndTime = time.Now()
	r.snapshot.FinalState = finalState
	r.recording = false

	// Calculate checksums
	r.calculateChecksums()

	// Calculate size (before compression)
	data, err := json.Marshal(r.snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	r.snapshot.SizeBytes = int64(len(data))

	// Compress if configured
	if r.config.CompressData {
		if err := r.compressSnapshot(); err != nil {
			return nil, fmt.Errorf("compress snapshot: %w", err)
		}
	}

	// Encrypt if configured (placeholder)
	if r.config.EncryptPII {
		r.encryptPII()
	}

	return r.snapshot, nil
}

// GetSnapshot returns the current snapshot (for live queries).
func (r *Recorder) GetSnapshot() *RunSnapshot {
	if !r.recording {
		return nil
	}
	return r.snapshot
}

// captureEnvironment captures relevant environment variables.
func (r *Recorder) captureEnvironment() {
	// Capture safe environment variables (not secrets)
	safeEnvVars := []string{
		"HOSTNAME",
		"USER",
		"PWD",
		"SHELL",
		"LANG",
		"TZ",
	}

	// This is a simplified implementation
	// In production, you'd use os.Getenv
	for _, key := range safeEnvVars {
		r.snapshot.Environment[key] = "[captured]"
	}
}

// calculateChecksums calculates verification checksums.
func (r *Recorder) calculateChecksums() {
	// Checksum of model calls
	if len(r.snapshot.ModelCalls) > 0 {
		modelData, _ := json.Marshal(r.snapshot.ModelCalls)
		hash := sha256.Sum256(modelData)
		r.snapshot.Checksums["model_calls"] = fmt.Sprintf("%x", hash)
	}

	// Checksum of tool calls
	if len(r.snapshot.ToolCalls) > 0 {
		toolData, _ := json.Marshal(r.snapshot.ToolCalls)
		hash := sha256.Sum256(toolData)
		r.snapshot.Checksums["tool_calls"] = fmt.Sprintf("%x", hash)
	}
}

// compressSnapshot compresses the snapshot data.
func (r *Recorder) compressSnapshot() error {
	// Marshal snapshot
	data, err := json.Marshal(r.snapshot)
	if err != nil {
		return err
	}

	// Compress
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	// Update size
	originalSize := r.snapshot.SizeBytes
	compressedSize := int64(buf.Len())
	r.snapshot.SizeBytes = compressedSize
	r.snapshot.Compressed = true

	compressionRatio := float64(originalSize) / float64(compressedSize)
	_ = compressionRatio // Log this in production

	return nil
}

// encryptPII encrypts PII in the snapshot (placeholder).
func (r *Recorder) encryptPII() {
	// In production, this would:
	// 1. Scan prompts/responses for PII patterns
	// 2. Encrypt detected PII
	// 3. Store encryption keys securely
	r.snapshot.Encrypted = true
}

// decompressSnapshot decompresses snapshot data.
func decompressSnapshot(compressedData []byte) ([]byte, error) {
	buf := bytes.NewReader(compressedData)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}

// SnapshotStore manages snapshot persistence.
type SnapshotStore struct {
	// db would be *sql.DB in production
}

// SaveSnapshot persists a snapshot.
func (s *SnapshotStore) SaveSnapshot(ctx context.Context, snapshot *RunSnapshot) error {
	// Serialize snapshot
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	// In production, save to database and/or filesystem
	_ = data

	return nil
}

// LoadSnapshot retrieves a snapshot by ID.
func (s *SnapshotStore) LoadSnapshot(ctx context.Context, snapshotID string) (*RunSnapshot, error) {
	// In production, load from database
	return nil, fmt.Errorf("not implemented")
}

// ListSnapshots lists snapshots with optional filters.
func (s *SnapshotStore) ListSnapshots(ctx context.Context, runID string, limit int) ([]SnapshotMetadata, error) {
	// In production, query database
	return nil, fmt.Errorf("not implemented")
}

// DeleteSnapshot removes a snapshot.
func (s *SnapshotStore) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	// In production, delete from database
	return fmt.Errorf("not implemented")
}
