package policy

import (
	"time"
)

// PolicyVersion represents a versioned policy document.
type PolicyVersion struct {
	ID          string            `json:"id"`
	PolicyName  string            `json:"policy_name"`
	Version     string            `json:"version"`
	Content     []byte            `json:"content"`
	Format      string            `json:"format"` // "yaml", "json"
	Author      string            `json:"author"`
	Changelog   string            `json:"changelog"`
	EffectiveAt time.Time         `json:"effective_at"`
	CreatedAt   time.Time         `json:"created_at"`
	Supersedes  string            `json:"supersedes,omitempty"` // Previous version ID
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PolicyVersionMetadata holds minimal version info for listings.
type PolicyVersionMetadata struct {
	PolicyName  string    `json:"policy_name"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Changelog   string    `json:"changelog"`
	EffectiveAt time.Time `json:"effective_at"`
	CreatedAt   time.Time `json:"created_at"`
	Active      bool      `json:"active"`
}

// PolicyDiff represents the difference between two policy versions.
type PolicyDiff struct {
	FromVersion string              `json:"from_version"`
	ToVersion   string              `json:"to_version"`
	Changes     []PolicyChange      `json:"changes"`
	Summary     PolicyChangeSummary `json:"summary"`
}

// PolicyChange represents a single change in a policy.
type PolicyChange struct {
	Type        string `json:"type"` // "added", "removed", "modified"
	Section     string `json:"section"`
	Description string `json:"description"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
	Impact      string `json:"impact"` // "breaking", "major", "minor"
}

// PolicyChangeSummary summarizes changes between versions.
type PolicyChangeSummary struct {
	TotalChanges    int `json:"total_changes"`
	Added           int `json:"added"`
	Removed         int `json:"removed"`
	Modified        int `json:"modified"`
	BreakingChanges int `json:"breaking_changes"`
}

// VersionConstraint represents a semantic version constraint.
type VersionConstraint struct {
	Operator string // "=", "!=", ">", ">=", "<", "<=", "~", "^"
	Version  string
}

// ParseVersion parses a semantic version string.
func ParseVersion(v string) (major, minor, patch int, err error) {
	// Simple implementation - can be enhanced with full semver library
	_, err = time.Parse("2006.01.02", v) // Also support date-based versions
	return
}

// CompareVersions compares two semantic versions.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Simple string comparison for now - can be enhanced
	if v1 == v2 {
		return 0
	}
	if v1 < v2 {
		return -1
	}
	return 1
}
