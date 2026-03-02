package tool

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPluginDiscoveryNew(t *testing.T) {
	dirs := []string{"./plugins", "/usr/local/plugins"}
	pd := NewPluginDiscovery(dirs)

	if pd == nil {
		t.Fatal("NewPluginDiscovery returned nil")
	}

	if len(pd.pluginDirs) != 2 {
		t.Errorf("Expected 2 plugin dirs, got %d", len(pd.pluginDirs))
	}
}

func TestScanPlugins(t *testing.T) {
	// Create temporary plugin directory structure
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "plugins", "test-plugin")

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin dir: %v", err)
	}

	// Create plugin manifest
	manifestYAML := `name: test-plugin
version: 1.0.0
type: goplugin
binary: ./test.so
description: Test plugin
author: Test Author
permissions:
  - fs:read
  - net:external
`

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestYAML), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Scan plugins
	pd := NewPluginDiscovery([]string{filepath.Join(tmpDir, "plugins")})
	manifests, err := pd.ScanPlugins()
	if err != nil {
		t.Fatalf("ScanPlugins failed: %v", err)
	}

	if len(manifests) != 1 {
		t.Fatalf("Expected 1 manifest, got %d", len(manifests))
	}

	manifest := manifests[0]
	if manifest.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", manifest.Name)
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", manifest.Version)
	}

	if manifest.Type != "goplugin" {
		t.Errorf("Expected type 'goplugin', got '%s'", manifest.Type)
	}

	if len(manifest.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(manifest.Permissions))
	}
}

func TestLoadPluginManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "plugin.yaml")

	manifestYAML := `name: example-tool
version: 2.0.0
type: wasm
binary: ./example.wasm
description: Example WASM tool
author: AgentRuntime Team
permissions:
  - compute:cpu
  - memory:allocate
`

	if err := os.WriteFile(manifestPath, []byte(manifestYAML), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := loadPluginManifest(manifestPath)
	if err != nil {
		t.Fatalf("loadPluginManifest failed: %v", err)
	}

	if manifest.Name != "example-tool" {
		t.Errorf("Expected name 'example-tool', got '%s'", manifest.Name)
	}

	if manifest.Type != "wasm" {
		t.Errorf("Expected type 'wasm', got '%s'", manifest.Type)
	}

	// Binary path should be resolved to absolute path
	if !filepath.IsAbs(manifest.Binary) {
		t.Error("Expected binary path to be absolute")
	}
}

func TestScanPluginsEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	pd := NewPluginDiscovery([]string{tmpDir})
	manifests, err := pd.ScanPlugins()

	if err != nil {
		t.Fatalf("ScanPlugins failed: %v", err)
	}

	if len(manifests) != 0 {
		t.Errorf("Expected 0 manifests in empty dir, got %d", len(manifests))
	}
}

func TestScanPluginsNonExistentDirectory(t *testing.T) {
	pd := NewPluginDiscovery([]string{"/nonexistent/path"})
	manifests, err := pd.ScanPlugins()

	// Should not error, just return empty
	if err != nil {
		t.Fatalf("ScanPlugins should not error on non-existent dir: %v", err)
	}

	if len(manifests) != 0 {
		t.Errorf("Expected 0 manifests, got %d", len(manifests))
	}
}

func TestPluginDiscoveryWatching(t *testing.T) {
	tmpDir := t.TempDir()
	pd := NewPluginDiscovery([]string{tmpDir})

	// Start watching
	err := pd.WatchForChanges(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("WatchForChanges failed: %v", err)
	}

	// Try to watch same directory again (should error)
	err = pd.WatchForChanges(tmpDir, 100*time.Millisecond)
	if err == nil {
		t.Error("Expected error when watching same directory twice")
	}

	// Stop watching
	pd.StopWatching(tmpDir)

	// Should be able to watch again after stopping
	err = pd.WatchForChanges(tmpDir, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Should be able to watch after stopping: %v", err)
	}

	pd.StopWatching(tmpDir)
}

func TestListAllPlugins(t *testing.T) {
	pd := NewPluginDiscovery([]string{})

	plugins := pd.ListAllPlugins()
	if plugins == nil {
		t.Error("ListAllPlugins should not return nil")
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}
