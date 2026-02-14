package suggest

import (
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// BuildSpecContent exposes buildSpecContent for external tests.
func BuildSpecContent(doc *converter.SpecDoc) string {
	return buildSpecContent(doc)
}

// AIService exposes aiService for external tests.
func (s *Suggester) AIService() ai.Service {
	if s == nil {
		return nil
	}
	return s.aiService
}
