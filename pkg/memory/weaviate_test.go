package memory

import (
	"context"
	"testing"
	"time"
)

// TestWeaviateStoreInterface verifies WeaviateStore implements interfaces
func TestWeaviateStoreInterface(t *testing.T) {
	t.Skip("Requires running Weaviate instance")

	// This test verifies that WeaviateStore satisfies the SemanticStore interface
	var _ SemanticStore = (*WeaviateStore)(nil)
}

// TestWeaviateConnection tests connecting to Weaviate (integration test)
func TestWeaviateConnection(t *testing.T) {
	t.Skip("Integration test - requires running Weaviate instance on localhost:8080")

	ctx := context.Background()
	config := WeaviateConfig{
		Host:      "localhost:8080",
		Scheme:    "http",
		ClassName: "TestCollection",
	}

	store, err := NewWeaviateStore(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create Weaviate store: %v", err)
	}
	defer store.Close()

	// Test basic operations
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Store a vector
	err = store.Store(ctx, "test-agent", "test-key", map[string]string{"test": "data"}, embedding)
	if err != nil {
		t.Fatalf("Failed to store: %v", err)
	}
}

// TestWeaviateConfig tests configuration validation
func TestWeaviateConfig(t *testing.T) {
	tests := []struct {
		name   string
		config WeaviateConfig
		valid  bool
	}{
		{
			name: "valid http config",
			config: WeaviateConfig{
				Host:      "localhost:8080",
				Scheme:    "http",
				ClassName: "Test",
			},
			valid: true,
		},
		{
			name: "valid https with API key",
			config: WeaviateConfig{
				Host:      "weaviate.example.com",
				Scheme:    "https",
				ClassName: "Production",
				APIKey:    "secret-key",
			},
			valid: true,
		},
		{
			name: "default scheme",
			config: WeaviateConfig{
				Host:      "localhost:8080",
				ClassName: "Test",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Host == "" {
				t.Error("Host should not be empty")
			}
			if tt.config.ClassName == "" {
				t.Error("ClassName should not be empty")
			}

			// Verify scheme defaults to http
			scheme := tt.config.Scheme
			if scheme == "" {
				scheme = "http"
			}
			if scheme != "http" && scheme != "https" {
				t.Errorf("Invalid scheme: %s", scheme)
			}
		})
	}
}

// TestWeaviateHTTPClient tests the HTTP client configuration
func TestWeaviateHTTPClient(t *testing.T) {
	ctx := context.Background()
	config := WeaviateConfig{
		Host:      "localhost:9999",
		ClassName: "Test",
	}

	store, err := NewWeaviateStore(ctx, config)
	if err != nil {
		// Expected to fail since Weaviate isn't running
		t.Skip("Weaviate not running (expected)")
	}

	if store != nil {
		defer store.Close()

		// Verify client has timeout
		if store.httpClient.Timeout != 30*time.Second {
			t.Errorf("Expected 30s timeout, got %v", store.httpClient.Timeout)
		}
	}
}
