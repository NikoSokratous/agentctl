package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingClient is the OpenAI embeddings API client.
type EmbeddingClient struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

// NewEmbeddingClient creates an OpenAI embedding client.
func NewEmbeddingClient(apiKey, model string) *EmbeddingClient {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &EmbeddingClient{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		Model:   model,
		HTTP:    http.DefaultClient,
	}
}

// Name implements llm.EmbeddingProvider.
func (c *EmbeddingClient) Name() string {
	return "openai-embeddings"
}

// Dimensions returns the embedding dimension size.
func (c *EmbeddingClient) Dimensions() int {
	// text-embedding-3-small has 1536 dimensions
	return 1536
}

// Embed implements llm.EmbeddingProvider.
func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	body := map[string]any{
		"model": c.Model,
		"input": texts,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/embeddings", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embeddings api error %d: %s", resp.StatusCode, string(msg))
	}

	var out embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(out.Data))
	for i, item := range out.Data {
		embeddings[i] = item.Embedding
	}

	return embeddings, nil
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}
