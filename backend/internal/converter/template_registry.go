package converter

import (
	"fmt"
	"sync"
)

// TemplateRegistry manages template configurations
// It loads templates from embedded resources
type TemplateRegistry struct {
	templates map[string]*TemplateConfig
	mu        sync.RWMutex
}

// NewTemplateRegistry creates a new template registry.
// Only "spec" and "table" templates are registered for conversion.
func NewTemplateRegistry() *TemplateRegistry {
	reg := &TemplateRegistry{
		templates: make(map[string]*TemplateConfig),
	}

	reg.templates["spec"] = reg.createSpecTemplate()
	reg.templates["table"] = reg.createTableTemplate()

	return reg
}

// LoadTemplate loads a template by name from registry
// Returns error if template not found
func (r *TemplateRegistry) LoadTemplate(name string) (*TemplateConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if template, exists := r.templates[name]; exists {
		return template, nil
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// LoadTemplateOrDefault loads a template by name or returns the default spec template.
func (r *TemplateRegistry) LoadTemplateOrDefault(name string) *TemplateConfig {
	if template, err := r.LoadTemplate(name); err == nil {
		return template
	}
	template, _ := r.LoadTemplate("spec")
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
		names = append(names, name)
	}
	return names
}

// createSpecTemplate creates the default spec template.
// Uses HeaderSynonyms from column_map.go to avoid duplication.
func (r *TemplateRegistry) createSpecTemplate() *TemplateConfig {
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
		Name:           "spec",
		Description:    "Structured specification output",
		HeaderSynonyms: headerSynonyms,
		RequiredFields: []string{"scenario"},
		Output: TemplateOutputConfig{
			Type:              "spec",
			UnmappedColumns:   "append_section",
			PreserveAllFields: true,
		},
		Metadata: map[string]interface{}{
			"version": "2.0",
			"source":  "embedded_spec",
		},
	}
}

// createTableTemplate creates the default table template.
func (r *TemplateRegistry) createTableTemplate() *TemplateConfig {
	headerSynonyms := make(map[string][]string)
	for synonym, field := range HeaderSynonyms {
		fieldName := fieldToName(field)
		if _, exists := headerSynonyms[fieldName]; !exists {
			headerSynonyms[fieldName] = []string{}
		}
		headerSynonyms[fieldName] = append(headerSynonyms[fieldName], synonym)
	}

	return &TemplateConfig{
		Name:           "table",
		Description:    "Simple markdown table output",
		HeaderSynonyms: headerSynonyms,
		RequiredFields: []string{},
		Output: TemplateOutputConfig{
			Type:              "table",
			UnmappedColumns:   "ignore",
			PreserveAllFields: true,
		},
		Metadata: map[string]interface{}{
			"version": "2.0",
			"source":  "embedded_table",
		},
	}
}

// fieldToName converts a CanonicalField to its string name
func fieldToName(field CanonicalField) string {
	switch field {
	case FieldID:
		return "id"
	case FieldTitle:
		return "title"
	case FieldDescription:
		return "description"
	case FieldAcceptance:
		return "acceptance_criteria"
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
	case FieldMethod:
		return "method"
	case FieldParameters:
		return "parameters"
	case FieldResponse:
		return "response"
	case FieldStatusCode:
		return "status_code"
	case FieldNotes:
		return "notes"
	case FieldComponent:
		return "component"
	case FieldAssignee:
		return "assignee"
	case FieldCategory:
		return "category"
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

// GetDefaultTemplate returns the default spec template.
func (r *TemplateRegistry) GetDefaultTemplate() *TemplateConfig {
	template, _ := r.LoadTemplate("spec")
	return template
}
