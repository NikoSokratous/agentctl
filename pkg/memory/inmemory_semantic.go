package memory

import (
	"context"
	"math"
	"sync"
)

// InMemorySemanticStore is a simple in-memory vector store with cosine similarity.
// For production, consider using HNSW library or external vector DB.
type InMemorySemanticStore struct {
	mu      sync.RWMutex
	vectors map[string]map[string]*vectorEntry // agentID -> id -> entry
}

type vectorEntry struct {
	ID        string
	Embedding Embedding
	Metadata  map[string]any
}

// NewInMemorySemanticStore creates an in-memory semantic store.
func NewInMemorySemanticStore() *InMemorySemanticStore {
	return &InMemorySemanticStore{
		vectors: make(map[string]map[string]*vectorEntry),
	}
}

// Upsert adds or updates a vector.
func (s *InMemorySemanticStore) Upsert(ctx context.Context, agentID string, id string, embedding Embedding, metadata map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vectors[agentID] == nil {
		s.vectors[agentID] = make(map[string]*vectorEntry)
	}

	s.vectors[agentID][id] = &vectorEntry{
		ID:        id,
		Embedding: embedding,
		Metadata:  metadata,
	}
	return nil
}

// Search finds the topK most similar vectors using cosine similarity.
func (s *InMemorySemanticStore) Search(ctx context.Context, agentID string, query Embedding, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agentVectors := s.vectors[agentID]
	if len(agentVectors) == 0 {
		return []SearchResult{}, nil
	}

	// Calculate similarities
	type scoredEntry struct {
		entry *vectorEntry
		score float64
	}

	var scored []scoredEntry
	for _, entry := range agentVectors {
		score := cosineSimilarity(query, entry.Embedding)
		scored = append(scored, scoredEntry{entry: entry, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Take topK
	if topK > len(scored) {
		topK = len(scored)
	}

	results := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = SearchResult{
			ID:       scored[i].entry.ID,
			Score:    scored[i].score,
			Metadata: scored[i].entry.Metadata,
		}
	}

	return results, nil
}

// DeleteAgent removes all vectors for an agent.
func (s *InMemorySemanticStore) DeleteAgent(ctx context.Context, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vectors, agentID)
	return nil
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b Embedding) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Stats returns statistics about the store.
func (s *InMemorySemanticStore) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for agentID, vectors := range s.vectors {
		stats[agentID] = len(vectors)
	}
	return stats
}
