package memory

import (
	"context"
	"testing"
	"time"
)

func TestQdrantStore(t *testing.T) {
	// This is an integration test that requires a running Qdrant instance
	// Skip if not available
	t.Skip("Requires Qdrant instance")

	config := QdrantConfig{
		Host:           "localhost",
		Port:           6333,
		CollectionName: "test_collection",
		VectorSize:     384,
	}

	ctx := context.Background()
	store, err := NewQdrantStore(ctx, config)
	if err != nil {
		t.Fatalf("NewQdrantStore failed: %v", err)
	}
	defer store.Close()

	// Test store
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = 0.1
	}

	err = store.Store(ctx, "test-agent", "test-key", map[string]interface{}{"data": "test"}, embedding)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Test search
	results, err := store.Search(ctx, "test-agent", embedding, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results")
	}
}

func TestWeaviateStore(t *testing.T) {
	// This is an integration test that requires a running Weaviate instance
	// Skip if not available
	t.Skip("Requires Weaviate instance")

	config := WeaviateConfig{
		Host:      "localhost:8080",
		Scheme:    "http",
		ClassName: "TestMemory",
	}

	ctx := context.Background()
	store, err := NewWeaviateStore(ctx, config)
	if err != nil {
		t.Fatalf("NewWeaviateStore failed: %v", err)
	}
	defer store.Close()

	// Test store
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = 0.1
	}

	err = store.Store(ctx, "test-agent", "test-key", map[string]interface{}{"data": "test"}, embedding)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Give time for indexing
	time.Sleep(100 * time.Millisecond)

	// Test search
	results, err := store.Search(ctx, "test-agent", embedding, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results")
	}
}
