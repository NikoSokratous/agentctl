package memory

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// TestSemanticStoreInterface ensures both implementations satisfy the interface
func TestSemanticStoreInterface(t *testing.T) {
	var _ SemanticStore = (*QdrantStore)(nil)
	var _ SemanticStore = (*WeaviateStore)(nil)
}

// TestQdrantStoreComprehensive is a comprehensive integration test for Qdrant
func TestQdrantStoreComprehensive(t *testing.T) {
	t.Skip("Requires Qdrant instance")

	ctx := context.Background()
	config := QdrantConfig{
		Host:           "localhost",
		Port:           6334,
		CollectionName: fmt.Sprintf("test_%d", time.Now().Unix()),
		VectorSize:     128,
	}

	store, err := NewQdrantStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Qdrant store: %v", err)
	}
	defer store.Close()

	testSemanticStoreOperations(t, store)
}

// TestWeaviateStoreComprehensive is a comprehensive integration test for Weaviate
func TestWeaviateStoreComprehensive(t *testing.T) {
	t.Skip("Requires Weaviate instance")

	ctx := context.Background()
	config := WeaviateConfig{
		Host:      "localhost:8080",
		Scheme:    "http",
		ClassName: fmt.Sprintf("Test%d", time.Now().Unix()),
	}

	store, err := NewWeaviateStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Weaviate store: %v", err)
	}
	defer store.Close()

	testSemanticStoreOperations(t, store)
}

// testSemanticStoreOperations runs a common test suite on any SemanticStore
func testSemanticStoreOperations(t *testing.T, store SemanticStore) {
	ctx := context.Background()

	// Test 1: Upsert and Search
	t.Run("UpsertAndSearch", func(t *testing.T) {
		// Create test vectors
		vectors := []struct {
			id       string
			content  string
			vector   Embedding
			metadata map[string]any
		}{
			{
				id:      "doc1",
				content: "The cat sat on the mat",
				vector:  generateTestVector(128, 0.1),
				metadata: map[string]any{
					"content": "The cat sat on the mat",
					"topic":   "animals",
				},
			},
			{
				id:      "doc2",
				content: "The dog ran in the park",
				vector:  generateTestVector(128, 0.2),
				metadata: map[string]any{
					"content": "The dog ran in the park",
					"topic":   "animals",
				},
			},
			{
				id:      "doc3",
				content: "Machine learning is fascinating",
				vector:  generateTestVector(128, 0.8),
				metadata: map[string]any{
					"content": "Machine learning is fascinating",
					"topic":   "technology",
				},
			},
		}

		// Upsert all vectors
		for _, v := range vectors {
			err := store.Upsert(ctx, "test-agent", v.id, v.vector, v.metadata)
			if err != nil {
				t.Fatalf("Failed to upsert %s: %v", v.id, err)
			}
		}

		// Search for similar vectors
		queryVector := generateTestVector(128, 0.15)
		results, err := store.Search(ctx, "test-agent", queryVector, 2)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		}

		t.Logf("Found %d results", len(results))
		for i, result := range results {
			t.Logf("Result %d: ID=%s, Score=%.3f", i+1, result.ID, result.Score)
		}
	})

	// Test 2: Update existing vector
	t.Run("UpdateVector", func(t *testing.T) {
		id := "update-test"
		metadata1 := map[string]any{"version": 1, "content": "original"}
		metadata2 := map[string]any{"version": 2, "content": "updated"}

		vector := generateTestVector(128, 0.5)

		// First upsert
		err := store.Upsert(ctx, "test-agent", id, vector, metadata1)
		if err != nil {
			t.Fatalf("First upsert failed: %v", err)
		}

		// Second upsert (update)
		err = store.Upsert(ctx, "test-agent", id, vector, metadata2)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Search to verify update
		results, err := store.Search(ctx, "test-agent", vector, 1)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected at least one result")
		}

		t.Logf("Updated vector found: %+v", results[0].Metadata)
	})

	// Test 3: Agent isolation
	t.Run("AgentIsolation", func(t *testing.T) {
		vector := generateTestVector(128, 0.3)
		metadata := map[string]any{"test": "isolation"}

		// Insert for agent1
		err := store.Upsert(ctx, "agent1", "iso-doc1", vector, metadata)
		if err != nil {
			t.Fatalf("Agent1 upsert failed: %v", err)
		}

		// Insert for agent2
		err = store.Upsert(ctx, "agent2", "iso-doc2", vector, metadata)
		if err != nil {
			t.Fatalf("Agent2 upsert failed: %v", err)
		}

		// Search for agent1 - should only find agent1's vectors
		results, err := store.Search(ctx, "agent1", vector, 10)
		if err != nil {
			t.Fatalf("Agent1 search failed: %v", err)
		}

		for _, result := range results {
			if result.ID == "iso-doc2" {
				t.Error("Agent1 search returned agent2's document")
			}
		}

		t.Logf("Agent isolation verified: %d results for agent1", len(results))
	})

	// Test 4: Delete agent
	t.Run("DeleteAgent", func(t *testing.T) {
		agentID := fmt.Sprintf("delete-test-%d", time.Now().Unix())
		vector := generateTestVector(128, 0.4)

		// Insert multiple vectors
		for i := 0; i < 5; i++ {
			metadata := map[string]any{"index": i}
			err := store.Upsert(ctx, agentID, fmt.Sprintf("del-doc%d", i), vector, metadata)
			if err != nil {
				t.Fatalf("Upsert %d failed: %v", i, err)
			}
		}

		// Verify vectors exist
		results, err := store.Search(ctx, agentID, vector, 10)
		if err != nil {
			t.Fatalf("Pre-delete search failed: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("Expected 5 vectors before delete, got %d", len(results))
		}

		// Delete all agent vectors
		err = store.DeleteAgent(ctx, agentID)
		if err != nil {
			t.Fatalf("DeleteAgent failed: %v", err)
		}

		// Verify all deleted
		results, err = store.Search(ctx, agentID, vector, 10)
		if err != nil {
			t.Fatalf("Post-delete search failed: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 vectors after delete, got %d", len(results))
		}

		t.Log("DeleteAgent verified successfully")
	})

	// Test 5: Search relevance
	t.Run("SearchRelevance", func(t *testing.T) {
		// Create vectors with different similarities
		baseVector := generateTestVector(128, 0.5)

		vectors := []struct {
			id     string
			vector Embedding
		}{
			{"exact", baseVector},                     // Exact match
			{"close", perturbVector(baseVector, 0.1)}, // Close match
			{"far", perturbVector(baseVector, 0.5)},   // Far match
		}

		for _, v := range vectors {
			metadata := map[string]any{"type": v.id}
			err := store.Upsert(ctx, "relevance-test", v.id, v.vector, metadata)
			if err != nil {
				t.Fatalf("Upsert %s failed: %v", v.id, err)
			}
		}

		// Search with base vector
		results, err := store.Search(ctx, "relevance-test", baseVector, 3)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected results")
		}

		// Verify ordering (scores should decrease)
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Errorf("Results not ordered by score: %f > %f", results[i].Score, results[i-1].Score)
			}
		}

		t.Logf("Search relevance verified with %d results", len(results))
		for i, r := range results {
			t.Logf("  %d. ID=%s Score=%.4f", i+1, r.ID, r.Score)
		}
	})
}

// generateTestVector creates a random test vector
func generateTestVector(size int, seed float64) Embedding {
	rand.Seed(int64(seed * 1000000))
	vector := make(Embedding, size)
	for i := range vector {
		vector[i] = float32(rand.NormFloat64())
	}
	return normalizeVector(vector)
}

// perturbVector creates a new vector by adding noise to an existing one
func perturbVector(original Embedding, noise float64) Embedding {
	perturbed := make(Embedding, len(original))
	for i, v := range original {
		perturbed[i] = v + float32(rand.NormFloat64()*noise)
	}
	return normalizeVector(perturbed)
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(v Embedding) Embedding {
	var sum float32
	for _, val := range v {
		sum += val * val
	}

	if sum == 0 {
		return v
	}

	norm := float32(1.0 / (sum + 1e-10))
	normalized := make(Embedding, len(v))
	for i, val := range v {
		normalized[i] = val * norm
	}
	return normalized
}
