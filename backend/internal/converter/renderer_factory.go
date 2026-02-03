package converter

import (
	"fmt"
)

// RendererFactory creates appropriate Renderer based on template output configuration
// This implements the factory pattern for Phase 4 output format abstraction
type RendererFactory struct {
	templateRegistry *TemplateRegistry
}

// NewRendererFactory creates a new RendererFactory
func NewRendererFactory(registry *TemplateRegistry) *RendererFactory {
	return &RendererFactory{
		templateRegistry: registry,
	}
}

// CreateRenderer instantiates the correct renderer based on template output type
// Returns error if output type is unknown or config is invalid
func (f *RendererFactory) CreateRenderer(template *TemplateConfig) (Renderer, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	outputType := template.Output.Type
	if outputType == "" {
		outputType = "test_spec_markdown" // Default
	}

	switch outputType {
	case "generic_table":
		return NewGenericTableRenderer(), nil

	case "test_spec_markdown", "test_spec":
		return NewTestSpecRenderer(template), nil

	case "row_cards":
		if template.Output.RowCards == nil {
			return nil, fmt.Errorf("row_cards output type requires RowCardsConfig in template")
		}
		return NewRowCardsRenderer(template), nil

	case "narrative":
		if template.Output.Narrative == nil {
			return nil, fmt.Errorf("narrative output type requires NarrativeConfig in template")
		}
		// TODO: Implement NarrativeRenderer in Phase 4.2
		return nil, fmt.Errorf("narrative output type not yet implemented")

	default:
		return nil, fmt.Errorf("unknown output type: %s", outputType)
	}
}

// ValidateOutputType checks if an output type is supported
func (f *RendererFactory) ValidateOutputType(outputType string) error {
	switch outputType {
	case "generic_table", "test_spec_markdown", "test_spec", "row_cards":
		return nil
	case "narrative":
		return fmt.Errorf("narrative output type not yet implemented")
	default:
		return fmt.Errorf("unknown output type: %s", outputType)
	}
}
