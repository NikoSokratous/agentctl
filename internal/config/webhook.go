package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WebhookConfig represents a webhook endpoint configuration.
type WebhookConfig struct {
	Path         string            `yaml:"path" json:"path"`
	Agent        string            `yaml:"agent" json:"agent"`
	GoalTemplate string            `yaml:"goal_template" json:"goal_template"`
	AuthSecret   string            `yaml:"auth_secret" json:"auth_secret"`
	CallbackURL  string            `yaml:"callback_url" json:"callback_url,omitempty"`
	Headers      map[string]string `yaml:"headers" json:"headers,omitempty"`
}

// Webhooks is a collection of webhook configurations.
type Webhooks struct {
	Webhooks []WebhookConfig `yaml:"webhooks" json:"webhooks"`
}

// Validate checks if the webhook configuration is valid.
func (w *WebhookConfig) Validate() error {
	if w.Path == "" {
		return fmt.Errorf("webhook path is required")
	}
	if w.Path[0] != '/' {
		return fmt.Errorf("webhook path must start with /")
	}
	if w.Agent == "" {
		return fmt.Errorf("webhook agent is required")
	}
	if w.GoalTemplate == "" {
		return fmt.Errorf("webhook goal_template is required")
	}
	return nil
}

// ResolveSecret resolves environment variable references in the secret.
// Supports format: $ENV_VAR or ${ENV_VAR}
func (w *WebhookConfig) ResolveSecret() string {
	secret := w.AuthSecret

	// Handle $ENV_VAR format
	if len(secret) > 0 && secret[0] == '$' {
		envVar := secret[1:]
		// Handle ${ENV_VAR} format
		if len(envVar) > 0 && envVar[0] == '{' && envVar[len(envVar)-1] == '}' {
			envVar = envVar[1 : len(envVar)-1]
		}
		return os.Getenv(envVar)
	}

	return secret
}

// LoadWebhooks loads webhook configurations from a YAML file.
func LoadWebhooks(path string) (*Webhooks, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read webhooks file: %w", err)
	}

	var webhooks Webhooks
	if err := yaml.Unmarshal(data, &webhooks); err != nil {
		return nil, fmt.Errorf("parse webhooks: %w", err)
	}

	// Validate all webhooks
	for i, wh := range webhooks.Webhooks {
		if err := wh.Validate(); err != nil {
			return nil, fmt.Errorf("webhook %d: %w", i, err)
		}
	}

	return &webhooks, nil
}
