package memory

import (
	"context"
)

// Embedding is a vector representation.
type Embedding []float32

// SemanticStore is the interface for vector similarity search.
type SemanticStore interface {
	Upsert(ctx context.Context, agentID string, id string, embedding Embedding, metadata map[string]any) error
	Search(ctx context.Context, agentID string, query Embedding, topK int) ([]SearchResult, error)
	DeleteAgent(ctx context.Context, agentID string) error
}

// SearchResult is a single search hit.
type SearchResult struct {
	ID       string
	Score    float64
	Metadata map[string]any
}
