package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WeaviateStore implements semantic memory using Weaviate vector database.
type WeaviateStore struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	className  string
}

// WeaviateConfig holds configuration for Weaviate connection.
type WeaviateConfig struct {
	Host      string `yaml:"host" json:"host"`
	Scheme    string `yaml:"scheme" json:"scheme"`
	ClassName string `yaml:"class_name" json:"class_name"`
	APIKey    string `yaml:"api_key" json:"api_key,omitempty"`
}

// NewWeaviateStore creates a new Weaviate-backed semantic store.
func NewWeaviateStore(ctx context.Context, config WeaviateConfig) (*WeaviateStore, error) {
	scheme := config.Scheme
	if scheme == "" {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, config.Host)

	store := &WeaviateStore{
		httpClient: &http.Client{Timeout: 30},
		baseURL:    baseURL,
		apiKey:     config.APIKey,
		className:  config.ClassName,
	}

	// Ensure class/schema exists
	if err := store.ensureClass(ctx); err != nil {
		return nil, fmt.Errorf("ensure class: %w", err)
	}

	return store, nil
}

// ensureClass creates the Weaviate class if it doesn't exist.
func (w *WeaviateStore) ensureClass(ctx context.Context) error {
	// Check if class exists
	req, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"/v1/schema/"+w.className, nil)
	if err != nil {
		return err
	}

	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Create class
	classSchema := map[string]interface{}{
		"class": w.className,
		"properties": []map[string]interface{}{
			{
				"name":     "agent",
				"dataType": []string{"text"},
			},
			{
				"name":     "key",
				"dataType": []string{"text"},
			},
			{
				"name":     "value",
				"dataType": []string{"text"},
			},
		},
		"vectorizer": "none",
	}

	body, err := json.Marshal(classSchema)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, "POST", w.baseURL+"/v1/schema", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err = w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create class failed: %s", body)
	}

	return nil
}

// Store adds a memory with embedding to Weaviate (custom interface).
func (w *WeaviateStore) Store(ctx context.Context, agentName, key string, value interface{}, embedding []float32) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	object := map[string]interface{}{
		"class": w.className,
		"properties": map[string]interface{}{
			"agent": agentName,
			"key":   key,
			"value": string(valueBytes),
		},
		"vector": embedding,
	}

	body, err := json.Marshal(object)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.baseURL+"/v1/objects", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("store failed: %s", body)
	}

	return nil
}

// Upsert implements SemanticStore interface.
func (w *WeaviateStore) Upsert(ctx context.Context, agentID string, id string, embedding Embedding, metadata map[string]any) error {
	return w.Store(ctx, agentID, id, metadata, embedding)
}

// Search implements SemanticStore.Search interface (uses Embedding type).
func (w *WeaviateStore) Search(ctx context.Context, agentID string, query Embedding, topK int) ([]SearchResult, error) {
	queryPayload := map[string]interface{}{
		"vector": query,
		"limit":  topK,
		"where": map[string]interface{}{
			"path":      []string{"agent"},
			"operator":  "Equal",
			"valueText": agentID,
		},
	}

	body, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/objects/%s", w.baseURL, w.className)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", body)
	}

	var response struct {
		Data struct {
			Objects []struct {
				ID         string                 `json:"id"`
				Properties map[string]interface{} `json:"properties"`
				Distance   float64                `json:"_additional"`
			} `json:"objects"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(response.Data.Objects))
	for _, obj := range response.Data.Objects {
		var value interface{}
		metadata := make(map[string]any)

		if valStr, ok := obj.Properties["value"].(string); ok {
			json.Unmarshal([]byte(valStr), &value)
			metadata["value"] = value
		}

		if key, ok := obj.Properties["key"].(string); ok {
			metadata["key"] = key
		}

		results = append(results, SearchResult{
			ID:       obj.ID,
			Score:    1.0 - obj.Distance,
			Metadata: metadata,
		})
	}

	return results, nil
}

// WeaviateResult is the legacy result format.
type WeaviateResult struct {
	Key   string
	Value interface{}
	Score float64
}

// Delete removes a memory from Weaviate.
func (w *WeaviateStore) Delete(ctx context.Context, agentName, key string) error {
	// Weaviate delete by filter
	filter := map[string]interface{}{
		"path":      []string{"key"},
		"operator":  "Equal",
		"valueText": key,
	}

	body, err := json.Marshal(map[string]interface{}{"where": filter})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/batch/objects", w.baseURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", body)
	}

	return nil
}

// DeleteAgent implements SemanticStore interface.
func (w *WeaviateStore) DeleteAgent(ctx context.Context, agentID string) error {
	filter := map[string]interface{}{
		"path":      []string{"agent"},
		"operator":  "Equal",
		"valueText": agentID,
	}

	body, err := json.Marshal(map[string]interface{}{"where": filter})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/batch/objects", w.baseURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if w.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+w.apiKey)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete agent failed: %s", body)
	}

	return nil
}

// Close closes the Weaviate connection.
func (w *WeaviateStore) Close() error {
	return nil
}
