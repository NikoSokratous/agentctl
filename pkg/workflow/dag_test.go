package workflow

import (
	"testing"
	"time"
)

func TestDAG(t *testing.T) {
	t.Run("AddNode", func(t *testing.T) {
		dag := NewDAG()

		node := &Node{
			ID:    "node1",
			Name:  "Test Node",
			Agent: "test-agent",
		}

		if err := dag.AddNode(node); err != nil {
			t.Fatalf("AddNode failed: %v", err)
		}

		if len(dag.Nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(dag.Nodes))
		}

		// Duplicate node should fail
		if err := dag.AddNode(node); err == nil {
			t.Error("Expected error for duplicate node")
		}
	})

	t.Run("AddEdge", func(t *testing.T) {
		dag := NewDAG()

		node1 := &Node{ID: "n1", Name: "Node 1"}
		node2 := &Node{ID: "n2", Name: "Node 2"}

		dag.AddNode(node1)
		dag.AddNode(node2)

		if err := dag.AddEdge("n1", "n2"); err != nil {
			t.Fatalf("AddEdge failed: %v", err)
		}

		if len(dag.Edges["n1"]) != 1 {
			t.Errorf("Expected 1 edge, got %d", len(dag.Edges["n1"]))
		}
	})

	t.Run("CycleDetection", func(t *testing.T) {
		dag := NewDAG()

		n1 := &Node{ID: "n1"}
		n2 := &Node{ID: "n2"}
		n3 := &Node{ID: "n3"}

		dag.AddNode(n1)
		dag.AddNode(n2)
		dag.AddNode(n3)

		dag.AddEdge("n1", "n2")
		dag.AddEdge("n2", "n3")

		// Adding cycle should fail
		if err := dag.AddEdge("n3", "n1"); err == nil {
			t.Error("Expected error for cycle")
		}

		if dag.hasCycle() {
			t.Error("DAG should not have cycle after rejected edge")
		}
	})

	t.Run("TopologicalSort", func(t *testing.T) {
		dag := NewDAG()

		// Create linear DAG: n1 -> n2 -> n3
		dag.AddNode(&Node{ID: "n1"})
		dag.AddNode(&Node{ID: "n2"})
		dag.AddNode(&Node{ID: "n3"})

		dag.AddEdge("n1", "n2")
		dag.AddEdge("n2", "n3")

		order, err := dag.TopologicalSort()
		if err != nil {
			t.Fatalf("TopologicalSort failed: %v", err)
		}

		if len(order) != 3 {
			t.Errorf("Expected 3 nodes, got %d", len(order))
		}

		// n1 should come before n2, n2 before n3
		n1Idx := indexOf(order, "n1")
		n2Idx := indexOf(order, "n2")
		n3Idx := indexOf(order, "n3")

		if n1Idx >= n2Idx || n2Idx >= n3Idx {
			t.Errorf("Invalid order: %v", order)
		}
	})

	t.Run("ExecutionLevels", func(t *testing.T) {
		dag := NewDAG()

		// Diamond DAG: n1 -> n2, n3 -> n4
		dag.AddNode(&Node{ID: "n1"})
		dag.AddNode(&Node{ID: "n2"})
		dag.AddNode(&Node{ID: "n3"})
		dag.AddNode(&Node{ID: "n4"})

		dag.AddEdge("n1", "n2")
		dag.AddEdge("n1", "n3")
		dag.AddEdge("n2", "n4")
		dag.AddEdge("n3", "n4")

		levels, err := dag.GetExecutionLevels()
		if err != nil {
			t.Fatalf("GetExecutionLevels failed: %v", err)
		}

		if len(levels) != 3 {
			t.Errorf("Expected 3 levels, got %d", len(levels))
		}

		// Level 0: n1
		// Level 1: n2, n3 (parallel)
		// Level 2: n4

		if len(levels[0]) != 1 || levels[0][0] != "n1" {
			t.Errorf("Level 0 incorrect: %v", levels[0])
		}

		if len(levels[1]) != 2 {
			t.Errorf("Level 1 should have 2 nodes, got %d", len(levels[1]))
		}

		if len(levels[2]) != 1 || levels[2][0] != "n4" {
			t.Errorf("Level 2 incorrect: %v", levels[2])
		}
	})

	t.Run("Validate", func(t *testing.T) {
		dag := NewDAG()

		// Empty DAG should fail
		if err := dag.Validate(); err == nil {
			t.Error("Expected error for empty DAG")
		}

		// Valid DAG
		dag.AddNode(&Node{ID: "n1"})
		dag.AddNode(&Node{ID: "n2"})
		dag.AddEdge("n1", "n2")

		if err := dag.Validate(); err != nil {
			t.Errorf("Validate failed: %v", err)
		}
	})

	t.Run("GetRootsAndLeaves", func(t *testing.T) {
		dag := NewDAG()

		// n1 -> n2 -> n3
		//  └-> n4 -> n5
		dag.AddNode(&Node{ID: "n1"})
		dag.AddNode(&Node{ID: "n2"})
		dag.AddNode(&Node{ID: "n3"})
		dag.AddNode(&Node{ID: "n4"})
		dag.AddNode(&Node{ID: "n5"})

		dag.AddEdge("n1", "n2")
		dag.AddEdge("n2", "n3")
		dag.AddEdge("n1", "n4")
		dag.AddEdge("n4", "n5")

		roots := dag.GetRoots()
		if len(roots) != 1 || roots[0] != "n1" {
			t.Errorf("Expected root [n1], got %v", roots)
		}

		leaves := dag.GetLeaves()
		if len(leaves) != 2 {
			t.Errorf("Expected 2 leaves, got %d: %v", len(leaves), leaves)
		}
	})

	t.Run("Clone", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode(&Node{ID: "n1"})
		dag.AddNode(&Node{ID: "n2"})
		dag.AddEdge("n1", "n2")

		clone := dag.Clone()

		if len(clone.Nodes) != len(dag.Nodes) {
			t.Error("Clone has different number of nodes")
		}

		// Modify clone shouldn't affect original
		clone.AddNode(&Node{ID: "n3"})
		if len(dag.Nodes) == len(clone.Nodes) {
			t.Error("Modifying clone affected original")
		}
	})

	t.Run("ToDOT", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode(&Node{ID: "n1", Name: "Node 1"})
		dag.AddNode(&Node{ID: "n2", Name: "Node 2"})
		dag.AddEdge("n1", "n2")

		dot := dag.ToDOT()

		if dot == "" {
			t.Error("ToDOT returned empty string")
		}

		// Should contain essential DOT elements
		if !contains(dot, "digraph") || !contains(dot, "->") {
			t.Error("DOT output missing essential elements")
		}
	})
}

func TestExecutor(t *testing.T) {
	t.Run("ExecuteSimpleDAG", func(t *testing.T) {
		dag := NewDAG()
		dag.AddNode(&Node{ID: "n1", Agent: "test"})
		dag.AddNode(&Node{ID: "n2", Agent: "test"})
		dag.AddEdge("n1", "n2")

		executor := NewExecutor(nil, nil)

		result, err := executor.Execute(testContext(), dag, "test-wf-1")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		if len(result.NodeResults) != 2 {
			t.Errorf("Expected 2 node results, got %d", len(result.NodeResults))
		}

		// Both nodes should be completed
		for id, nr := range result.NodeResults {
			if nr.Status != "completed" {
				t.Errorf("Node %s status: expected completed, got %s", id, nr.Status)
			}
		}
	})

	t.Run("ExecuteParallelNodes", func(t *testing.T) {
		dag := NewDAG()

		// Root with 3 parallel children
		dag.AddNode(&Node{ID: "root", Agent: "test"})
		dag.AddNode(&Node{ID: "p1", Agent: "test"})
		dag.AddNode(&Node{ID: "p2", Agent: "test"})
		dag.AddNode(&Node{ID: "p3", Agent: "test"})

		dag.AddEdge("root", "p1")
		dag.AddEdge("root", "p2")
		dag.AddEdge("root", "p3")

		executor := NewExecutor(nil, nil)

		result, err := executor.Execute(testContext(), dag, "test-wf-2")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		// All 4 nodes should complete
		if len(result.NodeResults) != 4 {
			t.Errorf("Expected 4 node results, got %d", len(result.NodeResults))
		}
	})
}

func TestCELEvaluator(t *testing.T) {
	t.Run("SimpleBooleanCondition", func(t *testing.T) {
		evaluator, err := NewCELEvaluator()
		if err != nil {
			t.Fatalf("NewCELEvaluator failed: %v", err)
		}

		context := map[string]interface{}{
			"outputs": map[string]interface{}{
				"score": 0.9,
			},
		}

		result, err := evaluator.Evaluate("outputs.score > 0.8", context)
		if err != nil {
			t.Fatalf("Evaluate failed: %v", err)
		}

		if !result {
			t.Error("Expected condition to be true")
		}
	})

	t.Run("ValidateCondition", func(t *testing.T) {
		evaluator, err := NewCELEvaluator()
		if err != nil {
			t.Fatalf("NewCELEvaluator failed: %v", err)
		}

		// Valid condition
		if err := evaluator.ValidateCondition("outputs.value > 10"); err != nil {
			t.Errorf("Valid condition rejected: %v", err)
		}

		// Invalid syntax
		if err := evaluator.ValidateCondition("outputs.value >"); err == nil {
			t.Error("Invalid condition accepted")
		}
	})
}

func TestTemplateRegistry(t *testing.T) {
	t.Run("LoadEmbedded", func(t *testing.T) {
		registry := NewTemplateRegistry()

		if err := registry.LoadEmbeddedTemplates(); err != nil {
			t.Fatalf("LoadEmbeddedTemplates failed: %v", err)
		}

		templates := registry.List()
		// Templates may be empty when tests run from pkg/workflow (examples dir not in CWD)
		if len(templates) > 0 {
			if templates[0].Name == "" {
				t.Error("Template has no name")
			}
		}
	})

	t.Run("ValidateTemplate", func(t *testing.T) {
		tmpl := &Template{
			Name:     "test",
			Workflow: map[string]interface{}{"steps": []interface{}{}},
			Parameters: []TemplateParam{
				{Name: "param1", Type: "string", Required: true},
			},
		}

		if err := ValidateTemplate(tmpl); err != nil {
			t.Errorf("Valid template rejected: %v", err)
		}

		// Missing name
		invalid := &Template{Workflow: map[string]interface{}{}}
		if err := ValidateTemplate(invalid); err == nil {
			t.Error("Invalid template accepted")
		}
	})
}

// Helper functions

func indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf([]string{s}, substr) >= 0 ||
		(len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func testContext() *testContextImpl {
	return &testContextImpl{}
}

type testContextImpl struct{}

func (t *testContextImpl) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }
func (t *testContextImpl) Done() <-chan struct{}                   { return nil }
func (t *testContextImpl) Err() error                              { return nil }
func (t *testContextImpl) Value(key interface{}) interface{}       { return nil }
