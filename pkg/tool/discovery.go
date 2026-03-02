package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PluginDiscovery handles plugin discovery and hot-reload.
type PluginDiscovery struct {
	pluginDirs []string
	goLoader   *PluginLoader
	wasmLoader *WASMLoader
	watchers   map[string]chan struct{}
	mu         sync.RWMutex
}

// PluginManifest describes a plugin's metadata.
type PluginManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Type        string   `yaml:"type"` // "goplugin" or "wasm"
	Binary      string   `yaml:"binary"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Permissions []string `yaml:"permissions"`
}

// NewPluginDiscovery creates a new plugin discovery manager.
func NewPluginDiscovery(pluginDirs []string) *PluginDiscovery {
	return &PluginDiscovery{
		pluginDirs: pluginDirs,
		goLoader:   NewPluginLoader(),
		wasmLoader: NewWASMLoader(),
		watchers:   make(map[string]chan struct{}),
	}
}

// ScanPlugins scans plugin directories for available plugins.
func (pd *PluginDiscovery) ScanPlugins() ([]PluginManifest, error) {
	var manifests []PluginManifest

	for _, dir := range pd.pluginDirs {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Walk directory
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Look for plugin.yaml or manifest.yaml
			if info.IsDir() || (info.Name() != "plugin.yaml" && info.Name() != "manifest.yaml") {
				return nil
			}

			// Load manifest
			manifest, err := loadPluginManifest(path)
			if err != nil {
				return fmt.Errorf("load manifest %s: %w", path, err)
			}

			manifests = append(manifests, *manifest)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("scan directory %s: %w", dir, err)
		}
	}

	return manifests, nil
}

// loadPluginManifest loads a plugin manifest file.
func loadPluginManifest(path string) (*PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	// Resolve binary path relative to manifest
	if !filepath.IsAbs(manifest.Binary) {
		manifestDir := filepath.Dir(path)
		manifest.Binary = filepath.Join(manifestDir, manifest.Binary)
	}

	return &manifest, nil
}

// LoadFromManifest loads a plugin from a manifest.
func (pd *PluginDiscovery) LoadFromManifest(manifest PluginManifest) error {
	switch strings.ToLower(manifest.Type) {
	case "goplugin":
		return pd.goLoader.LoadPlugin(manifest.Binary)
	case "wasm":
		// For WASM, we need metadata
		metadata := WASMToolMetadata{
			Name:        manifest.Name,
			Version:     manifest.Version,
			InputSchema: []byte("{}"),
			Permissions: []Permission{},
		}
		for _, p := range manifest.Permissions {
			perm := Permission{
				Scope:    p,
				Required: true,
			}
			metadata.Permissions = append(metadata.Permissions, perm)
		}
		return pd.wasmLoader.LoadWASM(context.Background(), manifest.Binary, metadata)
	default:
		return fmt.Errorf("unsupported plugin type: %s", manifest.Type)
	}
}

// WatchForChanges enables hot-reload for a plugin directory.
func (pd *PluginDiscovery) WatchForChanges(dir string, reloadInterval time.Duration) error {
	pd.mu.Lock()
	if _, exists := pd.watchers[dir]; exists {
		pd.mu.Unlock()
		return fmt.Errorf("already watching directory: %s", dir)
	}

	stopChan := make(chan struct{})
	pd.watchers[dir] = stopChan
	pd.mu.Unlock()

	// Launch watcher goroutine
	go func() {
		ticker := time.NewTicker(reloadInterval)
		defer ticker.Stop()

		// Track file modification times
		modTimes := make(map[string]time.Time)

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				// Scan for changes
				filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() {
						return nil
					}

					// Check plugin files
					if strings.HasSuffix(path, ".so") || strings.HasSuffix(path, ".wasm") {
						lastMod, exists := modTimes[path]
						if !exists || info.ModTime().After(lastMod) {
							// File changed, reload
							modTimes[path] = info.ModTime()
							if exists {
								// Hot reload: unload old plugin and reload new one
								pd.mu.Unlock() // Unlock temporarily

								// Note: Hot reload for Go plugins is limited due to Go's plugin system
								// For production, consider using WASM plugins or external processes
								fmt.Printf("Plugin changed: %s (restart required for Go plugins)\n", path)

								pd.mu.Lock() // Re-lock
							}
						}
					}

					return nil
				})
			}
		}
	}()

	return nil
}

// StopWatching stops watching a directory.
func (pd *PluginDiscovery) StopWatching(dir string) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if stopChan, exists := pd.watchers[dir]; exists {
		close(stopChan)
		delete(pd.watchers, dir)
	}
}

// GetGoPlugin retrieves a Go plugin.
func (pd *PluginDiscovery) GetGoPlugin(name, version string) (*PluginTool, error) {
	return pd.goLoader.GetPlugin(name, version)
}

// GetWASMPlugin retrieves a WASM plugin.
func (pd *PluginDiscovery) GetWASMPlugin(name, version string) (*WASMTool, error) {
	return pd.wasmLoader.GetWASMTool(name, version)
}

// ListAllPlugins returns all loaded plugins.
func (pd *PluginDiscovery) ListAllPlugins() []string {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	plugins := make([]string, 0)
	plugins = append(plugins, pd.goLoader.ListPlugins()...)
	plugins = append(plugins, pd.wasmLoader.ListWASMTools()...)
	return plugins
}
