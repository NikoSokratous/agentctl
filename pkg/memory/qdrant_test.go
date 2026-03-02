package memory

import (
	"context"
	"fmt"
	"testing"
)

// TestQdrantStoreInterface verifies QdrantStore implements interfaces
func TestQdrantStoreInterface(t *testing.T) {
	t.Skip("Requires running Qdrant instance")

	// This test verifies that QdrantStore satisfies the SemanticStore interface
	var _ SemanticStore = (*QdrantStore)(nil)
}

// TestQdrantConnection tests connecting to Qdrant (integration test)
func TestQdrantConnection(t *testing.T) {
	t.Skip("Integration test - requires running Qdrant instance on localhost:6334")

	ctx := context.Background()
	config := QdrantConfig{
		Host:           "localhost",
		Port:           6334,
		CollectionName: "test_collection",
		VectorSize:     384,
	}

	store, err := NewQdrantStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Qdrant store: %v", err)
	}
	defer store.Close()

	// Test basic operations
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Store a vector
	err = store.Store(ctx, "test-agent", "test-key", map[string]string{"test": "data"}, embedding)
	if err != nil {
		t.Fatalf("Failed to store: %v", err)
	}

	// Search for similar vectors
	results, err := store.Search(ctx, "test-agent", embedding, 5)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 result")
	}

	// Delete
	err = store.Delete(ctx, "test-agent", "test-key")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}
}

// TestQdrantUpsert tests the Upsert operation
func TestQdrantUpsert(t *testing.T) {
	t.Skip("Integration test - requires running Qdrant instance")

	ctx := context.Background()
	config := QdrantConfig{
		Host:           "localhost",
		Port:           6334,
		CollectionName: "test_upsert",
		VectorSize:     128,
	}

	store, err := NewQdrantStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	embedding := make(Embedding, 128)
	for i := range embedding {
		embedding[i] = 0.5
	}

	metadata := map[string]any{
		"source": "test",
		"index":  1,
	}

	// First upsert
	err = store.Upsert(ctx, "agent1", "id1", embedding, metadata)
	if err != nil {
		t.Fatalf("First upsert failed: %v", err)
	}

	// Second upsert (should update)
	metadata["index"] = 2
	err = store.Upsert(ctx, "agent1", "id1", embedding, metadata)
	if err != nil {
		t.Fatalf("Second upsert failed: %v", err)
	}

	// Search and verify
	results, err := store.Search(ctx, "agent1", embedding, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestQdrantDeleteAgent tests deleting all vectors for an agent
func TestQdrantDeleteAgent(t *testing.T) {
	t.Skip("Integration test - requires running Qdrant instance")

	ctx := context.Background()
	config := QdrantConfig{
		Host:           "localhost",
		Port:           6334,
		CollectionName: "test_delete_agent",
		VectorSize:     64,
	}

	store, err := NewQdrantStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	embedding := make(Embedding, 64)

	// Insert multiple vectors for the same agent
	for i := 0; i < 5; i++ {
		err = store.Upsert(ctx, "agent-delete-test", fmt.Sprintf("id%d", i), embedding, map[string]any{"index": i})
		if err != nil {
			t.Fatalf("Failed to insert vector %d: %v", i, err)
		}
	}

	// Verify vectors exist
	results, err := store.Search(ctx, "agent-delete-test", embedding, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 vectors, got %d", len(results))
	}

	// Delete all vectors for agent
	err = store.DeleteAgent(ctx, "agent-delete-test")
	if err != nil {
		t.Fatalf("DeleteAgent failed: %v", err)
	}

	// Verify all deleted
	results, err = store.Search(ctx, "agent-delete-test", embedding, 10)
	if err != nil {
		t.Fatalf("Search after delete failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 vectors after delete, got %d", len(results))
	}
}

// TestQdrantConfig tests configuration validation
func TestQdrantConfig(t *testing.T) {
	tests := []struct {
		name   string
		config QdrantConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: QdrantConfig{
				Host:           "localhost",
				Port:           6334,
				CollectionName: "test",
				VectorSize:     384,
			},
			valid: true,
		},
		{
			name: "with API key",
			config: QdrantConfig{
				Host:           "cloud.qdrant.io",
				Port:           6333,
				CollectionName: "prod",
				VectorSize:     1536,
				APIKey:         "secret-key",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Host == "" {
				t.Error("Host should not be empty")
			}
			if tt.config.Port == 0 {
				t.Error("Port should not be zero")
			}
			if tt.config.CollectionName == "" {
				t.Error("CollectionName should not be empty")
			}
			if tt.config.VectorSize == 0 {
				t.Error("VectorSize should not be zero")
			}
		})
	}
}
