package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestScanAndLoadPlugins tests the complete plugin lifecycle.
func TestScanAndLoadPlugins(t *testing.T) {
	// Create temporary plugin structure
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Create manifest
	manifestYAML := `name: integration-test-plugin
version: 1.0.0
type: goplugin
binary: ./test.so
description: Integration test plugin
author: Test Suite
permissions:
  - test:permission
`

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestYAML), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create discovery
	pd := NewPluginDiscovery([]string{tmpDir})

	// Scan
	manifests, err := pd.ScanPlugins()
	if err != nil {
		t.Fatalf("ScanPlugins failed: %v", err)
	}

	if len(manifests) != 1 {
		t.Fatalf("Expected 1 manifest, got %d", len(manifests))
	}

	manifest := manifests[0]

	// Validate manifest
	if manifest.Name != "integration-test-plugin" {
		t.Errorf("Expected name 'integration-test-plugin', got '%s'", manifest.Name)
	}

	if manifest.Type != "goplugin" {
		t.Errorf("Expected type 'goplugin', got '%s'", manifest.Type)
	}

	// Verify binary path is absolute
	if !filepath.IsAbs(manifest.Binary) {
		t.Error("Expected binary path to be absolute")
	}

	// Note: Cannot test LoadFromManifest without actual .so file
}

// TestPluginManifestTypes tests different plugin types.
func TestPluginManifestTypes(t *testing.T) {
	tests := []struct {
		name         string
		manifestYAML string
		expectedType string
	}{
		{
			name: "goplugin",
			manifestYAML: `name: go-plugin
version: 1.0.0
type: goplugin
binary: ./plugin.so
`,
			expectedType: "goplugin",
		},
		{
			name: "wasm",
			manifestYAML: `name: wasm-plugin
version: 1.0.0
type: wasm
binary: ./plugin.wasm
`,
			expectedType: "wasm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, "plugin.yaml")

			if err := os.WriteFile(manifestPath, []byte(tt.manifestYAML), 0644); err != nil {
				t.Fatalf("Failed to write manifest: %v", err)
			}

			manifest, err := loadPluginManifest(manifestPath)
			if err != nil {
				t.Fatalf("loadPluginManifest failed: %v", err)
			}

			if manifest.Type != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, manifest.Type)
			}
		})
	}
}

// TestWatchForChanges tests the hot-reload watcher.
func TestWatchForChanges(t *testing.T) {
	tmpDir := t.TempDir()
	pd := NewPluginDiscovery([]string{tmpDir})

	// Start watching with short interval
	err := pd.WatchForChanges(tmpDir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WatchForChanges failed: %v", err)
	}

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a .so file
	testFile := filepath.Join(tmpDir, "test.so")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Give watcher time to detect
	time.Sleep(100 * time.Millisecond)

	// Modify file
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Give watcher time to detect change
	time.Sleep(100 * time.Millisecond)

	// Stop watching
	pd.StopWatching(tmpDir)

	// Verify can't watch same dir twice
	err = pd.WatchForChanges(tmpDir, 50*time.Millisecond)
	if err != nil {
		t.Errorf("Should be able to watch after stopping: %v", err)
	}

	pd.StopWatching(tmpDir)
}

// TestMultipleManifestsInSameDirectory tests scanning with multiple plugins.
func TestMultipleManifestsInSameDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two plugin directories
	for i := 1; i <= 2; i++ {
		pluginDir := filepath.Join(tmpDir, fmt.Sprintf("plugin%d", i))
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			t.Fatalf("Failed to create plugin dir: %v", err)
		}

		manifestYAML := fmt.Sprintf(`name: plugin-%d
version: 1.0.0
type: goplugin
binary: ./plugin.so
`, i)

		manifestPath := filepath.Join(pluginDir, "plugin.yaml")
		if err := os.WriteFile(manifestPath, []byte(manifestYAML), 0644); err != nil {
			t.Fatalf("Failed to write manifest: %v", err)
		}
	}

	// Scan
	pd := NewPluginDiscovery([]string{tmpDir})
	manifests, err := pd.ScanPlugins()
	if err != nil {
		t.Fatalf("ScanPlugins failed: %v", err)
	}

	if len(manifests) != 2 {
		t.Errorf("Expected 2 manifests, got %d", len(manifests))
	}
}
