package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// TemplateRegistry manages template configurations
// It loads templates from embedded resources or filesystem
type TemplateRegistry struct {
	templates map[string]*TemplateConfig
	mu        sync.RWMutex
	basePath  string // Path to load external templates from
}

// NewTemplateRegistry creates a new template registry
// It loads the default test_spec_v1 template and discovers cookbook templates
func NewTemplateRegistry() *TemplateRegistry {
	reg := &TemplateRegistry{
		templates: make(map[string]*TemplateConfig),
		basePath:  "",
	}

	// Load default template
	defaultTemplate := reg.createDefaultTestSpecV1Template()
	reg.templates["test_spec_v1"] = defaultTemplate
	reg.templates["default"] = defaultTemplate // Alias to default

	// Load cookbook templates from backend/templates directory
	reg.loadCookbookTemplates()

	return reg
}

// NewTemplateRegistryWithPath creates a registry and sets the base path for external templates
func NewTemplateRegistryWithPath(basePath string) *TemplateRegistry {
	reg := NewTemplateRegistry()
	reg.basePath = basePath
	return reg
}

// LoadTemplate loads a template by name from registry or filesystem
// Returns error if template not found
func (r *TemplateRegistry) LoadTemplate(name string) (*TemplateConfig, error) {
	r.mu.RLock()
	if template, exists := r.templates[name]; exists {
		r.mu.RUnlock()
		return template, nil
	}
	r.mu.RUnlock()

	// Try to load from filesystem if base path is set
	if r.basePath != "" {
		return r.loadFromFile(name)
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// LoadTemplate loads a template by name or returns the default
func (r *TemplateRegistry) LoadTemplateOrDefault(name string) *TemplateConfig {
	if template, err := r.LoadTemplate(name); err == nil {
		return template
	}
	// Return default template
	template, _ := r.LoadTemplate("test_spec_v1")
	return template
}

// RegisterTemplate registers a template in the registry
func (r *TemplateRegistry) RegisterTemplate(template *TemplateConfig) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if errs := template.Validate(); len(errs) > 0 {
		return fmt.Errorf("template validation failed: %v", errs)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.templates[template.Name] = template
	return nil
}

// ListTemplates returns names of all registered templates
func (r *TemplateRegistry) ListTemplates() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.templates {
		if name != "default" { // Don't list alias
			names = append(names, name)
		}
	}
	return names
}

// loadFromFile loads a YAML template from filesystem
func (r *TemplateRegistry) loadFromFile(name string) (*TemplateConfig, error) {
	filePath := filepath.Join(r.basePath, name+".yaml")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file '%s': %w", filePath, err)
	}

	var template TemplateConfig
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML '%s': %w", filePath, err)
	}

	if errs := template.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("template validation failed for '%s': %v", filePath, errs)
	}

	// Cache the loaded template
	r.mu.Lock()
	r.templates[template.Name] = &template
	r.mu.Unlock()

	return &template, nil
}

// createDefaultTestSpecV1Template creates the default test_spec_v1 template
// Uses HeaderSynonyms from column_map.go to avoid duplication
func (r *TemplateRegistry) createDefaultTestSpecV1Template() *TemplateConfig {
	// Convert column_map.go HeaderSynonyms to template format
	headerSynonyms := make(map[string][]string)
	for synonym, field := range HeaderSynonyms {
		fieldName := fieldToName(field)
		if _, exists := headerSynonyms[fieldName]; !exists {
			headerSynonyms[fieldName] = []string{}
		}
		headerSynonyms[fieldName] = append(headerSynonyms[fieldName], synonym)
	}

	return &TemplateConfig{
		Name:           "test_spec_v1",
		Description:    "Legacy test specification format with 180+ column mappings",
		HeaderSynonyms: headerSynonyms,
		RequiredFields: []string{"scenario"},
		Output: TemplateOutputConfig{
			Type:              "test_spec_markdown",
			UnmappedColumns:   "append_section",
			PreserveAllFields: true,
		},
		Metadata: map[string]interface{}{
			"version": "1.0",
			"source":  "embedded_default",
		},
	}
}

// fieldToName converts a CanonicalField to its string name
func fieldToName(field CanonicalField) string {
	switch field {
	case FieldID:
		return "id"
	case FieldFeature:
		return "feature"
	case FieldScenario:
		return "scenario"
	case FieldInstructions:
		return "instructions"
	case FieldInputs:
		return "inputs"
	case FieldExpected:
		return "expected"
	case FieldPrecondition:
		return "precondition"
	case FieldPriority:
		return "priority"
	case FieldType:
		return "type"
	case FieldStatus:
		return "status"
	case FieldEndpoint:
		return "endpoint"
	case FieldNotes:
		return "notes"
	case FieldNo:
		return "no"
	case FieldItemName:
		return "item_name"
	case FieldItemType:
		return "item_type"
	case FieldRequiredOptional:
		return "required_optional"
	case FieldInputRestrictions:
		return "input_restrictions"
	case FieldDisplayConditions:
		return "display_conditions"
	case FieldAction:
		return "action"
	case FieldNavigationDest:
		return "navigation_destination"
	default:
		return "unknown"
	}
}

// GetDefaultTemplate returns the default test_spec_v1 template
func (r *TemplateRegistry) GetDefaultTemplate() *TemplateConfig {
	template, _ := r.LoadTemplate("test_spec_v1")
	return template
}

// loadCookbookTemplates discovers and loads cookbook templates from backend/templates directory
func (r *TemplateRegistry) loadCookbookTemplates() {
	// Look for backend/templates directory using env var or well-known paths
	possiblePaths := []string{
		os.Getenv("TEMPLATE_DIR"), // Check env var first
		"backend/templates",
		"./backend/templates",
		"../backend/templates",
	}

	for _, basePath := range possiblePaths {
		if basePath == "" {
			continue // Skip empty paths from env var
		}
		
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue // Try next path
		}

		// Load all .yaml files in the directory
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			filePath := filepath.Join(basePath, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue // Skip files we can't read
			}

			var config TemplateConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				continue // Skip files we can't parse
			}

			// Skip invalid templates
			if config.Name == "" {
				continue
			}

			// Cache the loaded template
			r.mu.Lock()
			r.templates[config.Name] = &config
			r.mu.Unlock()
		}

		// Success - stop trying other paths
		break
	}
}
