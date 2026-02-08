package converter

import (
	"fmt"
)

// RendererFactory creates appropriate Renderer based on format.
// Supported formats are only "spec" and "table".
type RendererFactory struct {
	templateRegistry *TemplateRegistry
}

// NewRendererFactory creates a new RendererFactory
func NewRendererFactory(registry *TemplateRegistry) *RendererFactory {
	return &RendererFactory{
		templateRegistry: registry,
	}
}

// CreateRenderer instantiates the correct renderer based on output type.
func (f *RendererFactory) CreateRenderer(template *TemplateConfig) (Renderer, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	outputType := template.Output.Type
	if outputType == "" {
		outputType = "spec" // Default to spec format
	}

	switch outputType {
	case "table":
		return NewTableRenderer(), nil

	case "spec":
		return NewSpecRenderer(), nil

	default:
		return nil, fmt.Errorf("unknown format: %s (supported: spec, table)", outputType)
	}
}

// NewRendererSimple creates a renderer by format name (no template required)
// Useful for Phase 4+ simple rendering
func NewRendererSimple(format string) (Renderer, error) {
	if format == "" {
		format = "spec"
	}

	switch format {
	case "table":
		return NewTableRenderer(), nil
	case "spec":
		return NewSpecRenderer(), nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: spec, table)", format)
	}
}

// ValidateOutputType checks if an output type is supported.
func (f *RendererFactory) ValidateOutputType(outputType string) error {
	if outputType == "" {
		return nil // Default is valid
	}

	switch outputType {
	case "spec", "table":
		return nil
	default:
		return fmt.Errorf("unknown format: %s (supported: spec, table)", outputType)
	}
}
