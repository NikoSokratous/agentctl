package registry

import (
	"fmt"
	"strings"
)

// SecurityValidator validates plugin security.
type SecurityValidator struct {
	allowedPermissions map[Permission]bool
	maxPermissions     int
	requireSignature   bool
}

// NewSecurityValidator creates a new security validator.
func NewSecurityValidator(config *SecurityConfig) *SecurityValidator {
	return &SecurityValidator{
		allowedPermissions: config.AllowedPermissions,
		maxPermissions:     config.MaxPermissions,
		requireSignature:   config.RequireSignature,
	}
}

// SecurityConfig defines security validation configuration.
type SecurityConfig struct {
	AllowedPermissions map[Permission]bool
	MaxPermissions     int
	RequireSignature   bool
	TrustedAuthors     []string
	BlockedPlugins     []string
}

// DefaultSecurityConfig returns default security configuration.
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedPermissions: map[Permission]bool{
			PermissionNetwork:   true,
			PermissionFileRead:  true,
			PermissionFileWrite: false, // Disabled by default
			PermissionFileExec:  false, // Disabled by default
			PermissionEnvRead:   true,
			PermissionEnvWrite:  false, // Disabled by default
			PermissionDatabase:  true,
			PermissionAPI:       true,
			PermissionSecrets:   false, // Disabled by default
		},
		MaxPermissions:   5,
		RequireSignature: false,
		TrustedAuthors:   []string{},
		BlockedPlugins:   []string{},
	}
}

// Validate validates plugin security.
func (v *SecurityValidator) Validate(meta *PluginMetadata) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}

	// Check signature if required
	if v.requireSignature && meta.Signature == "" {
		result.Errors = append(result.Errors, "plugin signature is required")
		result.Valid = false
	}

	// Validate permissions
	if len(meta.Permissions) > v.maxPermissions {
		result.Errors = append(result.Errors,
			fmt.Sprintf("too many permissions requested: %d (max: %d)",
				len(meta.Permissions), v.maxPermissions))
		result.Valid = false
	}

	for _, perm := range meta.Permissions {
		if !v.allowedPermissions[perm] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("permission not allowed: %s", perm))
			result.Valid = false
		}

		// Add warnings for dangerous permissions
		if isDangerousPermission(perm) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("dangerous permission requested: %s", perm))
		}
	}

	// Validate checksum
	if meta.Checksum == "" {
		result.Errors = append(result.Errors, "missing checksum")
		result.Valid = false
	}

	// Check for suspicious patterns
	if containsSuspiciousPattern(meta.Description) {
		result.Warnings = append(result.Warnings, "description contains suspicious content")
	}

	if containsSuspiciousPattern(meta.Name) {
		result.Warnings = append(result.Warnings, "plugin name contains suspicious content")
	}

	// Validate dependencies
	if len(meta.Dependencies) > 10 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("high number of dependencies: %d", len(meta.Dependencies)))
	}

	return result, nil
}

// ValidationResult contains validation results.
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

// ValidateChecksum validates artifact checksum.
func ValidateChecksum(data []byte, expectedChecksum string) bool {
	actualChecksum := CalculateChecksum(data)
	return actualChecksum == expectedChecksum
}

// ValidateSignature validates plugin signature.
func ValidateSignature(data []byte, signature string, publicKey []byte) (bool, error) {
	// Placeholder: would implement GPG/RSA signature verification
	// For now, just check signature is not empty
	return signature != "", nil
}

// isDangerousPermission checks if a permission is considered dangerous.
func isDangerousPermission(perm Permission) bool {
	dangerous := map[Permission]bool{
		PermissionFileWrite: true,
		PermissionFileExec:  true,
		PermissionEnvWrite:  true,
		PermissionSecrets:   true,
	}
	return dangerous[perm]
}

// containsSuspiciousPattern checks for suspicious content.
func containsSuspiciousPattern(text string) bool {
	suspicious := []string{
		"eval(",
		"exec(",
		"system(",
		"__import__",
		"subprocess",
		"os.system",
		"shell=True",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range suspicious {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}

	return false
}

// ScanForMalware performs basic malware scanning.
func ScanForMalware(data []byte) (*ScanResult, error) {
	result := &ScanResult{
		Clean:    true,
		Threats:  make([]string, 0),
		Warnings: make([]string, 0),
	}

	content := string(data)

	// Simple pattern matching for known malicious patterns
	maliciousPatterns := map[string]string{
		"eval(":           "Dynamic code execution",
		"exec(":           "Dynamic code execution",
		"__import__":      "Dynamic imports",
		"subprocess.call": "System command execution",
		"os.system":       "System command execution",
		"rm -rf":          "Destructive file operation",
		"dd if=/dev/zero": "Disk wipe operation",
	}

	for pattern, description := range maliciousPatterns {
		if strings.Contains(content, pattern) {
			result.Threats = append(result.Threats,
				fmt.Sprintf("%s detected: %s", description, pattern))
			result.Clean = false
		}
	}

	// Check for encoded content (potential obfuscation)
	if strings.Contains(content, "base64") && strings.Contains(content, "decode") {
		result.Warnings = append(result.Warnings, "base64 encoded content detected")
	}

	return result, nil
}

// ScanResult contains malware scan results.
type ScanResult struct {
	Clean    bool     `json:"clean"`
	Threats  []string `json:"threats,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
