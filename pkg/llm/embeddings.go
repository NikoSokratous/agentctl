package llm

import "context"

// EmbeddingProvider generates vector embeddings from text.
type EmbeddingProvider interface {
	Name() string
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
}

// EmbeddingRequest contains text to embed.
type EmbeddingRequest struct {
	Texts []string
	Model string // optional model override
}

// EmbeddingResponse contains generated embeddings.
type EmbeddingResponse struct {
	Embeddings [][]float32
	Model      string
	Usage      EmbeddingUsage
}

// EmbeddingUsage tracks token usage for embedding generation.
type EmbeddingUsage struct {
	PromptTokens int
	TotalTokens  int
}
