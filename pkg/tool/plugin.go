package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"plugin"
	"sync"
)

// PluginTool wraps a dynamically loaded Go plugin.
type PluginTool struct {
	name       string
	version    string
	pluginPath string
	plugin     *plugin.Plugin
	toolImpl   Tool
	mu         sync.RWMutex
}

// NewPluginTool loads a tool from a Go plugin file.
func NewPluginTool(pluginPath string) (*PluginTool, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	// Look for NewTool symbol
	sym, err := p.Lookup("NewTool")
	if err != nil {
		return nil, fmt.Errorf("plugin must export NewTool function: %w", err)
	}

	// Type assert to tool constructor
	newToolFn, ok := sym.(func() Tool)
	if !ok {
		return nil, fmt.Errorf("NewTool must be func() Tool")
	}

	// Create tool instance
	toolImpl := newToolFn()

	return &PluginTool{
		name:       toolImpl.Name(),
		version:    toolImpl.Version(),
		pluginPath: pluginPath,
		plugin:     p,
		toolImpl:   toolImpl,
	}, nil
}

// Name returns the plugin tool name.
func (p *PluginTool) Name() string {
	return p.name
}

// Version returns the plugin tool version.
func (p *PluginTool) Version() string {
	return p.version
}

// Description returns the tool description.
func (p *PluginTool) Description() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.toolImpl.Description()
}

// InputSchema returns the tool's input schema.
func (p *PluginTool) InputSchema() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.toolImpl.InputSchema()
}

// Permissions returns the tool's permission requirements.
func (p *PluginTool) Permissions() []Permission {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.toolImpl.Permissions()
}

// Execute runs the plugin tool.
func (p *PluginTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.toolImpl.Execute(ctx, input)
}

// PluginLoader manages loading and hot-reloading of plugins.
type PluginLoader struct {
	plugins map[string]*PluginTool
	mu      sync.RWMutex
}

// NewPluginLoader creates a new plugin loader.
func NewPluginLoader() *PluginLoader {
	return &PluginLoader{
		plugins: make(map[string]*PluginTool),
	}
}

// LoadPlugin loads a plugin from a file.
func (l *PluginLoader) LoadPlugin(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	pluginTool, err := NewPluginTool(path)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s@%s", pluginTool.Name(), pluginTool.Version())
	l.plugins[key] = pluginTool

	return nil
}

// GetPlugin retrieves a loaded plugin.
func (l *PluginLoader) GetPlugin(name, version string) (*PluginTool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	key := fmt.Sprintf("%s@%s", name, version)
	pluginTool, ok := l.plugins[key]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", key)
	}

	return pluginTool, nil
}

// ListPlugins returns all loaded plugins.
func (l *PluginLoader) ListPlugins() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	names := make([]string, 0, len(l.plugins))
	for key := range l.plugins {
		names = append(names, key)
	}
	return names
}

// UnloadPlugin removes a plugin from memory.
func (l *PluginLoader) UnloadPlugin(name, version string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := fmt.Sprintf("%s@%s", name, version)
	if _, ok := l.plugins[key]; !ok {
		return fmt.Errorf("plugin %s not found", key)
	}

	delete(l.plugins, key)
	return nil
}

// ReloadPlugin hot-reloads a plugin.
func (l *PluginLoader) ReloadPlugin(path, name, version string) error {
	// Unload old version
	key := fmt.Sprintf("%s@%s", name, version)
	l.mu.Lock()
	delete(l.plugins, key)
	l.mu.Unlock()

	// Load new version
	return l.LoadPlugin(path)
}
