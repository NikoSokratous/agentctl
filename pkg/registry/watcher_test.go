package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPluginWatcher(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "plugin-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create watcher
	registry := NewRegistry(nil, "", "") // nil DB for testing
	watcher, err := NewPluginWatcher(registry, []string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a test plugin file
	pluginFile := filepath.Join(tmpDir, "test-plugin.yaml")
	content := []byte("name: test-plugin\nversion: 1.0.0\n")
	if err := os.WriteFile(pluginFile, content, 0644); err != nil {
		t.Fatalf("Failed to write plugin file: %v", err)
	}

	// Give watcher time to detect the file
	time.Sleep(200 * time.Millisecond)

	// Check if plugin was loaded
	loaded := watcher.GetLoadedPlugins()
	if len(loaded) != 1 {
		t.Errorf("Expected 1 loaded plugin, got %d", len(loaded))
	}

	// Test hot-reload: modify the file
	newContent := []byte("name: test-plugin\nversion: 1.1.0\n")
	if err := os.WriteFile(pluginFile, newContent, 0644); err != nil {
		t.Fatalf("Failed to update plugin file: %v", err)
	}

	// Give watcher time to detect the change
	time.Sleep(200 * time.Millisecond)

	// Plugin should still be loaded (reloaded)
	loaded = watcher.GetLoadedPlugins()
	if len(loaded) != 1 {
		t.Errorf("Expected 1 loaded plugin after reload, got %d", len(loaded))
	}

	// Test removal
	if err := os.Remove(pluginFile); err != nil {
		t.Fatalf("Failed to remove plugin file: %v", err)
	}

	// Give watcher time to detect the removal
	time.Sleep(200 * time.Millisecond)

	// Plugin should be unloaded
	loaded = watcher.GetLoadedPlugins()
	if len(loaded) != 0 {
		t.Errorf("Expected 0 loaded plugins after removal, got %d", len(loaded))
	}
}

func TestPluginInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-info-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRegistry(nil, "", "")
	watcher, err := NewPluginWatcher(registry, []string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	ctx := context.Background()
	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	pluginFile := filepath.Join(tmpDir, "info-test.yaml")
	if err := os.WriteFile(pluginFile, []byte("test: data\n"), 0644); err != nil {
		t.Fatalf("Failed to write plugin file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	info, exists := watcher.GetPluginInfo(pluginFile)
	if !exists {
		t.Error("Plugin info should exist")
	}

	if info.FilePath != pluginFile {
		t.Errorf("Expected file path %s, got %s", pluginFile, info.FilePath)
	}

	if info.Metadata.Name != "info-test.yaml" {
		t.Errorf("Expected name info-test.yaml, got %s", info.Metadata.Name)
	}
}
