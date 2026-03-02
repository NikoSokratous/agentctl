package registry

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// PluginWatcher monitors plugin directories for changes
type PluginWatcher struct {
	registry      *Registry
	watcher       *fsnotify.Watcher
	pluginDirs    []string
	loadedPlugins map[string]*PluginInfo
	mu            sync.RWMutex
	stopChan      chan struct{}
}

// PluginInfo tracks loaded plugin information
type PluginInfo struct {
	Metadata    *PluginMetadata
	LoadedAt    time.Time
	FilePath    string
	FileModTime time.Time
}

// NewPluginWatcher creates a new plugin watcher
func NewPluginWatcher(registry *Registry, pluginDirs []string) (*PluginWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	pw := &PluginWatcher{
		registry:      registry,
		watcher:       watcher,
		pluginDirs:    pluginDirs,
		loadedPlugins: make(map[string]*PluginInfo),
		stopChan:      make(chan struct{}),
	}

	// Add directories to watch
	for _, dir := range pluginDirs {
		if err := watcher.Add(dir); err != nil {
			log.Printf("Failed to watch directory %s: %v", dir, err)
			continue
		}
		log.Printf("Watching plugin directory: %s", dir)
	}

	return pw, nil
}

// Start starts watching for changes
func (pw *PluginWatcher) Start(ctx context.Context) error {
	// Initial load of all plugins
	if err := pw.loadAllPlugins(ctx); err != nil {
		return fmt.Errorf("initial load: %w", err)
	}

	go pw.watch(ctx)
	return nil
}

// Stop stops the watcher
func (pw *PluginWatcher) Stop() {
	close(pw.stopChan)
	pw.watcher.Close()
}

// watch monitors file system events
func (pw *PluginWatcher) watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pw.stopChan:
			return
		case event, ok := <-pw.watcher.Events:
			if !ok {
				return
			}

			// Only handle .yaml or .yml files
			if filepath.Ext(event.Name) != ".yaml" && filepath.Ext(event.Name) != ".yml" {
				continue
			}

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write:
				log.Printf("Plugin modified: %s", event.Name)
				pw.reloadPlugin(ctx, event.Name)
			case event.Op&fsnotify.Create == fsnotify.Create:
				log.Printf("Plugin added: %s", event.Name)
				pw.loadPlugin(ctx, event.Name)
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				log.Printf("Plugin removed: %s", event.Name)
				pw.unloadPlugin(ctx, event.Name)
			}

		case err, ok := <-pw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// loadAllPlugins loads all plugins from watched directories
func (pw *PluginWatcher) loadAllPlugins(ctx context.Context) error {
	for _, dir := range pw.pluginDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() || (filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml") {
				return nil
			}

			return pw.loadPlugin(ctx, path)
		})

		if err != nil {
			log.Printf("Failed to load plugins from %s: %v", dir, err)
		}
	}

	return nil
}

// loadPlugin loads a single plugin
func (pw *PluginWatcher) loadPlugin(ctx context.Context, path string) error {
	// Read plugin metadata
	_, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plugin file: %w", err)
	}

	// Parse metadata (simplified - in real implementation, use YAML parser)
	// For now, we'll create a minimal metadata structure
	metadata := &PluginMetadata{
		ID:      filepath.Base(path),
		Name:    filepath.Base(path),
		Version: "1.0.0",
	}

	// Get file mod time
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	pw.mu.Lock()
	pw.loadedPlugins[path] = &PluginInfo{
		Metadata:    metadata,
		LoadedAt:    time.Now(),
		FilePath:    path,
		FileModTime: info.ModTime(),
	}
	pw.mu.Unlock()

	log.Printf("Loaded plugin: %s (version %s)", metadata.Name, metadata.Version)
	return nil
}

// reloadPlugin reloads a modified plugin
func (pw *PluginWatcher) reloadPlugin(ctx context.Context, path string) error {
	pw.mu.RLock()
	oldPlugin, exists := pw.loadedPlugins[path]
	pw.mu.RUnlock()

	if !exists {
		return pw.loadPlugin(ctx, path)
	}

	// Check if file actually changed
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	if !info.ModTime().After(oldPlugin.FileModTime) {
		return nil // No change
	}

	log.Printf("Reloading plugin: %s", path)

	// Graceful unload
	if err := pw.unloadPlugin(ctx, path); err != nil {
		log.Printf("Warning: Failed to unload plugin %s: %v", path, err)
	}

	// Small delay to ensure file operations complete
	time.Sleep(100 * time.Millisecond)

	// Load new version
	return pw.loadPlugin(ctx, path)
}

// unloadPlugin unloads a plugin
func (pw *PluginWatcher) unloadPlugin(ctx context.Context, path string) error {
	pw.mu.Lock()
	plugin, exists := pw.loadedPlugins[path]
	if !exists {
		pw.mu.Unlock()
		return nil
	}

	delete(pw.loadedPlugins, path)
	pw.mu.Unlock()

	log.Printf("Unloaded plugin: %s", plugin.Metadata.Name)
	return nil
}

// GetLoadedPlugins returns currently loaded plugins
func (pw *PluginWatcher) GetLoadedPlugins() map[string]*PluginInfo {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	// Return a copy
	result := make(map[string]*PluginInfo)
	for k, v := range pw.loadedPlugins {
		result[k] = v
	}
	return result
}

// GetPluginInfo returns info about a specific plugin
func (pw *PluginWatcher) GetPluginInfo(path string) (*PluginInfo, bool) {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	info, exists := pw.loadedPlugins[path]
	return info, exists
}
