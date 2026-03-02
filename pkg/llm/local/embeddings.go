package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// EmbeddingClient uses sentence-transformers via Python for local embeddings.
type EmbeddingClient struct {
	ModelName  string
	dimensions int
	PythonPath string
}

// NewEmbeddingClient creates a local embedding client.
func NewEmbeddingClient(modelName string) *EmbeddingClient {
	if modelName == "" {
		modelName = "all-MiniLM-L6-v2"
	}
	return &EmbeddingClient{
		ModelName:  modelName,
		dimensions: 384, // all-MiniLM-L6-v2 has 384 dimensions
		PythonPath: "python",
	}
}

// Name implements llm.EmbeddingProvider.
func (c *EmbeddingClient) Name() string {
	return "local-embeddings"
}

// Dimensions returns the embedding dimension size.
func (c *EmbeddingClient) Dimensions() int {
	return c.dimensions
}

// Embed implements llm.EmbeddingProvider using Python sentence-transformers.
func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Prepare JSON input for Python script
	input := map[string]any{
		"model": c.ModelName,
		"texts": texts,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	// Call Python helper script
	cmd := exec.CommandContext(ctx, c.PythonPath, "scripts/embedding_server.py")
	cmd.Stdin = strings.NewReader(string(inputJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run embedding script: %w (output: %s)", err, string(output))
	}

	// Parse output
	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
		Error      string      `json:"error"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse embedding output: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("embedding generation error: %s", result.Error)
	}

	return result.Embeddings, nil
}
