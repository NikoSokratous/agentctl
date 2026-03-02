package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads an agent config from a file.
func Load(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}

// Parse parses agent config from YAML bytes.
func Parse(data []byte) (*AgentConfig, error) {
	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	return &cfg, nil
}

// LoadWithPolicy loads the agent config and resolves the policy path.
func LoadWithPolicy(agentPath string) (*AgentConfig, string, error) {
	cfg, err := Load(agentPath)
	if err != nil {
		return nil, "", err
	}
	policyPath := cfg.Policy
	if policyPath != "" && !filepath.IsAbs(policyPath) {
		dir := filepath.Dir(agentPath)
		policyPath = filepath.Join(dir, policyPath)
	}
	return cfg, policyPath, nil
}
