package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/converter"
)

// SuggestionType represents the type of AI suggestion
type SuggestionType string

const (
	SuggestionMissingField     SuggestionType = "missing_field"
	SuggestionVagueDescription SuggestionType = "vague_description"
	SuggestionIncompleteSteps  SuggestionType = "incomplete_steps"
	SuggestionFormatting       SuggestionType = "formatting"
	SuggestionCoverage         SuggestionType = "coverage"
)

// AISuggestion represents a single AI-generated suggestion
type AISuggestion struct {
	Type       SuggestionType `json:"type"`
	Severity   string         `json:"severity"` // info, warn, error
	Message    string         `json:"message"`
	RowRef     *int           `json:"row_ref,omitempty"`
	Field      string         `json:"field,omitempty"`
	Suggestion string         `json:"suggestion"`
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
type Suggester struct {
	apiKey string
	model  string
}

// NewSuggester creates a new AI suggester instance
func NewSuggester(apiKey, model string) *Suggester {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Suggester{
		apiKey: apiKey,
		model:  model,
	}
}

// IsConfigured returns true if the suggester has an API key
func (s *Suggester) IsConfigured() bool {
	return s.apiKey != ""
}

// GetSuggestions analyzes a spec document and returns improvement suggestions
func (s *Suggester) GetSuggestions(req *SuggestionRequest) (*SuggestionResponse, error) {
	if !s.IsConfigured() {
		return &SuggestionResponse{
			Error: "OpenAI API key not configured",
		}, nil
	}

	if req.SpecDoc == nil || len(req.SpecDoc.Rows) == 0 {
		return &SuggestionResponse{
			Suggestions: []AISuggestion{},
		}, nil
	}

	// Build the prompt with spec data
	prompt := buildPrompt(req.SpecDoc, req.Template)

	// Call OpenAI API
	suggestions, err := s.callOpenAI(prompt)
	if err != nil {
		return &SuggestionResponse{
			Error: fmt.Sprintf("Failed to get AI suggestions: %v", err),
		}, nil
	}

	return &SuggestionResponse{
		Suggestions: suggestions,
	}, nil
}

func buildPrompt(doc *converter.SpecDoc, template string) string {
	var sb strings.Builder

	sb.WriteString(`You are a QA expert analyzing a test specification document. Review the following spec rows and identify quality issues.

For each issue found, provide a suggestion with:
- type: one of "missing_field", "vague_description", "incomplete_steps", "formatting", "coverage"
- severity: "info" for minor improvements, "warn" for important issues, "error" for critical problems
- message: a brief description of the issue
- row_ref: the row number (1-based) if applicable, or null for general issues
- field: the field name if applicable (e.g., "expected", "instructions")
- suggestion: specific actionable improvement text

Focus on these quality issues:
1. Missing required fields (ID, Expected results, Instructions)
2. Vague or incomplete descriptions (e.g., "do something", "check result", single word descriptions)
3. Missing preconditions for complex scenarios
4. Incomplete test steps that lack specific actions
5. Missing edge cases or negative test scenarios

Template: ` + template + `

Spec Rows:
`)

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

	sb.WriteString(`

Respond with a JSON object containing a "suggestions" array. Each suggestion should have the fields described above. Limit to the top 10 most important suggestions. If the spec is high quality with no significant issues, return an empty suggestions array.

Example response:
{
  "suggestions": [
    {
      "type": "vague_description",
      "severity": "warn",
      "message": "Expected result is too vague",
      "row_ref": 3,
      "field": "expected",
      "suggestion": "Instead of 'check result', specify exactly what should be verified, e.g., 'Verify success message \"Order confirmed\" is displayed and order ID is generated'"
    }
  ]
}`)

	return sb.String()
}

// OpenAI API structures
type openAIRequest struct {
	Model          string          `json:"model"`
	Messages       []openAIMessage `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type aiSuggestionsResponse struct {
	Suggestions []AISuggestion `json:"suggestions"`
}

func (s *Suggester) callOpenAI(prompt string) ([]AISuggestion, error) {
	reqBody := openAIRequest{
		Model: s.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are a QA expert that analyzes test specifications and provides actionable improvement suggestions. Always respond with valid JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
		MaxTokens:      2000,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the suggestions from the response content
	content := openAIResp.Choices[0].Message.Content
	var suggestionsResp aiSuggestionsResponse
	if err := json.Unmarshal([]byte(content), &suggestionsResp); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w (content: %s)", err, content)
	}

	// Validate and normalize suggestions
	validSuggestions := make([]AISuggestion, 0, len(suggestionsResp.Suggestions))
	for _, s := range suggestionsResp.Suggestions {
		// Validate type
		switch s.Type {
		case SuggestionMissingField, SuggestionVagueDescription, SuggestionIncompleteSteps, SuggestionFormatting, SuggestionCoverage:
			// Valid type
		default:
			s.Type = SuggestionVagueDescription // Default fallback
		}

		// Validate severity
		switch s.Severity {
		case "info", "warn", "error":
			// Valid severity
		default:
			s.Severity = "info" // Default fallback
		}

		// Skip empty suggestions
		if s.Message == "" && s.Suggestion == "" {
			continue
		}

		validSuggestions = append(validSuggestions, s)
	}

	return validSuggestions, nil
}
