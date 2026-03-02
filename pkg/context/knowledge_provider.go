package context

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// KnowledgeProvider retrieves from external knowledge bases (RAG).
type KnowledgeProvider struct {
	priority       int
	Sources        []string
	Enabled        bool
	KnowledgeStore *KnowledgeStore
	TopK           int
}

// NewKnowledgeProvider creates a new knowledge provider.
func NewKnowledgeProvider(priority int, sources []string) *KnowledgeProvider {
	return &KnowledgeProvider{
		priority: priority,
		Sources:  sources,
		Enabled:  false, // Disabled by default until knowledge store is configured
		TopK:     5,
	}
}

// Name returns the provider name.
func (p *KnowledgeProvider) Name() string {
	return "knowledge"
}

// Priority returns the provider priority.
func (p *KnowledgeProvider) Priority() int {
	return p.priority
}

// Fetch retrieves knowledge context using RAG.
func (p *KnowledgeProvider) Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error) {
	if !p.Enabled || p.KnowledgeStore == nil {
		return nil, nil
	}

	if input.StepInput.Goal == "" {
		return nil, nil
	}

	// Search knowledge base
	chunks, err := p.KnowledgeStore.Search(ctx, input.StepInput.Goal, p.TopK)
	if err != nil || len(chunks) == 0 {
		return nil, err
	}

	// Format chunks as context
	content := p.formatKnowledgeChunks(chunks)
	tokenCount := len(content) / 4

	return &ContextFragment{
		ProviderName: p.Name(),
		Type:         FragmentTypeKnowledge,
		Content:      content,
		Priority:     p.priority,
		TokenCount:   tokenCount,
		Metadata: map[string]any{
			"sources": p.extractSources(chunks),
			"chunks":  len(chunks),
		},
		Timestamp: time.Now(),
	}, nil
}

// SetEnabled enables or disables the knowledge provider.
func (p *KnowledgeProvider) SetEnabled(enabled bool) {
	p.Enabled = enabled
}

// AddSource adds a knowledge source.
func (p *KnowledgeProvider) AddSource(source string) {
	p.Sources = append(p.Sources, source)
}

// formatKnowledgeChunks formats document chunks as context.
func (p *KnowledgeProvider) formatKnowledgeChunks(chunks []DocumentChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	parts := []string{"Relevant knowledge:"}
	for i, chunk := range chunks {
		source := "unknown"
		if s, ok := chunk.Metadata["source"].(string); ok {
			source = s
		}
		fileName := "unknown"
		if f, ok := chunk.Metadata["file_path"].(string); ok {
			fileName = f
		}

		parts = append(parts, fmt.Sprintf("\n%d. [Source: %s, File: %s]\n%s",
			i+1, source, fileName, chunk.Content))
	}

	return strings.Join(parts, "\n")
}

// extractSources extracts unique sources from chunks.
func (p *KnowledgeProvider) extractSources(chunks []DocumentChunk) []string {
	sourceMap := make(map[string]bool)
	for _, chunk := range chunks {
		if s, ok := chunk.Metadata["source"].(string); ok {
			sourceMap[s] = true
		}
	}

	sources := make([]string, 0, len(sourceMap))
	for s := range sourceMap {
		sources = append(sources, s)
	}
	return sources
}
