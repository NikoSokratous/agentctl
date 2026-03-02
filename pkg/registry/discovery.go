package registry

import (
	"fmt"
	"strings"
)

// DiscoveryService discovers plugins from various sources.
type DiscoveryService struct {
	registry *Registry
	sources  []PluginSource
}

// PluginSource defines a plugin source.
type PluginSource interface {
	List() ([]PluginMetadata, error)
	Download(pluginID, version string) ([]byte, error)
}

// NewDiscoveryService creates a new discovery service.
func NewDiscoveryService(registry *Registry) *DiscoveryService {
	return &DiscoveryService{
		registry: registry,
		sources:  make([]PluginSource, 0),
	}
}

// AddSource adds a plugin source.
func (d *DiscoveryService) AddSource(source PluginSource) {
	d.sources = append(d.sources, source)
}

// Discover discovers plugins from all sources.
func (d *DiscoveryService) Discover() ([]PluginMetadata, error) {
	allPlugins := make([]PluginMetadata, 0)

	for _, source := range d.sources {
		plugins, err := source.List()
		if err != nil {
			continue // Skip failed sources
		}
		allPlugins = append(allPlugins, plugins...)
	}

	// Deduplicate by ID
	seen := make(map[string]bool)
	unique := make([]PluginMetadata, 0)

	for _, p := range allPlugins {
		key := fmt.Sprintf("%s@%s", p.ID, p.Version)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, p)
		}
	}

	return unique, nil
}

// SearchByTags searches plugins by tags.
func (d *DiscoveryService) SearchByTags(tags []string) ([]PluginMetadata, error) {
	// Placeholder: would implement tag-based search
	return []PluginMetadata{}, nil
}

// GetRecommendations returns recommended plugins.
func (d *DiscoveryService) GetRecommendations(userID string, limit int) ([]PluginMetadata, error) {
	// Placeholder: would implement recommendation algorithm
	// based on user history, ratings, and popularity
	return []PluginMetadata{}, nil
}

// GetTrending returns trending plugins.
func (d *DiscoveryService) GetTrending(limit int) ([]PluginMetadata, error) {
	// Placeholder: would calculate trending based on
	// recent downloads, ratings, and activity
	return []PluginMetadata{}, nil
}

// LocalSource is a local filesystem plugin source.
type LocalSource struct {
	path string
}

// NewLocalSource creates a local plugin source.
func NewLocalSource(path string) *LocalSource {
	return &LocalSource{path: path}
}

// List lists plugins from local source.
func (l *LocalSource) List() ([]PluginMetadata, error) {
	// Placeholder: would scan local directory for plugins
	return []PluginMetadata{}, nil
}

// Download downloads plugin from local source.
func (l *LocalSource) Download(pluginID, version string) ([]byte, error) {
	// Placeholder: would read plugin from local filesystem
	return nil, fmt.Errorf("not implemented")
}

// RemoteSource is a remote registry plugin source.
type RemoteSource struct {
	url    string
	client interface{} // HTTP client
}

// NewRemoteSource creates a remote plugin source.
func NewRemoteSource(url string) *RemoteSource {
	return &RemoteSource{url: url}
}

// List lists plugins from remote source.
func (r *RemoteSource) List() ([]PluginMetadata, error) {
	// Placeholder: would fetch from remote API
	return []PluginMetadata{}, nil
}

// Download downloads plugin from remote source.
func (r *RemoteSource) Download(pluginID, version string) ([]byte, error) {
	// Placeholder: would download from remote URL
	return nil, fmt.Errorf("not implemented")
}

// ParsePluginID parses a plugin identifier.
func ParsePluginID(id string) (name string, version string, err error) {
	parts := strings.Split(id, "@")

	if len(parts) == 1 {
		return parts[0], "latest", nil
	}

	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("invalid plugin ID format: %s", id)
}

// FormatPluginID formats a plugin identifier.
func FormatPluginID(name, version string) string {
	if version == "" || version == "latest" {
		return name
	}
	return fmt.Sprintf("%s@%s", name, version)
}
