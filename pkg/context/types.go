package context

import "time"

// TokenCounter estimates token count for text.
type TokenCounter interface {
	Count(text string) int
}

// SimpleTokenCounter is a basic token counter (approximation).
type SimpleTokenCounter struct{}

// Count estimates tokens using a simple heuristic (1 token ~= 4 chars).
func (c *SimpleTokenCounter) Count(text string) int {
	// Rough estimate: 1 token ≈ 4 characters
	return len(text) / 4
}

// ContextMetadata tracks context assembly details for observability.
type ContextMetadata struct {
	Fragments        []FragmentMetadata
	TotalTokens      int
	AssemblyDuration time.Duration
	Truncated        bool
	TruncatedCount   int
}

// FragmentMetadata describes a context fragment's metadata.
type FragmentMetadata struct {
	ProviderName  string
	Type          FragmentType
	Priority      int
	TokenCount    int
	Included      bool
	TruncatedTo   int // If truncated, new token count
	FetchDuration time.Duration
}

// ProviderConfig is a generic configuration for providers.
type ProviderConfig struct {
	Enabled  bool
	Priority int
	Config   map[string]any
}
