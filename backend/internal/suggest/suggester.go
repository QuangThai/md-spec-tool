package suggest

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// AISuggestion represents a single AI-generated suggestion (exported for handler compatibility)
type AISuggestion struct {
	Type       ai.SuggestionType `json:"type"`
	Severity   string            `json:"severity"` // info, warn, error
	Message    string            `json:"message"`
	RowRef     *int              `json:"row_ref,omitempty"`
	Field      string            `json:"field,omitempty"`
	Suggestion string            `json:"suggestion"`
}

// SuggestionRequest contains the input for AI analysis
type SuggestionRequest struct {
	SpecDoc  *converter.SpecDoc `json:"spec_doc"`
	Template string             `json:"template"`
}

// SuggestionResponse contains the AI analysis result
type SuggestionResponse struct {
	Suggestions []AISuggestion `json:"suggestions"`
	Error       string         `json:"error,omitempty"`
}

// Suggester provides AI-powered suggestions for spec documents
// Now uses ai.Service instead of direct HTTP calls for:
// - Retry logic with exponential backoff
// - Request timeout handling
// - Response caching
// - Consistent error handling
type Suggester struct {
	aiService ai.Service
}

// NewSuggester creates a new AI suggester instance using ai.Service
func NewSuggester(aiService ai.Service) *Suggester {
	return &Suggester{
		aiService: aiService,
	}
}

// IsConfigured returns true if the suggester has a valid AI service
func (s *Suggester) IsConfigured() bool {
	return s.aiService != nil && s.aiService.GetMode() != "off"
}

// GetSuggestions analyzes a spec document and returns improvement suggestions
func (s *Suggester) GetSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResponse, error) {
	if !s.IsConfigured() {
		return &SuggestionResponse{
			Error: "AI suggestions not configured",
		}, nil
	}

	if req.SpecDoc == nil || len(req.SpecDoc.Rows) == 0 {
		return &SuggestionResponse{
			Suggestions: []AISuggestion{},
		}, nil
	}

	// Build spec content for AI analysis
	specContent := buildSpecContent(req.SpecDoc)

	// Create AI request
	aiReq := ai.SuggestionsRequest{
		SpecContent: specContent,
		Template:    req.Template,
		RowCount:    len(req.SpecDoc.Rows),
	}

	// Call AI service (has retry/timeout/caching built-in)
	result, err := s.aiService.GetSuggestions(ctx, aiReq)
	if err != nil {
		return &SuggestionResponse{
			Error: fmt.Sprintf("Failed to get AI suggestions: %v", err),
		}, nil
	}

	// Convert ai.Suggestion to AISuggestion for backward compatibility
	suggestions := make([]AISuggestion, 0, len(result.Suggestions))
	for _, s := range result.Suggestions {
		suggestions = append(suggestions, AISuggestion{
			Type:       s.Type,
			Severity:   s.Severity,
			Message:    s.Message,
			RowRef:     s.RowRef,
			Field:      s.Field,
			Suggestion: s.Suggestion,
		})
	}

	return &SuggestionResponse{
		Suggestions: suggestions,
	}, nil
}

// buildSpecContent formats the spec document for AI analysis
func buildSpecContent(doc *converter.SpecDoc) string {
	var sb strings.Builder

	for i, row := range doc.Rows {
		sb.WriteString(fmt.Sprintf("\n--- Row %d ---\n", i+1))
		if row.ID != "" {
			sb.WriteString(fmt.Sprintf("ID: %s\n", row.ID))
		}
		if row.Feature != "" {
			sb.WriteString(fmt.Sprintf("Feature: %s\n", row.Feature))
		}
		if row.Scenario != "" {
			sb.WriteString(fmt.Sprintf("Scenario: %s\n", row.Scenario))
		}
		if row.Instructions != "" {
			sb.WriteString(fmt.Sprintf("Instructions: %s\n", row.Instructions))
		}
		if row.Inputs != "" {
			sb.WriteString(fmt.Sprintf("Inputs: %s\n", row.Inputs))
		}
		if row.Expected != "" {
			sb.WriteString(fmt.Sprintf("Expected: %s\n", row.Expected))
		}
		if row.Precondition != "" {
			sb.WriteString(fmt.Sprintf("Precondition: %s\n", row.Precondition))
		}
		if row.Priority != "" {
			sb.WriteString(fmt.Sprintf("Priority: %s\n", row.Priority))
		}
		if row.Notes != "" {
			sb.WriteString(fmt.Sprintf("Notes: %s\n", row.Notes))
		}
	}

	return sb.String()
}
