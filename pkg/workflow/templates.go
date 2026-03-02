package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Template represents a workflow template.
type Template struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Author      string                 `yaml:"author"`
	Category    string                 `yaml:"category"` // data-pipeline, code-review, research, etc.
	Parameters  []TemplateParam        `yaml:"parameters"`
	Workflow    map[string]interface{} `yaml:"workflow"`
}

// TemplateParam defines a template parameter.
type TemplateParam struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Type        string      `yaml:"type"` // string, int, bool, list
	Required    bool        `yaml:"required"`
	Default     interface{} `yaml:"default,omitempty"`
}

// TemplateRegistry manages workflow templates.
type TemplateRegistry struct {
	templates map[string]*Template
}

// NewTemplateRegistry creates a new template registry.
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*Template),
	}
}

// LoadEmbeddedTemplates loads built-in templates from examples directory.
func (r *TemplateRegistry) LoadEmbeddedTemplates() error {
	// Try to load from examples/workflows/templates
	templatesDir := "examples/workflows/templates"

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// No templates directory, silently return
		return nil
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return fmt.Errorf("read templates dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(templatesDir, entry.Name()))
		if err != nil {
			continue
		}

		var tmpl Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			continue
		}

		r.templates[tmpl.Name] = &tmpl
	}

	return nil
}

// LoadFromFile loads a template from a file.
func (r *TemplateRegistry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read template file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	r.templates[tmpl.Name] = &tmpl
	return nil
}

// Get retrieves a template by name.
func (r *TemplateRegistry) Get(name string) (*Template, error) {
	tmpl, exists := r.templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

// List returns all available templates.
func (r *TemplateRegistry) List() []Template {
	templates := make([]Template, 0, len(r.templates))
	for _, tmpl := range r.templates {
		templates = append(templates, *tmpl)
	}
	return templates
}

// Instantiate creates a workflow from a template with parameters.
func (r *TemplateRegistry) Instantiate(templateName string, params map[string]interface{}) ([]byte, error) {
	tmpl, err := r.Get(templateName)
	if err != nil {
		return nil, err
	}

	// Validate required parameters
	for _, param := range tmpl.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				if param.Default != nil {
					params[param.Name] = param.Default
				} else {
					return nil, fmt.Errorf("required parameter missing: %s", param.Name)
				}
			}
		}
	}

	// Convert workflow to YAML string
	workflowData, err := yaml.Marshal(tmpl.Workflow)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}

	// Apply template parameters
	tmplStr := string(workflowData)
	t, err := template.New("workflow").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, params); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return []byte(buf.String()), nil
}

// ValidateTemplate validates a template structure.
func ValidateTemplate(tmpl *Template) error {
	if tmpl.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if tmpl.Workflow == nil {
		return fmt.Errorf("template workflow is required")
	}

	// Validate parameters
	paramNames := make(map[string]bool)
	for _, param := range tmpl.Parameters {
		if param.Name == "" {
			return fmt.Errorf("parameter name is required")
		}
		if paramNames[param.Name] {
			return fmt.Errorf("duplicate parameter: %s", param.Name)
		}
		paramNames[param.Name] = true

		// Validate type
		validTypes := map[string]bool{
			"string": true,
			"int":    true,
			"bool":   true,
			"list":   true,
			"map":    true,
		}
		if !validTypes[param.Type] {
			return fmt.Errorf("invalid parameter type: %s", param.Type)
		}
	}

	return nil
}
