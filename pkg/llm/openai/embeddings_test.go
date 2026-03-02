package openai

import (
	"context"
	"os"
	"testing"
)

func TestNewEmbeddingClient(t *testing.T) {
	client := NewEmbeddingClient("test-key", "")

	if client.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", client.APIKey)
	}

	if client.Model != "text-embedding-3-small" {
		t.Errorf("expected default model 'text-embedding-3-small', got '%s'", client.Model)
	}
}

func TestOpenAIEmbeddingsIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	client := NewEmbeddingClient(apiKey, "text-embedding-3-small")

	embeddings, err := client.Embed(context.Background(), []string{"test"})
	if err != nil {
		t.Fatalf("failed to generate embeddings: %v", err)
	}

	if len(embeddings) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(embeddings))
	}

	if len(embeddings[0]) != 1536 {
		t.Fatalf("expected 1536 dimensions, got %d", len(embeddings[0]))
	}
}
