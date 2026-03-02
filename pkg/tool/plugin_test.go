package tool

import (
	"testing"
)

// TestPluginLoader tests the plugin loader functionality.
func TestPluginLoader(t *testing.T) {
	loader := NewPluginLoader()

	if loader == nil {
		t.Fatal("NewPluginLoader returned nil")
	}

	// Test empty list
	plugins := loader.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}

// TestPluginLoaderGetNotFound tests retrieving non-existent plugin.
func TestPluginLoaderGetNotFound(t *testing.T) {
	loader := NewPluginLoader()

	_, err := loader.GetPlugin("nonexistent", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}

	expectedMsg := "plugin nonexistent@1.0.0 not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestPluginLoaderUnloadNotFound tests unloading non-existent plugin.
func TestPluginLoaderUnloadNotFound(t *testing.T) {
	loader := NewPluginLoader()

	err := loader.UnloadPlugin("nonexistent", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
}

// TestWASMLoader tests the WASM loader functionality.
func TestWASMLoader(t *testing.T) {
	loader := NewWASMLoader()

	if loader == nil {
		t.Fatal("NewWASMLoader returned nil")
	}

	// Test GetWASMTool on empty loader
	_, err := loader.GetWASMTool("test", "1.0.0")
	if err == nil {
		t.Error("Expected error for non-existent WASM tool")
	}
}

// TestWASMToolMetadata tests WASM tool metadata structure.
func TestWASMToolMetadata(t *testing.T) {
	metadata := WASMToolMetadata{
		Name:        "test-tool",
		Version:     "1.0.0",
		InputSchema: []byte(`{"type":"object"}`),
		Permissions: []Permission{
			{Scope: "compute:cpu", Required: true},
		},
	}

	if metadata.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", metadata.Name)
	}

	if len(metadata.Permissions) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(metadata.Permissions))
	}
}

// Note: Full plugin loading tests require actual .so files (platform-specific)
// and .wasm files, which are built separately. These tests validate the structure.
