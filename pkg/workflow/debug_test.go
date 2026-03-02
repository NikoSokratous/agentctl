package workflow

import (
	"context"
	"testing"
	"time"
)

func TestDebugSession(t *testing.T) {
	workflowID := "test-workflow"
	session := NewDebugSession(workflowID)

	if session.WorkflowID != workflowID {
		t.Errorf("Expected workflow ID %s, got %s", workflowID, session.WorkflowID)
	}

	// Test adding breakpoint
	session.AddBreakpoint("node1", "")
	if len(session.Breakpoints) != 1 {
		t.Errorf("Expected 1 breakpoint, got %d", len(session.Breakpoints))
	}

	// Test conditional breakpoint
	session.AddBreakpoint("node2", "outputs.node1.status == 'success'")
	shouldPause, err := session.ShouldPause("node2", map[string]interface{}{
		"outputs": map[string]interface{}{
			"node1": map[string]interface{}{
				"status": "success",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error checking breakpoint: %v", err)
	}
	if !shouldPause {
		t.Error("Expected to pause on conditional breakpoint")
	}

	// Test removing breakpoint
	session.RemoveBreakpoint("node1")
	if len(session.Breakpoints) != 1 {
		t.Errorf("Expected 1 breakpoint after removal, got %d", len(session.Breakpoints))
	}

	// Test pause/resume
	session.Pause()
	if !session.Paused {
		t.Error("Expected session to be paused")
	}

	session.Resume()
	if session.Paused {
		t.Error("Expected session to be resumed")
	}
}

func TestDebugExecutor(t *testing.T) {
	// Create a simple DAG
	dag := NewDAG()
	dag.AddNode(&Node{ID: "start", Agent: "test", Name: "Start"})
	dag.AddNode(&Node{ID: "process", Agent: "test", Name: "Process"})
	dag.AddNode(&Node{ID: "end", Agent: "test", Name: "End"})
	dag.AddEdge("start", "process")
	dag.AddEdge("process", "end")

	// Create debug session
	session := NewDebugSession("test-workflow")
	session.AddBreakpoint("process", "")

	// Create executor
	stateStore := NewStateStore(nil) // nil DB for testing
	executor := NewDebugExecutor(stateStore, session)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Resume after short delay to prevent deadlock
	go func() {
		time.Sleep(100 * time.Millisecond)
		session.Resume()
	}()

	outputs, err := executor.ExecuteDAGWithDebug(ctx, dag, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execution error: %v", err)
	}

	if len(outputs) != 3 {
		t.Errorf("Expected 3 outputs, got %d", len(outputs))
	}
}

func TestDebugState(t *testing.T) {
	session := NewDebugSession("test-workflow")
	session.AddBreakpoint("node1", "")
	session.UpdateVariables(map[string]interface{}{
		"var1": "value1",
		"var2": 123,
	})

	state := session.GetState()

	if state["workflow_id"] != "test-workflow" {
		t.Error("Workflow ID mismatch in state")
	}

	vars, ok := state["variables"].(map[string]interface{})
	if !ok {
		t.Fatal("Variables not in state")
	}

	if vars["var1"] != "value1" {
		t.Error("Variable var1 mismatch")
	}
}
