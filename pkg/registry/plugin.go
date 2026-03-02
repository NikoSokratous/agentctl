package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// PluginMetadata represents plugin information.
type PluginMetadata struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Version      string                 `json:"version" yaml:"version"`
	Author       string                 `json:"author" yaml:"author"`
	Description  string                 `json:"description" yaml:"description"`
	Type         PluginType             `json:"type" yaml:"type"`
	Runtime      RuntimeType            `json:"runtime" yaml:"runtime"`
	License      string                 `json:"license" yaml:"license"`
	Repository   string                 `json:"repository" yaml:"repository"`
	Homepage     string                 `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Dependencies []Dependency           `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Permissions  []Permission           `json:"permissions" yaml:"permissions"`
	Checksum     string                 `json:"checksum" yaml:"checksum"`
	Signature    string                 `json:"signature,omitempty" yaml:"signature,omitempty"`
	CreatedAt    time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" yaml:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PluginType defines the type of plugin.
type PluginType string

const (
	PluginTypeTool        PluginType = "tool"
	PluginTypeAgent       PluginType = "agent"
	PluginTypeIntegration PluginType = "integration"
	PluginTypeMemory      PluginType = "memory"
	PluginTypeModel       PluginType = "model"
)

// RuntimeType defines the plugin runtime.
type RuntimeType string

const (
	RuntimeGo     RuntimeType = "go"
	RuntimeWASM   RuntimeType = "wasm"
	RuntimePython RuntimeType = "python"
	RuntimeNode   RuntimeType = "node"
)

// Dependency represents a plugin dependency.
type Dependency struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Type    string `json:"type,omitempty" yaml:"type,omitempty"` // plugin, system, library
}

// Permission represents a plugin permission.
type Permission string

const (
	PermissionNetwork   Permission = "network"
	PermissionFileRead  Permission = "file:read"
	PermissionFileWrite Permission = "file:write"
	PermissionFileExec  Permission = "file:exec"
	PermissionEnvRead   Permission = "env:read"
	PermissionEnvWrite  Permission = "env:write"
	PermissionDatabase  Permission = "database"
	PermissionAPI       Permission = "api"
	PermissionSecrets   Permission = "secrets"
)

// PluginStats represents plugin statistics.
type PluginStats struct {
	PluginID     string    `json:"plugin_id"`
	Downloads    int64     `json:"downloads"`
	Rating       float64   `json:"rating"`
	ReviewCount  int       `json:"review_count"`
	LastDownload time.Time `json:"last_download"`
	TrendScore   float64   `json:"trend_score"`
}

// PluginReview represents a plugin review.
type PluginReview struct {
	ID        string    `json:"id"`
	PluginID  string    `json:"plugin_id"`
	UserID    string    `json:"user_id"`
	Rating    int       `json:"rating"` // 1-5
	Comment   string    `json:"comment,omitempty"`
	Helpful   int       `json:"helpful"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PluginManifest is the manifest file structure.
type PluginManifest struct {
	Metadata    PluginMetadata    `json:"metadata" yaml:"metadata"`
	Entrypoint  string            `json:"entrypoint" yaml:"entrypoint"`
	Files       []string          `json:"files" yaml:"files"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Build       *BuildConfig      `json:"build,omitempty" yaml:"build,omitempty"`
}

// BuildConfig defines build configuration.
type BuildConfig struct {
	Command string   `json:"command" yaml:"command"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty"`
	Env     []string `json:"env,omitempty" yaml:"env,omitempty"`
}

// CalculateChecksum computes SHA-256 checksum of data.
func CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// ValidateMetadata validates plugin metadata.
func ValidateMetadata(meta *PluginMetadata) error {
	if meta.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if meta.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if meta.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if meta.Author == "" {
		return fmt.Errorf("plugin author is required")
	}

	// Validate type
	validTypes := map[PluginType]bool{
		PluginTypeTool:        true,
		PluginTypeAgent:       true,
		PluginTypeIntegration: true,
		PluginTypeMemory:      true,
		PluginTypeModel:       true,
	}
	if !validTypes[meta.Type] {
		return fmt.Errorf("invalid plugin type: %s", meta.Type)
	}

	// Validate runtime
	validRuntimes := map[RuntimeType]bool{
		RuntimeGo:     true,
		RuntimeWASM:   true,
		RuntimePython: true,
		RuntimeNode:   true,
	}
	if !validRuntimes[meta.Runtime] {
		return fmt.Errorf("invalid runtime: %s", meta.Runtime)
	}

	// Validate permissions
	validPerms := map[Permission]bool{
		PermissionNetwork:   true,
		PermissionFileRead:  true,
		PermissionFileWrite: true,
		PermissionFileExec:  true,
		PermissionEnvRead:   true,
		PermissionEnvWrite:  true,
		PermissionDatabase:  true,
		PermissionAPI:       true,
		PermissionSecrets:   true,
	}

	for _, perm := range meta.Permissions {
		if !validPerms[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	return nil
}

// SearchFilter defines search criteria for plugins.
type SearchFilter struct {
	Query     string      `json:"query,omitempty"`
	Type      PluginType  `json:"type,omitempty"`
	Runtime   RuntimeType `json:"runtime,omitempty"`
	Author    string      `json:"author,omitempty"`
	MinRating float64     `json:"min_rating,omitempty"`
	Tags      []string    `json:"tags,omitempty"`
	SortBy    string      `json:"sort_by,omitempty"`    // name, rating, downloads, created
	SortOrder string      `json:"sort_order,omitempty"` // asc, desc
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

// SearchResult contains search results.
type SearchResult struct {
	Plugins []PluginMetadata `json:"plugins"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}
