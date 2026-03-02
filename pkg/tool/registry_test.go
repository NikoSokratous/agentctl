package tool

import (
	"context"
	"encoding/json"
	"testing"
)

func TestToolRegistry(t *testing.T) {
	reg := NewRegistry()

	// Create a mock tool
	mockTool := &mockTool{
		name:    "test",
		version: "1",
	}

	reg.Register(mockTool)

	// Test Get
	tool, ok := reg.Get("test", "1")
	if !ok {
		t.Error("Tool not found after registration")
	}
	if tool.Name() != "test" {
		t.Errorf("Tool name = %v, want test", tool.Name())
	}

	// Test Get non-existent
	_, ok = reg.Get("nonexistent", "1")
	if ok {
		t.Error("Found non-existent tool")
	}

	// Test List
	list := reg.List()
	if len(list) != 1 {
		t.Errorf("List length = %v, want 1", len(list))
	}
	if list[0] != "test@1" {
		t.Errorf("List[0] = %v, want test@1", list[0])
	}
}

func TestToolExecution(t *testing.T) {
	reg := NewRegistry()
	mockTool := &mockTool{
		name:    "echo",
		version: "1",
		output:  map[string]any{"result": "success"},
	}
	reg.Register(mockTool)

	ctx := context.Background()
	input := json.RawMessage(`{"message":"hello"}`)

	result, err := reg.Execute(ctx, "echo", "1", input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ToolID != "echo@1" {
		t.Errorf("ToolID = %v, want echo@1", result.ToolID)
	}
	if result.Output["result"] != "success" {
		t.Errorf("Output result = %v, want success", result.Output["result"])
	}
}

// Mock tool for testing
type mockTool struct {
	name    string
	version string
	output  map[string]any
	err     error
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Version() string     { return m.version }
func (m *mockTool) Description() string { return "mock tool" }
func (m *mockTool) InputSchema() ([]byte, error) {
	return []byte(`{"type":"object"}`), nil
}
func (m *mockTool) Permissions() []Permission { return nil }
func (m *mockTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return map[string]any{}, nil
}
