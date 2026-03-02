package llm

import "context"

// Provider is the interface for LLM backends.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "anthropic", "ollama").
	Name() string

	// Chat sends messages and returns a completion.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
